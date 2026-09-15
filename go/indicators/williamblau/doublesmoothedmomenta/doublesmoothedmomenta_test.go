//nolint:testpackage
package doublesmoothedmomenta

import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

const tolerance = 1e-9

type dmCombo struct {
	name     string
	a        int
	y        int
	z        int
	expected []float64
}

func dmCombos() []dmCombo {
	return []dmCombo{
		{"A2_Y2_Z14", 2, 2, 14, expectedA2_Y2_Z14},
		{"A2_Y1_Z14", 2, 1, 14, expectedA2_Y1_Z14},
		{"A2_Y1_Z9", 2, 1, 9, expectedA2_Y1_Z9},
		{"A2_Y1_Z2", 2, 1, 2, expectedA2_Y1_Z2},
		{"A2_Y3_Z9", 2, 3, 9, expectedA2_Y3_Z9},
		{"A2_Y5_Z5", 2, 5, 5, expectedA2_Y5_Z5},
		{"A2_Y2_Z5", 2, 2, 5, expectedA2_Y2_Z5},
		{"A2_Y1_Z1", 2, 1, 1, expectedA2_Y1_Z1},
		{"A1_Y1_Z1", 1, 1, 1, expectedA1_Y1_Z1},
		{"A5_Y2_Z14", 5, 2, 14, expectedA5_Y2_Z14},
		{"A10_Y3_Z5", 10, 3, 5, expectedA10_Y3_Z5},
		{"A14_Y2_Z9", 14, 2, 9, expectedA14_Y2_Z9},
		{"A20_Y5_Z3", 20, 5, 3, expectedA20_Y5_Z3},
		{"A3_Y3_Z3", 3, 3, 3, expectedA3_Y3_Z3},
		{"A7_Y4_Z2", 7, 4, 2, expectedA7_Y4_Z2},
		{"A32_Y2_Z7", 32, 2, 7, expectedA32_Y2_Z7},
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

func testDoubleSmoothedMomentaCreate(t *testing.T, a, y, z int) *DoubleSmoothedMomenta {
	t.Helper()

	dm, err := NewDoubleSmoothedMomenta(&Params{A: a, Y: y, Z: z})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return dm
}

func TestDoubleSmoothedMomentaData(t *testing.T) {
	t.Parallel()

	for _, combo := range dmCombos() {
		combo := combo
		t.Run(combo.name, func(t *testing.T) {
			t.Parallel()

			dm := testDoubleSmoothedMomentaCreate(t, combo.a, combo.y, combo.z)

			for i := 0; i < len(testInput); i++ {
				checkVal(t, "dm", i, combo.expected[i], dm.Update(testInput[i]))
			}
		})
	}
}

func TestDoubleSmoothedMomentaWarmUp(t *testing.T) {
	t.Parallel()

	const a = 5

	dm := testDoubleSmoothedMomentaCreate(t, a, 2, 14)

	for i := 0; i < a-1; i++ {
		if !math.IsNaN(dm.Update(testInput[i])) {
			t.Errorf("dm[%d]: expected NaN during the a-bar warm-up", i)
		}
	}

	for i := a - 1; i < len(testInput); i++ {
		if math.IsNaN(dm.Update(testInput[i])) {
			t.Errorf("dm[%d]: unexpected NaN", i)
		}
	}
}

func TestDoubleSmoothedMomentaNoWarmUpWhenLookBackIsOne(t *testing.T) {
	t.Parallel()

	dm := testDoubleSmoothedMomentaCreate(t, 1, 1, 1)

	for i := 0; i < len(testInput); i++ {
		if math.IsNaN(dm.Update(testInput[i])) {
			t.Errorf("dm[%d]: unexpected NaN", i)
		}
	}
}

func TestDoubleSmoothedMomentaBounds(t *testing.T) {
	t.Parallel()

	dm, err := NewDoubleSmoothedMomenta(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < len(testInput); i++ {
		v := dm.Update(testInput[i])
		if math.IsNaN(v) {
			continue
		}

		if v < 0.0 || v > 100.0 {
			t.Errorf("dm[%d]: expected a value in [0, 100], got %v", i, v)
		}
	}
}

func TestDoubleSmoothedMomentaDegenerate(t *testing.T) {
	t.Parallel()

	dm := testDoubleSmoothedMomentaCreate(t, 1, 1, 1)

	for i := 0; i < len(testInput); i++ {
		if v := dm.Update(testInput[i]); v != 0.0 {
			t.Errorf("dm[%d]: expected exactly 0 when a=1, got %v", i, v)
		}
	}
}

// emaFormRsi is an independently-coded EMA-form RSI: 100 * EMA(up, z) / EMA(up+dn, z).
func emaFormRsi(closes []float64, z int) []float64 {
	alpha := 2.0 / (float64(z) + 1.0)
	result := make([]float64, len(closes))
	result[0] = math.NaN()

	var (
		numerator   float64
		denominator float64
		primed      bool
	)

	for k := 1; k < len(closes); k++ {
		diff := closes[k] - closes[k-1]

		var up, dn float64

		if diff > 0.0 {
			up = diff
		} else {
			dn = -diff
		}

		if primed {
			numerator = alpha*up + (1.0-alpha)*numerator
			denominator = alpha*(up+dn) + (1.0-alpha)*denominator
		} else {
			numerator = up
			denominator = up + dn
			primed = true
		}

		if denominator <= 0.0 {
			result[k] = 0.0
		} else {
			result[k] = 100.0 * numerator / denominator
		}
	}

	return result
}

func TestDoubleSmoothedMomentaRsiEquivalence(t *testing.T) {
	t.Parallel()

	for _, z := range []int{1, 2, 9, 14} {
		z := z
		t.Run("rsi", func(t *testing.T) {
			t.Parallel()

			expected := emaFormRsi(testInput, z)
			dm := testDoubleSmoothedMomentaCreate(t, 2, 1, z)

			for i := 0; i < len(testInput); i++ {
				checkVal(t, "rsi", i, expected[i], dm.Update(testInput[i]))
			}
		})
	}
}

func TestDoubleSmoothedMomentaIsPrimed(t *testing.T) {
	t.Parallel()

	const a = 5

	dm := testDoubleSmoothedMomentaCreate(t, a, 2, 14)

	for i := 0; i < a-1; i++ {
		dm.Update(testInput[i])

		if dm.IsPrimed() {
			t.Errorf("expected not primed after %d updates", i+1)
		}
	}

	for i := a - 1; i < len(testInput); i++ {
		dm.Update(testInput[i])

		if !dm.IsPrimed() {
			t.Errorf("expected primed after %d updates", i+1)
		}
	}
}

func TestDoubleSmoothedMomentaUpdateEntity(t *testing.T) {
	t.Parallel()

	const (
		a = 2
		y = 2
		z = 14
	)

	tm := time.Date(2021, time.April, 1, 0, 0, 0, 0, time.UTC)

	check := func(name string, i int, exp float64, act core.Output) {
		t.Helper()

		if len(act) != 1 {
			t.Fatalf("%s[%d]: expected 1 output, got %d", name, i, len(act))
		}

		s, ok := act[0].(entities.Scalar)
		if !ok {
			t.Fatalf("%s[%d]: expected a scalar output", name, i)
		}

		checkVal(t, name, i, exp, s.Value)
	}

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		dm := testDoubleSmoothedMomentaCreate(t, a, y, z)

		for i := 0; i < len(testInput); i++ {
			check("scalar", i, expectedA2_Y2_Z14[i],
				dm.UpdateScalar(&entities.Scalar{Time: tm, Value: testInput[i]}))
		}
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		dm := testDoubleSmoothedMomentaCreate(t, a, y, z)

		for i := 0; i < len(testInput); i++ {
			check("bar", i, expectedA2_Y2_Z14[i],
				dm.UpdateBar(&entities.Bar{Time: tm, Close: testInput[i]}))
		}
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		dm := testDoubleSmoothedMomentaCreate(t, a, y, z)

		for i := 0; i < len(testInput); i++ {
			check("quote", i, expectedA2_Y2_Z14[i],
				dm.UpdateQuote(&entities.Quote{Time: tm, Bid: testInput[i], Ask: testInput[i]}))
		}
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		dm := testDoubleSmoothedMomentaCreate(t, a, y, z)

		for i := 0; i < len(testInput); i++ {
			check("trade", i, expectedA2_Y2_Z14[i],
				dm.UpdateTrade(&entities.Trade{Time: tm, Price: testInput[i]}))
		}
	})
}

func TestDoubleSmoothedMomentaMetadata(t *testing.T) {
	t.Parallel()

	dm, err := NewDoubleSmoothedMomenta(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	act := dm.Metadata()

	if act.Identifier != core.DoubleSmoothedMomenta {
		t.Errorf("identifier: expected DoubleSmoothedMomenta, got %v", act.Identifier)
	}

	if act.Mnemonic != "dm(2,2,14)" {
		t.Errorf("mnemonic: expected 'dm(2,2,14)', got '%s'", act.Mnemonic)
	}

	if act.Description != "Double-Smoothed Momenta dm(2,2,14)" {
		t.Errorf("description: expected 'Double-Smoothed Momenta dm(2,2,14)', got '%s'", act.Description)
	}

	if len(act.Outputs) != 1 {
		t.Fatalf("outputs: expected 1, got %d", len(act.Outputs))
	}

	if act.Outputs[0].Shape != shape.Scalar {
		t.Errorf("Outputs[0].Shape: expected scalar, got %v", act.Outputs[0].Shape)
	}

	if act.Outputs[0].Mnemonic != "dm(2,2,14)" {
		t.Errorf("Outputs[0].Mnemonic: expected 'dm(2,2,14)', got '%s'", act.Outputs[0].Mnemonic)
	}
}

//nolint:funlen
func TestDoubleSmoothedMomentaMnemonic(t *testing.T) {
	t.Parallel()

	check := func(t *testing.T, params *Params, mnemonic string) {
		t.Helper()

		dm, err := NewDoubleSmoothedMomenta(params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if dm.LineIndicator.Mnemonic != mnemonic {
			t.Errorf("mnemonic: expected '%s', got '%s'", mnemonic, dm.LineIndicator.Mnemonic)
		}

		desc := "Double-Smoothed Momenta " + mnemonic
		if dm.LineIndicator.Description != desc {
			t.Errorf("description: expected '%s', got '%s'", desc, dm.LineIndicator.Description)
		}
	}

	t.Run("all components zero", func(t *testing.T) {
		t.Parallel()
		check(t, &Params{A: 2, Y: 2, Z: 14}, "dm(2,2,14)")
	})

	t.Run("custom periods", func(t *testing.T) {
		t.Parallel()
		check(t, &Params{A: 10, Y: 3, Z: 5}, "dm(10,3,5)")
	})

	t.Run("only bar component set", func(t *testing.T) {
		t.Parallel()
		check(t, &Params{A: 2, Y: 2, Z: 14, BarComponent: entities.BarMedianPrice}, "dm(2,2,14, hl/2)")
	})

	t.Run("only quote component set", func(t *testing.T) {
		t.Parallel()
		check(t, &Params{A: 2, Y: 2, Z: 14, QuoteComponent: entities.QuoteBidPrice}, "dm(2,2,14, b)")
	})

	t.Run("only trade component set", func(t *testing.T) {
		t.Parallel()
		check(t, &Params{A: 2, Y: 2, Z: 14, TradeComponent: entities.TradeVolume}, "dm(2,2,14, v)")
	})

	t.Run("bar and quote components set", func(t *testing.T) {
		t.Parallel()
		check(t, &Params{
			A: 2, Y: 2, Z: 14,
			BarComponent: entities.BarOpenPrice, QuoteComponent: entities.QuoteBidPrice,
		}, "dm(2,2,14, o, b)")
	})

	t.Run("bar and trade components set", func(t *testing.T) {
		t.Parallel()
		check(t, &Params{
			A: 2, Y: 2, Z: 14,
			BarComponent: entities.BarHighPrice, TradeComponent: entities.TradeVolume,
		}, "dm(2,2,14, h, v)")
	})

	t.Run("quote and trade components set", func(t *testing.T) {
		t.Parallel()
		check(t, &Params{
			A: 2, Y: 2, Z: 14,
			QuoteComponent: entities.QuoteAskPrice, TradeComponent: entities.TradeVolume,
		}, "dm(2,2,14, a, v)")
	})
}

func TestDoubleSmoothedMomentaInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params *Params
	}{
		{"a too small", &Params{A: -1, Y: 2, Z: 14}},
		{"y too small", &Params{A: 2, Y: -1, Z: 14}},
		{"z too small", &Params{A: 2, Y: 2, Z: -1}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := NewDoubleSmoothedMomenta(tt.params); err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}
