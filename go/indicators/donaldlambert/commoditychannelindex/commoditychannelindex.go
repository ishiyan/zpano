package commoditychannelindex

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// CommodityChannelIndex is Donald Lambert's Commodity Channel Index (CCI).
//
// CCI measures the deviation of the price from its statistical mean. High values
// indicate that prices are unusually high compared to average, and low values
// indicate that prices are unusually low.
//
//	CCI = (typicalPrice - SMA) / (scalingFactor * meanDeviation)
//
// where scalingFactor defaults to 0.015 so that approximately 70-80% of CCI values
// fall between -100 and +100.
//
// Reference:
//
// Lambert, Donald (1980). "Commodity Channel Index: Tools for Trading Cyclic Trends".
// Commodities (now Futures) magazine.
type CommodityChannelIndex struct {
	mu sync.RWMutex
	core.LineIndicator
	length        int
	scalingFactor float64
	window        []float64
	windowCount   int
	windowSum     float64
	value         float64
	primed        bool
}

// NewCommodityChannelIndex returns an instance of the indicator created using supplied parameters.
func NewCommodityChannelIndex(p *CommodityChannelIndexParams) (*CommodityChannelIndex, error) {
	const (
		invalid   = "invalid commodity channel index parameters"
		fmts      = "%s: %s"
		fmtw      = "%s: %w"
		minLength = 2
	)

	if p.Length < minLength {
		return nil, fmt.Errorf(fmts, invalid, "length should be greater than 1")
	}

	inverseFactor := p.InverseScalingFactor
	if inverseFactor == 0 {
		inverseFactor = DefaultInverseScalingFactor
	}

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

	mnemonic := fmt.Sprintf("cci(%d%s)", p.Length,
		core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Commodity Channel Index " + mnemonic

	cci := &CommodityChannelIndex{
		length:        p.Length,
		scalingFactor: float64(p.Length) / inverseFactor,
		window:        make([]float64, p.Length),
		value:         math.NaN(),
	}

	cci.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, cci.Update)

	return cci, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *CommodityChannelIndex) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *CommodityChannelIndex) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.CommodityChannelIndex,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the value of the indicator given the next sample.
func (s *CommodityChannelIndex) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	lastIndex := s.length - 1

	if s.primed {
		s.windowSum += sample - s.window[0]
		copy(s.window, s.window[1:])
		s.window[lastIndex] = sample

		average := s.windowSum / float64(s.length)

		var temp float64
		for i := 0; i < s.length; i++ {
			temp += math.Abs(s.window[i] - average)
		}

		if math.Abs(temp) < math.SmallestNonzeroFloat64 {
			s.value = 0
		} else {
			s.value = s.scalingFactor * (sample - average) / temp
		}
	} else {
		s.windowSum += sample
		s.window[s.windowCount] = sample
		s.windowCount++

		if s.windowCount == s.length {
			s.primed = true

			average := s.windowSum / float64(s.length)

			var temp float64
			for i := 0; i < s.length; i++ {
				temp += math.Abs(s.window[i] - average)
			}

			if math.Abs(temp) < math.SmallestNonzeroFloat64 {
				s.value = 0
			} else {
				s.value = s.scalingFactor * (sample - average) / temp
			}
		}
	}

	return s.value
}
