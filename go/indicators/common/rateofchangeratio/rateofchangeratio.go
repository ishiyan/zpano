package rateofchangeratio

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
)

// RateOfChangeRatio is the ratio of today's sample to the sample ℓ periods ago.
//
// The values are centered at 1 (or 100 when HundredScale is true) and are always positive.
//
// ROCRᵢ = Pᵢ / Pᵢ₋ℓ,
// ROCR100ᵢ = (Pᵢ / Pᵢ₋ℓ) * 100,
//
// where ℓ is the length.
//
// The indicator is not primed during the first ℓ updates.
type RateOfChangeRatio struct {
	mu sync.RWMutex
	core.LineIndicator
	window       []float64
	windowLength int
	windowCount  int
	lastIndex    int
	hundredScale bool
	primed       bool
}

// New returns an instance of the indicator created using supplied parameters.
func New(p *Params) (*RateOfChangeRatio, error) {
	const (
		invalid = "invalid rate of change ratio parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "rocr(%d%s)"
		fmtn100 = "rocr100(%d%s)"
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

	f := fmtn
	descPrefix := "Rate of Change Ratio "
	if p.HundredScale {
		f = fmtn100
		descPrefix = "Rate of Change Ratio 100 Scale "
	}

	mnemonic := fmt.Sprintf(f, length, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := descPrefix + mnemonic

	r := &RateOfChangeRatio{
		window:       make([]float64, length+1),
		windowLength: length + 1,
		lastIndex:    length,
		hundredScale: p.HundredScale,
	}

	r.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, r.Update)

	return r, nil
}

// IsPrimed indicates whether an indicator is primed.
func (s *RateOfChangeRatio) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes an output data of the indicator.
// It always has a single scalar output -- the calculated value of the rate of change ratio.
func (s *RateOfChangeRatio) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.RateOfChangeRatio,
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
func (s *RateOfChangeRatio) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	const epsilon = 1e-13

	scale := 1.0
	if s.hundredScale {
		scale = 100.0
	}

	if s.primed {
		if s.lastIndex > 1 {
			for i := 0; i < s.lastIndex; i++ {
				s.window[i] = s.window[i+1]
			}
		}

		s.window[s.lastIndex] = sample
		previous := s.window[0]
		if math.Abs(previous) > epsilon {
			return (sample / previous) * scale
		}

		return 0
	}

	s.window[s.windowCount] = sample
	s.windowCount++

	if s.windowLength == s.windowCount {
		s.primed = true
		previous := s.window[0]
		if math.Abs(previous) > epsilon {
			return (sample / previous) * scale
		}

		return 0
	}

	return math.NaN()
}
