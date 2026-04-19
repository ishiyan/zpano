package advancedecline

import (
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
)

// AdvanceDecline is Marc Chaikin's Advance-Decline (A/D) Line.
//
// The Accumulation/Distribution Line is a cumulative indicator that uses volume
// and price to assess whether a stock is being accumulated or distributed.
// The A/D line seeks to identify divergences between the stock price and volume flow.
//
// The value is calculated as:
//
//	CLV = ((Close - Low) - (High - Close)) / (High - Low)
//	AD  = AD_prev + CLV × Volume
//
// When High equals Low, the A/D value is unchanged (no division by zero).
//
// Reference:
//
// Chaikin, Marc. "Chaikin Accumulation/Distribution Line".
type AdvanceDecline struct {
	mu sync.RWMutex
	core.LineIndicator
	ad     float64
	value  float64
	primed bool
}

// NewAdvanceDecline returns an instance of the indicator created using supplied parameters.
func NewAdvanceDecline(_ *AdvanceDeclineParams) (*AdvanceDecline, error) {
	mnemonic := "ad"
	desc := "Advance-Decline"

	barFunc, _ := entities.BarComponentFunc(entities.BarClosePrice)
	quoteFunc, _ := entities.QuoteComponentFunc(entities.DefaultQuoteComponent)
	tradeFunc, _ := entities.TradeComponentFunc(entities.DefaultTradeComponent)

	a := &AdvanceDecline{
		value: math.NaN(),
	}

	// For scalar/quote/trade updates, H=L=C so AD is unchanged.
	a.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, a.Update)

	return a, nil
}

// IsPrimed indicates whether the indicator is primed.
func (a *AdvanceDecline) IsPrimed() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.primed
}

// Metadata describes the output data of the indicator.
func (a *AdvanceDecline) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.AdvanceDecline,
		Mnemonic:    a.LineIndicator.Mnemonic,
		Description: a.LineIndicator.Description,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(AdvanceDeclineValue),
				Type:        outputs.ScalarType,
				Mnemonic:    a.LineIndicator.Mnemonic,
				Description: a.LineIndicator.Description,
			},
		},
	}
}

// Update updates the indicator with the given sample.
// Since scalar updates use the same value for H, L, C, the range is 0 and AD is unchanged.
func (a *AdvanceDecline) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return math.NaN()
	}

	return a.UpdateHLCV(sample, sample, sample, 1)
}

// UpdateHLCV updates the indicator with the given high, low, close, and volume values.
func (a *AdvanceDecline) UpdateHLCV(high, low, close, volume float64) float64 {
	if math.IsNaN(high) || math.IsNaN(low) || math.IsNaN(close) || math.IsNaN(volume) {
		return math.NaN()
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	tmp := high - low
	if tmp > 0 {
		a.ad += ((close - low) - (high - close)) / tmp * volume
	}

	a.value = a.ad
	a.primed = true

	return a.value
}

// UpdateBar updates the indicator given the next bar sample.
// This shadows LineIndicator.UpdateBar to extract HLCV from the bar.
func (a *AdvanceDecline) UpdateBar(sample *entities.Bar) core.Output {
	value := a.UpdateHLCV(sample.High, sample.Low, sample.Close, sample.Volume)

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: value}

	return output
}
