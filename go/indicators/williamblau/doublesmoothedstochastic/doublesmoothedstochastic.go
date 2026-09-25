package doublesmoothedstochastic

import (
	"cmp"
	"fmt"
	"math"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
)

// ema is a stateful streaming exponential moving average: alpha = 2/(period+1),
// seeds e0 = x0. Inlined verbatim from the Blau exponential moving average so the
// indicator is a standalone porting unit. Do NOT change its numerics.
//
// period == 1 -> alpha == 1 -> pure passthrough (output == input).
type ema struct {
	alpha    float64
	previous float64
	primed   bool
}

func newEMA(period int) *ema {
	return &ema{alpha: 2.0 / (float64(period) + 1.0)}
}

func (e *ema) update(x float64) float64 {
	if !e.primed {
		e.previous = x
		e.primed = true

		return e.previous
	}

	e.previous = e.alpha*x + (1.0-e.alpha)*e.previous

	return e.previous
}

// sma is a stateful streaming simple moving average over the last period inputs.
//
// It returns the mean of the window's current contents on every update: an
// expanding window while fewer than period values have arrived, then a rolling
// period-bar window. There is no NaN warm-up (finite from the first input).
//
// period == 1 -> pure passthrough (output == input).
type sma struct {
	period int
	window []float64
	count  int
	index  int
}

func newSMA(period int) *sma {
	return &sma{period: period, window: make([]float64, period)}
}

func (s *sma) update(x float64) float64 {
	s.window[s.index] = x
	s.index = (s.index + 1) % s.period

	if s.count < s.period {
		s.count++
	}

	// Naive left-to-right sum from the oldest to the newest value (NOT a
	// compensated sum), so that every port reproduces the same values.
	// Once full, the oldest value sits at the next write position.
	start := 0
	if s.count == s.period {
		start = s.index
	}

	sum := 0.0
	for i := range s.count {
		sum += s.window[(start+i)%s.period]
	}

	return sum / float64(s.count)
}

// DoubleSmoothedStochastic is William Blau's Double Smoothed Stochastic (DSS) indicator.
//
// A classic double-smoothed stochastic oscillator bounded to [0, 100], paired
// with a short simple-moving-average signal line:
//
//	dss_k    = 100 * EMA(EMA(st, r), s) / EMA(EMA(rng, r), s)   (the oscillator)
//	signal_k = SMA(dss, g)_k                                    (g-period SMA)
//
// where, over the last q bars, HH_k is the highest high and LL_k is the lowest low,
// st_k = close_k - LL_k >= 0 is the raw stochastic (the close above the low), and
// rng_k = HH_k - LL_k >= 0 is the q-bar range.
//
// The raw stochastic and the range are smoothed separately with the same two-stage
// EMA cascade (r then s), then divided. Because 0 <= st <= rng on every bar, the
// ratio is bounded to [0, 100]. With q = 1 it is Blau's one-bar HLC index. It is
// exactly the MQL5 Blau_TStochI with its third EMA period u = 1. The inputs are the
// high, low and close prices.
//
// The indicator produces two outputs:
//   - DSS: the oscillator, range [0, 100];
//   - Signal: the g-period SMA of the oscillator.
//
// Priming convention (book / EasyLanguage): st and rng become valid once q bars of
// high/low exist, i.e. at bar q-1. All four cascade stages seed there together, so
// both outputs are NaN for bars 0..q-2 and finite from bar q-1; for q = 1 there is
// no NaN warm-up. The signal SMA seeds on the first finite oscillator value and
// returns the mean of the oscillator values seen so far (expanding window <= g),
// then the full g-bar rolling mean. Division guard: denominator <= 0 -> oscillator 0.0.
//
// Reference:
//
// Blau, William (1995). Momentum, Direction, and Divergence. Wiley.
type DoubleSmoothedStochastic struct {
	mu sync.RWMutex

	q           int
	highs       []float64
	lows        []float64
	windowCount int
	windowIndex int

	numR *ema
	numS *ema
	denR *ema
	denS *ema

	signalSMA *sma

	primed bool

	mnemonic string
}

// NewDoubleSmoothedStochastic returns an instance of the indicator created using supplied parameters.
func NewDoubleSmoothedStochastic(p *Params) (*DoubleSmoothedStochastic, error) {
	const (
		invalid  = "invalid double smoothed stochastic parameters"
		fmts     = "%s: %s"
		defaultQ = 5
		defaultR = 7
		defaultS = 3
		defaultG = 3
	)

	q := cmp.Or(p.Q, defaultQ)
	r := cmp.Or(p.R, defaultR)
	s := cmp.Or(p.S, defaultS)
	g := cmp.Or(p.G, defaultG)

	if q < 1 {
		return nil, fmt.Errorf(fmts, invalid, "q should be greater than 0")
	}

	if r < 1 {
		return nil, fmt.Errorf(fmts, invalid, "r should be greater than 0")
	}

	if s < 1 {
		return nil, fmt.Errorf(fmts, invalid, "s should be greater than 0")
	}

	if g < 1 {
		return nil, fmt.Errorf(fmts, invalid, "g should be greater than 0")
	}

	mnemonic := fmt.Sprintf("dss(%d,%d,%d,%d)", q, r, s, g)

	return &DoubleSmoothedStochastic{
		q:         q,
		highs:     make([]float64, q),
		lows:      make([]float64, q),
		numR:      newEMA(r),
		numS:      newEMA(s),
		denR:      newEMA(r),
		denS:      newEMA(s),
		signalSMA: newSMA(g),
		mnemonic:  mnemonic,
	}, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *DoubleSmoothedStochastic) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *DoubleSmoothedStochastic) Metadata() core.Metadata {
	desc := "Double Smoothed Stochastic " + s.mnemonic

	return core.BuildMetadata(
		core.DoubleSmoothedStochastic,
		s.mnemonic,
		desc,
		[]core.OutputText{
			{Mnemonic: s.mnemonic + " dss", Description: desc + " DSS"},
			{Mnemonic: s.mnemonic + " signal", Description: desc + " signal"},
		},
	)
}

// Update updates the indicator given the next bar's high, low and close values.
// Returns dss, signal values; both are NaN until q bars have been seen.
//
//nolint:nonamedreturns
func (s *DoubleSmoothedStochastic) Update(high, low, close float64) (dss, signal float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.highs[s.windowIndex] = high
	s.lows[s.windowIndex] = low
	s.windowIndex = (s.windowIndex + 1) % s.q

	if s.windowCount < s.q {
		s.windowCount++
	}

	// Need q bars of high/low before the stochastic is defined. Until then
	// neither output exists -- do NOT advance the EMA cascades or the SMA.
	if s.windowCount < s.q {
		nan := math.NaN()

		return nan, nan
	}

	// Rolling extremes over the last q bars.
	hh := s.highs[0]
	ll := s.lows[0]

	for i := 1; i < s.q; i++ {
		hh = max(hh, s.highs[i])
		ll = min(ll, s.lows[i])
	}

	// Raw stochastic and range (both non-negative).
	st := close - ll
	rng := hh - ll

	// Numerator cascade: EMA(EMA(st, r), s).
	n := s.numS.update(s.numR.update(st))
	// Denominator cascade: EMA(EMA(rng, r), s).
	d := s.denS.update(s.denR.update(rng))

	// Division guard: flat window so far -> oscillator 0.0.
	dss = 0.0
	if d > 0.0 {
		dss = 100.0 * n / d
	}

	// Signal line = SMA(dss, g); seeds on the first finite oscillator value.
	signal = s.signalSMA.update(dss)
	s.primed = true

	return dss, signal
}

// updateEntity updates the indicator and wraps the two outputs.
func (s *DoubleSmoothedStochastic) updateEntity(time time.Time, high, low, close float64) core.Output {
	dss, signal := s.Update(high, low, close)

	const outputCount = 2

	output := make([]any, outputCount)
	output[0] = entities.Scalar{Time: time, Value: dss}
	output[1] = entities.Scalar{Time: time, Value: signal}

	return output
}

// UpdateScalar updates the indicator given the next scalar sample.
//
// A scalar carries a single value, used as the high, the low and the close.
func (s *DoubleSmoothedStochastic) UpdateScalar(sample *entities.Scalar) core.Output {
	v := sample.Value

	return s.updateEntity(sample.Time, v, v, v)
}

// UpdateBar updates the indicator given the next bar sample.
func (s *DoubleSmoothedStochastic) UpdateBar(sample *entities.Bar) core.Output {
	return s.updateEntity(sample.Time, sample.High, sample.Low, sample.Close)
}

// UpdateQuote updates the indicator given the next quote sample.
//
// A quote maps the mid price to the high, the low and the close.
func (s *DoubleSmoothedStochastic) UpdateQuote(sample *entities.Quote) core.Output {
	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd

	return s.updateEntity(sample.Time, v, v, v)
}

// UpdateTrade updates the indicator given the next trade sample.
//
// A trade carries a single price, used as the high, the low and the close.
func (s *DoubleSmoothedStochastic) UpdateTrade(sample *entities.Trade) core.Output {
	v := sample.Price

	return s.updateEntity(sample.Time, v, v, v)
}
