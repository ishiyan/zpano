package coronaspectrum

import (
	"fmt"
	"math"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/johnehlers/corona"
)

// Params describes parameters to create an instance of the CoronaSpectrum indicator.
//
// Explicit zero values are NOT substituted with defaults; use [DefaultParams]
// to obtain Ehlers' November 2008 TASC article settings.
type Params struct {
	// MinRasterValue is the selection threshold in decibels (0 <= MinRasterValue < MaxRasterValue). Default: 6.
	MinRasterValue float64

	// MaxRasterValue is the intensity ceiling in decibels (> 0 and > MinRasterValue). Default: 20.
	MaxRasterValue float64

	// MinParameterValue is the minimal ordinate (y) value of the heatmap, rounded up to integer >= 6. Default: 6.
	MinParameterValue float64

	// MaxParameterValue is the maximal ordinate (y) value of the heatmap, rounded down to integer > minPeriod. Default: 30.
	MaxParameterValue float64

	// HighPassFilterCutoff is the high-pass filter cutoff (de-trending period) in bars (>= 5). Default: 30.
	HighPassFilterCutoff int

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
		MinRasterValue:       6,
		MaxRasterValue:       20,
		MinParameterValue:    6,
		MaxParameterValue:    30,
		HighPassFilterCutoff: 30,
	}
}

// Mnemonic returns the primary indicator mnemonic for p.
func (p Params) Mnemonic() string {
	bc, qc, tc := p.BarComponent, p.QuoteComponent, p.TradeComponent
	if bc == 0 {
		bc = entities.DefaultBarComponent
	}
	if qc == 0 {
		qc = entities.DefaultQuoteComponent
	}
	if tc == 0 {
		tc = entities.DefaultTradeComponent
	}
	comp := core.ComponentTripleMnemonic(bc, qc, tc)
	minPeriod := int(math.Ceil(p.MinParameterValue))
	maxPeriod := int(math.Floor(p.MaxParameterValue))
	return fmt.Sprintf("cspect(%s, %s, %d, %d, %d%s)",
		corona.FormatNumber(p.MinRasterValue),
		corona.FormatNumber(p.MaxRasterValue),
		minPeriod,
		maxPeriod,
		p.HighPassFilterCutoff,
		comp,
	)
}
