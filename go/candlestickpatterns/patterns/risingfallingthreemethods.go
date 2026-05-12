package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// RisingFallingThreeMethods detects a Rising/Falling Three Methods pattern:
// a five-candle continuation pattern.
//
// Uses TA-Lib logic: opposite-color check via color multiplication,
// real-body overlap (not full candle containment), sequential closes,
// 5th opens beyond 4th close.
//
// Category B: direction from 1st candle color (crisp sign).
func RisingFallingThreeMethods(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(5, cp.LongBody, cp.ShortBody) {
		return 0.0
	}

	b1 := cp.Bar(5)
	b2 := cp.Bar(4)
	b3 := cp.Bar(3)
	b4 := cp.Bar(2)
	b5 := cp.Bar(1)

	// Fuzzy: 1st long, 2nd-4th short, 5th long.
	muLong1 := cp.MuGreater(core.RealBodyLen(b1.O, b1.C), cp.LongBody, 5)
	muShort2 := cp.MuLess(core.RealBodyLen(b2.O, b2.C), cp.ShortBody, 4)
	muShort3 := cp.MuLess(core.RealBodyLen(b3.O, b3.C), cp.ShortBody, 3)
	muShort4 := cp.MuLess(core.RealBodyLen(b4.O, b4.C), cp.ShortBody, 2)
	muLong5 := cp.MuGreater(core.RealBodyLen(b5.O, b5.C), cp.LongBody, 1)

	// Determine color of 1st candle: +1 white, -1 black — crisp sign.
	color1 := 1.0
	if !core.IsWhite(b1.O, b1.C) {
		color1 = -1.0
	}

	// Color check: white, 3 black, white OR black, 3 white, black — crisp.
	c2 := 1.0
	if !core.IsWhite(b2.O, b2.C) {
		c2 = -1.0
	}
	c3 := 1.0
	if !core.IsWhite(b3.O, b3.C) {
		c3 = -1.0
	}
	c4 := 1.0
	if !core.IsWhite(b4.O, b4.C) {
		c4 = -1.0
	}
	c5 := 1.0
	if !core.IsWhite(b5.O, b5.C) {
		c5 = -1.0
	}

	if !(c2 == -color1 && c3 == c2 && c4 == c3 && c5 == -c4) {
		return 0.0
	}

	// 2nd to 4th hold within 1st: a part of the real body overlaps 1st range — crisp.
	if !(min(b2.O, b2.C) < b1.H && max(b2.O, b2.C) > b1.L &&
		min(b3.O, b3.C) < b1.H && max(b3.O, b3.C) > b1.L &&
		min(b4.O, b4.C) < b1.H && max(b4.O, b4.C) > b1.L) {
		return 0.0
	}

	// 2nd to 4th are falling (rising) — using color multiply trick — crisp.
	if !(b3.C*color1 < b2.C*color1 && b4.C*color1 < b3.C*color1) {
		return 0.0
	}

	// 5th opens above (below) the prior close — crisp.
	if !(b5.O*color1 > b4.C*color1) {
		return 0.0
	}

	// 5th closes above (below) the 1st close — crisp.
	if !(b5.C*color1 > b1.C*color1) {
		return 0.0
	}

	conf := fuzzy.TProductAll(muLong1, muShort2, muShort3, muShort4, muLong5)
	return color1 * conf * 100.0
}
