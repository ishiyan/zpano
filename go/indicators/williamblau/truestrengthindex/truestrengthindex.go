package truestrengthindex

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

// TrueStrengthIndex is William Blau's True Strength Index (TSI) indicator.
//
// A double-/triple-smoothed momentum oscillator bounded to [-100, +100], paired
// with an EMA signal line (the Ergodic form, Blau ch.1.4):
//
//	tsi_k    = 100 * TEMA(mtm, r, s, u) / TEMA(|mtm|, r, s, u)   (the oscillator)
//	signal_k = EMA(tsi, ul)_k                                    (ul-period EMA)
//
// where mtm_k = C_k - C_(k-(q-1)) and TEMA(x, r, s, u) = EMA(EMA(EMA(x, r), s), u).
//
// The indicator produces two outputs:
//   - TSI: the oscillator, range [-100, +100], NaN during warm-up (bars 0..q-2);
//   - Signal: the ul-period EMA of the oscillator (Blau's Ergodic signal line).
//
// Priming convention (book / EasyLanguage): each EMA stage seeds on its first
// received value, so all stages seed at bar q-1 together; the signal EMA seeds
// on the first finite oscillator. Division guard: denominator 0 -> oscillator 0.0.
//
// Reference:
//
// Blau, William (1995). Momentum, Direction, and Divergence, ch. 2. Wiley.
type TrueStrengthIndex struct {
	mu sync.RWMutex

	q int

	history    []float64
	historyLen int

	numR *ema
	numS *ema
	numU *ema
	denR *ema
	denS *ema
	denU *ema

	signalEMA *ema

	primed bool

	barFunc   entities.BarFunc
	quoteFunc entities.QuoteFunc
	tradeFunc entities.TradeFunc

	mnemonic string
}

// NewTrueStrengthIndex returns an instance of the indicator created using supplied parameters.
//
//nolint:funlen,cyclop
func NewTrueStrengthIndex(p *Params) (*TrueStrengthIndex, error) {
	const (
		invalid   = "invalid true strength index parameters"
		fmts      = "%s: %s"
		fmtw      = "%s: %w"
		defaultQ  = 2
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

	mnemonic := fmt.Sprintf("tsi(%d,%d,%d,%d%s)", q, r, s, u,
		core.ComponentTripleMnemonic(bc, qc, tc))

	return &TrueStrengthIndex{
		q:         q,
		history:   make([]float64, 0, q),
		numR:      newEMA(r),
		numS:      newEMA(s),
		numU:      newEMA(u),
		denR:      newEMA(r),
		denS:      newEMA(s),
		denU:      newEMA(u),
		signalEMA: newEMA(ul),
		barFunc:   barFunc,
		quoteFunc: quoteFunc,
		tradeFunc: tradeFunc,
		mnemonic:  mnemonic,
	}, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *TrueStrengthIndex) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *TrueStrengthIndex) Metadata() core.Metadata {
	desc := "True Strength Index " + s.mnemonic

	return core.BuildMetadata(
		core.TrueStrengthIndex,
		s.mnemonic,
		desc,
		[]core.OutputText{
			{Mnemonic: s.mnemonic + " tsi", Description: desc + " TSI"},
			{Mnemonic: s.mnemonic + " signal", Description: desc + " signal"},
		},
	)
}

// Update updates the indicator given the next sample value.
// Returns tsi, signal values.
//
//nolint:nonamedreturns
func (s *TrueStrengthIndex) Update(sample float64) (tsi, signal float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Maintain a rolling window of the last q prices; the leftmost element is
	// C_(k-(q-1)).
	if s.historyLen < s.q {
		s.history = append(s.history, sample)
		s.historyLen++
	} else {
		copy(s.history, s.history[1:])
		s.history[s.q-1] = sample
	}

	// Momentum needs a price from q-1 bars ago, available only once the window
	// holds q prices. Before then neither output is defined and the signal EMA
	// is NOT advanced.
	if s.historyLen < s.q {
		return math.NaN(), math.NaN()
	}

	// mtm_k = C_k - C_(k-(q-1)); the leftmost history element is C_(k-(q-1)).
	mtm := sample - s.history[0]
	absMtm := math.Abs(mtm)

	// Numerator cascade: TEMA(mtm, r, s, u).
	n := s.numU.update(s.numS.update(s.numR.update(mtm)))
	// Denominator cascade: TEMA(|mtm|, r, s, u).
	d := s.denU.update(s.denS.update(s.denR.update(absMtm)))

	// Division guard (Blau_TSI.mq5): denominator 0 -> oscillator 0.0.
	tsi = 0.0
	if d != 0.0 {
		tsi = 100.0 * n / d
	}

	// Signal line = EMA(tsi, ul); seeds here on the first finite oscillator.
	signal = s.signalEMA.update(tsi)
	s.primed = true

	return tsi, signal
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *TrueStrengthIndex) UpdateScalar(sample *entities.Scalar) core.Output {
	tsi, signal := s.Update(sample.Value)

	const outputCount = 2

	output := make([]any, outputCount)
	output[0] = entities.Scalar{Time: sample.Time, Value: tsi}
	output[1] = entities.Scalar{Time: sample.Time, Value: signal}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (s *TrueStrengthIndex) UpdateBar(sample *entities.Bar) core.Output {
	v := s.barFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *TrueStrengthIndex) UpdateQuote(sample *entities.Quote) core.Output {
	v := s.quoteFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *TrueStrengthIndex) UpdateTrade(sample *entities.Trade) core.Output {
	v := s.tradeFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}
