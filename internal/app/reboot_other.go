//go:build !windows

package app

import "fmt"

// scheduleSystemReboot is unsupported outside Windows.
func scheduleSystemReboot() error {
	return fmt.Errorf("reboot supported on Windows only")
}
