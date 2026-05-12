package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// UniqueThreeRiver: a three-candle bullish pattern.
//
// Must have:
//   - first candle: long black,
//   - second candle: black harami (body within first body) with a lower
//     low than the first candle,
//   - third candle: small white, opens not lower than the second candle's
//     low.
//
// The meaning of "long" is specified with LongBody.
// The meaning of "short" is specified with ShortBody.
//
// Category A: always bullish (continuous).
func UniqueThreeRiver(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(3, cp.LongBody, cp.ShortBody) {
		return 0.0
	}

	b1 := cp.Bar(3)
	b2 := cp.Bar(2)
	b3 := cp.Bar(1)

	// Crisp gates: colors.
	if !(core.IsBlack(b1.O, b1.C) && core.IsBlack(b2.O, b2.C) && core.IsWhite(b3.O, b3.C)) {
		return 0.0
	}

	// Crisp: harami body containment and lower low.
	if !(b2.C > b1.C && b2.O <= b1.O && b2.L < b1.L) {
		return 0.0
	}

	// Crisp: third opens not lower than second's low.
	if !(b3.O >= b2.L) {
		return 0.0
	}

	// Fuzzy: first candle is long.
	muLong1 := cp.MuGreater(core.RealBodyLen(b1.O, b1.C), cp.LongBody, 3)

	// Fuzzy: third candle is short.
	muShort3 := cp.MuLess(core.RealBodyLen(b3.O, b3.C), cp.ShortBody, 1)

	confidence := fuzzy.TProductAll(muLong1, muShort3)

	return confidence * 100.0
}
