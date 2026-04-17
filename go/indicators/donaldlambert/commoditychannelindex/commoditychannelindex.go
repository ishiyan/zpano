package commoditychannelindex

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
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
func (c *CommodityChannelIndex) IsPrimed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.primed
}

// Metadata describes the output data of the indicator.
func (c *CommodityChannelIndex) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.CommodityChannelIndex,
		Mnemonic:    c.LineIndicator.Mnemonic,
		Description: c.LineIndicator.Description,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(CommodityChannelIndexValue),
				Type:        outputs.ScalarType,
				Mnemonic:    c.LineIndicator.Mnemonic,
				Description: c.LineIndicator.Description,
			},
		},
	}
}

// Update updates the value of the indicator given the next sample.
func (c *CommodityChannelIndex) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	lastIndex := c.length - 1

	if c.primed {
		c.windowSum += sample - c.window[0]
		copy(c.window, c.window[1:])
		c.window[lastIndex] = sample

		average := c.windowSum / float64(c.length)

		var temp float64
		for i := 0; i < c.length; i++ {
			temp += math.Abs(c.window[i] - average)
		}

		if math.Abs(temp) < math.SmallestNonzeroFloat64 {
			c.value = 0
		} else {
			c.value = c.scalingFactor * (sample - average) / temp
		}
	} else {
		c.windowSum += sample
		c.window[c.windowCount] = sample
		c.windowCount++

		if c.windowCount == c.length {
			c.primed = true

			average := c.windowSum / float64(c.length)

			var temp float64
			for i := 0; i < c.length; i++ {
				temp += math.Abs(c.window[i] - average)
			}

			if math.Abs(temp) < math.SmallestNonzeroFloat64 {
				c.value = 0
			} else {
				c.value = c.scalingFactor * (sample - average) / temp
			}
		}
	}

	return c.value
}
