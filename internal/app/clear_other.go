//go:build !windows

package app

import "os"

func clearConsole() {
	_, _ = os.Stdout.WriteString("\n\n\n\n\n")
}
