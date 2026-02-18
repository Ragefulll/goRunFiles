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
	hungPID                      int
	enumHungCallback             = syscall.NewCallback(enumHungProc)
)

func isProcessHung(pid int) bool {
	if pid <= 0 {
		return false
	}
	enumMu.Lock()
	hungPID = pid
	r, _, _ := procEnumWindows.Call(enumHungCallback, 0)
	hungPID = 0
	enumMu.Unlock()
	return r == 0
}

func enumHungProc(hwnd uintptr, lparam uintptr) uintptr {
	if !windows.IsWindowVisible(windows.HWND(hwnd)) {
		return 1
	}
	var wpid uint32
	_, _, _ = procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&wpid)))
	if int(wpid) != hungPID {
		return 1
	}
	r, _, _ := procIsHungAppWindow.Call(hwnd)
	if r != 0 {
		return 0
	}
	return 1
}
