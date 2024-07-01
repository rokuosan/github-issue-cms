package config

import (
	"slices"
	"strings"
)

func validate() {
	for _, check := range []bool{
		// Validations
		config.Hugo.IsValidBundleType(),
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
