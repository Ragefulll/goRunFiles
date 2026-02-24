package app

import (
	"fmt"
	"strings"
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
	NetKBs    string `json:"net_kbs"`
	IOKBs     string `json:"io_kbs"`
}

// DisplaySnapshot is a UI-friendly snapshot of the current system state.
type DisplaySnapshot struct {
	Updated string          `json:"updated"`
	Version string          `json:"version"`
	NetUnit string          `json:"net_unit"`
	NetMode string          `json:"net_mode"`
	NetErr  string          `json:"net_err"`
	NetDbg  string          `json:"net_dbg"`
	Items   []DisplayStatus `json:"items"`
}

func buildDisplaySnapshot(version string, statuses []procStatus, now time.Time, netUnit, netMode, netErr, netDbg string) DisplaySnapshot {
	if strings.TrimSpace(netUnit) == "" {
		netUnit = "KB"
	}
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
			NetKBs:    formatRate(s.NetKBs, netUnit),
			IOKBs:     formatRate(s.IOKBs, netUnit),
		})
	}
	return DisplaySnapshot{
		Updated: now.Format("2006-01-02 15:04:05"),
		Version: version,
		NetUnit: netUnit,
		NetMode: netMode,
		NetErr:  netErr,
		NetDbg:  netDbg,
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

func formatRate(v float64, unit string) string {
	if v <= 0 {
		return "0"
	}
	if strings.EqualFold(unit, "MB") || strings.EqualFold(unit, "MB/s") {
		v = v / 1024.0
		if v < 10 {
			return fmt.Sprintf("%.1f", v)
		}
	}
	return fmt.Sprintf("%.0f", v)
}
