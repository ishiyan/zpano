// Package autocorrelationindicator implements Ehlers' Autocorrelation
// Indicator heatmap.
//
// The Autocorrelation Indicator (aci) displays a heatmap of Pearson
// correlation coefficients between the current filtered series and a lagged
// copy of itself, across a configurable lag range. The close series is
// pre-conditioned by a 2-pole Butterworth highpass (cutoff = MaxLag) followed
// by a 2-pole Super Smoother (cutoff = SmoothingPeriod) before the
// correlation bank is evaluated. Each bin's value is rescaled from the
// Pearson [-1, 1] range into [0, 1] via 0.5*(r + 1) for direct display.
//
// This implementation follows John Ehlers' EasyLanguage listing 8-2 from
// "Cycle Analytics for Traders". It is NOT a port of MBST's
// AutoCorrelationCoefficients / AutoCorrelationEstimator, which omit the
// HP + SS pre-filter, use a different Pearson formulation, and have an
// opposite AverageLength=0 convention (see the package README / conversion
// skill exemplar for details).
//
// Reference: John F. Ehlers, "Cycle Analytics for Traders",
// Code Listing 8-2 (Autocorrelation Indicator).
package autocorrelationindicator

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

// AutoCorrelationIndicator is Ehlers' Autocorrelation Indicator heatmap.
type AutoCorrelationIndicator struct {
	mu                  sync.RWMutex
	mnemonic            string
	description         string
	estimator           *estimator
	windowCount         int
	primeCount          int
	primed              bool
	minParameterValue   float64
	maxParameterValue   float64
	parameterResolution float64
	barFunc             entities.BarFunc
	quoteFunc           entities.QuoteFunc
	tradeFunc           entities.TradeFunc
}

// NewAutoCorrelationIndicatorDefault returns an instance created with default parameters.
func NewAutoCorrelationIndicatorDefault() (*AutoCorrelationIndicator, error) {
	return NewAutoCorrelationIndicatorParams(&Params{})
}

// NewAutoCorrelationIndicatorParams returns an instance created with the supplied parameters.
//
//nolint:funlen,cyclop
func NewAutoCorrelationIndicatorParams(p *Params) (*AutoCorrelationIndicator, error) {
	const (
		invalid         = "invalid autocorrelation indicator parameters"
		fmtMinLag       = "%s: MinLag should be >= 1"
		fmtMaxLag       = "%s: MaxLag should be > MinLag"
		fmtSmoothing    = "%s: SmoothingPeriod should be >= 2"
		fmtAverage      = "%s: AveragingLength should be >= 0"
		fmtw            = "%s: %w"
		descrPrefix     = "Autocorrelation indicator "
		defMinLag       = 3
		defMaxLag       = 48
		defSmoothing    = 10
		defAveragingLen = 0
	)

	cfg := *p

	if cfg.MinLag == 0 {
		cfg.MinLag = defMinLag
	}

	if cfg.MaxLag == 0 {
		cfg.MaxLag = defMaxLag
	}

	if cfg.SmoothingPeriod == 0 {
		cfg.SmoothingPeriod = defSmoothing
	}

	if cfg.MinLag < 1 {
		return nil, fmt.Errorf(fmtMinLag, invalid)
	}

	if cfg.MaxLag <= cfg.MinLag {
		return nil, fmt.Errorf(fmtMaxLag, invalid)
	}

	if cfg.SmoothingPeriod < 2 {
		return nil, fmt.Errorf(fmtSmoothing, invalid)
	}

	if cfg.AveragingLength < 0 {
		return nil, fmt.Errorf(fmtAverage, invalid)
	}

	// AutoCorrelationIndicator mirrors Ehlers' reference: BarMedianPrice default.
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

	flags := buildFlagTags(&cfg, defAveragingLen)
	mnemonic := fmt.Sprintf("aci(%d, %d, %d%s%s)",
		cfg.MinLag, cfg.MaxLag, cfg.SmoothingPeriod, flags, componentMnemonic)

	est := newEstimator(cfg.MinLag, cfg.MaxLag, cfg.SmoothingPeriod, cfg.AveragingLength)

	return &AutoCorrelationIndicator{
		mnemonic:            mnemonic,
		description:         descrPrefix + mnemonic,
		estimator:           est,
		primeCount:          est.filtBufferLen,
		minParameterValue:   float64(cfg.MinLag),
		maxParameterValue:   float64(cfg.MaxLag),
		parameterResolution: 1,
		barFunc:             barFunc,
		quoteFunc:           quoteFunc,
		tradeFunc:           tradeFunc,
	}, nil
}

// buildFlagTags encodes non-default settings as terse override-only tags.
// Returns an empty string when all flags are at their defaults. Emission
// order matches the Params field order.
func buildFlagTags(cfg *Params, defAveragingLen int) string {
	var s string

	if cfg.AveragingLength != defAveragingLen {
		s += fmt.Sprintf(", average=%d", cfg.AveragingLength)
	}

	return s
}

// IsPrimed indicates whether the indicator is primed.
func (s *AutoCorrelationIndicator) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *AutoCorrelationIndicator) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.AutoCorrelationIndicator,
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
func (s *AutoCorrelationIndicator) Update(sample float64, t time.Time) *outputs.Heatmap {
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

	values := make([]float64, lengthSpectrum)
	valueMin := math.Inf(1)
	valueMax := math.Inf(-1)

	// The estimator's spectrum is already in axis order (bin 0 = MinLag,
	// bin last = MaxLag), and values are already scaled to [0, 1] via
	// 0.5*(r + 1). No additional normalization is applied.
	for i := 0; i < lengthSpectrum; i++ {
		v := s.estimator.spectrum[i]
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
func (s *AutoCorrelationIndicator) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.updateEntity(sample.Time, sample.Value)
}

// UpdateBar updates the indicator given the next bar sample.
func (s *AutoCorrelationIndicator) UpdateBar(sample *entities.Bar) core.Output {
	return s.updateEntity(sample.Time, s.barFunc(sample))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *AutoCorrelationIndicator) UpdateQuote(sample *entities.Quote) core.Output {
	return s.updateEntity(sample.Time, s.quoteFunc(sample))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *AutoCorrelationIndicator) UpdateTrade(sample *entities.Trade) core.Output {
	return s.updateEntity(sample.Time, s.tradeFunc(sample))
}

func (s *AutoCorrelationIndicator) updateEntity(t time.Time, sample float64) core.Output {
	const length = 1

	heatmap := s.Update(sample, t)

	output := make([]any, length)
	output[0] = heatmap

	return output
}
