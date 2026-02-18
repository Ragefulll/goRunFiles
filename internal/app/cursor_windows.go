//go:build windows

package app

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type consoleCursorInfo struct {
	Size    uint32
	Visible int32
}

var (
	kernel32Cursor             = syscall.NewLazyDLL("kernel32.dll")
	procSetConsoleCursorInfo   = kernel32Cursor.NewProc("SetConsoleCursorInfo")
)

func hideCursor() {
	h, err := windows.GetStdHandle(windows.STD_OUTPUT_HANDLE)
	if err != nil {
		return
	}
	info := consoleCursorInfo{
		Size:    1,
		Visible: 0,
	}
	_, _, _ = procSetConsoleCursorInfo.Call(uintptr(h), uintptr(unsafe.Pointer(&info)))
}
