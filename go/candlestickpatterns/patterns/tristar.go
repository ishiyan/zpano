package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// Tristar: a three-candle reversal pattern with three dojis.
//
// Must have:
//   - three consecutive doji candles,
//   - if the second doji gaps up from the first and the third does not
//     close higher than the second: bearish,
//   - if the second doji gaps down from the first and the third does not
//     close lower than the second: bullish.
//
// Category A: fixed direction per branch (bullish or bearish).
func Tristar(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(3, cp.DojiBody) {
		return 0.0
	}

	b1 := cp.Bar(3)
	b2 := cp.Bar(2)
	b3 := cp.Bar(1)

	// Fuzzy: all three must be dojis.
	muDoji1 := cp.MuLess(core.RealBodyLen(b1.O, b1.C), cp.DojiBody, 3)
	muDoji2 := cp.MuLess(core.RealBodyLen(b2.O, b2.C), cp.DojiBody, 2)
	muDoji3 := cp.MuLess(core.RealBodyLen(b3.O, b3.C), cp.DojiBody, 1)

	// Bearish: second gaps up, third is not higher than second — crisp direction checks.
	if core.IsRealBodyGapUp(b1.O, b1.C, b2.O, b2.C) &&
		max(b3.O, b3.C) < max(b2.O, b2.C) {
		conf := fuzzy.TProductAll(muDoji1, muDoji2, muDoji3)
		return -conf * 100.0
	}

	// Bullish: second gaps down, third is not lower than second.
	if core.IsRealBodyGapDown(b1.O, b1.C, b2.O, b2.C) &&
		min(b3.O, b3.C) > min(b2.O, b2.C) {
		conf := fuzzy.TProductAll(muDoji1, muDoji2, muDoji3)
		return conf * 100.0
	}

	return 0.0
}
