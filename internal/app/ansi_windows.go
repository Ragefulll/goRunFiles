//go:build windows

package app

import "golang.org/x/sys/windows"

var ansiEnabled bool

func enableANSI() {
	if ansiEnabled {
		return
	}
	h, err := windows.GetStdHandle(windows.STD_OUTPUT_HANDLE)
	if err != nil {
		return
	}
	var mode uint32
	if err := windows.GetConsoleMode(h, &mode); err != nil {
		return
	}
	mode |= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
	if err := windows.SetConsoleMode(h, mode); err != nil {
		return
	}
	ansiEnabled = true
}
