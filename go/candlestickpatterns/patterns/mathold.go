package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

const matHoldPenetrationFactor = 0.5

// MatHold detects the Mat Hold pattern (5-candle bullish continuation).
//
// Must have:
//   - first candle: long white,
//   - second candle: small, black, gaps up from first,
//   - third and fourth candles: small,
//   - reaction candles (2-4) are falling, hold within first body
//     (penetration check),
//   - fifth candle: white, opens above prior close, closes above
//     highest high of reaction candles.
//
// Category A: always bullish (continuous).
func MatHold(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(5, cp.LongBody, cp.ShortBody) {
		return 0.0
	}

	b1 := cp.Bar(5)
	b2 := cp.Bar(4)
	b3 := cp.Bar(3)
	b4 := cp.Bar(2)
	b5 := cp.Bar(1)

	// Crisp gates: colors.
	if !(core.IsWhite(b1.O, b1.C) && core.IsBlack(b2.O, b2.C) && core.IsWhite(b5.O, b5.C)) {
		return 0.0
	}
	// Crisp: gap up from 1st to 2nd.
	if !core.IsRealBodyGapUp(b1.O, b1.C, b2.O, b2.C) {
		return 0.0
	}
	// Crisp: 3rd to 4th hold within 1st range.
	if !(min(b3.O, b3.C) < b1.C && min(b4.O, b4.C) < b1.C) {
		return 0.0
	}
	// Crisp: reaction days don't penetrate first body too much.
	rb1 := core.RealBodyLen(b1.O, b1.C)
	if !(min(b3.O, b3.C) > b1.C-rb1*matHoldPenetrationFactor &&
		min(b4.O, b4.C) > b1.C-rb1*matHoldPenetrationFactor) {
		return 0.0
	}
	// Crisp: 2nd to 4th are falling.
	if !(max(b3.O, b3.C) < b2.O && max(b4.O, b4.C) < max(b3.O, b3.C)) {
		return 0.0
	}
	// Crisp: 5th opens above prior close.
	if !(b5.O > b4.C) {
		return 0.0
	}
	// Crisp: 5th closes above highest high of reaction candles.
	if !(b5.C > max(b2.H, max(b3.H, b4.H))) {
		return 0.0
	}

	// Fuzzy: first candle long.
	muLong1 := cp.MuGreater(rb1, cp.LongBody, 5)
	// Fuzzy: 2nd, 3rd, 4th short.
	muShort2 := cp.MuLess(core.RealBodyLen(b2.O, b2.C), cp.ShortBody, 4)
	muShort3 := cp.MuLess(core.RealBodyLen(b3.O, b3.C), cp.ShortBody, 3)
	muShort4 := cp.MuLess(core.RealBodyLen(b4.O, b4.C), cp.ShortBody, 2)

	confidence := fuzzy.TProductAll(muLong1, muShort2, muShort3, muShort4)
	return confidence * 100.0
}
