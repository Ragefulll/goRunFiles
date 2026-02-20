//go:build !windows

package process

// GpuSample holds per-process GPU stats.
type GpuSample struct {
	Util  float64
	MemMB int
}

// GpuStatsByPid returns GPU utilization and memory per PID (unsupported).
func GpuStatsByPid() map[int]GpuSample {
	return map[int]GpuSample{}
}
