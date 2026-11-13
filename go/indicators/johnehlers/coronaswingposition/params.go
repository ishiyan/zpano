package coronaswingposition

import (
	"fmt"
	"strings"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/johnehlers/corona"
)

// Params describes parameters to create an instance of the CoronaSwingPosition indicator.
//
// Explicit zero values are NOT substituted with defaults; use [DefaultParams]
// to obtain Ehlers' TASC settings.
type Params struct {
	// RasterLength is the number of elements in the heatmap raster (>= 2). Default: 50.
	RasterLength int

	// MaxRasterValue is the maximal raster value (z) of the heatmap (> 6). Default: 20.
	MaxRasterValue float64

	// MinParameterValue is the minimal ordinate (y) value of the heatmap. Default: -5.
	MinParameterValue float64

	// MaxParameterValue is the maximal ordinate (y) value of the heatmap (> MinParameterValue). Default: 5.
	MaxParameterValue float64

	// HighPassFilterCutoff is the high-pass filter cutoff in bars (>= 5). Default: 30.
	HighPassFilterCutoff int

	// MinimalPeriod is the minimal period of the inner Corona engine (>= 6). Default: 6.
	MinimalPeriod int

	// MaximalPeriod is the maximal period of the inner Corona engine (> MinimalPeriod). Default: 30.
	MaximalPeriod int

	// LeadWindow is the rolling window length for 60-degree lead normalization (>= 2). Default: 51.
	LeadWindow int

	// PositionWindow is the rolling window length for position span calculation (>= 2). Default: 21.
	PositionWindow int

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
		RasterLength:         50,
		MaxRasterValue:       20,
		MinParameterValue:    -5,
		MaxParameterValue:    5,
		HighPassFilterCutoff: 30,
		MinimalPeriod:        6,
		MaximalPeriod:        30,
		LeadWindow:           51,
		PositionWindow:       21,
	}
}

func swingFlags(p Params) string {
	var builder strings.Builder
	if p.MinimalPeriod != 6 {
		fmt.Fprintf(&builder, ", pmin=%d", p.MinimalPeriod)
	}
	if p.MaximalPeriod != 30 {
		fmt.Fprintf(&builder, ", pmax=%d", p.MaximalPeriod)
	}
	if p.LeadWindow != 51 {
		fmt.Fprintf(&builder, ", lead=%d", p.LeadWindow)
	}
	if p.PositionWindow != 21 {
		fmt.Fprintf(&builder, ", pos=%d", p.PositionWindow)
	}
	return builder.String()
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
	flags := swingFlags(p)
	return fmt.Sprintf("cswing(%d, %s, %s, %s, %d%s%s)",
		p.RasterLength,
		corona.FormatNumber(p.MaxRasterValue),
		corona.FormatNumber(p.MinParameterValue),
		corona.FormatNumber(p.MaxParameterValue),
		p.HighPassFilterCutoff,
		flags,
		comp,
	)
}
