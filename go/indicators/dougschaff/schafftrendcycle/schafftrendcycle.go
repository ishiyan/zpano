package schafftrendcycle

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// ema is a stateful streaming exponential moving average: alpha = 2/(period+1),
// seeds e0 = x0. Inlined verbatim from the Blau exponential moving average so the
// indicator is a standalone porting unit. Do NOT change its numerics.
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

// window is a fixed-capacity ring buffer of the last n values, providing min/max.
type window struct {
	data  []float64
	size  int
	pos   int
	count int
}

func newWindow(n int) *window {
	return &window{data: make([]float64, n), size: n}
}

func (w *window) push(v float64) {
	w.data[w.pos] = v
	w.pos = (w.pos + 1) % w.size

	if w.count < w.size {
		w.count++
	}
}

func (w *window) minMax() (minVal, maxVal float64) {
	minVal = w.data[0]
	maxVal = w.data[0]

	for i := 1; i < w.count; i++ {
		v := w.data[i]
		if v < minVal {
			minVal = v
		}

		if v > maxVal {
			maxVal = v
		}
	}

	return minVal, maxVal
}

// SchaffTrendCycle is Doug Schaff's Schaff Trend Cycle (STC) indicator.
//
// STC runs a MACD line through two cascaded stochastics, each followed by an
// EMA-style smoothing, producing a cyclical oscillator bounded to [0, 100].
//
// The indicator produces three outputs:
//   - STC: the oscillator, range [0, 100], NaN during warm-up (bars 0..slow);
//   - MACD: the gated MACD line XMAC (0.0 pre-gate), exposed for stage testing;
//   - PF: the first smoothed %D (0.0 pre-gate), exposed for stage testing.
//
// Reference:
//
// Malagrida, F. (2017). Schaff Trend Cycle (schaff-trend-cycle2), ProRealCode.
type SchaffTrendCycle struct {
	mu sync.RWMutex

	emaFast *ema
	emaSlow *ema

	slow   int
	tclen  int
	factor float64

	bar int

	macdWin *window
	pfWin   *window

	frac1 float64
	frac2 float64
	pf    float64
	pff   float64

	primed bool

	barFunc   entities.BarFunc
	quoteFunc entities.QuoteFunc
	tradeFunc entities.TradeFunc

	mnemonic string
}

// NewSchaffTrendCycle returns an instance of the indicator created using supplied parameters.
//
//nolint:funlen,cyclop
func NewSchaffTrendCycle(p *Params) (*SchaffTrendCycle, error) {
	const (
		invalid       = "invalid schaff trend cycle parameters"
		fmts          = "%s: %s"
		fmtw          = "%s: %w"
		defaultFast   = 23
		defaultSlow   = 50
		defaultTclen  = 10
		defaultFactor = 0.5
	)

	fast := p.Fast
	if fast == 0 {
		fast = defaultFast
	}

	slow := p.Slow
	if slow == 0 {
		slow = defaultSlow
	}

	tclen := p.Tclen
	if tclen == 0 {
		tclen = defaultTclen
	}

	factor := p.Factor
	if factor == 0 {
		factor = defaultFactor
	}

	if fast < 1 {
		return nil, fmt.Errorf(fmts, invalid, "fast should be greater than 0")
	}

	if slow < 1 {
		return nil, fmt.Errorf(fmts, invalid, "slow should be greater than 0")
	}

	if tclen < 1 {
		return nil, fmt.Errorf(fmts, invalid, "tclen should be greater than 0")
	}

	if factor <= 0.0 || factor > 1.0 {
		return nil, fmt.Errorf(fmts, invalid, "factor should be in (0, 1]")
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

	mnemonic := fmt.Sprintf("stc(%d,%d,%d,%.2f%s)", fast, slow, tclen, factor,
		core.ComponentTripleMnemonic(bc, qc, tc))

	return &SchaffTrendCycle{
		emaFast:   newEMA(fast),
		emaSlow:   newEMA(slow),
		slow:      slow,
		tclen:     tclen,
		factor:    factor,
		bar:       -1,
		macdWin:   newWindow(tclen),
		pfWin:     newWindow(tclen),
		barFunc:   barFunc,
		quoteFunc: quoteFunc,
		tradeFunc: tradeFunc,
		mnemonic:  mnemonic,
	}, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *SchaffTrendCycle) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *SchaffTrendCycle) Metadata() core.Metadata {
	desc := "Schaff Trend Cycle " + s.mnemonic

	return core.BuildMetadata(
		core.SchaffTrendCycle,
		s.mnemonic,
		desc,
		[]core.OutputText{
			{Mnemonic: s.mnemonic + " stc", Description: desc + " STC"},
			{Mnemonic: s.mnemonic + " macd", Description: desc + " MACD"},
			{Mnemonic: s.mnemonic + " pf", Description: desc + " PF"},
		},
	)
}

// Update updates the indicator given the next sample value.
// Returns stc, macd, pf values.
//
//nolint:nonamedreturns
func (s *SchaffTrendCycle) Update(sample float64) (stc, macd, pf float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.bar++
	k := s.bar

	// Price EMAs always advance (they accumulate over the full history).
	emaFast := s.emaFast.update(sample)
	emaSlow := s.emaSlow.update(sample)

	// GATE: XMAC is only assigned while barindex > slow.
	gateOpen := k > s.slow

	macd = 0.0
	if gateOpen {
		macd = emaFast - emaSlow
	}

	s.macdWin.push(macd)

	if !gateOpen {
		s.pfWin.push(s.pf)

		return math.NaN(), macd, s.pf
	}

	// 1st stochastic of the MACD over tclen (guard on the range).
	ll1, hh1 := s.macdWin.minMax()
	rng1 := hh1 - ll1

	if rng1 > 0.0 {
		s.frac1 = ((macd - ll1) / rng1) * 100.0
	}

	// 1st smoothing: PF = EMA(Frac1, alpha=factor), seed 0.
	s.pf += s.factor * (s.frac1 - s.pf)
	s.pfWin.push(s.pf)

	// 2nd stochastic of PF over tclen.
	ll2, hh2 := s.pfWin.minMax()
	rng2 := hh2 - ll2

	if rng2 > 0.0 {
		s.frac2 = ((s.pf - ll2) / rng2) * 100.0
	}

	// 2nd smoothing: STC = PFF = EMA(Frac2, alpha=factor), seed 0.
	s.pff += s.factor * (s.frac2 - s.pff)
	s.primed = true

	return s.pff, macd, s.pf
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *SchaffTrendCycle) UpdateScalar(sample *entities.Scalar) core.Output {
	stc, macd, pf := s.Update(sample.Value)

	const outputCount = 3

	output := make([]any, outputCount)
	output[0] = entities.Scalar{Time: sample.Time, Value: stc}
	output[1] = entities.Scalar{Time: sample.Time, Value: macd}
	output[2] = entities.Scalar{Time: sample.Time, Value: pf}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (s *SchaffTrendCycle) UpdateBar(sample *entities.Bar) core.Output {
	v := s.barFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *SchaffTrendCycle) UpdateQuote(sample *entities.Quote) core.Output {
	v := s.quoteFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *SchaffTrendCycle) UpdateTrade(sample *entities.Trade) core.Output {
	v := s.tradeFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}
