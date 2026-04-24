package commoditychannelindex

import "zpano/entities"

// DefaultInverseScalingFactor is the default inverse scaling factor.
// The value of 0.015 ensures that approximately 70 to 80 percent of CCI values fall between -100 and +100.
const DefaultInverseScalingFactor = 0.015

// CommodityChannelIndexParams describes parameters to create an instance of the indicator.
type CommodityChannelIndexParams struct {
	// Length is the number of time periods of the commodity channel index.
	//
	// The value should be greater than 1.
	Length int

	// InverseScalingFactor is the factor to provide more readable value numbers.
	// The default value of 0.015 ensures that approximately 70 to 80 percent of CCI values
	// would fall between -100 and +100.
	//
	// If zero, the default (0.015) is used.
	InverseScalingFactor float64

	// BarComponent indicates the component of a bar to use when updating the indicator with a bar sample.
	//
	// If zero, the default (BarClosePrice) is used and the component is not shown in the indicator mnemonic.
	BarComponent entities.BarComponent

	// QuoteComponent indicates the component of a quote to use when updating the indicator with a quote sample.
	//
	// If zero, the default (QuoteMidPrice) is used and the component is not shown in the indicator mnemonic.
	QuoteComponent entities.QuoteComponent

	// TradeComponent indicates the component of a trade to use when updating the indicator with a trade sample.
	//
	// If zero, the default (TradePrice) is used and the component is not shown in the indicator mnemonic.
	TradeComponent entities.TradeComponent
}

// DefaultParams returns a [CommodityChannelIndexParams] value populated with conventional defaults.
func DefaultParams() *CommodityChannelIndexParams {
	return &CommodityChannelIndexParams{
		Length:               20,
		InverseScalingFactor: DefaultInverseScalingFactor,
	}
}
