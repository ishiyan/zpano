// Package coronaspectrum implements Ehlers' Corona Spectrum heatmap indicator.
//
// The Corona Spectrum measures cyclic activity over a cycle period range
// (default 6..30 bars) in a bank of contiguous bandpass filters. The amplitude
// of each filter output is compared to the strongest signal and displayed, in
// decibels, as a heatmap column. The filter having the strongest output is
// selected as the current dominant cycle period.
//
// Reference: John Ehlers, "Measuring Cycle Periods", Stocks & Commodities,
// November 2008.
package coronaspectrum

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
	"zpano/indicators/johnehlers/corona"
)

// CoronaSpectrum is Ehlers' Corona Spectrum heatmap indicator.
//
// It owns a private Corona spectral-analysis engine and exposes three outputs:
//
//   - Value: a per-bar heatmap column (decibels across the filter bank).
//   - DominantCycle: the weighted-center-of-gravity dominant cycle estimate.
//   - DominantCycleMedian: the 5-sample median of DominantCycle.
type CoronaSpectrum struct {
	mu                  sync.RWMutex
	mnemonic            string
	description         string
	mnemonicDC          string
	descriptionDC       string
	mnemonicDCM         string
	descriptionDCM      string
	c                   *corona.Corona
	minParameterValue   float64
	maxParameterValue   float64
	parameterResolution float64
	minRasterValue      float64
	maxRasterValue      float64
	barFunc             entities.BarFunc
	quoteFunc           entities.QuoteFunc
	tradeFunc           entities.TradeFunc
}

// NewCoronaSpectrumDefault returns an instance created with default parameters.
func NewCoronaSpectrumDefault() (*CoronaSpectrum, error) {
	return NewCoronaSpectrumParams(&Params{})
}

// NewCoronaSpectrumParams returns an instance created with the supplied parameters.
//
//nolint:funlen,cyclop
func NewCoronaSpectrumParams(p *Params) (*CoronaSpectrum, error) {
	const (
		invalid      = "invalid corona spectrum parameters"
		fmtMinRaster = "%s: MinRasterValue should be >= 0"
		fmtMaxRaster = "%s: MaxRasterValue should be > MinRasterValue"
		fmtMinParam  = "%s: MinParameterValue should be >= 2"
		fmtMaxParam  = "%s: MaxParameterValue should be > MinParameterValue"
		fmtHP        = "%s: HighPassFilterCutoff should be >= 2"
		fmtw         = "%s: %w"
		fmtnValue    = "cspect(%g, %g, %g, %g, %d%s)"
		fmtnDC       = "cspect-dc(%d%s)"
		fmtnDCM      = "cspect-dcm(%d%s)"
		descrValue   = "Corona spectrum "
		descrDC      = "Corona spectrum dominant cycle "
		descrDCM     = "Corona spectrum dominant cycle median "
		defMinRaster = 6.0
		defMaxRaster = 20.0
		defMinParam  = 6.0
		defMaxParam  = 30.0
		defHPCutoff  = 30
	)

	cfg := *p

	if cfg.MinRasterValue == 0 {
		cfg.MinRasterValue = defMinRaster
	}

	if cfg.MaxRasterValue == 0 {
		cfg.MaxRasterValue = defMaxRaster
	}

	if cfg.MinParameterValue == 0 {
		cfg.MinParameterValue = defMinParam
	}

	if cfg.MaxParameterValue == 0 {
		cfg.MaxParameterValue = defMaxParam
	}

	if cfg.HighPassFilterCutoff == 0 {
		cfg.HighPassFilterCutoff = defHPCutoff
	}

	if cfg.MinRasterValue < 0 {
		return nil, fmt.Errorf(fmtMinRaster, invalid)
	}

	if cfg.MaxRasterValue <= cfg.MinRasterValue {
		return nil, fmt.Errorf(fmtMaxRaster, invalid)
	}

	// MBST rounds min up and max down to integers.
	minParam := math.Ceil(cfg.MinParameterValue)
	maxParam := math.Floor(cfg.MaxParameterValue)

	if minParam < 2 {
		return nil, fmt.Errorf(fmtMinParam, invalid)
	}

	if maxParam <= minParam {
		return nil, fmt.Errorf(fmtMaxParam, invalid)
	}

	if cfg.HighPassFilterCutoff < 2 {
		return nil, fmt.Errorf(fmtHP, invalid)
	}

	// CoronaSpectrum mirrors Ehlers' reference: BarMedianPrice default.
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

	c, err := corona.NewCorona(&corona.Params{
		HighPassFilterCutoff:   cfg.HighPassFilterCutoff,
		MinimalPeriod:          int(minParam),
		MaximalPeriod:          int(maxParam),
		DecibelsLowerThreshold: cfg.MinRasterValue,
		DecibelsUpperThreshold: cfg.MaxRasterValue,
	})
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
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

	mnemonicValue := fmt.Sprintf(fmtnValue,
		cfg.MinRasterValue, cfg.MaxRasterValue, minParam, maxParam, cfg.HighPassFilterCutoff, componentMnemonic)
	mnemonicDC := fmt.Sprintf(fmtnDC, cfg.HighPassFilterCutoff, componentMnemonic)
	mnemonicDCM := fmt.Sprintf(fmtnDCM, cfg.HighPassFilterCutoff, componentMnemonic)

	// Parameter resolution: 2 samples per period (half-period stepping).
	// Values slice length = (maxParam*2 - minParam*2 + 1). The first sample sits
	// at minParam, the last at maxParam. Resolution satisfies:
	//   min + (length-1)/resolution = max  =>  resolution = (length-1) / (max - min).
	parameterResolution := float64(c.FilterBankLength()-1) / (maxParam - minParam)

	return &CoronaSpectrum{
		mnemonic:            mnemonicValue,
		description:         descrValue + mnemonicValue,
		mnemonicDC:          mnemonicDC,
		descriptionDC:       descrDC + mnemonicDC,
		mnemonicDCM:         mnemonicDCM,
		descriptionDCM:      descrDCM + mnemonicDCM,
		c:                   c,
		minParameterValue:   minParam,
		maxParameterValue:   maxParam,
		parameterResolution: parameterResolution,
		minRasterValue:      cfg.MinRasterValue,
		maxRasterValue:      cfg.MaxRasterValue,
		barFunc:             barFunc,
		quoteFunc:           quoteFunc,
		tradeFunc:           tradeFunc,
	}, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *CoronaSpectrum) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.c.IsPrimed()
}

// Metadata describes the output data of the indicator.
func (s *CoronaSpectrum) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.CoronaSpectrum,
		s.mnemonic,
		s.description,
		[]core.OutputText{
			{Mnemonic: s.mnemonic, Description: s.description},
			{Mnemonic: s.mnemonicDC, Description: s.descriptionDC},
			{Mnemonic: s.mnemonicDCM, Description: s.descriptionDCM},
		},
	)
}

// Update feeds the next sample to the engine and returns the heatmap column
// plus the current DominantCycle and DominantCycleMedian estimates.
//
// On unprimed bars the heatmap is an empty heatmap (with the indicator's
// parameter axis) and both scalar values are NaN. On a NaN input sample,
// state is left unchanged and all outputs are NaN / empty heatmap.
func (s *CoronaSpectrum) Update(sample float64, t time.Time) (*outputs.Heatmap, float64, float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if math.IsNaN(sample) {
		return outputs.NewEmptyHeatmap(t, s.minParameterValue, s.maxParameterValue, s.parameterResolution),
			math.NaN(), math.NaN()
	}

	primed := s.c.Update(sample)
	if !primed {
		return outputs.NewEmptyHeatmap(t, s.minParameterValue, s.maxParameterValue, s.parameterResolution),
			math.NaN(), math.NaN()
	}

	bank := s.c.FilterBank()
	values := make([]float64, len(bank))
	valueMin := math.Inf(1)
	valueMax := math.Inf(-1)

	for i := range bank {
		v := bank[i].Decibels
		values[i] = v

		if v < valueMin {
			valueMin = v
		}

		if v > valueMax {
			valueMax = v
		}
	}

	heatmap := outputs.NewHeatmap(t, s.minParameterValue, s.maxParameterValue, s.parameterResolution,
		valueMin, valueMax, values)

	return heatmap, s.c.DominantCycle(), s.c.DominantCycleMedian()
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *CoronaSpectrum) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.updateEntity(sample.Time, sample.Value)
}

// UpdateBar updates the indicator given the next bar sample.
func (s *CoronaSpectrum) UpdateBar(sample *entities.Bar) core.Output {
	return s.updateEntity(sample.Time, s.barFunc(sample))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *CoronaSpectrum) UpdateQuote(sample *entities.Quote) core.Output {
	return s.updateEntity(sample.Time, s.quoteFunc(sample))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *CoronaSpectrum) UpdateTrade(sample *entities.Trade) core.Output {
	return s.updateEntity(sample.Time, s.tradeFunc(sample))
}

func (s *CoronaSpectrum) updateEntity(t time.Time, sample float64) core.Output {
	const length = 3

	heatmap, dc, dcm := s.Update(sample, t)

	output := make([]any, length)
	output[0] = heatmap
	output[1] = entities.Scalar{Time: t, Value: dc}
	output[2] = entities.Scalar{Time: t, Value: dcm}

	return output
}
