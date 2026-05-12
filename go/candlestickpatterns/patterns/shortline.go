package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// ShortLine detects a Short Line pattern: a one-candle pattern.
//
// A candle with a short body, short upper shadow, and short lower shadow.
//
// The meaning of "short" for body is specified with ShortBody.
// The meaning of "short" for shadows is specified with ShortShadow.
//
// Category C: color determines sign.
func ShortLine(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(1, cp.ShortBody, cp.ShortShadow) {
		return 0.0
	}

	b := cp.Bar(1)

	muShortBody := cp.MuLess(core.RealBodyLen(b.O, b.C), cp.ShortBody, 1)
	muShortUS := cp.MuLess(core.UpperShadow(b.O, b.H, b.C), cp.ShortShadow, 1)
	muShortLS := cp.MuLess(core.LowerShadow(b.O, b.L, b.C), cp.ShortShadow, 1)

	confidence := fuzzy.TProductAll(muShortBody, muShortUS, muShortLS)

	if core.IsWhite(b.O, b.C) {
		return confidence * 100.0
	}
	if core.IsBlack(b.O, b.C) {
		return -confidence * 100.0
	}
	return 0.0
}
