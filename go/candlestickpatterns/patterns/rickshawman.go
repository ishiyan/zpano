package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// RickshawMan detects a Rickshaw Man pattern: a one-candle doji pattern.
//
// Must have:
//   - doji body (very small real body),
//   - two long shadows,
//   - body near the midpoint of the high-low range.
func RickshawMan(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(1, cp.DojiBody, cp.LongShadow, cp.Near) {
		return 0.0
	}

	b := cp.Bar(1)

	hlRange := b.H - b.L
	nearAvg := cp.AvgCS(cp.Near, 1)
	nearWidth := 0.0
	if nearAvg > 0.0 {
		nearWidth = cp.FuzzRatio * nearAvg
	}

	muDoji := cp.MuLess(core.RealBodyLen(b.O, b.C), cp.DojiBody, 1)
	muLongUS := cp.MuGreater(core.UpperShadow(b.O, b.H, b.C), cp.LongShadow, 1)
	muLongLS := cp.MuGreater(core.LowerShadow(b.O, b.L, b.C), cp.LongShadow, 1)
	midpoint := b.L + hlRange/2.0
	muNearMidLo := cp.MuLtRaw(min(b.O, b.C), midpoint+nearAvg, nearWidth)
	muNearMidHi := cp.MuGeRaw(max(b.O, b.C), midpoint-nearAvg, nearWidth)

	confidence := fuzzy.TProductAll(muDoji, muLongUS, muLongLS, muNearMidLo, muNearMidHi)
	return confidence * 100.0
}
