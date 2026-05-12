package patterns

import (
	"math"

	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// SeparatingLines detects a Separating Lines pattern: a two-candle continuation pattern.
//
// Opposite colors with the same open. The second candle is a belt hold
// (long body with no shadow on the opening side).
//
//   - bullish: first candle is black, second is white with same open,
//     long body, very short lower shadow,
//   - bearish: first candle is white, second is black with same open,
//     long body, very short upper shadow.
//
// The meaning of "long" is specified with LongBody.
// The meaning of "very short" for shadows is specified with VeryShortShadow.
// The meaning of "equal" is specified with Equal.
//
// Category C: both branches evaluated, return stronger signal.
func SeparatingLines(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(2, cp.LongBody, cp.VeryShortShadow, cp.Equal) {
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

	// Opens near equal — fuzzy (crisp was abs(o2-o1) <= eq).
	muEq := cp.MuLess(math.Abs(b2.O-b1.O), cp.Equal, 2)

	// Long body on 2nd candle — fuzzy.
	muLong := cp.MuGreater(core.RealBodyLen(b2.O, b2.C), cp.LongBody, 1)

	// Bullish: white belt hold (very short lower shadow).
	bullSignal := 0.0
	if color2 == 1 {
		muVS := cp.MuLess(core.LowerShadow(b2.O, b2.L, b2.C), cp.VeryShortShadow, 1)
		conf := fuzzy.TProductAll(muEq, muLong, muVS)
		bullSignal = conf * 100.0
	}

	// Bearish: black belt hold (very short upper shadow).
	bearSignal := 0.0
	if color2 == -1 {
		muVS := cp.MuLess(core.UpperShadow(b2.O, b2.H, b2.C), cp.VeryShortShadow, 1)
		conf := fuzzy.TProductAll(muEq, muLong, muVS)
		bearSignal = -conf * 100.0
	}

	if math.Abs(bullSignal) >= math.Abs(bearSignal) {
		return bullSignal
	}
	return bearSignal
}
