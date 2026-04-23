package weightedmovingaverage

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// WeightedMovingAverage computes the weighted moving average (WMA) that has multiplying factors
// to give arithmetically decreasing weights to the samples in the look back window.
//
//	WMAᵢ = (ℓPᵢ + (ℓ-1)Pᵢ₋₁ + ... + Pᵢ₋ℓ) / (ℓ + (ℓ-1) + ... + 2 + 1),
//
// where ℓ is the length.
//
// The denominator is a triangle number and can be computed as
//
//	½ℓ(ℓ+1).
//
// When calculating the WMA across successive values,
//
//	WMAᵢ₊₁ - WMAᵢ = ℓPᵢ₊₁ - Pᵢ - ... - Pᵢ₋ℓ₊₁
//
// If we denote the sum
//
//	Totalᵢ = Pᵢ + ... + Pᵢ₋ℓ₊₁
//
// then
//
//	Totalᵢ₊₁ = Totalᵢ + Pᵢ₊₁ - Pᵢ₋ℓ₊₁
//	Numeratorᵢ₊₁ = Numeratorᵢ + ℓPᵢ₊₁ - Totalᵢ
//	WMAᵢ₊₁ = Numeratorᵢ₊₁ / ½ℓ(ℓ+1)
//
// The WMA indicator is not primed during the first ℓ-1 updates.
type WeightedMovingAverage struct {
	mu sync.RWMutex
	core.LineIndicator
	window       []float64
	windowSum    float64
	windowSub    float64
	divider      float64
	windowLength int
	windowCount  int
	lastIndex    int
	primed       bool
}

// NewWeightedMovingAverage returns an instnce of the indicator created using supplied parameters.
func NewWeightedMovingAverage(p *WeightedMovingAverageParams) (*WeightedMovingAverage, error) {
	const (
		invalid = "invalid weighted moving average parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "wma(%d%s)"
		minlen  = 2
	)

	length := p.Length
	if length < minlen {
		return nil, fmt.Errorf(fmts, invalid, "length should be greater than 1")
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
	mnemonic := fmt.Sprintf(fmtn, length, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Weighted moving average " + mnemonic
	divider := float64(length) * float64(length+1) / 2. //nolint:gomnd

	wma := &WeightedMovingAverage{
		window:       make([]float64, length),
		divider:      divider,
		windowLength: length,
		lastIndex:    length - 1,
	}

	wma.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, wma.Update)

	return wma, nil
}

// IsPrimed indicates whether an indicator is primed.
func (s *WeightedMovingAverage) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes an output data of the indicator.
// It always has a single scalar output -- the calculated value of the weighted moving average.
func (s *WeightedMovingAverage) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.WeightedMovingAverage,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the value of the moving average given the next sample.
//
// The indicator is not primed during the first ℓ-1 updates.
func (s *WeightedMovingAverage) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	temp := sample

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.primed {
		s.windowSum -= s.windowSub
		s.windowSum += temp * float64(s.windowLength)
		s.windowSub -= s.window[0]
		s.windowSub += temp

		for i := 0; i < s.lastIndex; i++ {
			s.window[i] = s.window[i+1]
		}

		s.window[s.lastIndex] = temp
	} else { // Not primed.
		s.window[s.windowCount] = temp
		s.windowSub += temp
		s.windowCount++
		s.windowSum += temp * float64(s.windowCount)

		if s.windowLength > s.windowCount {
			return math.NaN()
		}

		s.primed = true
	}

	return s.windowSum / s.divider
}
