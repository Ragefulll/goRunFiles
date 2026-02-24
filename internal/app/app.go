package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"goRunFiles/internal/config"
	"goRunFiles/internal/process"
	"goRunFiles/internal/runner"

	"github.com/mattn/go-runewidth"
)

const LogTag = "[ART3D-CHEKER]:"

type App struct {
	cfg             config.Config
	logger          *log.Logger
	last            map[string]Status
	version         string
	defaultDisabled map[string]bool
	startTimes      map[int]int64
	lastRenderLines int
	lastRenderWidth int
	restartAt       map[string]time.Time
	hungSince       map[string]time.Time
	manualStop      map[string]bool
	mu              sync.Mutex
}

func New(cfg config.Config, logger *log.Logger, version string) *App {
	if logger == nil {
		logger = log.Default()
	}
	app := &App{
		cfg:             cfg,
		logger:          logger,
		last:            make(map[string]Status),
		version:         version,
		defaultDisabled: buildDefaultDisabledMap(cfg),
		startTimes:      make(map[int]int64),
		restartAt:       make(map[string]time.Time),
		hungSince:       make(map[string]time.Time),
		manualStop:      make(map[string]bool),
	}
	process.SetNetworkConfig(cfg.Settings.UseETWNetwork)
	process.SetNetworkScale(cfg.Settings.NetScale)
	return app
}

func (a *App) Run(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}
	if a.cfg.Settings.CheckTiming.Duration <= 0 {
		return fmt.Errorf("invalid check timing: %v\n\n", a.cfg.Settings.CheckTiming)
	}
	if a.cfg.Settings.RestartTiming.Duration <= 0 {
		return fmt.Errorf("invalid restart timing: %v\n\n", a.cfg.Settings.RestartTiming)
	}

	hideCursor()

	statuses := a.computeStatuses(true, time.Now())
	a.render(statuses)

	checkTicker := time.NewTicker(a.cfg.Settings.CheckTiming.Duration)
	defer checkTicker.Stop()

	restartTicker := time.NewTicker(a.cfg.Settings.RestartTiming.Duration)
	defer restartTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-checkTicker.C:
			statuses := a.computeStatuses(false, time.Now())
			a.render(statuses)
		case <-restartTicker.C:
			a.computeStatuses(true, time.Now())
		}
	}
}

// RunWithObserver runs the monitor loop and reports snapshots via callback.
func (a *App) RunWithObserver(ctx context.Context, onUpdate func(DisplaySnapshot)) error {
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}
	if a.cfg.Settings.CheckTiming.Duration <= 0 {
		return fmt.Errorf("invalid check timing: %v\n\n", a.cfg.Settings.CheckTiming)
	}
	if a.cfg.Settings.RestartTiming.Duration <= 0 {
		return fmt.Errorf("invalid restart timing: %v\n\n", a.cfg.Settings.RestartTiming)
	}
	if onUpdate == nil {
		return fmt.Errorf("onUpdate is nil")
	}

	now := time.Now()
	statuses := a.computeStatuses(true, now)
	onUpdate(buildDisplaySnapshot(a.version, statuses, now, a.cfg.Settings.NetUnit))

	checkTicker := time.NewTicker(a.cfg.Settings.CheckTiming.Duration)
	defer checkTicker.Stop()

	restartTicker := time.NewTicker(a.cfg.Settings.RestartTiming.Duration)
	defer restartTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-checkTicker.C:
			now := time.Now()
			statuses := a.computeStatuses(false, now)
			onUpdate(buildDisplaySnapshot(a.version, statuses, now, a.cfg.Settings.NetUnit))
		case <-restartTicker.C:
			now := time.Now()
			statuses := a.computeStatuses(true, now)
			onUpdate(buildDisplaySnapshot(a.version, statuses, now, a.cfg.Settings.NetUnit))
		}
	}
}

func (a *App) computeStatuses(doRestart bool, now time.Time) []procStatus {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cfg.Settings.AutoCloseErrorDialogs {
		titles := parseCSV(a.cfg.Settings.ErrorWindowTitles)
		titles = append(titles, a.buildAutoErrorTitles()...)
		closeErrorWindows(titles)
	}

	statuses := make([]procStatus, 0, len(a.cfg.Process))

	names := make([]string, 0, len(a.cfg.Process))
	for name := range a.cfg.Process {
		names = append(names, name)
	}
	sort.Strings(names)

	gpuByPid := process.GpuStatsByPid()

	for _, name := range names {
		item := a.cfg.Process[name]
		status := procStatus{
			Name: name,
			Type: item.Type,
		}
		var namesToCheck []string
		if item.Disabled {
			status.Status = StatusDisabled
			status.Err = ""
			statuses = append(statuses, status)
			continue
		}

		var alive bool
		switch item.Type {
		case config.TypeExe:
			if pathErr := validatePath(item.Path, item.Process); pathErr != "" {
				status.Err = pathErr
			}
			namesToCheck = parseProcessList(item.Process, item.CheckProcess)
			for _, procName := range namesToCheck {
				ok, pid, err := process.ByName(procName)
				if err != nil {
					status.Err = err.Error()
				}
				if ok {
					alive = true
					if pid > 0 {
						item.Pid = pid
					}
					break
				}
			}
			status.Target = buildExeTarget(item, namesToCheck)
			if alive && item.MonitorHang && item.HangTimeout.Duration > 0 {
				hung := isAnyProcessHung(namesToCheck)
				status.Hung = hung
				if hung {
					if _, ok := a.hungSince[name]; !ok {
						a.hungSince[name] = now
					}
					if now.Sub(a.hungSince[name]) >= item.HangTimeout.Duration {
						_ = process.KillByNames(namesToCheck)
						alive = false
						status.Err = "Not responding"
						a.restartAt[name] = now
						delete(a.hungSince, name)
					}
				} else {
					delete(a.hungSince, name)
				}
			}
		case config.TypeCmd:
			if strings.TrimSpace(item.CheckCmdline) != "" {
				ok, pid, err := process.ByNameAndCmdlineArgsExact(item.CheckProcess, item.CheckCmdline)
				if err != nil {
					status.Err = err.Error()
				}
				alive = ok
				if pid > 0 {
					item.Pid = pid
				}
			} else if strings.TrimSpace(item.CheckProcess) != "" {
				namesToCheck := parseProcessList("", item.CheckProcess)
				for _, procName := range namesToCheck {
					ok, pid, err := process.ByName(procName)
					if err != nil {
						status.Err = err.Error()
					}
					if ok {
						alive = true
						if pid > 0 {
							item.Pid = pid
						}
						break
					}
				}
			} else if item.Pid > 0 {
				alive = process.IsPidAlive(item.Pid)
			}
			status.Target = item.Command
		case config.TypeBat:
			if pathErr := validatePath(item.Path, item.Process); pathErr != "" {
				status.Err = pathErr
			}
			checkCmdline := strings.TrimSpace(item.CheckCmdline)
			if checkCmdline != "" {
				ok, pid, err := process.ByNameAndCmdlineArgsExact(item.CheckProcess, checkCmdline)
				if err != nil {
					status.Err = err.Error()
				}
				alive = ok
				if pid > 0 {
					item.Pid = pid
				}
			} else if strings.TrimSpace(item.CheckProcess) != "" {
				namesToCheck := parseProcessList("", item.CheckProcess)
				for _, procName := range namesToCheck {
					ok, pid, err := process.ByName(procName)
					if err != nil {
						status.Err = err.Error()
					}
					if ok {
						alive = true
						if pid > 0 {
							item.Pid = pid
						}
						break
					}
				}
			} else if strings.TrimSpace(item.Process) != "" {
				ok, pid, err := process.ByNameAndCmdlineArgsExact("", item.Process)
				if err != nil {
					status.Err = err.Error()
				}
				alive = ok
				if pid > 0 {
					item.Pid = pid
				}
			} else if item.Pid > 0 {
				alive = process.IsPidAlive(item.Pid)
			}
			status.Target = buildBatTarget(item)
		default:
			status.Status = StatusStopped
			status.Err = "unknown type: " + item.Type
			statuses = append(statuses, status)
			continue
		}

		if alive {
			if a.manualStop[name] {
				delete(a.manualStop, name)
			}
			if a.last[name] == StatusStarted {
				status.Status = StatusStarted
				a.last[name] = StatusRunning
			} else {
				status.Status = StatusRunning
				a.last[name] = StatusRunning
			}
			status.Pid = item.Pid
			metricsPid := status.Pid
			if status.Type == config.TypeExe {
				metricsPid = preferShippingPid(namesToCheck, status.Pid)
			}
			if metricsPid > 0 {
				status.Cpu = process.CPUPercent(metricsPid)
				status.MemMB = process.MemoryMB(metricsPid)
				status.NetKBs = process.NetKBs(metricsPid)
				if gpu, ok := gpuByPid[metricsPid]; ok {
					status.Gpu = gpu.Util
					status.GpuMemMB = gpu.MemMB
				}
			}
			a.fillTimes(&status, now)
			delete(a.restartAt, name)
			delete(a.hungSince, name)
			statuses = append(statuses, status)
			continue
		}

		status.Status = StatusStopped
		a.last[name] = StatusStopped

		if a.manualStop[name] {
			status.Uptime = "-"
			status.StartedAt = "-"
			delete(a.restartAt, name)
			statuses = append(statuses, status)
			continue
		}

		if _, ok := a.restartAt[name]; !ok {
			a.restartAt[name] = now.Add(a.cfg.Settings.RestartTiming.Duration)
		}

		if doRestart && !a.restartAt[name].After(now) {
			pid, err := runner.Start(item, a.cfg.Settings.LaunchInNewConsole)
			if err != nil {
				status.Err = err.Error()
				status.Uptime = formatCountdown(a.restartAt[name].Sub(now))
				statuses = append(statuses, status)
				continue
			}
			item.Pid = pid
			status.Pid = pid
			status.Status = StatusStarted
			a.last[name] = StatusStarted
			if pid > 0 {
				a.startTimes[pid] = time.Now().UnixMilli()
				a.fillTimes(&status, now)
			}
			delete(a.restartAt, name)
		}

		if status.Uptime == "" {
			status.Uptime = formatCountdown(a.restartAt[name].Sub(now))
			status.StartedAt = "-"
		}

		statuses = append(statuses, status)
	}

	return statuses
}

// UpdateConfig replaces the current config with a new one.
func (a *App) UpdateConfig(cfg config.Config) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg = cfg
	a.last = make(map[string]Status)
	a.defaultDisabled = buildDefaultDisabledMap(cfg)
	a.startTimes = make(map[int]int64)
	a.restartAt = make(map[string]time.Time)
	a.hungSince = make(map[string]time.Time)
	a.manualStop = make(map[string]bool)
	process.SetNetworkConfig(cfg.Settings.UseETWNetwork)
	process.SetNetworkScale(cfg.Settings.NetScale)
}

// StartProcess starts a process by config name.
func (a *App) StartProcess(name string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	item, ok := a.cfg.Process[name]
	if !ok {
		return fmt.Errorf("process %q not found", name)
	}
	// Manual START enables the process so it enters regular monitoring.
	item.Disabled = false
	pid, err := runner.Start(item, a.cfg.Settings.LaunchInNewConsole)
	if err != nil {
		return err
	}
	item.Pid = pid
	a.last[name] = StatusStarted
	delete(a.manualStop, name)
	return nil
}

// StopProcess stops a process by config name.
func (a *App) StopProcess(name string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	item, ok := a.cfg.Process[name]
	if !ok {
		return fmt.Errorf("process %q not found", name)
	}
	if a.defaultDisabled[name] {
		item.Disabled = true
	}
	a.manualStop[name] = true
	delete(a.restartAt, name)
	return stopProcessItem(item)
}

// RestartProcess restarts a process by config name.
func (a *App) RestartProcess(name string) error {
	if err := a.StopProcess(name); err != nil {
		return err
	}
	delete(a.manualStop, name)
	return a.StartProcess(name)
}

// RestartAll restarts all enabled processes.
func (a *App) RestartAll() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	var lastErr error
	a.manualStop = make(map[string]bool)
	// stop enabled
	for _, item := range a.cfg.Process {
		if item.Disabled {
			continue
		}
		if err := stopProcessItem(item); err != nil {
			lastErr = err
		}
	}
	// start enabled
	for name, item := range a.cfg.Process {
		if item.Disabled {
			continue
		}
		pid, err := runner.Start(item, a.cfg.Settings.LaunchInNewConsole)
		if err != nil {
			lastErr = err
			continue
		}
		item.Pid = pid
		a.last[name] = StatusStarted
		if pid > 0 {
			a.startTimes[pid] = time.Now().UnixMilli()
		}
	}
	return lastErr
}

// GetProcessPath returns configured working path for process by config name.
func (a *App) GetProcessPath(name string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	item, ok := a.cfg.Process[name]
	if !ok {
		return "", fmt.Errorf("process %q not found", name)
	}
	path := strings.TrimSpace(item.Path)
	if path == "" {
		return "", fmt.Errorf("process %q has empty path", name)
	}
	return path, nil
}

// StopAll stops all configured processes (including disabled).
func (a *App) StopAll() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	var lastErr error
	for _, item := range a.cfg.Process {
		if err := stopProcessItem(item); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

type procStatus struct {
	Name      string
	Type      string
	Status    Status
	Target    string
	Pid       int
	StartedAt string
	Uptime    string
	Hung      bool
	Cpu       float64
	Gpu       float64
	GpuMemMB  int
	MemMB     int
	NetKBs    float64
	Err       string
}

func (s procStatus) pidString() string {
	if s.Pid <= 0 {
		return "-"
	}
	return fmt.Sprintf("%d", s.Pid)
}

func truncateDisplay(s string, max int) string {
	if max <= 0 {
		return ""
	}
	return runewidth.Truncate(s, max, "...")
}

func padRight(s string, width int) string {
	if runewidth.StringWidth(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-runewidth.StringWidth(s))
}

func parseProcessList(defaultProcess, checkProcess string) []string {
	if strings.TrimSpace(checkProcess) == "" {
		return []string{defaultProcess}
	}
	parts := strings.Split(checkProcess, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	if len(out) == 0 {
		return []string{defaultProcess}
	}
	return out
}

func parseCSV(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

func buildDefaultDisabledMap(cfg config.Config) map[string]bool {
	out := make(map[string]bool, len(cfg.Process))
	for name, item := range cfg.Process {
		if item == nil {
			continue
		}
		out[name] = item.Disabled
	}
	return out
}

func preferShippingPid(namesToCheck []string, fallback int) int {
	for _, name := range namesToCheck {
		if strings.Contains(strings.ToLower(name), "win64-shipping.exe") {
			ok, pid, _ := process.ByName(name)
			if ok && pid > 0 {
				return pid
			}
		}
	}
	return fallback
}

func (a *App) buildAutoErrorTitles() []string {
	out := make([]string, 0, len(a.cfg.Process))
	for _, item := range a.cfg.Process {
		if item.Type != config.TypeExe || item.Process == "" {
			continue
		}
		name := strings.TrimSpace(item.Process)
		if strings.HasSuffix(strings.ToLower(name), ".exe") {
			name = name[:len(name)-4]
		}
		if name == "" {
			continue
		}
		out = append(out, "The UE-"+name+" Game has crashed and will close")
	}
	return out
}

func (a *App) fillTimes(status *procStatus, now time.Time) {
	if status.Pid <= 0 {
		status.StartedAt = "-"
		status.Uptime = "-"
		return
	}
	start, ok := a.getStartTime(status.Pid)
	if !ok {
		status.StartedAt = "-"
		status.Uptime = "-"
		return
	}
	status.StartedAt = start.Format("2006-01-02 15:04:05")
	status.Uptime = formatUptime(now.Sub(start))
}

func (a *App) getStartTime(pid int) (time.Time, bool) {
	if pid <= 0 {
		return time.Time{}, false
	}
	if ms, ok := a.startTimes[pid]; ok {
		return time.Unix(0, ms*int64(time.Millisecond)), true
	}
	t, ok := process.StartTime(pid)
	if !ok {
		return time.Time{}, false
	}
	a.startTimes[pid] = t.UnixMilli()
	return t, true
}

func formatUptime(d time.Duration) string {
	if d < 0 {
		return "-"
	}
	d = d.Truncate(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

func formatCountdown(d time.Duration) string {
	if d < 0 {
		return "restart now"
	}
	d = d.Truncate(time.Second)
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	if m > 99 {
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("restart in %02d:%02d", m, s)
}

func buildBatTarget(item *config.ProcessItem) string {
	if item.Args == "" {
		return item.Process
	}
	return item.Process
}

func buildExeTarget(item *config.ProcessItem, namesToCheck []string) string {
	if strings.TrimSpace(item.Process) == "" {
		return strings.Join(namesToCheck, ", ")
	}
	return item.Process
}

func isAnyProcessHung(names []string) bool {
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}
		pids, err := process.PidsByName(n)
		if err != nil {
			continue
		}
		for _, pid := range pids {
			if isProcessHung(pid) {
				return true
			}
		}
	}
	return false
}

func validatePath(dir, name string) string {
	if strings.TrimSpace(dir) == "" || strings.TrimSpace(name) == "" {
		return ""
	}
	full := filepath.Join(dir, name)
	if _, err := os.Stat(full); err != nil {
		if os.IsNotExist(err) {
			return "File not found"
		}
		return "path error: " + err.Error()
	}
	return ""
}

func stopProcessItem(item *config.ProcessItem) error {
	switch item.Type {
	case config.TypeExe:
		names := parseProcessList(item.Process, item.CheckProcess)
		if err := process.KillByNames(names); err != nil {
			return err
		}
		item.Pid = 0
		return nil
	case config.TypeCmd, config.TypeBat:
		if strings.TrimSpace(item.CheckCmdline) != "" {
			if err := process.KillByNameAndCmdlineArgsExact(item.CheckProcess, item.CheckCmdline); err != nil {
				return err
			}
			item.Pid = 0
			return nil
		}
		if strings.TrimSpace(item.CheckProcess) != "" {
			names := parseProcessList("", item.CheckProcess)
			if err := process.KillByNames(names); err != nil {
				return err
			}
			item.Pid = 0
			return nil
		}
		if item.Pid > 0 {
			if err := process.KillPid(item.Pid); err != nil {
				return err
			}
			item.Pid = 0
		}
		return nil
	default:
		return fmt.Errorf("unknown type %q", item.Type)
	}
}
