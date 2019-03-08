package config

import "github.com/BurntSushi/toml"

type Config struct {
	Sheets *ConfigSheets `toml:"google_sheets"`
}

// ConfigSheets Google sheets configuration
type ConfigSheets struct {
	SheetID        string `toml:"sheetID"`
	SecretJSONPath string `toml:"secretJSONPath"`
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
