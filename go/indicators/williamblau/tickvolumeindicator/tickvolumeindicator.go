package tickvolumeindicator

import (
	"cmp"
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

// TickVolumeIndicator is William Blau's Tick Volume Indicator (TVI).
//
// A normalized, double-/triple-smoothed oscillator built from the balance of
// upticks vs downticks inside each bar, bounded to [-100, +100] (Blau ch.4, ch.10):
//
//	tvi_k = 100 * (TEMA(up, r, s, u) - TEMA(down, r, s, u))
//	            / (TEMA(up, r, s, u) + TEMA(down, r, s, u))
//
// where TEMA(x, r, s, u) = EMA(EMA(EMA(x, r), s), u). Setting u=1 recovers the
// book's double-smoothed TVI(r, s), because EMA(., 1) is a passthrough. Because it
// is built from intra-bar tick direction rather than from the close vs a previous
// close, the TVI is immune to opening gaps.
//
// The inputs are two non-negative series, upticks and downticks. Genuine tick
// counts are fed through Update. The entity updates derive a deterministic proxy:
//   - Bar: up = close - low, down = high - close (the intra-bar range split);
//   - Scalar, Trade, Quote: a magnitude tick rule against the previous value
//     (value, price or mid price): up = max(x - previous, 0),
//     down = max(previous - x, 0). The first sample yields (0, 0). On such a
//     single-valued series the TVI reduces to a True Strength Index of the
//     one-step momentum.
//
// Priming convention (book / EasyLanguage): each EMA stage seeds on its first
// received value, so there is no NaN warm-up region and the output is finite for
// every update. Division guard: denominator 0 (a fully flat market) -> 0.0.
//
// Reference:
//
// Blau, William (1995). Momentum, Direction, and Divergence, ch. 4, ch. 10. Wiley.
type TickVolumeIndicator struct {
	mu sync.RWMutex

	upR   *ema
	upS   *ema
	upU   *ema
	downR *ema
	downS *ema
	downU *ema

	previous    float64
	hasPrevious bool

	primed bool

	mnemonic string
}

// NewTickVolumeIndicator returns an instance of the indicator created using supplied parameters.
func NewTickVolumeIndicator(p *Params) (*TickVolumeIndicator, error) {
	const (
		invalid  = "invalid tick volume indicator parameters"
		fmts     = "%s: %s"
		defaultR = 12
		defaultS = 12
		defaultU = 1
	)

	r := cmp.Or(p.R, defaultR)
	s := cmp.Or(p.S, defaultS)
	u := cmp.Or(p.U, defaultU)

	if r < 1 {
		return nil, fmt.Errorf(fmts, invalid, "r should be greater than 0")
	}

	if s < 1 {
		return nil, fmt.Errorf(fmts, invalid, "s should be greater than 0")
	}

	if u < 1 {
		return nil, fmt.Errorf(fmts, invalid, "u should be greater than 0")
	}

	mnemonic := fmt.Sprintf("tvi(%d,%d,%d)", r, s, u)

	return &TickVolumeIndicator{
		upR:      newEMA(r),
		upS:      newEMA(s),
		upU:      newEMA(u),
		downR:    newEMA(r),
		downS:    newEMA(s),
		downU:    newEMA(u),
		mnemonic: mnemonic,
	}, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *TickVolumeIndicator) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *TickVolumeIndicator) Metadata() core.Metadata {
	desc := "Tick Volume Indicator " + s.mnemonic

	return core.BuildMetadata(
		core.TickVolumeIndicator,
		s.mnemonic,
		desc,
		[]core.OutputText{
			{Mnemonic: s.mnemonic, Description: desc},
		},
	)
}

// Update updates the indicator given the next bar's upticks and downticks.
// Returns the TVI value.
func (s *TickVolumeIndicator) Update(upticks, downticks float64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.update(upticks, downticks)
}

// update is the unlocked core of Update; the caller must hold the write lock.
func (s *TickVolumeIndicator) update(upticks, downticks float64) float64 {
	// Upticks cascade: TEMA(up, r, s, u).
	up := s.upU.update(s.upS.update(s.upR.update(upticks)))
	// Downticks cascade: TEMA(down, r, s, u).
	down := s.downU.update(s.downS.update(s.downR.update(downticks)))

	s.primed = true

	// Division guard (Appendix B): fully flat smoothed volume -> 0.0.
	denominator := up + down
	if denominator == 0.0 {
		return 0.0
	}

	return 100.0 * (up - down) / denominator
}

// updateTickRule derives the upticks and downticks from the change of the value
// vs the previous value and updates the indicator.
func (s *TickVolumeIndicator) updateTickRule(value float64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	upticks, downticks := 0.0, 0.0
	if s.hasPrevious {
		diff := value - s.previous
		upticks = max(diff, 0.0)
		downticks = max(-diff, 0.0)
	}

	s.previous = value
	s.hasPrevious = true

	return s.update(upticks, downticks)
}

// wrap wraps the TVI value into the output.
func wrap(time time.Time, value float64) core.Output {
	output := make([]any, 1)
	output[0] = entities.Scalar{Time: time, Value: value}

	return output
}

// UpdateScalar updates the indicator given the next scalar sample.
//
// A scalar carries a single value, so the tick rule is applied to it.
func (s *TickVolumeIndicator) UpdateScalar(sample *entities.Scalar) core.Output {
	return wrap(sample.Time, s.updateTickRule(sample.Value))
}

// UpdateBar updates the indicator given the next bar sample.
//
// A bar splits its range: up = close - low, down = high - close.
func (s *TickVolumeIndicator) UpdateBar(sample *entities.Bar) core.Output {
	return wrap(sample.Time, s.Update(sample.Close-sample.Low, sample.High-sample.Close))
}

// UpdateQuote updates the indicator given the next quote sample.
//
// A quote applies the tick rule to its mid price.
func (s *TickVolumeIndicator) UpdateQuote(sample *entities.Quote) core.Output {
	return wrap(sample.Time, s.updateTickRule(sample.Mid()))
}

// UpdateTrade updates the indicator given the next trade sample.
//
// A trade applies the tick rule to its price.
func (s *TickVolumeIndicator) UpdateTrade(sample *entities.Trade) core.Output {
	return wrap(sample.Time, s.updateTickRule(sample.Price))
}
