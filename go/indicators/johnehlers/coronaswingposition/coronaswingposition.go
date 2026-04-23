// Package coronaswingposition implements Ehlers' Corona Swing Position heatmap
// indicator.
//
// The Corona swing position is estimated by correlating the prices with a
// perfect sine wave having the dominant cycle period. This correlation produces
// a smooth waveform that lets us better estimate the swing position and
// impending turning points.
//
// Reference: John Ehlers, "Measuring Cycle Periods", Stocks & Commodities,
// November 2008.
package coronaswingposition

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
	maxLeadListCount     = 50
	maxPositionListCount = 20

	// 60° phase-lead coefficients: lead60 = 0.5·BP + sin(60°)·Q ≈ 0.5·BP + 0.866·Q.
	lead60CoefBP = 0.5
	lead60CoefQ  = 0.866

	// Bandpass filter delta factor in γ = 1/cos(ω · 2 · 0.1).
	bpDelta = 0.1

	widthHighThreshold = 0.85
	widthHighSaturate  = 0.8
	widthNarrow        = 0.01
	widthScale         = 0.15

	rasterBlendExponent = 0.95
	rasterBlendHalf     = 0.5
)

// CoronaSwingPosition is Ehlers' Corona Swing Position indicator.
type CoronaSwingPosition struct {
	mu                  sync.RWMutex
	mnemonic            string
	description         string
	mnemonicSP          string
	descriptionSP       string
	c                   *corona.Corona
	rasterLength        int
	rasterStep          float64
	maxRasterValue      float64
	minParameterValue   float64
	maxParameterValue   float64
	parameterResolution float64
	raster              []float64
	leadList            []float64
	positionList        []float64
	samplePrevious      float64
	samplePrevious2     float64
	bandPassPrevious    float64
	bandPassPrevious2   float64
	swingPosition       float64
	isStarted           bool
	barFunc             entities.BarFunc
	quoteFunc           entities.QuoteFunc
	tradeFunc           entities.TradeFunc
}

// NewCoronaSwingPositionDefault returns an instance created with default parameters.
func NewCoronaSwingPositionDefault() (*CoronaSwingPosition, error) {
	return NewCoronaSwingPositionParams(&Params{})
}

// NewCoronaSwingPositionParams returns an instance created with the supplied parameters.
//
//nolint:funlen,cyclop
func NewCoronaSwingPositionParams(p *Params) (*CoronaSwingPosition, error) {
	const (
		invalid      = "invalid corona swing position parameters"
		fmtRaster    = "%s: RasterLength should be >= 2"
		fmtMaxRaster = "%s: MaxRasterValue should be > 0"
		fmtMaxParam  = "%s: MaxParameterValue should be > MinParameterValue"
		fmtHP        = "%s: HighPassFilterCutoff should be >= 2"
		fmtMinP      = "%s: MinimalPeriod should be >= 2"
		fmtMaxP      = "%s: MaximalPeriod should be > MinimalPeriod"
		fmtw         = "%s: %w"
		fmtnValue    = "cswing(%d, %g, %g, %g, %d%s)"
		fmtnSP       = "cswing-sp(%d%s)"
		descrValue   = "Corona swing position "
		descrSP      = "Corona swing position scalar "
		defRaster    = 50
		defMaxRast   = 20.0
		defMinParam  = -5.0
		defMaxParam  = 5.0
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

	// MinParameterValue and MaxParameterValue default to -5 and 5; since 0 is a
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
	mnemonicSP := fmt.Sprintf(fmtnSP, cfg.HighPassFilterCutoff, componentMnemonic)

	parameterResolution := float64(cfg.RasterLength-1) / (cfg.MaxParameterValue - cfg.MinParameterValue)

	return &CoronaSwingPosition{
		mnemonic:            mnemonicValue,
		description:         descrValue + mnemonicValue,
		mnemonicSP:          mnemonicSP,
		descriptionSP:       descrSP + mnemonicSP,
		c:                   c,
		rasterLength:        cfg.RasterLength,
		rasterStep:          cfg.MaxRasterValue / float64(cfg.RasterLength),
		maxRasterValue:      cfg.MaxRasterValue,
		minParameterValue:   cfg.MinParameterValue,
		maxParameterValue:   cfg.MaxParameterValue,
		parameterResolution: parameterResolution,
		raster:              make([]float64, cfg.RasterLength),
		leadList:            make([]float64, 0, maxLeadListCount),
		positionList:        make([]float64, 0, maxPositionListCount),
		swingPosition:       math.NaN(),
		barFunc:             barFunc,
		quoteFunc:           quoteFunc,
		tradeFunc:           tradeFunc,
	}, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *CoronaSwingPosition) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.c.IsPrimed()
}

// Metadata describes the output data of the indicator.
func (s *CoronaSwingPosition) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.CoronaSwingPosition,
		s.mnemonic,
		s.description,
		[]core.OutputText{
			{Mnemonic: s.mnemonic, Description: s.description},
			{Mnemonic: s.mnemonicSP, Description: s.descriptionSP},
		},
	)
}

// Update feeds the next sample and returns the heatmap column plus the current
// SwingPosition. On unprimed bars the heatmap is empty and the scalar is NaN.
//
//nolint:funlen,cyclop,gocognit
func (s *CoronaSwingPosition) Update(sample float64, t time.Time) (*outputs.Heatmap, float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if math.IsNaN(sample) {
		return outputs.NewEmptyHeatmap(t, s.minParameterValue, s.maxParameterValue, s.parameterResolution), math.NaN()
	}

	primed := s.c.Update(sample)

	if !s.isStarted {
		s.samplePrevious = sample
		s.isStarted = true

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

	// 60° lead: lead60 = 0.5·BP_previous2 + 0.866·Q
	lead60 := lead60CoefBP*s.bandPassPrevious2 + lead60CoefQ*quadrature2

	lowest, highest := appendRolling(&s.leadList, maxLeadListCount, lead60)

	// Normalised lead position ∈ [0, 1].
	position := highest - lowest
	if position > 0 {
		position = (lead60 - lowest) / position
	}

	lowest, highest = appendRolling(&s.positionList, maxPositionListCount, position)
	highest -= lowest

	width := 0.15 * highest
	if highest > widthHighThreshold {
		width = widthNarrow
	}

	s.swingPosition = (s.maxParameterValue-s.minParameterValue)*position + s.minParameterValue

	positionScaledToRasterLength := int(math.Round(position * float64(s.rasterLength)))
	positionScaledToMaxRasterValue := position * s.maxRasterValue

	for i := 0; i < s.rasterLength; i++ {
		value := s.raster[i]

		if i == positionScaledToRasterLength {
			value *= rasterBlendHalf
		} else {
			argument := positionScaledToMaxRasterValue - s.rasterStep*float64(i)
			if i > positionScaledToRasterLength {
				argument = -argument
			}

			if width > 0 {
				value = rasterBlendHalf *
					(math.Pow(argument/width, rasterBlendExponent) + rasterBlendHalf*value)
			}
		}

		if value < 0 {
			value = 0
		} else if value > s.maxRasterValue {
			value = s.maxRasterValue
		}

		if highest > widthHighSaturate {
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

	return heatmap, s.swingPosition
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *CoronaSwingPosition) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.updateEntity(sample.Time, sample.Value)
}

// UpdateBar updates the indicator given the next bar sample.
func (s *CoronaSwingPosition) UpdateBar(sample *entities.Bar) core.Output {
	return s.updateEntity(sample.Time, s.barFunc(sample))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *CoronaSwingPosition) UpdateQuote(sample *entities.Quote) core.Output {
	return s.updateEntity(sample.Time, s.quoteFunc(sample))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *CoronaSwingPosition) UpdateTrade(sample *entities.Trade) core.Output {
	return s.updateEntity(sample.Time, s.tradeFunc(sample))
}

func (s *CoronaSwingPosition) updateEntity(t time.Time, sample float64) core.Output {
	const length = 2

	heatmap, sp := s.Update(sample, t)

	output := make([]any, length)
	output[0] = heatmap
	output[1] = entities.Scalar{Time: t, Value: sp}

	return output
}

// appendRolling appends v to the list, drops the oldest element once the list
// reaches maxCount, and returns the current (lowest, highest) values.
func appendRolling(list *[]float64, maxCount int, v float64) (float64, float64) {
	if len(*list) >= maxCount {
		*list = (*list)[1:]
	}

	*list = append(*list, v)

	lowest := v
	highest := v

	for _, x := range *list {
		if x < lowest {
			lowest = x
		}

		if x > highest {
			highest = x
		}
	}

	return lowest, highest
}
