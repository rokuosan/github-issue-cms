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
func load() {
	viperInitialize()

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		panic(err)
	}

	config.validate()
}

// Generate generates a configuration file.
func Generate() {
	viperInitialize()

	c := NewConfig()

	data, err := yaml.Marshal(c)
	if err != nil {
		panic(err)
	}

	file, err := os.Create(GetConfigPath())
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		panic(err)
	}
}

// Get returns a configuration.
// If the configuration is not loaded, it loads it.
func Get() Config {
	if config == (Config{}) {
		load()
	}

	return config
}
