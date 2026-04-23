package mesaadaptivemovingaverage

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
	"zpano/indicators/johnehlers/hilberttransformer"
)

// MesaAdaptiveMovingAverage (Ehler's Mesa adaptive moving average, or Mother of All Moving Averages, MAMA)
// is an EMA with the smoothing factor, α, being changed with each new sample within the fast and the slow
// limit boundaries which are the constant parameters of MAMA:
//
// MAMAᵢ = αᵢPᵢ + (1 - αᵢ)*MAMAᵢ₋₁,  αs ≤ αᵢ ≤ αf
//
// The αf is the α of the fast (shortest, default suggested value 0.5 or 3 samples) limit boundary.
//
// The αs is the α of the slow (longest, default suggested value 0.05 or 39 samples) limit boundary.
//
// The concept of MAMA is to relate the phase rate of change, as measured by a Hilbert Transformer
// estimator, to the EMA smoothing factor α, thus making the EMA adaptive.
//
// The cycle phase is computed from the arctangent of the ratio of the Quadrature component to the
// InPhase component. The rate of change is obtained by taking the difference of successive phase
// measurements. The α is computed as the fast limit αf divided by the phase rate of change.
// Any time there is a negative phase rate of change the value of α is set to the fast limit αf;
// if the phase rate of change is large, the α is bounded at the slow limit αs.
//
// The Following Adaptive Moving Average (FAMA) is produced by applying the MAMA to the first
// MAMA indicator.
//
// By using an α in FAMA that is the half the value of the α in MAMA, the FAMA has steps in
// time synchronization with MAMA, but the vertical movement is not as great.
//
// As a result, MAMA and FAMA do not cross unless there has been a major change in the
// market direction. This suggests an adaptive moving average crossover system that is
// virtually free of whipsaw trades.
//
// Reference:
// John Ehlers, Rocket Science for Traders, Wiley, 2001, 0471405671, pp 177-184.
type MesaAdaptiveMovingAverage struct {
	mu              sync.RWMutex
	mnemonic        string
	description     string
	mnemonicFama    string
	descriptionFama string
	mnemonicBand    string
	descriptionBand string
	alphaFastLimit  float64
	alphaSlowLimit  float64
	previousPhase   float64
	mama            float64
	fama            float64
	htce            hilberttransformer.CycleEstimator
	isPhaseCached   bool
	primed          bool
	barFunc         entities.BarFunc
	quoteFunc       entities.QuoteFunc
	tradeFunc       entities.TradeFunc
}

// NewMesaAdaptiveMovingAverageDefault returns an instance of the indicator
// created using default values of the parameters.
func NewMesaAdaptiveMovingAverageDefault() (*MesaAdaptiveMovingAverage, error) {
	const (
		fastLimitLength           = 3
		slowLimitLength           = 39
		smoothingLength           = 4
		alphaEmaQuadratureInPhase = 0.2
		alphaEmaPeriod            = 0.2
		warmUpPeriod              = 0
	)

	return newMesaAdaptiveMovingAverage(
		hilberttransformer.HomodyneDiscriminator,
		&hilberttransformer.CycleEstimatorParams{
			SmoothingLength:           smoothingLength,
			AlphaEmaQuadratureInPhase: alphaEmaQuadratureInPhase,
			AlphaEmaPeriod:            alphaEmaPeriod,
			WarmUpPeriod:              warmUpPeriod,
		},
		fastLimitLength, slowLimitLength,
		math.NaN(), math.NaN(),
		0, 0, 0)
}

// NewMesaAdaptiveMovingAverageLength returns an instance of the indicator
// created using supplied parameters based on length.
func NewMesaAdaptiveMovingAverageLength(
	p *LengthParams,
) (*MesaAdaptiveMovingAverage, error) {
	return newMesaAdaptiveMovingAverage(
		p.EstimatorType, &p.EstimatorParams,
		p.FastLimitLength, p.SlowLimitLength,
		math.NaN(), math.NaN(),
		p.BarComponent, p.QuoteComponent, p.TradeComponent)
}

// NewMesaAdaptiveMovingAverageSmoothingFactor returns an instance of the indicator
// created using supplied parameters based on smoothing factor.
func NewMesaAdaptiveMovingAverageSmoothingFactor(
	p *SmoothingFactorParams,
) (*MesaAdaptiveMovingAverage, error) {
	return newMesaAdaptiveMovingAverage(
		p.EstimatorType, &p.EstimatorParams,
		0, 0,
		p.FastLimitSmoothingFactor, p.SlowLimitSmoothingFactor,
		p.BarComponent, p.QuoteComponent, p.TradeComponent)
}

//nolint:funlen,cyclop
func newMesaAdaptiveMovingAverage(
	estimatorType hilberttransformer.CycleEstimatorType,
	estimatorParams *hilberttransformer.CycleEstimatorParams,
	fastLimitLength int, slowLimitLength int,
	fastLimitSmoothingFactor float64, slowLimitSmoothingFactor float64,
	bc entities.BarComponent, qc entities.QuoteComponent, tc entities.TradeComponent,
) (*MesaAdaptiveMovingAverage, error) {
	const (
		invalid = "invalid mesa adaptive moving average parameters"
		fmtl    = "%s: %s length should be larger than 1"
		fmta    = "%s: %s smoothing factor should be in range [0, 1]"
		fmtw    = "%s: %w"
		fmtnl   = "mama(%d, %d%s%s)"
		fmtna   = "mama(%.3f, %.3f%s%s)"
		fmtnlf  = "fama(%d, %d%s%s)"
		fmtnaf  = "fama(%.3f, %.3f%s%s)"
		fmtnlb  = "mama-fama(%d, %d%s%s)"
		fmtnab  = "mama-fama(%.3f, %.3f%s%s)"
		two     = 2
		four    = 4
		alpha   = 0.2
		epsilon = 0.00000001
		flim    = "fast limit"
		slim    = "slow limit"
		descr   = "Mesa adaptive moving average "
	)

	var (
		mnemonic     string
		mnemonicFama string
		mnemonicBand string
		err          error
		barFunc      entities.BarFunc
		quoteFunc    entities.QuoteFunc
		tradeFunc    entities.TradeFunc
	)

	estimator, err := hilberttransformer.NewCycleEstimator(estimatorType, estimatorParams)
	if err != nil {
		return nil, err
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

	// Resolve defaults for component functions.
	// A zero value means "use default, don't show in mnemonic".
	if bc == 0 {
		bc = entities.DefaultBarComponent
	}

	if qc == 0 {
		qc = entities.DefaultQuoteComponent
	}

	if tc == 0 {
		tc = entities.DefaultTradeComponent
	}

	componentMnemonic := core.ComponentTripleMnemonic(bc, qc, tc)

	if math.IsNaN(fastLimitSmoothingFactor) { //nolint:nestif
		if fastLimitLength < two {
			return nil, fmt.Errorf(fmtl, invalid, flim)
		}

		if slowLimitLength < two {
			return nil, fmt.Errorf(fmtl, invalid, slim)
		}

		fastLimitSmoothingFactor = two / float64(1+fastLimitLength)
		slowLimitSmoothingFactor = two / float64(1+slowLimitLength)

		mnemonic = fmt.Sprintf(fmtnl,
			fastLimitLength, slowLimitLength, estimatorMoniker, componentMnemonic)
		mnemonicFama = fmt.Sprintf(fmtnlf,
			fastLimitLength, slowLimitLength, estimatorMoniker, componentMnemonic)
		mnemonicBand = fmt.Sprintf(fmtnlb,
			fastLimitLength, slowLimitLength, estimatorMoniker, componentMnemonic)
	} else {
		if fastLimitSmoothingFactor < 0. || fastLimitSmoothingFactor > 1. {
			return nil, fmt.Errorf(fmta, invalid, flim)
		}

		if slowLimitSmoothingFactor < 0. || slowLimitSmoothingFactor > 1. {
			return nil, fmt.Errorf(fmta, invalid, slim)
		}

		if fastLimitSmoothingFactor < epsilon {
			fastLimitSmoothingFactor = epsilon
		}

		if slowLimitSmoothingFactor < epsilon {
			slowLimitSmoothingFactor = epsilon
		}

		mnemonic = fmt.Sprintf(fmtna,
			fastLimitSmoothingFactor, slowLimitSmoothingFactor, estimatorMoniker, componentMnemonic)
		mnemonicFama = fmt.Sprintf(fmtnaf,
			fastLimitSmoothingFactor, slowLimitSmoothingFactor, estimatorMoniker, componentMnemonic)
		mnemonicBand = fmt.Sprintf(fmtnab,
			fastLimitSmoothingFactor, slowLimitSmoothingFactor, estimatorMoniker, componentMnemonic)
	}

	if barFunc, err = entities.BarComponentFunc(bc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	if quoteFunc, err = entities.QuoteComponentFunc(qc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	if tradeFunc, err = entities.TradeComponentFunc(tc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	return &MesaAdaptiveMovingAverage{
		mnemonic:        mnemonic,
		description:     descr + mnemonic,
		mnemonicFama:    mnemonicFama,
		descriptionFama: descr + mnemonicFama,
		mnemonicBand:    mnemonicBand,
		descriptionBand: descr + mnemonicBand,
		alphaFastLimit:  fastLimitSmoothingFactor,
		alphaSlowLimit:  slowLimitSmoothingFactor,
		htce:            estimator,
		barFunc:         barFunc,
		quoteFunc:       quoteFunc,
		tradeFunc:       tradeFunc,
	}, nil
}

// IsPrimed indicates whether an indicator is primed.
func (s *MesaAdaptiveMovingAverage) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes an output data of the indicator.
func (s *MesaAdaptiveMovingAverage) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.MesaAdaptiveMovingAverage,
		s.mnemonic,
		s.description,
		[]core.OutputText{
			{Mnemonic: s.mnemonic, Description: s.description},
			{Mnemonic: s.mnemonicFama, Description: s.descriptionFama},
			{Mnemonic: s.mnemonicBand, Description: s.descriptionBand},
		},
	)
}

// Update updates the value of the moving average given the next sample.
func (s *MesaAdaptiveMovingAverage) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.htce.Update(sample)

	if s.primed {
		return s.calculate(sample)
	}

	if s.htce.Primed() {
		if s.isPhaseCached {
			s.primed = true

			return s.calculate(sample)
		}

		s.isPhaseCached = true
		s.previousPhase = s.calculatePhase()
		s.mama = sample
		s.fama = sample
	}

	return math.NaN()
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *MesaAdaptiveMovingAverage) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.updateEntity(sample.Time, sample.Value)
}

// UpdateBar updates the indicator given the next bar sample.
func (s *MesaAdaptiveMovingAverage) UpdateBar(sample *entities.Bar) core.Output {
	return s.updateEntity(sample.Time, s.barFunc(sample))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *MesaAdaptiveMovingAverage) UpdateQuote(sample *entities.Quote) core.Output {
	return s.updateEntity(sample.Time, s.quoteFunc(sample))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *MesaAdaptiveMovingAverage) UpdateTrade(sample *entities.Trade) core.Output {
	return s.updateEntity(sample.Time, s.tradeFunc(sample))
}

func (s *MesaAdaptiveMovingAverage) updateEntity(time time.Time, sample float64) core.Output {
	const length = 3

	output := make([]any, length)
	mama := s.Update(sample)

	fama := s.fama
	if math.IsNaN(mama) {
		fama = math.NaN()
	}

	i := 0
	output[i] = entities.Scalar{Time: time, Value: mama}
	i++
	output[i] = entities.Scalar{Time: time, Value: fama}
	i++
	output[i] = outputs.Band{Time: time, Upper: mama, Lower: fama}

	return output
}

func (s *MesaAdaptiveMovingAverage) calculatePhase() float64 {
	if s.htce.InPhase() == 0 {
		return s.previousPhase
	}

	const rad2deg = 180.0 / math.Pi

	// The cycle phase is computed from the arctangent of the ratio
	// of the Quadrature component to the InPhase component.
	// phase := math.Atan2(s.htce.InPhase(), s.htce.Quadrature()) * rad2deg
	phase := math.Atan(s.htce.Quadrature()/s.htce.InPhase()) * rad2deg
	if !math.IsNaN(phase) && !math.IsInf(phase, 0) {
		return phase
	}

	return s.previousPhase
}

func (s *MesaAdaptiveMovingAverage) calculateMama(sample float64) float64 {
	phase := s.calculatePhase()

	// The phase rate of change is obtained by taking the
	// difference of successive previousPhase measurements.
	phaseRateOfChange := s.previousPhase - phase
	s.previousPhase = phase

	// Any negative rate change is theoretically impossible
	// because phase must advance as the time increases.
	// We therefore limit all rate changes of phase to be
	// no less than unity.
	if phaseRateOfChange < 1 {
		phaseRateOfChange = 1
	}

	// The α is computed as the fast limit divided
	// by the phase rate of change.
	alpha := min(max(s.alphaFastLimit/phaseRateOfChange, s.alphaSlowLimit), s.alphaFastLimit)

	s.mama = alpha*sample + (1.0-alpha)*s.mama

	return alpha
}

func (s *MesaAdaptiveMovingAverage) calculate(sample float64) float64 {
	const two = 2

	alpha := s.calculateMama(sample) / two
	s.fama = alpha*s.mama + (1.0-alpha)*s.fama

	return s.mama
}
