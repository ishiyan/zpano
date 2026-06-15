package instantaneoussinewaveperiod

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// InstantaneousSineWavePeriod is Don Mak's Instantaneous Sine Wave Period (ISWP) indicator.
//
// It estimates the dominant cycle period of price data by modeling it locally as a
// single sine wave superimposed on a constant level, combining a 4-point method
// (IF4) and a 5-point method (IF5) and selecting the one with the lower estimation
// error at each bar.
//
// The indicator produces seven outputs:
//   - Period: cycle period in bars (T = 2*pi/omega), NaN if invalid;
//   - Omega: circular frequency in radians/bar, NaN if invalid;
//   - Velocity: wave velocity, NaN if invalid;
//   - Acceleration: wave acceleration, NaN if invalid;
//   - Amplitude: sine wave amplitude, NaN if invalid;
//   - Phase: phase angle in radians, NaN if invalid;
//   - DcLevel: constant level D, NaN if invalid.
//
// Reference:
//
// Mak, Don K. (2006). Mathematical Techniques in Financial Market Trading.
type InstantaneousSineWavePeriod struct {
	mu sync.RWMutex

	smoothing      int
	minPeriod      float64
	maxPeriod      float64
	errorThreshold float64
	dx             float64

	emaAlpha  float64
	emaValue  float64
	emaPrimed bool

	buffer [5]float64
	count  int

	primed bool

	barFunc   entities.BarFunc
	quoteFunc entities.QuoteFunc
	tradeFunc entities.TradeFunc

	mnemonic string
}

// NewInstantaneousSineWavePeriod returns an instance of the indicator created using supplied parameters.
//
//nolint:funlen,cyclop
func NewInstantaneousSineWavePeriod(p *Params) (*InstantaneousSineWavePeriod, error) {
	const (
		invalid               = "invalid instantaneous sine wave period parameters"
		fmts                  = "%s: %s"
		fmtw                  = "%s: %w"
		defaultMinPeriod      = 4.0
		defaultMaxPeriod      = 50.0
		defaultErrorThreshold = 20.0
		defaultDx             = 0.01
	)

	smoothing := p.Smoothing

	minPeriod := p.MinPeriod
	if minPeriod == 0 {
		minPeriod = defaultMinPeriod
	}

	maxPeriod := p.MaxPeriod
	if maxPeriod == 0 {
		maxPeriod = defaultMaxPeriod
	}

	errorThreshold := p.ErrorThreshold
	if errorThreshold == 0 {
		errorThreshold = defaultErrorThreshold
	}

	dx := p.Dx
	if dx == 0 {
		dx = defaultDx
	}

	if smoothing < 0 {
		return nil, fmt.Errorf(fmts, invalid, "smoothing should be >= 0")
	}

	if minPeriod <= 0.0 {
		return nil, fmt.Errorf(fmts, invalid, "minPeriod should be > 0")
	}

	if maxPeriod <= minPeriod {
		return nil, fmt.Errorf(fmts, invalid, "maxPeriod should be > minPeriod")
	}

	if errorThreshold <= 0.0 {
		return nil, fmt.Errorf(fmts, invalid, "errorThreshold should be > 0")
	}

	if dx <= 0.0 {
		return nil, fmt.Errorf(fmts, invalid, "dx should be > 0")
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

	emaAlpha := 1.0
	if smoothing > 0 {
		emaAlpha = 2.0 / (float64(smoothing) + 1.0)
	}

	mnemonic := fmt.Sprintf("iswp(%d,%.2f,%.2f,%.2f,%.2f%s)", smoothing, minPeriod, maxPeriod,
		errorThreshold, dx, core.ComponentTripleMnemonic(bc, qc, tc))

	return &InstantaneousSineWavePeriod{
		smoothing:      smoothing,
		minPeriod:      minPeriod,
		maxPeriod:      maxPeriod,
		errorThreshold: errorThreshold,
		dx:             dx,
		emaAlpha:       emaAlpha,
		barFunc:        barFunc,
		quoteFunc:      quoteFunc,
		tradeFunc:      tradeFunc,
		mnemonic:       mnemonic,
	}, nil
}

func (s *InstantaneousSineWavePeriod) applyEMA(price float64) float64 {
	if !s.emaPrimed {
		s.emaValue = price
		s.emaPrimed = true
	} else {
		s.emaValue = s.emaAlpha*price + (1.0-s.emaAlpha)*s.emaValue
	}

	return s.emaValue
}

func (s *InstantaneousSineWavePeriod) pushBuffer(value float64) {
	for i := 4; i > 0; i-- {
		s.buffer[i] = s.buffer[i-1]
	}

	s.buffer[0] = value
}

//nolint:nonamedreturns
func (s *InstantaneousSineWavePeriod) calcOmega4() (omega, errVal float64) {
	x0 := s.buffer[0]
	xm1 := s.buffer[1]
	xm2 := s.buffer[2]
	xm3 := s.buffer[3]

	den := xm1 - xm2
	if den == 0.0 {
		return math.NaN(), s.errorThreshold
	}

	ratio := (x0 - xm3) / den

	sqrtArg := 3.0 - ratio
	if sqrtArg < 0.0 {
		return math.NaN(), s.errorThreshold
	}

	arg := 0.5 * math.Sqrt(sqrtArg)
	if arg > 1.0 {
		return math.NaN(), s.errorThreshold
	}

	omega4 := 2.0 * math.Asin(arg)

	dx2 := s.dx * s.dx

	denom1 := 1.0 - 0.25*sqrtArg
	if denom1 <= 0.0 || sqrtArg == 0.0 {
		return omega4, s.errorThreshold
	}

	f1 := 1.0 / (denom1 * sqrtArg)
	invDen2 := 1.0 / (den * den)
	q2 := invDen2*(dx2+dx2) + (ratio*ratio)*invDen2*(dx2+dx2)

	product := f1 * q2
	if product < 0.0 {
		return omega4, s.errorThreshold
	}

	return omega4, 0.5 * math.Sqrt(product)
}

//nolint:nonamedreturns
func (s *InstantaneousSineWavePeriod) calcOmega5() (omega, errVal float64) {
	x0 := s.buffer[0]
	xm1 := s.buffer[1]
	xm3 := s.buffer[3]
	xm4 := s.buffer[4]

	den1 := xm1 - xm3
	if den1 == 0.0 {
		return math.NaN(), s.errorThreshold
	}

	arg := 0.5 * (x0 - xm4) / den1
	if math.Abs(arg) > 1.0 {
		return math.NaN(), s.errorThreshold
	}

	omega5 := math.Acos(arg)

	dx2 := s.dx * s.dx

	denom := 1.0 - arg*arg
	if denom <= 0.0 {
		return omega5, s.errorThreshold
	}

	f1 := 1.0 / denom
	invDen1Sq := 1.0 / (den1 * den1)
	numeratorRatio := (x0 - xm4) / (den1 * den1)
	r2 := invDen1Sq*(dx2+dx2) + (numeratorRatio*numeratorRatio)*(dx2+dx2)

	product := f1 * r2
	if product < 0.0 {
		return omega5, s.errorThreshold
	}

	return omega5, 0.5 * math.Sqrt(product)
}

//nolint:nonamedreturns
func (s *InstantaneousSineWavePeriod) calcModelParams(omega float64) (amplitude, phase, velocity, acceleration, dcLevel float64) {
	x0 := s.buffer[0]
	xm1 := s.buffer[1]
	xm2 := s.buffer[2]

	halfW := omega / 2.0
	threeHalfW := 1.5 * omega

	sinHW := math.Sin(halfW)
	cosHW := math.Cos(halfW)
	sin3HW := math.Sin(threeHalfW)
	cos3HW := math.Cos(threeHalfW)

	d0 := sinHW*sinHW*cosHW*sin3HW - sinHW*sinHW*sinHW*cos3HW

	nan := math.NaN()
	if math.Abs(d0) < 1e-15 {
		return nan, nan, nan, nan, nan
	}

	invD0 := 1.0 / d0

	dx0M1 := x0 - xm1
	dxm1M2 := xm1 - xm2

	c := invD0 * (dx0M1*sinHW*sin3HW - dxm1M2*sinHW*sinHW)
	sineComponent := invD0 * (dxm1M2*sinHW*cosHW - dx0M1*sinHW*cos3HW)

	amplitude = 0.5 * math.Sqrt(c*c+sineComponent*sineComponent)
	phase = math.Atan2(sineComponent, c)
	velocity = amplitude * omega * math.Cos(phase)
	acceleration = -amplitude * omega * omega * math.Sin(phase)
	dcLevel = x0 - sineComponent/2.0

	return amplitude, phase, velocity, acceleration, dcLevel
}

// IsPrimed indicates whether the indicator is primed.
func (s *InstantaneousSineWavePeriod) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *InstantaneousSineWavePeriod) Metadata() core.Metadata {
	desc := "Instantaneous Sine Wave Period " + s.mnemonic

	return core.BuildMetadata(
		core.InstantaneousSineWavePeriod,
		s.mnemonic,
		desc,
		[]core.OutputText{
			{Mnemonic: s.mnemonic + " period", Description: desc + " Period"},
			{Mnemonic: s.mnemonic + " omega", Description: desc + " Omega"},
			{Mnemonic: s.mnemonic + " velocity", Description: desc + " Velocity"},
			{Mnemonic: s.mnemonic + " acceleration", Description: desc + " Acceleration"},
			{Mnemonic: s.mnemonic + " amplitude", Description: desc + " Amplitude"},
			{Mnemonic: s.mnemonic + " phase", Description: desc + " Phase"},
			{Mnemonic: s.mnemonic + " dcLevel", Description: desc + " DC Level"},
		},
	)
}

// Update updates the indicator given the next sample value.
// Returns period, omega, velocity, acceleration, amplitude, phase, dcLevel.
//
//nolint:nonamedreturns,cyclop
func (s *InstantaneousSineWavePeriod) Update(
	sample float64,
) (period, omega, velocity, acceleration, amplitude, phase, dcLevel float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	nan := math.NaN()

	smoothed := sample
	if s.smoothing > 0 {
		smoothed = s.applyEMA(sample)
	}

	s.pushBuffer(smoothed)
	s.count++

	if s.count < 5 {
		return nan, nan, nan, nan, nan, nan, nan
	}

	omega4, error4 := s.calcOmega4()
	omega5, error5 := s.calcOmega5()

	if error4 >= s.errorThreshold && error5 >= s.errorThreshold {
		return nan, nan, nan, nan, nan, nan, nan
	}

	omega = omega4
	if error5 < error4 {
		omega = omega5
	}

	if math.IsNaN(omega) || omega <= 0.0 {
		return nan, nan, nan, nan, nan, nan, nan
	}

	period = (2.0 * math.Pi) / omega
	if period < s.minPeriod || period > s.maxPeriod {
		return nan, nan, nan, nan, nan, nan, nan
	}

	amplitude, phase, velocity, acceleration, dcLevel = s.calcModelParams(omega)

	s.primed = true

	return period, omega, velocity, acceleration, amplitude, phase, dcLevel
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *InstantaneousSineWavePeriod) UpdateScalar(sample *entities.Scalar) core.Output {
	period, omega, velocity, acceleration, amplitude, phase, dcLevel := s.Update(sample.Value)

	const outputCount = 7

	output := make([]any, outputCount)
	output[0] = entities.Scalar{Time: sample.Time, Value: period}
	output[1] = entities.Scalar{Time: sample.Time, Value: omega}
	output[2] = entities.Scalar{Time: sample.Time, Value: velocity}
	output[3] = entities.Scalar{Time: sample.Time, Value: acceleration}
	output[4] = entities.Scalar{Time: sample.Time, Value: amplitude}
	output[5] = entities.Scalar{Time: sample.Time, Value: phase}
	output[6] = entities.Scalar{Time: sample.Time, Value: dcLevel}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (s *InstantaneousSineWavePeriod) UpdateBar(sample *entities.Bar) core.Output {
	v := s.barFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *InstantaneousSineWavePeriod) UpdateQuote(sample *entities.Quote) core.Output {
	v := s.quoteFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *InstantaneousSineWavePeriod) UpdateTrade(sample *entities.Trade) core.Output {
	v := s.tradeFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}
