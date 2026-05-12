package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// Stalled detects a Stalled (Deliberation) pattern: a three-candle bearish pattern.
//
// Three white candles with progressively higher closes:
//   - first candle: long white body,
//   - second candle: long white body, opens within or near the first
//     candle's body, very short upper shadow,
//   - third candle: small body that rides on the shoulder of the second
//     (opens near the second's close, accounting for its own body size).
//
// Category A: always bearish (continuous).
func Stalled(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(3, cp.LongBody, cp.ShortBody, cp.VeryShortShadow, cp.Near) {
		return 0.0
	}

	b1 := cp.Bar(3)
	b2 := cp.Bar(2)
	b3 := cp.Bar(1)

	// Crisp gates: all white, rising closes.
	if !(core.IsWhite(b1.O, b1.C) && core.IsWhite(b2.O, b2.C) && core.IsWhite(b3.O, b3.C)) {
		return 0.0
	}
	if !(b3.C > b2.C && b2.C > b1.C) {
		return 0.0
	}
	// Crisp: o2 > o1 (opens above prior open).
	if !(b2.O > b1.O) {
		return 0.0
	}

	rb3 := core.RealBodyLen(b3.O, b3.C)

	// Fuzzy conditions.
	muLong1 := cp.MuGreater(core.RealBodyLen(b1.O, b1.C), cp.LongBody, 3)
	muLong2 := cp.MuGreater(core.RealBodyLen(b2.O, b2.C), cp.LongBody, 2)
	muUS2 := cp.MuLess(core.UpperShadow(b2.O, b2.H, b2.C), cp.VeryShortShadow, 2)

	// o2 <= c1 + near_avg (opens within or near prior body).
	near3 := cp.AvgCS(cp.Near, 3)
	near3Width := 0.0
	if near3 > 0.0 {
		near3Width = cp.FuzzRatio * near3
	}
	muO2Near := cp.MuLtRaw(b2.O, b1.C+near3, near3Width)

	// Third candle: short body.
	muShort3 := cp.MuLess(rb3, cp.ShortBody, 1)

	// o3 >= c2 - rb3 - near_avg (rides on shoulder).
	near2 := cp.AvgCS(cp.Near, 2)
	near2Width := 0.0
	if near2 > 0.0 {
		near2Width = cp.FuzzRatio * near2
	}
	muO3Shoulder := cp.MuGeRaw(b3.O, b2.C-rb3-near2, near2Width)

	confidence := fuzzy.TProductAll(muLong1, muLong2, muUS2, muO2Near, muShort3, muO3Shoulder)

	return -1.0 * confidence * 100.0
}
