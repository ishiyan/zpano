package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// DojiStar detects a two-candle reversal pattern.
//
// Must have:
//   - first candle: long real body,
//   - second candle: doji that gaps away from the first candle.
//
//   - bearish: first candle is long white, doji gaps up,
//   - bullish: first candle is long black, doji gaps down.
//
// The meaning of "long" is specified with LongBody.
// The meaning of "doji" is specified with DojiBody.
//
// Category B: direction from 1st candle color (continuous).
func DojiStar(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(2, cp.LongBody, cp.DojiBody) {
		return 0.0
	}

	b1 := cp.Bar(2)
	b2 := cp.Bar(1)

	color1 := 1
	if b1.C < b1.O {
		color1 = -1
	}

	// Crisp gates: gap direction must match color.
	if color1 == 1 && !core.IsRealBodyGapUp(b1.O, b1.C, b2.O, b2.C) {
		return 0.0
	}
	if color1 == -1 && !core.IsRealBodyGapDown(b1.O, b1.C, b2.O, b2.C) {
		return 0.0
	}

	// Fuzzy conditions.
	muLong1 := cp.MuGreater(core.RealBodyLen(b1.O, b1.C), cp.LongBody, 2)
	muDoji2 := cp.MuLess(core.RealBodyLen(b2.O, b2.C), cp.DojiBody, 1)

	confidence := fuzzy.TProductAll(muLong1, muDoji2)
	// Direction: opposite of 1st candle color.
	direction := -1.0
	if color1 == -1 {
		direction = 1.0
	}
	return direction * confidence * 100.0
}
