package transformeddiscretefouriertransformspectrum

import (
	"fmt"

	"zpano/entities"
	"zpano/indicators/core"
)

// Params describes parameters to create an instance of Ehlers' TransformedDiscreteFourierTransformSpectrum indicator.
//
// Explicit zero values are NOT substituted with defaults; use [DefaultParams]
// to obtain Ehlers' January 2007 article settings.
type Params struct {
	// MinPeriod is the shortest evaluated integer cycle period in bars (>= 8). Default: 8.
	MinPeriod int `json:"minPeriod"`

	// MaxPeriod is the longest evaluated integer cycle period in bars (>= MinPeriod). Default: 50.
	MaxPeriod int `json:"maxPeriod"`

	// WindowLength is the moving observation window length in cleaned samples (>= 1). Default: 50.
	WindowLength int `json:"windowLength"`

	// HighpassPeriod is the high-pass filter cutoff period in bars (> 4.0). Default: 40.
	HighpassPeriod float64 `json:"highpassPeriod"`

	// BarComponent indicates the component of a bar to use when updating with a bar sample.
	// If zero, DefaultBarComponent (Close) is used and omitted from the mnemonic.
	BarComponent entities.BarComponent `json:"barComponent"`

	// QuoteComponent indicates the component of a quote to use when updating with a quote sample.
	// If zero, DefaultQuoteComponent (Mid) is used and omitted from the mnemonic.
	QuoteComponent entities.QuoteComponent `json:"quoteComponent"`

	// TradeComponent indicates the component of a trade to use when updating with a trade sample.
	// If zero, DefaultTradeComponent (Price) is used and omitted from the mnemonic.
	TradeComponent entities.TradeComponent `json:"tradeComponent"`
}

// DefaultParams returns a [Params] value populated with Ehlers' January 2007 article defaults.
func DefaultParams() *Params {
	return &Params{
		MinPeriod:      8,
		MaxPeriod:      50,
		WindowLength:   50,
		HighpassPeriod: 40,
	}
}

// FormatNumber formats a float parameter cleanly without negative zero.
func FormatNumber(value float64) string {
	if value == 0 {
		value = 0
	}
	return fmt.Sprintf("%g", value)
}

func formatMnemonic(prefix string, params Params) string {
	barComp, quoteComp, tradeComp := params.BarComponent, params.QuoteComponent, params.TradeComponent
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
	return fmt.Sprintf("%s(%d, %d, %d, %s%s)",
		prefix,
		params.MinPeriod,
		params.MaxPeriod,
		params.WindowLength,
		FormatNumber(params.HighpassPeriod),
		compSuffix,
	)
}

// Mnemonic returns the primary heatmap mnemonic for p.
func (p Params) Mnemonic() string {
	return formatMnemonic("tdfts", p)
}

// DominantCycleMnemonic returns the dominant-cycle mnemonic for p.
func (p Params) DominantCycleMnemonic() string {
	return formatMnemonic("tdfts-dc", p)
}
