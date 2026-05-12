package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// HomingPigeon: a two-candle bullish pattern.
//
// Must have:
// - first candle: long black,
// - second candle: short black, real body engulfed by first candle's
//   real body.
//
// The meaning of "long" is specified with cp.LongBody.
// The meaning of "short" is specified with cp.ShortBody.
//
// Category A: always bullish (continuous).
func HomingPigeon(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(2, cp.LongBody, cp.ShortBody) {
		return 0.0
	}

	b1 := cp.Bar(2)
	b2 := cp.Bar(1)

	// Crisp gates: both black.
	if !(core.IsBlack(b1.O, b1.C) && core.IsBlack(b2.O, b2.C)) {
		return 0.0
	}

	// Fuzzy conditions.
	muLong1 := cp.MuGreater(core.RealBodyLen(b1.O, b1.C), cp.LongBody, 2)
	muShort2 := cp.MuLess(core.RealBodyLen(b2.O, b2.C), cp.ShortBody, 1)

	// Containment: second body engulfed by first body.
	// For black candles: open > close, so upper = open, lower = close.
	eqWidth := cp.FuzzRatio * cp.AvgCS(cp.Equal, 2)
	muEncUpper := cp.MuLtRaw(b2.O, b1.O, eqWidth)
	muEncLower := cp.MuGtRaw(b2.C, b1.C, eqWidth)

	confidence := fuzzy.TProductAll(muLong1, muShort2, muEncUpper, muEncLower)
	return confidence * 100.0
}
