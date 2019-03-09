package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"google.golang.org/api/sheets/v4"

	"github.com/apex/log/handlers/text"
	"github.com/hink/SquadSheets/internal/pkg/config"

	"github.com/apex/log"
	"github.com/urfave/cli"
)

// CLIOpts command line options
type CLIOpts struct {
	ConfigPath     string
	LogPath        string
	Verbose        bool
	SquadConfigDir string
	ASCWhitelist   bool
}

// SheetProps allows quick reference to sheet properties
type SheetProps struct {
	AdminRoles *sheets.SheetProperties
	AdminVkng  *sheets.SheetProperties
	AdminOther *sheets.SheetProperties
	Whitelist  *sheets.SheetProperties
}

var cfg *config.Config

func main() {
	app := initApp()

	app.Action = func(c *cli.Context) error {
		// Parse CLI Args
		opts, err := validateArgs(c)
		if err != nil {
			log.Fatal(err.Error())
		}

		// configure logging
		logLevel := log.InfoLevel
		if opts.Verbose {
			logLevel = log.DebugLevel
		}
		log.SetLevel(logLevel)

		if opts.LogPath != "" {
			logFile, err := os.OpenFile(opts.LogPath, os.O_WRONLY|os.O_CREATE, 0755)
			if err != nil {
				log.Fatal(err.Error())
			}
			defer logFile.Close()
			log.SetHandler(text.New(logFile))
		} else {
			log.SetHandler(text.New(os.Stderr))
		}

		// Parse Configuration File
		log.WithField("path", opts.ConfigPath).Info("loading configuration")

		cfg, err = config.Load(opts.ConfigPath)
		if err != nil {
			log.Fatal(err.Error())
		}

		fmt.Println("xxx", opts.ConfigPath)
		// Initialize Sheets client
		log.Info("Initializing Google Sheets client")
		sheetsSrv, err := getSheetsService(cfg.Sheets.SecretJSONPath)
		if err != nil {
			log.Fatal(err.Error())
		}

		// Get sheet properties
		sheetProps := new(SheetProps)
		log.WithField("sheetid", cfg.Sheets.SheetID).Info("getting Google sheet properties")
		props, err := getSheetProperties(sheetsSrv, cfg.Sheets.SheetID)
		if err != nil {
			log.WithFields(log.Fields{
				"sheetid": cfg.Sheets.SheetID,
				"error":   err.Error(),
			}).Fatal("failed to get Google sheet properties")
		}
		for _, v := range props {
			switch strings.ToLower(v.Properties.Title) {
			case "admin_roles":
				sheetProps.AdminRoles = v.Properties
			case "admin_vkng":
				sheetProps.AdminVkng = v.Properties
			case "admin_valhalla":
				sheetProps.AdminOther = v.Properties
			case "whitelist":
				sheetProps.Whitelist = v.Properties
			default:
				// do nothing
			}
		}
		if sheetProps.AdminRoles == nil {
			log.WithField("sheet", "Customers").Fatal("could not find sheet")
		}
		if sheetProps.AdminVkng == nil {
			log.WithField("sheet", "DDOS").Fatal("could not find sheet")
		}
		if sheetProps.AdminOther == nil {
			log.WithField("sheet", "IPRM").Fatal("could not find sheet")
		}
		if sheetProps.Whitelist == nil {
			log.WithField("sheet", "NIDS").Fatal("could not find sheet")
		}

		return WriteAdminsFile(sheetsSrv, cfg.Sheets.SheetID, opts.SquadConfigDir, opts.ASCWhitelist)
	}

	app.Run(os.Args)
}

func initApp() *cli.App {
	app := cli.NewApp()
	app.Name = "squadsheets"
	app.Usage = "syncs google sheets with admin data to squad admin files"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "configuration file",
			Value: "squadsheets.cfg",
		},
		cli.StringFlag{
			Name:  "configdir, d",
			Usage: "Squad configuration directory",
		},
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "verbose logging",
		},
		cli.StringFlag{
			Name:  "log, l",
			Usage: "log file",
		},
		cli.BoolFlag{
			Name:  "whitelist, w",
			Usage: "ASC Whitelist",
		},
	}
	return app
}

func validateArgs(c *cli.Context) (*CLIOpts, error) {
	configDir := c.String("configdir")
	if configDir == "" {
		return nil, errors.New("no squad configuration directory provided")
	}

	return &CLIOpts{
		ConfigPath:     c.String("config"),
		SquadConfigDir: c.String("configdir"),
		LogPath:        c.String("log"),
		Verbose:        c.Bool("verbose"),
		ASCWhitelist:   c.Bool("whitelist"),
	}, nil
}
