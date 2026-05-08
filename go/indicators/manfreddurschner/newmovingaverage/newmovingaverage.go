package newmovingaverage

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// MAType enumerates the moving average types used in NMA.
type MAType int

const (
	// SMA is the simple moving average.
	SMA MAType = iota
	// EMA is the exponential moving average.
	EMA
	// SMMA is the smoothed moving average.
	SMMA
	// LWMA is the linear weighted moving average.
	LWMA
)

// streamingMA is a common interface for the streaming moving averages.
type streamingMA interface {
	update(sample float64) float64
}

// streamingSMA computes a streaming simple moving average.
type streamingSMA struct {
	period      int
	buffer      []float64
	bufferIndex int
	bufferCount int
	sum         float64
	primed      bool
}

func newStreamingSMA(period int) *streamingSMA {
	return &streamingSMA{
		period: period,
		buffer: make([]float64, period),
	}
}

func (s *streamingSMA) update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	if s.primed {
		s.sum -= s.buffer[s.bufferIndex]
	}

	s.buffer[s.bufferIndex] = sample
	s.sum += sample
	s.bufferIndex = (s.bufferIndex + 1) % s.period

	if !s.primed {
		s.bufferCount++
		if s.bufferCount < s.period {
			return math.NaN()
		}

		s.primed = true
	}

	return s.sum / float64(s.period)
}

// streamingEMA computes a streaming exponential moving average (SMA-seeded).
type streamingEMA struct {
	period     int
	multiplier float64
	count      int
	sum        float64
	value      float64
	primed     bool
}

func newStreamingEMA(period int) *streamingEMA {
	return &streamingEMA{
		period:     period,
		multiplier: 2.0 / float64(period+1),
		value:      math.NaN(),
	}
}

func (e *streamingEMA) update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	if !e.primed {
		e.count++
		e.sum += sample

		if e.count < e.period {
			return math.NaN()
		}

		e.value = e.sum / float64(e.period)
		e.primed = true

		return e.value
	}

	e.value = (sample-e.value)*e.multiplier + e.value

	return e.value
}

// streamingSMMA computes a streaming smoothed moving average (SMA-seeded).
type streamingSMMA struct {
	period int
	count  int
	sum    float64
	value  float64
	primed bool
}

func newStreamingSMMA(period int) *streamingSMMA {
	return &streamingSMMA{
		period: period,
		value:  math.NaN(),
	}
}

func (s *streamingSMMA) update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	if !s.primed {
		s.count++
		s.sum += sample

		if s.count < s.period {
			return math.NaN()
		}

		s.value = s.sum / float64(s.period)
		s.primed = true

		return s.value
	}

	s.value = (s.value*float64(s.period-1) + sample) / float64(s.period)

	return s.value
}

// streamingLWMA computes a streaming linear weighted moving average.
type streamingLWMA struct {
	period      int
	buffer      []float64
	bufferIndex int
	bufferCount int
	weightSum   float64
	primed      bool
}

func newStreamingLWMA(period int) *streamingLWMA {
	return &streamingLWMA{
		period:    period,
		buffer:    make([]float64, period),
		weightSum: float64(period) * float64(period+1) / 2.0,
	}
}

func (l *streamingLWMA) update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	l.buffer[l.bufferIndex] = sample
	l.bufferIndex = (l.bufferIndex + 1) % l.period

	if !l.primed {
		l.bufferCount++
		if l.bufferCount < l.period {
			return math.NaN()
		}

		l.primed = true
	}

	result := 0.0
	index := l.bufferIndex

	for i := 0; i < l.period; i++ {
		result += float64(i+1) * l.buffer[index]
		index = (index + 1) % l.period
	}

	return result / l.weightSum
}

func createStreamingMA(maType MAType, period int) streamingMA {
	switch maType {
	case SMA:
		return newStreamingSMA(period)
	case EMA:
		return newStreamingEMA(period)
	case SMMA:
		return newStreamingSMMA(period)
	case LWMA:
		return newStreamingLWMA(period)
	default:
		return newStreamingLWMA(period)
	}
}

// NewMovingAverage computes the New Moving Average (NMA) by Dürschner.
//
// NMA applies the Nyquist-Shannon sampling theorem to moving average design:
// by cascading two moving averages whose period ratio satisfies the Nyquist
// criterion (lambda = n1/n2 >= 2), the resulting lag can be extrapolated away
// geometrically.
//
// Formula: NMA = (1 + alpha) * MA1 - alpha * MA2
// where: alpha = lambda * (n1-1) / (n1-lambda), lambda = n1 // n2
type NewMovingAverage struct {
	mu sync.RWMutex
	core.LineIndicator
	alpha       float64
	maPrimary   streamingMA
	maSecondary streamingMA
	primed      bool
}

// NewNewMovingAverage returns an instance of the indicator created using supplied parameters.
func NewNewMovingAverage(p *NewMovingAverageParams) (*NewMovingAverage, error) {
	const (
		fmtn = "nma(%d, %d, %d%s)"
	)

	primaryPeriod := p.PrimaryPeriod
	secondaryPeriod := p.SecondaryPeriod

	// Enforce Nyquist constraint.
	if primaryPeriod < 4 {
		primaryPeriod = 4
	}

	if secondaryPeriod < 2 {
		secondaryPeriod = 2
	}

	if primaryPeriod < secondaryPeriod*2 {
		primaryPeriod = secondaryPeriod * 4
	}

	// Compute alpha.
	nyquistRatio := primaryPeriod / secondaryPeriod
	alpha := float64(nyquistRatio) * float64(primaryPeriod-1) / float64(primaryPeriod-nyquistRatio)

	// Resolve defaults for component functions.
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
		return nil, fmt.Errorf("invalid new moving average parameters: %w", err)
	}

	if quoteFunc, err = entities.QuoteComponentFunc(qc); err != nil {
		return nil, fmt.Errorf("invalid new moving average parameters: %w", err)
	}

	if tradeFunc, err = entities.TradeComponentFunc(tc); err != nil {
		return nil, fmt.Errorf("invalid new moving average parameters: %w", err)
	}

	mnemonic := fmt.Sprintf(fmtn, primaryPeriod, secondaryPeriod, int(p.MAType), core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "New moving average " + mnemonic

	nma := &NewMovingAverage{
		alpha:       alpha,
		maPrimary:   createStreamingMA(p.MAType, primaryPeriod),
		maSecondary: createStreamingMA(p.MAType, secondaryPeriod),
	}

	nma.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, nma.Update)

	return nma, nil
}

// IsPrimed indicates whether the indicator is primed.
func (n *NewMovingAverage) IsPrimed() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.primed
}

// Metadata describes the output data of the indicator.
func (n *NewMovingAverage) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.NewMovingAverage,
		n.LineIndicator.Mnemonic,
		n.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: n.LineIndicator.Mnemonic, Description: n.LineIndicator.Description},
		},
	)
}

// Update updates the value of the new moving average given the next sample.
func (n *NewMovingAverage) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	// First filter: MA of raw price.
	ma1Value := n.maPrimary.update(sample)
	if math.IsNaN(ma1Value) {
		return math.NaN()
	}

	// Second filter: MA of MA1 output.
	ma2Value := n.maSecondary.update(ma1Value)
	if math.IsNaN(ma2Value) {
		return math.NaN()
	}

	n.primed = true

	// Geometric extrapolation.
	return (1.0+n.alpha)*ma1Value - n.alpha*ma2Value
}
