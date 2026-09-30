//nolint:testpackage
package adxtypefilter

import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

const tolerance = 1e-13

type atfCombo struct {
	name     string
	source   Source
	q        int
	r        int
	s        int
	expected []float64
}

// closeCombos lists the expected arrays of the close-based (TSI) sources.
func closeCombos() []atfCombo {
	return []atfCombo{
		{"TSIMTM_Q2_R32_S32", TsiMomentum, 2, 32, 32, expectedTSIMTM_Q2_R32_S32},
		{"TSIMTM_Q2_R20_S5", TsiMomentum, 2, 20, 5, expectedTSIMTM_Q2_R20_S5},
		{"TSIMTM_Q2_R13_S1", TsiMomentum, 2, 13, 1, expectedTSIMTM_Q2_R13_S1},
		{"TSIMTM_Q2_R1_S1", TsiMomentum, 2, 1, 1, expectedTSIMTM_Q2_R1_S1},
		{"TSIMTM_Q5_R32_S32", TsiMomentum, 5, 32, 32, expectedTSIMTM_Q5_R32_S32},
		{"TSINORM_R32_S32", TsiNormalized, 2, 32, 32, expectedTSINORM_R32_S32},
		{"TSINORM_R20_S20", TsiNormalized, 2, 20, 20, expectedTSINORM_R20_S20},
	}
}

// barCombos lists the expected arrays of the high/low/close sources.
func barCombos() []atfCombo {
	return []atfCombo{
		{"SMIRAW_Q32_R32_S32", SmiMomentum, 32, 32, 32, expectedSMIRAW_Q32_R32_S32},
		{"SMIRAW_Q32_R20_S5", SmiMomentum, 32, 20, 5, expectedSMIRAW_Q32_R20_S5},
		{"SMIRAW_Q5_R32_S32", SmiMomentum, 5, 32, 32, expectedSMIRAW_Q5_R32_S32},
		{"DTINUM_Q2_R32_S32", DtiMomentum, 2, 32, 32, expectedDTINUM_Q2_R32_S32},
		{"DTINUM_Q2_R28_S28", DtiMomentum, 2, 28, 28, expectedDTINUM_Q2_R28_S28},
		{"DTINUM_Q5_R32_S32", DtiMomentum, 5, 32, 32, expectedDTINUM_Q5_R32_S32},
		{"TVI_R32_S32", TviBalance, 0, 32, 32, expectedTVI_R32_S32},
		{"TVI_R12_S12", TviBalance, 0, 12, 12, expectedTVI_R12_S12},
		{"TVI_R1_S1", TviBalance, 0, 1, 1, expectedTVI_R1_S1},
	}
}

var testTime = time.Date(2021, time.April, 1, 0, 0, 0, 0, time.UTC)

func testBar(i int) *entities.Bar {
	return &entities.Bar{
		Time: testTime, Open: testInput[i], High: testHigh[i], Low: testLow[i], Close: testInput[i],
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

func testAdxTypeFilterCreate(t *testing.T, p *Params) *AdxTypeFilter {
	t.Helper()

	atf, err := NewAdxTypeFilter(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return atf
}

func TestAdxTypeFilterCloseSources(t *testing.T) {
	t.Parallel()

	for _, combo := range closeCombos() {
		t.Run(combo.name, func(t *testing.T) {
			t.Parallel()

			atf := testAdxTypeFilterCreate(t, &Params{Source: combo.source, Q: combo.q, R: combo.r, S: combo.s})

			for i := range testInput {
				checkVal(t, "atf", i, combo.expected[i], atf.Update(testInput[i]))
			}
		})
	}
}

func TestAdxTypeFilterBarSources(t *testing.T) {
	t.Parallel()

	for _, combo := range barCombos() {
		t.Run(combo.name, func(t *testing.T) {
			t.Parallel()

			atf := testAdxTypeFilterCreate(t, &Params{Source: combo.source, Q: combo.q, R: combo.r, S: combo.s})

			for i := range testInput {
				checkOutput(t, "atf", i, combo.expected[i], atf.UpdateBar(testBar(i)))
			}
		})
	}
}

func TestAdxTypeFilterBarSourcesHighLowClose(t *testing.T) {
	t.Parallel()

	for _, combo := range barCombos() {
		t.Run(combo.name, func(t *testing.T) {
			t.Parallel()

			atf := testAdxTypeFilterCreate(t, &Params{Source: combo.source, Q: combo.q, R: combo.r, S: combo.s})

			for i := range testInput {
				checkVal(t, "atf", i, combo.expected[i], atf.UpdateHighLowClose(testHigh[i], testLow[i], testInput[i]))
			}
		})
	}
}

func TestAdxTypeFilterPassthrough(t *testing.T) {
	t.Parallel()

	// With r = s = 1 both EMAs are passthroughs, so ATF == |momentum|.
	t.Run("tsi momentum", func(t *testing.T) {
		t.Parallel()

		atf := testAdxTypeFilterCreate(t, &Params{Source: TsiMomentum, Q: 2, R: 1, S: 1})
		checkVal(t, "atf", 0, math.NaN(), atf.Update(10.0))
		checkVal(t, "atf", 1, 2.0, atf.Update(12.0))
		checkVal(t, "atf", 2, 1.0, atf.Update(11.0))
	})

	t.Run("smi momentum", func(t *testing.T) {
		t.Parallel()

		atf := testAdxTypeFilterCreate(t, &Params{Source: SmiMomentum, Q: 1, R: 1, S: 1})
		checkVal(t, "atf", 0, 0.5, atf.UpdateHighLowClose(11.0, 9.0, 10.5))
	})

	t.Run("tvi tick rule", func(t *testing.T) {
		t.Parallel()

		atf := testAdxTypeFilterCreate(t, &Params{Source: TviBalance, R: 1, S: 1})
		checkVal(t, "atf", 0, 0.0, atf.Update(10.0))
		checkVal(t, "atf", 1, 2.0, atf.Update(12.0))
		checkVal(t, "atf", 2, 3.0, atf.Update(9.0))
	})
}

func TestAdxTypeFilterWarmUp(t *testing.T) {
	t.Parallel()

	t.Run("smi default look-back", func(t *testing.T) {
		t.Parallel()

		atf := testAdxTypeFilterCreate(t, &Params{Source: SmiMomentum})

		for i := range testInput {
			v := atf.UpdateHighLowClose(testHigh[i], testLow[i], testInput[i])
			if math.IsNaN(v) != (i < 31) {
				t.Errorf("atf[%d]: unexpected NaN state %v", i, v)
			}
		}
	})

	t.Run("tvi has no warm-up", func(t *testing.T) {
		t.Parallel()

		atf := testAdxTypeFilterCreate(t, &Params{Source: TviBalance})

		for i := range testInput {
			if math.IsNaN(atf.UpdateHighLowClose(testHigh[i], testLow[i], testInput[i])) {
				t.Errorf("atf[%d]: unexpected NaN", i)
			}
		}
	})
}

func TestAdxTypeFilterNonNegative(t *testing.T) {
	t.Parallel()

	for _, source := range []Source{TsiMomentum, SmiMomentum, DtiMomentum, TviBalance, TsiNormalized} {
		atf := testAdxTypeFilterCreate(t, &Params{Source: source})

		for i := range testInput {
			output := atf.UpdateBar(testBar(i))

			v := output[0].(entities.Scalar).Value //nolint:forcetypeassert
			if v < 0.0 {
				t.Errorf("source %d atf[%d]: expected a non-negative value, got %v", source, i, v)
			}
		}
	}
}

func TestAdxTypeFilterIsPrimed(t *testing.T) {
	t.Parallel()

	const q = 5

	atf := testAdxTypeFilterCreate(t, &Params{Source: TsiMomentum, Q: q})

	for i := range q - 1 {
		atf.Update(testInput[i])

		if atf.IsPrimed() {
			t.Errorf("expected not primed after %d updates", i+1)
		}
	}

	for i := q - 1; i < len(testInput); i++ {
		atf.Update(testInput[i])

		if !atf.IsPrimed() {
			t.Errorf("expected primed after %d updates", i+1)
		}
	}
}

func TestAdxTypeFilterUpdateEntity(t *testing.T) {
	t.Parallel()

	expected := expectedTSIMTM_Q2_R32_S32

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		atf := testAdxTypeFilterCreate(t, DefaultParams())

		for i := range testInput {
			checkOutput(t, "scalar", i, expected[i],
				atf.UpdateScalar(&entities.Scalar{Time: testTime, Value: testInput[i]}))
		}
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		atf := testAdxTypeFilterCreate(t, DefaultParams())

		for i := range testInput {
			checkOutput(t, "bar", i, expected[i], atf.UpdateBar(testBar(i)))
		}
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		atf := testAdxTypeFilterCreate(t, DefaultParams())

		for i := range testInput {
			checkOutput(t, "quote", i, expected[i],
				atf.UpdateQuote(&entities.Quote{Time: testTime, Bid: testInput[i], Ask: testInput[i]}))
		}
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		atf := testAdxTypeFilterCreate(t, DefaultParams())

		for i := range testInput {
			checkOutput(t, "trade", i, expected[i],
				atf.UpdateTrade(&entities.Trade{Time: testTime, Price: testInput[i]}))
		}
	})

	t.Run("smi single values use the value as high, low and close", func(t *testing.T) {
		t.Parallel()

		p := &Params{Source: SmiMomentum, Q: 5}
		reference := testAdxTypeFilterCreate(t, p)
		scalar := testAdxTypeFilterCreate(t, p)
		quote := testAdxTypeFilterCreate(t, p)
		trade := testAdxTypeFilterCreate(t, p)

		for i := range testInput {
			v := testInput[i]
			exp := reference.UpdateHighLowClose(v, v, v)

			checkOutput(t, "scalar", i, exp, scalar.UpdateScalar(&entities.Scalar{Time: testTime, Value: v}))
			checkOutput(t, "quote", i, exp, quote.UpdateQuote(&entities.Quote{Time: testTime, Bid: v, Ask: v}))
			checkOutput(t, "trade", i, exp, trade.UpdateTrade(&entities.Trade{Time: testTime, Price: v}))
		}
	})
}

func TestAdxTypeFilterMetadata(t *testing.T) {
	t.Parallel()

	const (
		mnemonic = "atf.tsi(2,32,32)"
		desc     = "ADX-Type Filter " + mnemonic
	)

	atf := testAdxTypeFilterCreate(t, DefaultParams())
	act := atf.Metadata()

	if act.Identifier != core.AdxTypeFilter {
		t.Errorf("identifier: expected AdxTypeFilter, got %v", act.Identifier)
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
func TestAdxTypeFilterMnemonic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		params   *Params
		mnemonic string
	}{
		{"default", &Params{}, "atf.tsi(2,32,32)"},
		{"tsi source", &Params{Source: TsiMomentum}, "atf.tsi(2,32,32)"},
		{"smi source", &Params{Source: SmiMomentum}, "atf.smi(32,32,32)"},
		{"dti source", &Params{Source: DtiMomentum}, "atf.dti(2,32,32)"},
		{"tvi source", &Params{Source: TviBalance}, "atf.tvi(32,32)"},
		{"tsin source", &Params{Source: TsiNormalized}, "atf.tsin(2,32,32)"},
		{"custom periods", &Params{Source: SmiMomentum, Q: 5, R: 20, S: 5}, "atf.smi(5,20,5)"},
		{"only bar component set", &Params{BarComponent: entities.BarMedianPrice}, "atf.tsi(2,32,32, hl/2)"},
		{"only quote component set", &Params{QuoteComponent: entities.QuoteBidPrice}, "atf.tsi(2,32,32, b)"},
		{"only trade component set", &Params{TradeComponent: entities.TradeVolume}, "atf.tsi(2,32,32, v)"},
		{
			"bar and quote components set",
			&Params{BarComponent: entities.BarOpenPrice, QuoteComponent: entities.QuoteBidPrice},
			"atf.tsi(2,32,32, o, b)",
		},
		{
			"bar and trade components set",
			&Params{BarComponent: entities.BarHighPrice, TradeComponent: entities.TradeVolume},
			"atf.tsi(2,32,32, h, v)",
		},
		{
			"quote and trade components set",
			&Params{QuoteComponent: entities.QuoteAskPrice, TradeComponent: entities.TradeVolume},
			"atf.tsi(2,32,32, a, v)",
		},
		{
			"normalized source with component",
			&Params{Source: TsiNormalized, BarComponent: entities.BarMedianPrice},
			"atf.tsin(2,32,32, hl/2)",
		},
		{
			"components ignored by bar sources",
			&Params{Source: SmiMomentum, BarComponent: entities.BarMedianPrice},
			"atf.smi(32,32,32)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			atf := testAdxTypeFilterCreate(t, tt.params)

			if atf.mnemonic != tt.mnemonic {
				t.Errorf("mnemonic: expected '%s', got '%s'", tt.mnemonic, atf.mnemonic)
			}

			desc := "ADX-Type Filter " + tt.mnemonic
			if atf.description != desc {
				t.Errorf("description: expected '%s', got '%s'", desc, atf.description)
			}
		})
	}
}

func TestAdxTypeFilterInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params *Params
	}{
		{"q too small", &Params{Q: -1}},
		{"r too small", &Params{R: -1}},
		{"s too small", &Params{S: -1}},
		{"unknown source", &Params{Source: Source(99)}},
		{"invalid bar component", &Params{BarComponent: entities.BarComponent(9999)}},
		{"invalid quote component", &Params{QuoteComponent: entities.QuoteComponent(9999)}},
		{"invalid trade component", &Params{TradeComponent: entities.TradeComponent(9999)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := NewAdxTypeFilter(tt.params); err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}
