package patterns

import (
	"math"

	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

const abandonedBabyPenetrationFactor = 0.3

// AbandonedBaby: a three-candle reversal pattern.
//
// Must have:
// - first candle: long real body,
// - second candle: doji,
// - third candle: real body longer than short, opposite color to 1st,
//   closes well within 1st body,
// - upside/downside gap between 1st and doji (shadows don't touch),
// - downside/upside gap between doji and 3rd (shadows don't touch).
//
// Category C: both branches evaluated, return stronger signal.
func AbandonedBaby(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(3, cp.LongBody, cp.DojiBody, cp.ShortBody) {
		return 0.0
	}

	b1 := cp.Bar(3)
	b2 := cp.Bar(2)
	b3 := cp.Bar(1)

	// Shared fuzzy conditions: 1st long, 2nd doji, 3rd > short.
	muLong1 := cp.MuGreater(core.RealBodyLen(b1.O, b1.C), cp.LongBody, 3)
	muDoji2 := cp.MuLess(core.RealBodyLen(b2.O, b2.C), cp.DojiBody, 2)
	muShort3 := cp.MuGreater(core.RealBodyLen(b3.O, b3.C), cp.ShortBody, 1)

	penetration := abandonedBabyPenetrationFactor

	// Bearish: white-doji-black, gap up then gap down.
	bearSignal := 0.0
	if core.IsWhite(b1.O, b1.C) && core.IsBlack(b3.O, b3.C) {
		if core.IsHighLowGapUp(b1.H, b2.L) && core.IsHighLowGapDown(b2.L, b3.H) {
			rb1 := core.RealBodyLen(b1.O, b1.C)
			penThreshold := b1.C - rb1*penetration
			penWidth := cp.FuzzRatio * rb1
			if rb1 <= 0.0 {
				penWidth = 0.0
			}
			muPen := cp.MuLtRaw(b3.C, penThreshold, penWidth)
			confBear := fuzzy.TProductAll(muLong1, muDoji2, muShort3, muPen)
			bearSignal = -confBear * 100.0
		}
	}

	// Bullish: black-doji-white, gap down then gap up.
	bullSignal := 0.0
	if core.IsBlack(b1.O, b1.C) && core.IsWhite(b3.O, b3.C) {
		if core.IsHighLowGapDown(b1.L, b2.H) && core.IsHighLowGapUp(b2.H, b3.L) {
			rb1 := core.RealBodyLen(b1.O, b1.C)
			penThreshold := b1.C + rb1*penetration
			penWidth := cp.FuzzRatio * rb1
			if rb1 <= 0.0 {
				penWidth = 0.0
			}
			muPen := cp.MuGtRaw(b3.C, penThreshold, penWidth)
			confBull := fuzzy.TProductAll(muLong1, muDoji2, muShort3, muPen)
			bullSignal = confBull * 100.0
		}
	}

	if math.Abs(bullSignal) >= math.Abs(bearSignal) {
		return bullSignal
	}
	return bearSignal
}
