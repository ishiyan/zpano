// Package maximumentropyspectrum implements MBST's Maximum Entropy Spectrum heatmap indicator.
//
// The Maximum Entropy Spectrum (MESPECT) displays a power heatmap of the cyclic
// activity over a configurable cycle-period range using Burg's maximum-entropy
// auto-regressive method. It supports a fast-attack slow-decay automatic gain
// control and either floating or fixed (0-clamped) intensity normalization.
//
// Reference: MBST Mbs.Trading.Indicators.SpectralAnalysis.MaximumEntropySpectrum.
package maximumentropyspectrum

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

// MaximumEntropySpectrum is MBST's Maximum Entropy Spectrum heatmap indicator.
type MaximumEntropySpectrum struct {
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

// NewMaximumEntropySpectrumDefault returns an instance created with default parameters.
func NewMaximumEntropySpectrumDefault() (*MaximumEntropySpectrum, error) {
	return NewMaximumEntropySpectrumParams(&Params{})
}

// NewMaximumEntropySpectrumParams returns an instance created with the supplied parameters.
//
//nolint:funlen,cyclop
func NewMaximumEntropySpectrumParams(p *Params) (*MaximumEntropySpectrum, error) {
	const (
		invalid           = "invalid maximum entropy spectrum parameters"
		fmtLength         = "%s: Length should be >= 2"
		fmtDegree         = "%s: Degree should be > 0 and < Length"
		fmtMinPeriod      = "%s: MinPeriod should be >= 2"
		fmtMaxPeriod      = "%s: MaxPeriod should be > MinPeriod"
		fmtNyquist        = "%s: MaxPeriod should be <= 2 * Length"
		fmtResolution     = "%s: SpectrumResolution should be >= 1"
		fmtAgc            = "%s: AutomaticGainControlDecayFactor should be in (0, 1)"
		fmtw              = "%s: %w"
		descrPrefix       = "Maximum entropy spectrum "
		defLength         = 60
		defDegree         = 30
		defMinPeriod      = 2.0
		defMaxPeriod      = 59.0
		defSpectrumResult = 1
		defAgcDecayFactor = 0.995
		agcDecayEpsilon   = 1e-12
	)

	cfg := *p

	if cfg.Length == 0 {
		cfg.Length = defLength
	}

	if cfg.Degree == 0 {
		cfg.Degree = defDegree
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

	agcOn := !cfg.DisableAutomaticGainControl
	floatingNorm := !cfg.FixedNormalization

	if cfg.Length < 2 {
		return nil, fmt.Errorf(fmtLength, invalid)
	}

	if cfg.Degree <= 0 || cfg.Degree >= cfg.Length {
		return nil, fmt.Errorf(fmtDegree, invalid)
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

	// MaximumEntropySpectrum mirrors MBST's reference: BarMedianPrice default.
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

	flags := buildFlagTags(&cfg, agcOn, floatingNorm, defAgcDecayFactor, agcDecayEpsilon)
	mnemonic := fmt.Sprintf("mespect(%d, %d, %g, %g, %d%s%s)",
		cfg.Length, cfg.Degree, cfg.MinPeriod, cfg.MaxPeriod, cfg.SpectrumResolution,
		flags, componentMnemonic)

	est := newEstimator(
		cfg.Length, cfg.Degree, cfg.MinPeriod, cfg.MaxPeriod, cfg.SpectrumResolution,
		agcOn, cfg.AutomaticGainControlDecayFactor,
	)

	return &MaximumEntropySpectrum{
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
func buildFlagTags(cfg *Params, agcOn, floatingNorm bool, defAgc, eps float64) string {
	var s string

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
func (s *MaximumEntropySpectrum) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *MaximumEntropySpectrum) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.MaximumEntropySpectrum,
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
func (s *MaximumEntropySpectrum) Update(sample float64, t time.Time) *outputs.Heatmap {
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
func (s *MaximumEntropySpectrum) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.updateEntity(sample.Time, sample.Value)
}

// UpdateBar updates the indicator given the next bar sample.
func (s *MaximumEntropySpectrum) UpdateBar(sample *entities.Bar) core.Output {
	return s.updateEntity(sample.Time, s.barFunc(sample))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *MaximumEntropySpectrum) UpdateQuote(sample *entities.Quote) core.Output {
	return s.updateEntity(sample.Time, s.quoteFunc(sample))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *MaximumEntropySpectrum) UpdateTrade(sample *entities.Trade) core.Output {
	return s.updateEntity(sample.Time, s.tradeFunc(sample))
}

func (s *MaximumEntropySpectrum) updateEntity(t time.Time, sample float64) core.Output {
	const length = 1

	heatmap := s.Update(sample, t)

	output := make([]any, length)
	output[0] = heatmap

	return output
}
