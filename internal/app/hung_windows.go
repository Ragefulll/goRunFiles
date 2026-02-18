//go:build windows

package app

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procIsHungAppWindow          = user32.NewProc("IsHungAppWindow")
)

func isProcessHung(pid int) bool {
	if pid <= 0 {
		return false
	}
	cb := syscall.NewCallback(func(hwnd uintptr, lparam uintptr) uintptr {
		if !windows.IsWindowVisible(windows.HWND(hwnd)) {
			return 1
		}
		var wpid uint32
		_, _, _ = procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&wpid)))
		if int(wpid) != pid {
			return 1
		}
		r, _, _ := procIsHungAppWindow.Call(hwnd)
		if r != 0 {
			return 0
		}
		return 1
	})
	r, _, _ := procEnumWindows.Call(cb, 0)
	return r == 0
}
