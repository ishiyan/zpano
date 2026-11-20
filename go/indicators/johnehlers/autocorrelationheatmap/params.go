package autocorrelationheatmap

import (
	"fmt"

	"zpano/entities"
	"zpano/indicators/core"
)

// Params describes parameters to create an instance of Ehlers' AutoCorrelationHeatmap indicator.
//
// Explicit zero values are NOT substituted with defaults because MaxLag == 0 is a valid
// single-bin zero-lag configuration; use [DefaultParams] to obtain Ehlers' February 2025
// article settings.
type Params struct {
	// Length is the number of aligned sample pairs per Pearson correlation comparison (>= 2). Default: 20.
	Length int

	// SmoothingPeriod is the Ultimate Smoother critical period in bars (>= 1.0). Default: 20.
	SmoothingPeriod float64

	// MaxLag is the inclusive upper correlation lag in [0, 99]; minimum lag is always zero. Default: 99.
	MaxLag int

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

// DefaultParams returns a [Params] value populated with Ehlers' February 2025 article defaults.
func DefaultParams() *Params {
	return &Params{
		Length:          20,
		SmoothingPeriod: 20.0,
		MaxLag:          99,
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
	return fmt.Sprintf("ach(%d, %s, %d%s)",
		p.Length,
		FormatNumber(p.SmoothingPeriod),
		p.MaxLag,
		compSuffix,
	)
}
