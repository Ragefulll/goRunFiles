//go:build windows

package app

import (
	"fmt"
	"os/exec"
)

// scheduleSystemReboot triggers an immediate forced reboot of the PC.
func scheduleSystemReboot() error {
	out, err := exec.Command("shutdown", "/r", "/t", "0", "/f").CombinedOutput()
	if err != nil {
		return fmt.Errorf("shutdown: %v: %s", err, out)
	}
	return nil
}
