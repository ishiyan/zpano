// Package coronasignaltonoiseratio implements Ehlers' Corona Signal-to-Noise
// Ratio heatmap indicator.
//
// The Corona Signal to Noise Ratio is a measure of the cycle amplitude relative
// to noise. The "noise" is chosen to be the average bar height because there is
// not much trade information within the bar.
//
// Reference: John Ehlers, "Measuring Cycle Periods", Stocks & Commodities,
// November 2008.
package coronasignaltonoiseratio

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
	"zpano/indicators/johnehlers/corona"
)

const (
	highLowBufferSize       = 5
	highLowBufferSizeMinOne = highLowBufferSize - 1
	highLowMedianIndex      = 2
	averageSampleAlpha      = 0.1
	averageSampleOneMinus   = 0.9
	signalEmaAlpha          = 0.2
	signalEmaOneMinus       = 0.9 // Intentional: sums to 1.1, per Ehlers.
	noiseEmaAlpha           = 0.1
	noiseEmaOneMinus        = 0.9
	ratioOffsetDb           = 3.5
	ratioUpperDb            = 10.
	dbGain                  = 20.
	widthLowRatioThreshold  = 0.5
	widthBaseline           = 0.2
	widthSlope              = 0.4
	rasterBlendExponent     = 0.8
	rasterBlendHalf         = 0.5
	rasterNegativeArgCutoff = 1.
)

// CoronaSignalToNoiseRatio is Ehlers' Corona Signal-to-Noise Ratio indicator.
//
// It owns a private Corona spectral-analysis engine and exposes two outputs:
//
//   - Value: a per-bar heatmap column (intensity raster).
//   - SignalToNoiseRatio: the current SNR value mapped into
//     [MinParameterValue, MaxParameterValue].
type CoronaSignalToNoiseRatio struct {
	mu                    sync.RWMutex
	mnemonic              string
	description           string
	mnemonicSNR           string
	descriptionSNR        string
	c                     *corona.Corona
	rasterLength          int
	rasterStep            float64
	maxRasterValue        float64
	minParameterValue     float64
	maxParameterValue     float64
	parameterResolution   float64
	raster                []float64
	highLowBuffer         [highLowBufferSize]float64
	hlSorted              [highLowBufferSize]float64
	averageSamplePrevious float64
	signalPrevious        float64
	noisePrevious         float64
	signalToNoiseRatio    float64
	isStarted             bool
	barFunc               entities.BarFunc
	quoteFunc             entities.QuoteFunc
	tradeFunc             entities.TradeFunc
}

// NewCoronaSignalToNoiseRatioDefault returns an instance created with default parameters.
func NewCoronaSignalToNoiseRatioDefault() (*CoronaSignalToNoiseRatio, error) {
	return NewCoronaSignalToNoiseRatioParams(&Params{})
}

// NewCoronaSignalToNoiseRatioParams returns an instance created with the supplied parameters.
//
//nolint:funlen,cyclop
func NewCoronaSignalToNoiseRatioParams(p *Params) (*CoronaSignalToNoiseRatio, error) {
	const (
		invalid      = "invalid corona signal to noise ratio parameters"
		fmtRaster    = "%s: RasterLength should be >= 2"
		fmtMaxRaster = "%s: MaxRasterValue should be > 0"
		fmtMinParam  = "%s: MinParameterValue should be >= 0"
		fmtMaxParam  = "%s: MaxParameterValue should be > MinParameterValue"
		fmtHP        = "%s: HighPassFilterCutoff should be >= 2"
		fmtMinP      = "%s: MinimalPeriod should be >= 2"
		fmtMaxP      = "%s: MaximalPeriod should be > MinimalPeriod"
		fmtw         = "%s: %w"
		fmtnValue    = "csnr(%d, %g, %g, %g, %d%s)"
		fmtnSNR      = "csnr-snr(%d%s)"
		descrValue   = "Corona signal to noise ratio "
		descrSNR     = "Corona signal to noise ratio scalar "
		defRaster    = 50
		defMaxRast   = 20.0
		defMinParam  = 1.0
		defMaxParam  = 11.0
		defHPCutoff  = 30
		defMinPer    = 6
		defMaxPer    = 30
	)

	cfg := *p

	if cfg.RasterLength == 0 {
		cfg.RasterLength = defRaster
	}

	if cfg.MaxRasterValue == 0 {
		cfg.MaxRasterValue = defMaxRast
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

	if cfg.MinimalPeriod == 0 {
		cfg.MinimalPeriod = defMinPer
	}

	if cfg.MaximalPeriod == 0 {
		cfg.MaximalPeriod = defMaxPer
	}

	if cfg.RasterLength < 2 {
		return nil, fmt.Errorf(fmtRaster, invalid)
	}

	if cfg.MaxRasterValue <= 0 {
		return nil, fmt.Errorf(fmtMaxRaster, invalid)
	}

	if cfg.MinParameterValue < 0 {
		return nil, fmt.Errorf(fmtMinParam, invalid)
	}

	if cfg.MaxParameterValue <= cfg.MinParameterValue {
		return nil, fmt.Errorf(fmtMaxParam, invalid)
	}

	if cfg.HighPassFilterCutoff < 2 {
		return nil, fmt.Errorf(fmtHP, invalid)
	}

	if cfg.MinimalPeriod < 2 {
		return nil, fmt.Errorf(fmtMinP, invalid)
	}

	if cfg.MaximalPeriod <= cfg.MinimalPeriod {
		return nil, fmt.Errorf(fmtMaxP, invalid)
	}

	// Ehlers reference uses (High+Low)/2.
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
		HighPassFilterCutoff: cfg.HighPassFilterCutoff,
		MinimalPeriod:        cfg.MinimalPeriod,
		MaximalPeriod:        cfg.MaximalPeriod,
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
		cfg.RasterLength, cfg.MaxRasterValue, cfg.MinParameterValue, cfg.MaxParameterValue,
		cfg.HighPassFilterCutoff, componentMnemonic)
	mnemonicSNR := fmt.Sprintf(fmtnSNR, cfg.HighPassFilterCutoff, componentMnemonic)

	// Resolution satisfies: min + (rasterLength-1)/resolution = max.
	parameterResolution := float64(cfg.RasterLength-1) / (cfg.MaxParameterValue - cfg.MinParameterValue)

	return &CoronaSignalToNoiseRatio{
		mnemonic:            mnemonicValue,
		description:         descrValue + mnemonicValue,
		mnemonicSNR:         mnemonicSNR,
		descriptionSNR:      descrSNR + mnemonicSNR,
		c:                   c,
		rasterLength:        cfg.RasterLength,
		rasterStep:          cfg.MaxRasterValue / float64(cfg.RasterLength),
		maxRasterValue:      cfg.MaxRasterValue,
		minParameterValue:   cfg.MinParameterValue,
		maxParameterValue:   cfg.MaxParameterValue,
		parameterResolution: parameterResolution,
		raster:              make([]float64, cfg.RasterLength),
		signalToNoiseRatio:  math.NaN(),
		barFunc:             barFunc,
		quoteFunc:           quoteFunc,
		tradeFunc:           tradeFunc,
	}, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *CoronaSignalToNoiseRatio) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.c.IsPrimed()
}

// Metadata describes the output data of the indicator.
func (s *CoronaSignalToNoiseRatio) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.CoronaSignalToNoiseRatio,
		s.mnemonic,
		s.description,
		[]core.OutputText{
			{Mnemonic: s.mnemonic, Description: s.description},
			{Mnemonic: s.mnemonicSNR, Description: s.descriptionSNR},
		},
	)
}

// Update feeds the next sample plus bar extremes and returns the heatmap column
// and the current SignalToNoiseRatio. On unprimed bars the heatmap is empty and
// the scalar is NaN. On NaN sample input state is left unchanged.
//
//nolint:funlen,cyclop,gocognit
func (s *CoronaSignalToNoiseRatio) Update(sample, sampleLow, sampleHigh float64, t time.Time) (*outputs.Heatmap, float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if math.IsNaN(sample) {
		return outputs.NewEmptyHeatmap(t, s.minParameterValue, s.maxParameterValue, s.parameterResolution), math.NaN()
	}

	primed := s.c.Update(sample)

	if !s.isStarted {
		s.averageSamplePrevious = sample
		s.highLowBuffer[highLowBufferSizeMinOne] = sampleHigh - sampleLow
		s.isStarted = true

		return outputs.NewEmptyHeatmap(t, s.minParameterValue, s.maxParameterValue, s.parameterResolution), math.NaN()
	}

	maxAmpSq := s.c.MaximalAmplitudeSquared()

	averageSample := averageSampleAlpha*sample + averageSampleOneMinus*s.averageSamplePrevious
	s.averageSamplePrevious = averageSample

	if math.Abs(averageSample) > 0 || maxAmpSq > 0 {
		s.signalPrevious = signalEmaAlpha*math.Sqrt(maxAmpSq) + signalEmaOneMinus*s.signalPrevious
	}

	// Shift H-L ring buffer left; push new value.
	for i := 0; i < highLowBufferSizeMinOne; i++ {
		s.highLowBuffer[i] = s.highLowBuffer[i+1]
	}
	s.highLowBuffer[highLowBufferSizeMinOne] = sampleHigh - sampleLow

	ratio := 0.0
	if math.Abs(averageSample) > 0 {
		for i := 0; i < highLowBufferSize; i++ {
			s.hlSorted[i] = s.highLowBuffer[i]
		}

		sort.Float64s(s.hlSorted[:])
		s.noisePrevious = noiseEmaAlpha*s.hlSorted[highLowMedianIndex] + noiseEmaOneMinus*s.noisePrevious

		if math.Abs(s.noisePrevious) > 0 {
			ratio = dbGain*math.Log10(s.signalPrevious/s.noisePrevious) + ratioOffsetDb
			if ratio < 0 {
				ratio = 0
			} else if ratio > ratioUpperDb {
				ratio = ratioUpperDb
			}

			ratio /= ratioUpperDb // ∈ [0, 1]
		}
	}

	s.signalToNoiseRatio = (s.maxParameterValue-s.minParameterValue)*ratio + s.minParameterValue

	// Raster update.
	width := 0.0
	if ratio <= widthLowRatioThreshold {
		width = widthBaseline - widthSlope*ratio
	}

	ratioScaledToRasterLength := int(math.Round(ratio * float64(s.rasterLength)))
	ratioScaledToMaxRasterValue := ratio * s.maxRasterValue

	for i := 0; i < s.rasterLength; i++ {
		value := s.raster[i]

		switch {
		case i == ratioScaledToRasterLength:
			value *= 0.5
		case width == 0:
			// Above the high-ratio threshold: handled by the ratio>0.5 override below.
		default:
			argument := (ratioScaledToMaxRasterValue - s.rasterStep*float64(i)) / width
			if i < ratioScaledToRasterLength {
				value = rasterBlendHalf * (math.Pow(argument, rasterBlendExponent) + value)
			} else {
				argument = -argument
				if argument > rasterNegativeArgCutoff {
					value = rasterBlendHalf * (math.Pow(argument, rasterBlendExponent) + value)
				} else {
					value = s.maxRasterValue
				}
			}
		}

		if value < 0 {
			value = 0
		} else if value > s.maxRasterValue {
			value = s.maxRasterValue
		}

		if ratio > widthLowRatioThreshold {
			value = s.maxRasterValue
		}

		s.raster[i] = value
	}

	if !primed {
		return outputs.NewEmptyHeatmap(t, s.minParameterValue, s.maxParameterValue, s.parameterResolution), math.NaN()
	}

	values := make([]float64, s.rasterLength)
	valueMin := math.Inf(1)
	valueMax := math.Inf(-1)

	for i := 0; i < s.rasterLength; i++ {
		v := s.raster[i]
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

	return heatmap, s.signalToNoiseRatio
}

// UpdateScalar updates the indicator given the next scalar sample. Since no
// High/Low is available, the sample is used for both, yielding zero noise.
func (s *CoronaSignalToNoiseRatio) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.updateEntity(sample.Time, sample.Value, sample.Value, sample.Value)
}

// UpdateBar updates the indicator given the next bar sample.
func (s *CoronaSignalToNoiseRatio) UpdateBar(sample *entities.Bar) core.Output {
	return s.updateEntity(sample.Time, s.barFunc(sample), sample.Low, sample.High)
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *CoronaSignalToNoiseRatio) UpdateQuote(sample *entities.Quote) core.Output {
	v := s.quoteFunc(sample)

	return s.updateEntity(sample.Time, v, v, v)
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *CoronaSignalToNoiseRatio) UpdateTrade(sample *entities.Trade) core.Output {
	v := s.tradeFunc(sample)

	return s.updateEntity(sample.Time, v, v, v)
}

func (s *CoronaSignalToNoiseRatio) updateEntity(t time.Time, sample, low, high float64) core.Output {
	const length = 2

	heatmap, snr := s.Update(sample, low, high, t)

	output := make([]any, length)
	output[0] = heatmap
	output[1] = entities.Scalar{Time: t, Value: snr}

	return output
}
