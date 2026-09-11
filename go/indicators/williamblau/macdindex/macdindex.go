package macdindex

import (
	"fmt"
	"sync"

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

// MacdIndex is William Blau's MACD Index (MACD_I) indicator.
//
// Blau's MACD line is the difference of two EMAs of the close, optionally
// smoothed by a third EMA, paired with an EMA signal line (the Ergodic form,
// Blau ch. 5):
//
//	macd_k   = EMA(close, s)_k - EMA(close, r)_k          (MACD line; s fast, r slow)
//	macdi_k  = EMA(macd, u)_k                             (the MACD_I line)
//	signal_k = EMA(macdi, ul)_k                           (ul-period EMA)
//
// with the fast period s strictly shorter than the slow period r (s < r).
// Setting u=1 recovers the book's pure two-EMA MACD line. Blau notes the MACD
// and the MDI are both double-smoothed momentum indicators with nearly
// interchangeable shapes (within a scale factor).
//
// The index is NOT normalized: there is no 100 * TEMA/TEMA ratio and no fixed
// range, so the output is in the same price units as the input and may take any
// sign or magnitude. Because there is no division there is also no division guard.
//
// The indicator produces two outputs:
//   - MACDI: the index line, in raw price units, finite from bar 0;
//   - Signal: the ul-period EMA of the index (Blau's Ergodic signal line).
//
// Priming convention (book / EasyLanguage): each EMA stage seeds on its first
// received value. Both price EMAs are defined from bar 0, so macd_0 = 0 and the
// u smoothing and signal EMAs seed on that 0 -- there is no NaN warm-up region
// and bar 0 is exactly 0.0.
//
// Reference:
//
// Blau, William (1995). Momentum, Direction, and Divergence, ch. 5. Wiley.
type MacdIndex struct {
	mu sync.RWMutex

	emaFast  *ema
	emaSlow  *ema
	smoothU  *ema
	signalMA *ema

	primed bool

	barFunc   entities.BarFunc
	quoteFunc entities.QuoteFunc
	tradeFunc entities.TradeFunc

	mnemonic string
}

// NewMacdIndex returns an instance of the indicator created using supplied parameters.
//
//nolint:funlen,cyclop
func NewMacdIndex(p *Params) (*MacdIndex, error) {
	const (
		invalid   = "invalid macd index parameters"
		fmts      = "%s: %s"
		fmtw      = "%s: %w"
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

	if s >= r {
		return nil, fmt.Errorf(fmts, invalid, "s (fast) should be less than r (slow)")
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

	mnemonic := fmt.Sprintf("macdi(%d,%d,%d%s)", r, s, u,
		core.ComponentTripleMnemonic(bc, qc, tc))

	return &MacdIndex{
		emaFast:   newEMA(s),
		emaSlow:   newEMA(r),
		smoothU:   newEMA(u),
		signalMA:  newEMA(ul),
		barFunc:   barFunc,
		quoteFunc: quoteFunc,
		tradeFunc: tradeFunc,
		mnemonic:  mnemonic,
	}, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *MacdIndex) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *MacdIndex) Metadata() core.Metadata {
	desc := "MACD Index " + s.mnemonic

	return core.BuildMetadata(
		core.MacdIndex,
		s.mnemonic,
		desc,
		[]core.OutputText{
			{Mnemonic: s.mnemonic + " macdi", Description: desc + " MACDI"},
			{Mnemonic: s.mnemonic + " signal", Description: desc + " signal"},
		},
	)
}

// Update updates the indicator given the next sample value.
// Returns macdi, signal values.
//
//nolint:nonamedreturns
func (s *MacdIndex) Update(sample float64) (macdi, signal float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// MACD line = fast EMA - slow EMA. Both seed at bar 0 (to close_0), so the
	// line is 0.0 on bar 0 and defined on every bar thereafter.
	macd := s.emaFast.update(sample) - s.emaSlow.update(sample)

	// Smooth the MACD line: EMA(macd, u). No normalization, no guard.
	macdi = s.smoothU.update(macd)

	// Signal line = EMA(macdi, ul); seeds here on the bar-0 index value.
	signal = s.signalMA.update(macdi)
	s.primed = true

	return macdi, signal
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *MacdIndex) UpdateScalar(sample *entities.Scalar) core.Output {
	macdi, signal := s.Update(sample.Value)

	const outputCount = 2

	output := make([]any, outputCount)
	output[0] = entities.Scalar{Time: sample.Time, Value: macdi}
	output[1] = entities.Scalar{Time: sample.Time, Value: signal}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (s *MacdIndex) UpdateBar(sample *entities.Bar) core.Output {
	v := s.barFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *MacdIndex) UpdateQuote(sample *entities.Quote) core.Output {
	v := s.quoteFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *MacdIndex) UpdateTrade(sample *entities.Trade) core.Output {
	v := s.tradeFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}
