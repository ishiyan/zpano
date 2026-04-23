// Package autocorrelationperiodogram implements Ehlers' Autocorrelation
// Periodogram heatmap indicator.
//
// The Autocorrelation Periodogram (acp) displays a power heatmap of cyclic
// activity by taking a discrete Fourier transform of the autocorrelation
// function. The close series is pre-conditioned by a 2-pole Butterworth
// highpass (cutoff = MaxPeriod) followed by a 2-pole Super Smoother
// (cutoff = MinPeriod). The autocorrelation function is evaluated at lags
// 0..MaxPeriod using Pearson correlation with a fixed averaging length.
// Each period bin's squared-sum Fourier magnitude is exponentially smoothed,
// fast-attack / slow-decay AGC normalized, and displayed.
//
// This implementation follows John Ehlers' EasyLanguage listing 8-3 from
// "Cycle Analytics for Traders". It is NOT a port of MBST's
// AutoCorrelationSpectrum / AutoCorrelationSpectrumEstimator, which omits
// the HP + SS pre-filter, uses a different Pearson formulation, and smooths
// raw SqSum rather than SqSum² (see the package README / conversion skill
// exemplar for details).
//
// Reference: John F. Ehlers, "Cycle Analytics for Traders",
// Code Listing 8-3 (Autocorrelation Periodogram).
package autocorrelationperiodogram

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

// AutoCorrelationPeriodogram is Ehlers' Autocorrelation Periodogram heatmap.
type AutoCorrelationPeriodogram struct {
	mu                    sync.RWMutex
	mnemonic              string
	description           string
	estimator             *estimator
	windowCount           int
	primeCount            int
	primed                bool
	floatingNormalization bool
	minParameterValue     float64
	maxParameterValue     float64
	parameterResolution   float64
	barFunc               entities.BarFunc
	quoteFunc             entities.QuoteFunc
	tradeFunc             entities.TradeFunc
}

// NewAutoCorrelationPeriodogramDefault returns an instance created with default parameters.
func NewAutoCorrelationPeriodogramDefault() (*AutoCorrelationPeriodogram, error) {
	return NewAutoCorrelationPeriodogramParams(&Params{})
}

// NewAutoCorrelationPeriodogramParams returns an instance created with the supplied parameters.
//
//nolint:funlen,cyclop
func NewAutoCorrelationPeriodogramParams(p *Params) (*AutoCorrelationPeriodogram, error) {
	const (
		invalid           = "invalid autocorrelation periodogram parameters"
		fmtMinPeriod      = "%s: MinPeriod should be >= 2"
		fmtMaxPeriod      = "%s: MaxPeriod should be > MinPeriod"
		fmtAverage        = "%s: AveragingLength should be >= 1"
		fmtAgc            = "%s: AutomaticGainControlDecayFactor should be in (0, 1)"
		fmtw              = "%s: %w"
		descrPrefix       = "Autocorrelation periodogram "
		defMinPeriod      = 10
		defMaxPeriod      = 48
		defAveragingLen   = 3
		defAgcDecayFactor = 0.995
		agcDecayEpsilon   = 1e-12
	)

	cfg := *p

	if cfg.MinPeriod == 0 {
		cfg.MinPeriod = defMinPeriod
	}

	if cfg.MaxPeriod == 0 {
		cfg.MaxPeriod = defMaxPeriod
	}

	if cfg.AveragingLength == 0 {
		cfg.AveragingLength = defAveragingLen
	}

	if cfg.AutomaticGainControlDecayFactor == 0 {
		cfg.AutomaticGainControlDecayFactor = defAgcDecayFactor
	}

	squaringOn := !cfg.DisableSpectralSquaring
	smoothingOn := !cfg.DisableSmoothing
	agcOn := !cfg.DisableAutomaticGainControl
	floatingNorm := !cfg.FixedNormalization

	if cfg.MinPeriod < 2 {
		return nil, fmt.Errorf(fmtMinPeriod, invalid)
	}

	if cfg.MaxPeriod <= cfg.MinPeriod {
		return nil, fmt.Errorf(fmtMaxPeriod, invalid)
	}

	if cfg.AveragingLength < 1 {
		return nil, fmt.Errorf(fmtAverage, invalid)
	}

	if agcOn &&
		(cfg.AutomaticGainControlDecayFactor <= 0 || cfg.AutomaticGainControlDecayFactor >= 1) {
		return nil, fmt.Errorf(fmtAgc, invalid)
	}

	// AutoCorrelationPeriodogram mirrors Ehlers' reference: BarMedianPrice default.
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

	flags := buildFlagTags(
		&cfg, squaringOn, smoothingOn, agcOn, floatingNorm,
		defAveragingLen, defAgcDecayFactor, agcDecayEpsilon,
	)
	mnemonic := fmt.Sprintf("acp(%d, %d%s%s)",
		cfg.MinPeriod, cfg.MaxPeriod, flags, componentMnemonic)

	est := newEstimator(
		cfg.MinPeriod, cfg.MaxPeriod, cfg.AveragingLength,
		squaringOn, smoothingOn, agcOn, cfg.AutomaticGainControlDecayFactor,
	)

	return &AutoCorrelationPeriodogram{
		mnemonic:              mnemonic,
		description:           descrPrefix + mnemonic,
		estimator:             est,
		primeCount:            est.filtBufferLen,
		floatingNormalization: floatingNorm,
		minParameterValue:     float64(cfg.MinPeriod),
		maxParameterValue:     float64(cfg.MaxPeriod),
		parameterResolution:   1,
		barFunc:               barFunc,
		quoteFunc:             quoteFunc,
		tradeFunc:             tradeFunc,
	}, nil
}

// buildFlagTags encodes non-default settings as terse override-only tags.
// Returns an empty string when all flags are at their defaults. Emission
// order matches the Params field order.
//
//nolint:revive
func buildFlagTags(
	cfg *Params,
	squaringOn, smoothingOn, agcOn, floatingNorm bool,
	defAverage int,
	defAgc, agcEps float64,
) string {
	var s string

	if cfg.AveragingLength != defAverage {
		s += fmt.Sprintf(", average=%d", cfg.AveragingLength)
	}

	if !squaringOn {
		s += ", no-sqr"
	}

	if !smoothingOn {
		s += ", no-smooth"
	}

	if !agcOn {
		s += ", no-agc"
	}

	if agcOn && math.Abs(cfg.AutomaticGainControlDecayFactor-defAgc) > agcEps {
		s += fmt.Sprintf(", agc=%g", cfg.AutomaticGainControlDecayFactor)
	}

	if !floatingNorm {
		s += ", no-fn"
	}

	return s
}

// IsPrimed indicates whether the indicator is primed.
func (s *AutoCorrelationPeriodogram) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *AutoCorrelationPeriodogram) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.AutoCorrelationPeriodogram,
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
func (s *AutoCorrelationPeriodogram) Update(sample float64, t time.Time) *outputs.Heatmap {
	s.mu.Lock()
	defer s.mu.Unlock()

	if math.IsNaN(sample) {
		return outputs.NewEmptyHeatmap(t, s.minParameterValue, s.maxParameterValue, s.parameterResolution)
	}

	s.estimator.update(sample)

	if !s.primed {
		s.windowCount++

		if s.windowCount >= s.primeCount {
			s.primed = true
		} else {
			return outputs.NewEmptyHeatmap(t, s.minParameterValue, s.maxParameterValue, s.parameterResolution)
		}
	}

	lengthSpectrum := s.estimator.lengthSpectrum

	var minRef float64
	if s.floatingNormalization {
		minRef = s.estimator.spectrumMin
	}

	// Estimator spectrum is already AGC-normalized in [0, 1]. Apply optional
	// floating-minimum subtraction for display.
	maxRef := 1.0
	spectrumRange := maxRef - minRef

	values := make([]float64, lengthSpectrum)
	valueMin := math.Inf(1)
	valueMax := math.Inf(-1)

	for i := 0; i < lengthSpectrum; i++ {
		var v float64
		if spectrumRange > 0 {
			v = (s.estimator.spectrum[i] - minRef) / spectrumRange
		}

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
func (s *AutoCorrelationPeriodogram) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.updateEntity(sample.Time, sample.Value)
}

// UpdateBar updates the indicator given the next bar sample.
func (s *AutoCorrelationPeriodogram) UpdateBar(sample *entities.Bar) core.Output {
	return s.updateEntity(sample.Time, s.barFunc(sample))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *AutoCorrelationPeriodogram) UpdateQuote(sample *entities.Quote) core.Output {
	return s.updateEntity(sample.Time, s.quoteFunc(sample))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *AutoCorrelationPeriodogram) UpdateTrade(sample *entities.Trade) core.Output {
	return s.updateEntity(sample.Time, s.tradeFunc(sample))
}

func (s *AutoCorrelationPeriodogram) updateEntity(t time.Time, sample float64) core.Output {
	const length = 1

	heatmap := s.Update(sample, t)

	output := make([]any, length)
	output[0] = heatmap

	return output
}
