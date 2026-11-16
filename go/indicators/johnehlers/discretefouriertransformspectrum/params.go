package discretefouriertransformspectrum

import (
	"fmt"

	"zpano/entities"
	"zpano/indicators/core"
)

// Params describes parameters to create an instance of Ehlers' DiscreteFourierTransformSpectrum indicator.
//
// Explicit zero/false values are NOT substituted with defaults; use [DefaultParams]
// to obtain Ehlers' 2013 Book Listing 9-1 settings.
type Params struct {
	// MinPeriod is the shortest evaluated integer cycle period in bars (>= 3). Default: 10.
	MinPeriod int `json:"minPeriod"`

	// MaxPeriod is the longest evaluated integer cycle period in bars (>= MinPeriod). Default: 48.
	MaxPeriod int `json:"maxPeriod"`

	// WindowLength is the sliding roofed DFT window length in bars (>= 2, independent of MaxPeriod). Default: 48.
	WindowLength int `json:"windowLength"`

	// HighpassPeriod is the two-pole roofing high-pass cutoff period in bars (>= 5.0). Default: 48.
	HighpassPeriod float64 `json:"highpassPeriod"`

	// SmoothingPeriod is the two-pole SuperSmoother cutoff period in bars (>= 2.0). Default: 10.
	SmoothingPeriod float64 `json:"smoothingPeriod"`

	// SpectralDilationCompensation divides each cosine/sine contribution by period before summing/squaring. Default: true.
	SpectralDilationCompensation bool `json:"spectralDilationCompensation"`

	// EnhanceResolution cubes normalized display heatmap intensities after centroid extraction. Default: false.
	EnhanceResolution bool `json:"enhanceResolution"`

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

// DefaultParams returns a [Params] value populated with Ehlers' Listing 9-1 defaults.
func DefaultParams() *Params {
	return &Params{
		MinPeriod:                    10,
		MaxPeriod:                    48,
		WindowLength:                 48,
		HighpassPeriod:               48,
		SmoothingPeriod:              10,
		SpectralDilationCompensation: true,
		EnhanceResolution:            false,
	}
}

// FormatNumber formats a float parameter cleanly without negative zero.
func FormatNumber(value float64) string {
	if value == 0 {
		value = 0
	}
	return fmt.Sprintf("%g", value)
}

func formatMnemonic(prefix string, includeEnhance bool, params Params) string {
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
	sdcSuffix := ""
	if !params.SpectralDilationCompensation {
		sdcSuffix = ", no-sdc"
	}
	enhSuffix := ""
	if includeEnhance && params.EnhanceResolution {
		enhSuffix = ", enh"
	}
	return fmt.Sprintf("%s(%d, %d, %d, %s, %s%s%s%s)",
		prefix,
		params.MinPeriod,
		params.MaxPeriod,
		params.WindowLength,
		FormatNumber(params.HighpassPeriod),
		FormatNumber(params.SmoothingPeriod),
		sdcSuffix,
		enhSuffix,
		compSuffix,
	)
}

// Mnemonic returns the primary heatmap mnemonic for p.
func (p Params) Mnemonic() string {
	return formatMnemonic("dftps", true, p)
}

// DominantCycleMnemonic returns the dominant-cycle mnemonic for p.
func (p Params) DominantCycleMnemonic() string {
	return formatMnemonic("dftps-dc", false, p)
}
