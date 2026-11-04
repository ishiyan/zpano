package hodrickprescottsmoothinganddetrending

import (
	"strings"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
)

func TestComponentMnemonics(t *testing.T) {
	for _, components := range [][3]int{{0, 0, 0}, {1, 0, 0}, {0, 1, 0}, {0, 0, 2}, {1, 1, 0}, {1, 0, 2}, {0, 1, 2}, {1, 1, 2}} {
		p := DefaultParams()
		p.BarComponent = entities.BarComponent(components[0])
		p.QuoteComponent = entities.QuoteComponent(components[1])
		p.TradeComponent = entities.TradeComponent(components[2])
		indicator, err := NewHodrickPrescottSmoothingAndDetrending(p)
		if err != nil {
			t.Fatal(err)
		}
		bc, qc, tc := p.BarComponent, p.QuoteComponent, p.TradeComponent
		if bc == 0 {
			bc = entities.DefaultBarComponent
		}
		if qc == 0 {
			qc = entities.DefaultQuoteComponent
		}
		if tc == 0 {
			tc = entities.DefaultTradeComponent
		}
		suffix := core.ComponentTripleMnemonic(bc, qc, tc)
		if !strings.HasSuffix(indicator.Metadata().Mnemonic, suffix+")") {
			t.Fatal(indicator.Metadata().Mnemonic, suffix)
		}
		if indicator.IsPrimed() {
			t.Fatal("must begin unprimed")
		}
		indicator.Reset()
		if indicator.IsPrimed() {
			t.Fatal("reset must clear priming")
		}
	}
}
