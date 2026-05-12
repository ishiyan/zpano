package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// UpsideGapTwoCrows: a three-candle bearish pattern.
//
// Must have:
//   - first candle: long white,
//   - second candle: small black that gaps up from the first,
//   - third candle: black that engulfs the second candle's body and
//     closes above the first candle's close (gap not filled).
//
// The meaning of "long" is specified with LongBody.
// The meaning of "short" is specified with ShortBody.
//
// Category A: always bearish (continuous).
func UpsideGapTwoCrows(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(3, cp.LongBody, cp.ShortBody) {
		return 0.0
	}

	b1 := cp.Bar(3)
	b2 := cp.Bar(2)
	b3 := cp.Bar(1)

	// Crisp gates: colors.
	if !(core.IsWhite(b1.O, b1.C) && core.IsBlack(b2.O, b2.C) && core.IsBlack(b3.O, b3.C)) {
		return 0.0
	}

	// Crisp: gap up from first to second.
	if !core.IsRealBodyGapUp(b1.O, b1.C, b2.O, b2.C) {
		return 0.0
	}

	// Crisp: third engulfs second (o3 > o2 and c3 < c2) and closes above c1.
	if !(b3.O > b2.O && b3.C < b2.C && b3.C > b1.C) {
		return 0.0
	}

	// Fuzzy: first candle is long.
	muLong1 := cp.MuGreater(core.RealBodyLen(b1.O, b1.C), cp.LongBody, 3)

	// Fuzzy: second candle is short.
	muShort2 := cp.MuLess(core.RealBodyLen(b2.O, b2.C), cp.ShortBody, 2)

	confidence := fuzzy.TProductAll(muLong1, muShort2)

	return -confidence * 100.0
}
