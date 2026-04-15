package kaufmanadaptivemovingaverage

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
)

// KaufmanAdaptiveMovingAverage (Kaufman's adaptive moving average, KAMA) is an EMA with the smoothing
// factor, α, being changed with each new sample within the fastest and the slowest boundaries:
//
// KAMAᵢ = αPᵢ + (1 - α)*KAMAᵢ₋₁,  α = (αs + (αf - αs)ε)²
//
// where the αf is the α of the fastest (shortest, default 2 samples) period boundary,
// the αs is the α of the slowest (longest, default 30 samples) period boundary,
// and ε is the efficiency ratio:
//
// ε = |P - Pℓ| / ∑|Pᵢ - Pᵢ₊₁|,  i ≤ ℓ-1
//
// where ℓ is a number of samples used to calculate the ε.
// The recommended values of ℓ are in the range of 8 to 10.
//
// The efficiency ratio has the value of 1 when samples move in the same direction for
// the full ℓ periods, and a value of 0 when samples are unchanged over the ℓ periods.
// When samples move in wide swings within the interval, the sum of the denominator
// becomes very large compared with the numerator and the ε approaches 0.
// Smaller values of ε result in a smaller smoothing constant and a slower trend.
//
// The indicator is not primed during the first ℓ updates.
//
// Reference:
// Perry J. Kaufman, Smarter Trading, McGraw-Hill, Ney York, 1995, pp. 129-153.
type KaufmanAdaptiveMovingAverage struct {
	mu sync.RWMutex
	core.LineIndicator
	efficiencyRatioLength int
	windowCount           int
	window                []float64
	absoluteDelta         []float64
	absoluteDeltaSum      float64
	alphaFastest          float64
	alphaSlowest          float64
	alphaDiff             float64
	value                 float64
	efficiencyRatio       float64
	primed                bool
}

// NewKaufmanAdaptiveMovingAverageLength returns an instnce of the indicator
// created using supplied parameters based on length.
func NewKaufmanAdaptiveMovingAverageLength(
	p *KaufmanAdaptiveMovingAverageLengthParams,
) (*KaufmanAdaptiveMovingAverage, error) {
	return newKaufmanAdaptiveMovingAverage(p.EfficiencyRatioLength,
		p.FastestLength, p.SlowestLength,
		math.NaN(), math.NaN(),
		p.BarComponent, p.QuoteComponent, p.TradeComponent)
}

// NewKaufmanAdaptiveMovingAverageSmoothingFactor returns an instnce of the indicator
// created using supplied parameters based on smoothing factor.
func NewKaufmanAdaptiveMovingAverageSmoothingFactor(
	p *KaufmanAdaptiveMovingAverageSmoothingFactorParams,
) (*KaufmanAdaptiveMovingAverage, error) {
	return newKaufmanAdaptiveMovingAverage(p.EfficiencyRatioLength,
		0, 0,
		p.FastestSmoothingFactor, p.SlowestSmoothingFactor,
		p.BarComponent, p.QuoteComponent, p.TradeComponent)
}

//nolint:funlen,cyclop
func newKaufmanAdaptiveMovingAverage(efficiencyRatioLength int,
	fastestSmoothingLength int, slowestSmoothingLength int,
	fastestSmoothingFactor float64, slowestSmoothingFactor float64,
	bc entities.BarComponent, qc entities.QuoteComponent, tc entities.TradeComponent,
) (*KaufmanAdaptiveMovingAverage, error) {
	const (
		invalid = "invalid Kaufman adaptive moving average parameters"
		fmtl    = "%s: %s length should be larger than 1"
		fmta    = "%s: %s smoothing factor should be in range [0, 1]"
		fmtw    = "%s: %w"
		fmtnl   = "kama(%d, %d, %d%s)"
		fmtna   = "kama(%d, %.4f, %.4f%s)"
		two     = 2
		epsilon = 0.00000001
	)

	var (
		mnemonic  string
		err       error
		barFunc   entities.BarFunc
		quoteFunc entities.QuoteFunc
		tradeFunc entities.TradeFunc
	)

	if efficiencyRatioLength < two {
		return nil, fmt.Errorf(fmtl, invalid, "efficiency ratio")
	}

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

	if math.IsNaN(fastestSmoothingFactor) { //nolint:nestif
		if fastestSmoothingLength < two {
			return nil, fmt.Errorf(fmtl, invalid, "fastest smoothing")
		}

		if slowestSmoothingLength < two {
			return nil, fmt.Errorf(fmtl, invalid, "slowest smoothing")
		}

		fastestSmoothingFactor = float64(two) / float64(1+fastestSmoothingLength)
		slowestSmoothingFactor = float64(two) / float64(1+slowestSmoothingLength)

		mnemonic = fmt.Sprintf(fmtnl, efficiencyRatioLength,
			fastestSmoothingLength, slowestSmoothingLength,
			core.ComponentTripleMnemonic(bc, qc, tc))
	} else {
		if fastestSmoothingFactor < 0. || fastestSmoothingFactor > 1. {
			return nil, fmt.Errorf(fmta, invalid, "fastest")
		}

		if slowestSmoothingFactor < 0. || slowestSmoothingFactor > 1. {
			return nil, fmt.Errorf(fmta, invalid, "slowest")
		}

		if fastestSmoothingFactor < epsilon {
			fastestSmoothingFactor = epsilon
		}

		if slowestSmoothingFactor < epsilon {
			slowestSmoothingFactor = epsilon
		}

		mnemonic = fmt.Sprintf(fmtna, efficiencyRatioLength,
			fastestSmoothingFactor, slowestSmoothingFactor,
			core.ComponentTripleMnemonic(bc, qc, tc))
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

	desc := "Kaufman adaptive moving average " + mnemonic

	kama := &KaufmanAdaptiveMovingAverage{
		efficiencyRatioLength: efficiencyRatioLength,
		alphaFastest:          fastestSmoothingFactor,
		alphaSlowest:          slowestSmoothingFactor,
		alphaDiff:             fastestSmoothingFactor - slowestSmoothingFactor,
		value:                 math.NaN(),
		efficiencyRatio:       math.NaN(),

		// These slices will be automatically filled with zeroes.
		window:        make([]float64, efficiencyRatioLength+1),
		absoluteDelta: make([]float64, efficiencyRatioLength+1),
	}

	kama.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, kama.Update)

	return kama, nil
}

// IsPrimed indicates whether an indicator is primed.
func (s *KaufmanAdaptiveMovingAverage) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes an output data of the indicator.
// It always has a single scalar output -- the calculated value of the moving average.
func (s *KaufmanAdaptiveMovingAverage) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.KaufmanAdaptiveMovingAverage,
		Mnemonic:    s.LineIndicator.Mnemonic,
		Description: s.LineIndicator.Description,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(KaufmanAdaptiveMovingAverageValue),
				Type:        outputs.ScalarType,
				Mnemonic:    s.LineIndicator.Mnemonic,
				Description: s.LineIndicator.Description,
			},
		},
	}
}

// Update updates the value of the moving average given the next sample.
func (s *KaufmanAdaptiveMovingAverage) Update(sample float64) float64 { //nolint:funlen
	if math.IsNaN(sample) {
		return sample
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	const epsilon = 0.00000001

	var temp float64

	if s.primed { //nolint:nestif
		temp = math.Abs(sample - s.window[s.efficiencyRatioLength])
		s.absoluteDeltaSum += temp - s.absoluteDelta[1]

		for i := range s.efficiencyRatioLength {
			j := i + 1
			s.window[i] = s.window[j]
			s.absoluteDelta[i] = s.absoluteDelta[j]
		}

		s.window[s.efficiencyRatioLength] = sample
		s.absoluteDelta[s.efficiencyRatioLength] = temp
		delta := math.Abs(sample - s.window[0])

		if s.absoluteDeltaSum <= delta || s.absoluteDeltaSum < epsilon {
			temp = 1.0
		} else {
			temp = delta / s.absoluteDeltaSum
		}

		s.efficiencyRatio = temp
		temp = s.alphaSlowest + temp*s.alphaDiff
		s.value += (sample - s.value) * temp * temp

		return s.value
	} else {
		s.window[s.windowCount] = sample
		if 0 < s.windowCount {
			temp = math.Abs(sample - s.window[s.windowCount-1])
			s.absoluteDelta[s.windowCount] = temp
			s.absoluteDeltaSum += temp
		}

		if s.efficiencyRatioLength == s.windowCount {
			s.primed = true
			delta := math.Abs(sample - s.window[0])

			if s.absoluteDeltaSum <= delta || s.absoluteDeltaSum < epsilon {
				temp = 1.0
			} else {
				temp = delta / s.absoluteDeltaSum
			}

			s.efficiencyRatio = temp
			temp = s.alphaSlowest + temp*s.alphaDiff
			s.value = s.window[s.efficiencyRatioLength-1]
			s.value += (sample - s.value) * temp * temp

			return s.value
		} else {
			s.windowCount++
		}
	}

	return math.NaN()
}
