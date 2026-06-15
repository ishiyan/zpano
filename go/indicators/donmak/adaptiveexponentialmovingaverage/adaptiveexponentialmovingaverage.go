package adaptiveexponentialmovingaverage

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// iswp is an embedded Instantaneous Sine Wave Period omega estimator
// (omega-only reduction). It estimates the dominant circular frequency of price
// data by modeling it locally as a single sine wave, combining a 4-point and a
// 5-point method and selecting the one with the lower estimation error. Inlined
// so the indicator is a standalone porting unit. Do NOT change its numerics.
type iswp struct {
	smoothing int
	emaAlpha  float64
	emaValue  float64
	emaPrimed bool
	buffer    [5]float64
	count     int
}

const (
	iswpMinPeriod      = 4.0
	iswpMaxPeriod      = 50.0
	iswpErrorThreshold = 20.0
	iswpDx             = 0.01
)

func newISWP(smoothing int) *iswp {
	emaAlpha := 1.0
	if smoothing > 0 {
		emaAlpha = 2.0 / (float64(smoothing) + 1.0)
	}

	return &iswp{smoothing: smoothing, emaAlpha: emaAlpha}
}

func (s *iswp) applyEMA(price float64) float64 {
	if !s.emaPrimed {
		s.emaValue = price
		s.emaPrimed = true
	} else {
		s.emaValue = s.emaAlpha*price + (1.0-s.emaAlpha)*s.emaValue
	}

	return s.emaValue
}

func (s *iswp) pushBuffer(value float64) {
	for i := 4; i > 0; i-- {
		s.buffer[i] = s.buffer[i-1]
	}

	s.buffer[0] = value
}

func (s *iswp) calcOmega4() (omega, errVal float64) {
	x0 := s.buffer[0]
	xm1 := s.buffer[1]
	xm2 := s.buffer[2]
	xm3 := s.buffer[3]

	den := xm1 - xm2
	if den == 0.0 {
		return math.NaN(), iswpErrorThreshold
	}

	ratio := (x0 - xm3) / den

	sqrtArg := 3.0 - ratio
	if sqrtArg < 0.0 {
		return math.NaN(), iswpErrorThreshold
	}

	arg := 0.5 * math.Sqrt(sqrtArg)
	if arg > 1.0 {
		return math.NaN(), iswpErrorThreshold
	}

	omega4 := 2.0 * math.Asin(arg)

	dx2 := iswpDx * iswpDx

	denom1 := 1.0 - 0.25*sqrtArg
	if denom1 <= 0.0 || sqrtArg == 0.0 {
		return omega4, iswpErrorThreshold
	}

	f1 := 1.0 / (denom1 * sqrtArg)
	invDen2 := 1.0 / (den * den)
	q2 := invDen2*(dx2+dx2) + (ratio*ratio)*invDen2*(dx2+dx2)

	product := f1 * q2
	if product < 0.0 {
		return omega4, iswpErrorThreshold
	}

	return omega4, 0.5 * math.Sqrt(product)
}

func (s *iswp) calcOmega5() (omega, errVal float64) {
	x0 := s.buffer[0]
	xm1 := s.buffer[1]
	xm3 := s.buffer[3]
	xm4 := s.buffer[4]

	den1 := xm1 - xm3
	if den1 == 0.0 {
		return math.NaN(), iswpErrorThreshold
	}

	arg := 0.5 * (x0 - xm4) / den1
	if math.Abs(arg) > 1.0 {
		return math.NaN(), iswpErrorThreshold
	}

	omega5 := math.Acos(arg)

	dx2 := iswpDx * iswpDx

	denom := 1.0 - arg*arg
	if denom <= 0.0 {
		return omega5, iswpErrorThreshold
	}

	f1 := 1.0 / denom
	invDen1Sq := 1.0 / (den1 * den1)
	numeratorRatio := (x0 - xm4) / (den1 * den1)
	r2 := invDen1Sq*(dx2+dx2) + (numeratorRatio*numeratorRatio)*(dx2+dx2)

	product := f1 * r2
	if product < 0.0 {
		return omega5, iswpErrorThreshold
	}

	return omega5, 0.5 * math.Sqrt(product)
}

// update processes one price and returns the omega estimate (NaN if unavailable).
func (s *iswp) update(price float64) float64 {
	smoothed := price
	if s.smoothing > 0 {
		smoothed = s.applyEMA(price)
	}

	s.pushBuffer(smoothed)
	s.count++

	if s.count < 5 {
		return math.NaN()
	}

	omega4, error4 := s.calcOmega4()
	omega5, error5 := s.calcOmega5()

	if error4 >= iswpErrorThreshold && error5 >= iswpErrorThreshold {
		return math.NaN()
	}

	omega := omega4
	if error5 < error4 {
		omega = omega5
	}

	if math.IsNaN(omega) || omega <= 0.0 {
		return math.NaN()
	}

	period := (2.0 * math.Pi) / omega
	if period < iswpMinPeriod || period > iswpMaxPeriod {
		return math.NaN()
	}

	return omega
}

// AdaptiveExponentialMovingAverage is Don Mak's Adaptive Exponential Moving Average (AEMA).
//
// It is an EMA with a time-varying smoothing factor alpha that adapts based on
// the instantaneous frequency of the price data, estimated by an embedded ISWP.
//
// The indicator produces three outputs:
//   - Value: the adaptively smoothed price (never NaN);
//   - Omega: the instantaneous frequency estimate (may be NaN);
//   - Alpha: the smoothing factor used for this bar.
//
// Reference:
//
// Mak, D.K. (2006). Mathematical Techniques in Financial Market Trading.
type AdaptiveExponentialMovingAverage struct {
	mu sync.RWMutex

	alphaMax float64
	alphaMin float64
	omega0   float64
	a        float64
	b        float64

	iswp *iswp

	emaValue    float64
	initialized bool
	primed      bool

	barFunc   entities.BarFunc
	quoteFunc entities.QuoteFunc
	tradeFunc entities.TradeFunc

	mnemonic string
}

// NewAdaptiveExponentialMovingAverage returns an instance of the indicator created using supplied parameters.
//
//nolint:funlen,cyclop
func NewAdaptiveExponentialMovingAverage(p *Params) (*AdaptiveExponentialMovingAverage, error) {
	const (
		invalid          = "invalid adaptive exponential moving average parameters"
		fmts             = "%s: %s"
		fmtw             = "%s: %w"
		defaultAlphaMax  = 0.5
		defaultAlphaMin  = 0.05
		defaultOmega0    = 1.0
		defaultSmoothing = 3
	)

	alphaMax := p.AlphaMax
	if alphaMax == 0 {
		alphaMax = defaultAlphaMax
	}

	alphaMin := p.AlphaMin
	if alphaMin == 0 {
		alphaMin = defaultAlphaMin
	}

	omega0 := p.Omega0
	if omega0 == 0 {
		omega0 = defaultOmega0
	}

	smoothing := p.Smoothing

	if !(alphaMin > 0.0 && alphaMin < alphaMax && alphaMax <= 1.0) {
		return nil, fmt.Errorf(fmts, invalid, "need 0 < alphaMin < alphaMax <= 1")
	}

	if !(omega0 > 0.0 && omega0 < math.Pi) {
		return nil, fmt.Errorf(fmts, invalid, "need 0 < omega0 < pi")
	}

	if smoothing < 0 {
		return nil, fmt.Errorf(fmts, invalid, "smoothing should be >= 0")
	}

	bc := p.BarComponent
	if bc == 0 {
		bc = entities.DefaultBarComponent
	}

	qc := p.QuoteComponent
	if qc == 0 {
		qc = entities.DefaultQuoteComponent
	}

	tc := p.TradeComponent
	if tc == 0 {
		tc = entities.DefaultTradeComponent
	}

	var (
		err       error
		barFunc   entities.BarFunc
		quoteFunc entities.QuoteFunc
		tradeFunc entities.TradeFunc
	)

	if barFunc, err = entities.BarComponentFunc(bc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	if quoteFunc, err = entities.QuoteComponentFunc(qc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	if tradeFunc, err = entities.TradeComponentFunc(tc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	a := (alphaMax - alphaMin) * omega0 * math.Pi / (math.Pi - omega0)
	b := alphaMin - a/math.Pi

	mnemonic := fmt.Sprintf("aema(%.2f,%.2f,%.2f,%d%s)", alphaMax, alphaMin, omega0, smoothing,
		core.ComponentTripleMnemonic(bc, qc, tc))

	return &AdaptiveExponentialMovingAverage{
		alphaMax:  alphaMax,
		alphaMin:  alphaMin,
		omega0:    omega0,
		a:         a,
		b:         b,
		iswp:      newISWP(smoothing),
		barFunc:   barFunc,
		quoteFunc: quoteFunc,
		tradeFunc: tradeFunc,
		mnemonic:  mnemonic,
	}, nil
}

func (s *AdaptiveExponentialMovingAverage) computeAlpha(omega float64) float64 {
	if math.IsNaN(omega) {
		return s.alphaMin
	}

	if omega <= s.omega0 {
		return s.alphaMax
	}

	if omega >= math.Pi {
		return s.alphaMin
	}

	alpha := s.a/omega + s.b
	if alpha > s.alphaMax {
		return s.alphaMax
	}

	if alpha < s.alphaMin {
		return s.alphaMin
	}

	return alpha
}

// IsPrimed indicates whether the indicator is primed.
func (s *AdaptiveExponentialMovingAverage) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *AdaptiveExponentialMovingAverage) Metadata() core.Metadata {
	desc := "Adaptive Exponential Moving Average " + s.mnemonic

	return core.BuildMetadata(
		core.AdaptiveExponentialMovingAverage,
		s.mnemonic,
		desc,
		[]core.OutputText{
			{Mnemonic: s.mnemonic + " value", Description: desc + " Value"},
			{Mnemonic: s.mnemonic + " omega", Description: desc + " Omega"},
			{Mnemonic: s.mnemonic + " alpha", Description: desc + " Alpha"},
		},
	)
}

// Update updates the indicator given the next sample value.
// Returns value, omega, alpha.
//
//nolint:nonamedreturns
func (s *AdaptiveExponentialMovingAverage) Update(sample float64) (value, omega, alpha float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	omega = s.iswp.update(sample)
	alpha = s.computeAlpha(omega)

	if !s.initialized {
		s.emaValue = sample
		s.initialized = true
	} else {
		s.emaValue = alpha*sample + (1.0-alpha)*s.emaValue
	}

	if !math.IsNaN(omega) {
		s.primed = true
	}

	return s.emaValue, omega, alpha
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *AdaptiveExponentialMovingAverage) UpdateScalar(sample *entities.Scalar) core.Output {
	value, omega, alpha := s.Update(sample.Value)

	const outputCount = 3

	output := make([]any, outputCount)
	output[0] = entities.Scalar{Time: sample.Time, Value: value}
	output[1] = entities.Scalar{Time: sample.Time, Value: omega}
	output[2] = entities.Scalar{Time: sample.Time, Value: alpha}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (s *AdaptiveExponentialMovingAverage) UpdateBar(sample *entities.Bar) core.Output {
	v := s.barFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *AdaptiveExponentialMovingAverage) UpdateQuote(sample *entities.Quote) core.Output {
	v := s.quoteFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *AdaptiveExponentialMovingAverage) UpdateTrade(sample *entities.Trade) core.Output {
	v := s.tradeFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}
