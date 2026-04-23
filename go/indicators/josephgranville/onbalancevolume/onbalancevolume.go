package onbalancevolume

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// OnBalanceVolume is Joseph Granville's On-Balance Volume (OBV).
//
// OBV is a cumulative volume indicator. On each update, if the price is higher
// than the previous price, the volume is added to the running total; if the price
// is lower, the volume is subtracted. If the price is unchanged, the total remains
// the same.
//
// Reference:
//
// Granville, Joseph (1963). "Granville's New Key to Stock Market Profits".
type OnBalanceVolume struct {
	mu sync.RWMutex
	core.LineIndicator
	barFunc        entities.BarFunc
	previousSample float64
	value          float64
	primed         bool
}

// NewOnBalanceVolume returns an instance of the indicator created using supplied parameters.
func NewOnBalanceVolume(p *OnBalanceVolumeParams) (*OnBalanceVolume, error) {
	const (
		invalid = "invalid on-balance volume parameters"
		fmtw    = "%s: %w"
	)

	bc := p.BarComponent
	if bc == 0 {
		bc = entities.BarClosePrice
	}

	qc := p.QuoteComponent
	if qc == 0 {
		qc = entities.DefaultQuoteComponent
	}

	tc := p.TradeComponent
	if tc == 0 {
		tc = entities.DefaultTradeComponent
	}

	var (
		err       error
		barFunc   entities.BarFunc
		quoteFunc entities.QuoteFunc
		tradeFunc entities.TradeFunc
	)

	if barFunc, err = entities.BarComponentFunc(bc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	if quoteFunc, err = entities.QuoteComponentFunc(qc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	if tradeFunc, err = entities.TradeComponentFunc(tc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	mnemonic := "obv"
	if suffix := core.ComponentTripleMnemonic(bc, qc, tc); suffix != "" {
		mnemonic = "obv(" + suffix[2:] + ")" // strip leading ", "
	}

	desc := "On-Balance Volume OBV"

	o := &OnBalanceVolume{
		barFunc: barFunc,
		value:   math.NaN(),
	}

	// LineIndicator's Update uses volume=1 (scalar-only path).
	o.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, o.Update)

	return o, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *OnBalanceVolume) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *OnBalanceVolume) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.OnBalanceVolume,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the indicator with the given sample using volume = 1.
// This satisfies the LineIndicator updateFn signature.
func (s *OnBalanceVolume) Update(sample float64) float64 {
	return s.UpdateWithVolume(sample, 1)
}

// UpdateWithVolume updates the indicator with the given sample and volume.
func (s *OnBalanceVolume) UpdateWithVolume(sample, volume float64) float64 {
	if math.IsNaN(sample) || math.IsNaN(volume) {
		return math.NaN()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.primed {
		s.value = volume
		s.primed = true
	} else {
		if sample > s.previousSample {
			s.value += volume
		} else if sample < s.previousSample {
			s.value -= volume
		}
	}

	s.previousSample = sample

	return s.value
}

// UpdateBar updates the indicator given the next bar sample.
// This shadows LineIndicator.UpdateBar to use bar volume.
func (s *OnBalanceVolume) UpdateBar(sample *entities.Bar) core.Output {
	price := s.barFunc(sample)
	value := s.UpdateWithVolume(price, sample.Volume)

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: value}

	return output
}
