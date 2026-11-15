package zerolagerrorcorrectingexponentialmovingaverage

import (
	"fmt"
	"math"

	"zpano/entities"
	"zpano/indicators/core"
)

const alphaEpsilon = 1e-8

// LengthParams describes length-based parameters to create an instance of
// ZeroLagErrorCorrectingExponentialMovingAverage.
//
// Explicit zero values are NOT substituted with defaults (`GainLimit: 0`
// selects zero gain); use [DefaultLengthParams] or [DefaultParams] to obtain
// the article's default settings (`Length: 20, GainLimit: 50`).
type LengthParams struct {
	// Length is the equivalent EMA length L (>= 1), defining alpha = 2 / (L + 1).
	//
	// The default value in the 2010 article is 20 (alpha = 2/21).
	Length int

	// GainLimit is the nonnegative integer gain bound in tenths.
	//
	// Candidate gains span [-GainLimit/10, +GainLimit/10] in increments of 0.1.
	// The default value in the 2010 article is 50 (actual gains -5.0..+5.0).
	GainLimit int

	// BarComponent indicates the component of a bar to use when updating the indicator with a bar sample.
	//
	// If zero, DefaultBarComponent (Close) is used and omitted from the indicator mnemonic.
	BarComponent entities.BarComponent

	// QuoteComponent indicates the component of a quote to use when updating the indicator with a quote sample.
	//
	// If zero, DefaultQuoteComponent (Mid) is used and omitted from the indicator mnemonic.
	QuoteComponent entities.QuoteComponent

	// TradeComponent indicates the component of a trade to use when updating the indicator with a trade sample.
	//
	// If zero, DefaultTradeComponent (Price) is used and omitted from the indicator mnemonic.
	TradeComponent entities.TradeComponent
}

// Params aliases the primary length-based parameter struct.
type Params = LengthParams

// SmoothingFactorParams describes smoothing-factor-based parameters to create an
// instance of ZeroLagErrorCorrectingExponentialMovingAverage.
//
// Explicit zero values are NOT substituted with defaults (`GainLimit: 0`
// selects zero gain); use [DefaultSmoothingFactorParams] to obtain Ehlers'
// default settings (`SmoothingFactor: 2/21, GainLimit: 50`).
type SmoothingFactorParams struct {
	// SmoothingFactor is the baseline EMA smoothing factor alpha in (0, 1].
	//
	// The equivalent length is round(2/alpha) - 1. The default value is 2/21.
	SmoothingFactor float64

	// GainLimit is the nonnegative integer gain bound in tenths.
	//
	// Candidate gains span [-GainLimit/10, +GainLimit/10] in increments of 0.1.
	// The default value in the 2010 article is 50 (actual gains -5.0..+5.0).
	GainLimit int

	// BarComponent indicates the component of a bar to use when updating the indicator with a bar sample.
	//
	// If zero, DefaultBarComponent (Close) is used and omitted from the indicator mnemonic.
	BarComponent entities.BarComponent

	// QuoteComponent indicates the component of a quote to use when updating the indicator with a quote sample.
	//
	// If zero, DefaultQuoteComponent (Mid) is used and omitted from the indicator mnemonic.
	QuoteComponent entities.QuoteComponent

	// TradeComponent indicates the component of a trade to use when updating the indicator with a trade sample.
	//
	// If zero, DefaultTradeComponent (Price) is used and omitted from the indicator mnemonic.
	TradeComponent entities.TradeComponent
}

// DefaultLengthParams returns a [LengthParams] value populated with Ehlers defaults.
func DefaultLengthParams() *LengthParams {
	return &LengthParams{
		Length:    20,
		GainLimit: 50,
	}
}

// DefaultParams returns a [Params] value populated with Ehlers defaults.
func DefaultParams() *Params {
	return DefaultLengthParams()
}

// DefaultSmoothingFactorParams returns a [SmoothingFactorParams] value populated with Ehlers defaults.
func DefaultSmoothingFactorParams() *SmoothingFactorParams {
	return &SmoothingFactorParams{
		SmoothingFactor: 2.0 / 21.0,
		GainLimit:       50,
	}
}

func equivalentLengthFromAlpha(alpha float64) int {
	if !(alpha >= alphaEpsilon) {
		return math.MaxInt
	}
	rounded := math.Round(2.0 / alpha)
	if rounded >= float64(math.MaxInt) {
		return math.MaxInt
	}
	return int(rounded) - 1
}

func resolveComponentSuffix(
	barComp entities.BarComponent,
	quoteComp entities.QuoteComponent,
	tradeComp entities.TradeComponent,
) string {
	if barComp == 0 {
		barComp = entities.DefaultBarComponent
	}
	if quoteComp == 0 {
		quoteComp = entities.DefaultQuoteComponent
	}
	if tradeComp == 0 {
		tradeComp = entities.DefaultTradeComponent
	}
	return core.ComponentTripleMnemonic(barComp, quoteComp, tradeComp)
}

// Mnemonic returns the primary EC output mnemonic for p.
func (p LengthParams) Mnemonic() string {
	suffix := resolveComponentSuffix(p.BarComponent, p.QuoteComponent, p.TradeComponent)
	return fmt.Sprintf("zecema(%d, %d%s)", p.Length, p.GainLimit, suffix)
}

// EmaMnemonic returns the baseline EMA output mnemonic for p.
func (p LengthParams) EmaMnemonic() string {
	suffix := resolveComponentSuffix(p.BarComponent, p.QuoteComponent, p.TradeComponent)
	return fmt.Sprintf("zecemaEma(%d, %d%s)", p.Length, p.GainLimit, suffix)
}

// Mnemonic returns the primary EC output mnemonic for p.
func (p SmoothingFactorParams) Mnemonic() string {
	equivLength := equivalentLengthFromAlpha(p.SmoothingFactor)
	suffix := resolveComponentSuffix(p.BarComponent, p.QuoteComponent, p.TradeComponent)
	return fmt.Sprintf("zecema(%d, %.8f, %d%s)", equivLength, p.SmoothingFactor, p.GainLimit, suffix)
}

// EmaMnemonic returns the baseline EMA output mnemonic for p.
func (p SmoothingFactorParams) EmaMnemonic() string {
	equivLength := equivalentLengthFromAlpha(p.SmoothingFactor)
	suffix := resolveComponentSuffix(p.BarComponent, p.QuoteComponent, p.TradeComponent)
	return fmt.Sprintf("zecemaEma(%d, %.8f, %d%s)", equivLength, p.SmoothingFactor, p.GainLimit, suffix)
}
