//nolint:testpackage
package truestrengthindex

import (
	"math"
	"testing"

	"zpano/indicators/core"
)

const tolerance = 1e-9

// ul is the signal-line EMA period used for every expected signal array.
const ul = 3

type tsiCombo struct {
	name      string
	q         int
	r         int
	s         int
	u         int
	expTsi    []float64
	expSignal []float64
}

func tsiCombos() []tsiCombo {
	return []tsiCombo{
		{"Q2_R20_S5_U3", 2, 20, 5, 3, expectedQ2_R20_S5_U3, expectedQ2_R20_S5_U3_SIG_UL3},
		{"Q2_R25_S13_U1", 2, 25, 13, 1, expectedQ2_R25_S13_U1, expectedQ2_R25_S13_U1_SIG_UL3},
		{"Q2_R20_S5_U1", 2, 20, 5, 1, expectedQ2_R20_S5_U1, expectedQ2_R20_S5_U1_SIG_UL3},
		{"Q2_R32_S5_U1", 2, 32, 5, 1, expectedQ2_R32_S5_U1, expectedQ2_R32_S5_U1_SIG_UL3},
		{"Q2_R13_S13_U1", 2, 13, 13, 1, expectedQ2_R13_S13_U1, expectedQ2_R13_S13_U1_SIG_UL3},
		{"Q2_R20_S40_U1", 2, 20, 40, 1, expectedQ2_R20_S40_U1, expectedQ2_R20_S40_U1_SIG_UL3},
		{"Q2_R40_S20_U1", 2, 40, 20, 1, expectedQ2_R40_S20_U1, expectedQ2_R40_S20_U1_SIG_UL3},
		{"Q2_R64_S64_U1", 2, 64, 64, 1, expectedQ2_R64_S64_U1, expectedQ2_R64_S64_U1_SIG_UL3},
		{"Q2_R100_S5_U1", 2, 100, 5, 1, expectedQ2_R100_S5_U1, expectedQ2_R100_S5_U1_SIG_UL3},
		{"Q2_R1_S1_U1", 2, 1, 1, 1, expectedQ2_R1_S1_U1, expectedQ2_R1_S1_U1_SIG_UL3},
		{"Q2_R1_S5_U3", 2, 1, 5, 3, expectedQ2_R1_S5_U3, expectedQ2_R1_S5_U3_SIG_UL3},
		{"Q2_R20_S1_U1", 2, 20, 1, 1, expectedQ2_R20_S1_U1, expectedQ2_R20_S1_U1_SIG_UL3},
		{"Q2_R5_S5_U5", 2, 5, 5, 5, expectedQ2_R5_S5_U5, expectedQ2_R5_S5_U5_SIG_UL3},
		{"Q3_R20_S5_U3", 3, 20, 5, 3, expectedQ3_R20_S5_U3, expectedQ3_R20_S5_U3_SIG_UL3},
		{"Q5_R20_S5_U3", 5, 20, 5, 3, expectedQ5_R20_S5_U3, expectedQ5_R20_S5_U3_SIG_UL3},
		{"Q10_R20_S5_U1", 10, 20, 5, 1, expectedQ10_R20_S5_U1, expectedQ10_R20_S5_U1_SIG_UL3},
		{"Q2_R9_S3_U1", 2, 9, 3, 1, expectedQ2_R9_S3_U1, expectedQ2_R9_S3_U1_SIG_UL3},
		{"Q2_R7_S4_U2", 2, 7, 4, 2, expectedQ2_R7_S4_U2, expectedQ2_R7_S4_U2_SIG_UL3},
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

func TestTrueStrengthIndexData(t *testing.T) {
	t.Parallel()

	for _, combo := range tsiCombos() {
		combo := combo
		t.Run(combo.name, func(t *testing.T) {
			t.Parallel()

			tsi, err := NewTrueStrengthIndex(&Params{
				Q: combo.q, R: combo.r, S: combo.s, U: combo.u, Ul: ul,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for i := 0; i < len(testInput); i++ {
				tsiVal, signalVal := tsi.Update(testInput[i])

				checkVal(t, "tsi", i, combo.expTsi[i], tsiVal)
				checkVal(t, "signal", i, combo.expSignal[i], signalVal)
			}
		})
	}
}

func TestTrueStrengthIndexMnemonic(t *testing.T) {
	t.Parallel()

	tsi, err := NewTrueStrengthIndex(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tsi.mnemonic != "tsi(2,20,5,3)" {
		t.Errorf("mnemonic: expected 'tsi(2,20,5,3)', got '%s'", tsi.mnemonic)
	}
}

func TestTrueStrengthIndexMetadata(t *testing.T) {
	t.Parallel()

	tsi, err := NewTrueStrengthIndex(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	meta := tsi.Metadata()

	if meta.Identifier != core.TrueStrengthIndex {
		t.Errorf("identifier: expected TrueStrengthIndex, got %v", meta.Identifier)
	}

	if meta.Mnemonic != "tsi(2,20,5,3)" {
		t.Errorf("mnemonic: expected 'tsi(2,20,5,3)', got '%s'", meta.Mnemonic)
	}

	if len(meta.Outputs) != 2 {
		t.Errorf("outputs: expected 2, got %d", len(meta.Outputs))
	}
}

func TestTrueStrengthIndexInvalidParams(t *testing.T) {
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := NewTrueStrengthIndex(tt.params); err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}
