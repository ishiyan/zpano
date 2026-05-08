package jurikcommoditychannelindex

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/markjurik/jurikmovingaverage"
)

// JurikCommodityChannelIndex computes the Jurik Commodity Channel Index (JCCX).
// Uses fast JMA(4) and slow JMA(length), normalizes their difference by 1.5× MAD.
type JurikCommodityChannelIndex struct {
	mu sync.RWMutex
	core.LineIndicator
	primed      bool
	fastJMA     *jurikmovingaverage.JurikMovingAverage
	slowJMA     *jurikmovingaverage.JurikMovingAverage
	diffBuffer  []float64
	diffBufSize int
}

// NewJurikCommodityChannelIndex returns an instance of the indicator.
func NewJurikCommodityChannelIndex(p *JurikCommodityChannelIndexParams) (*JurikCommodityChannelIndex, error) {
	return newJurikCommodityChannelIndex(p.Length, p.BarComponent, p.QuoteComponent, p.TradeComponent)
}

func newJurikCommodityChannelIndex(length int,
	bc entities.BarComponent, qc entities.QuoteComponent, tc entities.TradeComponent,
) (*JurikCommodityChannelIndex, error) {
	const (
		invalid = "invalid jurik commodity channel index parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "jccx(%d%s)"
	)

	var (
		err       error
		barFunc   entities.BarFunc
		quoteFunc entities.QuoteFunc
		tradeFunc entities.TradeFunc
	)

	if length < 2 {
		return nil, fmt.Errorf(fmts, invalid, "length must be >= 2")
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
	desc := "Jurik commodity channel index " + mnemonic

	fastParams := &jurikmovingaverage.JurikMovingAverageParams{Length: 4, Phase: 0}
	slowParams := &jurikmovingaverage.JurikMovingAverageParams{Length: length, Phase: 0}

	fastJMA, err := jurikmovingaverage.NewJurikMovingAverage(fastParams)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	slowJMA, err := jurikmovingaverage.NewJurikMovingAverage(slowParams)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	ind := &JurikCommodityChannelIndex{
		fastJMA:     fastJMA,
		slowJMA:     slowJMA,
		diffBufSize: 3 * length,
	}

	ind.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, ind.Update)

	return ind, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *JurikCommodityChannelIndex) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *JurikCommodityChannelIndex) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.JurikCommodityChannelIndex,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the indicator given the next sample.
func (s *JurikCommodityChannelIndex) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	fastVal := s.fastJMA.Update(sample)
	slowVal := s.slowJMA.Update(sample)

	if math.IsNaN(fastVal) || math.IsNaN(slowVal) {
		return math.NaN()
	}

	diff := fastVal - slowVal

	s.diffBuffer = append(s.diffBuffer, diff)
	if len(s.diffBuffer) > s.diffBufSize {
		s.diffBuffer = s.diffBuffer[1:]
	}

	s.primed = true

	// Compute MAD.
	n := len(s.diffBuffer)
	var mad float64

	for _, d := range s.diffBuffer {
		mad += math.Abs(d)
	}

	mad /= float64(n)

	if mad < 0.00001 {
		return 0.0
	}

	return diff / (1.5 * mad)
}
