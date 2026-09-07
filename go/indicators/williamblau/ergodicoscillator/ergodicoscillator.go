package ergodicoscillator

import (
	"fmt"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/williamblau/truestrengthindex"
)

// ErgodicOscillator is William Blau's Ergodic Oscillator indicator.
//
// The Ergodic is the True Strength Index plotted together with a signal line --
// the EMA of the oscillator that Blau introduces as the trading vehicle for the
// TSI (ch. 2, Fig. 2-14):
//
//	ergodic_k = TSI(q, r, s, u)_k                    (the oscillator)
//	signal_k  = EMA(ergodic, ul)_k                   (ul-period EMA of it)
//
// The indicator produces two outputs:
//   - Ergodic: the oscillator, range [-100, +100], NaN during warm-up (bars 0..q-2);
//   - Signal: the ul-period EMA of the oscillator.
//
// The numerics are exactly those of the True Strength Index, so this indicator
// wraps a [truestrengthindex.TrueStrengthIndex] instance instead of duplicating
// the triple EMA cascade. Only the mnemonic, the identifier, and the output
// naming differ.
//
// Priming convention (book / EasyLanguage): both outputs are NaN for bars
// 0..q-2 and finite from bar q-1 onward; the signal EMA seeds on the first
// finite oscillator value. Division guard: denominator 0 -> oscillator 0.0.
//
// Reference:
//
// Blau, William (1995). Momentum, Direction, and Divergence, ch. 2. Wiley.
type ErgodicOscillator struct {
	mu sync.RWMutex

	tsi *truestrengthindex.TrueStrengthIndex

	barFunc   entities.BarFunc
	quoteFunc entities.QuoteFunc
	tradeFunc entities.TradeFunc

	mnemonic string
}

// NewErgodicOscillator returns an instance of the indicator created using supplied parameters.
func NewErgodicOscillator(p *Params) (*ErgodicOscillator, error) {
	const (
		invalid   = "invalid ergodic oscillator parameters"
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

	// The oscillator and its signal line are exactly the True Strength Index
	// outputs; wrap an instance rather than duplicating its numerics. The
	// resolved components are passed through so both agree on the mnemonic.
	// Parameter validation is performed by the wrapped indicator.
	tsi, err := truestrengthindex.NewTrueStrengthIndex(&truestrengthindex.Params{
		Q: q, R: r, S: s, U: u, Ul: ul,
		BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
	})
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	mnemonic := fmt.Sprintf("ergodic(%d,%d,%d,%d,%d%s)", q, r, s, u, ul,
		core.ComponentTripleMnemonic(bc, qc, tc))

	return &ErgodicOscillator{
		tsi:       tsi,
		barFunc:   barFunc,
		quoteFunc: quoteFunc,
		tradeFunc: tradeFunc,
		mnemonic:  mnemonic,
	}, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *ErgodicOscillator) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.tsi.IsPrimed()
}

// Metadata describes the output data of the indicator.
func (s *ErgodicOscillator) Metadata() core.Metadata {
	desc := "Ergodic Oscillator " + s.mnemonic

	return core.BuildMetadata(
		core.ErgodicOscillator,
		s.mnemonic,
		desc,
		[]core.OutputText{
			{Mnemonic: s.mnemonic + " ergodic", Description: desc + " ergodic"},
			{Mnemonic: s.mnemonic + " signal", Description: desc + " signal"},
		},
	)
}

// Update updates the indicator given the next sample value.
// Returns ergodic, signal values.
//
//nolint:nonamedreturns
func (s *ErgodicOscillator) Update(sample float64) (ergodic, signal float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.tsi.Update(sample)
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *ErgodicOscillator) UpdateScalar(sample *entities.Scalar) core.Output {
	ergodic, signal := s.Update(sample.Value)

	const outputCount = 2

	output := make([]any, outputCount)
	output[0] = entities.Scalar{Time: sample.Time, Value: ergodic}
	output[1] = entities.Scalar{Time: sample.Time, Value: signal}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (s *ErgodicOscillator) UpdateBar(sample *entities.Bar) core.Output {
	v := s.barFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *ErgodicOscillator) UpdateQuote(sample *entities.Quote) core.Output {
	v := s.quoteFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *ErgodicOscillator) UpdateTrade(sample *entities.Trade) core.Output {
	v := s.tradeFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}
