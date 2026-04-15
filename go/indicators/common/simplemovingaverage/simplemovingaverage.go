package simplemovingaverage

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
)

// SimpleMovingAverage computes the simple, or arithmetic, moving average (SMA) by adding the samples
// for a number of time periods (length, ℓ) and then dividing this total by the number of time periods.
//
// In other words, this is an unweighted mean (gives equal weight to each sample) of the previous ℓ samples.
//
// This implementation updates the value of the SMA incrementally using the formula:
//
//	SMAᵢ = SMAᵢ₋₁ + (Pᵢ - Pᵢ₋ℓ) / ℓ,
//
// where ℓ is the length.
//
// The indicator is not primed during the first ℓ-1 updates.
type SimpleMovingAverage struct {
	mu sync.RWMutex
	core.LineIndicator
	window       []float64
	windowSum    float64
	windowLength int
	windowCount  int
	lastIndex    int
	primed       bool
}

// NewSimpleMovingAverage returns an instnce of the indicator created using supplied parameters.
func NewSimpleMovingAverage(p *SimpleMovingAverageParams) (*SimpleMovingAverage, error) {
	const (
		invalid = "invalid simple moving average parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "sma(%d%s)"
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
	desc := "Simple moving average " + mnemonic

	sma := &SimpleMovingAverage{
		window:       make([]float64, length),
		windowLength: length,
		lastIndex:    length - 1,
	}

	sma.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, sma.Update)

	return sma, nil
}

// IsPrimed indicates whether an indicator is primed.
func (s *SimpleMovingAverage) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes an output data of the indicator.
// It always has a single scalar output -- the calculated value of the simple moving average.
func (s *SimpleMovingAverage) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.SimpleMovingAverage,
		Mnemonic:    s.LineIndicator.Mnemonic,
		Description: s.LineIndicator.Description,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(SimpleMovingAverageValue),
				Type:        outputs.ScalarType,
				Mnemonic:    s.LineIndicator.Mnemonic,
				Description: s.LineIndicator.Description,
			},
		},
	}
}

// Update updates the value of the simple moving average given the next sample.
//
// The indicator is not primed during the first ℓ-1 updates.
func (s *SimpleMovingAverage) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	temp := sample

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.primed {
		s.windowSum += temp - s.window[0]

		for i := 0; i < s.lastIndex; i++ {
			s.window[i] = s.window[i+1]
		}

		s.window[s.lastIndex] = temp
	} else {
		s.windowSum += temp
		s.window[s.windowCount] = temp
		s.windowCount++

		if s.windowLength > s.windowCount {
			return math.NaN()
		}

		s.primed = true
	}

	return s.windowSum / float64(s.windowLength)
}
