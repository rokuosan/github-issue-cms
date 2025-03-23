package converter

import "reflect"

// definedOr returns the v if the p is nil, otherwise it returns the value of the p.
func definedOr[T any](p *T, v T) T {
	if p == nil {
		return v
	}
	return *p
}

// definedOrV returns the l if the r is nil, otherwise it returns the value of the r.
func definedOrV[T any](l T, r T) T {
	v := reflect.ValueOf(l)
	if (v.Kind() == reflect.Ptr || v.Kind() == reflect.Slice || v.Kind() == reflect.Map || v.Kind() == reflect.Interface) && v.IsNil() {
		return r
	}
	return l
}
