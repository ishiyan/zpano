package dominantcycle

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

// DominantCycle (Ehlers' dominant cycle) computes the instantaneous cycle period and phase
// derived from a Hilbert transformer cycle estimator.
//
// It exposes three outputs:
//
//   - RawPeriod: the raw instantaneous cycle period produced by the Hilbert transformer estimator.
//   - Period: the dominant cycle period obtained by additional EMA smoothing of the raw period.
//     Periodᵢ = α·RawPeriodᵢ + (1 − α)·Periodᵢ₋₁, 0 < α ≤ 1.
//   - Phase: the dominant cycle phase, in degrees.
//
// The smoothed data are multiplied by the real (cosine) component of the dominant cycle
// and independently by the imaginary (sine) component of the dominant cycle. The products
// are summed then over one full dominant cycle. The phase angle is computed as the arctangent
// of the ratio of the real part to the imaginary part.
//
// Reference:
// John Ehlers, Rocket Science for Traders, Wiley, 2001, 0471405671, pp 52-77.
type DominantCycle struct {
	mu                             sync.RWMutex
	mnemonicRawPeriod              string
	descriptionRawPeriod           string
	mnemonicPeriod                 string
	descriptionPeriod              string
	mnemonicPhase                  string
	descriptionPhase               string
	alphaEmaPeriodAdditional       float64
	oneMinAlphaEmaPeriodAdditional float64
	smoothedPeriod                 float64
	smoothedPhase                  float64
	smoothedInput                  []float64
	smoothedInputLengthMin1        int
	htce                           hilberttransformer.CycleEstimator
	primed                         bool
	barFunc                        entities.BarFunc
	quoteFunc                      entities.QuoteFunc
	tradeFunc                      entities.TradeFunc
}

// NewDominantCycleDefault returns an instance of the indicator
// created using default values of the parameters.
func NewDominantCycleDefault() (*DominantCycle, error) {
	const (
		smoothingLength           = 4
		alphaEmaQuadratureInPhase = 0.2
		alphaEmaPeriod            = 0.2
		alphaEmaPeriodAdditional  = 0.33
		warmUpPeriod              = 100
	)

	return newDominantCycle(
		hilberttransformer.HomodyneDiscriminator,
		&hilberttransformer.CycleEstimatorParams{
			SmoothingLength:           smoothingLength,
			AlphaEmaQuadratureInPhase: alphaEmaQuadratureInPhase,
			AlphaEmaPeriod:            alphaEmaPeriod,
			WarmUpPeriod:              warmUpPeriod,
		},
		alphaEmaPeriodAdditional,
		0, 0, 0)
}

// NewDominantCycleParams returns an instance of the indicator created using supplied parameters.
func NewDominantCycleParams(p *Params) (*DominantCycle, error) {
	return newDominantCycle(
		p.EstimatorType, &p.EstimatorParams,
		p.AlphaEmaPeriodAdditional,
		p.BarComponent, p.QuoteComponent, p.TradeComponent)
}

func newDominantCycle(
	estimatorType hilberttransformer.CycleEstimatorType,
	estimatorParams *hilberttransformer.CycleEstimatorParams,
	alphaEmaPeriodAdditional float64,
	bc entities.BarComponent, qc entities.QuoteComponent, tc entities.TradeComponent,
) (*DominantCycle, error) {
	const (
		invalid  = "invalid dominant cycle parameters"
		fmta     = "%s: α for additional smoothing should be in range (0, 1]"
		fmtw     = "%s: %w"
		fmtnRaw  = "dcp-raw(%.3f%s%s)"
		fmtnPer  = "dcp(%.3f%s%s)"
		fmtnPha  = "dcph(%.3f%s%s)"
		four     = 4
		alpha    = 0.2
		descrRaw = "Dominant cycle raw period "
		descrPer = "Dominant cycle period "
		descrPha = "Dominant cycle phase "
	)

	if alphaEmaPeriodAdditional <= 0. || alphaEmaPeriodAdditional > 1. {
		return nil, fmt.Errorf(fmta, invalid)
	}

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

	mnemonicRawPeriod := fmt.Sprintf(fmtnRaw, alphaEmaPeriodAdditional, estimatorMoniker, componentMnemonic)
	mnemonicPeriod := fmt.Sprintf(fmtnPer, alphaEmaPeriodAdditional, estimatorMoniker, componentMnemonic)
	mnemonicPhase := fmt.Sprintf(fmtnPha, alphaEmaPeriodAdditional, estimatorMoniker, componentMnemonic)

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

	maxPeriod := estimator.MaxPeriod()

	return &DominantCycle{
		mnemonicRawPeriod:              mnemonicRawPeriod,
		descriptionRawPeriod:           descrRaw + mnemonicRawPeriod,
		mnemonicPeriod:                 mnemonicPeriod,
		descriptionPeriod:              descrPer + mnemonicPeriod,
		mnemonicPhase:                  mnemonicPhase,
		descriptionPhase:               descrPha + mnemonicPhase,
		alphaEmaPeriodAdditional:       alphaEmaPeriodAdditional,
		oneMinAlphaEmaPeriodAdditional: 1. - alphaEmaPeriodAdditional,
		htce:                           estimator,
		smoothedInput:                  make([]float64, maxPeriod),
		smoothedInputLengthMin1:        maxPeriod - 1,
		barFunc:                        barFunc,
		quoteFunc:                      quoteFunc,
		tradeFunc:                      tradeFunc,
	}, nil
}

// IsPrimed indicates whether an indicator is primed.
func (s *DominantCycle) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// SmoothedPrice returns the current WMA-smoothed price value produced by the
// underlying Hilbert transformer cycle estimator. Returns NaN if the indicator
// is not yet primed.
//
// This accessor is intended for composite indicators (e.g. TrendCycleMode) that
// wrap a DominantCycle and need to consult the same smoothed input stream that
// drives the dominant-cycle computation.
func (s *DominantCycle) SmoothedPrice() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.primed {
		return math.NaN()
	}

	return s.htce.Smoothed()
}

// MaxPeriod returns the maximum cycle period supported by the underlying
// Hilbert transformer cycle estimator (also the size of the internal
// smoothed-input buffer).
func (s *DominantCycle) MaxPeriod() int {
	return s.htce.MaxPeriod()
}

// Metadata describes an output data of the indicator.
func (s *DominantCycle) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.DominantCycle,
		s.mnemonicPeriod,
		s.descriptionPeriod,
		[]core.OutputText{
			{Mnemonic: s.mnemonicRawPeriod, Description: s.descriptionRawPeriod},
			{Mnemonic: s.mnemonicPeriod, Description: s.descriptionPeriod},
			{Mnemonic: s.mnemonicPhase, Description: s.descriptionPhase},
		},
	)
}

// Update updates the value of the indicator given the next sample, returning the
// (rawPeriod, period, phase) triple. Returns NaN values if the indicator is not yet primed.
func (s *DominantCycle) Update(sample float64) (float64, float64, float64) {
	if math.IsNaN(sample) {
		return sample, sample, sample
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.htce.Update(sample)
	s.pushSmoothedInput(s.htce.Smoothed())

	if s.primed {
		s.smoothedPeriod = s.alphaEmaPeriodAdditional*s.htce.Period() +
			s.oneMinAlphaEmaPeriodAdditional*s.smoothedPeriod
		s.calculateSmoothedPhase()

		return s.htce.Period(), s.smoothedPeriod, s.smoothedPhase
	}

	if s.htce.Primed() {
		s.primed = true
		s.smoothedPeriod = s.htce.Period()
		s.calculateSmoothedPhase()

		return s.htce.Period(), s.smoothedPeriod, s.smoothedPhase
	}

	nan := math.NaN()

	return nan, nan, nan
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *DominantCycle) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.updateEntity(sample.Time, sample.Value)
}

// UpdateBar updates the indicator given the next bar sample.
func (s *DominantCycle) UpdateBar(sample *entities.Bar) core.Output {
	return s.updateEntity(sample.Time, s.barFunc(sample))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *DominantCycle) UpdateQuote(sample *entities.Quote) core.Output {
	return s.updateEntity(sample.Time, s.quoteFunc(sample))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *DominantCycle) UpdateTrade(sample *entities.Trade) core.Output {
	return s.updateEntity(sample.Time, s.tradeFunc(sample))
}

func (s *DominantCycle) updateEntity(time time.Time, sample float64) core.Output {
	const length = 3

	rawPeriod, period, phase := s.Update(sample)

	output := make([]any, length)
	i := 0
	output[i] = entities.Scalar{Time: time, Value: rawPeriod}
	i++
	output[i] = entities.Scalar{Time: time, Value: period}
	i++
	output[i] = entities.Scalar{Time: time, Value: phase}

	return output
}

func (s *DominantCycle) pushSmoothedInput(value float64) {
	copy(s.smoothedInput[1:], s.smoothedInput[:s.smoothedInputLengthMin1])
	s.smoothedInput[0] = value
}

//nolint:cyclop
func (s *DominantCycle) calculateSmoothedPhase() {
	const (
		rad2deg    = 180.0 / math.Pi
		twoPi      = 2.0 * math.Pi
		epsilon    = 0.01
		half       = 0.5
		ninety     = 90.0
		oneEighty  = 180.0
		threeSixty = 360.0
	)

	// The smoothed data are multiplied by the real (cosine) component of the dominant cycle
	// and independently by the imaginary (sine) component of the dominant cycle.
	// The products are summed then over one full dominant cycle.
	length := int(math.Floor(s.smoothedPeriod + half))
	if length > s.smoothedInputLengthMin1 {
		length = s.smoothedInputLengthMin1
	}

	var realPart, imagPart float64

	for i := 0; i < length; i++ {
		temp := twoPi * float64(i) / float64(length)
		smoothed := s.smoothedInput[i]
		realPart += smoothed * math.Sin(temp)
		imagPart += smoothed * math.Cos(temp)
	}

	// We compute the phase angle as the arctangent of the ratio of the real part to the imaginary part.
	// The phase increases from the left to right.
	previous := s.smoothedPhase
	// phase := math.Atan2(realPart, imagPart) * rad2deg
	phase := math.Atan(realPart/imagPart) * rad2deg
	if math.IsNaN(phase) || math.IsInf(phase, 0) {
		phase = previous
	}

	if math.Abs(imagPart) <= epsilon {
		if realPart > 0 {
			phase += ninety
		} else if realPart < 0 {
			phase -= ninety
		}
	}

	// Introduce the 90 degree reference shift.
	phase += ninety
	// Compensate for one bar lag of the smoothed input price (weighted moving average).
	// This is done by adding the phase corresponding to a 1-bar lag of the smoothed dominant cycle period.
	phase += threeSixty / s.smoothedPeriod
	// Resolve phase ambiguity when the imaginary part is negative to provide a 360 degree phase presentation.
	if imagPart < 0 {
		phase += oneEighty
	}
	// Perform the cycle wraparound.
	if phase > threeSixty {
		phase -= threeSixty
	}

	s.smoothedPhase = phase
}
