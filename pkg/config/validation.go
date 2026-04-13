package config

import (
	"fmt"
	"log/slog"
)

func (c *Config) validate() error {
	constraints := []struct {
		errorMessage string
		fn           func() bool
	}{
		// Constraints
		{"Failed to validate deprecated options", c.WarnDeprecatedOptions},
	}

	// Check
	for _, constraint := range constraints {
		if !constraint.fn() {
			slog.Error(constraint.errorMessage)
			return fmt.Errorf("failed to validate the configuration")
		}
	}

	return nil
}

func (c *Config) WarnDeprecatedOptions() bool {
	if c.Hugo.Url.AppendSlash {
		slog.Warn("hugo.url.appendSlash is deprecated. This option will be ignored.")
	}
	if c.Hugo.Bundle != "" {
		slog.Warn("hugo.bundle is deprecated. This option will be ignored.")
	}

	return true
}
