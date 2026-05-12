package patterns

import (
	"zpano/candlestickpatterns/core"
)

// HikkakeModified: a four-candle pattern with near criterion.
//
// Detection outputs +100.0/-100.0, confirmation outputs +200.0/-200.0,
// 0.0 otherwise.
func HikkakeModified(cp *core.CandlestickPatterns) float64 {
	if cp.Count < 4 {
		return 0.0
	}

	if cp.HikmodPatternIdx == cp.Count && cp.HikmodPatternResult != 0 {
		return cp.HikmodPatternResult
	}

	if cp.HikmodConfirmed {
		return cp.HikmodLastSignal
	}

	return 0.0
}
