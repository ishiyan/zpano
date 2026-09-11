//nolint:testpackage
package macdindex

import (
	"math"
	"testing"

	"zpano/indicators/core"
)

const tolerance = 1e-9

// ul is the signal-line EMA period used for every expected signal array.
const ul = 3

type macdiCombo struct {
	name      string
	r         int
	s         int
	u         int
	expMacdi  []float64
	expSignal []float64
}

func macdiCombos() []macdiCombo {
	return []macdiCombo{
		{"R20_S5_U3", 20, 5, 3, expectedR20_S5_U3, expectedR20_S5_U3_SIG_UL3},
		{"R20_S5_U1", 20, 5, 1, expectedR20_S5_U1, expectedR20_S5_U1_SIG_UL3},
		{"R26_S12_U3", 26, 12, 3, expectedR26_S12_U3, expectedR26_S12_U3_SIG_UL3},
		{"R26_S12_U1", 26, 12, 1, expectedR26_S12_U1, expectedR26_S12_U1_SIG_UL3},
		{"R35_S5_U3", 35, 5, 3, expectedR35_S5_U3, expectedR35_S5_U3_SIG_UL3},
		{"R10_S3_U5", 10, 3, 5, expectedR10_S3_U5, expectedR10_S3_U5_SIG_UL3},
		{"R32_S12_U5", 32, 12, 5, expectedR32_S12_U5, expectedR32_S12_U5_SIG_UL3},
		{"R17_S8_U1", 17, 8, 1, expectedR17_S8_U1, expectedR17_S8_U1_SIG_UL3},
		{"R20_S10_U3", 20, 10, 3, expectedR20_S10_U3, expectedR20_S10_U3_SIG_UL3},
		{"R8_S4_U2", 8, 4, 2, expectedR8_S4_U2, expectedR8_S4_U2_SIG_UL3},
		{"R30_S15_U1", 30, 15, 1, expectedR30_S15_U1, expectedR30_S15_U1_SIG_UL3},
		{"R3_S2_U3", 3, 2, 3, expectedR3_S2_U3, expectedR3_S2_U3_SIG_UL3},
		{"R50_S12_U1", 50, 12, 1, expectedR50_S12_U1, expectedR50_S12_U1_SIG_UL3},
		{"R19_S6_U3", 19, 6, 3, expectedR19_S6_U3, expectedR19_S6_U3_SIG_UL3},
		{"R20_S5_U5", 20, 5, 5, expectedR20_S5_U5, expectedR20_S5_U5_SIG_UL3},
		{"R60_S30_U10", 60, 30, 10, expectedR60_S30_U10, expectedR60_S30_U10_SIG_UL3},
	}
}

func checkVal(t *testing.T, name string, i int, exp, act float64) {
	t.Helper()

	if math.Abs(act-exp) > tolerance {
		t.Errorf("%s[%d]: expected %v, got %v", name, i, exp, act)
	}
}

func TestMacdIndexData(t *testing.T) {
	t.Parallel()

	for _, combo := range macdiCombos() {
		combo := combo
		t.Run(combo.name, func(t *testing.T) {
			t.Parallel()

			macdi, err := NewMacdIndex(&Params{
				R: combo.r, S: combo.s, U: combo.u, Ul: ul,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for i := 0; i < len(testInput); i++ {
				macdiValue, signalValue := macdi.Update(testInput[i])

				checkVal(t, "macdi", i, combo.expMacdi[i], macdiValue)
				checkVal(t, "signal", i, combo.expSignal[i], signalValue)
			}
		})
	}
}

func TestMacdIndexNoWarmUp(t *testing.T) {
	t.Parallel()

	macdi, err := NewMacdIndex(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	macdiValue, signalValue := macdi.Update(testInput[0])

	if macdiValue != 0.0 {
		t.Errorf("macdi[0]: expected exactly 0, got %v", macdiValue)
	}

	if signalValue != 0.0 {
		t.Errorf("signal[0]: expected exactly 0, got %v", signalValue)
	}

	for i := 1; i < len(testInput); i++ {
		macdiValue, signalValue = macdi.Update(testInput[i])

		if math.IsNaN(macdiValue) {
			t.Errorf("macdi[%d]: unexpected NaN", i)
		}

		if math.IsNaN(signalValue) {
			t.Errorf("signal[%d]: unexpected NaN", i)
		}
	}
}

func TestMacdIndexPassthrough(t *testing.T) {
	t.Parallel()

	macdi, err := NewMacdIndex(&Params{R: 2, S: 1, U: 1, Ul: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Bar 0 seeds both price EMAs, so the MACD line is exactly 0.
	macdiValue, signalValue := macdi.Update(10.0)
	checkVal(t, "macdi", 0, 0.0, macdiValue)
	checkVal(t, "signal", 0, 0.0, signalValue)

	// EMA(1) is a passthrough -> fast = 12. EMA(2) = (2/3)*12 + (1/3)*10 = 11.3333.
	macdiValue, signalValue = macdi.Update(12.0)
	checkVal(t, "macdi", 1, 12.0-(2.0/3.0*12.0+1.0/3.0*10.0), macdiValue)
	checkVal(t, "signal", 1, macdiValue, signalValue)
}

func TestMacdIndexSignalPassthrough(t *testing.T) {
	t.Parallel()

	macdi, err := NewMacdIndex(&Params{R: 20, S: 5, U: 3, Ul: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < len(testInput); i++ {
		macdiValue, signalValue := macdi.Update(testInput[i])

		if macdiValue != signalValue {
			t.Errorf("signal[%d]: expected %v (== macdi) when ul=1, got %v", i, macdiValue, signalValue)
		}
	}
}

func TestMacdIndexIsPrimed(t *testing.T) {
	t.Parallel()

	macdi, err := NewMacdIndex(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if macdi.IsPrimed() {
		t.Errorf("expected not primed before any update")
	}

	macdi.Update(testInput[0])

	if !macdi.IsPrimed() {
		t.Errorf("expected primed after the first update")
	}
}

func TestMacdIndexMnemonic(t *testing.T) {
	t.Parallel()

	macdi, err := NewMacdIndex(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if macdi.mnemonic != "macdi(20,5,3)" {
		t.Errorf("mnemonic: expected 'macdi(20,5,3)', got '%s'", macdi.mnemonic)
	}
}

func TestMacdIndexMetadata(t *testing.T) {
	t.Parallel()

	macdi, err := NewMacdIndex(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	meta := macdi.Metadata()

	if meta.Identifier != core.MacdIndex {
		t.Errorf("identifier: expected MacdIndex, got %v", meta.Identifier)
	}

	if meta.Mnemonic != "macdi(20,5,3)" {
		t.Errorf("mnemonic: expected 'macdi(20,5,3)', got '%s'", meta.Mnemonic)
	}

	if len(meta.Outputs) != 2 {
		t.Errorf("outputs: expected 2, got %d", len(meta.Outputs))
	}
}

func TestMacdIndexInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params *Params
	}{
		{"r too small", &Params{R: -1, S: 5, U: 3, Ul: 3}},
		{"s too small", &Params{R: 20, S: -1, U: 3, Ul: 3}},
		{"u too small", &Params{R: 20, S: 5, U: -1, Ul: 3}},
		{"ul too small", &Params{R: 20, S: 5, U: 3, Ul: -1}},
		{"s not less than r", &Params{R: 5, S: 5, U: 3, Ul: 3}},
		{"s greater than r", &Params{R: 5, S: 6, U: 3, Ul: 3}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := NewMacdIndex(tt.params); err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}
