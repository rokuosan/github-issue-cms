package config

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func validate() {
	for _, check := range []bool{
		// Validations
		config.Hugo.IsValidBundleType(),
		config.Hugo.Direcotry.IsValidArticlesPath(),
	} {
		if !check {
			panic("Invalid configuration")
		}
	}
}

func (c *HugoConfig) IsValidBundleType() bool {
	allowType := []string{
		"none",
		"leaf",
	}

	return slices.Contains(allowType, strings.ToLower(c.Bundle))
}

func (c *HugoDirectoryConfig) IsValidArticlesPath() bool {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	abs, err := filepath.Abs(c.Articles)
	if err != nil {
		panic(err)
	}

	return strings.HasPrefix(abs, cwd)
}
