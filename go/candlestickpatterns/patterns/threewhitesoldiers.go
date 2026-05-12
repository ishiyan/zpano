package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// ThreeWhiteSoldiers: a three-candle bullish pattern.
//
// Must have:
//   - three consecutive white candles with consecutively higher closes,
//   - all three have very short upper shadows,
//   - each opens within or near the prior candle's real body,
//   - none is far shorter than the prior candle,
//   - third candle is not short.
//
// Category A: always bullish (continuous).
func ThreeWhiteSoldiers(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(3, cp.ShortBody, cp.VeryShortShadow, cp.Near, cp.Far) {
		return 0.0
	}

	b1 := cp.Bar(3)
	b2 := cp.Bar(2)
	b3 := cp.Bar(1)

	// Crisp gates: all white with consecutively higher closes.
	if !(core.IsWhite(b1.O, b1.C) && core.IsWhite(b2.O, b2.C) && core.IsWhite(b3.O, b3.C) &&
		b3.C > b2.C && b2.C > b1.C) {
		return 0.0
	}

	rb1 := core.RealBodyLen(b1.O, b1.C)
	rb2 := core.RealBodyLen(b2.O, b2.C)
	rb3 := core.RealBodyLen(b3.O, b3.C)

	// Crisp: each opens above the prior open (ordering).
	if !(b2.O > b1.O && b3.O > b2.O) {
		return 0.0
	}

	// Fuzzy: very short upper shadows (all three).
	muUS1 := cp.MuLess(core.UpperShadow(b1.O, b1.H, b1.C), cp.VeryShortShadow, 3)
	muUS2 := cp.MuLess(core.UpperShadow(b2.O, b2.H, b2.C), cp.VeryShortShadow, 2)
	muUS3 := cp.MuLess(core.UpperShadow(b3.O, b3.H, b3.C), cp.VeryShortShadow, 1)

	// Fuzzy: each opens within or near the prior body (upper bound).
	near3 := cp.AvgCS(cp.Near, 3)
	near3Width := 0.0
	if near3 > 0.0 {
		near3Width = cp.FuzzRatio * near3
	}
	muO2Near := cp.MuLtRaw(b2.O, b1.C+near3, near3Width)

	near2 := cp.AvgCS(cp.Near, 2)
	near2Width := 0.0
	if near2 > 0.0 {
		near2Width = cp.FuzzRatio * near2
	}
	muO3Near := cp.MuLtRaw(b3.O, b2.C+near2, near2Width)

	// Fuzzy: not far shorter than prior candle.
	far3 := cp.AvgCS(cp.Far, 3)
	far3Width := 0.0
	if far3 > 0.0 {
		far3Width = cp.FuzzRatio * far3
	}
	muNotFar2 := cp.MuGtRaw(rb2, rb1-far3, far3Width)

	far2 := cp.AvgCS(cp.Far, 2)
	far2Width := 0.0
	if far2 > 0.0 {
		far2Width = cp.FuzzRatio * far2
	}
	muNotFar3 := cp.MuGtRaw(rb3, rb2-far2, far2Width)

	// Fuzzy: third candle is not short.
	muNotShort3 := cp.MuGreater(rb3, cp.ShortBody, 1)

	confidence := fuzzy.TProductAll(muUS1, muUS2, muUS3,
		muO2Near, muO3Near,
		muNotFar2, muNotFar3,
		muNotShort3)

	return confidence * 100.0
}
