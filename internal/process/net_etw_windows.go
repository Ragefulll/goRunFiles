//go:build windows && cgo

package process

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/Velocidex/etw"
	"github.com/Velocidex/ordereddict"
)

var (
	etwOnce    sync.Once
	etwErr     error
	etwTotals  = map[int]uint64{}
	etwMu      sync.RWMutex
	etwRunning atomic.Bool
)

// StartETWNetwork starts a kernel ETW session for network events.
func StartETWNetwork() error {
	etwOnce.Do(func() {
		etwErr = startETWNetwork()
	})
	return etwErr
}

func startETWNetwork() error {
	cb := func(e *etw.Event) {
		switch e.Header.KernelLoggerType {
		case etw.SendTCPv4, etw.RecvTCPv4, etw.SendUDPv4, etw.RecvUDPv4,
			etw.SendTCPv6, etw.RecvTCPv6, etw.SendUDPv6, etw.RecvUDPv6:
		default:
			return
		}

		size := extractNetSize(e.Props())
		if size == 0 {
			return
		}
		pid := int(e.Header.ProcessID)
		if pid <= 0 {
			if p := extractPID(e.Props()); p > 0 {
				pid = p
			}
		}
		if pid <= 0 {
			return
		}
		etwMu.Lock()
		etwTotals[pid] += size
		etwMu.Unlock()
	}

	opts := etw.RundownOptions{Network: true}
	session, err := etw.NewKernelTraceSession(opts, cb)
	if err != nil {
		var exists etw.ExistsError
		if ok := errors.As(err, &exists); ok {
			_ = etw.KillSession(exists.SessionName)
			session, err = etw.NewKernelTraceSession(opts, cb)
		}
	}
	if err != nil {
		return err
	}

	etwRunning.Store(true)
	go func() {
		_ = session.Process()
	}()
	return nil
}

func extractPID(props *ordereddict.Dict) int {
	if props == nil {
		return 0
	}
	if v, ok := props.GetString("PID"); ok {
		if pid, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return pid
		}
	}
	if v, ok := props.GetString("ProcessID"); ok {
		if pid, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return pid
		}
	}
	return 0
}

func extractNetSize(props *ordereddict.Dict) uint64 {
	if props == nil {
		return 0
	}
	keys := []string{
		"Size", "size", "PayloadSize", "payloadSize", "TransferSize", "transferSize",
		"Bytes", "bytes", "Length", "length", "PacketSize", "packetSize",
	}
	for _, k := range keys {
		if v, ok := props.GetString(k); ok {
			if n := parseUint(v); n > 0 {
				return n
			}
		}
	}
	for _, k := range props.Keys() {
		if !strings.Contains(strings.ToLower(k), "size") && !strings.Contains(strings.ToLower(k), "length") {
			continue
		}
		if v, ok := props.GetString(k); ok {
			if n := parseUint(v); n > 0 {
				return n
			}
		}
	}
	return 0
}

func parseUint(s string) uint64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	if n, err := strconv.ParseUint(s, 0, 64); err == nil {
		return n
	}
	// Strip non-digits to handle "1234 bytes" formats.
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return 0
	}
	n, err := strconv.ParseUint(b.String(), 10, 64)
	if err != nil {
		return 0
	}
	return n
}

// etwTotalBytes returns total bytes per pid from ETW.
func etwTotalBytes(pid int) (uint64, bool) {
	if !etwRunning.Load() {
		return 0, false
	}
	etwMu.RLock()
	defer etwMu.RUnlock()
	v, ok := etwTotals[pid]
	return v, ok
}
