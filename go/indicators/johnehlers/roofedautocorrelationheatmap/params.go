package roofedautocorrelationheatmap

import (
	"fmt"

	"zpano/entities"
	"zpano/indicators/core"
)

// Params describes parameters to create an instance of Ehlers' RoofedAutoCorrelationHeatmap indicator.
//
// Explicit zero values are NOT substituted with defaults because MinLag == 0, MaxLag == 0,
// and AveragingLength == 0 are valid configurations; use [DefaultParams] to obtain Ehlers'
// 2013 Book Listing 8-2 settings.
type Params struct {
	// MinLag is the first displayed correlation lag on the heatmap axis (>= 0). Default: 3.
	MinLag int

	// MaxLag is the last displayed correlation lag on the heatmap axis (>= MinLag). Default: 48.
	MaxLag int

	// AveragingLength is the number of sample pairs per Pearson correlation comparison (>= 0; 0 selects M = lag). Default: 0.
	AveragingLength int

	// HighpassPeriod is the two-pole roofing high-pass cutoff period in bars (>= 5.0). Default: 48.0.
	HighpassPeriod float64

	// SmoothingPeriod is the two-pole SuperSmoother cutoff period in bars (>= 2.0). Default: 10.0.
	SmoothingPeriod float64

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

// DefaultParams returns a [Params] value populated with Ehlers' 2013 Book Listing 8-2 defaults.
func DefaultParams() *Params {
	return &Params{
		MinLag:          3,
		MaxLag:          48,
		AveragingLength: 0,
		HighpassPeriod:  48.0,
		SmoothingPeriod: 10.0,
	}
}

// FormatNumber formats a float parameter cleanly without negative zero.
func FormatNumber(value float64) string {
	if value == 0 {
		value = 0
	}
	return fmt.Sprintf("%g", value)
}

// Mnemonic returns the indicator mnemonic for p.
func (p Params) Mnemonic() string {
	barComp, quoteComp, tradeComp := p.BarComponent, p.QuoteComponent, p.TradeComponent
	if barComp == 0 {
		barComp = entities.DefaultBarComponent
	}
	if quoteComp == 0 {
		quoteComp = entities.DefaultQuoteComponent
	}
	if tradeComp == 0 {
		tradeComp = entities.DefaultTradeComponent
	}
	compSuffix := core.ComponentTripleMnemonic(barComp, quoteComp, tradeComp)
	return fmt.Sprintf("rach(%d, %d, %d, %s, %s%s)",
		p.MinLag,
		p.MaxLag,
		p.AveragingLength,
		FormatNumber(p.HighpassPeriod),
		FormatNumber(p.SmoothingPeriod),
		compSuffix,
	)
}
