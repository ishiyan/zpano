package core

import (
	"zpano/entities"
)

// LineIndicator provides UpdateScalar, UpdateBar, UpdateQuote and UpdateTrade
// methods for indicators that take a single numeric input and produce a single
// scalar output.
//
// Embed this struct in a concrete indicator and initialize it using
// NewLineIndicator, passing the concrete indicator's Update method as updateFn.
// The four Update* methods will be promoted to the embedding struct and satisfy
// the Indicator interface.
type LineIndicator struct {
	// Mnemonic is a short name of the indicator.
	Mnemonic string

	// Description is a description of the indicator.
	Description string

	barFunc   entities.BarFunc
	quoteFunc entities.QuoteFunc
	tradeFunc entities.TradeFunc
	updateFn  func(float64) float64
}

// NewLineIndicator creates a LineIndicator with the given component functions
// and an update function that computes the indicator value from a single sample.
func NewLineIndicator(
	mnemonic, description string,
	barFunc entities.BarFunc, quoteFunc entities.QuoteFunc, tradeFunc entities.TradeFunc,
	updateFn func(float64) float64,
) LineIndicator {
	return LineIndicator{
		Mnemonic:    mnemonic,
		Description: description,
		barFunc:     barFunc,
		quoteFunc:   quoteFunc,
		tradeFunc:   tradeFunc,
		updateFn:    updateFn,
	}
}

// UpdateScalar updates the indicator given the next scalar sample.
func (l *LineIndicator) UpdateScalar(sample *entities.Scalar) Output {
	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: l.updateFn(sample.Value)}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (l *LineIndicator) UpdateBar(sample *entities.Bar) Output {
	return l.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: l.barFunc(sample)})
}

// UpdateQuote updates the indicator given the next quote sample.
func (l *LineIndicator) UpdateQuote(sample *entities.Quote) Output {
	return l.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: l.quoteFunc(sample)})
}

// UpdateTrade updates the indicator given the next trade sample.
func (l *LineIndicator) UpdateTrade(sample *entities.Trade) Output {
	return l.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: l.tradeFunc(sample)})
}
