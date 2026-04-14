package core

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func assertEqualCmp(t *testing.T, want, got any, opts ...cmp.Option) {
	t.Helper()
	if diff := cmp.Diff(want, got, opts...); diff != "" {
		t.Fatalf("unexpected diff (-want +got):\n%s", diff)
	}
}
