package util

import (
	"testing"
)

func TestDefaultOrItself(t *testing.T) {
	t.Run("nil pointer returns default value", func(t *testing.T) {
		var nilPtr *string
		defaultVal := "default"

		result := DefaultOrItself(nilPtr, defaultVal)

		if result != defaultVal {
			t.Errorf("Expected default value %q, got %q", defaultVal, result)
		}
	})

	t.Run("non-nil pointer returns its value", func(t *testing.T) {
		value := "actual"
		ptr := &value
		defaultVal := "default"

		result := DefaultOrItself(ptr, defaultVal)

		if result != value {
			t.Errorf("Expected pointer value %q, got %q", value, result)
		}
	})

	t.Run("works with integer type", func(t *testing.T) {
		value := 42
		ptr := &value
		defaultVal := 0

		result := DefaultOrItself(ptr, defaultVal)

		if result != value {
			t.Errorf("Expected pointer value %d, got %d", value, result)
		}

		var nilPtr *int
		nilResult := DefaultOrItself(nilPtr, defaultVal)

		if nilResult != defaultVal {
			t.Errorf("Expected default value %d, got %d", defaultVal, nilResult)
		}
	})
}
