package core

import "zpano/entities"

// BarComponentMnemonic returns a short mnemonic for the given bar component.
func BarComponentMnemonic(c entities.BarComponent) string {
	return c.Mnemonic()
}

// QuoteComponentMnemonic returns a short mnemonic for the given quote component.
func QuoteComponentMnemonic(c entities.QuoteComponent) string {
	return c.Mnemonic()
}

// TradeComponentMnemonic returns a short mnemonic for the given trade component.
func TradeComponentMnemonic(c entities.TradeComponent) string {
	return c.Mnemonic()
}

// ComponentTripleMnemonic builds a mnemonic suffix string from bar, quote and trade
// components. A component equal to its default value is omitted from the mnemonic.
//
// For example, if bar component is BarMedianPrice (non-default), the result is ", hl/2".
// If bar component is BarClosePrice (default), it is omitted.
func ComponentTripleMnemonic(
	bc entities.BarComponent, qc entities.QuoteComponent, tc entities.TradeComponent,
) string {
	var s string

	if bc != entities.DefaultBarComponent {
		s += ", " + bc.Mnemonic()
	}

	if qc != entities.DefaultQuoteComponent {
		s += ", " + qc.Mnemonic()
	}

	if tc != entities.DefaultTradeComponent {
		s += ", " + tc.Mnemonic()
	}

	return s
}
