package advancedecline

import (
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
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
//	AD  = AD_previous + CLV × Volume
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
func (s *AdvanceDecline) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *AdvanceDecline) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.AdvanceDecline,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the indicator with the given sample.
// Since scalar updates use the same value for H, L, C, the range is 0 and AD is unchanged.
func (s *AdvanceDecline) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return math.NaN()
	}

	return s.UpdateHLCV(sample, sample, sample, 1)
}

// UpdateHLCV updates the indicator with the given high, low, close, and volume values.
func (s *AdvanceDecline) UpdateHLCV(high, low, close, volume float64) float64 {
	if math.IsNaN(high) || math.IsNaN(low) || math.IsNaN(close) || math.IsNaN(volume) {
		return math.NaN()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	temp := high - low
	if temp > 0 {
		s.ad += ((close - low) - (high - close)) / temp * volume
	}

	s.value = s.ad
	s.primed = true

	return s.value
}

// UpdateBar updates the indicator given the next bar sample.
// This shadows LineIndicator.UpdateBar to extract HLCV from the bar.
func (s *AdvanceDecline) UpdateBar(sample *entities.Bar) core.Output {
	value := s.UpdateHLCV(sample.High, sample.Low, sample.Close, sample.Volume)

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: value}

	return output
}
