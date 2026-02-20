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

type netSample struct {
	at    time.Time
	bytes uint64
}

var (
	netMu      sync.Mutex
	netSamples = map[int]netSample{}
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

// NetKBs returns approximate network throughput in KB/s for a PID.
func NetKBs(pid int) float64 {
	if pid <= 0 {
		return 0
	}
	now := time.Now()
	var rate float64
	if useETWNetwork.Load() {
		if total, ok := etwTotalBytes(pid); ok {
			rate = netRateFromSample(pid, now, total)
		}
	}
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return applyNetScale(rate)
	}
	io, err := p.IOCounters()
	if err != nil {
		return applyNetScale(rate)
	}
	var totalBytes uint64
	totalBytes = io.ReadBytes + io.WriteBytes
	rate = netRateFromSample(pid, now, totalBytes)
	return applyNetScale(rate)
}

func netRateFromSample(pid int, now time.Time, total uint64) float64 {
	netMu.Lock()
	defer netMu.Unlock()

	prev, ok := netSamples[pid]
	netSamples[pid] = netSample{at: now, bytes: total}
	if !ok {
		return 0
	}
	dt := now.Sub(prev.at).Seconds()
	if dt <= 0 {
		return 0
	}
	delta := float64(total - prev.bytes)
	if delta <= 0 {
		return 0
	}
	return (delta / 1024.0) / dt
}

func applyNetScale(rate float64) float64 {
	scale := getNetworkScale()
	if scale <= 0 {
		scale = 1
	}
	return rate / scale
}
