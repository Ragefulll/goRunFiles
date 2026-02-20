//go:build windows

package process

import (
	"bytes"
	"context"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// GpuSample holds per-process GPU stats.
type GpuSample struct {
	Util  float64
	MemMB int
}

// GpuStatsByPid returns GPU utilization and memory per PID (NVIDIA only).
func GpuStatsByPid() map[int]GpuSample {
	gpuMu.Lock()
	if time.Since(gpuLastAt) < time.Second && gpuLast != nil {
		defer gpuMu.Unlock()
		return cloneGpu(gpuLast)
	}
	gpuMu.Unlock()

	stats := queryGpuStats()

	gpuMu.Lock()
	gpuLastAt = time.Now()
	gpuLast = stats
	gpuMu.Unlock()
	return cloneGpu(stats)
}

var (
	gpuMu     sync.Mutex
	gpuLastAt time.Time
	gpuLast   map[int]GpuSample
)

func cloneGpu(src map[int]GpuSample) map[int]GpuSample {
	out := make(map[int]GpuSample, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func queryGpuStats() map[int]GpuSample {
	path, err := exec.LookPath("nvidia-smi")
	if err != nil {
		return map[int]GpuSample{}
	}
	stats := queryGpuComputeApps(path)
	if len(stats) > 0 {
		return stats
	}
	return queryGpuPmon(path)
}

func queryGpuComputeApps(path string) map[int]GpuSample {
	out := runHidden(path, "--query-compute-apps=pid,utilization.gpu,used_memory", "--format=csv,noheader,nounits")
	lines := strings.Split(out, "\n")
	stats := make(map[int]GpuSample, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			continue
		}
		pid, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil || pid <= 0 {
			continue
		}
		util, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		mem, _ := strconv.Atoi(strings.TrimSpace(parts[2]))
		cur := stats[pid]
		if util > cur.Util {
			cur.Util = util
		}
		cur.MemMB += mem
		stats[pid] = cur
	}
	return stats
}

func queryGpuPmon(path string) map[int]GpuSample {
	out := runHidden(path, "pmon", "-c", "1")
	lines := strings.Split(out, "\n")
	stats := make(map[int]GpuSample, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		pid, err := strconv.Atoi(fields[1])
		if err != nil || pid <= 0 {
			continue
		}
		util := parseMaybePercent(fields[3])
		mem := parseMaybePercentInt(fields[4])
		cur := stats[pid]
		if util > cur.Util {
			cur.Util = util
		}
		cur.MemMB += mem
		stats[pid] = cur
	}
	return stats
}

func runHidden(path string, args ...string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	var out bytes.Buffer
	cmd.Stdout = &out
	_ = cmd.Run()
	return out.String()
}

func parseMaybePercent(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

func parseMaybePercentInt(s string) int {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return 0
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return v
}
