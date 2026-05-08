package jurikdirectionalmovementindex

import (
	"fmt"
	"math"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/markjurik/jurikmovingaverage"
)

// JurikDirectionalMovementIndex computes the Jurik directional movement index (DMX),
// see http://jurikres.com/.
//
// It produces three output lines:
//   - Bipolar: 100*(Plus-Minus)/(Plus+Minus)
//   - Plus: JMA(upward) / JMA(TrueRange)
//   - Minus: JMA(downward) / JMA(TrueRange)
//
// The internal JMA instances use phase=-100 (maximum lag, no overshoot).
type JurikDirectionalMovementIndex struct {
	mu          sync.RWMutex
	mnemonic    string
	description string
	primed      bool
	bar         int
	prevHigh    float64
	prevLow     float64
	prevClose   float64
	jmaPlus     *jurikmovingaverage.JurikMovingAverage
	jmaMinus    *jurikmovingaverage.JurikMovingAverage
	jmaDenom    *jurikmovingaverage.JurikMovingAverage
	plusVal     float64
	minusVal    float64
	bipolarVal  float64
}

// NewJurikDirectionalMovementIndex returns an instance of the indicator created using supplied parameters.
func NewJurikDirectionalMovementIndex(p *JurikDirectionalMovementIndexParams) (*JurikDirectionalMovementIndex, error) {
	return newJurikDirectionalMovementIndex(p.Length)
}

func newJurikDirectionalMovementIndex(length int) (*JurikDirectionalMovementIndex, error) {
	const (
		invalid = "invalid jurik directional movement index parameters"
		fmts    = "%s: %s"
fmtn = "jdmx(%d)"
		phase   = -100
		minlen  = 1
	)

	if length < minlen {
		return nil, fmt.Errorf(fmts, invalid, "length should be positive")
	}

	jmaParams := &jurikmovingaverage.JurikMovingAverageParams{
		Length: length,
		Phase:  phase,
	}

	jmaPlus, err := jurikmovingaverage.NewJurikMovingAverage(jmaParams)
	if err != nil {
		return nil, fmt.Errorf(fmts, invalid, err.Error())
	}

	jmaMinus, err := jurikmovingaverage.NewJurikMovingAverage(jmaParams)
	if err != nil {
		return nil, fmt.Errorf(fmts, invalid, err.Error())
	}

	jmaDenom, err := jurikmovingaverage.NewJurikMovingAverage(jmaParams)
	if err != nil {
		return nil, fmt.Errorf(fmts, invalid, err.Error())
	}

	mnemonic := fmt.Sprintf(fmtn, length)

	return &JurikDirectionalMovementIndex{
		mnemonic:    mnemonic,
		description: "Jurik directional movement index " + mnemonic,
		jmaPlus:     jmaPlus,
		jmaMinus:    jmaMinus,
		jmaDenom:    jmaDenom,
		prevHigh:    math.NaN(),
		prevLow:     math.NaN(),
		prevClose:   math.NaN(),
	}, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *JurikDirectionalMovementIndex) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *JurikDirectionalMovementIndex) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.JurikDirectionalMovementIndex,
		s.mnemonic,
		s.description,
		[]core.OutputText{
			{Mnemonic: s.mnemonic + ":bipolar", Description: s.description + " bipolar"},
			{Mnemonic: s.mnemonic + ":plus", Description: s.description + " plus"},
			{Mnemonic: s.mnemonic + ":minus", Description: s.description + " minus"},
		},
	)
}

// Update updates the indicator given the next high, low, and close values.
func (s *JurikDirectionalMovementIndex) Update(high, low, close float64) (bipolar, plus, minus float64) { //nolint:cyclop
	s.mu.Lock()
	defer s.mu.Unlock()

	const (
		warmup  = 41
		epsilon = 0.00001
		hundred = 100.0
	)

	s.bar++

	// TrueRange starts from bar 2 (needs prevClose from bar before previous).
	// Upward/downward starts from bar 1 (needs prevHigh/prevLow).
	var trueRange, upward, downward float64

	if s.bar >= 2 { //nolint:mnd
		// Upward/downward movement.
		v1 := hundred * (high - s.prevHigh)
		v2 := hundred * (s.prevLow - low)

		if v1 > v2 && v1 > 0 {
			upward = v1
		}

		if v2 > v1 && v2 > 0 {
			downward = v2
		}
	}

	if s.bar >= 3 { //nolint:mnd
		// True range (starts from bar 2 in reference — needs 2 previous bars).
		m1 := math.Abs(high - low)
		m2 := math.Abs(high - s.prevClose)
		m3 := math.Abs(low - s.prevClose)
		trueRange = math.Max(math.Max(m1, m2), m3)
	}

	s.prevHigh = high
	s.prevLow = low
	s.prevClose = close

	// Feed into JMA instances.
	numerPlus := s.jmaPlus.Update(upward)
	numerMinus := s.jmaMinus.Update(downward)
	denom := s.jmaDenom.Update(trueRange)

	if s.bar <= warmup {
		s.bipolarVal = math.NaN()
		s.plusVal = math.NaN()
		s.minusVal = math.NaN()

		return math.NaN(), math.NaN(), math.NaN()
	}

	s.primed = true

	// Compute Plus and Minus.
	if denom > epsilon {
		s.plusVal = numerPlus / denom
	} else {
		s.plusVal = 0
	}

	if denom > epsilon {
		s.minusVal = numerMinus / denom
	} else {
		s.minusVal = 0
	}

	// Compute Bipolar.
	sum := s.plusVal + s.minusVal
	if sum > epsilon {
		s.bipolarVal = hundred * (s.plusVal - s.minusVal) / sum
	} else {
		s.bipolarVal = 0
	}

	return s.bipolarVal, s.plusVal, s.minusVal
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *JurikDirectionalMovementIndex) UpdateScalar(sample *entities.Scalar) core.Output {
	v := sample.Value

	return s.updateEntity(sample.Time, v, v, v)
}

// UpdateBar updates the indicator given the next bar sample.
func (s *JurikDirectionalMovementIndex) UpdateBar(sample *entities.Bar) core.Output {
	return s.updateEntity(sample.Time, sample.High, sample.Low, sample.Close)
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *JurikDirectionalMovementIndex) UpdateQuote(sample *entities.Quote) core.Output {
	return s.updateEntity(sample.Time, sample.Ask, sample.Bid, (sample.Ask+sample.Bid)/2) //nolint:mnd
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *JurikDirectionalMovementIndex) UpdateTrade(sample *entities.Trade) core.Output {
	v := sample.Price

	return s.updateEntity(sample.Time, v, v, v)
}

func (s *JurikDirectionalMovementIndex) updateEntity(
	t time.Time, high, low, close float64,
) core.Output {
	const outputCount = 3

	output := make([]any, outputCount)
	bipolar, plus, minus := s.Update(high, low, close)

	i := 0
	output[i] = entities.Scalar{Time: t, Value: bipolar}
	i++
	output[i] = entities.Scalar{Time: t, Value: plus}
	i++
	output[i] = entities.Scalar{Time: t, Value: minus}

	return output
}
