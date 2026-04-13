package config

import (
	"os"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var config Config
var GitHubToken string

// viperInitialize initializes viper.
func viperInitialize() {
	viper.SetConfigName(ConfigFileName)
	viper.SetConfigType(ConfigFileType)
	viper.AddConfigPath(ConfigFileDir)
}

// load loads a configuration file.
func load() error {
	viperInitialize()

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	if err := viper.Unmarshal(&config); err != nil {
		return err
	}

	return config.validate()
}

// Reload clears the cached config and reloads it from disk.
func Reload() (Config, error) {
	config = Config{}
	viper.Reset()
	return Get()
}

// Generate generates a configuration file.
func Generate() error {
	viperInitialize()

	c := NewConfig()

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	file, err := os.Create(GetConfigPath())
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

// Get returns a configuration.
// If the configuration is not loaded, it loads it.
func Get() (Config, error) {
	if config == (Config{}) {
		if err := load(); err != nil {
			return Config{}, err
		}
	}

	return config, nil
}
