package stochasticmomentumindex

import (
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
	alpha  float64
	prev   float64
	primed bool
}

func newEMA(period int) *ema {
	return &ema{alpha: 2.0 / (float64(period) + 1.0)}
}

func (e *ema) update(x float64) float64 {
	if !e.primed {
		e.prev = x
		e.primed = true

		return e.prev
	}

	e.prev = e.alpha*x + (1.0-e.alpha)*e.prev

	return e.prev
}

// StochasticMomentumIndex is William Blau's Stochastic Momentum Index (SMI) indicator.
//
// A double-/triple-smoothed stochastic oscillator bounded to [-100, +100],
// paired with an EMA signal line (the Ergodic form, Blau ch.3.4):
//
//	smi_k    = 100 * TEMA(sm, r, s, u) / TEMA(hr, r, s, u)   (the oscillator)
//	signal_k = EMA(smi, ul)_k                                (ul-period EMA)
//
// where, over the last q bars, HH_k is the highest high and LL_k is the lowest low,
// sm_k = close_k - 0.5*(HH_k + LL_k) is the distance of the close from the range
// midpoint, hr_k = 0.5*(HH_k - LL_k) >= 0 is the half-range, and
// TEMA(x, r, s, u) = EMA(EMA(EMA(x, r), s), u).
//
// Where the ordinary stochastic measures where the close sits inside the recent
// high-low range, the SMI measures the close relative to the midpoint of that range.
// Because |sm| <= hr on every bar, the ratio is bounded to [-100, +100]. With q = 1 it
// is Blau's one-day stochastic (sentiment indicator). The inputs are the high, low and
// close prices.
//
// The indicator produces two outputs:
//   - SMI: the oscillator, range [-100, +100];
//   - Signal: the ul-period EMA of the oscillator (Blau's Ergodic signal line).
//
// Priming convention (book / EasyLanguage): sm and hr become valid once q bars of
// high/low exist, i.e. at bar q-1. All six cascade stages seed there together, so
// both outputs are NaN for bars 0..q-2 and finite from bar q-1; for q = 1 there is
// no NaN warm-up. Division guard: denominator <= 0 -> oscillator 0.0.
//
// Reference:
//
// Blau, William (1995). Momentum, Direction, and Divergence, ch. 3. Wiley.
type StochasticMomentumIndex struct {
	mu sync.RWMutex

	q           int
	highs       []float64
	lows        []float64
	windowCount int
	windowIndex int

	numR *ema
	numS *ema
	numU *ema
	denR *ema
	denS *ema
	denU *ema

	signalEMA *ema

	primed bool

	mnemonic string
}

// NewStochasticMomentumIndex returns an instance of the indicator created using supplied parameters.
func NewStochasticMomentumIndex(p *Params) (*StochasticMomentumIndex, error) {
	const (
		invalid   = "invalid stochastic momentum index parameters"
		fmts      = "%s: %s"
		defaultQ  = 5
		defaultR  = 20
		defaultS  = 5
		defaultU  = 3
		defaultUl = 3
	)

	q := p.Q
	if q == 0 {
		q = defaultQ
	}

	r := p.R
	if r == 0 {
		r = defaultR
	}

	s := p.S
	if s == 0 {
		s = defaultS
	}

	u := p.U
	if u == 0 {
		u = defaultU
	}

	ul := p.Ul
	if ul == 0 {
		ul = defaultUl
	}

	if q < 1 {
		return nil, fmt.Errorf(fmts, invalid, "q should be greater than 0")
	}

	if r < 1 {
		return nil, fmt.Errorf(fmts, invalid, "r should be greater than 0")
	}

	if s < 1 {
		return nil, fmt.Errorf(fmts, invalid, "s should be greater than 0")
	}

	if u < 1 {
		return nil, fmt.Errorf(fmts, invalid, "u should be greater than 0")
	}

	if ul < 1 {
		return nil, fmt.Errorf(fmts, invalid, "ul should be greater than 0")
	}

	mnemonic := fmt.Sprintf("smi(%d,%d,%d,%d,%d)", q, r, s, u, ul)

	return &StochasticMomentumIndex{
		q:         q,
		highs:     make([]float64, q),
		lows:      make([]float64, q),
		numR:      newEMA(r),
		numS:      newEMA(s),
		numU:      newEMA(u),
		denR:      newEMA(r),
		denS:      newEMA(s),
		denU:      newEMA(u),
		signalEMA: newEMA(ul),
		mnemonic:  mnemonic,
	}, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *StochasticMomentumIndex) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *StochasticMomentumIndex) Metadata() core.Metadata {
	desc := "Stochastic Momentum Index " + s.mnemonic

	return core.BuildMetadata(
		core.StochasticMomentumIndex,
		s.mnemonic,
		desc,
		[]core.OutputText{
			{Mnemonic: s.mnemonic + " smi", Description: desc + " SMI"},
			{Mnemonic: s.mnemonic + " signal", Description: desc + " signal"},
		},
	)
}

// Update updates the indicator given the next bar's high, low and close values.
// Returns smi, signal values; both are NaN until q bars have been seen.
//
//nolint:nonamedreturns
func (s *StochasticMomentumIndex) Update(high, low, close float64) (smi, signal float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.highs[s.windowIndex] = high
	s.lows[s.windowIndex] = low
	s.windowIndex = (s.windowIndex + 1) % s.q

	if s.windowCount < s.q {
		s.windowCount++
	}

	// Need q bars of high/low before the stochastic is defined. Until then
	// neither output exists -- do NOT advance the EMA cascades.
	if s.windowCount < s.q {
		nan := math.NaN()

		return nan, nan
	}

	// Rolling extremes over the last q bars.
	hh := s.highs[0]
	ll := s.lows[0]

	for i := 1; i < s.q; i++ {
		hh = math.Max(hh, s.highs[i])
		ll = math.Min(ll, s.lows[i])
	}

	// Stochastic momentum (signed) and half-range (non-negative).
	sm := close - 0.5*(hh+ll)
	hr := 0.5 * (hh - ll)

	// Numerator cascade: TEMA(sm, r, s, u).
	n := s.numU.update(s.numS.update(s.numR.update(sm)))
	// Denominator cascade: TEMA(hr, r, s, u).
	d := s.denU.update(s.denS.update(s.denR.update(hr)))

	// Division guard: flat window so far -> oscillator 0.0.
	smi = 0.0
	if d > 0.0 {
		smi = 100.0 * n / d
	}

	// Signal line = EMA(smi, ul); seeds on the first finite oscillator value.
	signal = s.signalEMA.update(smi)
	s.primed = true

	return smi, signal
}

// updateEntity updates the indicator and wraps the two outputs.
func (s *StochasticMomentumIndex) updateEntity(time time.Time, high, low, close float64) core.Output {
	smi, signal := s.Update(high, low, close)

	const outputCount = 2

	output := make([]any, outputCount)
	output[0] = entities.Scalar{Time: time, Value: smi}
	output[1] = entities.Scalar{Time: time, Value: signal}

	return output
}

// UpdateScalar updates the indicator given the next scalar sample.
//
// A scalar carries a single value, used as the high, the low and the close.
func (s *StochasticMomentumIndex) UpdateScalar(sample *entities.Scalar) core.Output {
	v := sample.Value

	return s.updateEntity(sample.Time, v, v, v)
}

// UpdateBar updates the indicator given the next bar sample.
func (s *StochasticMomentumIndex) UpdateBar(sample *entities.Bar) core.Output {
	return s.updateEntity(sample.Time, sample.High, sample.Low, sample.Close)
}

// UpdateQuote updates the indicator given the next quote sample.
//
// A quote maps the mid price to the high, the low and the close.
func (s *StochasticMomentumIndex) UpdateQuote(sample *entities.Quote) core.Output {
	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd

	return s.updateEntity(sample.Time, v, v, v)
}

// UpdateTrade updates the indicator given the next trade sample.
//
// A trade carries a single price, used as the high, the low and the close.
func (s *StochasticMomentumIndex) UpdateTrade(sample *entities.Trade) core.Output {
	v := sample.Price

	return s.updateEntity(sample.Time, v, v, v)
}
