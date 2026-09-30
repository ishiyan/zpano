package nonambiguoustrendfilter

import (
	"cmp"
	"fmt"
	"math"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/williamblau/candlestickmomentumindex"
	"zpano/indicators/williamblau/candlestickstrengthindex"
	"zpano/indicators/williamblau/directionaltrendindex"
	"zpano/indicators/williamblau/meandeviationindex"
	"zpano/indicators/williamblau/stochasticmomentumindex"
	"zpano/indicators/williamblau/tickvolumeindicator"
	"zpano/indicators/williamblau/truestrengthindex"
)

// baseDefaults holds the book named-instance defaults of a base.
type baseDefaults struct {
	name       string
	q, r, s, u int
	usesQ      bool
}

// defaults lists the book named-instance defaults per base.
var defaults = map[Base]baseDefaults{
	Tsi: {name: "tsi", q: 2, r: 32, s: 13, u: 3, usesQ: true},
	Smi: {name: "smi", q: 32, r: 64, s: 7, u: 1, usesQ: true},
	Dti: {name: "dti", q: 2, r: 28, s: 28, u: 5, usesQ: true},
	Tvi: {name: "tvi", r: 32, s: 32, u: 5},
	Mdi: {name: "mdi", r: 20, s: 5, u: 3},
	Cmi: {name: "cmi", r: 20, s: 5, u: 3},
	Csi: {name: "csi", r: 32, s: 32, u: 1},
}

// NonambiguousTrendFilter is William Blau's Nonambiguous Trend Filter (_Trade).
//
// A post-processing transform applied to a normalized, signed base oscillator X
// (TSI, SMI, DTI, TVI, MDI, CMI, CSI). It keeps X only where its sign and slope
// agree and zeroes every ambiguous bar (book Ch. 8, Appendix B Figs. B-20..B-23):
//
//	X_Trade[k] = X[k]   if X[k] > 0 and X[k] - X[k-1] > 0   (positive and rising)
//	           = X[k]   if X[k] < 0 and X[k] - X[k-1] < 0   (negative and falling)
//	           = 0      otherwise                           (ambiguous)
//
// The nonzero stretches correspond one-to-one with genuine up/down trends;
// congestion and flat regions are blanked to zero.
//
// The filter wraps an instance of the base indicator: every sample is routed to
// the base with its own entity mapping, and the base's primary output is filtered.
//
// Conventions: a NaN base value (the base's own look-back warm-up) yields NaN and
// leaves the filter state untouched; the first finite base value has no prior
// slope, so the output is 0.0; a flat step (delta == 0) is neither rising nor
// falling, so it is zeroed.
//
// Reference:
//
// Blau, William (1995). Momentum, Direction, and Divergence, ch. 8, Appendix B
// Figs. B-20..B-23. Wiley.
type NonambiguousTrendFilter struct {
	mu sync.RWMutex

	base core.Indicator

	previous float64
	primed   bool

	mnemonic    string
	description string
}

// NewNonambiguousTrendFilter returns an instance of the indicator created using supplied parameters.
//
//nolint:funlen,cyclop
func NewNonambiguousTrendFilter(p *Params) (*NonambiguousTrendFilter, error) {
	const (
		invalid = "invalid nonambiguous trend filter parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
	)

	d, ok := defaults[p.Base]
	if !ok {
		return nil, fmt.Errorf("%s: unknown base %d", invalid, p.Base)
	}

	q := cmp.Or(p.Q, d.q)
	r := cmp.Or(p.R, d.r)
	s := cmp.Or(p.S, d.s)
	u := cmp.Or(p.U, d.u)

	if d.usesQ && q < 1 {
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

	// Price components are meaningful only for the single-price Tsi and Mdi bases.
	bc := entities.DefaultBarComponent
	qc := entities.DefaultQuoteComponent
	tc := entities.DefaultTradeComponent

	if p.Base == Tsi || p.Base == Mdi {
		bc = cmp.Or(p.BarComponent, bc)
		qc = cmp.Or(p.QuoteComponent, qc)
		tc = cmp.Or(p.TradeComponent, tc)
	}

	// The base signal line is unused, so its period is 1.
	base, err := newBase(p.Base, q, r, s, u, bc, qc, tc)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	var mnemonic string
	if d.usesQ {
		mnemonic = fmt.Sprintf("ntf.%s(%d,%d,%d,%d%s)", d.name, q, r, s, u, core.ComponentTripleMnemonic(bc, qc, tc))
	} else {
		mnemonic = fmt.Sprintf("ntf.%s(%d,%d,%d%s)", d.name, r, s, u, core.ComponentTripleMnemonic(bc, qc, tc))
	}

	return &NonambiguousTrendFilter{
		base:        base,
		mnemonic:    mnemonic,
		description: "Nonambiguous Trend Filter " + mnemonic,
	}, nil
}

// newBase creates the base indicator.
//
//nolint:cyclop
func newBase(b Base, q, r, s, u int,
	bc entities.BarComponent, qc entities.QuoteComponent, tc entities.TradeComponent,
) (core.Indicator, error) {
	switch b {
	case Tsi:
		return asIndicator(truestrengthindex.NewTrueStrengthIndex(&truestrengthindex.Params{
			Q: q, R: r, S: s, U: u, Ul: 1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}))
	case Smi:
		return asIndicator(stochasticmomentumindex.NewStochasticMomentumIndex(&stochasticmomentumindex.Params{
			Q: q, R: r, S: s, U: u, Ul: 1,
		}))
	case Dti:
		return asIndicator(directionaltrendindex.NewDirectionalTrendIndex(&directionaltrendindex.Params{
			Q: q, R: r, S: s, U: u, Ul: 1,
		}))
	case Tvi:
		return asIndicator(tickvolumeindicator.NewTickVolumeIndicator(&tickvolumeindicator.Params{
			R: r, S: s, U: u,
		}))
	case Mdi:
		return asIndicator(meandeviationindex.NewMeanDeviationIndex(&meandeviationindex.Params{
			R: r, S: s, U: u, Ul: 1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}))
	case Cmi:
		return asIndicator(candlestickmomentumindex.NewCandlestickMomentumIndex(&candlestickmomentumindex.Params{
			R: r, S: s, U: u, Ul: 1,
		}))
	default:
		return asIndicator(candlestickstrengthindex.NewCandlestickStrengthIndex(&candlestickstrengthindex.Params{
			R: r, S: s, U: u, Ul: 1,
		}))
	}
}

// asIndicator converts a constructor result to the Indicator interface without
// producing a non-nil interface holding a nil pointer.
func asIndicator[T core.Indicator](ind T, err error) (core.Indicator, error) {
	if err != nil {
		return nil, err
	}

	return ind, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *NonambiguousTrendFilter) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
// It always has a single scalar output -- the calculated value of the filter.
func (s *NonambiguousTrendFilter) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.NonambiguousTrendFilter,
		s.mnemonic,
		s.description,
		[]core.OutputText{
			{Mnemonic: s.mnemonic, Description: s.description},
		},
	)
}

// filter keeps x when positive-and-rising or negative-and-falling, else 0; the
// caller must hold the write lock.
func (s *NonambiguousTrendFilter) filter(x float64) float64 {
	if math.IsNaN(x) {
		// The base is still warming up: do not touch the filter state.
		return math.NaN()
	}

	if !s.primed {
		// First finite value: no prior slope, hence ambiguous.
		s.previous = x
		s.primed = true

		return 0.0
	}

	delta := x - s.previous
	s.previous = x

	switch {
	case x > 0.0 && delta > 0.0: // Positive and rising.
		return x
	case x < 0.0 && delta < 0.0: // Negative and falling.
		return x
	default: // Ambiguous, flat or congestion.
		return 0.0
	}
}

// filterOutput filters the primary output of the base; the caller must hold
// the write lock.
func (s *NonambiguousTrendFilter) filterOutput(output core.Output) float64 {
	scalar, ok := output[0].(entities.Scalar)
	if !ok {
		return math.NaN()
	}

	return s.filter(scalar.Value)
}

// wrap wraps the filter value into the output.
func wrap(time time.Time, value float64) core.Output {
	output := make([]any, 1)
	output[0] = entities.Scalar{Time: time, Value: value}

	return output
}

// Update updates the value of the indicator given the next single sample value,
// which is fed to the base as a scalar.
func (s *NonambiguousTrendFilter) Update(sample float64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.filterOutput(s.base.UpdateScalar(&entities.Scalar{Value: sample}))
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *NonambiguousTrendFilter) UpdateScalar(sample *entities.Scalar) core.Output {
	s.mu.Lock()
	defer s.mu.Unlock()

	return wrap(sample.Time, s.filterOutput(s.base.UpdateScalar(sample)))
}

// UpdateBar updates the indicator given the next bar sample.
func (s *NonambiguousTrendFilter) UpdateBar(sample *entities.Bar) core.Output {
	s.mu.Lock()
	defer s.mu.Unlock()

	return wrap(sample.Time, s.filterOutput(s.base.UpdateBar(sample)))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *NonambiguousTrendFilter) UpdateQuote(sample *entities.Quote) core.Output {
	s.mu.Lock()
	defer s.mu.Unlock()

	return wrap(sample.Time, s.filterOutput(s.base.UpdateQuote(sample)))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *NonambiguousTrendFilter) UpdateTrade(sample *entities.Trade) core.Output {
	s.mu.Lock()
	defer s.mu.Unlock()

	return wrap(sample.Time, s.filterOutput(s.base.UpdateTrade(sample)))
}
