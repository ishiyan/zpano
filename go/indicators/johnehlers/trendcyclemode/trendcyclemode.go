package trendcyclemode

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/johnehlers/dominantcycle"
	"zpano/indicators/johnehlers/hilberttransformer"
)

// Deg2Rad converts degrees to radians.
const deg2Rad = math.Pi / 180.0

// TrendCycleMode is Ehlers' Trend-versus-Cycle Mode indicator.
//
// It wraps a DominantCycle (for instantaneous period / phase / WMA-smoothed price)
// and exposes eight outputs:
//
//   - Value: +1 in trend mode, -1 in cycle mode.
//   - IsTrendMode: 1 if trend mode is declared, 0 otherwise.
//   - IsCycleMode: 1 if cycle mode is declared, 0 otherwise.
//   - InstantaneousTrendLine: WMA-smoothed trend line.
//   - SineWave: sin(phase·Deg2Rad).
//   - SineWaveLead: sin((phase+45)·Deg2Rad).
//   - DominantCyclePeriod: smoothed dominant cycle period.
//   - DominantCyclePhase: dominant cycle phase, in degrees.
//
// Reference:
// John Ehlers, Rocket Science for Traders, Wiley, 2001, 0471405671, pp 113-118.
type TrendCycleMode struct {
	mu                             sync.RWMutex
	mnemonic                       string
	description                    string
	mnemonicTrend                  string
	descriptionTrend               string
	mnemonicCycle                  string
	descriptionCycle               string
	mnemonicITL                    string
	descriptionITL                 string
	mnemonicSine                   string
	descriptionSine                string
	mnemonicSineLead               string
	descriptionSineLead            string
	mnemonicDCP                    string
	descriptionDCP                 string
	mnemonicDCPhase                string
	descriptionDCPhase             string
	dc                             *dominantcycle.DominantCycle
	cyclePartMultiplier            float64
	separationPercentage           float64
	separationFactor               float64
	trendLineSmoothingLength       int
	coeff0                         float64
	coeff1                         float64
	coeff2                         float64
	coeff3                         float64
	trendline                      float64
	trendAverage1                  float64
	trendAverage2                  float64
	trendAverage3                  float64
	sinWave                        float64
	sinWaveLead                    float64
	previousDcPhase                float64
	previousSineLeadWaveDifference float64
	samplesInTrend                 int
	isTrendMode                    bool
	input                          []float64
	inputLength                    int
	inputLengthMin1                int
	primed                         bool
	barFunc                        entities.BarFunc
	quoteFunc                      entities.QuoteFunc
	tradeFunc                      entities.TradeFunc
}

// NewTrendCycleModeDefault returns an instance of the indicator
// created using default values of the parameters.
func NewTrendCycleModeDefault() (*TrendCycleMode, error) {
	const (
		smoothingLength           = 4
		alphaEmaQuadratureInPhase = 0.2
		alphaEmaPeriod            = 0.2
		alphaEmaPeriodAdditional  = 0.33
		warmUpPeriod              = 100
		trendLineSmoothingLength  = 4
		cyclePartMultiplier       = 1.0
		separationPercentage      = 1.5
	)

	return newTrendCycleMode(
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
		separationPercentage,
		0, 0, 0)
}

// NewTrendCycleModeParams returns an instance of the indicator created using supplied parameters.
func NewTrendCycleModeParams(p *Params) (*TrendCycleMode, error) {
	return newTrendCycleMode(
		p.EstimatorType, &p.EstimatorParams,
		p.AlphaEmaPeriodAdditional,
		p.TrendLineSmoothingLength,
		p.CyclePartMultiplier,
		p.SeparationPercentage,
		p.BarComponent, p.QuoteComponent, p.TradeComponent)
}

//nolint:funlen,cyclop
func newTrendCycleMode(
	estimatorType hilberttransformer.CycleEstimatorType,
	estimatorParams *hilberttransformer.CycleEstimatorParams,
	alphaEmaPeriodAdditional float64,
	trendLineSmoothingLength int,
	cyclePartMultiplier float64,
	separationPercentage float64,
	bc entities.BarComponent, qc entities.QuoteComponent, tc entities.TradeComponent,
) (*TrendCycleMode, error) {
	const (
		invalid       = "invalid trend cycle mode parameters"
		fmta          = "%s: α for additional smoothing should be in range (0, 1]"
		fmttlsl       = "%s: trend line smoothing length should be 2, 3, or 4"
		fmtcpm        = "%s: cycle part multiplier should be in range (0, 10]"
		fmtsep        = "%s: separation percentage should be in range (0, 100]"
		fmtw          = "%s: %w"
		fmtnValue     = "tcm(%.3f, %d, %.3f, %.3f%%%s%s)"
		fmtnTrend     = "tcm-trend(%.3f, %d, %.3f, %.3f%%%s%s)"
		fmtnCycle     = "tcm-cycle(%.3f, %d, %.3f, %.3f%%%s%s)"
		fmtnITL       = "tcm-itl(%.3f, %d, %.3f, %.3f%%%s%s)"
		fmtnSine      = "tcm-sine(%.3f, %d, %.3f, %.3f%%%s%s)"
		fmtnSineLead  = "tcm-sineLead(%.3f, %d, %.3f, %.3f%%%s%s)"
		fmtnDCP       = "dcp(%.3f%s%s)"
		fmtnDCPha     = "dcph(%.3f%s%s)"
		four          = 4
		alpha         = 0.2
		descrValue    = "Trend versus cycle mode "
		descrTrend    = "Trend versus cycle mode, is-trend flag "
		descrCycle    = "Trend versus cycle mode, is-cycle flag "
		descrITL      = "Trend versus cycle mode instantaneous trend line "
		descrSine     = "Trend versus cycle mode sine wave "
		descrSineLead = "Trend versus cycle mode sine wave lead "
		descrDCP      = "Dominant cycle period "
		descrDCPha    = "Dominant cycle phase "
		tlslTwo       = 2
		tlslThree     = 3
		tlslFour      = 4
		cpmMax        = 10.0
		sepMax        = 100.0
		sepDenom      = 100.0
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

	if separationPercentage <= 0. || separationPercentage > sepMax {
		return nil, fmt.Errorf(fmtsep, invalid)
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

	// Build the inner DominantCycle with explicit components.
	dcParams := &dominantcycle.Params{
		EstimatorType:            estimatorType,
		EstimatorParams:          *estimatorParams,
		AlphaEmaPeriodAdditional: alphaEmaPeriodAdditional,
		BarComponent:             bc,
		QuoteComponent:           qc,
		TradeComponent:           tc,
	}

	dc, err := dominantcycle.NewDominantCycleParams(dcParams)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	// Compose the estimator moniker (same logic as DominantCycle / HTITL).
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

	fmtArgs := []any{
		alphaEmaPeriodAdditional, trendLineSmoothingLength,
		cyclePartMultiplier, separationPercentage,
		estimatorMoniker, componentMnemonic,
	}

	mnemonicValue := fmt.Sprintf(fmtnValue, fmtArgs...)
	mnemonicTrend := fmt.Sprintf(fmtnTrend, fmtArgs...)
	mnemonicCycle := fmt.Sprintf(fmtnCycle, fmtArgs...)
	mnemonicITL := fmt.Sprintf(fmtnITL, fmtArgs...)
	mnemonicSine := fmt.Sprintf(fmtnSine, fmtArgs...)
	mnemonicSineLead := fmt.Sprintf(fmtnSineLead, fmtArgs...)
	mnemonicDCP := fmt.Sprintf(fmtnDCP, alphaEmaPeriodAdditional, estimatorMoniker, componentMnemonic)
	mnemonicDCPha := fmt.Sprintf(fmtnDCPha, alphaEmaPeriodAdditional, estimatorMoniker, componentMnemonic)

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

	maxPeriod := dc.MaxPeriod()
	nan := math.NaN()

	return &TrendCycleMode{
		mnemonic:                 mnemonicValue,
		description:              descrValue + mnemonicValue,
		mnemonicTrend:            mnemonicTrend,
		descriptionTrend:         descrTrend + mnemonicTrend,
		mnemonicCycle:            mnemonicCycle,
		descriptionCycle:         descrCycle + mnemonicCycle,
		mnemonicITL:              mnemonicITL,
		descriptionITL:           descrITL + mnemonicITL,
		mnemonicSine:             mnemonicSine,
		descriptionSine:          descrSine + mnemonicSine,
		mnemonicSineLead:         mnemonicSineLead,
		descriptionSineLead:      descrSineLead + mnemonicSineLead,
		mnemonicDCP:              mnemonicDCP,
		descriptionDCP:           descrDCP + mnemonicDCP,
		mnemonicDCPhase:          mnemonicDCPha,
		descriptionDCPhase:       descrDCPha + mnemonicDCPha,
		dc:                       dc,
		cyclePartMultiplier:      cyclePartMultiplier,
		separationPercentage:     separationPercentage,
		separationFactor:         separationPercentage / sepDenom,
		trendLineSmoothingLength: trendLineSmoothingLength,
		coeff0:                   c0,
		coeff1:                   c1,
		coeff2:                   c2,
		coeff3:                   c3,
		trendline:                nan,
		sinWave:                  nan,
		sinWaveLead:              nan,
		isTrendMode:              true,
		input:                    make([]float64, maxPeriod),
		inputLength:              maxPeriod,
		inputLengthMin1:          maxPeriod - 1,
		barFunc:                  barFunc,
		quoteFunc:                quoteFunc,
		tradeFunc:                tradeFunc,
	}, nil
}

// IsPrimed indicates whether an indicator is primed.
func (s *TrendCycleMode) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes an output data of the indicator.
func (s *TrendCycleMode) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.TrendCycleMode,
		s.mnemonic,
		s.description,
		[]core.OutputText{
			{Mnemonic: s.mnemonic, Description: s.description},
			{Mnemonic: s.mnemonicTrend, Description: s.descriptionTrend},
			{Mnemonic: s.mnemonicCycle, Description: s.descriptionCycle},
			{Mnemonic: s.mnemonicITL, Description: s.descriptionITL},
			{Mnemonic: s.mnemonicSine, Description: s.descriptionSine},
			{Mnemonic: s.mnemonicSineLead, Description: s.descriptionSineLead},
			{Mnemonic: s.mnemonicDCP, Description: s.descriptionDCP},
			{Mnemonic: s.mnemonicDCPhase, Description: s.descriptionDCPhase},
		},
	)
}

// Update updates the value of the indicator given the next sample, returning the
// (value, isTrendMode, isCycleMode, trendline, sineWave, sineWaveLead, period, phase) tuple.
// The isTrendMode / isCycleMode values are encoded as 1 / 0 scalars. Returns NaN for all
// outputs if the indicator is not yet primed.
//
//nolint:funlen,cyclop,gocognit
func (s *TrendCycleMode) Update(sample float64) (float64, float64, float64, float64, float64, float64, float64, float64) {
	if math.IsNaN(sample) {
		return sample, sample, sample, sample, sample, sample, sample, sample
	}

	// Delegate to the inner DominantCycle. Its own mutex is independent.
	_, period, phase := s.dc.Update(sample)
	// SmoothedPrice tracks the WMA smoothed price inside the Hilbert transformer.
	smoothedPrice := s.dc.SmoothedPrice()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.pushInput(sample)

	if s.primed {
		smoothedPeriod := period
		average := s.calculateTrendAverage(smoothedPeriod)
		s.trendline = s.coeff0*average + s.coeff1*s.trendAverage1 +
			s.coeff2*s.trendAverage2 + s.coeff3*s.trendAverage3
		s.trendAverage3 = s.trendAverage2
		s.trendAverage2 = s.trendAverage1
		s.trendAverage1 = average

		diff := s.calculateSineLeadWaveDifference(phase)

		// Condition 1: a cycle mode exists for the half-period of a dominant cycle
		// after the SineWave vs SineWaveLead crossing.
		s.isTrendMode = true

		if (diff > 0 && s.previousSineLeadWaveDifference < 0) ||
			(diff < 0 && s.previousSineLeadWaveDifference > 0) {
			s.isTrendMode = false
			s.samplesInTrend = 0
		}

		s.previousSineLeadWaveDifference = diff
		s.samplesInTrend++

		const half = 0.5
		if float64(s.samplesInTrend) < half*smoothedPeriod {
			s.isTrendMode = false
		}

		// Condition 2: cycle mode if the measured phase rate of change is more than 2/3
		// the phase rate of change of the dominant cycle (360/period) and less than 1.5 times it.
		phaseDelta := phase - s.previousDcPhase
		s.previousDcPhase = phase

		if math.Abs(smoothedPeriod) > epsilon {
			const (
				minFactor = 2.0 / 3.0
				maxFactor = 1.5
				fullCycle = 360.0
			)

			dcRate := fullCycle / smoothedPeriod
			if phaseDelta > minFactor*dcRate && phaseDelta < maxFactor*dcRate {
				s.isTrendMode = false
			}
		}

		// Condition 3: if the WMA smoothed price is separated by more than the separation
		// percentage from the instantaneous trend line, force the trend mode.
		if math.Abs(s.trendline) > epsilon &&
			math.Abs((smoothedPrice-s.trendline)/s.trendline) >= s.separationFactor {
			s.isTrendMode = true
		}

		return s.mode(), s.isTrendFloat(), s.isCycleFloat(), s.trendline, s.sinWave, s.sinWaveLead, period, phase
	}

	if s.dc.IsPrimed() {
		s.primed = true
		smoothedPeriod := period
		s.trendline = s.calculateTrendAverage(smoothedPeriod)
		s.trendAverage1 = s.trendline
		s.trendAverage2 = s.trendline
		s.trendAverage3 = s.trendline

		s.previousDcPhase = phase
		s.previousSineLeadWaveDifference = s.calculateSineLeadWaveDifference(phase)

		s.isTrendMode = true
		s.samplesInTrend++

		const half = 0.5
		if float64(s.samplesInTrend) < half*smoothedPeriod {
			s.isTrendMode = false
		}

		return s.mode(), s.isTrendFloat(), s.isCycleFloat(), s.trendline, s.sinWave, s.sinWaveLead, period, phase
	}

	nan := math.NaN()

	return nan, nan, nan, nan, nan, nan, nan, nan
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *TrendCycleMode) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.updateEntity(sample.Time, sample.Value)
}

// UpdateBar updates the indicator given the next bar sample.
func (s *TrendCycleMode) UpdateBar(sample *entities.Bar) core.Output {
	return s.updateEntity(sample.Time, s.barFunc(sample))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *TrendCycleMode) UpdateQuote(sample *entities.Quote) core.Output {
	return s.updateEntity(sample.Time, s.quoteFunc(sample))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *TrendCycleMode) UpdateTrade(sample *entities.Trade) core.Output {
	return s.updateEntity(sample.Time, s.tradeFunc(sample))
}

func (s *TrendCycleMode) updateEntity(t time.Time, sample float64) core.Output {
	const length = 8

	value, trend, cycle, itl, sine, sineLead, period, phase := s.Update(sample)

	output := make([]any, length)
	i := 0
	output[i] = entities.Scalar{Time: t, Value: value}
	i++
	output[i] = entities.Scalar{Time: t, Value: trend}
	i++
	output[i] = entities.Scalar{Time: t, Value: cycle}
	i++
	output[i] = entities.Scalar{Time: t, Value: itl}
	i++
	output[i] = entities.Scalar{Time: t, Value: sine}
	i++
	output[i] = entities.Scalar{Time: t, Value: sineLead}
	i++
	output[i] = entities.Scalar{Time: t, Value: period}
	i++
	output[i] = entities.Scalar{Time: t, Value: phase}

	return output
}

// epsilon is the floating-point equality threshold (matches C#'s double.Epsilon behaviour for zero checks).
const epsilon = 1e-308

func (s *TrendCycleMode) pushInput(value float64) {
	copy(s.input[1:], s.input[:s.inputLengthMin1])
	s.input[0] = value
}

func (s *TrendCycleMode) calculateTrendAverage(smoothedPeriod float64) float64 {
	const half = 0.5

	length := int(math.Floor(smoothedPeriod*s.cyclePartMultiplier + half))
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

func (s *TrendCycleMode) calculateSineLeadWaveDifference(phase float64) float64 {
	const leadOffset = 45.0

	p := phase * deg2Rad
	s.sinWave = math.Sin(p)
	s.sinWaveLead = math.Sin(p + leadOffset*deg2Rad)

	return s.sinWave - s.sinWaveLead
}

func (s *TrendCycleMode) mode() float64 {
	if s.isTrendMode {
		return 1.0
	}

	return -1.0
}

func (s *TrendCycleMode) isTrendFloat() float64 {
	if s.isTrendMode {
		return 1.0
	}

	return 0.0
}

func (s *TrendCycleMode) isCycleFloat() float64 {
	if s.isTrendMode {
		return 0.0
	}

	return 1.0
}
