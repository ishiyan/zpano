package patterns

import (
	"zpano/candlestickpatterns/core"
)

// Doji detects a doji: open quite equal to close.
//
// Output is positive but this does not mean it is bullish:
// doji shows uncertainty and is neither bullish nor bearish when
// considered alone.
//
// The meaning of "doji" is specified with DojiBody.
func Doji(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(1, cp.DojiBody) {
		return 0.0
	}
	b := cp.Bar(1)
	// Fuzzy: degree to which real_body <= doji_avg.
	confidence := cp.MuLess(core.RealBodyLen(b.O, b.C), cp.DojiBody, 1)
	return confidence * 100.0
}
