package patterns

import (
	"math"

	"zpano/candlestickpatterns/core"
)

// StickSandwich detects a Stick Sandwich pattern: a three-candle bullish pattern.
//
// Must have:
//   - first candle: black,
//   - second candle: white, trades above the first candle's close
//     (low > first close),
//   - third candle: black, close equals the first candle's close.
//
// The meaning of "equal" is specified with Equal.
//
// Category A: always bullish (continuous).
func StickSandwich(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(3, cp.Equal) {
		return 0.0
	}

	b1 := cp.Bar(3)
	b2 := cp.Bar(2)
	b3 := cp.Bar(1)

	// Crisp gates: colors and gap.
	if !(core.IsBlack(b1.O, b1.C) && core.IsWhite(b2.O, b2.C) && core.IsBlack(b3.O, b3.C) && b2.L > b1.C) {
		return 0.0
	}

	// Fuzzy: third close equals first close (two-sided band).
	muEq := cp.MuLess(math.Abs(b3.C-b1.C), cp.Equal, 3)

	confidence := muEq

	return confidence * 100.0
}
