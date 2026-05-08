package jurikwaveletsampler

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// nmEntry holds the (n, M) parameters for each wavelet column.
type nmEntry struct {
	n int
	m int
}

// nmTable defines the (n, M) parameters for columns 0..17.
var nmTable = []nmEntry{
	{1, 0}, {2, 0}, {3, 0}, {4, 0}, {5, 0},
	{7, 2}, {10, 2}, {14, 4}, {19, 4}, {26, 8},
	{35, 8}, {48, 16}, {65, 16}, {90, 32}, {123, 32},
	{172, 64}, {237, 64}, {334, 128},
}

// JurikWaveletSampler computes the Jurik wavelet sampler.
// Produces `index` output columns per bar, each representing a different
// multi-resolution scale. The framework output is the first column value.
type JurikWaveletSampler struct {
	mu sync.RWMutex
	core.LineIndicator
	primed      bool
	index       int
	maxLookback int
	prices      []float64
	barCount    int
	columns     []float64
}

// NewJurikWaveletSampler returns an instance of the indicator.
func NewJurikWaveletSampler(p *JurikWaveletSamplerParams) (*JurikWaveletSampler, error) {
	return newJurikWaveletSampler(p.Index, p.BarComponent, p.QuoteComponent, p.TradeComponent)
}

func newJurikWaveletSampler(index int,
	bc entities.BarComponent, qc entities.QuoteComponent, tc entities.TradeComponent,
) (*JurikWaveletSampler, error) {
	const (
		invalid = "invalid jurik wavelet sampler parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
fmtn = "jwav(%d%s)"
	)

	var (
		err       error
		barFunc   entities.BarFunc
		quoteFunc entities.QuoteFunc
		tradeFunc entities.TradeFunc
	)

	if index < 1 || index > 18 {
		return nil, fmt.Errorf(fmts, invalid, "index must be in range [1, 18]")
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

	// Compute max lookback.
	maxLookback := 0

	for c := 0; c < index; c++ {
		lb := nmTable[c].n + nmTable[c].m/2
		if lb > maxLookback {
			maxLookback = lb
		}
	}

	mnemonic := fmt.Sprintf(fmtn, index, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Jurik wavelet sampler " + mnemonic

	ind := &JurikWaveletSampler{
		index:       index,
		maxLookback: maxLookback,
		columns:     make([]float64, index),
	}

	ind.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, ind.Update)

	return ind, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *JurikWaveletSampler) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *JurikWaveletSampler) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.JurikWaveletSampler,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Columns returns the current column values after the last update.
func (s *JurikWaveletSampler) Columns() []float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]float64, len(s.columns))
	copy(result, s.columns)

	return result
}

// Update updates the indicator given the next sample.
func (s *JurikWaveletSampler) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.prices = append(s.prices, sample)
	s.barCount++

	allValid := true

	for c := 0; c < s.index; c++ {
		n := nmTable[c].n
		m := nmTable[c].m
		deadZone := n + m/2

		if s.barCount <= deadZone {
			s.columns[c] = math.NaN()
			allValid = false
		} else {
			if m == 0 {
				// Simple lag.
				s.columns[c] = s.prices[s.barCount-1-n]
			} else {
				// Mean of (M+1) prices centered at lag n.
				half := m / 2
				centerIdx := s.barCount - 1 - n
				total := 0.0

				for k := centerIdx - half; k <= centerIdx+half; k++ {
					total += s.prices[k]
				}

				s.columns[c] = total / float64(m+1)
			}
		}
	}

	if allValid {
		s.primed = true
	}

	// Return first column as the framework output.
	return s.columns[0]
}
