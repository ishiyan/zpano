package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// HaramiCross: a two-candle reversal pattern.
//
// Like Harami, but the second candle is a doji instead of just short.
//
// Must have:
// - first candle: long real body,
// - second candle: doji body contained within the first candle's real body.
//
// Category B: direction from 1st candle color (continuous).
func HaramiCross(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(2, cp.LongBody, cp.DojiBody) {
		return 0.0
	}

	b1 := cp.Bar(2)
	b2 := cp.Bar(1)

	// Fuzzy size conditions.
	muLong1 := cp.MuGreater(core.RealBodyLen(b1.O, b1.C), cp.LongBody, 2)
	muDoji2 := cp.MuLess(core.RealBodyLen(b2.O, b2.C), cp.DojiBody, 1)

	// Fuzzy containment: 1st body encloses 2nd body.
	eqAvg := cp.AvgCS(cp.Equal, 1)
	eqWidth := cp.FuzzRatio * eqAvg
	if eqAvg <= 0.0 {
		eqWidth = 0.0
	}

	muEncUpper := cp.MuGeRaw(max(b1.O, b1.C), max(b2.O, b2.C), eqWidth)
	muEncLower := cp.MuLtRaw(min(b1.O, b1.C), min(b2.O, b2.C), eqWidth)

	confidence := fuzzy.TProductAll(muLong1, muDoji2, muEncUpper, muEncLower)

	// Direction: opposite of 1st candle color.
	direction := -1.0
	if b1.C < b1.O {
		direction = 1.0
	}
	return direction * confidence * 100.0
}
