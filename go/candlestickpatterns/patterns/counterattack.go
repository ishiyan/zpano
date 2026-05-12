package patterns

import (
	"math"

	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// Counterattack detects a two-candle reversal pattern.
//
// Two long candles of opposite color with closes that are equal
// (or very near equal).
//
//   - bullish: first candle is long black, second is long white,
//     closes are equal,
//   - bearish: first candle is long white, second is long black,
//     closes are equal.
//
// The meaning of "long" is specified with LongBody.
// The meaning of "equal" is specified with Equal.
//
// Category B: direction from 2nd candle color (continuous).
func Counterattack(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(2, cp.LongBody, cp.Equal) {
		return 0.0
	}

	b1 := cp.Bar(2)
	b2 := cp.Bar(1)

	// Opposite colors — crisp gate.
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

	// Fuzzy conditions.
	muLong1 := cp.MuGreater(core.RealBodyLen(b1.O, b1.C), cp.LongBody, 2)
	muLong2 := cp.MuGreater(core.RealBodyLen(b2.O, b2.C), cp.LongBody, 1)
	// Closes near equal: mu_less(abs_diff, eq_avg) — crossover at eq boundary.
	muEq := cp.MuLess(math.Abs(b2.C-b1.C), cp.Equal, 2)

	confidence := fuzzy.TProductAll(muLong1, muLong2, muEq)
	// Direction from 2nd candle color.
	direction := 1.0
	if b2.C < b2.O {
		direction = -1.0
	}
	return direction * confidence * 100.0
}
