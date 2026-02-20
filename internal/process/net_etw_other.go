//go:build !windows || !cgo

package process

// StartETWNetwork is a no-op on non-windows or non-cgo builds.
func StartETWNetwork() error {
	return nil
}

func etwTotalBytes(pid int) (uint64, bool) {
	return 0, false
}
