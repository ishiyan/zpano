package moneyflowindex

import "zpano/entities"

// MoneyFlowIndexParams describes parameters to create an instance of the indicator.
type MoneyFlowIndexParams struct {
	// Length is the number of time periods of the Money Flow Index.
	//
	// The value should be greater than 0. The default value is 14.
	Length int

	// BarComponent indicates the component of a bar to use when updating the indicator with a bar sample.
	//
	// If zero, the default (BarTypicalPrice) is used.
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

// DefaultParams returns a [MoneyFlowIndexParams] value populated with conventional defaults.
func DefaultParams() *MoneyFlowIndexParams {
	return &MoneyFlowIndexParams{
		Length: 14,
	}
}
