package percentagepriceoscillator

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

// PercentagePriceOscillator is Gerald Appel's Percentage Price Oscillator (PPO).
//
// PPO is calculated by subtracting the slower moving average from the faster moving
// average and then dividing the result by the slower moving average, expressed as a percentage.
//
//	PPO = 100 * (fast_ma - slow_ma) / slow_ma
//
// Reference:
//
// Appel, Gerald (2005). Technical Analysis: Power Tools for Active Investors.
type PercentagePriceOscillator struct {
	mu sync.RWMutex
	core.LineIndicator
	fastMA lineUpdater
	slowMA lineUpdater
	value  float64
	primed bool
}

// NewPercentagePriceOscillator returns an instance of the indicator created using supplied parameters.
func NewPercentagePriceOscillator(p *PercentagePriceOscillatorParams) (*PercentagePriceOscillator, error) {
	const (
		invalid   = "invalid percentage price oscillator parameters"
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

	mnemonic := fmt.Sprintf("ppo(%s%d/%s%d%s)", maLabel, p.FastLength, maLabel, p.SlowLength,
		core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Percentage Price Oscillator " + mnemonic

	ppo := &PercentagePriceOscillator{
		fastMA: fastMA,
		slowMA: slowMA,
		value:  math.NaN(),
	}

	ppo.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, ppo.Update)

	return ppo, nil
}

// IsPrimed indicates whether the indicator is primed.
func (p *PercentagePriceOscillator) IsPrimed() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.primed
}

// Metadata describes the output data of the indicator.
func (p *PercentagePriceOscillator) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.PercentagePriceOscillator,
		Mnemonic:    p.LineIndicator.Mnemonic,
		Description: p.LineIndicator.Description,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(PercentagePriceOscillatorValue),
				Type:        outputs.ScalarType,
				Mnemonic:    p.LineIndicator.Mnemonic,
				Description: p.LineIndicator.Description,
			},
		},
	}
}

// Update updates the value of the indicator given the next sample.
func (p *PercentagePriceOscillator) Update(sample float64) float64 {
	const epsilon = 1e-8

	if math.IsNaN(sample) {
		return sample
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	slow := p.slowMA.Update(sample)
	fast := p.fastMA.Update(sample)
	p.primed = p.slowMA.IsPrimed() && p.fastMA.IsPrimed()

	if math.IsNaN(fast) || math.IsNaN(slow) {
		p.value = math.NaN()

		return p.value
	}

	if math.Abs(slow) < epsilon {
		p.value = 0
	} else {
		p.value = 100 * (fast - slow) / slow
	}

	return p.value
}
