package onbalancevolume

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
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
func (o *OnBalanceVolume) IsPrimed() bool {
	o.mu.RLock()
	defer o.mu.RUnlock()

	return o.primed
}

// Metadata describes the output data of the indicator.
func (o *OnBalanceVolume) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.OnBalanceVolume,
		Mnemonic:    o.LineIndicator.Mnemonic,
		Description: o.LineIndicator.Description,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(OnBalanceVolumeValue),
				Type:        outputs.ScalarType,
				Mnemonic:    o.LineIndicator.Mnemonic,
				Description: o.LineIndicator.Description,
			},
		},
	}
}

// Update updates the indicator with the given sample using volume = 1.
// This satisfies the LineIndicator updateFn signature.
func (o *OnBalanceVolume) Update(sample float64) float64 {
	return o.UpdateWithVolume(sample, 1)
}

// UpdateWithVolume updates the indicator with the given sample and volume.
func (o *OnBalanceVolume) UpdateWithVolume(sample, volume float64) float64 {
	if math.IsNaN(sample) || math.IsNaN(volume) {
		return math.NaN()
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	if !o.primed {
		o.value = volume
		o.primed = true
	} else {
		if sample > o.previousSample {
			o.value += volume
		} else if sample < o.previousSample {
			o.value -= volume
		}
	}

	o.previousSample = sample

	return o.value
}

// UpdateBar updates the indicator given the next bar sample.
// This shadows LineIndicator.UpdateBar to use bar volume.
func (o *OnBalanceVolume) UpdateBar(sample *entities.Bar) core.Output {
	price := o.barFunc(sample)
	value := o.UpdateWithVolume(price, sample.Volume)

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: value}

	return output
}
