package coronaspectrum

import "zpano/entities"

// Params describes parameters to create an instance of the CoronaSpectrum indicator.
type Params struct {
	// MinRasterValue is the minimal raster value (z) of the heatmap, in decibels.
	// Corresponds to the CoronaLowerDecibels threshold. The default value is 6.
	MinRasterValue float64

	// MaxRasterValue is the maximal raster value (z) of the heatmap, in decibels.
	// Corresponds to the CoronaUpperDecibels threshold. The default value is 20.
	MaxRasterValue float64

	// MinParameterValue is the minimal ordinate (y) value of the heatmap, representing
	// the minimal cycle period covered by the filter bank. The default value is 6.
	MinParameterValue float64

	// MaxParameterValue is the maximal ordinate (y) value of the heatmap, representing
	// the maximal cycle period covered by the filter bank. The default value is 30.
	MaxParameterValue float64

	// HighPassFilterCutoff is the high-pass filter cutoff (de-trending period) used by
	// the inner Corona engine. Suggested values are 20, 30, 100. The default value is 30.
	HighPassFilterCutoff int

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
