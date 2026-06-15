//nolint:testpackage
package adaptiveexponentialmovingaverage

import (
	"math"
	"testing"

	"zpano/indicators/core"
)

const tolerance = 1e-9

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

func TestAdaptiveExponentialMovingAverageValue(t *testing.T) {
	t.Parallel()

	combos := []struct {
		name     string
		alphaMax float64
		alphaMin float64
		omega0   float64
		smooth   int
		expected []float64
	}{
		{"DEFAULT", 0.5, 0.05, 1.0, 3, expectedDEFAULT},
		{"A0_8_A0_02", 0.8, 0.02, 1.0, 3, expectedA0_8_A0_02},
		{"W0_5", 0.5, 0.05, 0.5, 3, expectedW0_5},
		{"W1_5", 0.5, 0.05, 1.5, 3, expectedW1_5},
		{"S0", 0.5, 0.05, 1.0, 0, expectedS0},
		{"S6", 0.5, 0.05, 1.0, 6, expectedS6},
	}

	for _, combo := range combos {
		combo := combo
		t.Run(combo.name, func(t *testing.T) {
			t.Parallel()

			aema, err := NewAdaptiveExponentialMovingAverage(&Params{
				AlphaMax: combo.alphaMax, AlphaMin: combo.alphaMin, Omega0: combo.omega0, Smoothing: combo.smooth,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for i := 0; i < len(testInput); i++ {
				value, _, _ := aema.Update(testInput[i])
				checkVal(t, "value", i, combo.expected[i], value)
			}
		})
	}
}

func TestAdaptiveExponentialMovingAverageOmegaAlpha(t *testing.T) {
	t.Parallel()

	aema, err := NewAdaptiveExponentialMovingAverage(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < len(testInput); i++ {
		_, omega, alpha := aema.Update(testInput[i])
		checkVal(t, "omega", i, expectedDEFAULT_OMEGA[i], omega)
		checkVal(t, "alpha", i, expectedDEFAULT_ALPHA[i], alpha)
	}
}

func TestAdaptiveExponentialMovingAverageSine(t *testing.T) {
	t.Parallel()

	aema, err := NewAdaptiveExponentialMovingAverage(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < len(test1InputSine); i++ {
		value, omega, alpha := aema.Update(test1InputSine[i])
		checkVal(t, "value", i, test1Expected[i], value)
		checkVal(t, "omega", i, test1ExpectedOmega[i], omega)
		checkVal(t, "alpha", i, test1ExpectedAlpha[i], alpha)
	}
}

func TestAdaptiveExponentialMovingAverageMnemonic(t *testing.T) {
	t.Parallel()

	aema, err := NewAdaptiveExponentialMovingAverage(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if aema.mnemonic != "aema(0.50,0.05,1.00,3)" {
		t.Errorf("mnemonic: expected 'aema(0.50,0.05,1.00,3)', got '%s'", aema.mnemonic)
	}
}

func TestAdaptiveExponentialMovingAverageMetadata(t *testing.T) {
	t.Parallel()

	aema, err := NewAdaptiveExponentialMovingAverage(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	meta := aema.Metadata()

	if meta.Identifier != core.AdaptiveExponentialMovingAverage {
		t.Errorf("identifier: expected AdaptiveExponentialMovingAverage, got %v", meta.Identifier)
	}

	if meta.Mnemonic != "aema(0.50,0.05,1.00,3)" {
		t.Errorf("mnemonic: expected 'aema(0.50,0.05,1.00,3)', got '%s'", meta.Mnemonic)
	}

	if len(meta.Outputs) != 3 {
		t.Errorf("outputs: expected 3, got %d", len(meta.Outputs))
	}
}

func TestAdaptiveExponentialMovingAverageInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params *Params
	}{
		{"alpha order", &Params{AlphaMax: 0.05, AlphaMin: 0.5, Omega0: 1.0, Smoothing: 3}},
		{"alpha max too large", &Params{AlphaMax: 1.5, AlphaMin: 0.05, Omega0: 1.0, Smoothing: 3}},
		{"omega0 too large", &Params{AlphaMax: 0.5, AlphaMin: 0.05, Omega0: 4.0, Smoothing: 3}},
		{"smoothing negative", &Params{AlphaMax: 0.5, AlphaMin: 0.05, Omega0: 1.0, Smoothing: -1}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := NewAdaptiveExponentialMovingAverage(tt.params); err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}
