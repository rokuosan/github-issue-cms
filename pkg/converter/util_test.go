package converter

import "testing"

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

		value := testStruct{Field: "actual"}
		ptr := &value

		result := definedOr(ptr, testStruct{Field: "default"})

		if result != value {
			t.Errorf("Expected pointer value %v, got %v", value, result)
		}
	})
}

func Test_definedOrV(t *testing.T) {
	t.Run("nil slice returns default slice", func(t *testing.T) {
		var nilSlice []string
		defaultVal := []string{"default"}

		result := definedOrV(nilSlice, defaultVal)

		if len(result) != len(defaultVal) || result[0] != defaultVal[0] {
			t.Errorf("Expected default value %v, got %v", defaultVal, result)
		}
	})

	t.Run("non-nil slice returns its value", func(t *testing.T) {
		value := []string{"actual"}
		defaultVal := []string{"default"}

		result := definedOrV(value, defaultVal)

		if len(result) != len(value) || result[0] != value[0] {
			t.Errorf("Expected value %v, got %v", value, result)
		}
	})

	t.Run("nil map returns default map", func(t *testing.T) {
		var nilMap map[string]int
		defaultVal := map[string]int{"key": 42}

		result := definedOrV(nilMap, defaultVal)

		if len(result) != len(defaultVal) || result["key"] != defaultVal["key"] {
			t.Errorf("Expected default value %v, got %v", defaultVal, result)
		}
	})

	t.Run("non-nil map returns its value", func(t *testing.T) {
		value := map[string]int{"key": 100}
		defaultVal := map[string]int{"key": 42}

		result := definedOrV(value, defaultVal)

		if len(result) != len(value) || result["key"] != value["key"] {
			t.Errorf("Expected value %v, got %v", value, result)
		}
	})

	t.Run("zero string does not returns default string", func(t *testing.T) {
		var zeroStr string
		defaultVal := "default"

		result := definedOrV(zeroStr, defaultVal)

		if result == defaultVal {
			t.Errorf("Expected default value %v, got %v", defaultVal, result)
		}
	})

	t.Run("non-zero string returns its value", func(t *testing.T) {
		value := "actual"
		defaultVal := "default"

		result := definedOrV(value, defaultVal)

		if result != value {
			t.Errorf("Expected value %v, got %v", value, result)
		}
	})
}
