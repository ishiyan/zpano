package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// Engulfing: a two-candle reversal pattern.
//
// Must have:
// - first candle and second candle have opposite colors,
// - second candle's real body engulfs the first (at least one end strictly
//   exceeds, the other can match).
//
// Category B: direction from 2nd candle color (continuous).
// Opposite-color check stays crisp (doji edge case).
func Engulfing(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(2) {
		return 0.0
	}

	b1 := cp.Bar(2)
	b2 := cp.Bar(1)

	// Opposite colors — crisp gate (TA-Lib convention: c >= o is white).
	color1 := 1
	if b1.C < b1.O {
		color1 = -1
	}
	color2 := 1
	if b2.C < b2.O {
		color2 = -1
	}
	if color1 == color2 {
		return 0.0
	}

	// Fuzzy engulfment: 2nd body upper >= 1st body upper AND
	//                    2nd body lower <= 1st body lower.
	upper1 := max(b1.O, b1.C)
	lower1 := min(b1.O, b1.C)
	upper2 := max(b2.O, b2.C)
	lower2 := min(b2.O, b2.C)

	// Width based on the equal criterion for tight comparisons.
	eqAvg := cp.AvgCS(cp.Equal, 1)
	eqWidth := cp.FuzzRatio * eqAvg
	if eqAvg <= 0.0 {
		eqWidth = 0.0
	}

	muUpper := cp.MuGeRaw(upper2, upper1, eqWidth)
	muLower := cp.MuLtRaw(lower2, lower1, eqWidth)

	confidence := fuzzy.TProductAll(muUpper, muLower)

	// Direction sign from 2nd candle (TA-Lib: c >= o is bullish).
	direction := 1.0
	if b2.C < b2.O {
		direction = -1.0
	}
	return direction * confidence * 100.0
}
