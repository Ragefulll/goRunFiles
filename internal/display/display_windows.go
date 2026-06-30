//go:build windows

package display

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"github.com/shirou/gopsutil/v3/process"
)

type rect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type monitorInfoEx struct {
	Size    uint32
	Monitor rect
	Work    rect
	Flags   uint32
	Device  [32]uint16
}

var (
	user32                    = syscall.NewLazyDLL("user32.dll")
	procEnumDisplayMonitors   = user32.NewProc("EnumDisplayMonitors")
	procGetMonitorInfoW       = user32.NewProc("GetMonitorInfoW")
	procEnumWindows           = user32.NewProc("EnumWindows")
	procGetWindowThreadProcID = user32.NewProc("GetWindowThreadProcessId")
	procIsWindowVisible       = user32.NewProc("IsWindowVisible")
	procGetWindow             = user32.NewProc("GetWindow")
	procSetWindowPos          = user32.NewProc("SetWindowPos")
)

const (
	monitorinfofPrimary  = 0x00000001
	gwOwner              = 4
	swpNoZOrder          = 0x0004
	swpNoActivate        = 0x0010
	swpShowWindow        = 0x0040
	swpNoSendChanging    = 0x0400
)

func ListScreens() ([]Screen, error) {
	var screens []Screen
	cb := syscall.NewCallback(func(hMonitor uintptr, hdc uintptr, lprcMonitor uintptr, lparam uintptr) uintptr {
		var info monitorInfoEx
		info.Size = uint32(unsafe.Sizeof(info))
		ret, _, _ := procGetMonitorInfoW.Call(hMonitor, uintptr(unsafe.Pointer(&info)))
		if ret == 0 {
			return 1
		}
		r := info.Monitor
		index := len(screens) + 1
		name := syscall.UTF16ToString(info.Device[:])
		if name == "" {
			name = fmt.Sprintf("Screen %d", index)
		}
		screens = append(screens, Screen{
			Index:   index,
			Name:    name,
			Primary: info.Flags&monitorinfofPrimary != 0,
			X:       int(r.Left),
			Y:       int(r.Top),
			Width:   int(r.Right - r.Left),
			Height:  int(r.Bottom - r.Top),
		})
		return 1
	})
	ret, _, err := procEnumDisplayMonitors.Call(0, 0, cb, 0)
	if ret == 0 {
		return nil, err
	}
	return screens, nil
}

func MoveProcessWindowToScreen(pid int, screenIndex int) error {
	if pid <= 0 || screenIndex <= 0 {
		return nil
	}
	screens, err := ListScreens()
	if err != nil {
		return err
	}
	if screenIndex > len(screens) {
		return fmt.Errorf("screen %d not found", screenIndex)
	}
	screen := screens[screenIndex-1]

	var hwnd uintptr
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		hwnd = findMainWindow(processTreePIDs(pid))
		if hwnd != 0 {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if hwnd == 0 {
		return fmt.Errorf("window for pid %d not found", pid)
	}

	ret, _, err := procSetWindowPos.Call(
		hwnd,
		0,
		uintptr(int32(screen.X)),
		uintptr(int32(screen.Y)),
		uintptr(int32(screen.Width)),
		uintptr(int32(screen.Height)),
		uintptr(swpNoZOrder|swpShowWindow|swpNoActivate),
	)
	if ret == 0 {
		return err
	}
	return nil
}

func findMainWindow(targetPIDs []int) uintptr {
	targets := make(map[int]struct{}, len(targetPIDs))
	for _, pid := range targetPIDs {
		if pid > 0 {
			targets[pid] = struct{}{}
		}
	}
	var found uintptr
	cb := syscall.NewCallback(func(hwnd uintptr, lparam uintptr) uintptr {
		if found != 0 {
			return 0
		}
		visible, _, _ := procIsWindowVisible.Call(hwnd)
		if visible == 0 {
			return 1
		}
		owner, _, _ := procGetWindow.Call(hwnd, gwOwner)
		if owner != 0 {
			return 1
		}
		var pid uint32
		procGetWindowThreadProcID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
		if _, ok := targets[int(pid)]; ok {
			found = hwnd
			return 0
		}
		return 1
	})
	procEnumWindows.Call(cb, 0)
	return found
}

func processTreePIDs(root int) []int {
	if root <= 0 {
		return nil
	}
	seen := map[int]bool{root: true}
	out := []int{root}
	queue := []int{root}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		p, err := process.NewProcess(int32(cur))
		if err != nil {
			continue
		}
		children, err := p.Children()
		if err != nil {
			continue
		}
		for _, child := range children {
			pid := int(child.Pid)
			if pid <= 0 || seen[pid] {
				continue
			}
			seen[pid] = true
			out = append(out, pid)
			queue = append(queue, pid)
		}
	}
	return out
}
