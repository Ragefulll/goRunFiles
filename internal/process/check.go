package process

import (
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

// ByName reports if a process with the given name is running and returns one PID if found.
func ByName(name string) (bool, int, error) {
	processes, err := process.Processes()
	if err != nil {
		return false, 0, err
	}
	for _, p := range processes {
		n, err := p.Name()
		if err != nil {
			continue
		}
		if n == name {
			return true, int(p.Pid), nil
		}
	}
	return false, 0, nil
}

// IsPidAlive reports if a PID is running.
func IsPidAlive(pid int) bool {
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return false
	}
	running, _ := p.IsRunning()
	return running
}

// StartTime returns the process start time for a PID.
func StartTime(pid int) (time.Time, bool) {
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return time.Time{}, false
	}
	ms, err := p.CreateTime()
	if err != nil {
		return time.Time{}, false
	}
	return time.Unix(0, ms*int64(time.Millisecond)), true
}

// ByNameAndCmdlineContains reports if a process with the given name has cmdline containing substr.
// If name is empty, matches any process by cmdline substring.
func ByNameAndCmdlineContains(name, substr string) (bool, int, error) {
	substr = strings.ToLower(strings.TrimSpace(substr))
	name = strings.ToLower(strings.TrimSpace(name))
	if substr == "" {
		return false, 0, nil
	}
	processes, err := process.Processes()
	if err != nil {
		return false, 0, err
	}
	for _, p := range processes {
		if name != "" {
			n, err := p.Name()
			if err != nil {
				continue
			}
			if strings.ToLower(n) != name {
				continue
			}
		}
		cmd, err := p.Cmdline()
		if err != nil || cmd == "" {
			continue
		}
		if strings.Contains(strings.ToLower(cmd), substr) {
			return true, int(p.Pid), nil
		}
	}
	return false, 0, nil
}
