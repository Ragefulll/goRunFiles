package process

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
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
		if sameProcessName(n, name) {
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
		if sameProcessName(n, name) {
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
		if err := KillPid(pid); err != nil {
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
	if runtime.GOOS == "windows" {
		// Kill entire process tree for cmd/bat wrappers (e.g. npm/node children).
		cmd := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(pid))
		hideCmdWindow(cmd)
		return cmd.Run()
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
// containing the argument sequence (token match). Matching is performed against
// the process command line only (cwd is intentionally ignored to avoid false
// positives for unrelated processes started from the same project folder). If
// multiple matches exist, returns the newest (largest start time). If name is
// empty, matches any process by cmdline.
func ByNameAndCmdlineArgsExact(name, args string) (bool, int, error) {
	return ByNameAndCmdlineArgsExactWithExclude(name, args, "")
}

// ByNameAndCmdlineArgsExactWithExclude reports if a process with the given name has
// cmdline/cwd containing args sequence and does not match exclude sequence(s).
// Exclude accepts comma-separated patterns.
func ByNameAndCmdlineArgsExactWithExclude(name, args, exclude string) (bool, int, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	args = strings.TrimSpace(args)
	if args == "" {
		return false, 0, nil
	}
	needle := parseCmdlineTokens(args)
	if len(needle) == 0 {
		return false, 0, nil
	}
	excludeGroups := parsePatternGroups(exclude)
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
			if !sameProcessName(n, name) {
				continue
			}
		}
		tokens := processMatchTokens(p)
		if !containsTokenSequence(tokens, needle) {
			continue
		}
		if matchesAnyPatternGroup(tokens, excludeGroups) {
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
	return PidsByNameAndCmdlineArgsExactWithExclude(name, args, "")
}

// PidsByNameAndCmdlineArgsExactWithExclude returns all PIDs with cmdline/cwd containing exact args sequence.
// Exclude accepts comma-separated patterns.
func PidsByNameAndCmdlineArgsExactWithExclude(name, args, exclude string) ([]int, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	args = strings.TrimSpace(args)
	if args == "" {
		return nil, nil
	}
	needle := parseCmdlineTokens(args)
	if len(needle) == 0 {
		return nil, nil
	}
	excludeGroups := parsePatternGroups(exclude)
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
			if !sameProcessName(n, name) {
				continue
			}
		}
		tokens := processMatchTokens(p)
		if !containsTokenSequence(tokens, needle) {
			continue
		}
		if matchesAnyPatternGroup(tokens, excludeGroups) {
			continue
		}
		out = append(out, int(p.Pid))
	}
	return out, nil
}

func processMatchTokens(p *process.Process) []string {
	out := make([]string, 0, 32)
	if cmd, err := p.Cmdline(); err == nil && cmd != "" {
		out = append(out, parseCmdlineTokens(cmd)...)
	}
	// Some Windows node child processes can expose empty cmdline, but cwd still
	// points to project folder and is useful for matching.
	if cwd, err := p.Cwd(); err == nil && cwd != "" {
		out = append(out, parseCmdlineTokens(cwd)...)
	}
	return out
}

func sameProcessName(actual, expected string) bool {
	a := normalizeProcessName(actual)
	e := normalizeProcessName(expected)
	if a == "" || e == "" {
		return a == e
	}
	if a == e {
		return true
	}
	if strings.TrimSuffix(a, ".exe") == strings.TrimSuffix(e, ".exe") {
		return true
	}
	return false
}

func normalizeProcessName(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	n = strings.Trim(n, "\"'")
	return n
}

// KillByNameAndCmdlineArgsExact terminates all matching processes.
func KillByNameAndCmdlineArgsExact(name, args string) error {
	return KillByNameAndCmdlineArgsExactWithExclude(name, args, "")
}

// KillByNameAndCmdlineArgsExactWithExclude terminates matching processes and applies exclude patterns.
func KillByNameAndCmdlineArgsExactWithExclude(name, args, exclude string) error {
	pids, err := PidsByNameAndCmdlineArgsExactWithExclude(name, args, exclude)
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

func parsePatternGroups(raw string) [][]string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([][]string, 0, len(parts))
	for _, p := range parts {
		tokens := parseCmdlineTokens(strings.TrimSpace(p))
		if len(tokens) > 0 {
			out = append(out, tokens)
		}
	}
	return out
}

func matchesAnyPatternGroup(haystack []string, patterns [][]string) bool {
	for _, p := range patterns {
		if containsTokenSequence(haystack, p) {
			return true
		}
	}
	return false
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
			if tokenMatchesNeedle(haystack[h], n) {
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

func tokenMatchesNeedle(token, needle string) bool {
	t := strings.ToLower(strings.TrimSpace(token))
	n := strings.ToLower(strings.TrimSpace(needle))
	t = strings.Trim(t, "\"'")
	n = strings.Trim(n, "\"'")
	if t == "" || n == "" {
		return false
	}
	if t == n {
		return true
	}
	if strings.TrimSuffix(t, ".exe") == strings.TrimSuffix(n, ".exe") {
		return true
	}
	parts := splitTokenParts(t)
	for _, part := range parts {
		if part == n {
			return true
		}
		if strings.TrimSuffix(part, ".exe") == strings.TrimSuffix(n, ".exe") {
			return true
		}
	}
	return false
}

func splitTokenParts(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	seps := func(r rune) bool {
		switch r {
		case '\\', '/', '=', ':', ';', ',', '(', ')', '[', ']', '{', '}':
			return true
		default:
			return false
		}
	}
	parts := strings.FieldsFunc(s, seps)
	uniq := make(map[string]struct{}, len(parts)+2)
	out := make([]string, 0, len(parts)+2)
	add := func(v string) {
		v = strings.TrimSpace(strings.ToLower(v))
		if v == "" {
			return
		}
		if _, ok := uniq[v]; ok {
			return
		}
		uniq[v] = struct{}{}
		out = append(out, v)
	}
	add(s)
	add(filepath.Base(s))
	for _, p := range parts {
		add(p)
	}
	return out
}
