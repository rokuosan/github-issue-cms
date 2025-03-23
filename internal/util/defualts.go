package util

// DefaultOrItself returns the default value if the pointer is nil, otherwise it returns the value of the pointer.
func DefaultOrItself[T any](p *T, def T) T {
	if p == nil {
		return def
	}
	return *p
}
