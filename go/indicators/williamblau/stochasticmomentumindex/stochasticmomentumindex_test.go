//nolint:testpackage
package stochasticmomentumindex

import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
)

const tolerance = 1e-9

// ul is the signal-line EMA period used for every expected signal array.
const ul = 3

type smiCombo struct {
	name      string
	q         int
	r         int
	s         int
	u         int
	expSmi    []float64
	expSignal []float64
}

func smiCombos() []smiCombo {
	return []smiCombo{
		{"Q5_R20_S5_U3", 5, 20, 5, 3, expectedQ5_R20_S5_U3, expectedQ5_R20_S5_U3_SIG_UL3},
		{"Q13_R25_S2_U1", 13, 25, 2, 1, expectedQ13_R25_S2_U1, expectedQ13_R25_S2_U1_SIG_UL3},
		{"Q2_R20_S20_U1", 2, 20, 20, 1, expectedQ2_R20_S20_U1, expectedQ2_R20_S20_U1_SIG_UL3},
		{"Q13_R25_S2_U3", 13, 25, 2, 3, expectedQ13_R25_S2_U3, expectedQ13_R25_S2_U3_SIG_UL3},
		{"Q5_R20_S5_U1", 5, 20, 5, 1, expectedQ5_R20_S5_U1, expectedQ5_R20_S5_U1_SIG_UL3},
		{"Q8_R5_S3_U1", 8, 5, 3, 1, expectedQ8_R5_S3_U1, expectedQ8_R5_S3_U1_SIG_UL3},
		{"Q21_R13_S4_U1", 21, 13, 4, 1, expectedQ21_R13_S4_U1, expectedQ21_R13_S4_U1_SIG_UL3},
		{"Q1_R20_S5_U3", 1, 20, 5, 3, expectedQ1_R20_S5_U3, expectedQ1_R20_S5_U3_SIG_UL3},
		{"Q1_R40_S20_U1", 1, 40, 20, 1, expectedQ1_R40_S20_U1, expectedQ1_R40_S20_U1_SIG_UL3},
		{"Q1_R100_S20_U1", 1, 100, 20, 1, expectedQ1_R100_S20_U1, expectedQ1_R100_S20_U1_SIG_UL3},
		{"Q1_R1_S1_U1", 1, 1, 1, 1, expectedQ1_R1_S1_U1, expectedQ1_R1_S1_U1_SIG_UL3},
		{"Q5_R1_S1_U1", 5, 1, 1, 1, expectedQ5_R1_S1_U1, expectedQ5_R1_S1_U1_SIG_UL3},
		{"Q3_R10_S10_U1", 3, 10, 10, 1, expectedQ3_R10_S10_U1, expectedQ3_R10_S10_U1_SIG_UL3},
		{"Q34_R5_S5_U1", 34, 5, 5, 1, expectedQ34_R5_S5_U1, expectedQ34_R5_S5_U1_SIG_UL3},
		{"Q2_R2_S2_U2", 2, 2, 2, 2, expectedQ2_R2_S2_U2, expectedQ2_R2_S2_U2_SIG_UL3},
		{"Q50_R20_S5_U3", 50, 20, 5, 3, expectedQ50_R20_S5_U3, expectedQ50_R20_S5_U3_SIG_UL3},
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

func TestStochasticMomentumIndexData(t *testing.T) {
	t.Parallel()

	for _, combo := range smiCombos() {
		combo := combo
		t.Run(combo.name, func(t *testing.T) {
			t.Parallel()

			smi, err := NewStochasticMomentumIndex(&Params{
				Q: combo.q, R: combo.r, S: combo.s, U: combo.u, Ul: ul,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for i := 0; i < len(testInput); i++ {
				smiVal, signalVal := smi.Update(testHigh[i], testLow[i], testInput[i])

				checkVal(t, "smi", i, combo.expSmi[i], smiVal)
				checkVal(t, "signal", i, combo.expSignal[i], signalVal)
			}
		})
	}
}

func TestStochasticMomentumIndexPassthrough(t *testing.T) {
	t.Parallel()

	// One-day stochastic with all EMA stages and the signal EMA as passthroughs,
	// so the oscillator is 100*(close-mid)/half-range and the signal equals it.
	smi, err := NewStochasticMomentumIndex(&Params{Q: 1, R: 1, S: 1, U: 1, Ul: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name  string
		high  float64
		low   float64
		close float64
		exp   float64
	}{
		{"close at high", 12.0, 10.0, 12.0, 100.0},
		{"close at low", 12.0, 10.0, 10.0, -100.0},
		{"close at midpoint", 12.0, 10.0, 11.0, 0.0},
		{"flat window", 11.0, 11.0, 11.0, 0.0},
	}

	for i, tt := range tests {
		smiVal, signalVal := smi.Update(tt.high, tt.low, tt.close)

		checkVal(t, tt.name+" smi", i, tt.exp, smiVal)
		checkVal(t, tt.name+" signal", i, tt.exp, signalVal)
	}
}

func TestStochasticMomentumIndexIsPrimed(t *testing.T) {
	t.Parallel()

	t.Run("default q", func(t *testing.T) {
		t.Parallel()

		smi, err := NewStochasticMomentumIndex(DefaultParams())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if smi.IsPrimed() {
			t.Errorf("IsPrimed: expected false before the first update")
		}

		// Bars 0..q-2 are the NaN warm-up region.
		const q = 5

		for i := 0; i < q-1; i++ {
			smiVal, signalVal := smi.Update(testHigh[i], testLow[i], testInput[i])

			if smi.IsPrimed() {
				t.Errorf("[%d] IsPrimed: expected false", i)
			}

			checkVal(t, "smi", i, math.NaN(), smiVal)
			checkVal(t, "signal", i, math.NaN(), signalVal)
		}

		for i := q - 1; i < len(testInput); i++ {
			smi.Update(testHigh[i], testLow[i], testInput[i])

			if !smi.IsPrimed() {
				t.Errorf("[%d] IsPrimed: expected true", i)
			}
		}
	})

	t.Run("q = 1 has no warm-up", func(t *testing.T) {
		t.Parallel()

		smi, err := NewStochasticMomentumIndex(&Params{Q: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if smi.IsPrimed() {
			t.Errorf("IsPrimed: expected false before the first update")
		}

		smi.Update(testHigh[0], testLow[0], testInput[0])

		if !smi.IsPrimed() {
			t.Errorf("IsPrimed: expected true after the first update")
		}
	})
}

func TestStochasticMomentumIndexUpdateBar(t *testing.T) {
	t.Parallel()

	smi, err := NewStochasticMomentumIndex(&Params{Q: 5, R: 20, S: 5, U: 3, Ul: ul})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tm := time.Now()

	var output core.Output

	for i := 0; i < len(testInput); i++ {
		output = smi.UpdateBar(&entities.Bar{
			Time: tm, High: testHigh[i], Low: testLow[i], Close: testInput[i],
		})
	}

	if len(output) != 2 {
		t.Fatalf("outputs: expected 2, got %d", len(output))
	}

	last := len(testInput) - 1

	if act := output[0].(entities.Scalar).Time; !act.Equal(tm) {
		t.Errorf("time: expected %v, got %v", tm, act)
	}

	checkVal(t, "smi", last, expectedQ5_R20_S5_U3[last], output[0].(entities.Scalar).Value)
	checkVal(t, "signal", last, expectedQ5_R20_S5_U3_SIG_UL3[last], output[1].(entities.Scalar).Value)
}

func TestStochasticMomentumIndexUpdateEntities(t *testing.T) {
	t.Parallel()

	tm := time.Now()

	t.Run("scalar is used as high, low and close", func(t *testing.T) {
		t.Parallel()

		smi, err := NewStochasticMomentumIndex(&Params{Q: 2, R: 1, S: 1, U: 1, Ul: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		output := smi.UpdateScalar(&entities.Scalar{Time: tm, Value: 10.0})

		checkVal(t, "smi", 0, math.NaN(), output[0].(entities.Scalar).Value)
		checkVal(t, "signal", 0, math.NaN(), output[1].(entities.Scalar).Value)

		// Close at the 2-bar high.
		output = smi.UpdateScalar(&entities.Scalar{Time: tm, Value: 12.0})

		checkVal(t, "smi", 1, 100.0, output[0].(entities.Scalar).Value)
		checkVal(t, "signal", 1, 100.0, output[1].(entities.Scalar).Value)
	})

	t.Run("quote mid price is used as high, low and close", func(t *testing.T) {
		t.Parallel()

		smi, err := NewStochasticMomentumIndex(&Params{Q: 2, R: 1, S: 1, U: 1, Ul: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		smi.UpdateQuote(&entities.Quote{Time: tm, Bid: 12.0, Ask: 14.0})

		// Mid 11 is at the 2-bar low.
		output := smi.UpdateQuote(&entities.Quote{Time: tm, Bid: 10.0, Ask: 12.0})

		checkVal(t, "smi", 1, -100.0, output[0].(entities.Scalar).Value)
		checkVal(t, "signal", 1, -100.0, output[1].(entities.Scalar).Value)
	})

	t.Run("trade price is used as high, low and close", func(t *testing.T) {
		t.Parallel()

		smi, err := NewStochasticMomentumIndex(&Params{Q: 2, R: 1, S: 1, U: 1, Ul: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		smi.UpdateTrade(&entities.Trade{Time: tm, Price: 10.0, Volume: 1.0})
		output := smi.UpdateTrade(&entities.Trade{Time: tm, Price: 12.0, Volume: 1.0})

		checkVal(t, "smi", 1, 100.0, output[0].(entities.Scalar).Value)
		checkVal(t, "signal", 1, 100.0, output[1].(entities.Scalar).Value)
	})
}

func TestStochasticMomentumIndexMnemonic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params *Params
		exp    string
	}{
		{"all parameters zero", &Params{}, "smi(5,20,5,3,3)"},
		{"default parameters", DefaultParams(), "smi(5,20,5,3,3)"},
		{"book basic", &Params{Q: 13, R: 25, S: 2, U: 1, Ul: 3}, "smi(13,25,2,1,3)"},
		{"passthrough signal", &Params{Q: 5, R: 20, S: 5, U: 3, Ul: 1}, "smi(5,20,5,3,1)"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			smi, err := NewStochasticMomentumIndex(tt.params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if smi.mnemonic != tt.exp {
				t.Errorf("mnemonic: expected '%s', got '%s'", tt.exp, smi.mnemonic)
			}

			expDesc := "Stochastic Momentum Index " + tt.exp
			if meta := smi.Metadata(); meta.Description != expDesc {
				t.Errorf("description: expected '%s', got '%s'", expDesc, meta.Description)
			}
		})
	}
}

func TestStochasticMomentumIndexMetadata(t *testing.T) {
	t.Parallel()

	smi, err := NewStochasticMomentumIndex(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	meta := smi.Metadata()

	if meta.Identifier != core.StochasticMomentumIndex {
		t.Errorf("identifier: expected StochasticMomentumIndex, got %v", meta.Identifier)
	}

	if meta.Mnemonic != "smi(5,20,5,3,3)" {
		t.Errorf("mnemonic: expected 'smi(5,20,5,3,3)', got '%s'", meta.Mnemonic)
	}

	if len(meta.Outputs) != 2 {
		t.Fatalf("outputs: expected 2, got %d", len(meta.Outputs))
	}

	if act := meta.Outputs[0].Mnemonic; act != "smi(5,20,5,3,3) smi" {
		t.Errorf("outputs[0].Mnemonic: expected 'smi(5,20,5,3,3) smi', got '%s'", act)
	}

	if act := meta.Outputs[1].Mnemonic; act != "smi(5,20,5,3,3) signal" {
		t.Errorf("outputs[1].Mnemonic: expected 'smi(5,20,5,3,3) signal', got '%s'", act)
	}
}

func TestStochasticMomentumIndexInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params *Params
	}{
		{"q too small", &Params{Q: -1, R: 20, S: 5, U: 3, Ul: 3}},
		{"r too small", &Params{Q: 5, R: -1, S: 5, U: 3, Ul: 3}},
		{"s too small", &Params{Q: 5, R: 20, S: -1, U: 3, Ul: 3}},
		{"u too small", &Params{Q: 5, R: 20, S: 5, U: -1, Ul: 3}},
		{"ul too small", &Params{Q: 5, R: 20, S: 5, U: 3, Ul: -1}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := NewStochasticMomentumIndex(tt.params); err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}
