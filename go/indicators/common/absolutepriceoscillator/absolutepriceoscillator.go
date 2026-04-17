package absolutepriceoscillator

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/common/exponentialmovingaverage"
	"zpano/indicators/common/simplemovingaverage"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
)

// lineUpdater is an interface for indicators that accept a single scalar and return a value.
type lineUpdater interface {
	Update(float64) float64
	IsPrimed() bool
}

// AbsolutePriceOscillator is the Absolute Price Oscillator (APO).
//
// APO is calculated by subtracting the slower moving average from the faster moving average.
//
//	APO = fast_ma - slow_ma
type AbsolutePriceOscillator struct {
	mu sync.RWMutex
	core.LineIndicator
	fastMA lineUpdater
	slowMA lineUpdater
	value  float64
	primed bool
}

// NewAbsolutePriceOscillator returns an instance of the indicator created using supplied parameters.
func NewAbsolutePriceOscillator(p *AbsolutePriceOscillatorParams) (*AbsolutePriceOscillator, error) {
	const (
		invalid   = "invalid absolute price oscillator parameters"
		fmts      = "%s: %s"
		fmtw      = "%s: %w"
		minLength = 2
	)

	if p.FastLength < minLength {
		return nil, fmt.Errorf(fmts, invalid, "fast length should be greater than 1")
	}

	if p.SlowLength < minLength {
		return nil, fmt.Errorf(fmts, invalid, "slow length should be greater than 1")
	}

	bc := p.BarComponent
	if bc == 0 {
		bc = entities.DefaultBarComponent
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

	var fastMA, slowMA lineUpdater

	var maLabel string

	switch p.MovingAverageType {
	case EMA:
		maLabel = "EMA"

		fast, e := exponentialmovingaverage.NewExponentialMovingAverageLength(
			&exponentialmovingaverage.ExponentialMovingAverageLengthParams{
				Length:         p.FastLength,
				FirstIsAverage: p.FirstIsAverage,
			})
		if e != nil {
			return nil, fmt.Errorf(fmtw, invalid, e)
		}

		slow, e := exponentialmovingaverage.NewExponentialMovingAverageLength(
			&exponentialmovingaverage.ExponentialMovingAverageLengthParams{
				Length:         p.SlowLength,
				FirstIsAverage: p.FirstIsAverage,
			})
		if e != nil {
			return nil, fmt.Errorf(fmtw, invalid, e)
		}

		fastMA = fast
		slowMA = slow
	default:
		maLabel = "SMA"

		fast, e := simplemovingaverage.NewSimpleMovingAverage(
			&simplemovingaverage.SimpleMovingAverageParams{Length: p.FastLength})
		if e != nil {
			return nil, fmt.Errorf(fmtw, invalid, e)
		}

		slow, e := simplemovingaverage.NewSimpleMovingAverage(
			&simplemovingaverage.SimpleMovingAverageParams{Length: p.SlowLength})
		if e != nil {
			return nil, fmt.Errorf(fmtw, invalid, e)
		}

		fastMA = fast
		slowMA = slow
	}

	mnemonic := fmt.Sprintf("apo(%s%d/%s%d%s)", maLabel, p.FastLength, maLabel, p.SlowLength,
		core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Absolute Price Oscillator " + mnemonic

	apo := &AbsolutePriceOscillator{
		fastMA: fastMA,
		slowMA: slowMA,
		value:  math.NaN(),
	}

	apo.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, apo.Update)

	return apo, nil
}

// IsPrimed indicates whether the indicator is primed.
func (a *AbsolutePriceOscillator) IsPrimed() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.primed
}

// Metadata describes the output data of the indicator.
func (a *AbsolutePriceOscillator) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.AbsolutePriceOscillator,
		Mnemonic:    a.LineIndicator.Mnemonic,
		Description: a.LineIndicator.Description,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(AbsolutePriceOscillatorValue),
				Type:        outputs.ScalarType,
				Mnemonic:    a.LineIndicator.Mnemonic,
				Description: a.LineIndicator.Description,
			},
		},
	}
}

// Update updates the value of the indicator given the next sample.
func (a *AbsolutePriceOscillator) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	slow := a.slowMA.Update(sample)
	fast := a.fastMA.Update(sample)
	a.primed = a.slowMA.IsPrimed() && a.fastMA.IsPrimed()

	if math.IsNaN(fast) || math.IsNaN(slow) {
		a.value = math.NaN()

		return a.value
	}

	a.value = fast - slow

	return a.value
}
