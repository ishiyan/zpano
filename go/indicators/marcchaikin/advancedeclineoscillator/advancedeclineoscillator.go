package advancedeclineoscillator

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/common/exponentialmovingaverage"
	"zpano/indicators/common/simplemovingaverage"
	"zpano/indicators/core"
)

// lineUpdater is an interface for indicators that accept a single scalar and return a value.
type lineUpdater interface {
	Update(float64) float64
	IsPrimed() bool
}

// AdvanceDeclineOscillator is Marc Chaikin's Advance-Decline (A/D) Oscillator.
//
// The Chaikin Oscillator is the difference between a fast and slow moving average
// of the Accumulation/Distribution Line. It is used to anticipate changes in the A/D Line
// by measuring the momentum behind accumulation/distribution movements.
//
// The value is calculated as:
//
//	CLV = ((Close - Low) - (High - Close)) / (High - Low)
//	AD  = AD_previous + CLV × Volume
//	ADOSC = FastMA(AD) - SlowMA(AD)
//
// When High equals Low, the A/D value is unchanged (no division by zero).
//
// Reference:
//
// Chaikin, Marc. "Chaikin Oscillator".
type AdvanceDeclineOscillator struct {
	mu sync.RWMutex
	core.LineIndicator
	ad     float64
	fastMA lineUpdater
	slowMA lineUpdater
	value  float64
	primed bool
}

// NewAdvanceDeclineOscillator returns an instance of the indicator created using supplied parameters.
func NewAdvanceDeclineOscillator(p *AdvanceDeclineOscillatorParams) (*AdvanceDeclineOscillator, error) {
	const (
		invalid   = "invalid advance-decline oscillator parameters"
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

	var fastMA, slowMA lineUpdater

	var maLabel string

	switch p.MovingAverageType {
	case SMA:
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
	default:
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
	}

	mnemonic := fmt.Sprintf("adosc(%s%d/%s%d)", maLabel, p.FastLength, maLabel, p.SlowLength)
	desc := "Chaikin Advance-Decline Oscillator " + mnemonic

	barFunc, _ := entities.BarComponentFunc(entities.BarClosePrice)
	quoteFunc, _ := entities.QuoteComponentFunc(entities.DefaultQuoteComponent)
	tradeFunc, _ := entities.TradeComponentFunc(entities.DefaultTradeComponent)

	a := &AdvanceDeclineOscillator{
		fastMA: fastMA,
		slowMA: slowMA,
		value:  math.NaN(),
	}

	a.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, a.Update)

	return a, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *AdvanceDeclineOscillator) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *AdvanceDeclineOscillator) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.AdvanceDeclineOscillator,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the indicator with the given sample.
// Since scalar updates use the same value for H, L, C, the range is 0 and AD is unchanged,
// but the unchanged AD value is still fed to the MAs.
func (s *AdvanceDeclineOscillator) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return math.NaN()
	}

	return s.UpdateHLCV(sample, sample, sample, 1)
}

// UpdateHLCV updates the indicator with the given high, low, close, and volume values.
func (s *AdvanceDeclineOscillator) UpdateHLCV(high, low, close, volume float64) float64 {
	if math.IsNaN(high) || math.IsNaN(low) || math.IsNaN(close) || math.IsNaN(volume) {
		return math.NaN()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Compute cumulative AD.
	temp := high - low
	if temp > 0 {
		s.ad += ((close - low) - (high - close)) / temp * volume
	}

	// Feed AD to both MAs.
	fast := s.fastMA.Update(s.ad)
	slow := s.slowMA.Update(s.ad)
	s.primed = s.fastMA.IsPrimed() && s.slowMA.IsPrimed()

	if math.IsNaN(fast) || math.IsNaN(slow) {
		s.value = math.NaN()

		return s.value
	}

	s.value = fast - slow

	return s.value
}

// UpdateBar updates the indicator given the next bar sample.
// This shadows LineIndicator.UpdateBar to extract HLCV from the bar.
func (s *AdvanceDeclineOscillator) UpdateBar(sample *entities.Bar) core.Output {
	value := s.UpdateHLCV(sample.High, sample.Low, sample.Close, sample.Volume)

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: value}

	return output
}
