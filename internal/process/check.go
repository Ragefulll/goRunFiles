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

// ByNameAndCmdlineArgsExact reports if a process with the given name has cmdline
// containing the exact argument sequence (token match). If multiple matches
// exist, returns the newest (largest start time). If name is empty, matches any
// process by cmdline.
func ByNameAndCmdlineArgsExact(name, args string) (bool, int, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	args = strings.TrimSpace(args)
	if args == "" {
		return false, 0, nil
	}
	needle := parseCmdlineTokens(args)
	if len(needle) == 0 {
		return false, 0, nil
	}
	processes, err := process.Processes()
	if err != nil {
		return false, 0, err
	}
	var bestPid int
	var bestStart int64
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
		tokens := parseCmdlineTokens(cmd)
		if !containsTokenSequence(tokens, needle) {
			continue
		}
		start, err := p.CreateTime()
		if err == nil && start >= bestStart {
			bestStart = start
			bestPid = int(p.Pid)
		} else if bestPid == 0 {
			bestPid = int(p.Pid)
		}
	}
	if bestPid > 0 {
		return true, bestPid, nil
	}
	return false, 0, nil
}

// PidsByNameAndCmdlineArgsExact returns all PIDs with cmdline containing exact args sequence.
func PidsByNameAndCmdlineArgsExact(name, args string) ([]int, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	args = strings.TrimSpace(args)
	if args == "" {
		return nil, nil
	}
	needle := parseCmdlineTokens(args)
	if len(needle) == 0 {
		return nil, nil
	}
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}
	out := make([]int, 0, 4)
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
		tokens := parseCmdlineTokens(cmd)
		if !containsTokenSequence(tokens, needle) {
			continue
		}
		out = append(out, int(p.Pid))
	}
	return out, nil
}

// KillByNameAndCmdlineArgsExact terminates all matching processes.
func KillByNameAndCmdlineArgsExact(name, args string) error {
	pids, err := PidsByNameAndCmdlineArgsExact(name, args)
	if err != nil {
		return err
	}
	var lastErr error
	for _, pid := range pids {
		if err := KillPid(pid); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

func parseCmdlineTokens(s string) []string {
	var out []string
	var b strings.Builder
	inQuote := byte(0)
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '"' || ch == '\'' {
			if inQuote == 0 {
				inQuote = ch
				continue
			}
			if inQuote == ch {
				inQuote = 0
				continue
			}
		}
		if inQuote == 0 && (ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r') {
			if b.Len() > 0 {
				out = append(out, strings.ToLower(b.String()))
				b.Reset()
			}
			continue
		}
		b.WriteByte(ch)
	}
	if b.Len() > 0 {
		out = append(out, strings.ToLower(b.String()))
	}
	return out
}

func containsTokenSequence(haystack, needle []string) bool {
	if len(needle) == 0 || len(haystack) == 0 {
		return false
	}
	h := 0
	for _, n := range needle {
		found := false
		for h < len(haystack) {
			if strings.Contains(haystack[h], n) {
				found = true
				break
			}
			h++
		}
		if !found {
			return false
		}
	}
	return true
}
