//go:build windows

package app

import (
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const wmClose = 0x0010

func closeErrorWindows(titleSubstrings []string) {
	if len(titleSubstrings) == 0 {
		return
	}
	lower := make([]string, 0, len(titleSubstrings))
	for _, s := range titleSubstrings {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		lower = append(lower, strings.ToLower(s))
	}
	if len(lower) == 0 {
		return
	}

	cb := syscall.NewCallback(func(hwnd uintptr, lparam uintptr) uintptr {
		if !windows.IsWindowVisible(windows.HWND(hwnd)) {
			return 1
		}
		title := windowTitle(windows.HWND(hwnd))
		if title == "" {
			return 1
		}
		lt := strings.ToLower(title)
		for _, s := range lower {
			if strings.Contains(lt, s) {
				sendMessage(windows.HWND(hwnd), wmClose, 0, 0)
				break
			}
		}
		return 1
	})

	_, _, _ = procEnumWindows.Call(cb, 0)
}

func windowTitle(hwnd windows.HWND) string {
	n := getWindowTextLength(hwnd)
	if n == 0 {
		return ""
	}
	buf := make([]uint16, n+1)
	_ = getWindowText(hwnd, &buf[0], int32(len(buf)))
	return windows.UTF16ToString(buf)
}

var (
	user32                   = syscall.NewLazyDLL("user32.dll")
	procEnumWindows          = user32.NewProc("EnumWindows")
	procGetWindowTextLengthW = user32.NewProc("GetWindowTextLengthW")
	procGetWindowTextW       = user32.NewProc("GetWindowTextW")
	procSendMessageW         = user32.NewProc("SendMessageW")
)

func getWindowTextLength(hwnd windows.HWND) int {
	r, _, _ := procGetWindowTextLengthW.Call(uintptr(hwnd))
	return int(r)
}

func getWindowText(hwnd windows.HWND, buf *uint16, max int32) int32 {
	r, _, _ := procGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(buf)), uintptr(max))
	return int32(r)
}

func sendMessage(hwnd windows.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	r, _, _ := procSendMessageW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return r
}
