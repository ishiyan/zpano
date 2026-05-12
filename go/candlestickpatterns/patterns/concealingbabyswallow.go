package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// ConcealingBabySwallow detects a four-candle bullish pattern.
//
// Must have:
//   - first candle: black marubozu (very short shadows),
//   - second candle: black marubozu (very short shadows),
//   - third candle: black, opens gapping down, upper shadow extends into
//     the prior real body (upper shadow > very-short avg),
//   - fourth candle: black, completely engulfs the third candle including
//     shadows (strict > / <).
//
// The meaning of "very short" for shadows is specified with VeryShortShadow.
//
// Category A: always bullish (continuous).
func ConcealingBabySwallow(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(4, cp.VeryShortShadow) {
		return 0.0
	}

	b1 := cp.Bar(4)
	b2 := cp.Bar(3)
	b3 := cp.Bar(2)
	b4 := cp.Bar(1)

	// Crisp gates: all black.
	if !(core.IsBlack(b1.O, b1.C) && core.IsBlack(b2.O, b2.C) &&
		core.IsBlack(b3.O, b3.C) && core.IsBlack(b4.O, b4.C)) {
		return 0.0
	}
	// Crisp: gap down and upper shadow extends into prior body.
	if !(core.IsRealBodyGapDown(b2.O, b2.C, b3.O, b3.C) && b3.H > b2.C) {
		return 0.0
	}
	// Crisp: fourth engulfs third including shadows (strict).
	if !(b4.H > b3.H && b4.L < b3.L) {
		return 0.0
	}

	// Fuzzy: first and second are marubozu (very short shadows).
	muLS1 := cp.MuLess(core.LowerShadow(b1.O, b1.L, b1.C), cp.VeryShortShadow, 4)
	muUS1 := cp.MuLess(core.UpperShadow(b1.O, b1.H, b1.C), cp.VeryShortShadow, 4)
	muLS2 := cp.MuLess(core.LowerShadow(b2.O, b2.L, b2.C), cp.VeryShortShadow, 3)
	muUS2 := cp.MuLess(core.UpperShadow(b2.O, b2.H, b2.C), cp.VeryShortShadow, 3)
	// Fuzzy: third candle upper shadow > very-short avg.
	muUS3Long := cp.MuGreater(core.UpperShadow(b3.O, b3.H, b3.C), cp.VeryShortShadow, 2)

	confidence := fuzzy.TProductAll(muLS1, muUS1, muLS2, muUS2, muUS3Long)
	return confidence * 100.0
}
