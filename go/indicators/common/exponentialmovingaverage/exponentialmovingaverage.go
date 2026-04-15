package exponentialmovingaverage

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
)

// ExponentialMovingAverage computes the exponential, or exponentially weighted, moving average (EMA).
//
// Given a constant smoothing percentage factor 0 < α ≤ 1, EMA is calculated by applying a constant
// smoothing factor α to a difference of today's sample and yesterday's EMA value:
//
//	EMAᵢ = αPᵢ + (1-α)EMAᵢ₋₁ = EMAᵢ₋₁ + α(Pᵢ - EMAᵢ₋₁), 0 < α ≤ 1.
//
// Thus, the weighting for each older sample is given by the geometric progression 1, α, α², α³, …,
// giving much more importance to recent observations while not discarding older ones: all data
// previously used are always part of the new EMA value.
//
// α may be expressed as a percentage, so a smoothing factor of 10% is equivalent to α = 0.1. A higher α
// discounts older observations faster. Alternatively, α may be expressed in terms of ℓ time periods (length),
// where:
//
//	α = 2 / (ℓ + 1) and ℓ = 2/α - 1.
//
// The indicator is not primed during the first ℓ-1 updates.
//
// The 12- and 26-day EMAs are the most popular short-term averages, and they are used to create indicators
// like MACD and PPO. In general, the 50- and 200-day EMAs are used as signals of long-term trends.
//
// The very first EMA value (the seed for subsequent values) is calculated differently. This implementation
// allows for two algorithms for this seed.
// ❶ Use a simple average of the first 'period'. This is the most widely documented approach.
// ❷ Use first sample value as a seed. This is used in Metastock.
type ExponentialMovingAverage struct {
	mu sync.RWMutex
	core.LineIndicator
	value           float64
	sum             float64
	smoothingFactor float64
	length          int
	count           int
	firstIsAverage  bool
	primed          bool
}

// NewExponentialMovingAverageLength returns an instnce of the indicator
// created using supplied parameters based on length.
func NewExponentialMovingAverageLength(
	p *ExponentialMovingAverageLengthParams,
) (*ExponentialMovingAverage, error) {
	return newExponentialMovingAverage(p.Length, math.NaN(), p.FirstIsAverage,
		p.BarComponent, p.QuoteComponent, p.TradeComponent)
}

// NewExponentialMovingAverageSmoothingFactor returns an instnce of the indicator
// created using supplied parameters based on smoothing factor.
func NewExponentialMovingAverageSmoothingFactor(
	p *ExponentialMovingAverageSmoothingFactorParams,
) (*ExponentialMovingAverage, error) {
	return newExponentialMovingAverage(0, p.SmoothingFactor, p.FirstIsAverage,
		p.BarComponent, p.QuoteComponent, p.TradeComponent)
}

//nolint:funlen,cyclop
func newExponentialMovingAverage(length int, alpha float64, firstIsAverage bool,
	bc entities.BarComponent, qc entities.QuoteComponent, tc entities.TradeComponent,
) (*ExponentialMovingAverage, error) {
	const (
		invalid = "invalid exponential moving average parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtnl   = "ema(%d%s)"
		fmtna   = "ema(%d, %.8f%s)"
		minlen  = 1
		two     = 2.
		epsilon = 0.00000001
	)

	var (
		mnemonic  string
		err       error
		barFunc   entities.BarFunc
		quoteFunc entities.QuoteFunc
		tradeFunc entities.TradeFunc
	)

	// Resolve defaults for component functions.
	// A zero value means "use default, don't show in mnemonic".
	if bc == 0 {
		bc = entities.DefaultBarComponent
	}

	if qc == 0 {
		qc = entities.DefaultQuoteComponent
	}

	if tc == 0 {
		tc = entities.DefaultTradeComponent
	}

	if math.IsNaN(alpha) {
		if length < minlen {
			return nil, fmt.Errorf(fmts, invalid, "length should be positive")
		}

		alpha = two / float64(1+length)
		mnemonic = fmt.Sprintf(fmtnl, length, core.ComponentTripleMnemonic(bc, qc, tc))
	} else {
		if alpha < 0. || alpha > 1. {
			return nil, fmt.Errorf(fmts, invalid, "smoothing factor should be in range [0, 1]")
		}

		if alpha < epsilon {
			alpha = epsilon
		}

		length = int(math.Round(two/alpha)) - 1
		mnemonic = fmt.Sprintf(fmtna, length, alpha, core.ComponentTripleMnemonic(bc, qc, tc))
	}

	if barFunc, err = entities.BarComponentFunc(bc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	if quoteFunc, err = entities.QuoteComponentFunc(qc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	if tradeFunc, err = entities.TradeComponentFunc(tc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	desc := "Exponential moving average " + mnemonic

	ema := &ExponentialMovingAverage{
		smoothingFactor: alpha,
		length:          length,
		firstIsAverage:  firstIsAverage,
	}

	ema.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, ema.Update)

	return ema, nil
}

// IsPrimed indicates whether an indicator is primed.
func (s *ExponentialMovingAverage) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes an output data of the indicator.
// It always has a single scalar output -- the calculated value of the exponential moving average.
func (s *ExponentialMovingAverage) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.ExponentialMovingAverage,
		Mnemonic:    s.LineIndicator.Mnemonic,
		Description: s.LineIndicator.Description,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(ExponentialMovingAverageValue),
				Type:        outputs.ScalarType,
				Mnemonic:    s.LineIndicator.Mnemonic,
				Description: s.LineIndicator.Description,
			},
		},
	}
}

// Update updates the value of the exponential moving average given the next sample.
//
// The indicator is not primed during the first ℓ-1 updates.
func (s *ExponentialMovingAverage) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	temp := sample

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.primed { //nolint:nestif
		s.value += (temp - s.value) * s.smoothingFactor
	} else {
		s.count++
		if s.firstIsAverage {
			s.sum += temp
			if s.count < s.length {
				return math.NaN()
			}

			s.value = s.sum / float64(s.length)
		} else {
			if s.count == 1 {
				s.value = temp
			} else {
				s.value += (temp - s.value) * s.smoothingFactor
			}

			if s.count < s.length {
				return math.NaN()
			}
		}

		s.primed = true
	}

	return s.value
}
