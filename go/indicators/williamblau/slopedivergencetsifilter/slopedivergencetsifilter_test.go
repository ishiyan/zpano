//nolint:testpackage
package slopedivergencetsifilter

import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

const tolerance = 1e-13

type sdtsiCombo struct {
	name     string
	r        int
	s        int
	u        int
	x        int
	y        int
	expected []float64
}

// sdtsiCombos lists every expected array; all use the book momentum q=2.
func sdtsiCombos() []sdtsiCombo {
	return []sdtsiCombo{
		{"R32_S32_U7_X32_Y7", 32, 32, 7, 32, 7, expectedR32_S32_U7_X32_Y7},
		{"R32_S32_U1_X32_Y1", 32, 32, 1, 32, 1, expectedR32_S32_U1_X32_Y1},
		{"R32_S32_U7_X32_Y1", 32, 32, 7, 32, 1, expectedR32_S32_U7_X32_Y1},
		{"R32_S32_U1_X32_Y7", 32, 32, 1, 32, 7, expectedR32_S32_U1_X32_Y7},
		{"R1_S1_U1_X1_Y1", 1, 1, 1, 1, 1, expectedR1_S1_U1_X1_Y1},
		{"R20_S5_U3_X20_Y3", 20, 5, 3, 20, 3, expectedR20_S5_U3_X20_Y3},
		{"R32_S13_U3_X32_Y7", 32, 13, 3, 32, 7, expectedR32_S13_U3_X32_Y7},
		{"R12_S12_U1_X12_Y1", 12, 12, 1, 12, 1, expectedR12_S12_U1_X12_Y1},
		{"R25_S13_U1_X25_Y1", 25, 13, 1, 25, 1, expectedR25_S13_U1_X25_Y1},
		{"R64_S64_U7_X32_Y7", 64, 64, 7, 32, 7, expectedR64_S64_U7_X32_Y7},
		{"R32_S32_U7_X16_Y3", 32, 32, 7, 16, 3, expectedR32_S32_U7_X16_Y3},
		{"R5_S5_U5_X5_Y5", 5, 5, 5, 5, 5, expectedR5_S5_U5_X5_Y5},
		{"R10_S10_U1_X10_Y1", 10, 10, 1, 10, 1, expectedR10_S10_U1_X10_Y1},
		{"R40_S20_U5_X32_Y7", 40, 20, 5, 32, 7, expectedR40_S20_U5_X32_Y7},
		{"R32_S5_U1_X32_Y1", 32, 5, 1, 32, 1, expectedR32_S5_U1_X32_Y1},
		{"R50_S25_U1_X50_Y1", 50, 25, 1, 50, 1, expectedR50_S25_U1_X50_Y1},
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

func checkOutput(t *testing.T, name string, i int, tm time.Time, exp float64, output core.Output) {
	t.Helper()

	if len(output) != 1 {
		t.Fatalf("%s[%d]: outputs: expected 1, got %d", name, i, len(output))
	}

	scalar, ok := output[0].(entities.Scalar)
	if !ok {
		t.Fatalf("%s[%d]: expected a scalar output", name, i)
	}

	if scalar.Time != tm {
		t.Errorf("%s[%d]: time: expected %v, got %v", name, i, tm, scalar.Time)
	}

	checkVal(t, name, i, exp, scalar.Value)
}

func testSlopeDivergenceTsiFilterCreate(t *testing.T, p *Params) *SlopeDivergenceTsiFilter {
	t.Helper()

	sd, err := NewSlopeDivergenceTsiFilter(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return sd
}

func TestSlopeDivergenceTsiFilterData(t *testing.T) {
	t.Parallel()

	for _, combo := range sdtsiCombos() {
		t.Run(combo.name, func(t *testing.T) {
			t.Parallel()

			sd := testSlopeDivergenceTsiFilterCreate(t, &Params{
				Q: 2, R: combo.r, S: combo.s, U: combo.u, X: combo.x, Y: combo.y,
			})

			for i := range testInput {
				checkVal(t, "sdtsi", i, combo.expected[i], sd.Update(testInput[i]))
			}
		})
	}
}

func TestSlopeDivergenceTsiFilterPassthrough(t *testing.T) {
	t.Parallel()

	// All EMA stages are passthroughs: TSI is +/-100, the reference is the close.
	sd := testSlopeDivergenceTsiFilterCreate(t, &Params{Q: 2, R: 1, S: 1, U: 1, X: 1, Y: 1})

	tests := []struct {
		name   string
		sample float64
		exp    float64
	}{
		{"momentum undefined", 10.0, math.NaN()},
		{"first finite tsi, no slope", 12.0, 0.0},
		{"both falling", 11.0, -100.0},
		{"both rising", 13.0, 100.0},
		{"tsi flat", 14.0, 0.0},
	}

	for i, tt := range tests {
		checkVal(t, tt.name, i, tt.exp, sd.Update(tt.sample))
	}
}

func TestSlopeDivergenceTsiFilterWarmUp(t *testing.T) {
	t.Parallel()

	const q = 5

	sd := testSlopeDivergenceTsiFilterCreate(t, &Params{Q: q})

	for i := range q - 1 {
		if !math.IsNaN(sd.Update(testInput[i])) {
			t.Errorf("sdtsi[%d]: expected NaN during the momentum warm-up", i)
		}
	}

	if v := sd.Update(testInput[q-1]); v != 0.0 {
		t.Errorf("sdtsi[%d]: expected 0 at the first finite bar, got %v", q-1, v)
	}

	for i := q; i < len(testInput); i++ {
		if math.IsNaN(sd.Update(testInput[i])) {
			t.Errorf("sdtsi[%d]: unexpected NaN", i)
		}
	}
}

func TestSlopeDivergenceTsiFilterBounds(t *testing.T) {
	t.Parallel()

	sd := testSlopeDivergenceTsiFilterCreate(t, DefaultParams())

	for i := range testInput {
		v := sd.Update(testInput[i])
		if math.IsNaN(v) {
			continue
		}

		if v < -100.0 || v > 100.0 {
			t.Errorf("sdtsi[%d]: expected a value in [-100, 100], got %v", i, v)
		}
	}
}

func TestSlopeDivergenceTsiFilterIsPrimed(t *testing.T) {
	t.Parallel()

	const q = 5

	sd := testSlopeDivergenceTsiFilterCreate(t, &Params{Q: q})

	for i := range q - 1 {
		sd.Update(testInput[i])

		if sd.IsPrimed() {
			t.Errorf("expected not primed after %d updates", i+1)
		}
	}

	for i := q - 1; i < len(testInput); i++ {
		sd.Update(testInput[i])

		if !sd.IsPrimed() {
			t.Errorf("expected primed after %d updates", i+1)
		}
	}
}

func TestSlopeDivergenceTsiFilterUpdateEntity(t *testing.T) {
	t.Parallel()

	tm := time.Date(2021, time.April, 1, 0, 0, 0, 0, time.UTC)
	expected := expectedR32_S32_U7_X32_Y7

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		sd := testSlopeDivergenceTsiFilterCreate(t, DefaultParams())

		for i := range testInput {
			checkOutput(t, "scalar", i, tm, expected[i],
				sd.UpdateScalar(&entities.Scalar{Time: tm, Value: testInput[i]}))
		}
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		sd := testSlopeDivergenceTsiFilterCreate(t, DefaultParams())

		for i := range testInput {
			checkOutput(t, "bar", i, tm, expected[i],
				sd.UpdateBar(&entities.Bar{Time: tm, Close: testInput[i]}))
		}
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		sd := testSlopeDivergenceTsiFilterCreate(t, DefaultParams())

		for i := range testInput {
			checkOutput(t, "quote", i, tm, expected[i],
				sd.UpdateQuote(&entities.Quote{Time: tm, Bid: testInput[i], Ask: testInput[i]}))
		}
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		sd := testSlopeDivergenceTsiFilterCreate(t, DefaultParams())

		for i := range testInput {
			checkOutput(t, "trade", i, tm, expected[i],
				sd.UpdateTrade(&entities.Trade{Time: tm, Price: testInput[i]}))
		}
	})
}

func TestSlopeDivergenceTsiFilterMetadata(t *testing.T) {
	t.Parallel()

	const (
		mnemonic = "sdtsi(2,32,32,7,32,7)"
		desc     = "Slope Divergence TSI Filter " + mnemonic
	)

	sd := testSlopeDivergenceTsiFilterCreate(t, DefaultParams())
	act := sd.Metadata()

	if act.Identifier != core.SlopeDivergenceTsiFilter {
		t.Errorf("identifier: expected SlopeDivergenceTsiFilter, got %v", act.Identifier)
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
func TestSlopeDivergenceTsiFilterMnemonic(t *testing.T) {
	t.Parallel()

	check := func(t *testing.T, params *Params, mnemonic string) {
		t.Helper()

		sd := testSlopeDivergenceTsiFilterCreate(t, params)

		if sd.LineIndicator.Mnemonic != mnemonic {
			t.Errorf("mnemonic: expected '%s', got '%s'", mnemonic, sd.LineIndicator.Mnemonic)
		}

		desc := "Slope Divergence TSI Filter " + mnemonic
		if sd.LineIndicator.Description != desc {
			t.Errorf("description: expected '%s', got '%s'", desc, sd.LineIndicator.Description)
		}
	}

	t.Run("all components zero", func(t *testing.T) {
		t.Parallel()
		check(t, &Params{}, "sdtsi(2,32,32,7,32,7)")
	})

	t.Run("custom periods", func(t *testing.T) {
		t.Parallel()
		check(t, &Params{Q: 3, R: 20, S: 5, U: 3, X: 20, Y: 3}, "sdtsi(3,20,5,3,20,3)")
	})

	t.Run("only bar component set", func(t *testing.T) {
		t.Parallel()
		check(t, &Params{BarComponent: entities.BarMedianPrice}, "sdtsi(2,32,32,7,32,7, hl/2)")
	})

	t.Run("only quote component set", func(t *testing.T) {
		t.Parallel()
		check(t, &Params{QuoteComponent: entities.QuoteBidPrice}, "sdtsi(2,32,32,7,32,7, b)")
	})

	t.Run("only trade component set", func(t *testing.T) {
		t.Parallel()
		check(t, &Params{TradeComponent: entities.TradeVolume}, "sdtsi(2,32,32,7,32,7, v)")
	})

	t.Run("bar and quote components set", func(t *testing.T) {
		t.Parallel()
		check(t, &Params{
			BarComponent: entities.BarOpenPrice, QuoteComponent: entities.QuoteBidPrice,
		}, "sdtsi(2,32,32,7,32,7, o, b)")
	})

	t.Run("bar and trade components set", func(t *testing.T) {
		t.Parallel()
		check(t, &Params{
			BarComponent: entities.BarHighPrice, TradeComponent: entities.TradeVolume,
		}, "sdtsi(2,32,32,7,32,7, h, v)")
	})

	t.Run("quote and trade components set", func(t *testing.T) {
		t.Parallel()
		check(t, &Params{
			QuoteComponent: entities.QuoteAskPrice, TradeComponent: entities.TradeVolume,
		}, "sdtsi(2,32,32,7,32,7, a, v)")
	})
}

func TestSlopeDivergenceTsiFilterInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params *Params
	}{
		{"q too small", &Params{Q: -1}},
		{"r too small", &Params{R: -1}},
		{"s too small", &Params{S: -1}},
		{"u too small", &Params{U: -1}},
		{"x too small", &Params{X: -1}},
		{"y too small", &Params{Y: -1}},
		{"invalid bar component", &Params{BarComponent: entities.BarComponent(9999)}},
		{"invalid quote component", &Params{QuoteComponent: entities.QuoteComponent(9999)}},
		{"invalid trade component", &Params{TradeComponent: entities.TradeComponent(9999)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := NewSlopeDivergenceTsiFilter(tt.params); err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}
