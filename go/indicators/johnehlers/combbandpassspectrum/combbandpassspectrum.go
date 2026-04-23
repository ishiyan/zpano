// Package combbandpassspectrum implements Ehlers' Comb Band-Pass Spectrum
// heatmap indicator.
//
// The Comb Band-Pass Spectrum (cbps) displays a power heatmap of cyclic
// activity over a configurable cycle-period range. Each cycle bin is
// estimated by a dedicated 2-pole band-pass filter tuned to that period,
// forming a "comb" filter bank. The close series is pre-conditioned by a
// 2-pole Butterworth highpass (cutoff = MaxPeriod) followed by a 2-pole
// Super Smoother (cutoff = MinPeriod) before it enters the comb. Each bin's
// power is the sum of squared band-pass outputs over the last N samples,
// optionally compensated for spectral dilation (divide by N) and normalized
// by a fast-attack slow-decay automatic gain control.
//
// This implementation follows John Ehlers' EasyLanguage listing 10-1 from
// "Cycle Analytics for Traders". It is NOT a port of MBST's
// CombBandPassSpectrumEstimator, which is misnamed and actually implements
// a plain DFT (see the DiscreteFourierTransformSpectrum indicator for a
// faithful MBST DFT port).
//
// Reference: John F. Ehlers, "Cycle Analytics for Traders",
// Code Listing 10-1 (Comb BandPass Spectrum).
package combbandpassspectrum

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

// CombBandPassSpectrum is Ehlers' Comb Band-Pass Spectrum heatmap indicator.
type CombBandPassSpectrum struct {
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

// NewCombBandPassSpectrumDefault returns an instance created with default parameters.
func NewCombBandPassSpectrumDefault() (*CombBandPassSpectrum, error) {
	return NewCombBandPassSpectrumParams(&Params{})
}

// NewCombBandPassSpectrumParams returns an instance created with the supplied parameters.
//
//nolint:funlen,cyclop
func NewCombBandPassSpectrumParams(p *Params) (*CombBandPassSpectrum, error) {
	const (
		invalid           = "invalid comb band-pass spectrum parameters"
		fmtMinPeriod      = "%s: MinPeriod should be >= 2"
		fmtMaxPeriod      = "%s: MaxPeriod should be > MinPeriod"
		fmtBandwidth      = "%s: Bandwidth should be in (0, 1)"
		fmtAgc            = "%s: AutomaticGainControlDecayFactor should be in (0, 1)"
		fmtw              = "%s: %w"
		descrPrefix       = "Comb band-pass spectrum "
		defMinPeriod      = 10
		defMaxPeriod      = 48
		defBandwidth      = 0.3
		defAgcDecayFactor = 0.995
		agcDecayEpsilon   = 1e-12
		bandwidthEpsilon  = 1e-12
	)

	cfg := *p

	if cfg.MinPeriod == 0 {
		cfg.MinPeriod = defMinPeriod
	}

	if cfg.MaxPeriod == 0 {
		cfg.MaxPeriod = defMaxPeriod
	}

	if cfg.Bandwidth == 0 {
		cfg.Bandwidth = defBandwidth
	}

	if cfg.AutomaticGainControlDecayFactor == 0 {
		cfg.AutomaticGainControlDecayFactor = defAgcDecayFactor
	}

	sdcOn := !cfg.DisableSpectralDilationCompensation
	agcOn := !cfg.DisableAutomaticGainControl
	floatingNorm := !cfg.FixedNormalization

	if cfg.MinPeriod < 2 {
		return nil, fmt.Errorf(fmtMinPeriod, invalid)
	}

	if cfg.MaxPeriod <= cfg.MinPeriod {
		return nil, fmt.Errorf(fmtMaxPeriod, invalid)
	}

	if cfg.Bandwidth <= 0 || cfg.Bandwidth >= 1 {
		return nil, fmt.Errorf(fmtBandwidth, invalid)
	}

	if agcOn &&
		(cfg.AutomaticGainControlDecayFactor <= 0 || cfg.AutomaticGainControlDecayFactor >= 1) {
		return nil, fmt.Errorf(fmtAgc, invalid)
	}

	// CombBandPassSpectrum mirrors Ehlers' reference: BarMedianPrice default.
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
		&cfg, sdcOn, agcOn, floatingNorm, defBandwidth, defAgcDecayFactor,
		bandwidthEpsilon, agcDecayEpsilon,
	)
	mnemonic := fmt.Sprintf("cbps(%d, %d%s%s)",
		cfg.MinPeriod, cfg.MaxPeriod, flags, componentMnemonic)

	est := newEstimator(
		cfg.MinPeriod, cfg.MaxPeriod, cfg.Bandwidth,
		sdcOn, agcOn, cfg.AutomaticGainControlDecayFactor,
	)

	return &CombBandPassSpectrum{
		mnemonic:              mnemonic,
		description:           descrPrefix + mnemonic,
		estimator:             est,
		primeCount:            cfg.MaxPeriod,
		floatingNormalization: floatingNorm,
		minParameterValue:     float64(cfg.MinPeriod),
		maxParameterValue:     float64(cfg.MaxPeriod),
		parameterResolution:   1,
		barFunc:               barFunc,
		quoteFunc:             quoteFunc,
		tradeFunc:             tradeFunc,
	}, nil
}

// buildFlagTags encodes non-default boolean/decay/bandwidth settings as terse
// override-only tags. Returns an empty string when all flags are at their
// defaults. Emission order matches the Params field order.
//
//nolint:revive
func buildFlagTags(
	cfg *Params,
	sdcOn, agcOn, floatingNorm bool,
	defBandwidth, defAgc, bwEps, agcEps float64,
) string {
	var s string

	if math.Abs(cfg.Bandwidth-defBandwidth) > bwEps {
		s += fmt.Sprintf(", bw=%g", cfg.Bandwidth)
	}

	if !sdcOn {
		s += ", no-sdc"
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
func (s *CombBandPassSpectrum) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *CombBandPassSpectrum) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.CombBandPassSpectrum,
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
func (s *CombBandPassSpectrum) Update(sample float64, t time.Time) *outputs.Heatmap {
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

	maxRef := s.estimator.spectrumMax
	spectrumRange := maxRef - minRef

	values := make([]float64, lengthSpectrum)
	valueMin := math.Inf(1)
	valueMax := math.Inf(-1)

	// The estimator's spectrum is already in axis order (bin 0 = MinPeriod,
	// bin last = MaxPeriod), matching the heatmap axis.
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
func (s *CombBandPassSpectrum) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.updateEntity(sample.Time, sample.Value)
}

// UpdateBar updates the indicator given the next bar sample.
func (s *CombBandPassSpectrum) UpdateBar(sample *entities.Bar) core.Output {
	return s.updateEntity(sample.Time, s.barFunc(sample))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *CombBandPassSpectrum) UpdateQuote(sample *entities.Quote) core.Output {
	return s.updateEntity(sample.Time, s.quoteFunc(sample))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *CombBandPassSpectrum) UpdateTrade(sample *entities.Trade) core.Output {
	return s.updateEntity(sample.Time, s.tradeFunc(sample))
}

func (s *CombBandPassSpectrum) updateEntity(t time.Time, sample float64) core.Output {
	const length = 1

	heatmap := s.Update(sample, t)

	output := make([]any, length)
	output[0] = heatmap

	return output
}
