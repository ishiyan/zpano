package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// AdvanceBlock detects a bearish three-candle pattern.
//
// Three white candles with consecutively higher closes and opens, but
// showing signs of weakening (diminishing bodies, growing upper shadows).
//
// Category A: always bearish (continuous).
func AdvanceBlock(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(3, cp.LongBody, cp.ShortShadow, cp.LongShadow, cp.Near, cp.Far) {
		return 0.0
	}

	b1 := cp.Bar(3)
	b2 := cp.Bar(2)
	b3 := cp.Bar(1)

	// Crisp gates: all white with rising closes.
	if !(core.IsWhite(b1.O, b1.C) && core.IsWhite(b2.O, b2.C) && core.IsWhite(b3.O, b3.C) &&
		b3.C > b2.C && b2.C > b1.C) {
		return 0.0
	}
	if !(b2.O > b1.O) {
		return 0.0
	}
	if !(b3.O > b2.O) {
		return 0.0
	}

	rb1 := core.RealBodyLen(b1.O, b1.C)
	rb2 := core.RealBodyLen(b2.O, b2.C)
	rb3 := core.RealBodyLen(b3.O, b3.C)

	// Fuzzy: 2nd opens within/near 1st body (upper bound).
	near3 := cp.AvgCS(cp.Near, 3)
	near3Width := cp.FuzzRatio * near3
	if near3 <= 0.0 {
		near3Width = 0.0
	}
	muO2Near := cp.MuLtRaw(b2.O, b1.C+near3, near3Width)

	// Fuzzy: 3rd opens within/near 2nd body (upper bound).
	near2 := cp.AvgCS(cp.Near, 2)
	near2Width := cp.FuzzRatio * near2
	if near2 <= 0.0 {
		near2Width = 0.0
	}
	muO3Near := cp.MuLtRaw(b3.O, b2.C+near2, near2Width)

	// Fuzzy: first candle long body.
	muLong1 := cp.MuGreater(rb1, cp.LongBody, 3)
	// Fuzzy: first candle short upper shadow.
	muUS1 := cp.MuLess(core.UpperShadow(b1.O, b1.H, b1.C), cp.ShortShadow, 3)

	// At least one weakness condition must hold (OR → max).
	far2 := cp.AvgCS(cp.Far, 3)
	far2Width := cp.FuzzRatio * far2
	if far2 <= 0.0 {
		far2Width = 0.0
	}
	far1 := cp.AvgCS(cp.Far, 2)
	far1Width := cp.FuzzRatio * far1
	if far1 <= 0.0 {
		far1Width = 0.0
	}
	near1 := cp.AvgCS(cp.Near, 2)
	near1Width := cp.FuzzRatio * near1
	if near1 <= 0.0 {
		near1Width = 0.0
	}

	// Branch 1: 2 far smaller than 1 AND 3 not longer than 2
	muB1A := cp.MuLtRaw(rb2, rb1-far2, far2Width)
	muB1B := cp.MuLtRaw(rb3, rb2+near1, near1Width)
	branch1 := fuzzy.TProductAll(muB1A, muB1B)

	// Branch 2: 3 far smaller than 2
	branch2 := cp.MuLtRaw(rb3, rb2-far1, far1Width)

	// Branch 3: 3 < 2 AND 2 < 1 AND (3 or 2 has non-short upper shadow)
	rb3Width := cp.FuzzRatio * rb2
	if rb2 <= 0.0 {
		rb3Width = 0.0
	}
	rb2Width := cp.FuzzRatio * rb1
	if rb1 <= 0.0 {
		rb2Width = 0.0
	}
	muB3A := cp.MuLtRaw(rb3, rb2, rb3Width)
	muB3B := cp.MuLtRaw(rb2, rb1, rb2Width)
	muB3US3 := cp.MuGreater(core.UpperShadow(b3.O, b3.H, b3.C), cp.ShortShadow, 1)
	muB3US2 := cp.MuGreater(core.UpperShadow(b2.O, b2.H, b2.C), cp.ShortShadow, 2)
	branch3 := fuzzy.TProductAll(muB3A, muB3B, max(muB3US3, muB3US2))

	// Branch 4: 3 < 2 AND 3 has long upper shadow
	muB4A := cp.MuLtRaw(rb3, rb2, rb3Width)
	muB4B := cp.MuGreater(core.UpperShadow(b3.O, b3.H, b3.C), cp.LongShadow, 1)
	branch4 := fuzzy.TProductAll(muB4A, muB4B)

	weakness := max(branch1, branch2, branch3, branch4)

	confidence := fuzzy.TProductAll(muO2Near, muO3Near, muLong1, muUS1, weakness)

	return -confidence * 100.0
}
