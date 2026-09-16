package candlestickmomentumindex

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

// CandlestickMomentumIndex is William Blau's Candlestick Momentum Index (CMI) indicator.
//
// A double-/triple-smoothed intra-bar momentum oscillator bounded to [-100, +100],
// paired with an EMA signal line (the Ergodic form, Blau ch.6.4):
//
//	cmi_k    = 100 * TEMA(cmtm, r, s, u) / TEMA(|cmtm|, r, s, u)   (the oscillator)
//	signal_k = EMA(cmi, ul)_k                                      (ul-period EMA)
//
// where the candle momentum is the signed candle body cmtm_k = close_k - open_k
// and TEMA(x, r, s, u) = EMA(EMA(EMA(x, r), s), u).
//
// This is the True Strength Index structure applied to the candle body instead of
// price momentum. Because it only looks inside each bar it is immune to inter-bar
// gaps. The inputs are the open and close prices only.
//
// The indicator produces two outputs:
//   - CandlestickMomentumIndex: the oscillator, range [-100, +100];
//   - Signal: the ul-period EMA of the oscillator (Blau's Ergodic signal line).
//
// Priming convention (book / EasyLanguage): each EMA stage seeds on its first
// received value. The candle momentum is defined from bar 0, so there is no NaN
// warm-up region -- all stages seed on bar 0 and both outputs are finite for
// every bar. Division guard: denominator 0 -> oscillator 0.0.
//
// Reference:
//
// Blau, William (1995). Momentum, Direction, and Divergence, ch. 6. Wiley.
type CandlestickMomentumIndex struct {
	mu sync.RWMutex

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

// NewCandlestickMomentumIndex returns an instance of the indicator created using supplied parameters.
func NewCandlestickMomentumIndex(p *Params) (*CandlestickMomentumIndex, error) {
	const (
		invalid   = "invalid candlestick momentum index parameters"
		fmts      = "%s: %s"
		defaultR  = 20
		defaultS  = 5
		defaultU  = 3
		defaultUl = 3
	)

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

	mnemonic := fmt.Sprintf("cmi(%d,%d,%d,%d)", r, s, u, ul)

	return &CandlestickMomentumIndex{
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
func (s *CandlestickMomentumIndex) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *CandlestickMomentumIndex) Metadata() core.Metadata {
	desc := "Candlestick Momentum Index " + s.mnemonic

	return core.BuildMetadata(
		core.CandlestickMomentumIndex,
		s.mnemonic,
		desc,
		[]core.OutputText{
			{Mnemonic: s.mnemonic + " cmi", Description: desc + " CMI"},
			{Mnemonic: s.mnemonic + " signal", Description: desc + " signal"},
		},
	)
}

// Update updates the indicator given the next bar's open and close values.
// Returns cmi, signal values.
//
//nolint:nonamedreturns
func (s *CandlestickMomentumIndex) Update(open, close float64) (cmi, signal float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Candle momentum: the signed body of the candle.
	cmtm := close - open
	absCmtm := math.Abs(cmtm)

	// Numerator cascade: TEMA(cmtm, r, s, u).
	n := s.numU.update(s.numS.update(s.numR.update(cmtm)))
	// Denominator cascade: TEMA(|cmtm|, r, s, u).
	d := s.denU.update(s.denS.update(s.denR.update(absCmtm)))

	// Division guard (Blau_CMI.mq5): denominator 0 -> oscillator 0.0.
	cmi = 0.0
	if d != 0.0 {
		cmi = 100.0 * n / d
	}

	// Signal line = EMA(cmi, ul); seeds on bar 0's oscillator value.
	signal = s.signalEMA.update(cmi)
	s.primed = true

	return cmi, signal
}

// updateEntity updates the indicator and wraps the two outputs.
func (s *CandlestickMomentumIndex) updateEntity(time time.Time, open, close float64) core.Output {
	cmi, signal := s.Update(open, close)

	const outputCount = 2

	output := make([]any, outputCount)
	output[0] = entities.Scalar{Time: time, Value: cmi}
	output[1] = entities.Scalar{Time: time, Value: signal}

	return output
}

// UpdateScalar updates the indicator given the next scalar sample.
//
// A scalar carries a single value, so open == close and the candle momentum is zero.
func (s *CandlestickMomentumIndex) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.updateEntity(sample.Time, sample.Value, sample.Value)
}

// UpdateBar updates the indicator given the next bar sample.
func (s *CandlestickMomentumIndex) UpdateBar(sample *entities.Bar) core.Output {
	return s.updateEntity(sample.Time, sample.Open, sample.Close)
}

// UpdateQuote updates the indicator given the next quote sample.
//
// A quote maps the bid to the open and the ask to the close, so the candle body is the spread.
func (s *CandlestickMomentumIndex) UpdateQuote(sample *entities.Quote) core.Output {
	return s.updateEntity(sample.Time, sample.Bid, sample.Ask)
}

// UpdateTrade updates the indicator given the next trade sample.
//
// A trade carries a single price, so open == close and the candle momentum is zero.
func (s *CandlestickMomentumIndex) UpdateTrade(sample *entities.Trade) core.Output {
	return s.updateEntity(sample.Time, sample.Price, sample.Price)
}
