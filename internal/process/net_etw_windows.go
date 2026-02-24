//go:build windows && cgo

package process

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/Velocidex/etw"
	"github.com/Velocidex/ordereddict"
	"golang.org/x/sys/windows"
)

var (
	etwOnce    sync.Once
	etwErr     error
	etwTotals  = map[int]uint64{}
	etwMu      sync.RWMutex
	etwRunning atomic.Bool
	sizeRx     = regexp.MustCompile(`(?i)(?:\bsize\b|payloadsize|transfersize|bytes|length)\D{0,8}(\d{1,12})`)
	pidRx      = regexp.MustCompile(`(?i)(?:\bpid\b|processid|process_id)\D{0,8}(\d{1,10})`)
)

const etwInvalidPID = int(^uint32(0)) // 4294967295

// StartETWNetwork starts a kernel ETW session for network events.
func StartETWNetwork() error {
	etwOnce.Do(func() {
		etwErr = startETWNetwork()
	})
	return etwErr
}

func startETWNetwork() error {
	cb := func(e *etw.Event) {
		parsed := e.Parsed()
		props := e.Props()
		kernelType := strings.ToLower(e.Header.KernelLoggerType.String())

		// Session is configured only for network events, so do not rely on a
		// hardcoded enum subset that can vary across Windows versions.
		size := extractNetSize(props)
		if size == 0 {
			size = extractNetSizeParsed(parsed)
		}
		if size == 0 {
			size = extractSizeTextFallback(props, parsed)
		}

		pid := int(e.Header.ProcessID)
		if !isReasonablePID(pid) {
			pid = 0
		}
		if pid <= 0 {
			if p := extractPID(props); p > 0 {
				pid = p
			}
		}
		if pid <= 0 {
			if p := extractPIDParsed(parsed); p > 0 {
				pid = p
			}
		}
		if pid <= 0 {
			pid = extractPIDTextFallback(props, parsed)
		}
		if !isReasonablePID(pid) {
			pid = 0
		}
		if pid <= 0 {
			return
		}
		if size == 0 {
			// Some ETW decoders omit payload size for network events on specific
			// systems; still count such events minimally so NET is not pinned to 0.
			if strings.Contains(kernelType, "udp") ||
				strings.Contains(kernelType, "tcp") ||
				strings.Contains(kernelType, "send") ||
				strings.Contains(kernelType, "recv") {
				size = 1
			} else {
				return
			}
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
		} else if strings.Contains(strings.ToLower(err.Error()), "already exist") {
			_ = etw.KillSession(etw.KernelTraceSessionName)
			session, err = etw.NewKernelTraceSession(opts, cb)
		}
	}
	if err != nil {
		return err
	}

	etwRunning.Store(true)

	// On some systems network kernel events appear only when the session
	// explicitly subscribes to SystemTraceProvider
	// ({9E814AAD-3204-11D2-9A82-006008A86939}), similar to tracer usage.
	// Errors are non-fatal here because sessions may still deliver events
	// depending on OS/session state.
	if guid, gerr := windows.GUIDFromString(etw.KernelTraceControlGUIDString); gerr == nil {
		_ = session.SubscribeToProvider(etw.SessionOptions{
			Guid:          guid,
			Level:         etw.TraceLevel(255),
			CaptureState:  true,
			EnableMapInfo: false,
		})
	}

	go func() {
		if err := session.Process(); err != nil {
			etwRunning.Store(false)
			etwNetworkActive.Store(false)
			etwLastError.Store(err.Error())
		}
	}()
	return nil
}

func extractPID(props *ordereddict.Dict) int {
	if props == nil {
		return 0
	}
	keys := []string{"PID", "ProcessID", "ProcessId", "processId", "processID", "pid"}
	for _, k := range keys {
		if n, ok := propUint(props, k); ok && n > 0 {
			return int(n)
		}
	}
	for _, k := range props.Keys() {
		if !strings.Contains(strings.ToLower(k), "pid") {
			continue
		}
		if n, ok := propUint(props, k); ok && n > 0 {
			return int(n)
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
		"DataLength", "dataLength", "DataLen", "dataLen", "NumBytes", "numBytes",
	}
	for _, k := range keys {
		if n, ok := propUint(props, k); ok && n > 0 {
			return n
		}
	}
	for _, k := range props.Keys() {
		if !strings.Contains(strings.ToLower(k), "size") && !strings.Contains(strings.ToLower(k), "length") {
			continue
		}
		if n, ok := propUint(props, k); ok && n > 0 {
			return n
		}
	}
	// Fallback heuristic for provider/schema variations: choose the largest
	// numeric payload-like field, excluding obvious identity/address keys.
	var best uint64
	for _, k := range props.Keys() {
		lk := strings.ToLower(k)
		if strings.Contains(lk, "pid") ||
			strings.Contains(lk, "port") ||
			strings.Contains(lk, "addr") ||
			strings.Contains(lk, "ip") {
			continue
		}
		n, ok := propUint(props, k)
		if !ok || n == 0 {
			continue
		}
		// Network event payload sizes are typically bounded; ignore huge ids.
		if n > 16*1024*1024 {
			continue
		}
		if n > best {
			best = n
		}
	}
	if best > 0 {
		return best
	}
	return 0
}

func propUint(props *ordereddict.Dict, key string) (uint64, bool) {
	if props == nil {
		return 0, false
	}
	if v, ok := props.GetInt64(key); ok {
		if v <= 0 {
			return 0, false
		}
		return uint64(v), true
	}
	raw, ok := props.Get(key)
	if !ok || raw == nil {
		return 0, false
	}
	switch t := raw.(type) {
	case uint8:
		return uint64(t), true
	case uint16:
		return uint64(t), true
	case uint32:
		return uint64(t), true
	case uint64:
		return t, true
	case uint:
		return uint64(t), true
	case int8:
		if t <= 0 {
			return 0, false
		}
		return uint64(t), true
	case int16:
		if t <= 0 {
			return 0, false
		}
		return uint64(t), true
	case int32:
		if t <= 0 {
			return 0, false
		}
		return uint64(t), true
	case int64:
		if t <= 0 {
			return 0, false
		}
		return uint64(t), true
	case int:
		if t <= 0 {
			return 0, false
		}
		return uint64(t), true
	case float32:
		if t <= 0 {
			return 0, false
		}
		return uint64(math.Round(float64(t))), true
	case float64:
		if t <= 0 {
			return 0, false
		}
		return uint64(math.Round(t)), true
	case string:
		if n := parseUint(t); n > 0 {
			return n, true
		}
	case []byte:
		if n := parseUint(string(t)); n > 0 {
			return n, true
		}
	default:
		if n := parseUint(fmt.Sprint(raw)); n > 0 {
			return n, true
		}
	}
	return 0, false
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

func extractPIDParsed(parsed *ordereddict.Dict) int {
	if parsed == nil {
		return 0
	}
	// Header.ProcessID
	if headerAny, ok := parsed.Get("Header"); ok {
		if n := dictValueToInt(headerAny, "ProcessID"); n > 0 {
			return n
		}
		if n := dictValueToInt(headerAny, "PID"); n > 0 {
			return n
		}
	}
	// EventProperties.PID
	if propsAny, ok := parsed.Get("EventProperties"); ok {
		if n := dictValueToInt(propsAny, "PID"); n > 0 {
			return n
		}
		if n := dictValueToInt(propsAny, "ProcessID"); n > 0 {
			return n
		}
	}
	// Generic recursive fallback for provider/schema variations.
	if n := findIntRecursive(parsed, "pid", "processid", "process_id"); n > 0 {
		return n
	}
	return 0
}

func extractNetSizeParsed(parsed *ordereddict.Dict) uint64 {
	if parsed == nil {
		return 0
	}
	if propsAny, ok := parsed.Get("EventProperties"); ok {
		for _, k := range []string{
			"size", "Size", "PayloadSize", "payloadSize", "TransferSize", "transferSize",
			"Bytes", "bytes", "Length", "length", "PacketSize", "packetSize",
		} {
			if n := dictValueToUint64(propsAny, k); n > 0 {
				return n
			}
		}
	}
	// Generic recursive fallback for provider/schema variations.
	if n := findUintRecursive(parsed, "size", "payloadsize", "transfersize", "bytes", "length", "packetsize"); n > 0 {
		return n
	}
	return 0
}

func dictValueToInt(container interface{}, key string) int {
	return anyToInt(dictValue(container, key))
}

func dictValueToUint64(container interface{}, key string) uint64 {
	return anyToUint64(dictValue(container, key))
}

func dictValue(container interface{}, key string) interface{} {
	switch t := container.(type) {
	case *ordereddict.Dict:
		if v, ok := t.Get(key); ok {
			return v
		}
	case ordereddict.Dict:
		if v, ok := t.Get(key); ok {
			return v
		}
	case map[string]interface{}:
		return t[key]
	}
	return nil
}

func findIntRecursive(container interface{}, keys ...string) int {
	n := findUintRecursive(container, keys...)
	if n == 0 || n > math.MaxInt32 {
		return 0
	}
	return int(n)
}

func findUintRecursive(container interface{}, keys ...string) uint64 {
	if container == nil {
		return 0
	}
	keySet := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		keySet[strings.ToLower(k)] = struct{}{}
	}
	var scan func(v interface{}) uint64
	scan = func(v interface{}) uint64 {
		if v == nil {
			return 0
		}
		switch t := v.(type) {
		case *ordereddict.Dict:
			for _, k := range t.Keys() {
				val, _ := t.Get(k)
				if _, ok := keySet[strings.ToLower(k)]; ok {
					if n := anyToUint64(val); n > 0 {
						return n
					}
				}
				if n := scan(val); n > 0 {
					return n
				}
			}
			return 0
		case ordereddict.Dict:
			tt := t
			return scan(&tt)
		case map[string]interface{}:
			for k, val := range t {
				if _, ok := keySet[strings.ToLower(k)]; ok {
					if n := anyToUint64(val); n > 0 {
						return n
					}
				}
				if n := scan(val); n > 0 {
					return n
				}
			}
			return 0
		case []interface{}:
			for _, it := range t {
				if n := scan(it); n > 0 {
					return n
				}
			}
			return 0
		default:
			rv := reflect.ValueOf(v)
			if rv.Kind() == reflect.Pointer && !rv.IsNil() {
				return scan(rv.Elem().Interface())
			}
		}
		return 0
	}
	return scan(container)
}

func extractSizeTextFallback(props *ordereddict.Dict, parsed *ordereddict.Dict) uint64 {
	if props != nil {
		if n := firstMatchUint(sizeRx, fmt.Sprintf("%v", props)); n > 0 {
			return n
		}
	}
	if parsed != nil {
		if n := firstMatchUint(sizeRx, fmt.Sprintf("%v", parsed)); n > 0 {
			return n
		}
	}
	return 0
}

func extractPIDTextFallback(props *ordereddict.Dict, parsed *ordereddict.Dict) int {
	if props != nil {
		if n := firstMatchUint(pidRx, fmt.Sprintf("%v", props)); n > 0 && n <= math.MaxInt32 {
			return int(n)
		}
	}
	if parsed != nil {
		if n := firstMatchUint(pidRx, fmt.Sprintf("%v", parsed)); n > 0 && n <= math.MaxInt32 {
			return int(n)
		}
	}
	return 0
}

func firstMatchUint(rx *regexp.Regexp, s string) uint64 {
	m := rx.FindStringSubmatch(s)
	if len(m) < 2 {
		return 0
	}
	return parseUint(m[1])
}

func isReasonablePID(pid int) bool {
	if pid <= 0 {
		return false
	}
	if pid == etwInvalidPID {
		return false
	}
	// Practical upper bound for Windows PIDs in normal operation.
	if pid > 10_000_000 {
		return false
	}
	return true
}

func anyToInt(v interface{}) int {
	n := anyToUint64(v)
	if n == 0 || n > math.MaxInt32 {
		return 0
	}
	return int(n)
}

func anyToUint64(v interface{}) uint64 {
	switch t := v.(type) {
	case nil:
		return 0
	case uint8:
		return uint64(t)
	case uint16:
		return uint64(t)
	case uint32:
		return uint64(t)
	case uint64:
		return t
	case uint:
		return uint64(t)
	case int8:
		if t <= 0 {
			return 0
		}
		return uint64(t)
	case int16:
		if t <= 0 {
			return 0
		}
		return uint64(t)
	case int32:
		if t <= 0 {
			return 0
		}
		return uint64(t)
	case int64:
		if t <= 0 {
			return 0
		}
		return uint64(t)
	case int:
		if t <= 0 {
			return 0
		}
		return uint64(t)
	case float32:
		if t <= 0 {
			return 0
		}
		return uint64(math.Round(float64(t)))
	case float64:
		if t <= 0 {
			return 0
		}
		return uint64(math.Round(t))
	case string:
		return parseUint(t)
	case []byte:
		return parseUint(string(t))
	default:
		return parseUint(fmt.Sprint(v))
	}
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

func etwSnapshotTotals() map[int]uint64 {
	if !etwRunning.Load() {
		return nil
	}
	etwMu.RLock()
	defer etwMu.RUnlock()
	out := make(map[int]uint64, len(etwTotals))
	for pid, total := range etwTotals {
		out[pid] = total
	}
	return out
}

func etwDebugSummary(limit int) string {
	if !etwRunning.Load() {
		return "etw:off"
	}
	etwMu.RLock()
	defer etwMu.RUnlock()
	if len(etwTotals) == 0 {
		return "etw:pids=0 bytes=0"
	}
	type kv struct {
		pid int
		b   uint64
	}
	list := make([]kv, 0, len(etwTotals))
	var total uint64
	for pid, b := range etwTotals {
		total += b
		list = append(list, kv{pid: pid, b: b})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].b > list[j].b })
	if limit <= 0 {
		limit = 3
	}
	if limit > len(list) {
		limit = len(list)
	}
	parts := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		parts = append(parts, fmt.Sprintf("%d:%d", list[i].pid, list[i].b))
	}
	return fmt.Sprintf("etw:pids=%d bytes=%d top=%s", len(list), total, strings.Join(parts, ","))
}
