//nolint:testpackage
package ergodicoscillator

import (
	"math"
	"testing"

	"zpano/indicators/core"
)

const tolerance = 1e-9

type ergodicCombo struct {
	name       string
	q          int
	r          int
	s          int
	u          int
	ul         int
	expErgodic []float64
	expSignal  []float64
}

func ergodicCombos() []ergodicCombo {
	return []ergodicCombo{
		{"Q2_R20_S5_U3_L3", 2, 20, 5, 3, 3, expectedErgQ2_R20_S5_U3_L3, expectedSigQ2_R20_S5_U3_L3},
		{"Q2_R32_S5_U1_L5", 2, 32, 5, 1, 5, expectedErgQ2_R32_S5_U1_L5, expectedSigQ2_R32_S5_U1_L5},
		{"Q2_R20_S5_U1_L5", 2, 20, 5, 1, 5, expectedErgQ2_R20_S5_U1_L5, expectedSigQ2_R20_S5_U1_L5},
		{"Q2_R32_S5_U1_L7", 2, 32, 5, 1, 7, expectedErgQ2_R32_S5_U1_L7, expectedSigQ2_R32_S5_U1_L7},
		{"Q2_R25_S13_U1_L5", 2, 25, 13, 1, 5, expectedErgQ2_R25_S13_U1_L5, expectedSigQ2_R25_S13_U1_L5},
		{"Q2_R20_S5_U3_L1", 2, 20, 5, 3, 1, expectedErgQ2_R20_S5_U3_L1, expectedSigQ2_R20_S5_U3_L1},
		{"Q2_R1_S1_U1_L1", 2, 1, 1, 1, 1, expectedErgQ2_R1_S1_U1_L1, expectedSigQ2_R1_S1_U1_L1},
		{"Q2_R20_S5_U3_L9", 2, 20, 5, 3, 9, expectedErgQ2_R20_S5_U3_L9, expectedSigQ2_R20_S5_U3_L9},
		{"Q2_R64_S64_U1_L5", 2, 64, 64, 1, 5, expectedErgQ2_R64_S64_U1_L5, expectedSigQ2_R64_S64_U1_L5},
		{"Q2_R9_S3_U1_L3", 2, 9, 3, 1, 3, expectedErgQ2_R9_S3_U1_L3, expectedSigQ2_R9_S3_U1_L3},
		{"Q3_R20_S5_U3_L3", 3, 20, 5, 3, 3, expectedErgQ3_R20_S5_U3_L3, expectedSigQ3_R20_S5_U3_L3},
		{"Q5_R20_S5_U3_L5", 5, 20, 5, 3, 5, expectedErgQ5_R20_S5_U3_L5, expectedSigQ5_R20_S5_U3_L5},
		{"Q2_R13_S7_U1_L7", 2, 13, 7, 1, 7, expectedErgQ2_R13_S7_U1_L7, expectedSigQ2_R13_S7_U1_L7},
		{"Q2_R20_S5_U1_L3", 2, 20, 5, 1, 3, expectedErgQ2_R20_S5_U1_L3, expectedSigQ2_R20_S5_U1_L3},
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

func TestErgodicOscillatorData(t *testing.T) {
	t.Parallel()

	for _, combo := range ergodicCombos() {
		combo := combo
		t.Run(combo.name, func(t *testing.T) {
			t.Parallel()

			erg, err := NewErgodicOscillator(&Params{
				Q: combo.q, R: combo.r, S: combo.s, U: combo.u, Ul: combo.ul,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for i := 0; i < len(testInput); i++ {
				ergodicValue, signalValue := erg.Update(testInput[i])

				checkVal(t, "ergodic", i, combo.expErgodic[i], ergodicValue)
				checkVal(t, "signal", i, combo.expSignal[i], signalValue)
			}
		})
	}
}

func TestErgodicOscillatorSignalPassthrough(t *testing.T) {
	t.Parallel()

	erg, err := NewErgodicOscillator(&Params{Q: 2, R: 20, S: 5, U: 3, Ul: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < len(testInput); i++ {
		ergodicValue, signalValue := erg.Update(testInput[i])

		if math.IsNaN(ergodicValue) {
			if !math.IsNaN(signalValue) {
				t.Errorf("signal[%d]: expected NaN, got %v", i, signalValue)
			}

			continue
		}

		if ergodicValue != signalValue {
			t.Errorf("signal[%d]: expected %v (== ergodic) when ul=1, got %v", i, ergodicValue, signalValue)
		}
	}
}

func TestErgodicOscillatorIsPrimed(t *testing.T) {
	t.Parallel()

	erg, err := NewErgodicOscillator(&Params{Q: 3, R: 20, S: 5, U: 3, Ul: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if erg.IsPrimed() {
		t.Errorf("expected not primed before any update")
	}

	erg.Update(testInput[0])

	if erg.IsPrimed() {
		t.Errorf("expected not primed after 1 update with q=3")
	}

	erg.Update(testInput[1])

	if erg.IsPrimed() {
		t.Errorf("expected not primed after 2 updates with q=3")
	}

	erg.Update(testInput[2])

	if !erg.IsPrimed() {
		t.Errorf("expected primed after 3 updates with q=3")
	}
}

func TestErgodicOscillatorMnemonic(t *testing.T) {
	t.Parallel()

	erg, err := NewErgodicOscillator(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if erg.mnemonic != "ergodic(2,20,5,3,3)" {
		t.Errorf("mnemonic: expected 'ergodic(2,20,5,3,3)', got '%s'", erg.mnemonic)
	}
}

func TestErgodicOscillatorMetadata(t *testing.T) {
	t.Parallel()

	erg, err := NewErgodicOscillator(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	meta := erg.Metadata()

	if meta.Identifier != core.ErgodicOscillator {
		t.Errorf("identifier: expected ErgodicOscillator, got %v", meta.Identifier)
	}

	if meta.Mnemonic != "ergodic(2,20,5,3,3)" {
		t.Errorf("mnemonic: expected 'ergodic(2,20,5,3,3)', got '%s'", meta.Mnemonic)
	}

	if len(meta.Outputs) != 2 {
		t.Errorf("outputs: expected 2, got %d", len(meta.Outputs))
	}
}

func TestErgodicOscillatorInvalidParams(t *testing.T) {
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

			if _, err := NewErgodicOscillator(tt.params); err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}
