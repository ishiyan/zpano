package fractaladaptivemovingaverage

import (
	"fmt"

	"zpano/entities"
	"zpano/indicators/core"
)

// Params describes parameters to create an instance of the FractalAdaptiveMovingAverage indicator.
//
// Explicit zero values are NOT substituted with defaults (`CoefficientMode: 0`
// selects Ehlers' literal 2005 `-4.6` coefficient); use [DefaultParams] to
// obtain the full-precision default settings.
type Params struct {
	// Length is the even lookback length (>= 2). Default: 16.
	Length int

	// SlowestSmoothingFactor is the slowest boundary smoothing factor (finite, 0 < A <= 1). Default: 0.01.
	SlowestSmoothingFactor float64

	// CoefficientMode selects the exponent scale: 1 for full-precision ln(SlowestSmoothingFactor),
	// 0 for the literal 2005 `-4.6` constant (valid only when SlowestSmoothingFactor == 0.01). Default: 1.
	CoefficientMode int

	// BarComponent indicates the component of a bar to use when updating with a bar sample.
	// If zero, DefaultBarComponent (Close) is used and omitted from the mnemonic.
	BarComponent entities.BarComponent

	// QuoteComponent indicates the component of a quote to use when updating with a quote sample.
	// If zero, DefaultQuoteComponent (Mid) is used and omitted from the mnemonic.
	QuoteComponent entities.QuoteComponent

	// TradeComponent indicates the component of a trade to use when updating with a trade sample.
	// If zero, DefaultTradeComponent (Price) is used and omitted from the mnemonic.
	TradeComponent entities.TradeComponent
}

// DefaultParams returns a [Params] value populated with Ehlers defaults.
func DefaultParams() *Params {
	return &Params{
		Length:                 16,
		SlowestSmoothingFactor: 0.01,
		CoefficientMode:        1,
	}
}

func formatNumber(value float64) string {
	if value == 0 {
		value = 0
	}
	return fmt.Sprintf("%g", value)
}

func formatMnemonic(prefix string, params Params) string {
	barComp := params.BarComponent
	if barComp == 0 {
		barComp = entities.DefaultBarComponent
	}
	quoteComp := params.QuoteComponent
	if quoteComp == 0 {
		quoteComp = entities.DefaultQuoteComponent
	}
	tradeComp := params.TradeComponent
	if tradeComp == 0 {
		tradeComp = entities.DefaultTradeComponent
	}
	modeTag := ""
	if params.CoefficientMode == 0 {
		modeTag = ", mode=0"
	}
	suffix := core.ComponentTripleMnemonic(barComp, quoteComp, tradeComp)
	return fmt.Sprintf(
		"%s(%d, %s%s%s)",
		prefix,
		params.Length,
		formatNumber(params.SlowestSmoothingFactor),
		modeTag,
		suffix,
	)
}

// Mnemonic returns the primary indicator mnemonic for p.
func (p Params) Mnemonic() string {
	return formatMnemonic("frama", p)
}

// DimensionMnemonic returns the fractal-dimension output mnemonic for p.
func (p Params) DimensionMnemonic() string {
	return formatMnemonic("framaDim", p)
}
