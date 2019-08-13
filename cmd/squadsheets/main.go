package main

import (
	"errors"
	"os"

	"github.com/hink/SquadSheets/pkg/models"

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

		// parse sheets
		roles := make(map[string]models.AdminRole)
		admins := []models.User{}
		whitelist := []models.User{}
		for _, s := range cfg.Sheets {
			// Initialize Sheets client
			log.Info("Initializing Google Sheets client")
			sheetsSrv, err := getSheetsService(s.SecretJSONPath)
			if err != nil {
				log.Fatal(err.Error())
			}

			// Get sheet properties
			sheetProps := make(map[string]*sheets.SheetProperties)
			log.WithField("sheetid", s.SheetID).Info("getting Google sheet properties")
			props, err := getSheetProperties(sheetsSrv, s.SheetID)
			if err != nil {
				log.WithFields(log.Fields{
					"sheetid": s.SheetID,
					"error":   err.Error(),
				}).Fatal("failed to get Google sheet properties")
			}

			for _, v := range props {
				// roles
				if v.Properties.Title == s.SheetAdminRoles {
					sheetProps[s.SheetAdminRoles] = v.Properties
					continue
				}
				// whitelist
				for _, s := range s.SheetsWhitelist {
					if v.Properties.Title == s {
						sheetProps[s] = v.Properties
						continue
					}
				}
				// Admins
				for _, s := range s.SheetsAdmin {
					if v.Properties.Title == s {
						sheetProps[s] = v.Properties
						continue
					}
				}
			}

			// Double check for props
			// roles
			if sheetProps[s.SheetAdminRoles] == nil {
				log.WithField("sheet", s.SheetAdminRoles).Fatal("could not find sheet")
			}

			// Admins
			for _, sx := range s.SheetsAdmin {
				if sheetProps[sx] == nil {
					log.WithField("sheet", sx).Fatal("could not find sheet")
				}
			}
			// whitelist
			for _, sx := range s.SheetsWhitelist {
				if sheetProps[sx] == nil {
					log.WithField("sheet", sx).Fatal("could not find sheet")
				}
			}

			log.Info("Retrieving data and writing admins file")
			r, a, w, err := getFileLines(sheetsSrv, s)
			for k, v := range r {
				roles[k] = v
			}
			admins = append(admins, a...)
			whitelist = append(whitelist, w...)
		}
		return WriteAdminsFile(roles, admins, whitelist, opts.SquadConfigDir, opts.ASCWhitelist)
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
