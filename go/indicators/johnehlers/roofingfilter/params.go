package roofingfilter

import (
	"fmt"

	"zpano/entities"
	"zpano/indicators/core"
)

// Params describes parameters to create an instance of the RoofingFilter indicator.
//
// Explicit zero values are NOT substituted with defaults; use [DefaultParams]
// to obtain Ehlers' TASC settings.
type Params struct {
	// ShortestCyclePeriod is the shortest (SuperSmoother low-pass) cycle period in bars (>= 2). Default: 10.
	ShortestCyclePeriod float64

	// LongestCyclePeriod is the longest (high-pass) cycle period in bars (> ShortestCyclePeriod). Default: 48.
	LongestCyclePeriod float64

	// HighpassPoles is the number of poles in the initial high-pass stage (1 or 2). Default: 2.
	HighpassPoles int

	// ZeroMean enables a post-smoothing 1-pole high-pass stage (valid only when HighpassPoles == 1). Default: false.
	ZeroMean bool

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

// DefaultParams returns a [Params] value populated with Ehlers' January 2014 article defaults.
func DefaultParams() *Params {
	return &Params{
		ShortestCyclePeriod: 10,
		LongestCyclePeriod:  48,
		HighpassPoles:       2,
	}
}

func formatNumber(value float64) string {
	if value == 0 {
		value = 0
	}
	return fmt.Sprintf("%g", value)
}

// Mnemonic returns the indicator mnemonic for p.
func (p Params) Mnemonic() string {
	barComp := p.BarComponent
	if barComp == 0 {
		barComp = entities.DefaultBarComponent
	}
	quoteComp := p.QuoteComponent
	if quoteComp == 0 {
		quoteComp = entities.DefaultQuoteComponent
	}
	tradeComp := p.TradeComponent
	if tradeComp == 0 {
		tradeComp = entities.DefaultTradeComponent
	}
	zmTag := ""
	if p.ZeroMean {
		zmTag = "zm"
	}
	suffix := core.ComponentTripleMnemonic(barComp, quoteComp, tradeComp)
	return fmt.Sprintf(
		"roof%dhp%s(%s, %s%s)",
		p.HighpassPoles,
		zmTag,
		formatNumber(p.ShortestCyclePeriod),
		formatNumber(p.LongestCyclePeriod),
		suffix,
	)
}
