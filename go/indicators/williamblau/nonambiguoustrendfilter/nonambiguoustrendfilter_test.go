//nolint:testpackage
package nonambiguoustrendfilter

import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

const tolerance = 1e-13

type ntfCombo struct {
	name     string
	base     Base
	q        int
	r        int
	s        int
	u        int
	expected []float64
}

// ntfCombos lists every expected array; q=0 where the base does not use it.
func ntfCombos() []ntfCombo {
	return []ntfCombo{
		{"TSI_R32_S13_U3", Tsi, 2, 32, 13, 3, expectedTSI_R32_S13_U3},
		{"TSI_R20_S5_U3", Tsi, 2, 20, 5, 3, expectedTSI_R20_S5_U3},
		{"TSI_R40_S20_U5", Tsi, 2, 40, 20, 5, expectedTSI_R40_S20_U5},
		{"SMI_Q32_R64_S7_U1", Smi, 32, 64, 7, 1, expectedSMI_Q32_R64_S7_U1},
		{"SMI_Q5_R20_S5_U3", Smi, 5, 20, 5, 3, expectedSMI_Q5_R20_S5_U3},
		{"SMI_Q13_R25_S2_U1", Smi, 13, 25, 2, 1, expectedSMI_Q13_R25_S2_U1},
		{"DTI_Q2_R28_S28_U5", Dti, 2, 28, 28, 5, expectedDTI_Q2_R28_S28_U5},
		{"DTI_Q2_R20_S5_U3", Dti, 2, 20, 5, 3, expectedDTI_Q2_R20_S5_U3},
		{"DTI_Q4_R14_S14_U3", Dti, 4, 14, 14, 3, expectedDTI_Q4_R14_S14_U3},
		{"MDI_R20_S5_U3", Mdi, 0, 20, 5, 3, expectedMDI_R20_S5_U3},
		{"MDI_R40_S5_U3", Mdi, 0, 40, 5, 3, expectedMDI_R40_S5_U3},
		{"CMI_R20_S5_U3", Cmi, 0, 20, 5, 3, expectedCMI_R20_S5_U3},
		{"CMI_R10_S5_U3", Cmi, 0, 10, 5, 3, expectedCMI_R10_S5_U3},
		{"CSI_R32_S32_U1", Csi, 0, 32, 32, 1, expectedCSI_R32_S32_U1},
		{"CSI_R20_S5_U3", Csi, 0, 20, 5, 3, expectedCSI_R20_S5_U3},
		{"CSI_R1_S1_U1", Csi, 0, 1, 1, 1, expectedCSI_R1_S1_U1},
		{"TVI_R32_S32_U5", Tvi, 0, 32, 32, 5, expectedTVI_R32_S32_U5},
		{"TVI_R12_S12_U1", Tvi, 0, 12, 12, 1, expectedTVI_R12_S12_U1},
		{"TVI_R25_S13_U1", Tvi, 0, 25, 13, 1, expectedTVI_R25_S13_U1},
	}
}

var testTime = time.Date(2021, time.April, 1, 0, 0, 0, 0, time.UTC)

func testBar(i int) *entities.Bar {
	return &entities.Bar{
		Time: testTime, Open: testOpen[i], High: testHigh[i], Low: testLow[i], Close: testInput[i],
	}
}

func checkVal(t *testing.T, name string, i int, exp, act float64) {
	t.Helper()

	if math.IsNaN(exp) {
		if !math.IsNaN(act) {
			t.Errorf("%s[%d]: expected NaN, got %v", name, i, act)
		}

		return
	}

	if math.Abs(act-exp) > tolerance {
		t.Errorf("%s[%d]: expected %v, got %v", name, i, exp, act)
	}
}

func checkOutput(t *testing.T, name string, i int, exp float64, output core.Output) {
	t.Helper()

	if len(output) != 1 {
		t.Fatalf("%s[%d]: outputs: expected 1, got %d", name, i, len(output))
	}

	scalar, ok := output[0].(entities.Scalar)
	if !ok {
		t.Fatalf("%s[%d]: expected a scalar output", name, i)
	}

	if scalar.Time != testTime {
		t.Errorf("%s[%d]: time: expected %v, got %v", name, i, testTime, scalar.Time)
	}

	checkVal(t, name, i, exp, scalar.Value)
}

func testNonambiguousTrendFilterCreate(t *testing.T, p *Params) *NonambiguousTrendFilter {
	t.Helper()

	ntf, err := NewNonambiguousTrendFilter(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return ntf
}

func TestNonambiguousTrendFilterDataFromBars(t *testing.T) {
	t.Parallel()

	for _, combo := range ntfCombos() {
		t.Run(combo.name, func(t *testing.T) {
			t.Parallel()

			ntf := testNonambiguousTrendFilterCreate(t, &Params{
				Base: combo.base, Q: combo.q, R: combo.r, S: combo.s, U: combo.u,
			})

			for i := range testInput {
				checkOutput(t, "ntf", i, combo.expected[i], ntf.UpdateBar(testBar(i)))
			}
		})
	}
}

func TestNonambiguousTrendFilterDataFromSamples(t *testing.T) {
	t.Parallel()

	for _, combo := range ntfCombos() {
		if combo.base != Tsi && combo.base != Mdi {
			continue
		}

		t.Run(combo.name, func(t *testing.T) {
			t.Parallel()

			ntf := testNonambiguousTrendFilterCreate(t, &Params{
				Base: combo.base, Q: combo.q, R: combo.r, S: combo.s, U: combo.u,
			})

			for i := range testInput {
				checkVal(t, "ntf", i, combo.expected[i], ntf.Update(testInput[i]))
			}
		})
	}
}

func TestNonambiguousTrendFilterRule(t *testing.T) {
	t.Parallel()

	// A passthrough CSI base is 100*(C-O)/(H-L).
	ntf := testNonambiguousTrendFilterCreate(t, &Params{Base: Csi, R: 1, S: 1, U: 1})

	tests := []struct {
		name  string
		close float64
		exp   float64
	}{
		{"first finite value", 11.0, 0.0},
		{"positive and rising", 12.0, 50.0},
		{"positive and falling", 11.0, 0.0},
		{"negative and falling", 9.0, -25.0},
		{"flat", 9.0, 0.0},
		{"negative and rising", 9.5, 0.0},
	}

	for i, tt := range tests {
		bar := &entities.Bar{Time: testTime, Open: 10.0, High: 12.0, Low: 8.0, Close: tt.close}
		checkOutput(t, tt.name, i, tt.exp, ntf.UpdateBar(bar))
	}
}

func TestNonambiguousTrendFilterWarmUp(t *testing.T) {
	t.Parallel()

	const q = 13

	ntf := testNonambiguousTrendFilterCreate(t, &Params{Base: Smi, Q: q, R: 25, S: 2, U: 1})

	for i := range q - 1 {
		checkOutput(t, "ntf", i, math.NaN(), ntf.UpdateBar(testBar(i)))

		if ntf.IsPrimed() {
			t.Errorf("expected not primed after %d updates", i+1)
		}
	}

	checkOutput(t, "ntf", q-1, 0.0, ntf.UpdateBar(testBar(q-1)))

	if !ntf.IsPrimed() {
		t.Errorf("expected primed after %d updates", q)
	}
}

func TestNonambiguousTrendFilterUpdateEntity(t *testing.T) {
	t.Parallel()

	expected := expectedTSI_R32_S13_U3

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		ntf := testNonambiguousTrendFilterCreate(t, DefaultParams())

		for i := range testInput {
			checkOutput(t, "scalar", i, expected[i],
				ntf.UpdateScalar(&entities.Scalar{Time: testTime, Value: testInput[i]}))
		}
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		ntf := testNonambiguousTrendFilterCreate(t, DefaultParams())

		for i := range testInput {
			checkOutput(t, "bar", i, expected[i], ntf.UpdateBar(testBar(i)))
		}
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		ntf := testNonambiguousTrendFilterCreate(t, DefaultParams())

		for i := range testInput {
			checkOutput(t, "quote", i, expected[i],
				ntf.UpdateQuote(&entities.Quote{Time: testTime, Bid: testInput[i], Ask: testInput[i]}))
		}
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		ntf := testNonambiguousTrendFilterCreate(t, DefaultParams())

		for i := range testInput {
			checkOutput(t, "trade", i, expected[i],
				ntf.UpdateTrade(&entities.Trade{Time: testTime, Price: testInput[i]}))
		}
	})
}

func TestNonambiguousTrendFilterMetadata(t *testing.T) {
	t.Parallel()

	const (
		mnemonic = "ntf.tsi(2,32,13,3)"
		desc     = "Nonambiguous Trend Filter " + mnemonic
	)

	act := testNonambiguousTrendFilterCreate(t, DefaultParams()).Metadata()

	if act.Identifier != core.NonambiguousTrendFilter {
		t.Errorf("identifier: expected NonambiguousTrendFilter, got %v", act.Identifier)
	}

	if act.Mnemonic != mnemonic {
		t.Errorf("mnemonic: expected '%s', got '%s'", mnemonic, act.Mnemonic)
	}

	if act.Description != desc {
		t.Errorf("description: expected '%s', got '%s'", desc, act.Description)
	}

	if len(act.Outputs) != 1 {
		t.Fatalf("outputs: expected 1, got %d", len(act.Outputs))
	}

	if act.Outputs[0].Kind != int(Value) {
		t.Errorf("Outputs[0].Kind: expected %d, got %d", Value, act.Outputs[0].Kind)
	}

	if act.Outputs[0].Shape != shape.Scalar {
		t.Errorf("Outputs[0].Shape: expected scalar, got %v", act.Outputs[0].Shape)
	}

	if act.Outputs[0].Mnemonic != mnemonic {
		t.Errorf("Outputs[0].Mnemonic: expected '%s', got '%s'", mnemonic, act.Outputs[0].Mnemonic)
	}

	if act.Outputs[0].Description != desc {
		t.Errorf("Outputs[0].Description: expected '%s', got '%s'", desc, act.Outputs[0].Description)
	}
}

//nolint:funlen
func TestNonambiguousTrendFilterMnemonic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		params   *Params
		mnemonic string
	}{
		{"default", &Params{}, "ntf.tsi(2,32,13,3)"},
		{"smi base", &Params{Base: Smi}, "ntf.smi(32,64,7,1)"},
		{"dti base", &Params{Base: Dti}, "ntf.dti(2,28,28,5)"},
		{"tvi base", &Params{Base: Tvi}, "ntf.tvi(32,32,5)"},
		{"mdi base", &Params{Base: Mdi}, "ntf.mdi(20,5,3)"},
		{"cmi base", &Params{Base: Cmi}, "ntf.cmi(20,5,3)"},
		{"csi base", &Params{Base: Csi}, "ntf.csi(32,32,1)"},
		{"custom periods", &Params{Base: Smi, Q: 5, R: 20, S: 5, U: 3}, "ntf.smi(5,20,5,3)"},
		{"only bar component set", &Params{BarComponent: entities.BarMedianPrice}, "ntf.tsi(2,32,13,3, hl/2)"},
		{"only quote component set", &Params{QuoteComponent: entities.QuoteBidPrice}, "ntf.tsi(2,32,13,3, b)"},
		{"only trade component set", &Params{TradeComponent: entities.TradeVolume}, "ntf.tsi(2,32,13,3, v)"},
		{
			"bar and quote components set",
			&Params{BarComponent: entities.BarOpenPrice, QuoteComponent: entities.QuoteBidPrice},
			"ntf.tsi(2,32,13,3, o, b)",
		},
		{
			"bar and trade components set",
			&Params{BarComponent: entities.BarHighPrice, TradeComponent: entities.TradeVolume},
			"ntf.tsi(2,32,13,3, h, v)",
		},
		{
			"quote and trade components set",
			&Params{QuoteComponent: entities.QuoteAskPrice, TradeComponent: entities.TradeVolume},
			"ntf.tsi(2,32,13,3, a, v)",
		},
		{"mdi base with component", &Params{Base: Mdi, BarComponent: entities.BarMedianPrice}, "ntf.mdi(20,5,3, hl/2)"},
		{"components ignored by bar bases", &Params{Base: Csi, BarComponent: entities.BarMedianPrice}, "ntf.csi(32,32,1)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ntf := testNonambiguousTrendFilterCreate(t, tt.params)

			if ntf.mnemonic != tt.mnemonic {
				t.Errorf("mnemonic: expected '%s', got '%s'", tt.mnemonic, ntf.mnemonic)
			}

			desc := "Nonambiguous Trend Filter " + tt.mnemonic
			if ntf.description != desc {
				t.Errorf("description: expected '%s', got '%s'", desc, ntf.description)
			}
		})
	}
}

func TestNonambiguousTrendFilterInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params *Params
	}{
		{"q too small", &Params{Q: -1}},
		{"r too small", &Params{R: -1}},
		{"s too small", &Params{S: -1}},
		{"u too small", &Params{U: -1}},
		{"unknown base", &Params{Base: Base(99)}},
		{"invalid bar component", &Params{BarComponent: entities.BarComponent(9999)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := NewNonambiguousTrendFilter(tt.params); err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}
