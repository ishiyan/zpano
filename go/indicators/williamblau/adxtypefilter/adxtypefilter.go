package adxtypefilter

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/williamblau/truestrengthindex"
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

// push appends a sample to a rolling window holding at most length values,
// dropping the oldest one when the window is full.
func push(window []float64, length int, sample float64) []float64 {
	if len(window) < length {
		return append(window, sample)
	}

	copy(window, window[1:])
	window[length-1] = sample

	return window
}

// AdxTypeFilter is William Blau's ADX-Type Filter (ATF).
//
// A non-negative trend-strength filter, analogous to Wilder's ADX, built by
// rectifying and double-smoothing a bipolar momentum series (book Fig. B-24):
//
//	ATF(Price, r, s) = EMA(|EMA(Price, r)|, s)
//
// The inner EMA(r) smooths the signed momentum, the absolute value discards the
// direction and keeps the amplitude, and the outer EMA(s) smooths the amplitude.
// A rising ATF signals a strengthening trend, a falling ATF a ranging market.
//
// The bipolar momentum is selected by the source:
//   - TsiMomentum:   C - C[q-1]                               (TSI_ATF);
//   - SmiMomentum:   C - 0.5*(HH(q) + LL(q))                  (SMI_ATF);
//   - DtiMomentum:   max(H - H[q-1], 0) - max(L[q-1] - L, 0);
//   - TviBalance:    upticks - downticks (2C - H - L for a bar);
//   - TsiNormalized: TSI(q, r, 1, 1), which replaces the inner EMA (r = 1).
//
// Priming convention (book / EasyLanguage): each EMA stage seeds on its first
// finite momentum value. A NaN momentum (the q-bar look-back warm-up) is
// propagated: the output is NaN and the EMAs do not advance. The output is
// always >= 0.
//
// Reference:
//
// Blau, William (1995). Momentum, Direction, and Divergence, Appendix B Fig. B-24. Wiley.
type AdxTypeFilter struct {
	mu sync.RWMutex

	source         Source
	usesComponents bool
	q              int
	closes         []float64
	highs          []float64
	lows           []float64
	previous       float64
	hasPrevious    bool
	tsi            *truestrengthindex.TrueStrengthIndex
	inner          *ema
	outer          *ema
	primed         bool
	barFunc        entities.BarFunc
	quoteFunc      entities.QuoteFunc
	tradeFunc      entities.TradeFunc
	mnemonic       string
	description    string
}

// NewAdxTypeFilter returns an instance of the indicator created using supplied parameters.
//
//nolint:funlen,cyclop
func NewAdxTypeFilter(p *Params) (*AdxTypeFilter, error) {
	const (
		invalid     = "invalid adx type filter parameters"
		fmts        = "%s: %s"
		fmtw        = "%s: %w"
		defaultQ    = 2
		defaultQSmi = 32
		defaultR    = 32
		defaultS    = 32
	)

	var name string

	switch p.Source {
	case TsiMomentum:
		name = "tsi"
	case SmiMomentum:
		name = "smi"
	case DtiMomentum:
		name = "dti"
	case TviBalance:
		name = "tvi"
	case TsiNormalized:
		name = "tsin"
	default:
		return nil, fmt.Errorf("%s: unknown source %d", invalid, p.Source)
	}

	q := p.Q
	if q == 0 {
		q = defaultQ
		if p.Source == SmiMomentum {
			q = defaultQSmi
		}
	}

	r := cmp.Or(p.R, defaultR)
	s := cmp.Or(p.S, defaultS)

	if q < 1 {
		return nil, fmt.Errorf(fmts, invalid, "q should be greater than 0")
	}

	if r < 1 {
		return nil, fmt.Errorf(fmts, invalid, "r should be greater than 0")
	}

	if s < 1 {
		return nil, fmt.Errorf(fmts, invalid, "s should be greater than 0")
	}

	// Price components are meaningful only for the single-price TSI sources;
	// the other sources use the bar's high/low/close or the single value of the
	// sample (scalar value, quote mid price, trade price).
	usesComponents := p.Source == TsiMomentum || p.Source == TsiNormalized

	bc := entities.DefaultBarComponent
	qc := entities.DefaultQuoteComponent
	tc := entities.DefaultTradeComponent

	if usesComponents {
		bc = cmp.Or(p.BarComponent, bc)
		qc = cmp.Or(p.QuoteComponent, qc)
		tc = cmp.Or(p.TradeComponent, tc)
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

	var mnemonic string
	if p.Source == TviBalance {
		mnemonic = fmt.Sprintf("atf.%s(%d,%d)", name, r, s)
	} else {
		mnemonic = fmt.Sprintf("atf.%s(%d,%d,%d%s)", name, q, r, s, core.ComponentTripleMnemonic(bc, qc, tc))
	}

	a := &AdxTypeFilter{
		source:         p.Source,
		usesComponents: usesComponents,
		q:              q,
		closes:         make([]float64, 0, q),
		highs:          make([]float64, 0, q),
		lows:           make([]float64, 0, q),
		inner:          newEMA(r),
		outer:          newEMA(s),
		barFunc:        barFunc,
		quoteFunc:      quoteFunc,
		tradeFunc:      tradeFunc,
		mnemonic:       mnemonic,
		description:    "ADX-Type Filter " + mnemonic,
	}

	// The normalized TSI replaces the inner EMA, which becomes a passthrough.
	if p.Source == TsiNormalized {
		if a.tsi, err = truestrengthindex.NewTrueStrengthIndex(&truestrengthindex.Params{
			Q: q, R: r, S: 1, U: 1, Ul: 1,
		}); err != nil {
			return nil, fmt.Errorf(fmtw, invalid, err)
		}

		a.inner = newEMA(1)
	}

	return a, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *AdxTypeFilter) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
// It always has a single scalar output -- the calculated value of the filter.
func (s *AdxTypeFilter) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.AdxTypeFilter,
		s.mnemonic,
		s.description,
		[]core.OutputText{
			{Mnemonic: s.mnemonic, Description: s.description},
		},
	)
}

// Update updates the value of the indicator given the next single sample value.
//
// The SmiMomentum and DtiMomentum sources use the value as the high, the low
// and the close; the TviBalance source applies the tick rule
// (balance = value - previous value).
func (s *AdxTypeFilter) Update(sample float64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.update(sample)
}

// UpdateHighLowClose updates the value of the indicator given the next bar's
// high, low and close. The TsiMomentum and TsiNormalized sources use the close only.
func (s *AdxTypeFilter) UpdateHighLowClose(high, low, close float64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.updateHighLowClose(high, low, close)
}

// filter is the inner smooth -> rectify -> outer smooth pipeline; the caller
// must hold the write lock. A NaN momentum propagates without advancing the EMAs.
func (s *AdxTypeFilter) filter(momentum float64) float64 {
	if math.IsNaN(momentum) {
		return math.NaN()
	}

	s.primed = true

	return s.outer.update(math.Abs(s.inner.update(momentum)))
}

// update is the unlocked core of Update; the caller must hold the write lock.
func (s *AdxTypeFilter) update(sample float64) float64 {
	switch s.source {
	case TsiMomentum:
		s.closes = push(s.closes, s.q, sample)
		if len(s.closes) < s.q {
			return math.NaN()
		}

		// mtm_k = C_k - C_(k-(q-1)); the leftmost element is C_(k-(q-1)).
		return s.filter(sample - s.closes[0])
	case TsiNormalized:
		tsi, _ := s.tsi.Update(sample)

		return s.filter(tsi)
	case TviBalance:
		balance := 0.0
		if s.hasPrevious {
			balance = sample - s.previous
		}

		s.previous = sample
		s.hasPrevious = true

		return s.filter(balance)
	default:
		return s.updateHighLowClose(sample, sample, sample)
	}
}

// updateHighLowClose is the unlocked core of UpdateHighLowClose; the caller
// must hold the write lock.
func (s *AdxTypeFilter) updateHighLowClose(high, low, close float64) float64 {
	switch s.source {
	case SmiMomentum:
		s.highs = push(s.highs, s.q, high)
		s.lows = push(s.lows, s.q, low)

		if len(s.highs) < s.q {
			return math.NaN()
		}

		// sm = C - 0.5*(HH(q) + LL(q)).
		return s.filter(close - 0.5*(slices.Max(s.highs)+slices.Min(s.lows)))
	case DtiMomentum:
		s.highs = push(s.highs, s.q, high)
		s.lows = push(s.lows, s.q, low)

		if len(s.highs) < s.q {
			return math.NaN()
		}

		// HMU - LMD; the leftmost elements are H_(k-(q-1)) and L_(k-(q-1)).
		hmu := max(high-s.highs[0], 0.0)
		lmd := max(s.lows[0]-low, 0.0)

		return s.filter(hmu - lmd)
	case TviBalance:
		// up - down = (C - L) - (H - C) = 2C - H - L.
		return s.filter(2.0*close - high - low)
	default:
		return s.update(close)
	}
}

// wrap wraps the ATF value into the output.
func wrap(time time.Time, value float64) core.Output {
	output := make([]any, 1)
	output[0] = entities.Scalar{Time: time, Value: value}

	return output
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *AdxTypeFilter) UpdateScalar(sample *entities.Scalar) core.Output {
	return wrap(sample.Time, s.Update(sample.Value))
}

// UpdateBar updates the indicator given the next bar sample.
//
// The TSI sources use the bar component; the other sources use the bar's high,
// low and close.
func (s *AdxTypeFilter) UpdateBar(sample *entities.Bar) core.Output {
	if s.usesComponents {
		return wrap(sample.Time, s.Update(s.barFunc(sample)))
	}

	return wrap(sample.Time, s.UpdateHighLowClose(sample.High, sample.Low, sample.Close))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *AdxTypeFilter) UpdateQuote(sample *entities.Quote) core.Output {
	return wrap(sample.Time, s.Update(s.quoteFunc(sample)))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *AdxTypeFilter) UpdateTrade(sample *entities.Trade) core.Output {
	return wrap(sample.Time, s.Update(s.tradeFunc(sample)))
}
