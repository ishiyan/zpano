package coronatrendvigor

import "zpano/entities"

// Params describes parameters to create an instance of the CoronaTrendVigor indicator.
type Params struct {
	// RasterLength is the number of elements in the heatmap raster. The default value is 50.
	RasterLength int

	// MaxRasterValue is the maximal raster value (z) of the heatmap. The default value is 20.
	MaxRasterValue float64

	// MinParameterValue is the minimal ordinate (y) value of the heatmap. The default value is -10.
	MinParameterValue float64

	// MaxParameterValue is the maximal ordinate (y) value of the heatmap. The default value is 10.
	MaxParameterValue float64

	// HighPassFilterCutoff is the high-pass filter cutoff (de-trending period) used by
	// the inner Corona engine. Suggested values are 20, 30, 100. The default value is 30.
	HighPassFilterCutoff int

	// MinimalPeriod is the minimal period of the inner Corona engine. The default value is 6.
	MinimalPeriod int

	// MaximalPeriod is the maximal period of the inner Corona engine. The default value is 30.
	MaximalPeriod int

	// BarComponent indicates the component of a bar to use when updating the indicator with a bar sample.
	//
	// If zero, the default (BarMedianPrice) is used, matching Ehlers' reference which operates on
	// (High+Low)/2. Since this differs from the framework-wide default, it is always shown in the
	// indicator mnemonic.
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

// DefaultParams returns a [Params] value populated with Ehlers defaults.
func DefaultParams() *Params {
	return &Params{
		RasterLength:         50,
		MaxRasterValue:       20,
		MinParameterValue:    -10,
		MaxParameterValue:    10,
		HighPassFilterCutoff: 30,
		MinimalPeriod:        6,
		MaximalPeriod:        30,
	}
}
