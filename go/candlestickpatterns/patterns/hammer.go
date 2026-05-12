package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// Hammer: a two-candle bullish reversal pattern.
//
// Must have:
// - small real body,
// - long lower shadow,
// - no or very short upper shadow,
// - body is below or near the lows of the previous candle.
func Hammer(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(2, cp.ShortBody, cp.LongShadow, cp.VeryShortShadow, cp.Near) {
		return 0.0
	}

	b1 := cp.Bar(2)
	b2 := cp.Bar(1)

	nearAvg := cp.AvgCS(cp.Near, 2)
	nearWidth := cp.FuzzRatio * nearAvg
	if nearAvg <= 0.0 {
		nearWidth = 0.0
	}

	muShort := cp.MuLess(core.RealBodyLen(b2.O, b2.C), cp.ShortBody, 1)
	muLongLS := cp.MuGreater(core.LowerShadow(b2.O, b2.L, b2.C), cp.LongShadow, 1)
	muShortUS := cp.MuLess(core.UpperShadow(b2.O, b2.H, b2.C), cp.VeryShortShadow, 1)
	muNearLow := cp.MuLtRaw(min(b2.O, b2.C), b1.L+nearAvg, nearWidth)

	confidence := fuzzy.TProductAll(muShort, muLongLS, muShortUS, muNearLow)
	return confidence * 100.0
}
