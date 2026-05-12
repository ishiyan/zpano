package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// SpinningTop detects a Spinning Top pattern: a one-candle pattern.
//
// A candle with a small body and shadows longer than the body on both sides.
//
// The meaning of "short" is specified with ShortBody.
//
// Category C: color determines sign.
func SpinningTop(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(1, cp.ShortBody) {
		return 0.0
	}

	b := cp.Bar(1)

	rb := core.RealBodyLen(b.O, b.C)

	muShort := cp.MuLess(rb, cp.ShortBody, 1)

	// Shadows > body: positional comparisons.
	us := core.UpperShadow(b.O, b.H, b.C)
	ls := core.LowerShadow(b.O, b.L, b.C)
	widthUS := 0.0
	if rb > 0.0 {
		widthUS = cp.FuzzRatio * rb
	}
	widthLS := 0.0
	if rb > 0.0 {
		widthLS = cp.FuzzRatio * rb
	}
	muUSGtRB := cp.MuGtRaw(us, rb, widthUS)
	muLSGtRB := cp.MuGtRaw(ls, rb, widthLS)

	confidence := fuzzy.TProductAll(muShort, muUSGtRB, muLSGtRB)

	if core.IsWhite(b.O, b.C) {
		return confidence * 100.0
	}
	if core.IsBlack(b.O, b.C) {
		return -confidence * 100.0
	}
	return 0.0
}
