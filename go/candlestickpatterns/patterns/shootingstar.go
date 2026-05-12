package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// ShootingStar detects a Shooting Star pattern: a two-candle bearish reversal pattern.
//
// Must have:
//   - gap up from the previous candle (real body gap up),
//   - small real body,
//   - long upper shadow,
//   - very short lower shadow.
func ShootingStar(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(2, cp.ShortBody, cp.LongShadow, cp.VeryShortShadow) {
		return 0.0
	}

	b1 := cp.Bar(2)
	b2 := cp.Bar(1)

	if !core.IsRealBodyGapUp(b1.O, b1.C, b2.O, b2.C) {
		return 0.0
	}

	muShort := cp.MuLess(core.RealBodyLen(b2.O, b2.C), cp.ShortBody, 1)
	muLongUS := cp.MuGreater(core.UpperShadow(b2.O, b2.H, b2.C), cp.LongShadow, 1)
	muShortLS := cp.MuLess(core.LowerShadow(b2.O, b2.L, b2.C), cp.VeryShortShadow, 1)

	confidence := fuzzy.TProductAll(muShort, muLongUS, muShortLS)
	return -confidence * 100.0
}
