package slopedivergencetsifilter

import (
	"cmp"
	"fmt"
	"math"
	"sync"

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

// SlopeDivergenceTsiFilter is William Blau's Slope Divergence TSI Filter (SD_TSI).
//
// A trend/congestion prefilter built on the True Strength Index. It keeps the
// TSI value only when the slope of the TSI agrees in sign with the slope of a
// separate double EMA of price; otherwise it outputs 0 (a slope divergence, or
// congestion zone):
//
//	ind_k = TSI(close, q, r, s, u)_k
//	ref_k = DEMA(close, x, y)_k = EMA(EMA(close, x), y)_k
//
//	SD_TSI_k = ind_k   if ind_k - ind_(k-1) > 0 and ref_k - ref_(k-1) > 0
//	         = ind_k   if ind_k - ind_(k-1) < 0 and ref_k - ref_(k-1) < 0
//	         = 0       otherwise
//
// The gate is strict (book Fig. B-25): a flat slope on either series yields 0.
// The output range is [-100, +100].
//
// Priming convention (book / EasyLanguage): the TSI is NaN for bars 0..q-2
// (momentum look-back), so SD_TSI is NaN there. The price DEMA seeds at bar 0
// and advances every bar, including through the TSI warm-up. At the first finite
// TSI bar there is no prior TSI value, hence no slope, so the output is 0.0.
//
// Reference:
//
// Blau, William (1995). Momentum, Direction, and Divergence, ch. 12, Appendix B
// Fig. B-25. Wiley.
type SlopeDivergenceTsiFilter struct {
	mu sync.RWMutex
	core.LineIndicator

	tsi *truestrengthindex.TrueStrengthIndex

	referenceX *ema
	referenceY *ema

	previousTsi       float64
	hasPreviousTsi    bool
	previousReference float64

	primed bool
}

// NewSlopeDivergenceTsiFilter returns an instance of the indicator created using supplied parameters.
//
//nolint:funlen,cyclop
func NewSlopeDivergenceTsiFilter(p *Params) (*SlopeDivergenceTsiFilter, error) {
	const (
		invalid  = "invalid slope divergence tsi filter parameters"
		fmts     = "%s: %s"
		fmtw     = "%s: %w"
		fmtn     = "sdtsi(%d,%d,%d,%d,%d,%d%s)"
		defaultQ = 2
		defaultR = 32
		defaultS = 32
		defaultU = 7
		defaultX = 32
		defaultY = 7
	)

	q := cmp.Or(p.Q, defaultQ)
	r := cmp.Or(p.R, defaultR)
	s := cmp.Or(p.S, defaultS)
	u := cmp.Or(p.U, defaultU)
	x := cmp.Or(p.X, defaultX)
	y := cmp.Or(p.Y, defaultY)

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

	if x < 1 {
		return nil, fmt.Errorf(fmts, invalid, "x should be greater than 0")
	}

	if y < 1 {
		return nil, fmt.Errorf(fmts, invalid, "y should be greater than 0")
	}

	// Resolve defaults for component functions.
	// A zero value means "use default, don't show in mnemonic".
	bc := cmp.Or(p.BarComponent, entities.DefaultBarComponent)
	qc := cmp.Or(p.QuoteComponent, entities.DefaultQuoteComponent)
	tc := cmp.Or(p.TradeComponent, entities.DefaultTradeComponent)

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

	// The TSI oscillator (its signal line is unused, so ul=1).
	tsi, err := truestrengthindex.NewTrueStrengthIndex(&truestrengthindex.Params{
		Q: q, R: r, S: s, U: u, Ul: 1,
	})
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	// Build mnemonic using resolved components — defaults are omitted by ComponentTripleMnemonic.
	mnemonic := fmt.Sprintf(fmtn, q, r, s, u, x, y, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Slope Divergence TSI Filter " + mnemonic

	sd := &SlopeDivergenceTsiFilter{
		tsi: tsi,

		// Price reference: DEMA(close, x, y) = EMA(EMA(close, x), y).
		referenceX: newEMA(x),
		referenceY: newEMA(y),
	}

	sd.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, sd.Update)

	return sd, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *SlopeDivergenceTsiFilter) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
// It always has a single scalar output -- the calculated value of the filter.
func (s *SlopeDivergenceTsiFilter) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.SlopeDivergenceTsiFilter,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the value of the indicator given the next sample.
//
// The indicator is not primed during the first q-1 updates.
func (s *SlopeDivergenceTsiFilter) Update(sample float64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	// The price reference advances every bar (it has no NaN warm-up).
	reference := s.referenceY.update(s.referenceX.update(sample))

	tsi, _ := s.tsi.Update(sample)
	if math.IsNaN(tsi) {
		// TSI momentum warm-up: keep the previous-bar reference current.
		s.previousReference = reference

		return math.NaN()
	}

	result := 0.0

	// At the first finite TSI there is no prior TSI, hence no slope.
	if s.hasPreviousTsi {
		deltaTsi := tsi - s.previousTsi
		deltaReference := reference - s.previousReference

		// Keep the TSI only when both slopes are strictly same-signed.
		if (deltaTsi > 0.0 && deltaReference > 0.0) || (deltaTsi < 0.0 && deltaReference < 0.0) {
			result = tsi
		}
	}

	s.previousTsi = tsi
	s.hasPreviousTsi = true
	s.previousReference = reference
	s.primed = true

	return result
}
