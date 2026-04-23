// Package coronatrendvigor implements Ehlers' Corona Trend Vigor heatmap
// indicator.
//
// The Corona trend vigor is computed as the slope of the momentum taken over a
// full dominant cycle period, normalised by the cycle amplitude. The ratio is
// scaled into the range [-10, 10]. A value of ±2 means the trend slope equals
// twice the cycle amplitude; values between -2 and +2 form the "corona" and
// suggest the trend should not be traded.
//
// Reference: John Ehlers, "Measuring Cycle Periods", Stocks & Commodities,
// November 2008.
package coronatrendvigor

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

const (
	// Bandpass filter delta factor in γ = 1/cos(ω · 2 · 0.1).
	bpDelta = 0.1

	// Ratio EMA coefficients (0.33 * current + 0.67 * previous).
	ratioNewCoef      = 0.33
	ratioPreviousCoef = 0.67

	// Vigor band thresholds.
	vigorMidLow  = 0.3
	vigorMidHigh = 0.7
	vigorMid     = 0.5
	widthEdge    = 0.01

	// Raster update blend factors.
	rasterBlendScale    = 0.8
	rasterBlendPrevious = 0.2
	rasterBlendHalf     = 0.5
	rasterBlendExponent = 0.85

	// Ratio clamp bounds.
	ratioLimit = 10.0

	// vigor = vigorScale*(ratio+ratioLimit); with ratioLimit=10 and range 20,
	// vigorScale=0.05 maps ratio ∈ [-10,10] to vigor ∈ [0,1].
	vigorScale = 0.05
)

// CoronaTrendVigor is Ehlers' Corona Trend Vigor indicator.
type CoronaTrendVigor struct {
	mu                  sync.RWMutex
	mnemonic            string
	description         string
	mnemonicTV          string
	descriptionTV       string
	c                   *corona.Corona
	rasterLength        int
	rasterStep          float64
	maxRasterValue      float64
	minParameterValue   float64
	maxParameterValue   float64
	parameterResolution float64
	raster              []float64
	sampleBuffer        []float64
	sampleCount         int
	samplePrevious      float64
	samplePrevious2     float64
	bandPassPrevious    float64
	bandPassPrevious2   float64
	ratioPrevious       float64
	trendVigor          float64
	barFunc             entities.BarFunc
	quoteFunc           entities.QuoteFunc
	tradeFunc           entities.TradeFunc
}

// NewCoronaTrendVigorDefault returns an instance created with default parameters.
func NewCoronaTrendVigorDefault() (*CoronaTrendVigor, error) {
	return NewCoronaTrendVigorParams(&Params{})
}

// NewCoronaTrendVigorParams returns an instance created with the supplied parameters.
//
//nolint:funlen,cyclop
func NewCoronaTrendVigorParams(p *Params) (*CoronaTrendVigor, error) {
	const (
		invalid      = "invalid corona trend vigor parameters"
		fmtRaster    = "%s: RasterLength should be >= 2"
		fmtMaxRaster = "%s: MaxRasterValue should be > 0"
		fmtMaxParam  = "%s: MaxParameterValue should be > MinParameterValue"
		fmtHP        = "%s: HighPassFilterCutoff should be >= 2"
		fmtMinP      = "%s: MinimalPeriod should be >= 2"
		fmtMaxP      = "%s: MaximalPeriod should be > MinimalPeriod"
		fmtw         = "%s: %w"
		fmtnValue    = "ctv(%d, %g, %g, %g, %d%s)"
		fmtnTV       = "ctv-tv(%d%s)"
		descrValue   = "Corona trend vigor "
		descrTV      = "Corona trend vigor scalar "
		defRaster    = 50
		defMaxRast   = 20.0
		defMinParam  = -10.0
		defMaxParam  = 10.0
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

	// MinParameterValue and MaxParameterValue default to -10 and 10; since 0 is a
	// valid user value for either, we only substitute when both are zero (the
	// unconfigured case).
	if cfg.MinParameterValue == 0 && cfg.MaxParameterValue == 0 {
		cfg.MinParameterValue = defMinParam
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
	mnemonicTV := fmt.Sprintf(fmtnTV, cfg.HighPassFilterCutoff, componentMnemonic)

	parameterResolution := float64(cfg.RasterLength-1) / (cfg.MaxParameterValue - cfg.MinParameterValue)

	return &CoronaTrendVigor{
		mnemonic:            mnemonicValue,
		description:         descrValue + mnemonicValue,
		mnemonicTV:          mnemonicTV,
		descriptionTV:       descrTV + mnemonicTV,
		c:                   c,
		rasterLength:        cfg.RasterLength,
		rasterStep:          cfg.MaxRasterValue / float64(cfg.RasterLength),
		maxRasterValue:      cfg.MaxRasterValue,
		minParameterValue:   cfg.MinParameterValue,
		maxParameterValue:   cfg.MaxParameterValue,
		parameterResolution: parameterResolution,
		raster:              make([]float64, cfg.RasterLength),
		sampleBuffer:        make([]float64, c.MaximalPeriodTimesTwo()),
		trendVigor:          math.NaN(),
		barFunc:             barFunc,
		quoteFunc:           quoteFunc,
		tradeFunc:           tradeFunc,
	}, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *CoronaTrendVigor) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.c.IsPrimed()
}

// Metadata describes the output data of the indicator.
func (s *CoronaTrendVigor) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.CoronaTrendVigor,
		s.mnemonic,
		s.description,
		[]core.OutputText{
			{Mnemonic: s.mnemonic, Description: s.description},
			{Mnemonic: s.mnemonicTV, Description: s.descriptionTV},
		},
	)
}

// Update feeds the next sample and returns the heatmap column plus the current
// TrendVigor. On unprimed bars the heatmap is empty and the scalar is NaN.
//
//nolint:funlen,cyclop,gocognit
func (s *CoronaTrendVigor) Update(sample float64, t time.Time) (*outputs.Heatmap, float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if math.IsNaN(sample) {
		return outputs.NewEmptyHeatmap(t, s.minParameterValue, s.maxParameterValue, s.parameterResolution), math.NaN()
	}

	primed := s.c.Update(sample)
	s.sampleCount++

	bufLast := len(s.sampleBuffer) - 1

	if s.sampleCount == 1 {
		s.samplePrevious = sample
		s.sampleBuffer[bufLast] = sample

		return outputs.NewEmptyHeatmap(t, s.minParameterValue, s.maxParameterValue, s.parameterResolution), math.NaN()
	}

	// Bandpass InPhase filter at the dominant cycle median period.
	omega := 2. * math.Pi / s.c.DominantCycleMedian()
	beta2 := math.Cos(omega)
	gamma2 := 1. / math.Cos(omega*2*bpDelta)
	alpha2 := gamma2 - math.Sqrt(gamma2*gamma2-1.)
	bandPass := 0.5*(1-alpha2)*(sample-s.samplePrevious2) +
		beta2*(1+alpha2)*s.bandPassPrevious -
		alpha2*s.bandPassPrevious2

	// Quadrature = derivative / omega.
	quadrature2 := (bandPass - s.bandPassPrevious) / omega

	s.bandPassPrevious2 = s.bandPassPrevious
	s.bandPassPrevious = bandPass
	s.samplePrevious2 = s.samplePrevious
	s.samplePrevious = sample

	// Left-shift sampleBuffer and append the new sample.
	for i := 0; i < bufLast; i++ {
		s.sampleBuffer[i] = s.sampleBuffer[i+1]
	}

	s.sampleBuffer[bufLast] = sample

	// Cycle amplitude.
	amplitude2 := math.Sqrt(bandPass*bandPass + quadrature2*quadrature2)

	// Trend amplitude taken over the cycle period. Use DominantCycleMedian-1
	// directly (the MBST implementation clamps to sampleBuffer length which
	// negates the intent; see the reference impls).
	cyclePeriod := int(s.c.DominantCycleMedian() - 1)
	if cyclePeriod > len(s.sampleBuffer) {
		cyclePeriod = len(s.sampleBuffer)
	}

	if cyclePeriod < 1 {
		cyclePeriod = 1
	}

	lookback := cyclePeriod
	if s.sampleCount < lookback {
		lookback = s.sampleCount
	}

	trend := sample - s.sampleBuffer[len(s.sampleBuffer)-lookback]

	ratio := 0.0
	if math.Abs(trend) > 0 && amplitude2 > 0 {
		ratio = ratioNewCoef*trend/amplitude2 + ratioPreviousCoef*s.ratioPrevious
	}

	if ratio > ratioLimit {
		ratio = ratioLimit
	} else if ratio < -ratioLimit {
		ratio = -ratioLimit
	}

	s.ratioPrevious = ratio

	// ratio ∈ [-10, 10] ⇒ vigor ∈ [0, 1].
	vigor := vigorScale * (ratio + ratioLimit)

	var width float64

	switch {
	case vigor >= vigorMidLow && vigor < vigorMid:
		width = vigor - (vigorMidLow - widthEdge)
	case vigor >= vigorMid && vigor <= vigorMidHigh:
		width = (vigorMidHigh + widthEdge) - vigor
	default:
		width = widthEdge
	}

	s.trendVigor = (s.maxParameterValue-s.minParameterValue)*vigor + s.minParameterValue

	vigorScaledToRasterLength := int(math.Round(float64(s.rasterLength) * vigor))
	vigorScaledToMaxRasterValue := vigor * s.maxRasterValue

	for i := 0; i < s.rasterLength; i++ {
		value := s.raster[i]

		if i == vigorScaledToRasterLength {
			value *= rasterBlendHalf
		} else {
			argument := vigorScaledToMaxRasterValue - s.rasterStep*float64(i)
			if i > vigorScaledToRasterLength {
				argument = -argument
			}

			if width > 0 {
				value = rasterBlendScale *
					(math.Pow(argument/width, rasterBlendExponent) + rasterBlendPrevious*value)
			}
		}

		switch {
		case value < 0:
			value = 0
		case value > s.maxRasterValue, vigor < vigorMidLow, vigor > vigorMidHigh:
			value = s.maxRasterValue
		}

		if math.IsNaN(value) {
			value = 0
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

	return heatmap, s.trendVigor
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *CoronaTrendVigor) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.updateEntity(sample.Time, sample.Value)
}

// UpdateBar updates the indicator given the next bar sample.
func (s *CoronaTrendVigor) UpdateBar(sample *entities.Bar) core.Output {
	return s.updateEntity(sample.Time, s.barFunc(sample))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *CoronaTrendVigor) UpdateQuote(sample *entities.Quote) core.Output {
	return s.updateEntity(sample.Time, s.quoteFunc(sample))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *CoronaTrendVigor) UpdateTrade(sample *entities.Trade) core.Output {
	return s.updateEntity(sample.Time, s.tradeFunc(sample))
}

func (s *CoronaTrendVigor) updateEntity(t time.Time, sample float64) core.Output {
	const length = 2

	heatmap, tv := s.Update(sample, t)

	output := make([]any, length)
	output[0] = heatmap
	output[1] = entities.Scalar{Time: t, Value: tv}

	return output
}
