package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// ThreeStarsInTheSouth: a three-candle bullish pattern.
//
// Must have:
//   - all three candles are black,
//   - first candle: long body with long lower shadow,
//   - second candle: smaller body, opens within or above prior range,
//     trades lower but its low does not go below the first candle's low,
//   - third candle: small marubozu (very short shadows) engulfed by the
//     second candle's range.
//
// The meaning of "long" is specified with LongBody.
// The meaning of "short" is specified with ShortBody.
// The meaning of "long" for shadows is specified with LongShadow.
// The meaning of "very short" for shadows is specified with VeryShortShadow.
//
// Category A: always bullish (continuous).
func ThreeStarsInTheSouth(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(3, cp.LongBody, cp.ShortBody, cp.LongShadow, cp.VeryShortShadow) {
		return 0.0
	}

	b1 := cp.Bar(3)
	b2 := cp.Bar(2)
	b3 := cp.Bar(1)

	// Crisp gates: all black.
	if !(core.IsBlack(b1.O, b1.C) && core.IsBlack(b2.O, b2.C) && core.IsBlack(b3.O, b3.C)) {
		return 0.0
	}

	rb1 := core.RealBodyLen(b1.O, b1.C)
	rb2 := core.RealBodyLen(b2.O, b2.C)

	// Crisp: second body smaller than first.
	if !(rb2 < rb1) {
		return 0.0
	}

	// Crisp: second opens within or above prior range, low not below first's low.
	if !(b2.O <= b1.H && b2.O >= b1.L && b2.L >= b1.L) {
		return 0.0
	}

	// Crisp: third engulfed by second's range.
	if !(b3.H <= b2.H && b3.L >= b2.L) {
		return 0.0
	}

	// Fuzzy: first candle long body.
	muLong1 := cp.MuGreater(rb1, cp.LongBody, 3)

	// Fuzzy: first candle long lower shadow.
	muLS1 := cp.MuGreater(core.LowerShadow(b1.O, b1.L, b1.C), cp.LongShadow, 3)

	// Fuzzy: third candle short body.
	muShort3 := cp.MuLess(core.RealBodyLen(b3.O, b3.C), cp.ShortBody, 1)

	// Fuzzy: third candle very short shadows (marubozu).
	muVSUS3 := cp.MuLess(core.UpperShadow(b3.O, b3.H, b3.C), cp.VeryShortShadow, 1)
	muVSLS3 := cp.MuLess(core.LowerShadow(b3.O, b3.L, b3.C), cp.VeryShortShadow, 1)

	confidence := fuzzy.TProductAll(muLong1, muLS1, muShort3, muVSUS3, muVSLS3)

	return confidence * 100.0
}
