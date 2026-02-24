package process

import (
	"math"
	"sync/atomic"
)

var useETWNetwork atomic.Bool
var etwNetworkActive atomic.Bool
var etwLastError atomic.Value
var netScaleBits atomic.Uint64

// SetNetworkConfig toggles ETW-based network collection.
func SetNetworkConfig(enable bool) error {
	useETWNetwork.Store(enable)
	etwNetworkActive.Store(false)
	etwLastError.Store("")
	if enable {
		if err := StartETWNetwork(); err != nil {
			etwLastError.Store(err.Error())
			return err
		}
		etwNetworkActive.Store(true)
	}
	return nil
}

// SetNetworkScale applies a divisor for network rate (e.g. 100 or 1000).
func SetNetworkScale(scale float64) {
	if scale <= 0 {
		scale = 1
	}
	netScaleBits.Store(math.Float64bits(scale))
}

func getNetworkScale() float64 {
	b := netScaleBits.Load()
	if b == 0 {
		return 1
	}
	return math.Float64frombits(b)
}

func isETWActive() bool {
	return etwNetworkActive.Load()
}

// NetSource returns human-readable network source mode.
func NetSource() string {
	if useETWNetwork.Load() && isETWActive() {
		return "ETW"
	}
	if useETWNetwork.Load() {
		return "Unavailable"
	}
	return "Disabled"
}

// NetSourceError returns ETW startup error if any.
func NetSourceError() string {
	v := etwLastError.Load()
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

// NetDebug returns lightweight diagnostics for current network backend state.
func NetDebug() string {
	if useETWNetwork.Load() && !isETWActive() {
		return "etw:inactive"
	}
	return etwDebugSummary(3)
}
