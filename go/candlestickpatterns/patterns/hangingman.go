package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// HangingMan: a two-candle bearish pattern.
//
// Must have:
// - small real body,
// - long lower shadow,
// - no or very short upper shadow,
// - body is above or near the highs of the previous candle.
func HangingMan(cp *core.CandlestickPatterns) float64 {
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
	muNearHigh := cp.MuGeRaw(min(b2.O, b2.C), b1.H-nearAvg, nearWidth)

	confidence := fuzzy.TProductAll(muShort, muLongLS, muShortUS, muNearHigh)
	return -confidence * 100.0
}
