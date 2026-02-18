//go:build windows

package app

import (
	"os"

	"golang.org/x/sys/windows"
)

func clearConsole() {
	if ansiEnabled {
		_, _ = os.Stdout.WriteString("\x1b[H")
		return
	}
	cursorHome()
}

func cursorHome() {
	h, err := windows.GetStdHandle(windows.STD_OUTPUT_HANDLE)
	if err != nil {
		return
	}
	pos := windows.Coord{X: 0, Y: 0}
	_ = windows.SetConsoleCursorPosition(h, pos)
}
