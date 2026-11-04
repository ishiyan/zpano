package metonicrelativestrengthindex

import (
	"math"
	"strings"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
)

func TestComponentMnemonics(t *testing.T) {
	for _, components := range [][3]int{{0, 0, 0}, {1, 0, 0}, {0, 1, 0}, {0, 0, 2}, {1, 1, 0}, {1, 0, 2}, {0, 1, 2}, {1, 1, 2}} {
		p := DefaultParams()
		p.BarComponent = entities.BarComponent(components[0])
		p.QuoteComponent = entities.QuoteComponent(components[1])
		p.TradeComponent = entities.TradeComponent(components[2])
		indicator, err := NewMetonicRelativeStrengthIndex(p)
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

func TestProductionCalendarDelays(t *testing.T) {
	p := DefaultParams()
	p.Length = 1
	indicator, err := NewMetonicRelativeStrengthIndex(p)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 6945; i++ {
		actual := indicator.Update(testInput[i%len(testInput)], time.Time{}).Value
		if i <= 6939 {
			if !math.IsNaN(actual) {
				t.Fatal(i, actual)
			}
		} else {
			sum := 0.0
			for _, lag := range []int{6939, 4016, 2922} {
				if testInput[(i-lag)%len(testInput)] >= testInput[(i-lag-1)%len(testInput)] {
					sum += 100
				}
			}
			if math.Abs(actual-sum/3) > 1e-13 {
				t.Fatal(i, actual, sum/3)
			}
		}
	}
}
