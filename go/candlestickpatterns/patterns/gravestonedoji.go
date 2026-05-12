package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// GravestoneDoji: a one-candle pattern.
//
// Must have:
// - doji body (very small real body relative to high-low range),
// - no or very short lower shadow,
// - upper shadow is not very short.
func GravestoneDoji(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(1, cp.DojiBody, cp.VeryShortShadow) {
		return 0.0
	}

	b := cp.Bar(1)
	muDoji := cp.MuLess(core.RealBodyLen(b.O, b.C), cp.DojiBody, 1)
	muShortLS := cp.MuLess(core.LowerShadow(b.O, b.L, b.C), cp.VeryShortShadow, 1)
	muLongUS := cp.MuGreater(core.UpperShadow(b.O, b.H, b.C), cp.VeryShortShadow, 1)

	confidence := fuzzy.TProductAll(muDoji, muShortLS, muLongUS)
	return confidence * 100.0
}
