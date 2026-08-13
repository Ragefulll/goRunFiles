//go:build !windows || !cgo

package process

import "fmt"

// StartETWNetwork is a no-op on non-windows or non-cgo builds.
func StartETWNetwork() error {
	return fmt.Errorf("etw network requires windows+cgo build")
}

// CleanupETWTotals is a no-op on non-windows.
func CleanupETWTotals() {}

func etwTotalBytes(pid int) (uint64, bool) {
	return 0, false
}

func etwSnapshotTotals() map[int]uint64 {
	return nil
}

func etwDebugSummary(limit int) string {
	return "etw:unsupported"
}
