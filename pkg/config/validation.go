package config

import (
	"log/slog"
	"slices"
	"strings"
)

func (c *Config) validate() {
	constraints := []struct {
		errorMessage string
		fn           func() bool
	}{
		// Constraints
		{"Invalid bandle type: " + c.Hugo.Bundle, c.Hugo.IsValidBundleType},
		{"Failed to validate deprecated options", c.WarnDeprecatedOptions},
	}

	// Check
	for _, constraint := range constraints {
		if !constraint.fn() {
			slog.Error(constraint.errorMessage)
			panic("Failed to validate the configuration")
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

func (c *Config) WarnDeprecatedOptions() bool {
	if c.Hugo.Url.AppendSlash {
		slog.Warn("hugo.url.appendSlash is deprecated. This option will be ignored.")
	}

	return true
}
