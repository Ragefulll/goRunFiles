package process

import (
	"runtime"
	"strings"
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
	at   time.Time
	bytes uint64
	rate float64
}

var (
	netMu      sync.Mutex
	netSamples = map[int]netSample{}
)

type ioSample struct {
	at   time.Time
	bytes uint64
	rate float64
}

var (
	ioMu      sync.Mutex
	ioSamples = map[int]ioSample{}
)

const minRateSampleWindow = 250 * time.Millisecond

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
	pids := processTreePIDs(pid)
	if len(pids) == 0 {
		return 0
	}

	// Prefer ETW network counters when enabled.
	if useETWNetwork.Load() && isETWActive() {
		var etwTotal uint64
		var hasETW bool
		for _, p := range pids {
			if total, ok := etwTotalBytes(p); ok {
				etwTotal += total
				hasETW = true
			}
		}
		if hasETW {
			return applyNetScale(netRateFromSample(pid, now, etwTotal))
		}
	}

	// No ETW => no reliable per-process network counters in this mode.
	_ = netRateFromSample(pid, now, 0)
	return 0
}

func netRateFromSample(pid int, now time.Time, total uint64) float64 {
	netMu.Lock()
	defer netMu.Unlock()

	prev, ok := netSamples[pid]
	if !ok {
		netSamples[pid] = netSample{at: now, bytes: total}
		return 0
	}
	dur := now.Sub(prev.at)
	if dur < minRateSampleWindow {
		return prev.rate
	}
	dt := dur.Seconds()
	if dt <= 0 {
		return 0
	}
	delta := float64(total - prev.bytes)
	if total < prev.bytes || delta <= 0 {
		netSamples[pid] = netSample{at: now, bytes: total, rate: 0}
		return 0
	}
	rate := (delta / 1024.0) / dt
	netSamples[pid] = netSample{at: now, bytes: total, rate: rate}
	return rate
}

func applyNetScale(rate float64) float64 {
	scale := getNetworkScale()
	if scale <= 0 {
		scale = 1
	}
	return rate / scale
}

// NetKBsByNames returns aggregate network throughput in KB/s for all
// processes matching provided executable names.
func NetKBsByNames(names []string) float64 {
	pids := pidsFromNames(names)
	if len(pids) == 0 {
		return 0
	}
	now := time.Now()

	if useETWNetwork.Load() && isETWActive() {
		var totalRate float64
		for _, pid := range pids {
			if total, ok := etwTotalBytes(pid); ok {
				totalRate += netRateFromSample(pid, now, total)
			} else {
				_ = netRateFromSample(pid, now, 0)
			}
		}
		if totalRate <= 0 {
			// Fallback: match ETW totals by current process name for this sample.
			// This helps when name->pid resolution lags behind spawned child pids.
			totalRate = etwRateByNamesFallback(names, now)
		}
		return applyNetScale(totalRate)
	}

	// No ETW => no reliable per-process network counters in this mode.
	for _, pid := range pids {
		_ = netRateFromSample(pid, now, 0)
	}
	return 0
}

func etwRateByNamesFallback(names []string, now time.Time) float64 {
	totals := etwSnapshotTotals()
	if len(totals) == 0 {
		return 0
	}
	want := make([]string, 0, len(names))
	for _, n := range names {
		n = strings.TrimSpace(strings.Trim(n, "\"'"))
		if n != "" {
			want = append(want, n)
		}
	}
	if len(want) == 0 {
		return 0
	}
	var totalRate float64
	for pid, total := range totals {
		proc, err := process.NewProcess(int32(pid))
		if err != nil {
			continue
		}
		n, err := proc.Name()
		if err != nil {
			continue
		}
		for _, w := range want {
			if sameProcessName(n, w) {
				totalRate += netRateFromSample(pid, now, total)
				break
			}
		}
	}
	return totalRate
}

// IOKBs returns aggregate process I/O throughput in KB/s (parent + children).
func IOKBs(pid int) float64 {
	if pid <= 0 {
		return 0
	}
	now := time.Now()
	pids := processTreePIDs(pid)
	if len(pids) == 0 {
		return 0
	}
	var totalBytes uint64
	for _, p := range pids {
		proc, err := process.NewProcess(int32(p))
		if err != nil {
			continue
		}
		io, err := proc.IOCounters()
		if err != nil {
			continue
		}
		totalBytes += io.ReadBytes + io.WriteBytes
	}
	return applyNetScale(ioRateFromSample(pid, now, totalBytes))
}

// IOKBsByNames returns aggregate process I/O throughput in KB/s for all
// processes matching provided executable names.
func IOKBsByNames(names []string) float64 {
	pids := pidsFromNames(names)
	if len(pids) == 0 {
		return 0
	}
	now := time.Now()
	var totalRate float64
	for _, pid := range pids {
		proc, err := process.NewProcess(int32(pid))
		if err != nil {
			continue
		}
		io, err := proc.IOCounters()
		if err != nil {
			continue
		}
		total := io.ReadBytes + io.WriteBytes
		totalRate += ioRateFromSample(pid, now, total)
	}
	return applyNetScale(totalRate)
}

func ioRateFromSample(pid int, now time.Time, total uint64) float64 {
	ioMu.Lock()
	defer ioMu.Unlock()

	prev, ok := ioSamples[pid]
	if !ok {
		ioSamples[pid] = ioSample{at: now, bytes: total}
		return 0
	}
	dur := now.Sub(prev.at)
	if dur < minRateSampleWindow {
		return prev.rate
	}
	dt := dur.Seconds()
	if dt <= 0 {
		return 0
	}
	delta := float64(total - prev.bytes)
	if total < prev.bytes || delta <= 0 {
		ioSamples[pid] = ioSample{at: now, bytes: total, rate: 0}
		return 0
	}
	rate := (delta / 1024.0) / dt
	ioSamples[pid] = ioSample{at: now, bytes: total, rate: rate}
	return rate
}

func processTreePIDs(root int) []int {
	if root <= 0 {
		return nil
	}
	seen := map[int]bool{root: true}
	queue := []int{root}
	out := make([]int, 0, 8)

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		out = append(out, cur)

		p, err := process.NewProcess(int32(cur))
		if err != nil {
			continue
		}
		children, err := p.Children()
		if err != nil {
			continue
		}
		for _, ch := range children {
			cpid := int(ch.Pid)
			if cpid <= 0 || seen[cpid] {
				continue
			}
			seen[cpid] = true
			queue = append(queue, cpid)
		}
	}

	return out
}

func pidsFromNames(names []string) []int {
	seen := map[int]bool{}
	out := make([]int, 0, len(names))
	for _, n := range names {
		n = strings.TrimSpace(n)
		n = strings.Trim(n, "\"'")
		if n == "" {
			continue
		}
		pids, err := PidsByName(n)
		if err != nil {
			continue
		}
		for _, pid := range pids {
			addPIDWithFamily(pid, seen, &out)
		}
	}
	return out
}

func addPIDWithFamily(pid int, seen map[int]bool, out *[]int) {
	if pid <= 0 {
		return
	}
	// Include the process itself and its descendants.
	for _, p := range processTreePIDs(pid) {
		if p <= 0 || seen[p] {
			continue
		}
		seen[p] = true
		*out = append(*out, p)
	}
	// Also include direct parent: some launches/proxies may be attributed there.
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		return
	}
	ppid, err := proc.Ppid()
	if err != nil {
		return
	}
	parent := int(ppid)
	if parent > 0 && !seen[parent] {
		seen[parent] = true
		*out = append(*out, parent)
	}
}
