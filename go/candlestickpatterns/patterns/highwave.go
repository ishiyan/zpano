package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// HighWave: a one-candle pattern.
//
// Must have:
// - short real body,
// - very long upper shadow,
// - very long lower shadow.
//
// The meaning of "short" is specified with cp.ShortBody.
// The meaning of "very long" (shadow) is specified with cp.VeryLongShadow.
//
// Category C: color determines sign.
func HighWave(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(1, cp.ShortBody, cp.VeryLongShadow) {
		return 0.0
	}

	b := cp.Bar(1)
	muShort := cp.MuLess(core.RealBodyLen(b.O, b.C), cp.ShortBody, 1)
	muLongUS := cp.MuGreater(core.UpperShadow(b.O, b.H, b.C), cp.VeryLongShadow, 1)
	muLongLS := cp.MuGreater(core.LowerShadow(b.O, b.L, b.C), cp.VeryLongShadow, 1)

	confidence := fuzzy.TProductAll(muShort, muLongUS, muLongLS)
	if core.IsWhite(b.O, b.C) {
		return confidence * 100.0
	}
	return -confidence * 100.0
}
