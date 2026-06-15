//nolint:testpackage
package instantaneoussinewaveperiod

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

func TestInstantaneousSineWavePeriodData(t *testing.T) {
	t.Parallel()

	combos := []struct {
		name         string
		smoothing    int
		period       []float64
		omega        []float64
		velocity     []float64
		acceleration []float64
	}{
		{"S0", 0, expectedS0_PERIOD, expectedS0_OMEGA, expectedS0_VELOCITY, expectedS0_ACCELERATION},
		{"S3", 3, expectedS3_PERIOD, expectedS3_OMEGA, expectedS3_VELOCITY, expectedS3_ACCELERATION},
		{"S6", 6, expectedS6_PERIOD, expectedS6_OMEGA, expectedS6_VELOCITY, expectedS6_ACCELERATION},
		{"S12", 12, expectedS12_PERIOD, expectedS12_OMEGA, expectedS12_VELOCITY, expectedS12_ACCELERATION},
	}

	for _, combo := range combos {
		combo := combo
		t.Run(combo.name, func(t *testing.T) {
			t.Parallel()

			iswp, err := NewInstantaneousSineWavePeriod(&Params{Smoothing: combo.smoothing})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for i := 0; i < len(testInput); i++ {
				period, omega, velocity, acceleration, _, _, _ := iswp.Update(testInput[i])
				checkVal(t, "period", i, combo.period[i], period)
				checkVal(t, "omega", i, combo.omega[i], omega)
				checkVal(t, "velocity", i, combo.velocity[i], velocity)
				checkVal(t, "acceleration", i, combo.acceleration[i], acceleration)
			}
		})
	}
}

func TestInstantaneousSineWavePeriodMnemonic(t *testing.T) {
	t.Parallel()

	iswp, err := NewInstantaneousSineWavePeriod(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if iswp.mnemonic != "iswp(0,4.00,50.00,20.00,0.01)" {
		t.Errorf("mnemonic: expected 'iswp(0,4.00,50.00,20.00,0.01)', got '%s'", iswp.mnemonic)
	}
}

func TestInstantaneousSineWavePeriodMetadata(t *testing.T) {
	t.Parallel()

	iswp, err := NewInstantaneousSineWavePeriod(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	meta := iswp.Metadata()

	if meta.Identifier != core.InstantaneousSineWavePeriod {
		t.Errorf("identifier: expected InstantaneousSineWavePeriod, got %v", meta.Identifier)
	}

	if meta.Mnemonic != "iswp(0,4.00,50.00,20.00,0.01)" {
		t.Errorf("mnemonic: expected 'iswp(0,4.00,50.00,20.00,0.01)', got '%s'", meta.Mnemonic)
	}

	if len(meta.Outputs) != 7 {
		t.Errorf("outputs: expected 7, got %d", len(meta.Outputs))
	}
}

func TestInstantaneousSineWavePeriodInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params *Params
	}{
		{"smoothing negative", &Params{Smoothing: -1, MinPeriod: 4, MaxPeriod: 50, ErrorThreshold: 20, Dx: 0.01}},
		{"min period zero", &Params{Smoothing: 0, MinPeriod: -1, MaxPeriod: 50, ErrorThreshold: 20, Dx: 0.01}},
		{"max le min", &Params{Smoothing: 0, MinPeriod: 50, MaxPeriod: 50, ErrorThreshold: 20, Dx: 0.01}},
		{"error threshold negative", &Params{Smoothing: 0, MinPeriod: 4, MaxPeriod: 50, ErrorThreshold: -1, Dx: 0.01}},
		{"dx negative", &Params{Smoothing: 0, MinPeriod: 4, MaxPeriod: 50, ErrorThreshold: 20, Dx: -1}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := NewInstantaneousSineWavePeriod(tt.params); err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}
