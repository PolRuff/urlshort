//go:build compilefail
// +build compilefail

package compilefail

import "github.com/PolRuff/urlshort/internal/pool"

// NoReset is a type that intentionally does not implement pool.Resettable.
// It is used to verify generic compile-time constraints.
type NoReset struct{}

func _() {
	_ = pool.New(func() *NoReset {
		return &NoReset{}
	})
}
