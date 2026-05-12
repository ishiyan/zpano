package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// Takuri (Dragonfly Doji with very long lower shadow): a one-candle pattern.
//
// A doji body with a very short upper shadow and a very long lower shadow.
//
// Must have:
//   - doji body (real body smaller than doji threshold),
//   - very short upper shadow,
//   - very long lower shadow.
func Takuri(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(1, cp.DojiBody, cp.VeryShortShadow, cp.VeryLongShadow) {
		return 0.0
	}

	b := cp.Bar(1)

	muDoji := cp.MuLess(core.RealBodyLen(b.O, b.C), cp.DojiBody, 1)
	muShortUS := cp.MuLess(core.UpperShadow(b.O, b.H, b.C), cp.VeryShortShadow, 1)
	muLongLS := cp.MuGreater(core.LowerShadow(b.O, b.L, b.C), cp.VeryLongShadow, 1)

	confidence := fuzzy.TProductAll(muDoji, muShortUS, muLongLS)
	return confidence * 100.0
}
