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

// PidsByName returns all PIDs that match the given process name.
func PidsByName(name string) ([]int, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}
	out := make([]int, 0, 4)
	for _, p := range processes {
		n, err := p.Name()
		if err != nil {
			continue
		}
		if n == name {
			out = append(out, int(p.Pid))
		}
	}
	return out, nil
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

// KillByName attempts to terminate all processes matching the given name.
func KillByName(name string) error {
	pids, err := PidsByName(name)
	if err != nil {
		return err
	}
	var lastErr error
	for _, pid := range pids {
		p, err := process.NewProcess(int32(pid))
		if err != nil {
			continue
		}
		if err := p.Terminate(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// KillByNames terminates processes for all provided names.
func KillByNames(names []string) error {
	var lastErr error
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}
		if err := KillByName(n); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// KillPid attempts to terminate a process by PID.
func KillPid(pid int) error {
	if pid <= 0 {
		return nil
	}
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return err
	}
	return p.Terminate()
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
