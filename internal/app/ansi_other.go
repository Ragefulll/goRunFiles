//go:build !windows

package app

var ansiEnabled bool

func enableANSI() {
	// no-op on non-windows
}
