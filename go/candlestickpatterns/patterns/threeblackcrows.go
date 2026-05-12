package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// ThreeBlackCrows: a four-candle bearish reversal pattern.
//
// Must have:
//   - preceding candle (oldest) is white,
//   - three consecutive black candles with declining closes,
//   - each opens within the prior black candle's real body,
//   - each has a very short lower shadow,
//   - 1st black closes under the prior white candle's high.
//
// Category A: always bearish (continuous).
func ThreeBlackCrows(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(4, cp.VeryShortShadow) {
		return 0.0
	}

	b0 := cp.Bar(4) // prior white
	b1 := cp.Bar(3) // 1st black
	b2 := cp.Bar(2) // 2nd black
	b3 := cp.Bar(1) // 3rd black

	// Crisp gates: colors, declining closes, opens within prior body.
	if !core.IsWhite(b0.O, b0.C) {
		return 0.0
	}
	if !(core.IsBlack(b1.O, b1.C) && core.IsBlack(b2.O, b2.C) && core.IsBlack(b3.O, b3.C)) {
		return 0.0
	}
	if !(b1.C > b2.C && b2.C > b3.C) {
		return 0.0
	}
	// Opens within prior black body (crisp containment for strict ordering).
	if !(b2.O < b1.O && b2.O > b1.C && b3.O < b2.O && b3.O > b2.C) {
		return 0.0
	}
	// Prior white's high > 1st black's close (crisp).
	if !(b0.H > b1.C) {
		return 0.0
	}

	// Fuzzy: very short lower shadows.
	muLS1 := cp.MuLess(core.LowerShadow(b1.O, b1.L, b1.C), cp.VeryShortShadow, 3)
	muLS2 := cp.MuLess(core.LowerShadow(b2.O, b2.L, b2.C), cp.VeryShortShadow, 2)
	muLS3 := cp.MuLess(core.LowerShadow(b3.O, b3.L, b3.C), cp.VeryShortShadow, 1)

	confidence := fuzzy.TProductAll(muLS1, muLS2, muLS3)

	return -1.0 * confidence * 100.0
}
