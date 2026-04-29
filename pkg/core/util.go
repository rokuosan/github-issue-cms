package core

// Use this file only as a last resort for helpers that are genuinely shared
// and hard to classify elsewhere.
//
// Prefer keeping functions close to the domain that owns them. Helpers added
// here should usually be small compatibility shims that can be replaced by
// standard library features when the project can adopt them. Ptr is one such
// example: it can be replaced in the future when moving to a Go 1.26-friendly
// approach, but for now it lives here as a temporary shared helper.

// Ptr returns a pointer to the given value.
func Ptr[T any](v T) *T {
	return &v
}
