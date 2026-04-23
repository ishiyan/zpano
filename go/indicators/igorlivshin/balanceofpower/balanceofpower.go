package balanceofpower

import (
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

const epsilon = 1e-8

// BalanceOfPower is Igor Livshin's Balance of Power (BOP).
//
// The Balance of Market Power captures the struggles of bulls vs. bears
// throughout the trading day. It assigns scores to both bulls and bears
// based on how much they were able to move prices throughout the trading day.
//
// The value is calculated as:
//
//	BOP = (Close - Open) / (High - Low)
//
// When the range (High - Low) is less than epsilon, the value is 0.
//
// Reference:
//
// Livshin, Igor. "Balance of Market Power".
type BalanceOfPower struct {
	mu sync.RWMutex
	core.LineIndicator
	value float64
}

// NewBalanceOfPower returns an instance of the indicator created using supplied parameters.
func NewBalanceOfPower(_ *BalanceOfPowerParams) (*BalanceOfPower, error) {
	mnemonic := "bop"
	desc := "Balance of Power"

	barFunc, _ := entities.BarComponentFunc(entities.BarClosePrice)
	quoteFunc, _ := entities.QuoteComponentFunc(entities.DefaultQuoteComponent)
	tradeFunc, _ := entities.TradeComponentFunc(entities.DefaultTradeComponent)

	b := &BalanceOfPower{
		value: math.NaN(),
	}

	// For scalar/quote/trade updates, O=H=L=C so BOP is always 0.
	b.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, b.Update)

	return b, nil
}

// IsPrimed indicates whether the indicator is primed.
// Balance of Power is always primed.
func (s *BalanceOfPower) IsPrimed() bool {
	return true
}

// Metadata describes the output data of the indicator.
func (s *BalanceOfPower) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.BalanceOfPower,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the indicator with the given sample.
// Since scalar updates use the same value for O, H, L, C, the result is always 0.
func (s *BalanceOfPower) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return math.NaN()
	}

	return s.UpdateOHLC(sample, sample, sample, sample)
}

// UpdateOHLC updates the indicator with the given OHLC values.
func (s *BalanceOfPower) UpdateOHLC(open, high, low, close float64) float64 {
	if math.IsNaN(open) || math.IsNaN(high) || math.IsNaN(low) || math.IsNaN(close) {
		return math.NaN()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	r := high - low
	if r < epsilon {
		s.value = 0
	} else {
		s.value = (close - open) / r
	}

	return s.value
}

// UpdateBar updates the indicator given the next bar sample.
// This shadows LineIndicator.UpdateBar to extract OHLC from the bar.
func (s *BalanceOfPower) UpdateBar(sample *entities.Bar) core.Output {
	value := s.UpdateOHLC(sample.Open, sample.High, sample.Low, sample.Close)

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: value}

	return output
}
