package jurikturningpointoscillator

import (
	"fmt"
	"math"
	"sort"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// JurikTurningPointOscillator computes Spearman rank correlation between
// price ranks and time positions. Output is in [-1, +1].
type JurikTurningPointOscillator struct {
	mu sync.RWMutex
	core.LineIndicator
	primed bool
	length int
	buffer []float64
	bufIdx int
	count  int
	f18    float64
	mid    float64
}

// NewJurikTurningPointOscillator returns an instance of the indicator created using supplied parameters.
func NewJurikTurningPointOscillator(p *JurikTurningPointOscillatorParams) (*JurikTurningPointOscillator, error) {
	return newJurikTurningPointOscillator(p.Length, p.BarComponent, p.QuoteComponent, p.TradeComponent)
}

func newJurikTurningPointOscillator(length int,
	bc entities.BarComponent, qc entities.QuoteComponent, tc entities.TradeComponent,
) (*JurikTurningPointOscillator, error) {
	const (
		invalid = "invalid jurik turning point oscillator parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "jtpo(%d%s)"
	)

	var (
		err       error
		barFunc   entities.BarFunc
		quoteFunc entities.QuoteFunc
		tradeFunc entities.TradeFunc
	)

	if length < 2 {
		return nil, fmt.Errorf(fmts, invalid, "length should be at least 2")
	}

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
	desc := "Jurik turning point oscillator " + mnemonic

	n := float64(length)
	ind := &JurikTurningPointOscillator{
		length: length,
		buffer: make([]float64, length),
		f18:    12.0 / (n * (n - 1) * (n + 1)),
		mid:    (n + 1) / 2.0,
	}

	ind.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, ind.Update)

	return ind, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *JurikTurningPointOscillator) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *JurikTurningPointOscillator) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.JurikTurningPointOscillator,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the indicator given the next sample.
func (s *JurikTurningPointOscillator) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	length := s.length

	s.buffer[s.bufIdx] = sample
	s.bufIdx = (s.bufIdx + 1) % length
	s.count++

	if s.count < length {
		return math.NaN()
	}

	// Extract window in chronological order.
	window := make([]float64, length)
	for i := 0; i < length; i++ {
		window[i] = s.buffer[(s.bufIdx+i)%length]
	}

	// Check if all values are identical.
	allSame := true
	for i := 1; i < length; i++ {
		if window[i] != window[0] {
			allSame = false

			break
		}
	}

	if allSame {
		if !s.primed {
			s.primed = true
		}

		return math.NaN()
	}

	// Build indices sorted by price.
	type indexedPrice struct {
		idx   int
		price float64
	}

	items := make([]indexedPrice, length)
	for i := 0; i < length; i++ {
		items[i] = indexedPrice{idx: i, price: window[i]}
	}

	sort.SliceStable(items, func(a, b int) bool {
		return items[a].price < items[b].price
	})

	// arr2[i] = original time position (1-based) of the i-th sorted element.
	arr2 := make([]float64, length)
	for i := 0; i < length; i++ {
		arr2[i] = float64(items[i].idx + 1)
	}

	// Assign fractional ranks for ties.
	arr3 := make([]float64, length)
	i := 0

	for i < length {
		j := i
		for j < length-1 && items[j+1].price == items[j].price {
			j++
		}

		avgRank := float64(i+1+j+1) / 2.0
		for k := i; k <= j; k++ {
			arr3[k] = avgRank
		}

		i = j + 1
	}

	// Compute correlation sum.
	var corrSum float64

	for i := 0; i < length; i++ {
		corrSum += (arr3[i] - s.mid) * (arr2[i] - s.mid)
	}

	if !s.primed {
		s.primed = true
	}

	return s.f18 * corrSum
}
