package doublesmoothedmomenta

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

// DoubleSmoothedMomenta is William Blau's Double-Smoothed Momenta (DM) indicator,
// also known as the Double-Smoothed RSI (DRSI) when the look-back is fixed at 2.
//
// A close-based, double-smoothed momentum oscillator bounded to [0, 100]:
//
//	LCa_k = min(Close over the last a bars)          (lowest close)
//	HCa_k = max(Close over the last a bars)          (highest close)
//	st_k  = Close_k - LCa_k                          (close above the low close)
//	rng_k = HCa_k - LCa_k                            (a-bar close range)
//
//	DM(a, y, z) = 100 * EMA(EMA(st, y), z) / EMA(EMA(rng, y), z)
//
// Each of the numerator (st) and denominator (rng) series is double-smoothed by an
// inner EMA of period y then an outer EMA of period z (Blau's Ez(Ey(.)) ), and the
// ratio is scaled by 100.
//
// This is structurally the Double-Smoothed Stochastic computed on the CLOSE -- it
// uses the highest/lowest close over a bars instead of the high/low of the bar --
// and it has no signal line, so it produces a single scalar output per bar.
//
// Named instances:
//   - RSI equivalence:     DM(2, 1, z) == RSI(z), the EMA-form RSI;
//   - Double-smoothed RSI: DRSI(y, z) = DM(2, y, z).
//
// The equivalence DM(2,1,z) == RSI(z) holds for the EMA-form RSI built from the
// Blau EMA (alpha = 2/(z+1)), NOT Wilder's classic RSI (which uses RMA smoothing,
// alpha = 1/z).
//
// Priming convention (book / EasyLanguage): st/rng are valid once a closes exist
// (bar a-1); all four EMA stages seed there. DM is NaN for bars 0..a-2 and finite
// from bar a-1. For a == 1 there is no NaN warm-up, but the a-bar close range is
// then always 0, so DM is 0.0 on every bar via the guard (a degenerate setting).
//
// Division guard: EMA(EMA(rng)) <= 0 -> DM = 0.0.
//
// Reference:
//
// Blau, William (1995). Momentum, Direction, and Divergence. Wiley.
type DoubleSmoothedMomenta struct {
	mu sync.RWMutex
	core.LineIndicator

	window       []float64
	windowLength int
	windowCount  int
	lastIndex    int

	numeratorY   *ema
	numeratorZ   *ema
	denominatorY *ema
	denominatorZ *ema

	primed bool
}

// NewDoubleSmoothedMomenta returns an instance of the indicator created using supplied parameters.
//
//nolint:funlen,cyclop
func NewDoubleSmoothedMomenta(p *Params) (*DoubleSmoothedMomenta, error) {
	const (
		invalid  = "invalid double smoothed momenta parameters"
		fmts     = "%s: %s"
		fmtw     = "%s: %w"
		fmtn     = "dm(%d,%d,%d%s)"
		defaultA = 2
		defaultY = 2
		defaultZ = 14
	)

	a := p.A
	if a == 0 {
		a = defaultA
	}

	y := p.Y
	if y == 0 {
		y = defaultY
	}

	z := p.Z
	if z == 0 {
		z = defaultZ
	}

	if a < 1 {
		return nil, fmt.Errorf(fmts, invalid, "a should be greater than 0")
	}

	if y < 1 {
		return nil, fmt.Errorf(fmts, invalid, "y should be greater than 0")
	}

	if z < 1 {
		return nil, fmt.Errorf(fmts, invalid, "z should be greater than 0")
	}

	// Resolve defaults for component functions.
	// A zero value means "use default, don't show in mnemonic".
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

	// Build mnemonic using resolved components — defaults are omitted by ComponentTripleMnemonic.
	mnemonic := fmt.Sprintf(fmtn, a, y, z, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Double-Smoothed Momenta " + mnemonic

	dm := &DoubleSmoothedMomenta{
		// Rolling window of the last a closes (for the highest/lowest close).
		window:       make([]float64, a),
		windowLength: a,
		lastIndex:    a - 1,

		// Two independent 2-stage EMA cascades (double smoothing), each wired
		// inner(y) -> outer(z): EMA(EMA(x, y), z).
		numeratorY:   newEMA(y),
		numeratorZ:   newEMA(z),
		denominatorY: newEMA(y),
		denominatorZ: newEMA(z),
	}

	dm.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, dm.Update)

	return dm, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *DoubleSmoothedMomenta) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
// It always has a single scalar output -- the calculated value of the oscillator.
func (s *DoubleSmoothedMomenta) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.DoubleSmoothedMomenta,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the value of the indicator given the next sample.
//
// The indicator is not primed during the first a-1 updates.
func (s *DoubleSmoothedMomenta) Update(sample float64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.primed {
		for i := 0; i < s.lastIndex; i++ {
			s.window[i] = s.window[i+1]
		}

		s.window[s.lastIndex] = sample
	} else {
		s.window[s.windowCount] = sample
		s.windowCount++

		// Need a closes before the highest/lowest close is defined. While
		// unprimed, the EMA cascades must NOT advance (they seed at bar a-1).
		if s.windowLength > s.windowCount {
			return math.NaN()
		}

		s.primed = true
	}

	// Highest/lowest close over the last a bars.
	hc := s.window[0]
	lc := s.window[0]

	for i := 1; i < s.windowLength; i++ {
		v := s.window[i]

		if v > hc {
			hc = v
		}

		if v < lc {
			lc = v
		}
	}

	// Raw close-above-low (>= 0) and a-bar close range (>= 0).
	st := sample - lc
	rng := hc - lc

	// Double-smooth each separately (inner y, then outer z), then divide.
	num := s.numeratorZ.update(s.numeratorY.update(st))
	den := s.denominatorZ.update(s.denominatorY.update(rng))

	// Division guard: smoothed range <= 0 -> DM = 0.0.
	if den <= 0.0 {
		return 0.0
	}

	return 100.0 * num / den
}
