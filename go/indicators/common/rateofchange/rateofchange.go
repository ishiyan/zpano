package rateofchange

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
)

// RateOfChange is the difference between today's sample and the sample ℓ periods ago
// scaled by the old sample so as to represent the increase as a fraction.
//
// The values are centered at zero and can be positive and negative.
//
// ROCᵢ = 100 (Pᵢ - Pᵢ₋ℓ) / Pᵢ₋ℓ = 100 (Pᵢ/Pᵢ₋ℓ -1),
//
// where ℓ is the length.
//
// The indicator is not primed during the first ℓ updates.
type RateOfChange struct {
	mu sync.RWMutex
	core.LineIndicator
	window       []float64
	windowLength int
	windowCount  int
	lastIndex    int
	primed       bool
}

// New returns an instance of the indicator created using supplied parameters.
func New(p *Params) (*RateOfChange, error) {
	const (
		invalid = "invalid rate of change parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "roc(%d%s)"
		minlen  = 1
	)

	length := p.Length
	if length < minlen {
		return nil, fmt.Errorf(fmts, invalid, "length should be positive")
	}

	var (
		err       error
		barFunc   entities.BarFunc
		quoteFunc entities.QuoteFunc
		tradeFunc entities.TradeFunc
	)

	// Resolve defaults for component functions.
	// A zero value means "use default, don't show in mnemonic".
	bc := p.BarComponent
	qc := p.QuoteComponent
	tc := p.TradeComponent

	if bc == 0 {
		bc = entities.DefaultBarComponent
	}

	if qc == 0 {
		qc = entities.DefaultQuoteComponent
	}

	if tc == 0 {
		tc = entities.DefaultTradeComponent
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

	mnemonic := fmt.Sprintf(fmtn, length, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Rate of Change " + mnemonic

	r := &RateOfChange{
		window:       make([]float64, length+1),
		windowLength: length + 1,
		lastIndex:    length,
	}

	r.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, r.Update)

	return r, nil
}

// IsPrimed indicates whether an indicator is primed.
func (s *RateOfChange) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes an output data of the indicator.
// It always has a single scalar output -- the calculated value of the rate of change.
func (s *RateOfChange) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.RateOfChange,
		Mnemonic:    s.LineIndicator.Mnemonic,
		Description: s.LineIndicator.Description,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(Value),
				Type:        outputs.ScalarType,
				Mnemonic:    s.LineIndicator.Mnemonic,
				Description: s.LineIndicator.Description,
			},
		},
	}
}

// Update updates the value of the indicator given the next sample.
//
// The indicator is not primed during the first ℓ updates.
func (s *RateOfChange) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	const epsilon = 1e-13
	const c100 = 100

	if s.primed {
		if s.lastIndex > 1 {
			for i := 0; i < s.lastIndex; i++ {
				s.window[i] = s.window[i+1]
			}
		}

		s.window[s.lastIndex] = sample
		previous := s.window[0]
		if math.Abs(previous) > epsilon {
			return (sample/previous - 1) * c100
		}

		return 0
	}

	s.window[s.windowCount] = sample
	s.windowCount++

	if s.windowLength == s.windowCount {
		s.primed = true
		previous := s.window[0]
		if math.Abs(previous) > epsilon {
			return (sample/previous - 1) * c100
		}

		return 0
	}

	return math.NaN()
}
