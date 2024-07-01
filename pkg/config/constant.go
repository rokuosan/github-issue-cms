package config

import "path/filepath"

const (
	// ConfigFileDir is the directory of the configuration file.
	ConfigFileDir = "."

	// ConfigFileName is the name of the configuration file.
	ConfigFileName = "gic.config"

	// ConfigFileType is the type of the configuration file.
	ConfigFileType = "yaml"
)

func GetConfigPath() string {
	return filepath.Join(ConfigFileDir, ConfigFileName+"."+ConfigFileType)
}
