package process

import (
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

// CPUPercent returns the CPU usage percentage for a PID.
func CPUPercent(pid int) float64 {
	if pid <= 0 {
		return 0
	}
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return 0
	}
	times, err := p.Times()
	if err != nil {
		return 0
	}
	now := time.Now()
	total := times.User + times.System
	return cpuPercentFromSample(pid, now, total)
}

// MemoryMB returns the working set (RSS) in MB for a PID.
func MemoryMB(pid int) int {
	if pid <= 0 {
		return 0
	}
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return 0
	}
	mem, err := p.MemoryInfo()
	if err != nil {
		return 0
	}
	return int(mem.RSS / (1024 * 1024))
}

type cpuSample struct {
	at    time.Time
	total float64
}

var (
	cpuMu      sync.Mutex
	cpuSamples = map[int]cpuSample{}
)

func cpuPercentFromSample(pid int, now time.Time, total float64) float64 {
	cpuMu.Lock()
	defer cpuMu.Unlock()

	prev, ok := cpuSamples[pid]
	cpuSamples[pid] = cpuSample{at: now, total: total}
	if !ok {
		return 0
	}
	dt := now.Sub(prev.at).Seconds()
	if dt <= 0 {
		return 0
	}
	dproc := total - prev.total
	if dproc <= 0 {
		return 0
	}
	cores := float64(runtime.NumCPU())
	if cores <= 0 {
		cores = 1
	}
	pct := (dproc / dt) / cores * 100.0
	if pct < 0 {
		return 0
	}
	return pct
}
