package onbalancevolume

import "zpano/entities"

// OnBalanceVolumeParams describes parameters to create an instance of the indicator.
type OnBalanceVolumeParams struct {
	// BarComponent indicates the component of a bar to use when updating the indicator with a bar sample.
	//
	// If zero, the default (BarClosePrice) is used.
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

// DefaultParams returns a [OnBalanceVolumeParams] value populated with conventional defaults.
func DefaultParams() *OnBalanceVolumeParams { return &OnBalanceVolumeParams{} }
