package chandemomentumoscillator

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

const epsilon = 1e-12

// ChandeMomentumOscillator is a momentum indicator based on the average
// of up samples and down samples over a specified length ℓ.
//
// The calculation formula is:
//
// CMOᵢ = 100 (SUᵢ-SDᵢ) / (SUᵢ + SDᵢ),
//
// where SUᵢ (sum up) is the sum of gains and SDᵢ (sum down)
// is the sum of losses over the chosen length [i-ℓ, i].
//
// The indicator is not primed during the first ℓ updates.
type ChandeMomentumOscillator struct {
	mu sync.RWMutex
	core.LineIndicator
	length         int
	count          int
	ringBuffer     []float64
	ringHead       int
	previousSample float64
	gainSum        float64
	lossSum        float64
	primed         bool
}

// New returns an instance of the indicator created using supplied parameters.
func New(p *Params) (*ChandeMomentumOscillator, error) {
	const (
		invalid = "invalid Chande momentum oscillator parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "cmo(%d%s)"
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
	desc := "Chande Momentum Oscillator " + mnemonic

	c := &ChandeMomentumOscillator{
		length:     length,
		ringBuffer: make([]float64, length),
	}

	c.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, c.Update)

	return c, nil
}

// IsPrimed indicates whether an indicator is primed.
func (s *ChandeMomentumOscillator) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes an output data of the indicator.
// It always has a single scalar output -- the calculated value of the indicator.
func (s *ChandeMomentumOscillator) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.ChandeMomentumOscillator,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the value of the Chande momentum oscillator given the next sample.
//
// The indicator is not primed during the first ℓ updates.
func (s *ChandeMomentumOscillator) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.count++
	if s.count == 1 {
		s.previousSample = sample

		return math.NaN()
	}

	// New delta
	delta := sample - s.previousSample
	s.previousSample = sample

	if !s.primed {
		// Fill until we have s.length deltas (i.e., s.length+1 samples)
		s.ringBuffer[s.ringHead] = delta
		s.ringHead = (s.ringHead + 1) % s.length

		if delta > 0 {
			s.gainSum += delta
		} else if delta < 0 {
			s.lossSum += -delta
		}

		if s.count <= s.length {
			return math.NaN()
		}

		// Now we have exactly s.length deltas in the buffer
		s.primed = true
	} else {
		// Remove oldest delta and add the new one
		old := s.ringBuffer[s.ringHead]
		if old > 0 {
			s.gainSum -= old
		} else if old < 0 {
			s.lossSum -= -old
		}

		s.ringBuffer[s.ringHead] = delta
		s.ringHead = (s.ringHead + 1) % s.length

		if delta > 0 {
			s.gainSum += delta
		} else if delta < 0 {
			s.lossSum += -delta
		}

		// Clamp to avoid tiny negative sums from FP noise
		if s.gainSum < 0 {
			s.gainSum = 0
		}

		if s.lossSum < 0 {
			s.lossSum = 0
		}
	}

	den := s.gainSum + s.lossSum
	if math.Abs(den) < epsilon {
		return 0
	}

	return 100.0 * (s.gainSum - s.lossSum) / den
}
