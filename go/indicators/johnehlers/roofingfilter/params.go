package roofingfilter

import "zpano/entities"

// RoofingFilterParams describes parameters to create an instance of the indicator.
type RoofingFilterParams struct {
	// ShortestCyclePeriod is the shortest cycle period in bars.
	// The Roofing Filter attenuates all cycle periods shorter than this one.
	//
	// The value should be greater than 1. The default value is 10.
	ShortestCyclePeriod int

	// LongestCyclePeriod is the longest cycle period in bars.
	// The Roofing Filter attenuates all cycle periods longer than this one.
	//
	// The value should be greater than ShortestCyclePeriod. The default value is 48.
	LongestCyclePeriod int

	// HasTwoPoleHighpassFilter indicates whether to use a two-pole high-pass filter
	// instead of the default one-pole high-pass filter.
	HasTwoPoleHighpassFilter bool

	// HasZeroMean indicates whether to apply a zero-mean filter after the super smoother.
	// Only applicable when HasTwoPoleHighpassFilter is false.
	HasZeroMean bool

	// BarComponent indicates the component of a bar to use when updating the indicator with a bar sample.
	//
	// If zero, the default (BarMedianPrice) is used and the component is not shown in the indicator mnemonic.
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

// DefaultParams returns a [RoofingFilterParams] value populated with Ehlers defaults.
func DefaultParams() *RoofingFilterParams {
	return &RoofingFilterParams{
		ShortestCyclePeriod: 10,
		LongestCyclePeriod:  48,
	}
}
