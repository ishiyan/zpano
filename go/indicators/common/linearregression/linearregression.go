package linearregression

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

// LinearRegression computes the least-squares regression line over a rolling window
// and produces five outputs per sample:
//
//   - Value:     b + m*(period-1)  — the regression value at the last bar of the window
//   - Forecast:  b + m*period      — the time series forecast (one bar ahead)
//   - Intercept: b                 — the y-intercept of the regression line
//   - SlopeRad:  m                 — the slope of the regression line
//   - SlopeDeg:  atan(m)*180/π    — the slope expressed in degrees
//
// where y = b + m*x is the best-fit line (x = 0 … period-1).
//
// The indicator is not primed during the first (period-1) updates.
type LinearRegression struct {
	mu          sync.RWMutex
	mnemonic    string
	description string
	length      int
	lengthF     float64
	sumX        float64
	divisor     float64
	window      []float64
	windowCount int
	primed      bool
	// Current output values (stored for updateEntity).
	curValue     float64
	curForecast  float64
	curIntercept float64
	curSlopeRad  float64
	curSlopeDeg  float64
	barFunc      entities.BarFunc
	quoteFunc    entities.QuoteFunc
	tradeFunc    entities.TradeFunc
}

// New returns an instance of the indicator created using supplied parameters.
func New(p *Params) (*LinearRegression, error) {
	const (
		invalid = "invalid linear regression parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "linreg(%d%s)"
		descPfx = "Linear Regression "
		minlen  = 2
	)

	length := p.Length
	if length < minlen {
		return nil, fmt.Errorf(fmts, invalid, "length should be greater than 1")
	}

	var (
		err       error
		barFunc   entities.BarFunc
		quoteFunc entities.QuoteFunc
		tradeFunc entities.TradeFunc
	)

	bc := p.BarComponent
	qc := p.QuoteComponent
	tc := p.TradeComponent

	if bc == 0 {
		bc = entities.DefaultBarComponent
	}

	if qc == 0 {
		qc = entities.DefaultQuoteComponent
	}

	if tc == 0 {
		tc = entities.DefaultTradeComponent
	}

	if barFunc, err = entities.BarComponentFunc(bc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	if quoteFunc, err = entities.QuoteComponentFunc(qc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	if tradeFunc, err = entities.TradeComponentFunc(tc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	mnemonic := fmt.Sprintf(fmtn, length, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := descPfx + mnemonic

	n := float64(length)
	sumX := n * (n - 1) * 0.5
	sumXSqr := n * (n - 1) * (2*n - 1) / 6
	divisor := sumX*sumX - n*sumXSqr

	return &LinearRegression{
		mnemonic:    mnemonic,
		description: desc,
		length:      length,
		lengthF:     n,
		sumX:        sumX,
		divisor:     divisor,
		window:      make([]float64, length),
		barFunc:     barFunc,
		quoteFunc:   quoteFunc,
		tradeFunc:   tradeFunc,
	}, nil
}

// IsPrimed indicates whether an indicator is primed.
func (s *LinearRegression) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *LinearRegression) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.LinearRegression,
		Mnemonic:    s.mnemonic,
		Description: s.description,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(Value),
				Type:        outputs.ScalarType,
				Mnemonic:    s.mnemonic,
				Description: s.description + " value",
			},
			{
				Kind:        int(Forecast),
				Type:        outputs.ScalarType,
				Mnemonic:    s.mnemonic,
				Description: s.description + " forecast",
			},
			{
				Kind:        int(Intercept),
				Type:        outputs.ScalarType,
				Mnemonic:    s.mnemonic,
				Description: s.description + " intercept",
			},
			{
				Kind:        int(SlopeRad),
				Type:        outputs.ScalarType,
				Mnemonic:    s.mnemonic,
				Description: s.description + " slope",
			},
			{
				Kind:        int(SlopeDeg),
				Type:        outputs.ScalarType,
				Mnemonic:    s.mnemonic,
				Description: s.description + " angle",
			},
		},
	}
}

// Update updates the indicator given the next sample and returns the Value output.
//
// To obtain all five outputs, use UpdateScalar/UpdateBar/UpdateQuote/UpdateTrade.
func (s *LinearRegression) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.primed {
		s.calculate(sample)

		return s.curValue
	}

	s.window[s.windowCount] = sample
	s.windowCount++

	if s.windowCount == s.length {
		s.primed = true
		s.computeFromWindow()

		return s.curValue
	}

	return math.NaN()
}

func (s *LinearRegression) calculate(sample float64) {
	// Shift window.
	for i := range s.length - 1 {
		s.window[i] = s.window[i+1]
	}

	s.window[s.length-1] = sample
	s.computeFromWindow()
}

func (s *LinearRegression) computeFromWindow() {
	const radToDeg = 180.0 / math.Pi

	var sumXY, sumY float64

	for i := s.length; i > 0; i-- {
		v := s.window[s.length-i]
		sumY += v
		sumXY += float64(i-1) * v
	}

	m := (s.lengthF*sumXY - s.sumX*sumY) / s.divisor
	b := (sumY - m*s.sumX) / s.lengthF

	s.curSlopeRad = m
	s.curSlopeDeg = math.Atan(m) * radToDeg
	s.curIntercept = b
	s.curValue = b + m*(s.lengthF-1)
	s.curForecast = b + m*s.lengthF
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *LinearRegression) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.updateEntity(sample.Time, sample.Value)
}

// UpdateBar updates the indicator given the next bar sample.
func (s *LinearRegression) UpdateBar(sample *entities.Bar) core.Output {
	return s.updateEntity(sample.Time, s.barFunc(sample))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *LinearRegression) UpdateQuote(sample *entities.Quote) core.Output {
	return s.updateEntity(sample.Time, s.quoteFunc(sample))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *LinearRegression) UpdateTrade(sample *entities.Trade) core.Output {
	return s.updateEntity(sample.Time, s.tradeFunc(sample))
}

func (s *LinearRegression) updateEntity(
	t time.Time, sample float64,
) core.Output {
	const numOutputs = 5

	output := make([]any, numOutputs)
	val := s.Update(sample)

	if math.IsNaN(val) {
		nan := math.NaN()

		i := 0
		output[i] = entities.Scalar{Time: t, Value: nan}
		i++
		output[i] = entities.Scalar{Time: t, Value: nan}
		i++
		output[i] = entities.Scalar{Time: t, Value: nan}
		i++
		output[i] = entities.Scalar{Time: t, Value: nan}
		i++
		output[i] = entities.Scalar{Time: t, Value: nan}

		return output
	}

	i := 0
	output[i] = entities.Scalar{Time: t, Value: s.curValue}
	i++
	output[i] = entities.Scalar{Time: t, Value: s.curForecast}
	i++
	output[i] = entities.Scalar{Time: t, Value: s.curIntercept}
	i++
	output[i] = entities.Scalar{Time: t, Value: s.curSlopeRad}
	i++
	output[i] = entities.Scalar{Time: t, Value: s.curSlopeDeg}

	return output
}
