//nolint:testpackage
package schafftrendcycle

import (
	"math"
	"testing"

	"zpano/indicators/core"
)

const tolerance = 1e-9

type stcCombo struct {
	name     string
	fast     int
	slow     int
	tclen    int
	factor   float64
	expStc   []float64
	expMacd  []float64
	expPf    []float64
}

func stcCombos() []stcCombo {
	return []stcCombo{
		{"F23_S50_T10_C50", 23, 50, 10, 0.5, expectedStcF23_S50_T10_C50, expectedMacdF23_S50_T10_C50, expectedPfF23_S50_T10_C50},
		{"F12_S26_T10_C50", 12, 26, 10, 0.5, expectedStcF12_S26_T10_C50, expectedMacdF12_S26_T10_C50, expectedPfF12_S26_T10_C50},
		{"F5_S10_T5_C50", 5, 10, 5, 0.5, expectedStcF5_S10_T5_C50, expectedMacdF5_S10_T5_C50, expectedPfF5_S10_T5_C50},
		{"F3_S7_T3_C50", 3, 7, 3, 0.5, expectedStcF3_S7_T3_C50, nil, nil},
		{"F8_S21_T10_C50", 8, 21, 10, 0.5, expectedStcF8_S21_T10_C50, nil, nil},
		{"F10_S30_T10_C50", 10, 30, 10, 0.5, expectedStcF10_S30_T10_C50, nil, nil},
		{"F15_S40_T14_C50", 15, 40, 14, 0.5, expectedStcF15_S40_T14_C50, nil, nil},
		{"F6_S13_T8_C60", 6, 13, 8, 0.6, expectedStcF6_S13_T8_C60, nil, nil},
		{"F23_S50_T23_C50", 23, 50, 23, 0.5, expectedStcF23_S50_T23_C50, nil, nil},
		{"F23_S50_T5_C50", 23, 50, 5, 0.5, expectedStcF23_S50_T5_C50, nil, nil},
		{"F12_S26_T10_C25", 12, 26, 10, 0.25, expectedStcF12_S26_T10_C25, nil, nil},
		{"F12_S26_T10_C80", 12, 26, 10, 0.8, expectedStcF12_S26_T10_C80, nil, nil},
		{"F12_S26_T10_C100", 12, 26, 10, 1.0, expectedStcF12_S26_T10_C100, nil, nil},
		{"F20_S40_T10_C50", 20, 40, 10, 0.5, expectedStcF20_S40_T10_C50, nil, nil},
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

func TestSchaffTrendCycleData(t *testing.T) {
	t.Parallel()

	for _, combo := range stcCombos() {
		combo := combo
		t.Run(combo.name, func(t *testing.T) {
			t.Parallel()

			stc, err := NewSchaffTrendCycle(&Params{
				Fast: combo.fast, Slow: combo.slow, Tclen: combo.tclen, Factor: combo.factor,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for i := 0; i < len(testInput); i++ {
				stcVal, macdVal, pfVal := stc.Update(testInput[i])

				checkVal(t, "stc", i, combo.expStc[i], stcVal)

				if combo.expMacd != nil {
					checkVal(t, "macd", i, combo.expMacd[i], macdVal)
				}

				if combo.expPf != nil {
					checkVal(t, "pf", i, combo.expPf[i], pfVal)
				}
			}
		})
	}
}

func TestSchaffTrendCycleMnemonic(t *testing.T) {
	t.Parallel()

	stc, err := NewSchaffTrendCycle(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stc.mnemonic != "stc(23,50,10,0.50)" {
		t.Errorf("mnemonic: expected 'stc(23,50,10,0.50)', got '%s'", stc.mnemonic)
	}
}

func TestSchaffTrendCycleMetadata(t *testing.T) {
	t.Parallel()

	stc, err := NewSchaffTrendCycle(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	meta := stc.Metadata()

	if meta.Identifier != core.SchaffTrendCycle {
		t.Errorf("identifier: expected SchaffTrendCycle, got %v", meta.Identifier)
	}

	if meta.Mnemonic != "stc(23,50,10,0.50)" {
		t.Errorf("mnemonic: expected 'stc(23,50,10,0.50)', got '%s'", meta.Mnemonic)
	}

	if len(meta.Outputs) != 3 {
		t.Errorf("outputs: expected 3, got %d", len(meta.Outputs))
	}
}

func TestSchaffTrendCycleInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params *Params
	}{
		{"fast too small", &Params{Fast: -1, Slow: 50, Tclen: 10, Factor: 0.5}},
		{"slow too small", &Params{Fast: 23, Slow: -1, Tclen: 10, Factor: 0.5}},
		{"tclen too small", &Params{Fast: 23, Slow: 50, Tclen: -1, Factor: 0.5}},
		{"factor too large", &Params{Fast: 23, Slow: 50, Tclen: 10, Factor: 1.5}},
		{"factor negative", &Params{Fast: 23, Slow: 50, Tclen: 10, Factor: -0.5}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := NewSchaffTrendCycle(tt.params); err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}
