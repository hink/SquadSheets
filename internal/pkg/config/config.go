package config

import "github.com/BurntSushi/toml"

type Config struct {
	Sheets    []*ConfigSheets  `toml:"google_sheets"`
	OtherDocs *ConfigOtherDocs `toml:"other_docs"`
}

// ConfigSheets Google sheets configuration
type ConfigSheets struct {
	SheetID         string   `toml:"sheetID"`
	SecretJSONPath  string   `toml:"secretJSONPath"`
	SheetAdminRoles string   `toml:"rolesSheet"`
	SheetsAdmin     []string `toml:"adminSheets"`
	SheetsWhitelist []string `toml:"whitelistSheets"`
}

// ConfigOtherDocs struct
type ConfigOtherDocs struct {
	Docs []string `toml:"otherDocs"`
}

// Load configuration
func Load(path string) (*Config, error) {
	config := new(Config)
	_, err := toml.DecodeFile(path, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
