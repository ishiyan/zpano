package patterns

import (
	"math"

	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// TasukiGap: a three-candle continuation pattern.
//
// Upside Tasuki Gap (bullish):
//   - real-body gap up between 1st and 2nd candles,
//   - 2nd candle: white,
//   - 3rd candle: black, opens within 2nd white body, closes below 2nd
//     open but above 1st candle's real body top (inside the gap),
//   - 2nd and 3rd have near-equal body sizes.
//
// Downside Tasuki Gap (bearish):
//   - real-body gap down between 1st and 2nd candles,
//   - 2nd candle: black,
//   - 3rd candle: white, opens within 2nd black body, closes above 2nd
//     open but below 1st candle's real body bottom (inside the gap),
//   - 2nd and 3rd have near-equal body sizes.
//
// Category C: both branches evaluated, return stronger signal.
func TasukiGap(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(3, cp.Near) {
		return 0.0
	}

	b1 := cp.Bar(3)
	b2 := cp.Bar(2)
	b3 := cp.Bar(1)

	// Upside Tasuki Gap (bullish).
	bullSignal := 0.0
	if core.IsRealBodyGapUp(b1.O, b1.C, b2.O, b2.C) &&
		core.IsWhite(b2.O, b2.C) && core.IsBlack(b3.O, b3.C) {
		rb2 := core.RealBodyLen(b2.O, b2.C)
		rb3 := core.RealBodyLen(b3.O, b3.C)
		width := 0.0
		if rb2 > 0.0 {
			width = cp.FuzzRatio * rb2
		}
		// o3 within 2nd body: o3 < c2 and o3 > o2
		muO3LtC2 := cp.MuLtRaw(b3.O, b2.C, width)
		muO3GtO2 := cp.MuGtRaw(b3.O, b2.O, width)
		// c3 below o2
		muC3LtO2 := cp.MuLtRaw(b3.C, b2.O, width)
		// c3 above 1st body top (inside gap)
		body1Top := max(b1.C, b1.O)
		muC3GtTop1 := cp.MuGtRaw(b3.C, body1Top, width)
		// near-equal bodies
		muNear := cp.MuLess(math.Abs(rb2-rb3), cp.Near, 2)
		conf := fuzzy.TProductAll(muO3LtC2, muO3GtO2, muC3LtO2, muC3GtTop1, muNear)
		bullSignal = conf * 100.0
	}

	// Downside Tasuki Gap (bearish).
	bearSignal := 0.0
	if core.IsRealBodyGapDown(b1.O, b1.C, b2.O, b2.C) &&
		core.IsBlack(b2.O, b2.C) && core.IsWhite(b3.O, b3.C) {
		rb2 := core.RealBodyLen(b2.O, b2.C)
		rb3 := core.RealBodyLen(b3.O, b3.C)
		width := 0.0
		if rb2 > 0.0 {
			width = cp.FuzzRatio * rb2
		}
		// o3 within 2nd body: o3 < o2 and o3 > c2
		muO3LtO2 := cp.MuLtRaw(b3.O, b2.O, width)
		muO3GtC2 := cp.MuGtRaw(b3.O, b2.C, width)
		// c3 above o2
		muC3GtO2 := cp.MuGtRaw(b3.C, b2.O, width)
		// c3 below 1st body bottom (inside gap)
		body1Bot := min(b1.C, b1.O)
		muC3LtBot1 := cp.MuLtRaw(b3.C, body1Bot, width)
		// near-equal bodies
		muNear := cp.MuLess(math.Abs(rb2-rb3), cp.Near, 2)
		conf := fuzzy.TProductAll(muO3LtO2, muO3GtC2, muC3GtO2, muC3LtBot1, muNear)
		bearSignal = -conf * 100.0
	}

	if math.Abs(bullSignal) >= math.Abs(bearSignal) {
		return bullSignal
	}
	return bearSignal
}
