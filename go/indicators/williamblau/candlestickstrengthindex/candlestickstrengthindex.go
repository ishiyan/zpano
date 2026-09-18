package candlestickstrengthindex

import (
	"fmt"
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

// CandlestickStrengthIndex is William Blau's Candlestick Strength Index (CSI) indicator,
// known in the book as the CandleStick Indicator.
//
// A double-/triple-smoothed candle-body-vs-range oscillator bounded to [-100, +100],
// paired with an EMA signal line (the Ergodic form, Blau ch.6.4):
//
//	csi_k    = 100 * TEMA(close-open, r, s, u) / TEMA(high-low, r, s, u)   (the oscillator)
//	signal_k = EMA(csi, ul)_k                                              (ul-period EMA)
//
// where the two intra-bar quantities are the signed candle body co_k = close_k - open_k
// and the bar range hl_k = high_k - low_k >= 0, and TEMA(x, r, s, u) = EMA(EMA(EMA(x, r), s), u).
//
// It is the range-normalized sibling of the Candlestick Momentum Index (CMI): both share
// the signed numerator TEMA(close-open), but the CSI divides by the smoothed range while
// the CMI divides by the smoothed absolute body. Because every bar has |close-open| <= high-low,
// the ratio is bounded to [-100, +100]: +100 when closes pin the high while opens pin the low
// (relentless bullish bodies), -100 in the mirror-image bearish case, and 0 when bodies net out.
// The inputs are the open, high, low and close prices.
//
// The indicator produces two outputs:
//   - CandlestickStrengthIndex: the oscillator, range [-100, +100];
//   - Signal: the ul-period EMA of the oscillator (Blau's Ergodic signal line).
//
// Priming convention (book / EasyLanguage): each EMA stage seeds on its first
// received value. Both intra-bar series are defined from bar 0, so there is no NaN
// warm-up region -- all stages seed on bar 0 and both outputs are finite for
// every bar. Division guard: denominator <= 0 -> oscillator 0.0.
//
// Reference:
//
// Blau, William (1995). Momentum, Direction, and Divergence, ch. 6 and Appendix B
// (Figure B-15). Wiley.
type CandlestickStrengthIndex struct {
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

// NewCandlestickStrengthIndex returns an instance of the indicator created using supplied parameters.
func NewCandlestickStrengthIndex(p *Params) (*CandlestickStrengthIndex, error) {
	const (
		invalid   = "invalid candlestick strength index parameters"
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

	mnemonic := fmt.Sprintf("csi(%d,%d,%d,%d)", r, s, u, ul)

	return &CandlestickStrengthIndex{
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
func (s *CandlestickStrengthIndex) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *CandlestickStrengthIndex) Metadata() core.Metadata {
	desc := "Candlestick Strength Index " + s.mnemonic

	return core.BuildMetadata(
		core.CandlestickStrengthIndex,
		s.mnemonic,
		desc,
		[]core.OutputText{
			{Mnemonic: s.mnemonic + " csi", Description: desc + " CSI"},
			{Mnemonic: s.mnemonic + " signal", Description: desc + " signal"},
		},
	)
}

// Update updates the indicator given the next bar's open, high, low and close values.
// Returns csi, signal values.
//
//nolint:nonamedreturns
func (s *CandlestickStrengthIndex) Update(open, high, low, close float64) (csi, signal float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Two intra-bar quantities: the signed candle body and the (non-negative) range.
	co := close - open
	hl := high - low

	// Numerator cascade: TEMA(close-open, r, s, u).
	n := s.numU.update(s.numS.update(s.numR.update(co)))
	// Denominator cascade: TEMA(high-low, r, s, u).
	d := s.denU.update(s.denS.update(s.denR.update(hl)))

	// Division guard: zero range so far -> oscillator 0.0.
	csi = 0.0
	if d > 0.0 {
		csi = 100.0 * n / d
	}

	// Signal line = EMA(csi, ul); seeds on bar 0's oscillator value.
	signal = s.signalEMA.update(csi)
	s.primed = true

	return csi, signal
}

// updateEntity updates the indicator and wraps the two outputs.
func (s *CandlestickStrengthIndex) updateEntity(
	time time.Time, open, high, low, close float64,
) core.Output {
	csi, signal := s.Update(open, high, low, close)

	const outputCount = 2

	output := make([]any, outputCount)
	output[0] = entities.Scalar{Time: time, Value: csi}
	output[1] = entities.Scalar{Time: time, Value: signal}

	return output
}

// UpdateScalar updates the indicator given the next scalar sample.
//
// A scalar carries a single value, so the candle body and the bar range are both zero.
func (s *CandlestickStrengthIndex) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.updateEntity(sample.Time, sample.Value, sample.Value, sample.Value, sample.Value)
}

// UpdateBar updates the indicator given the next bar sample.
func (s *CandlestickStrengthIndex) UpdateBar(sample *entities.Bar) core.Output {
	return s.updateEntity(sample.Time, sample.Open, sample.High, sample.Low, sample.Close)
}

// UpdateQuote updates the indicator given the next quote sample.
//
// A quote maps the bid to the open and the low, and the ask to the close and the high,
// so the candle body and the bar range are both the spread.
func (s *CandlestickStrengthIndex) UpdateQuote(sample *entities.Quote) core.Output {
	return s.updateEntity(sample.Time, sample.Bid, sample.Ask, sample.Bid, sample.Ask)
}

// UpdateTrade updates the indicator given the next trade sample.
//
// A trade carries a single price, so the candle body and the bar range are both zero.
func (s *CandlestickStrengthIndex) UpdateTrade(sample *entities.Trade) core.Output {
	return s.updateEntity(sample.Time, sample.Price, sample.Price, sample.Price, sample.Price)
}
