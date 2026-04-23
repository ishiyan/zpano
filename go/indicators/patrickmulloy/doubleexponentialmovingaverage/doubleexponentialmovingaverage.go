package doubleexponentialmovingaverage

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// DoubleExponentialMovingAverage computes the Double Exponential Moving Average (DEMA),
// a smoothing indicator with less lag than a straight exponential moving average.
//
// The DEMA was developed by Patrick G. Mulloy and is described in two articles:
//
//	❶ Technical Analysis of Stocks & Commodities v.12:1 (11-19), Smoothing Data With Faster Moving Averages.
//	❷ Technical Analysis of Stocks & Commodities v.12:2 (72-80), Smoothing Data With Less Lag.
//
// The calculation is as follows:
//
//	EMA¹ᵢ = EMA(Pᵢ) = αPᵢ + (1-α)EMA¹ᵢ₋₁ = EMA¹ᵢ₋₁ + α(Pᵢ - EMA¹ᵢ₋₁), 0 < α ≤ 1
//	EMA²ᵢ = EMA(EMA¹ᵢ) = αEMA¹ᵢ + (1-α)EMA²ᵢ₋₁ = EMA²ᵢ₋₁ + α(EMA¹ᵢ - EMA²ᵢ₋₁), 0 < α ≤ 1
//	DEMAᵢ = 2 EMA¹ᵢ - EMA²ᵢ
//
// The very first EMA value (the seed for subsequent values) is calculated differently.
// This implementation allows for two algorithms for this seed.
//
//	❶ Use a simple average of the first 'period'. This is the most widely documented approach.
//	❷ Use first sample value as a seed. This is used in Metastock.
type DoubleExponentialMovingAverage struct {
	mu sync.RWMutex
	core.LineIndicator
	smoothingFactor float64
	sum             float64
	ema1            float64
	ema2            float64
	length          int
	length2         int
	count           int
	firstIsAverage  bool
	primed          bool
}

// NewDoubleExponentialMovingAverageLength returns an instnce of the indicator
// created using supplied parameters based on length.
func NewDoubleExponentialMovingAverageLength(
	p *DoubleExponentialMovingAverageLengthParams,
) (*DoubleExponentialMovingAverage, error) {
	return newDoubleExponentialMovingAverage(p.Length, math.NaN(), p.FirstIsAverage,
		p.BarComponent, p.QuoteComponent, p.TradeComponent)
}

// NewDoubleExponentialMovingAverageSmoothingFactor returns an instnce of the indicator
// created using supplied parameters based on smoothing factor.
func NewDoubleExponentialMovingAverageSmoothingFactor(
	p *DoubleExponentialMovingAverageSmoothingFactorParams,
) (*DoubleExponentialMovingAverage, error) {
	return newDoubleExponentialMovingAverage(0, p.SmoothingFactor, p.FirstIsAverage,
		p.BarComponent, p.QuoteComponent, p.TradeComponent)
}

//nolint:funlen,cyclop
func newDoubleExponentialMovingAverage(length int, alpha float64, firstIsAverage bool,
	bc entities.BarComponent, qc entities.QuoteComponent, tc entities.TradeComponent,
) (*DoubleExponentialMovingAverage, error) {
	const (
		invalid = "invalid double exponential moving average parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtnl   = "dema(%d%s)"
		fmtna   = "dema(%d, %.8f%s)"
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

	desc := "Double exponential moving average " + mnemonic

	dema := &DoubleExponentialMovingAverage{
		smoothingFactor: alpha,
		length:          length,
		length2:         int(two)*length - 1,
		firstIsAverage:  firstIsAverage,
	}

	dema.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, dema.Update)

	return dema, nil
}

// IsPrimed indicates whether an indicator is primed.
func (s *DoubleExponentialMovingAverage) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes an output data of the indicator.
// It always has a single scalar output -- the calculated value of the double exponential moving average.
func (s *DoubleExponentialMovingAverage) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.DoubleExponentialMovingAverage,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the value of the indicator given the next sample.
func (s *DoubleExponentialMovingAverage) Update(sample float64) float64 { //nolint:cyclop
	const two = 2.

	if math.IsNaN(sample) {
		return sample
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.primed {
		v1 := s.ema1
		v2 := s.ema2
		sf := s.smoothingFactor
		v1 += (sample - v1) * sf
		v2 += (v1 - v2) * sf
		s.ema1 = v1
		s.ema2 = v2

		return two*v1 - v2
	}

	s.count++
	if s.firstIsAverage { //nolint:nestif
		if s.count == 1 {
			s.sum = sample
		} else if s.length >= s.count {
			s.sum += sample
			if s.length == s.count {
				s.ema1 = s.sum / float64(s.length)
				s.sum = s.ema1
			}
		} else { // if s.length2 >= s.count {
			s.ema1 += (sample - s.ema1) * s.smoothingFactor
			s.sum += s.ema1

			if s.length2 == s.count {
				s.primed = true
				s.ema2 = s.sum / float64(s.length)

				return two*s.ema1 - s.ema2
			}
		}
	} else { // Metastock
		if s.count == 1 {
			s.ema1 = sample
		} else if s.length >= s.count {
			s.ema1 += (sample - s.ema1) * s.smoothingFactor
			if s.length == s.count {
				s.ema2 = s.ema1
			}
		} else { // if s.length2 >= s.count {
			s.ema1 += (sample - s.ema1) * s.smoothingFactor
			s.ema2 += (s.ema1 - s.ema2) * s.smoothingFactor

			if s.length2 == s.count {
				s.primed = true

				return two*s.ema1 - s.ema2
			}
		}
	}

	return math.NaN()
}
