package patterns

import (
	"zpano/candlestickpatterns/core"
)

// LadderBottom detects the Ladder Bottom pattern (5-candle bullish).
//
// Must have:
//   - first three candles: descending black candles (each closes lower),
//   - fourth candle: black with a long upper shadow,
//   - fifth candle: white, opens above the fourth candle's real body,
//     closes above the fourth candle's high.
//
// The meaning of "long" for shadows is specified with cp.VeryShortShadow.
//
// Category A: always bullish (continuous).
func LadderBottom(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(5, cp.VeryShortShadow) {
		return 0.0
	}

	b1 := cp.Bar(5)
	b2 := cp.Bar(4)
	b3 := cp.Bar(3)
	b4 := cp.Bar(2)
	b5 := cp.Bar(1)

	// Crisp gates: colors.
	if !(core.IsBlack(b1.O, b1.C) && core.IsBlack(b2.O, b2.C) &&
		core.IsBlack(b3.O, b3.C) && core.IsBlack(b4.O, b4.C) &&
		core.IsWhite(b5.O, b5.C)) {
		return 0.0
	}
	// Crisp: three descending opens and closes.
	if !(b1.O > b2.O && b2.O > b3.O && b1.C > b2.C && b2.C > b3.C) {
		return 0.0
	}
	// Crisp: fifth opens above fourth's open, closes above fourth's high.
	if !(b5.O > b4.O && b5.C > b4.H) {
		return 0.0
	}

	// Fuzzy: fourth candle has upper shadow > very short avg.
	muUS4 := cp.MuGreater(core.UpperShadow(b4.O, b4.H, b4.C), cp.VeryShortShadow, 2)
	return muUS4 * 100.0
}
