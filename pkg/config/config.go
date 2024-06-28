package config

import (
	"os"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var config Config

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

	validate()
}

// Generate generates a configuration file.
func Generate() {
	viperInitialize()

	c := Config{
		GitHub: &GitHubConfig{
			Username:   "<YOUR_USERNAME>",
			Repository: "<YOUR_REPOSITORY>",
		},
		Hugo: &HugoConfig{
			Direcotry: &HugoDirectoryConfig{
				Articles: "content/posts",
			},
			Url: &HugoURLConfig{
				AppendSlash: false,
				Images:      "/images",
			},
		},
	}

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
