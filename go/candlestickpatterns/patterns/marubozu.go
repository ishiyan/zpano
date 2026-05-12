package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// Marubozu detects the Marubozu pattern (1-candle).
//
// Must have:
//   - long real body,
//   - very short upper shadow,
//   - very short lower shadow.
//
// The meaning of "long" is specified with cp.LongBody.
// The meaning of "very short" for shadows is specified with cp.VeryShortShadow.
//
// Category B: direction from candle color.
func Marubozu(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(1, cp.LongBody, cp.VeryShortShadow) {
		return 0.0
	}

	b := cp.Bar(1)
	// Fuzzy: long body, very short shadows.
	muLong := cp.MuGreater(core.RealBodyLen(b.O, b.C), cp.LongBody, 1)
	muUS := cp.MuLess(core.UpperShadow(b.O, b.H, b.C), cp.VeryShortShadow, 1)
	muLS := cp.MuLess(core.LowerShadow(b.O, b.L, b.C), cp.VeryShortShadow, 1)

	confidence := fuzzy.TProductAll(muLong, muUS, muLS)
	// Crisp direction from color.
	direction := 1
	if !core.IsWhite(b.O, b.C) {
		direction = -1
	}
	return float64(direction) * confidence * 100.0
}
