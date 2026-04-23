// Package goertzelspectrum implements MBST's Goertzel Spectrum heatmap indicator.
//
// The Goertzel Spectrum displays a power heatmap of the cyclic activity over a
// configurable cycle-period range using the Goertzel algorithm. It supports
// first- and second-order Goertzel estimators, optional spectral-dilation
// compensation, a fast-attack slow-decay automatic gain control, and either
// floating or fixed (0-clamped) intensity normalization.
//
// Reference: MBST Mbs.Trading.Indicators.SpectralAnalysis.GoertzelSpectrum.
package goertzelspectrum

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
)

// GoertzelSpectrum is MBST's Goertzel Spectrum heatmap indicator.
type GoertzelSpectrum struct {
	mu                    sync.RWMutex
	mnemonic              string
	description           string
	estimator             *estimator
	windowCount           int
	lastIndex             int
	primed                bool
	floatingNormalization bool
	minParameterValue     float64
	maxParameterValue     float64
	parameterResolution   float64
	barFunc               entities.BarFunc
	quoteFunc             entities.QuoteFunc
	tradeFunc             entities.TradeFunc
}

// NewGoertzelSpectrumDefault returns an instance created with default parameters.
func NewGoertzelSpectrumDefault() (*GoertzelSpectrum, error) {
	return NewGoertzelSpectrumParams(&Params{})
}

// NewGoertzelSpectrumParams returns an instance created with the supplied parameters.
//
//nolint:funlen,cyclop
func NewGoertzelSpectrumParams(p *Params) (*GoertzelSpectrum, error) {
	const (
		invalid           = "invalid goertzel spectrum parameters"
		fmtLength         = "%s: Length should be >= 2"
		fmtMinPeriod      = "%s: MinPeriod should be >= 2"
		fmtMaxPeriod      = "%s: MaxPeriod should be > MinPeriod"
		fmtNyquist        = "%s: MaxPeriod should be <= 2 * Length"
		fmtResolution     = "%s: SpectrumResolution should be >= 1"
		fmtAgc            = "%s: AutomaticGainControlDecayFactor should be in (0, 1)"
		fmtw              = "%s: %w"
		descrPrefix       = "Goertzel spectrum "
		defLength         = 64
		defMinPeriod      = 2.0
		defMaxPeriod      = 64.0
		defSpectrumResult = 1
		defAgcDecayFactor = 0.991
		agcDecayEpsilon   = 1e-12
	)

	cfg := *p

	if cfg.Length == 0 {
		cfg.Length = defLength
	}

	if cfg.MinPeriod == 0 {
		cfg.MinPeriod = defMinPeriod
	}

	if cfg.MaxPeriod == 0 {
		cfg.MaxPeriod = defMaxPeriod
	}

	if cfg.SpectrumResolution == 0 {
		cfg.SpectrumResolution = defSpectrumResult
	}

	if cfg.AutomaticGainControlDecayFactor == 0 {
		cfg.AutomaticGainControlDecayFactor = defAgcDecayFactor
	}

	// Resolve the inverted-sentinel bool flags to their MBST-semantic form.
	sdcOn := !cfg.DisableSpectralDilationCompensation
	agcOn := !cfg.DisableAutomaticGainControl
	floatingNorm := !cfg.FixedNormalization

	if cfg.Length < 2 {
		return nil, fmt.Errorf(fmtLength, invalid)
	}

	if cfg.MinPeriod < 2 {
		return nil, fmt.Errorf(fmtMinPeriod, invalid)
	}

	if cfg.MaxPeriod <= cfg.MinPeriod {
		return nil, fmt.Errorf(fmtMaxPeriod, invalid)
	}

	if cfg.MaxPeriod > 2*float64(cfg.Length) {
		return nil, fmt.Errorf(fmtNyquist, invalid)
	}

	if cfg.SpectrumResolution < 1 {
		return nil, fmt.Errorf(fmtResolution, invalid)
	}

	if agcOn &&
		(cfg.AutomaticGainControlDecayFactor <= 0 || cfg.AutomaticGainControlDecayFactor >= 1) {
		return nil, fmt.Errorf(fmtAgc, invalid)
	}

	// GoertzelSpectrum mirrors MBST's reference: BarMedianPrice default.
	bc := cfg.BarComponent
	if bc == 0 {
		bc = entities.BarMedianPrice
	}

	qc := cfg.QuoteComponent
	if qc == 0 {
		qc = entities.DefaultQuoteComponent
	}

	tc := cfg.TradeComponent
	if tc == 0 {
		tc = entities.DefaultTradeComponent
	}

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

	componentMnemonic := core.ComponentTripleMnemonic(bc, qc, tc)

	flags := buildFlagTags(&cfg, sdcOn, agcOn, floatingNorm, defAgcDecayFactor, agcDecayEpsilon)
	mnemonic := fmt.Sprintf("gspect(%d, %g, %g, %d%s%s)",
		cfg.Length, cfg.MinPeriod, cfg.MaxPeriod, cfg.SpectrumResolution, flags, componentMnemonic)

	est := newEstimator(
		cfg.Length, cfg.MinPeriod, cfg.MaxPeriod, cfg.SpectrumResolution,
		cfg.IsFirstOrder, sdcOn, agcOn, cfg.AutomaticGainControlDecayFactor,
	)

	return &GoertzelSpectrum{
		mnemonic:              mnemonic,
		description:           descrPrefix + mnemonic,
		estimator:             est,
		lastIndex:             cfg.Length - 1,
		floatingNormalization: floatingNorm,
		minParameterValue:     cfg.MinPeriod,
		maxParameterValue:     cfg.MaxPeriod,
		parameterResolution:   float64(cfg.SpectrumResolution),
		barFunc:               barFunc,
		quoteFunc:             quoteFunc,
		tradeFunc:             tradeFunc,
	}, nil
}

// buildFlagTags encodes non-default boolean/decay settings as terse override-only tags.
// Returns an empty string when all flags are at their defaults.
func buildFlagTags(cfg *Params, sdcOn, agcOn, floatingNorm bool, defAgc, eps float64) string {
	var s string

	if cfg.IsFirstOrder {
		s += ", fo"
	}

	if !sdcOn {
		s += ", no-sdc"
	}

	if !agcOn {
		s += ", no-agc"
	}

	if agcOn && math.Abs(cfg.AutomaticGainControlDecayFactor-defAgc) > eps {
		s += fmt.Sprintf(", agc=%g", cfg.AutomaticGainControlDecayFactor)
	}

	if !floatingNorm {
		s += ", no-fn"
	}

	return s
}

// IsPrimed indicates whether the indicator is primed.
func (s *GoertzelSpectrum) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *GoertzelSpectrum) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.GoertzelSpectrum,
		s.mnemonic,
		s.description,
		[]core.OutputText{
			{Mnemonic: s.mnemonic, Description: s.description},
		},
	)
}

// Update feeds the next sample to the engine and returns the heatmap column.
//
// Before priming the heatmap is empty (with the indicator's parameter axis).
// On a NaN input sample the state is left unchanged and an empty heatmap is
// returned.
func (s *GoertzelSpectrum) Update(sample float64, t time.Time) *outputs.Heatmap {
	s.mu.Lock()
	defer s.mu.Unlock()

	if math.IsNaN(sample) {
		return outputs.NewEmptyHeatmap(t, s.minParameterValue, s.maxParameterValue, s.parameterResolution)
	}

	window := s.estimator.inputSeries

	if s.primed {
		copy(window[:s.lastIndex], window[1:])
		window[s.lastIndex] = sample
	} else {
		window[s.windowCount] = sample
		s.windowCount++

		if s.windowCount == s.estimator.length {
			s.primed = true
		}
	}

	if !s.primed {
		return outputs.NewEmptyHeatmap(t, s.minParameterValue, s.maxParameterValue, s.parameterResolution)
	}

	s.estimator.calculate()

	lengthSpectrum := s.estimator.lengthSpectrum

	var minRef float64
	if s.floatingNormalization {
		minRef = s.estimator.spectrumMin
	}

	maxRef := s.estimator.spectrumMax
	spectrumRange := maxRef - minRef

	// MBST fills spectrum[0] at MaxPeriod and spectrum[last] at MinPeriod.
	// The heatmap axis runs MinPeriod -> MaxPeriod, so reverse on output.
	values := make([]float64, lengthSpectrum)
	valueMin := math.Inf(1)
	valueMax := math.Inf(-1)

	for i := 0; i < lengthSpectrum; i++ {
		v := (s.estimator.spectrum[lengthSpectrum-1-i] - minRef) / spectrumRange
		values[i] = v

		if v < valueMin {
			valueMin = v
		}

		if v > valueMax {
			valueMax = v
		}
	}

	return outputs.NewHeatmap(t, s.minParameterValue, s.maxParameterValue, s.parameterResolution,
		valueMin, valueMax, values)
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *GoertzelSpectrum) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.updateEntity(sample.Time, sample.Value)
}

// UpdateBar updates the indicator given the next bar sample.
func (s *GoertzelSpectrum) UpdateBar(sample *entities.Bar) core.Output {
	return s.updateEntity(sample.Time, s.barFunc(sample))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *GoertzelSpectrum) UpdateQuote(sample *entities.Quote) core.Output {
	return s.updateEntity(sample.Time, s.quoteFunc(sample))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *GoertzelSpectrum) UpdateTrade(sample *entities.Trade) core.Output {
	return s.updateEntity(sample.Time, s.tradeFunc(sample))
}

func (s *GoertzelSpectrum) updateEntity(t time.Time, sample float64) core.Output {
	const length = 1

	heatmap := s.Update(sample, t)

	output := make([]any, length)
	output[0] = heatmap

	return output
}
