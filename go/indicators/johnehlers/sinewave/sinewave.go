package sinewave

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
	"zpano/indicators/johnehlers/dominantcycle"
	"zpano/indicators/johnehlers/hilberttransformer"
)

// Deg2Rad converts degrees to radians.
const deg2Rad = math.Pi / 180.0

// SineWave is the Ehlers' Sine Wave indicator.
//
// It exposes five outputs:
//
//   - Value: the sine wave value, sin(phase·Deg2Rad).
//   - Lead: the sine wave lead value, sin((phase+45)·Deg2Rad).
//   - Band: a band with Upper=Value and Lower=Lead.
//   - DominantCyclePeriod: the smoothed dominant cycle period.
//   - DominantCyclePhase: the dominant cycle phase, in degrees.
//
// Reference:
// John Ehlers, Rocket Science for Traders, Wiley, 2001, 0471405671, pp 95-105.
type SineWave struct {
	mu                 sync.RWMutex
	mnemonic           string
	description        string
	mnemonicLead       string
	descriptionLead    string
	mnemonicBand       string
	descriptionBand    string
	mnemonicDCP        string
	descriptionDCP     string
	mnemonicDCPhase    string
	descriptionDCPhase string
	dc                 *dominantcycle.DominantCycle
	primed             bool
	value              float64
	lead               float64
	barFunc            entities.BarFunc
	quoteFunc          entities.QuoteFunc
	tradeFunc          entities.TradeFunc
}

// NewSineWaveDefault returns an instance of the indicator
// created using default values of the parameters.
func NewSineWaveDefault() (*SineWave, error) {
	const (
		smoothingLength           = 4
		alphaEmaQuadratureInPhase = 0.2
		alphaEmaPeriod            = 0.2
		alphaEmaPeriodAdditional  = 0.33
		warmUpPeriod              = 100
	)

	return newSineWave(
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

// NewSineWaveParams returns an instance of the indicator created using supplied parameters.
func NewSineWaveParams(p *Params) (*SineWave, error) {
	return newSineWave(
		p.EstimatorType, &p.EstimatorParams,
		p.AlphaEmaPeriodAdditional,
		p.BarComponent, p.QuoteComponent, p.TradeComponent)
}

//nolint:funlen
func newSineWave(
	estimatorType hilberttransformer.CycleEstimatorType,
	estimatorParams *hilberttransformer.CycleEstimatorParams,
	alphaEmaPeriodAdditional float64,
	bc entities.BarComponent, qc entities.QuoteComponent, tc entities.TradeComponent,
) (*SineWave, error) {
	const (
		invalid    = "invalid sine wave parameters"
		fmta       = "%s: α for additional smoothing should be in range (0, 1]"
		fmtw       = "%s: %w"
		fmtnValue  = "sw(%.3f%s%s)"
		fmtnLead   = "sw-lead(%.3f%s%s)"
		fmtnBand   = "sw-band(%.3f%s%s)"
		fmtnDCP    = "dcp(%.3f%s%s)"
		fmtnDCPha  = "dcph(%.3f%s%s)"
		four       = 4
		alpha      = 0.2
		descrValue = "Sine wave "
		descrLead  = "Sine wave lead "
		descrBand  = "Sine wave band "
		descrDCP   = "Dominant cycle period "
		descrDCPha = "Dominant cycle phase "
	)

	if alphaEmaPeriodAdditional <= 0. || alphaEmaPeriodAdditional > 1. {
		return nil, fmt.Errorf(fmta, invalid)
	}

	// Resolve defaults for component functions.
	// SineWave defaults to BarMedianPrice (not the framework default), so it always shows in the mnemonic.
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

	// Compose the estimator moniker (same logic as DominantCycle).
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

	mnemonicValue := fmt.Sprintf(fmtnValue, alphaEmaPeriodAdditional, estimatorMoniker, componentMnemonic)
	mnemonicLead := fmt.Sprintf(fmtnLead, alphaEmaPeriodAdditional, estimatorMoniker, componentMnemonic)
	mnemonicBand := fmt.Sprintf(fmtnBand, alphaEmaPeriodAdditional, estimatorMoniker, componentMnemonic)
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

	nan := math.NaN()

	return &SineWave{
		mnemonic:           mnemonicValue,
		description:        descrValue + mnemonicValue,
		mnemonicLead:       mnemonicLead,
		descriptionLead:    descrLead + mnemonicLead,
		mnemonicBand:       mnemonicBand,
		descriptionBand:    descrBand + mnemonicBand,
		mnemonicDCP:        mnemonicDCP,
		descriptionDCP:     descrDCP + mnemonicDCP,
		mnemonicDCPhase:    mnemonicDCPha,
		descriptionDCPhase: descrDCPha + mnemonicDCPha,
		dc:                 dc,
		value:              nan,
		lead:               nan,
		barFunc:            barFunc,
		quoteFunc:          quoteFunc,
		tradeFunc:          tradeFunc,
	}, nil
}

// IsPrimed indicates whether an indicator is primed.
func (s *SineWave) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes an output data of the indicator.
func (s *SineWave) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.SineWave,
		s.mnemonic,
		s.description,
		[]core.OutputText{
			{Mnemonic: s.mnemonic, Description: s.description},
			{Mnemonic: s.mnemonicLead, Description: s.descriptionLead},
			{Mnemonic: s.mnemonicBand, Description: s.descriptionBand},
			{Mnemonic: s.mnemonicDCP, Description: s.descriptionDCP},
			{Mnemonic: s.mnemonicDCPhase, Description: s.descriptionDCPhase},
		},
	)
}

// Update updates the value of the indicator given the next sample, returning
// the (value, lead, period, phase) tuple. Returns NaN values if the indicator is not yet primed.
func (s *SineWave) Update(sample float64) (float64, float64, float64, float64) {
	if math.IsNaN(sample) {
		return sample, sample, sample, sample
	}

	// Delegate to the inner DominantCycle. Its own mutex is independent.
	_, period, phase := s.dc.Update(sample)

	s.mu.Lock()
	defer s.mu.Unlock()

	if math.IsNaN(phase) {
		nan := math.NaN()

		return nan, nan, nan, nan
	}

	const leadOffset = 45.0

	s.primed = true
	s.value = math.Sin(phase * deg2Rad)
	s.lead = math.Sin((phase + leadOffset) * deg2Rad)

	return s.value, s.lead, period, phase
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *SineWave) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.updateEntity(sample.Time, sample.Value)
}

// UpdateBar updates the indicator given the next bar sample.
func (s *SineWave) UpdateBar(sample *entities.Bar) core.Output {
	return s.updateEntity(sample.Time, s.barFunc(sample))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *SineWave) UpdateQuote(sample *entities.Quote) core.Output {
	return s.updateEntity(sample.Time, s.quoteFunc(sample))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *SineWave) UpdateTrade(sample *entities.Trade) core.Output {
	return s.updateEntity(sample.Time, s.tradeFunc(sample))
}

func (s *SineWave) updateEntity(time time.Time, sample float64) core.Output {
	const length = 5

	value, lead, period, phase := s.Update(sample)

	output := make([]any, length)
	i := 0
	output[i] = entities.Scalar{Time: time, Value: value}
	i++
	output[i] = entities.Scalar{Time: time, Value: lead}
	i++
	output[i] = outputs.Band{Time: time, Upper: value, Lower: lead}
	i++
	output[i] = entities.Scalar{Time: time, Value: period}
	i++
	output[i] = entities.Scalar{Time: time, Value: phase}

	return output
}
