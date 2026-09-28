//nolint:testpackage
package directionaltrendindex

import (
	"fmt"
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
)

const tolerance = 1e-13

type dtiCombo struct {
	name      string
	q         int
	r         int
	s         int
	u         int
	ul        int
	expDti    []float64
	expSignal []float64
}

func dtiCombos() []dtiCombo {
	return []dtiCombo{
		{"Q2_R20_S5_U3", 2, 20, 5, 3, 3, expectedQ2_R20_S5_U3, expectedQ2_R20_S5_U3_SIG_UL3},
		{"Q2_R25_S13_U1", 2, 25, 13, 1, 3, expectedQ2_R25_S13_U1, expectedQ2_R25_S13_U1_SIG_UL3},
		{"Q2_R20_S5_U1", 2, 20, 5, 1, 3, expectedQ2_R20_S5_U1, expectedQ2_R20_S5_U1_SIG_UL3},
		{"Q2_R28_S28_U5", 2, 28, 28, 5, 3, expectedQ2_R28_S28_U5, expectedQ2_R28_S28_U5_SIG_UL3},
		{"Q2_R1_S1_U1", 2, 1, 1, 1, 3, expectedQ2_R1_S1_U1, expectedQ2_R1_S1_U1_SIG_UL3},
		{"Q3_R20_S5_U3", 3, 20, 5, 3, 3, expectedQ3_R20_S5_U3, expectedQ3_R20_S5_U3_SIG_UL3},
		{"Q5_R20_S5_U3", 5, 20, 5, 3, 3, expectedQ5_R20_S5_U3, expectedQ5_R20_S5_U3_SIG_UL3},
		{"Q2_R13_S13_U1", 2, 13, 13, 1, 3, expectedQ2_R13_S13_U1, expectedQ2_R13_S13_U1_SIG_UL3},
		{"Q2_R40_S20_U1", 2, 40, 20, 1, 3, expectedQ2_R40_S20_U1, expectedQ2_R40_S20_U1_SIG_UL3},
		{"Q2_R5_S5_U5", 2, 5, 5, 5, 3, expectedQ2_R5_S5_U5, expectedQ2_R5_S5_U5_SIG_UL3},
		{"Q1_R20_S5_U3", 1, 20, 5, 3, 3, expectedQ1_R20_S5_U3, expectedQ1_R20_S5_U3_SIG_UL3},
		{"Q10_R20_S5_U1", 10, 20, 5, 1, 3, expectedQ10_R20_S5_U1, expectedQ10_R20_S5_U1_SIG_UL3},
		{"Q2_R9_S3_U1", 2, 9, 3, 1, 3, expectedQ2_R9_S3_U1, expectedQ2_R9_S3_U1_SIG_UL3},
		{"Q2_R64_S64_U1", 2, 64, 64, 1, 3, expectedQ2_R64_S64_U1, expectedQ2_R64_S64_U1_SIG_UL3},
		{"Q4_R28_S28_U5", 4, 28, 28, 5, 3, expectedQ4_R28_S28_U5, expectedQ4_R28_S28_U5_SIG_UL3},
		{"Q2_R7_S4_U2", 2, 7, 4, 2, 3, expectedQ2_R7_S4_U2, expectedQ2_R7_S4_U2_SIG_UL3},
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

func TestDirectionalTrendIndexData(t *testing.T) {
	t.Parallel()

	for _, combo := range dtiCombos() {
		t.Run(combo.name, func(t *testing.T) {
			t.Parallel()

			dti, err := NewDirectionalTrendIndex(&Params{
				Q: combo.q, R: combo.r, S: combo.s, U: combo.u, Ul: combo.ul,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for i := range testHigh {
				dtiValue, signalValue := dti.Update(testHigh[i], testLow[i])

				checkVal(t, "dti", i, combo.expDti[i], dtiValue)
				checkVal(t, "signal", i, combo.expSignal[i], signalValue)
			}
		})
	}
}

func TestDirectionalTrendIndexPassthrough(t *testing.T) {
	t.Parallel()

	t.Run("all stages passthrough", func(t *testing.T) {
		t.Parallel()

		// One-bar HLM with all EMA stages and the signal EMA as passthroughs,
		// so the oscillator is 100*sign(HLM) and the signal equals it.
		dti, err := NewDirectionalTrendIndex(&Params{Q: 2, R: 1, S: 1, U: 1, Ul: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		tests := []struct {
			name string
			high float64
			low  float64
			exp  float64
		}{
			{"first bar", 10.0, 9.0, math.NaN()},
			{"high rises", 12.0, 11.0, 100.0}, // HMU=+2, LMD=0
			{"low falls", 11.0, 8.0, -100.0},  // HMU=0, LMD=3
			{"inside bar", 10.0, 9.0, 0.0},    // HMU=0, LMD=0 -> guard
		}

		for i, tt := range tests {
			dtiValue, signalValue := dti.Update(tt.high, tt.low)

			checkVal(t, tt.name+" dti", i, tt.exp, dtiValue)
			checkVal(t, tt.name+" signal", i, tt.exp, signalValue)
		}
	})

	t.Run("q = 1 yields zero on every bar", func(t *testing.T) {
		t.Parallel()

		// HLM == 0 on every bar -> the division guard yields 0.0 from bar 0.
		dti, err := NewDirectionalTrendIndex(&Params{Q: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for i := range testHigh {
			dtiValue, signalValue := dti.Update(testHigh[i], testLow[i])

			if dtiValue != 0.0 || signalValue != 0.0 {
				t.Errorf("[%d]: expected 0, 0, got %v, %v", i, dtiValue, signalValue)
			}
		}
	})

	t.Run("ul = 1 makes signal equal dti", func(t *testing.T) {
		t.Parallel()

		dti, err := NewDirectionalTrendIndex(&Params{Q: 2, R: 20, S: 5, U: 3, Ul: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for i := range testHigh {
			dtiValue, signalValue := dti.Update(testHigh[i], testLow[i])

			if math.IsNaN(dtiValue) {
				checkVal(t, "signal", i, math.NaN(), signalValue)
			} else if dtiValue != signalValue {
				t.Errorf("[%d]: expected signal %v to equal dti %v", i, signalValue, dtiValue)
			}
		}
	})
}

func TestDirectionalTrendIndexIsPrimed(t *testing.T) {
	t.Parallel()

	for _, q := range []int{2, 3, 5, 10} {
		t.Run(fmt.Sprintf("q=%d", q), func(t *testing.T) {
			t.Parallel()

			dti, err := NewDirectionalTrendIndex(&Params{Q: q})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if dti.IsPrimed() {
				t.Errorf("IsPrimed: expected false before the first update")
			}

			// Bars 0..q-2 are the NaN warm-up region.
			for i := range q - 1 {
				dtiValue, signalValue := dti.Update(testHigh[i], testLow[i])

				if dti.IsPrimed() {
					t.Errorf("q=%d [%d] IsPrimed: expected false", q, i)
				}

				checkVal(t, "dti", i, math.NaN(), dtiValue)
				checkVal(t, "signal", i, math.NaN(), signalValue)
			}

			for i := q - 1; i < len(testHigh); i++ {
				dti.Update(testHigh[i], testLow[i])

				if !dti.IsPrimed() {
					t.Errorf("q=%d [%d] IsPrimed: expected true", q, i)
				}
			}
		})
	}

	t.Run("q = 1 has no warm-up", func(t *testing.T) {
		t.Parallel()

		dti, err := NewDirectionalTrendIndex(&Params{Q: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if dti.IsPrimed() {
			t.Errorf("IsPrimed: expected false before the first update")
		}

		dti.Update(testHigh[0], testLow[0])

		if !dti.IsPrimed() {
			t.Errorf("IsPrimed: expected true after the first update")
		}
	})
}

func TestDirectionalTrendIndexUpdateBar(t *testing.T) {
	t.Parallel()

	dti, err := NewDirectionalTrendIndex(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tm := time.Now()

	var output core.Output

	for i := range testHigh {
		output = dti.UpdateBar(&entities.Bar{
			Time: tm, High: testHigh[i], Low: testLow[i],
		})
	}

	if len(output) != 2 {
		t.Fatalf("outputs: expected 2, got %d", len(output))
	}

	last := len(testHigh) - 1

	if act := output[0].(entities.Scalar).Time; !act.Equal(tm) {
		t.Errorf("time: expected %v, got %v", tm, act)
	}

	checkVal(t, "dti", last, expectedQ2_R20_S5_U3[last], output[0].(entities.Scalar).Value)
	checkVal(t, "signal", last, expectedQ2_R20_S5_U3_SIG_UL3[last], output[1].(entities.Scalar).Value)
}

func TestDirectionalTrendIndexUpdateEntities(t *testing.T) {
	t.Parallel()

	tm := time.Now()

	t.Run("scalar is used as high and low", func(t *testing.T) {
		t.Parallel()

		dti, err := NewDirectionalTrendIndex(&Params{Q: 2, R: 1, S: 1, U: 1, Ul: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		output := dti.UpdateScalar(&entities.Scalar{Time: tm, Value: 10.0})

		checkVal(t, "dti", 0, math.NaN(), output[0].(entities.Scalar).Value)
		checkVal(t, "signal", 0, math.NaN(), output[1].(entities.Scalar).Value)

		// Rising value: HLM is the plain one-bar momentum.
		output = dti.UpdateScalar(&entities.Scalar{Time: tm, Value: 12.0})

		checkVal(t, "dti", 1, 100.0, output[0].(entities.Scalar).Value)
		checkVal(t, "signal", 1, 100.0, output[1].(entities.Scalar).Value)
	})

	t.Run("quote mid price is used as high and low", func(t *testing.T) {
		t.Parallel()

		dti, err := NewDirectionalTrendIndex(&Params{Q: 2, R: 1, S: 1, U: 1, Ul: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		dti.UpdateQuote(&entities.Quote{Time: tm, Bid: 12.0, Ask: 14.0})

		// Mid falls from 13 to 11.
		output := dti.UpdateQuote(&entities.Quote{Time: tm, Bid: 10.0, Ask: 12.0})

		checkVal(t, "dti", 1, -100.0, output[0].(entities.Scalar).Value)
		checkVal(t, "signal", 1, -100.0, output[1].(entities.Scalar).Value)
	})

	t.Run("trade price is used as high and low", func(t *testing.T) {
		t.Parallel()

		dti, err := NewDirectionalTrendIndex(&Params{Q: 2, R: 1, S: 1, U: 1, Ul: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		dti.UpdateTrade(&entities.Trade{Time: tm, Price: 10.0, Volume: 1.0})
		output := dti.UpdateTrade(&entities.Trade{Time: tm, Price: 12.0, Volume: 1.0})

		checkVal(t, "dti", 1, 100.0, output[0].(entities.Scalar).Value)
		checkVal(t, "signal", 1, 100.0, output[1].(entities.Scalar).Value)
	})
}

func TestDirectionalTrendIndexMnemonic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params *Params
		exp    string
	}{
		{"all parameters zero", &Params{}, "dti(2,20,5,3,3)"},
		{"default parameters", DefaultParams(), "dti(2,20,5,3,3)"},
		{"book alternative", &Params{Q: 2, R: 25, S: 13, U: 1, Ul: 3}, "dti(2,25,13,1,3)"},
		{"passthrough signal", &Params{Q: 4, R: 28, S: 28, U: 5, Ul: 1}, "dti(4,28,28,5,1)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dti, err := NewDirectionalTrendIndex(tt.params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if dti.mnemonic != tt.exp {
				t.Errorf("mnemonic: expected '%s', got '%s'", tt.exp, dti.mnemonic)
			}

			expDesc := "Directional Trend Index " + tt.exp
			if meta := dti.Metadata(); meta.Description != expDesc {
				t.Errorf("description: expected '%s', got '%s'", expDesc, meta.Description)
			}
		})
	}
}

func TestDirectionalTrendIndexMetadata(t *testing.T) {
	t.Parallel()

	dti, err := NewDirectionalTrendIndex(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	meta := dti.Metadata()

	if meta.Identifier != core.DirectionalTrendIndex {
		t.Errorf("identifier: expected DirectionalTrendIndex, got %v", meta.Identifier)
	}

	if meta.Mnemonic != "dti(2,20,5,3,3)" {
		t.Errorf("mnemonic: expected 'dti(2,20,5,3,3)', got '%s'", meta.Mnemonic)
	}

	if len(meta.Outputs) != 2 {
		t.Fatalf("outputs: expected 2, got %d", len(meta.Outputs))
	}

	if act := meta.Outputs[0].Mnemonic; act != "dti(2,20,5,3,3) dti" {
		t.Errorf("outputs[0].Mnemonic: expected 'dti(2,20,5,3,3) dti', got '%s'", act)
	}

	if act := meta.Outputs[0].Description; act != "Directional Trend Index dti(2,20,5,3,3) DTI" {
		t.Errorf("outputs[0].Description: expected 'Directional Trend Index dti(2,20,5,3,3) DTI', got '%s'", act)
	}

	if act := meta.Outputs[1].Mnemonic; act != "dti(2,20,5,3,3) signal" {
		t.Errorf("outputs[1].Mnemonic: expected 'dti(2,20,5,3,3) signal', got '%s'", act)
	}

	if act := meta.Outputs[1].Description; act != "Directional Trend Index dti(2,20,5,3,3) signal" {
		t.Errorf("outputs[1].Description: expected 'Directional Trend Index dti(2,20,5,3,3) signal', got '%s'", act)
	}
}

func TestDirectionalTrendIndexInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params *Params
	}{
		{"q too small", &Params{Q: -1, R: 20, S: 5, U: 3, Ul: 3}},
		{"r too small", &Params{Q: 2, R: -1, S: 5, U: 3, Ul: 3}},
		{"s too small", &Params{Q: 2, R: 20, S: -1, U: 3, Ul: 3}},
		{"u too small", &Params{Q: 2, R: 20, S: 5, U: -1, Ul: 3}},
		{"ul too small", &Params{Q: 2, R: 20, S: 5, U: 3, Ul: -1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := NewDirectionalTrendIndex(tt.params); err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}
