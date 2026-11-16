package autocorrelationperiodogram

import (
	"encoding/json"
	"fmt"

	"zpano/entities"
	"zpano/indicators/core"
)

// Source selects the Ehlers source profile for AutoCorrelationPeriodogram.
type Source int

const (
	// Article selects the TASC September 2016 "Measuring Market Cycles" sidebar profile (default).
	Article Source = iota
	// Book selects the corrected Cycle Analytics for Traders (2013), Listing 8-3 profile.
	Book
	sourceLast
)

const (
	articleSourceStr = "article"
	bookSourceStr    = "book"
	unknownSourceStr = "unknown"
)

// String implements the Stringer interface.
func (s Source) String() string {
	switch s {
	case Article:
		return articleSourceStr
	case Book:
		return bookSourceStr
	default:
		return unknownSourceStr
	}
}

// IsKnown reports whether s is a supported Source variant.
func (s Source) IsKnown() bool {
	return s >= Article && s < sourceLast
}

// MarshalJSON implements json.Marshaler.
func (s Source) MarshalJSON() ([]byte, error) {
	text := s.String()
	if text == unknownSourceStr {
		return nil, fmt.Errorf("cannot marshal %q: unknown autocorrelation periodogram source", text)
	}
	return json.Marshal(text)
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *Source) UnmarshalJSON(data []byte) error {
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return fmt.Errorf("cannot unmarshal %s: source must be a JSON string", string(data))
	}
	switch text {
	case articleSourceStr:
		*s = Article
		return nil
	case bookSourceStr:
		*s = Book
		return nil
	default:
		return fmt.Errorf("cannot unmarshal %q: unknown autocorrelation periodogram source", text)
	}
}

// Params describes parameters to create an instance of AutoCorrelationPeriodogram.
//
// MinPeriod == 0 selects the profile default (8 for Article, 10 for Book).
// AveragingLength == 0 selects lag-dependent Pearson windows (M = lag); use
// [DefaultParams] to obtain the default 3-sample window and MaxPeriod = 48.
type Params struct {
	// Source selects the 2016 article (default) or 2013 corrected book profile.
	Source Source `json:"source"`

	// MinPeriod is the shortest spectrum bin and SuperSmoother cutoff in bars (>= 3).
	// Zero selects the profile default: 8 for Article or 10 for Book.
	MinPeriod int `json:"minPeriod"`

	// MaxPeriod is the longest spectrum bin, high-pass cutoff, and largest lag (>= 5, > MinPeriod). Default: 48.
	MaxPeriod int `json:"maxPeriod"`

	// AveragingLength is the Pearson window length: 0 for lag-dependent (M = lag) or >= 2. Default: 3.
	AveragingLength int `json:"averagingLength"`

	// EnhanceResolution cubes normalized power before centroid in Article mode or squares display power after centroid in Book mode. Default: false.
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

// DefaultParams returns a [Params] value populated with Ehlers' 2016 article defaults.
//
// MinPeriod is left at 0 so that overriding Source to [Book] automatically resolves MinPeriod to 10.
func DefaultParams() *Params {
	return &Params{
		Source:            Article,
		MinPeriod:         0,
		MaxPeriod:         48,
		AveragingLength:   3,
		EnhanceResolution: false,
	}
}

func resolveMinPeriod(source Source, minPeriod int) int {
	if minPeriod != 0 {
		return minPeriod
	}
	if source == Book {
		return 10
	}
	return 8
}

func formatMnemonic(prefix string, includeBookEnhance bool, params Params) string {
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
	resolvedMin := resolveMinPeriod(params.Source, params.MinPeriod)
	sourceTag := ""
	if params.Source == Book {
		sourceTag = ", book"
	}
	enhTag := ""
	if params.EnhanceResolution && (params.Source != Book || includeBookEnhance) {
		enhTag = ", enh"
	}
	return fmt.Sprintf("%s(%d, %d, %d%s%s%s)",
		prefix,
		resolvedMin,
		params.MaxPeriod,
		params.AveragingLength,
		sourceTag,
		enhTag,
		compSuffix,
	)
}

// Mnemonic returns the primary heatmap mnemonic for p.
func (p Params) Mnemonic() string {
	return formatMnemonic("acp", true, p)
}

// DominantCycleMnemonic returns the dominant-cycle mnemonic for p.
func (p Params) DominantCycleMnemonic() string {
	return formatMnemonic("acp-dc", false, p)
}
