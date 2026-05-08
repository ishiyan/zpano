package jurikcommoditychannelindex

import "zpano/entities"

// JurikCommodityChannelIndexParams describes the parameters of the indicator.
type JurikCommodityChannelIndexParams struct {
	// Length controls the slow JMA length.
	Length int `json:"length"`

	// BarComponent specifies the bar component to use.
	BarComponent entities.BarComponent `json:"bar_component"`

	// QuoteComponent specifies the quote component to use.
	QuoteComponent entities.QuoteComponent `json:"quote_component"`

	// TradeComponent specifies the trade component to use.
	TradeComponent entities.TradeComponent `json:"trade_component"`
}

// DefaultParams returns default parameters for the indicator.
func DefaultParams() *JurikCommodityChannelIndexParams {
	return &JurikCommodityChannelIndexParams{
		Length: 20,
	}
}
