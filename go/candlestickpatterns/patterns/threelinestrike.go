package patterns

import (
	"math"

	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// ThreeLineStrike: a four-candle pattern.
//
// Bullish: three white candles with rising closes, each opening within/near
// the prior body, 4th black opens above 3rd close and closes below 1st open.
//
// Bearish: three black candles with falling closes, each opening within/near
// the prior body, 4th white opens below 3rd close and closes above 1st open.
//
// Category C: both branches evaluated, return stronger signal.
func ThreeLineStrike(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(4, cp.Near) {
		return 0.0
	}

	b1 := cp.Bar(4)
	b2 := cp.Bar(3)
	b3 := cp.Bar(2)
	b4 := cp.Bar(1)

	// Three same color — crisp gate.
	color1 := 1
	if !core.IsWhite(b1.O, b1.C) {
		color1 = -1
	}
	color2 := 1
	if !core.IsWhite(b2.O, b2.C) {
		color2 = -1
	}
	color3 := 1
	if !core.IsWhite(b3.O, b3.C) {
		color3 = -1
	}
	color4 := 1
	if !core.IsWhite(b4.O, b4.C) {
		color4 = -1
	}

	if !(color1 == color2 && color2 == color3 && color4 == -color3) {
		return 0.0
	}

	// 2nd opens within/near 1st real body — fuzzy.
	near4 := cp.AvgCS(cp.Near, 4)
	near3 := cp.AvgCS(cp.Near, 3)
	nearWidth4 := 0.0
	if near4 > 0.0 {
		nearWidth4 = cp.FuzzRatio * near4
	}
	nearWidth3 := 0.0
	if near3 > 0.0 {
		nearWidth3 = cp.FuzzRatio * near3
	}

	muO2Ge := cp.MuGeRaw(b2.O, min(b1.O, b1.C)-near4, nearWidth4)
	muO2Le := cp.MuLtRaw(b2.O, max(b1.O, b1.C)+near4, nearWidth4)

	// 3rd opens within/near 2nd real body — fuzzy.
	muO3Ge := cp.MuGeRaw(b3.O, min(b2.O, b2.C)-near3, nearWidth3)
	muO3Le := cp.MuLtRaw(b3.O, max(b2.O, b2.C)+near3, nearWidth3)

	// Bullish: three white, rising closes, 4th opens above 3rd close, closes below 1st open.
	bullSignal := 0.0
	if color3 == 1 && b3.C > b2.C && b2.C > b1.C {
		rb1 := math.Abs(b1.C - b1.O)
		width := 0.0
		if rb1 > 0.0 {
			width = cp.FuzzRatio * rb1
		}
		muO4Above := cp.MuGtRaw(b4.O, b3.C, width)
		muC4Below := cp.MuLtRaw(b4.C, b1.O, width)
		conf := fuzzy.TProductAll(muO2Ge, muO2Le, muO3Ge, muO3Le, muO4Above, muC4Below)
		bullSignal = conf * 100.0
	}

	// Bearish: three black, falling closes, 4th opens below 3rd close, closes above 1st open.
	bearSignal := 0.0
	if color3 == -1 && b3.C < b2.C && b2.C < b1.C {
		rb1 := math.Abs(b1.C - b1.O)
		width := 0.0
		if rb1 > 0.0 {
			width = cp.FuzzRatio * rb1
		}
		muO4Below := cp.MuLtRaw(b4.O, b3.C, width)
		muC4Above := cp.MuGtRaw(b4.C, b1.O, width)
		conf := fuzzy.TProductAll(muO2Ge, muO2Le, muO3Ge, muO3Le, muO4Below, muC4Above)
		bearSignal = -conf * 100.0
	}

	if math.Abs(bullSignal) >= math.Abs(bearSignal) {
		return bullSignal
	}
	return bearSignal
}
