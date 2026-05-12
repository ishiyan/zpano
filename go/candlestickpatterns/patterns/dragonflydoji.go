package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// DragonflyDoji detects a one-candle pattern.
//
// Must have:
//   - doji body (very small real body relative to high-low range),
//   - no or very short upper shadow,
//   - lower shadow is not very short.
func DragonflyDoji(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(1, cp.DojiBody, cp.VeryShortShadow) {
		return 0.0
	}

	b := cp.Bar(1)
	muDoji := cp.MuLess(core.RealBodyLen(b.O, b.C), cp.DojiBody, 1)
	muShortUS := cp.MuLess(core.UpperShadow(b.O, b.H, b.C), cp.VeryShortShadow, 1)
	muLongLS := cp.MuGreater(core.LowerShadow(b.O, b.L, b.C), cp.VeryShortShadow, 1)

	confidence := fuzzy.TProductAll(muDoji, muShortUS, muLongLS)
	return confidence * 100.0
}
