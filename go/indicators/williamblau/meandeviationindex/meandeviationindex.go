package meandeviationindex

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

// MeanDeviationIndex is William Blau's Mean Deviation Index (MDI) indicator.
//
// A detrended, double-/triple-smoothed momentum line in raw price units, paired
// with an EMA signal line (the Ergodic form, Blau ch. 5):
//
//	md_k     = price_k - EMA(price, r)_k                  (deviation from trend)
//	mdi_k    = EMA(EMA(md, s), u)_k                       (the MDI line)
//	signal_k = EMA(mdi, ul)_k                             (ul-period EMA)
//
// The price series is detrended by subtracting its own r-period EMA, then the
// deviation is smoothed by an s-period EMA and an optional u-period EMA. Blau
// notes the MDI approximates the MACD when r is long and s is short.
//
// The index is NOT normalized: there is no 100 * TEMA/TEMA ratio and no fixed
// range, so the output is in the same price units as the input and may take any
// sign or magnitude. Because there is no division there is also no division guard.
//
// The indicator produces two outputs:
//   - MDI: the index line, in raw price units, finite from bar 0;
//   - Signal: the ul-period EMA of the index (Blau's Ergodic signal line).
//
// Priming convention (book / EasyLanguage): each EMA stage seeds on its first
// received value. The detrending EMA is defined from bar 0, so md_0 = 0 and both
// smoothing EMAs seed on that 0 -- there is no NaN warm-up region and bar 0 is
// exactly 0.0. Degenerate case: r=1 makes the detrend a passthrough, so the
// index is identically 0.0.
//
// Reference:
//
// Blau, William (1995). Momentum, Direction, and Divergence, ch. 5. Wiley.
type MeanDeviationIndex struct {
	mu sync.RWMutex

	trend    *ema
	smoothS  *ema
	smoothU  *ema
	signalMA *ema

	primed bool

	barFunc   entities.BarFunc
	quoteFunc entities.QuoteFunc
	tradeFunc entities.TradeFunc

	mnemonic string
}

// NewMeanDeviationIndex returns an instance of the indicator created using supplied parameters.
//
//nolint:funlen,cyclop
func NewMeanDeviationIndex(p *Params) (*MeanDeviationIndex, error) {
	const (
		invalid   = "invalid mean deviation index parameters"
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

	mnemonic := fmt.Sprintf("mdi(%d,%d,%d%s)", r, s, u,
		core.ComponentTripleMnemonic(bc, qc, tc))

	return &MeanDeviationIndex{
		trend:     newEMA(r),
		smoothS:   newEMA(s),
		smoothU:   newEMA(u),
		signalMA:  newEMA(ul),
		barFunc:   barFunc,
		quoteFunc: quoteFunc,
		tradeFunc: tradeFunc,
		mnemonic:  mnemonic,
	}, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *MeanDeviationIndex) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *MeanDeviationIndex) Metadata() core.Metadata {
	desc := "Mean Deviation Index " + s.mnemonic

	return core.BuildMetadata(
		core.MeanDeviationIndex,
		s.mnemonic,
		desc,
		[]core.OutputText{
			{Mnemonic: s.mnemonic + " mdi", Description: desc + " MDI"},
			{Mnemonic: s.mnemonic + " signal", Description: desc + " signal"},
		},
	)
}

// Update updates the indicator given the next sample value.
// Returns mdi, signal values.
//
//nolint:nonamedreturns
func (s *MeanDeviationIndex) Update(sample float64) (mdi, signal float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Mean deviation: the price minus its own r-period EMA trend. The baseline
	// EMA seeds on bar 0, so the bar-0 deviation is exactly 0.
	md := sample - s.trend.update(sample)

	// Smooth the deviation: EMA(EMA(md, s), u). No normalization, no guard.
	mdi = s.smoothU.update(s.smoothS.update(md))

	// Signal line = EMA(mdi, ul); seeds here on the bar-0 index value.
	signal = s.signalMA.update(mdi)
	s.primed = true

	return mdi, signal
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *MeanDeviationIndex) UpdateScalar(sample *entities.Scalar) core.Output {
	mdi, signal := s.Update(sample.Value)

	const outputCount = 2

	output := make([]any, outputCount)
	output[0] = entities.Scalar{Time: sample.Time, Value: mdi}
	output[1] = entities.Scalar{Time: sample.Time, Value: signal}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (s *MeanDeviationIndex) UpdateBar(sample *entities.Bar) core.Output {
	v := s.barFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *MeanDeviationIndex) UpdateQuote(sample *entities.Quote) core.Output {
	v := s.quoteFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *MeanDeviationIndex) UpdateTrade(sample *entities.Trade) core.Output {
	v := s.tradeFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}
