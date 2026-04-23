package momentum

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// Momentum is the absolute (not normalized) difference between today's sample and the sample l periods ago.
//
// This implementation calculates the value of the MOM using the formula:
//
// MOMi = Pi - Pi-l,
//
// where l is the length.
//
// The indicator is not primed during the first l updates.
type Momentum struct {
	mu sync.RWMutex
	core.LineIndicator
	window       []float64
	windowLength int
	windowCount  int
	lastIndex    int
	primed       bool
}

// New returns an instance of the indicator created using supplied parameters.
func New(p *Params) (*Momentum, error) {
	const (
		invalid = "invalid momentum parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "mom(%d%s)"
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
	desc := "Momentum " + mnemonic

	m := &Momentum{
		window:       make([]float64, length+1),
		windowLength: length + 1,
		lastIndex:    length,
	}

	m.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, m.Update)

	return m, nil
}

// IsPrimed indicates whether an indicator is primed.
func (m *Momentum) IsPrimed() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.primed
}

// Metadata describes an output data of the indicator.
// It always has a single scalar output -- the calculated value of the momentum.
func (m *Momentum) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.Momentum,
		m.LineIndicator.Mnemonic,
		m.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: m.LineIndicator.Mnemonic, Description: m.LineIndicator.Description},
		},
	)
}

// Update updates the value of the momentum given the next sample.
//
// The indicator is not primed during the first l updates.
func (m *Momentum) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.primed {
		for i := 0; i < m.lastIndex; i++ {
			m.window[i] = m.window[i+1]
		}

		m.window[m.lastIndex] = sample

		return sample - m.window[0]
	}

	m.window[m.windowCount] = sample
	m.windowCount++

	if m.windowLength == m.windowCount {
		m.primed = true

		return sample - m.window[0]
	}

	return math.NaN()
}
