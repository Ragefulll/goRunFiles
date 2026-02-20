package app

import (
	"fmt"
	"time"
)

// DisplayStatus is a stable, UI-friendly view of procStatus.
type DisplayStatus struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	Icon      string `json:"icon"`
	Pid       string `json:"pid"`
	StartedAt string `json:"started_at"`
	Uptime    string `json:"uptime"`
	Target    string `json:"target"`
	Error     string `json:"error"`
	Hung      bool   `json:"hung"`
	Cpu       string `json:"cpu"`
	Gpu       string `json:"gpu"`
	GpuMemMB  string `json:"gpu_mem_mb"`
	MemMB     string `json:"mem_mb"`
}

// DisplaySnapshot is a UI-friendly snapshot of the current system state.
type DisplaySnapshot struct {
	Updated string          `json:"updated"`
	Version string          `json:"version"`
	Items   []DisplayStatus `json:"items"`
}

func buildDisplaySnapshot(version string, statuses []procStatus, now time.Time) DisplaySnapshot {
	items := make([]DisplayStatus, 0, len(statuses))
	for _, s := range statuses {
		items = append(items, DisplayStatus{
			Name:      s.Name,
			Type:      s.Type,
			Status:    string(s.Status),
			Icon:      s.Status.Icon(),
			Pid:       s.pidString(),
			StartedAt: s.StartedAt,
			Uptime:    s.Uptime,
			Target:    s.Target,
			Error:     s.Err,
			Hung:      s.Hung,
			Cpu:       formatPercent(s.Cpu),
			Gpu:       formatPercent(s.Gpu),
			GpuMemMB:  formatMemMB(s.GpuMemMB),
			MemMB:     formatMemMB(s.MemMB),
		})
	}
	return DisplaySnapshot{
		Updated: now.Format("2006-01-02 15:04:05"),
		Version: version,
		Items:   items,
	}
}

func formatPercent(v float64) string {
	if v <= 0 {
		return "0"
	}
	if v >= 999 {
		return "999"
	}
	return fmt.Sprintf("%.0f", v)
}

func formatMemMB(v int) string {
	if v <= 0 {
		return "0"
	}
	return fmt.Sprintf("%d", v)
}
