package main

import (
	"errors"
	"os"

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

		// Initialize Sheets client
		log.Info("Initializing Google Sheets client")
		sheetsSrv, err := getSheetsService(cfg.Sheets.SecretJSONPath)
		if err != nil {
			log.Fatal(err.Error())
		}

		// Get sheet properties
		sheetProps := make(map[string]*sheets.SheetProperties)
		log.WithField("sheetid", cfg.Sheets.SheetID).Info("getting Google sheet properties")
		props, err := getSheetProperties(sheetsSrv, cfg.Sheets.SheetID)
		if err != nil {
			log.WithFields(log.Fields{
				"sheetid": cfg.Sheets.SheetID,
				"error":   err.Error(),
			}).Fatal("failed to get Google sheet properties")
		}

		for _, v := range props {
			// roles
			if v.Properties.Title == cfg.Sheets.SheetAdminRoles {
				sheetProps[cfg.Sheets.SheetAdminRoles] = v.Properties
				continue
			}
			// whitelist
			for _, s := range cfg.Sheets.SheetsWhitelist {
				if v.Properties.Title == s {
					sheetProps[s] = v.Properties
					continue
				}
			}
			// Admins
			for _, s := range cfg.Sheets.SheetsAdmin {
				if v.Properties.Title == s {
					sheetProps[s] = v.Properties
					continue
				}
			}
		}

		// Double check for props
		// roles
		if sheetProps[cfg.Sheets.SheetAdminRoles] == nil {
			log.WithField("sheet", cfg.Sheets.SheetAdminRoles).Fatal("could not find sheet")
		}
		// Admins
		for _, s := range cfg.Sheets.SheetsAdmin {
			if sheetProps[s] == nil {
				log.WithField("sheet", s).Fatal("could not find sheet")
			}
		}
		// whitelist
		for _, s := range cfg.Sheets.SheetsWhitelist {
			if sheetProps[s] == nil {
				log.WithField("sheet", s).Fatal("could not find sheet")
			}
		}

		log.Info("Retrieving data and writing admins file")
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
