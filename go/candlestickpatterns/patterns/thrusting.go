package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// Thrusting: a two-candle bearish continuation pattern.
//
// Must have:
//   - first candle: long black,
//   - second candle: white, opens below the prior candle's low, closes
//     into the prior candle's real body but below the midpoint, and the
//     close is not equal to the prior candle's close (to distinguish
//     from in-neck).
//
// The meaning of "long" is specified with LongBody.
// The meaning of "equal" is specified with Equal.
//
// Category A: always bearish (continuous).
func Thrusting(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(2, cp.LongBody, cp.Equal) {
		return 0.0
	}

	b1 := cp.Bar(2)
	b2 := cp.Bar(1)

	rb1 := core.RealBodyLen(b1.O, b1.C)

	// Crisp gates: color checks and open below prior low.
	if !(core.IsBlack(b1.O, b1.C) && core.IsWhite(b2.O, b2.C) && b2.O < b1.L) {
		return 0.0
	}

	// Fuzzy conditions.
	muLong1 := cp.MuGreater(rb1, cp.LongBody, 2)

	// Close above prior close + equal avg (not equal to prior close).
	eq := cp.AvgCS(cp.Equal, 2)
	eqWidth := 0.0
	if eq > 0.0 {
		eqWidth = cp.FuzzRatio * eq
	}
	muAboveClose := cp.MuGtRaw(b2.C, b1.C+eq, eqWidth)

	// Close at or below midpoint of prior body: c2 <= c1 + rb1 * 0.5
	mid := b1.C + rb1*0.5
	midWidth := 0.0
	if rb1 > 0.0 {
		midWidth = cp.FuzzRatio * rb1 * 0.5
	}
	muBelowMid := cp.MuLtRaw(b2.C, mid, midWidth)

	confidence := fuzzy.TProductAll(muLong1, muAboveClose, muBelowMid)

	return -1.0 * confidence * 100.0
}
