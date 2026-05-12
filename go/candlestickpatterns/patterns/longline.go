package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// LongLine detects the Long Line pattern (1-candle).
//
// Must have:
//   - long real body,
//   - short upper shadow,
//   - short lower shadow.
//
// The meaning of "long" is specified with cp.LongBody.
// The meaning of "short" for shadows is specified with cp.ShortShadow.
//
// Category B: direction from candle color.
func LongLine(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(1, cp.LongBody, cp.ShortShadow) {
		return 0.0
	}

	b := cp.Bar(1)
	// Fuzzy: long body, short shadows.
	muLong := cp.MuGreater(core.RealBodyLen(b.O, b.C), cp.LongBody, 1)
	muUS := cp.MuLess(core.UpperShadow(b.O, b.H, b.C), cp.ShortShadow, 1)
	muLS := cp.MuLess(core.LowerShadow(b.O, b.L, b.C), cp.ShortShadow, 1)

	confidence := fuzzy.TProductAll(muLong, muUS, muLS)
	// Crisp direction from color.
	direction := 1
	if !core.IsWhite(b.O, b.C) {
		direction = -1
	}
	return float64(direction) * confidence * 100.0
}
