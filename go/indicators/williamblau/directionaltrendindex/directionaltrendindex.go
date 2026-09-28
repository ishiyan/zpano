package directionaltrendindex

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

// DirectionalTrendIndex is William Blau's Directional Trend Index (DTI) indicator.
//
// A double-/triple-smoothed High-Low Momentum oscillator bounded to [-100, +100],
// paired with an EMA signal line (the Ergodic form):
//
//	dti_k    = 100 * TEMA(HLM, r, s, u)_k / TEMA(|HLM|, r, s, u)_k   (the oscillator)
//	signal_k = EMA(dti, ul)_k                                       (ul-period EMA)
//
// where the High-Low Momentum is built from how far the high rose and the low
// fell relative to q-1 bars ago:
//
//	HMU_k = max(high_k - high_(k-(q-1)), 0)      (upward high movement)
//	LMD_k = max(low_(k-(q-1)) - low_k, 0)        (downward low movement)
//	HLM_k = HMU_k - LMD_k                         (composite high-low momentum)
//	TEMA(x, r, s, u) = EMA(EMA(EMA(x, r), s), u)  (triple EMA cascade)
//
// This is the True Strength Index structure applied to HLM instead of price
// momentum. The inputs are the high and low prices only (no close).
//
// The indicator produces two outputs:
//   - DTI: the oscillator, range [-100, +100];
//   - Signal: the ul-period EMA of the oscillator.
//
// Priming convention (book / EasyLanguage): HLM is valid from bar q-1 (it needs a
// high/low from q-1 bars ago), so all cascade stages seed there together; both
// outputs are NaN for bars 0..q-2 and finite from bar q-1. For q = 1 there is no
// NaN warm-up, but HLM == 0 on every bar, so the division guard yields dti == 0.0
// for all bars. The signal EMA seeds on the first finite oscillator value.
// Division guard: denominator == 0 -> oscillator 0.0.
//
// Reference:
//
// Blau, William (1995). Momentum, Direction, and Divergence. Wiley.
type DirectionalTrendIndex struct {
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

// NewDirectionalTrendIndex returns an instance of the indicator created using supplied parameters.
func NewDirectionalTrendIndex(p *Params) (*DirectionalTrendIndex, error) {
	const (
		invalid   = "invalid directional trend index parameters"
		fmts      = "%s: %s"
		defaultQ  = 2
		defaultR  = 20
		defaultS  = 5
		defaultU  = 3
		defaultUl = 3
	)

	q := cmp.Or(p.Q, defaultQ)
	r := cmp.Or(p.R, defaultR)
	s := cmp.Or(p.S, defaultS)
	u := cmp.Or(p.U, defaultU)
	ul := cmp.Or(p.Ul, defaultUl)

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

	mnemonic := fmt.Sprintf("dti(%d,%d,%d,%d,%d)", q, r, s, u, ul)

	return &DirectionalTrendIndex{
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
func (s *DirectionalTrendIndex) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *DirectionalTrendIndex) Metadata() core.Metadata {
	desc := "Directional Trend Index " + s.mnemonic

	return core.BuildMetadata(
		core.DirectionalTrendIndex,
		s.mnemonic,
		desc,
		[]core.OutputText{
			{Mnemonic: s.mnemonic + " dti", Description: desc + " DTI"},
			{Mnemonic: s.mnemonic + " signal", Description: desc + " signal"},
		},
	)
}

// Update updates the indicator given the next bar's high and low values.
// Returns dti, signal values; both are NaN until q bars have been seen.
//
//nolint:nonamedreturns
func (s *DirectionalTrendIndex) Update(high, low float64) (dti, signal float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.highs[s.windowIndex] = high
	s.lows[s.windowIndex] = low
	s.windowIndex = (s.windowIndex + 1) % s.q

	if s.windowCount < s.q {
		s.windowCount++
	}

	// HLM needs a high/low from q-1 bars ago. Until then neither output
	// exists -- do NOT advance the EMA cascades.
	if s.windowCount < s.q {
		nan := math.NaN()

		return nan, nan
	}

	// The oldest value in the full window sits at the next write position:
	// high_(k-(q-1)) and low_(k-(q-1)).
	previousHigh := s.highs[s.windowIndex]
	previousLow := s.lows[s.windowIndex]

	// Upward high movement and downward low movement, each floored at 0.
	hmu := max(high-previousHigh, 0.0)
	lmd := max(previousLow-low, 0.0)

	// Composite high-low momentum and its magnitude.
	hlm := hmu - lmd
	absHlm := math.Abs(hlm)

	// Numerator cascade: TEMA(HLM, r, s, u).
	n := s.numU.update(s.numS.update(s.numR.update(hlm)))
	// Denominator cascade: TEMA(|HLM|, r, s, u).
	d := s.denU.update(s.denS.update(s.denR.update(absHlm)))

	// Division guard: denominator 0 -> oscillator 0.0.
	dti = 0.0
	if d != 0.0 {
		dti = 100.0 * n / d
	}

	// Signal line = EMA(dti, ul); seeds on the first finite oscillator value.
	signal = s.signalEMA.update(dti)
	s.primed = true

	return dti, signal
}

// updateEntity updates the indicator and wraps the two outputs.
func (s *DirectionalTrendIndex) updateEntity(time time.Time, high, low float64) core.Output {
	dti, signal := s.Update(high, low)

	const outputCount = 2

	output := make([]any, outputCount)
	output[0] = entities.Scalar{Time: time, Value: dti}
	output[1] = entities.Scalar{Time: time, Value: signal}

	return output
}

// UpdateScalar updates the indicator given the next scalar sample.
//
// A scalar carries a single value, used as both the high and the low.
func (s *DirectionalTrendIndex) UpdateScalar(sample *entities.Scalar) core.Output {
	v := sample.Value

	return s.updateEntity(sample.Time, v, v)
}

// UpdateBar updates the indicator given the next bar sample.
func (s *DirectionalTrendIndex) UpdateBar(sample *entities.Bar) core.Output {
	return s.updateEntity(sample.Time, sample.High, sample.Low)
}

// UpdateQuote updates the indicator given the next quote sample.
//
// A quote maps the mid price to both the high and the low.
func (s *DirectionalTrendIndex) UpdateQuote(sample *entities.Quote) core.Output {
	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd

	return s.updateEntity(sample.Time, v, v)
}

// UpdateTrade updates the indicator given the next trade sample.
//
// A trade carries a single price, used as both the high and the low.
func (s *DirectionalTrendIndex) UpdateTrade(sample *entities.Trade) core.Output {
	v := sample.Price

	return s.updateEntity(sample.Time, v, v)
}
