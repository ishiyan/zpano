package hilberttransformerinstantaneoustrendline

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/johnehlers/hilberttransformer"
)

// HilbertTransformerInstantaneousTrendLine is Ehlers' Instantaneous Trend Line indicator
// built on top of a Hilbert transformer cycle estimator.
//
// It exposes two outputs:
//
//   - Value: the instantaneous trend line value, computed as a WMA of simple averages
//     over windows whose length tracks the smoothed dominant cycle period.
//   - DominantCyclePeriod: the additionally EMA-smoothed dominant cycle period.
//
// Reference:
// John Ehlers, Rocket Science for Traders, Wiley, 2001, 0471405671, pp 107-112.
type HilbertTransformerInstantaneousTrendLine struct {
	mu                             sync.RWMutex
	mnemonic                       string
	description                    string
	mnemonicDCP                    string
	descriptionDCP                 string
	htce                           hilberttransformer.CycleEstimator
	alphaEmaPeriodAdditional       float64
	oneMinAlphaEmaPeriodAdditional float64
	cyclePartMultiplier            float64
	trendLineSmoothingLength       int
	coeff0                         float64
	coeff1                         float64
	coeff2                         float64
	coeff3                         float64
	smoothedPeriod                 float64
	value                          float64
	average1                       float64
	average2                       float64
	average3                       float64
	input                          []float64
	inputLength                    int
	inputLengthMin1                int
	primed                         bool
	barFunc                        entities.BarFunc
	quoteFunc                      entities.QuoteFunc
	tradeFunc                      entities.TradeFunc
}

// NewHilbertTransformerInstantaneousTrendLineDefault returns an instance of the indicator
// created using default values of the parameters.
func NewHilbertTransformerInstantaneousTrendLineDefault() (*HilbertTransformerInstantaneousTrendLine, error) {
	const (
		smoothingLength           = 4
		alphaEmaQuadratureInPhase = 0.2
		alphaEmaPeriod            = 0.2
		alphaEmaPeriodAdditional  = 0.33
		warmUpPeriod              = 100
		trendLineSmoothingLength  = 4
		cyclePartMultiplier       = 1.0
	)

	return newHilbertTransformerInstantaneousTrendLine(
		hilberttransformer.HomodyneDiscriminator,
		&hilberttransformer.CycleEstimatorParams{
			SmoothingLength:           smoothingLength,
			AlphaEmaQuadratureInPhase: alphaEmaQuadratureInPhase,
			AlphaEmaPeriod:            alphaEmaPeriod,
			WarmUpPeriod:              warmUpPeriod,
		},
		alphaEmaPeriodAdditional,
		trendLineSmoothingLength,
		cyclePartMultiplier,
		0, 0, 0)
}

// NewHilbertTransformerInstantaneousTrendLineParams returns an instance of the indicator
// created using supplied parameters.
func NewHilbertTransformerInstantaneousTrendLineParams(
	p *Params,
) (*HilbertTransformerInstantaneousTrendLine, error) {
	return newHilbertTransformerInstantaneousTrendLine(
		p.EstimatorType, &p.EstimatorParams,
		p.AlphaEmaPeriodAdditional,
		p.TrendLineSmoothingLength,
		p.CyclePartMultiplier,
		p.BarComponent, p.QuoteComponent, p.TradeComponent)
}

//nolint:funlen,cyclop
func newHilbertTransformerInstantaneousTrendLine(
	estimatorType hilberttransformer.CycleEstimatorType,
	estimatorParams *hilberttransformer.CycleEstimatorParams,
	alphaEmaPeriodAdditional float64,
	trendLineSmoothingLength int,
	cyclePartMultiplier float64,
	bc entities.BarComponent, qc entities.QuoteComponent, tc entities.TradeComponent,
) (*HilbertTransformerInstantaneousTrendLine, error) {
	const (
		invalid    = "invalid hilbert transformer instantaneous trend line parameters"
		fmta       = "%s: α for additional smoothing should be in range (0, 1]"
		fmttlsl    = "%s: trend line smoothing length should be 2, 3, or 4"
		fmtcpm     = "%s: cycle part multiplier should be in range (0, 10]"
		fmtw       = "%s: %w"
		fmtn       = "htitl(%.3f, %d, %.3f%s%s)"
		fmtnDCP    = "dcp(%.3f%s%s)"
		four       = 4
		alpha      = 0.2
		descrValue = "Hilbert transformer instantaneous trend line "
		descrDCP   = "Dominant cycle period "
		tlslTwo    = 2
		tlslThree  = 3
		tlslFour   = 4
		cpmMax     = 10.0
	)

	if alphaEmaPeriodAdditional <= 0. || alphaEmaPeriodAdditional > 1. {
		return nil, fmt.Errorf(fmta, invalid)
	}

	if trendLineSmoothingLength < tlslTwo || trendLineSmoothingLength > tlslFour {
		return nil, fmt.Errorf(fmttlsl, invalid)
	}

	if cyclePartMultiplier <= 0. || cyclePartMultiplier > cpmMax {
		return nil, fmt.Errorf(fmtcpm, invalid)
	}

	// SineWave-style default: BarMedianPrice always shown in mnemonic.
	if bc == 0 {
		bc = entities.BarMedianPrice
	}

	if qc == 0 {
		qc = entities.DefaultQuoteComponent
	}

	if tc == 0 {
		tc = entities.DefaultTradeComponent
	}

	estimator, err := hilberttransformer.NewCycleEstimator(estimatorType, estimatorParams)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	estimatorMoniker := ""
	if estimatorType != hilberttransformer.HomodyneDiscriminator ||
		estimatorParams.SmoothingLength != four ||
		estimatorParams.AlphaEmaQuadratureInPhase != alpha ||
		estimatorParams.AlphaEmaPeriod != alpha {
		estimatorMoniker = hilberttransformer.EstimatorMoniker(estimatorType, estimator)
		if len(estimatorMoniker) > 0 {
			estimatorMoniker = ", " + estimatorMoniker
		}
	}

	componentMnemonic := core.ComponentTripleMnemonic(bc, qc, tc)

	mnemonic := fmt.Sprintf(fmtn, alphaEmaPeriodAdditional, trendLineSmoothingLength,
		cyclePartMultiplier, estimatorMoniker, componentMnemonic)
	mnemonicDCP := fmt.Sprintf(fmtnDCP, alphaEmaPeriodAdditional, estimatorMoniker, componentMnemonic)

	barFunc, err := entities.BarComponentFunc(bc)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	quoteFunc, err := entities.QuoteComponentFunc(qc)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	tradeFunc, err := entities.TradeComponentFunc(tc)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	var c0, c1, c2, c3 float64

	switch trendLineSmoothingLength {
	case tlslTwo:
		c0, c1 = 2.0/3.0, 1.0/3.0
	case tlslThree:
		c0, c1, c2 = 3.0/6.0, 2.0/6.0, 1.0/6.0
	default: // tlslFour
		c0, c1, c2, c3 = 4.0/10.0, 3.0/10.0, 2.0/10.0, 1.0/10.0
	}

	maxPeriod := estimator.MaxPeriod()

	return &HilbertTransformerInstantaneousTrendLine{
		mnemonic:                       mnemonic,
		description:                    descrValue + mnemonic,
		mnemonicDCP:                    mnemonicDCP,
		descriptionDCP:                 descrDCP + mnemonicDCP,
		htce:                           estimator,
		alphaEmaPeriodAdditional:       alphaEmaPeriodAdditional,
		oneMinAlphaEmaPeriodAdditional: 1. - alphaEmaPeriodAdditional,
		cyclePartMultiplier:            cyclePartMultiplier,
		trendLineSmoothingLength:       trendLineSmoothingLength,
		coeff0:                         c0,
		coeff1:                         c1,
		coeff2:                         c2,
		coeff3:                         c3,
		input:                          make([]float64, maxPeriod),
		inputLength:                    maxPeriod,
		inputLengthMin1:                maxPeriod - 1,
		barFunc:                        barFunc,
		quoteFunc:                      quoteFunc,
		tradeFunc:                      tradeFunc,
	}, nil
}

// IsPrimed indicates whether an indicator is primed.
func (s *HilbertTransformerInstantaneousTrendLine) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes an output data of the indicator.
func (s *HilbertTransformerInstantaneousTrendLine) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.HilbertTransformerInstantaneousTrendLine,
		s.mnemonic,
		s.description,
		[]core.OutputText{
			{Mnemonic: s.mnemonic, Description: s.description},
			{Mnemonic: s.mnemonicDCP, Description: s.descriptionDCP},
		},
	)
}

// Update updates the value of the indicator given the next sample, returning the
// (value, period) pair. Returns NaN values if the indicator is not yet primed.
func (s *HilbertTransformerInstantaneousTrendLine) Update(sample float64) (float64, float64) {
	if math.IsNaN(sample) {
		return sample, sample
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.htce.Update(sample)
	s.pushInput(sample)

	if s.primed {
		s.smoothedPeriod = s.alphaEmaPeriodAdditional*s.htce.Period() +
			s.oneMinAlphaEmaPeriodAdditional*s.smoothedPeriod
		average := s.calculateAverage()
		s.value = s.coeff0*average + s.coeff1*s.average1 + s.coeff2*s.average2 + s.coeff3*s.average3
		s.average3 = s.average2
		s.average2 = s.average1
		s.average1 = average

		return s.value, s.smoothedPeriod
	}

	if s.htce.Primed() {
		s.primed = true
		s.smoothedPeriod = s.htce.Period()
		average := s.calculateAverage()
		s.value = average
		s.average1 = average
		s.average2 = average
		s.average3 = average

		return s.value, s.smoothedPeriod
	}

	nan := math.NaN()

	return nan, nan
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *HilbertTransformerInstantaneousTrendLine) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.updateEntity(sample.Time, sample.Value)
}

// UpdateBar updates the indicator given the next bar sample.
func (s *HilbertTransformerInstantaneousTrendLine) UpdateBar(sample *entities.Bar) core.Output {
	return s.updateEntity(sample.Time, s.barFunc(sample))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *HilbertTransformerInstantaneousTrendLine) UpdateQuote(sample *entities.Quote) core.Output {
	return s.updateEntity(sample.Time, s.quoteFunc(sample))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *HilbertTransformerInstantaneousTrendLine) UpdateTrade(sample *entities.Trade) core.Output {
	return s.updateEntity(sample.Time, s.tradeFunc(sample))
}

func (s *HilbertTransformerInstantaneousTrendLine) updateEntity(t time.Time, sample float64) core.Output {
	const length = 2

	value, period := s.Update(sample)

	output := make([]any, length)
	output[0] = entities.Scalar{Time: t, Value: value}
	output[1] = entities.Scalar{Time: t, Value: period}

	return output
}

func (s *HilbertTransformerInstantaneousTrendLine) pushInput(value float64) {
	copy(s.input[1:], s.input[:s.inputLengthMin1])
	s.input[0] = value
}

func (s *HilbertTransformerInstantaneousTrendLine) calculateAverage() float64 {
	const half = 0.5

	// Compute the trend line as a simple average over the measured dominant cycle period.
	length := int(math.Floor(s.smoothedPeriod*s.cyclePartMultiplier + half))
	if length > s.inputLength {
		length = s.inputLength
	} else if length < 1 {
		length = 1
	}

	var sum float64
	for i := 0; i < length; i++ {
		sum += s.input[i]
	}

	return sum / float64(length)
}
