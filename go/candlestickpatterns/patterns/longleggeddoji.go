package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// LongLeggedDoji detects the Long Legged Doji pattern (1-candle).
//
// Must have:
//   - doji body (very small real body),
//   - one or both shadows are long.
func LongLeggedDoji(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(1, cp.DojiBody, cp.LongShadow) {
		return 0.0
	}

	b := cp.Bar(1)
	muDoji := cp.MuLess(core.RealBodyLen(b.O, b.C), cp.DojiBody, 1)
	muLongUS := cp.MuGreater(core.UpperShadow(b.O, b.H, b.C), cp.LongShadow, 1)
	muLongLS := cp.MuGreater(core.LowerShadow(b.O, b.L, b.C), cp.LongShadow, 1)
	muAnyLong := fuzzy.SMax(muLongUS, muLongLS)

	confidence := fuzzy.TProductAll(muDoji, muAnyLong)
	return confidence * 100.0
}
