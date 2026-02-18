package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
	last            map[string]string
	version         string
	startTimes      map[int]int64
	lastRenderLines int
	lastRenderWidth int
	restartAt       map[string]time.Time
}

func New(cfg config.Config, logger *log.Logger, version string) *App {
	if logger == nil {
		logger = log.Default()
	}
	return &App{
		cfg:        cfg,
		logger:     logger,
		last:       make(map[string]string),
		version:    version,
		startTimes: make(map[int]int64),
		restartAt:  make(map[string]time.Time),
	}
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

func (a *App) computeStatuses(doRestart bool, now time.Time) []procStatus {
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

	for _, name := range names {
		item := a.cfg.Process[name]
		status := procStatus{
			Name: name,
			Type: item.Type,
		}

		var alive bool
		switch item.Type {
		case config.TypeExe:
			if pathErr := validatePath(item.Path, item.Process); pathErr != "" {
				status.Err = pathErr
			}
			namesToCheck := parseProcessList(item.Process, item.CheckProcess)
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
		case config.TypeCmd:
			if strings.TrimSpace(item.CheckCmdline) != "" {
				ok, pid, err := process.ByNameAndCmdlineContains(item.CheckProcess, item.CheckCmdline)
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
			if checkCmdline == "" {
				checkCmdline = item.Process
			}
			if checkCmdline != "" {
				ok, pid, err := process.ByNameAndCmdlineContains(item.CheckProcess, checkCmdline)
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
			status.Target = buildBatTarget(item)
		default:
			status.Status = "❌"
			status.Err = "unknown type: " + item.Type
			statuses = append(statuses, status)
			continue
		}

		if alive {
			if a.last[name] == "*" {
				status.Status = "*"
				a.last[name] = "✔"
			} else {
				status.Status = "✔"
				a.last[name] = "✔"
			}
			status.Pid = item.Pid
			a.fillTimes(&status, now)
			delete(a.restartAt, name)
			statuses = append(statuses, status)
			continue
		}

		status.Status = "❌"
		a.last[name] = "❌"

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
			status.Status = "*"
			a.last[name] = "*"
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

type procStatus struct {
	Name      string
	Type      string
	Status    string
	Target    string
	Pid       int
	StartedAt string
	Uptime    string
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
