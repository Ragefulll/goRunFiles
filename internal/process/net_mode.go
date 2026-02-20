package process

import (
	"math"
	"sync/atomic"
)

var useETWNetwork atomic.Bool
var netScaleBits atomic.Uint64

// SetNetworkConfig toggles ETW-based network collection.
func SetNetworkConfig(enable bool) {
	useETWNetwork.Store(enable)
	if enable {
		_ = StartETWNetwork()
	}
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
