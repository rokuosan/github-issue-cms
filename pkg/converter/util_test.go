package converter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_definedOr(t *testing.T) {
	t.Run("nil pointer returns default value", func(t *testing.T) {
		var nilPtr *string
		defaultVal := "default"

		result := definedOr(nilPtr, defaultVal)

		if result != defaultVal {
			t.Errorf("Expected default value %q, got %q", defaultVal, result)
		}
	})

	t.Run("non-nil pointer returns its value", func(t *testing.T) {
		value := "actual"
		ptr := &value
		defaultVal := "default"

		result := definedOr(ptr, defaultVal)

		if result != value {
			t.Errorf("Expected pointer value %q, got %q", value, result)
		}
	})

	t.Run("works with integer type", func(t *testing.T) {
		value := 42
		ptr := &value
		defaultVal := 0

		result := definedOr(ptr, defaultVal)

		if result != value {
			t.Errorf("Expected pointer value %d, got %d", value, result)
		}

		var nilPtr *int
		nilResult := definedOr(nilPtr, defaultVal)

		if nilResult != defaultVal {
			t.Errorf("Expected default value %d, got %d", defaultVal, nilResult)
		}
	})

	t.Run("works with struct type", func(t *testing.T) {
		type testStruct struct {
			Field string
		}

		t.Run("nil pointer returns default value", func(t *testing.T) {
			var nilPtr *testStruct = nil
			defaultVal := testStruct{"default"}

			result := definedOr(nilPtr, defaultVal)

			if result != defaultVal {
				t.Errorf("Expected default value %v, got %v", defaultVal, result)
			}
		})

		t.Run("non-nil pointer returns its value", func(t *testing.T) {
			value := testStruct{"actual"}
			defaultVal := testStruct{"default"}

			result := definedOr(&value, defaultVal)

			if result != value {
				t.Errorf("Expected pointer value %v, got %v", value, result)
			}
		})
	})

	t.Run("works with slice type", func(t *testing.T) {
		t.Run("nil pointer returns default value", func(t *testing.T) {
			var nilPtr []string = nil
			defaultVal := []string{"default"}

			result := definedOr(&nilPtr, defaultVal)

			assert.Equal(t, defaultVal, result)
		})

		t.Run("non-nil pointer returns its value", func(t *testing.T) {
			value := []string{"actual"}
			defaultVal := []string{"default"}

			result := definedOr(&value, defaultVal)

			assert.Equal(t, value, result)
		})
	})

	t.Run("works with map type", func(t *testing.T) {
		t.Run("nil pointer returns default value", func(t *testing.T) {
			var nilPtr map[string]int = nil
			defaultVal := map[string]int{"key": 42}

			result := definedOr(&nilPtr, defaultVal)

			assert.Equal(t, defaultVal, result)
		})

		t.Run("non-nil pointer returns its value", func(t *testing.T) {
			value := map[string]int{"key": 100}
			defaultVal := map[string]int{"key": 42}

			result := definedOr(&value, defaultVal)

			assert.Equal(t, value, result)
		})
	})

	t.Run("works with channel type", func(t *testing.T) {
		t.Run("nil pointer returns default value", func(t *testing.T) {
			var nilPtr chan int = nil
			defaultVal := make(chan int)

			result := definedOr(&nilPtr, defaultVal)

			assert.Equal(t, defaultVal, result)
		})

		t.Run("non-nil pointer returns its value", func(t *testing.T) {
			value := make(chan int)
			defaultVal := make(chan int)

			result := definedOr(&value, defaultVal)

			assert.Equal(t, value, result)
		})
	})

	t.Run("works with function type", func(t *testing.T) {
		t.Run("nil pointer returns default value", func(t *testing.T) {
			var nilPtr func() string = nil
			defaultVal := func() string { return "default" }

			result := definedOr(&nilPtr, defaultVal)

			assert.Equal(t, defaultVal(), result())
		})

		t.Run("non-nil pointer returns its value", func(t *testing.T) {
			value := func() string { return "actual" }
			defaultVal := func() string { return "default" }

			result := definedOr(&value, defaultVal)

			assert.Equal(t, value(), result())
		})
	})

	t.Run("works with interface type", func(t *testing.T) {
		t.Run("nil pointer returns default value", func(t *testing.T) {
			var nilPtr interface{} = nil
			var defaultVal interface{} = "default"

			result := definedOr(&nilPtr, defaultVal)

			assert.Equal(t, defaultVal, result)
		})

		t.Run("non-nil pointer returns its value", func(t *testing.T) {
			var value interface{} = "actual"
			var defaultVal interface{} = "default"

			result := definedOr(&value, defaultVal)

			assert.Equal(t, value, result)
		})
	})
}
