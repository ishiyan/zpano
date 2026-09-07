//nolint:testpackage
package meandeviationindex

import (
	"math"
	"testing"

	"zpano/indicators/core"
)

const tolerance = 1e-9

// ul is the signal-line EMA period used for every expected signal array.
const ul = 3

type mdiCombo struct {
	name      string
	r         int
	s         int
	u         int
	expMdi    []float64
	expSignal []float64
}

func mdiCombos() []mdiCombo {
	return []mdiCombo{
		{"R20_S5_U3", 20, 5, 3, expectedR20_S5_U3, expectedR20_S5_U3_SIG_UL3},
		{"R20_S5_U1", 20, 5, 1, expectedR20_S5_U1, expectedR20_S5_U1_SIG_UL3},
		{"R1_S5_U3", 1, 5, 3, expectedR1_S5_U3, expectedR1_S5_U3_SIG_UL3},
		{"R40_S5_U3", 40, 5, 3, expectedR40_S5_U3, expectedR40_S5_U3_SIG_UL3},
		{"R10_S5_U3", 10, 5, 3, expectedR10_S5_U3, expectedR10_S5_U3_SIG_UL3},
		{"R5_S5_U5", 5, 5, 5, expectedR5_S5_U5, expectedR5_S5_U5_SIG_UL3},
		{"R20_S9_U1", 20, 9, 1, expectedR20_S9_U1, expectedR20_S9_U1_SIG_UL3},
		{"R26_S12_U9", 26, 12, 9, expectedR26_S12_U9, expectedR26_S12_U9_SIG_UL3},
		{"R50_S13_U1", 50, 13, 1, expectedR50_S13_U1, expectedR50_S13_U1_SIG_UL3},
		{"R30_S5_U3", 30, 5, 3, expectedR30_S5_U3, expectedR30_S5_U3_SIG_UL3},
		{"R3_S3_U3", 3, 3, 3, expectedR3_S3_U3, expectedR3_S3_U3_SIG_UL3},
		{"R7_S4_U2", 7, 4, 2, expectedR7_S4_U2, expectedR7_S4_U2_SIG_UL3},
		{"R2_S5_U3", 2, 5, 3, expectedR2_S5_U3, expectedR2_S5_U3_SIG_UL3},
		{"R20_S1_U1", 20, 1, 1, expectedR20_S1_U1, expectedR20_S1_U1_SIG_UL3},
		{"R20_S20_U5", 20, 20, 5, expectedR20_S20_U5, expectedR20_S20_U5_SIG_UL3},
		{"R60_S30_U10", 60, 30, 10, expectedR60_S30_U10, expectedR60_S30_U10_SIG_UL3},
	}
}

func checkVal(t *testing.T, name string, i int, exp, act float64) {
	t.Helper()

	if math.Abs(act-exp) > tolerance {
		t.Errorf("%s[%d]: expected %v, got %v", name, i, exp, act)
	}
}

func TestMeanDeviationIndexData(t *testing.T) {
	t.Parallel()

	for _, combo := range mdiCombos() {
		combo := combo
		t.Run(combo.name, func(t *testing.T) {
			t.Parallel()

			mdi, err := NewMeanDeviationIndex(&Params{
				R: combo.r, S: combo.s, U: combo.u, Ul: ul,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for i := 0; i < len(testInput); i++ {
				mdiValue, signalValue := mdi.Update(testInput[i])

				checkVal(t, "mdi", i, combo.expMdi[i], mdiValue)
				checkVal(t, "signal", i, combo.expSignal[i], signalValue)
			}
		})
	}
}

func TestMeanDeviationIndexNoWarmUp(t *testing.T) {
	t.Parallel()

	mdi, err := NewMeanDeviationIndex(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mdiValue, signalValue := mdi.Update(testInput[0])

	if mdiValue != 0.0 {
		t.Errorf("mdi[0]: expected exactly 0, got %v", mdiValue)
	}

	if signalValue != 0.0 {
		t.Errorf("signal[0]: expected exactly 0, got %v", signalValue)
	}

	for i := 1; i < len(testInput); i++ {
		mdiValue, signalValue = mdi.Update(testInput[i])

		if math.IsNaN(mdiValue) {
			t.Errorf("mdi[%d]: unexpected NaN", i)
		}

		if math.IsNaN(signalValue) {
			t.Errorf("signal[%d]: unexpected NaN", i)
		}
	}
}

func TestMeanDeviationIndexDegenerate(t *testing.T) {
	t.Parallel()

	mdi, err := NewMeanDeviationIndex(&Params{R: 1, S: 5, U: 3, Ul: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < len(testInput); i++ {
		mdiValue, signalValue := mdi.Update(testInput[i])

		if mdiValue != 0.0 {
			t.Errorf("mdi[%d]: expected exactly 0 when r=1, got %v", i, mdiValue)
		}

		if signalValue != 0.0 {
			t.Errorf("signal[%d]: expected exactly 0 when r=1, got %v", i, signalValue)
		}
	}
}

func TestMeanDeviationIndexPassthrough(t *testing.T) {
	t.Parallel()

	mdi, err := NewMeanDeviationIndex(&Params{R: 2, S: 1, U: 1, Ul: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Bar 0 seeds the baseline, so the deviation is exactly 0.
	mdiValue, signalValue := mdi.Update(10.0)
	checkVal(t, "mdi", 0, 0.0, mdiValue)
	checkVal(t, "signal", 0, 0.0, signalValue)

	// EMA(2) = (2/3)*13 + (1/3)*10 = 12, so the deviation is 13-12 = 1.
	mdiValue, signalValue = mdi.Update(13.0)
	checkVal(t, "mdi", 1, 1.0, mdiValue)
	checkVal(t, "signal", 1, 1.0, signalValue)
}

func TestMeanDeviationIndexSignalPassthrough(t *testing.T) {
	t.Parallel()

	mdi, err := NewMeanDeviationIndex(&Params{R: 20, S: 5, U: 3, Ul: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < len(testInput); i++ {
		mdiValue, signalValue := mdi.Update(testInput[i])

		if mdiValue != signalValue {
			t.Errorf("signal[%d]: expected %v (== mdi) when ul=1, got %v", i, mdiValue, signalValue)
		}
	}
}

func TestMeanDeviationIndexIsPrimed(t *testing.T) {
	t.Parallel()

	mdi, err := NewMeanDeviationIndex(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mdi.IsPrimed() {
		t.Errorf("expected not primed before any update")
	}

	mdi.Update(testInput[0])

	if !mdi.IsPrimed() {
		t.Errorf("expected primed after the first update")
	}
}

func TestMeanDeviationIndexMnemonic(t *testing.T) {
	t.Parallel()

	mdi, err := NewMeanDeviationIndex(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mdi.mnemonic != "mdi(20,5,3)" {
		t.Errorf("mnemonic: expected 'mdi(20,5,3)', got '%s'", mdi.mnemonic)
	}
}

func TestMeanDeviationIndexMetadata(t *testing.T) {
	t.Parallel()

	mdi, err := NewMeanDeviationIndex(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	meta := mdi.Metadata()

	if meta.Identifier != core.MeanDeviationIndex {
		t.Errorf("identifier: expected MeanDeviationIndex, got %v", meta.Identifier)
	}

	if meta.Mnemonic != "mdi(20,5,3)" {
		t.Errorf("mnemonic: expected 'mdi(20,5,3)', got '%s'", meta.Mnemonic)
	}

	if len(meta.Outputs) != 2 {
		t.Errorf("outputs: expected 2, got %d", len(meta.Outputs))
	}
}

func TestMeanDeviationIndexInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params *Params
	}{
		{"r too small", &Params{R: -1, S: 5, U: 3, Ul: 3}},
		{"s too small", &Params{R: 20, S: -1, U: 3, Ul: 3}},
		{"u too small", &Params{R: 20, S: 5, U: -1, Ul: 3}},
		{"ul too small", &Params{R: 20, S: 5, U: 3, Ul: -1}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := NewMeanDeviationIndex(tt.params); err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}
