//nolint:testpackage
package modifiedexponentialmovingaverage

import (
	"math"
	"testing"

	"zpano/indicators/core"
)

const tolerance = 1e-9

type memaCombo struct {
	period   int
	degree   int
	skip     int
	expected []float64
}

func memaCombos() []memaCombo {
	return []memaCombo{
		{3, 3, 1, expectedP3D3Sk1}, {3, 3, 2, expectedP3D3Sk2}, {3, 3, 4, expectedP3D3Sk4},
		{3, 4, 1, expectedP3D4Sk1}, {3, 4, 2, expectedP3D4Sk2}, {3, 4, 4, expectedP3D4Sk4},
		{6, 3, 1, expectedP6D3Sk1}, {6, 3, 2, expectedP6D3Sk2}, {6, 3, 4, expectedP6D3Sk4},
		{6, 4, 1, expectedP6D4Sk1}, {6, 4, 2, expectedP6D4Sk2}, {6, 4, 4, expectedP6D4Sk4},
		{12, 3, 1, expectedP12D3Sk1}, {12, 3, 2, expectedP12D3Sk2}, {12, 3, 4, expectedP12D3Sk4},
		{12, 4, 1, expectedP12D4Sk1}, {12, 4, 2, expectedP12D4Sk2}, {12, 4, 4, expectedP12D4Sk4},
	}
}

func checkSeries(t *testing.T, name string, params *Params, inputs, expected []float64) {
	t.Helper()

	mema, err := NewModifiedExponentialMovingAverage(params)
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", name, err)
	}

	if len(inputs) != len(expected) {
		t.Fatalf("%s: length mismatch", name)
	}

	for i := 0; i < len(inputs); i++ {
		value := mema.Update(inputs[i])
		exp := expected[i]

		if math.IsNaN(exp) {
			if !math.IsNaN(value) {
				t.Errorf("%s[%d]: expected NaN, got %v", name, i, value)
			}

			continue
		}

		if math.Abs(value-exp) > tolerance {
			t.Errorf("%s[%d]: expected %v, got %v", name, i, exp, value)
		}
	}
}

func TestModifiedExponentialMovingAverageData(t *testing.T) {
	t.Parallel()

	for _, combo := range memaCombos() {
		combo := combo
		params := &Params{Period: combo.period, Degree: combo.degree, Skip: combo.skip}
		mema, _ := NewModifiedExponentialMovingAverage(params)
		name := mema.LineIndicator.Mnemonic

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			checkSeries(t, name, params, inputClose, combo.expected)
		})
	}

	checkSeries(t, "TEST1", &Params{Period: 6, Degree: 3, Skip: 1}, test1InputLinear, test1ExpectedP6D3Sk1)
}

func TestModifiedExponentialMovingAverageMnemonic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		params *Params
		want   string
	}{
		{DefaultParams(), "mema(6,3,1)"},
		{&Params{Period: 12, Degree: 4, Skip: 2}, "mema(12,4,2)"},
	}

	for _, tt := range tests {
		mema, err := NewModifiedExponentialMovingAverage(tt.params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if mema.LineIndicator.Mnemonic != tt.want {
			t.Errorf("mnemonic: expected '%s', got '%s'", tt.want, mema.LineIndicator.Mnemonic)
		}
	}
}

func TestModifiedExponentialMovingAverageMetadata(t *testing.T) {
	t.Parallel()

	mema, err := NewModifiedExponentialMovingAverage(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	meta := mema.Metadata()

	if meta.Identifier != core.ModifiedExponentialMovingAverage {
		t.Errorf("identifier: expected ModifiedExponentialMovingAverage, got %v", meta.Identifier)
	}

	if meta.Mnemonic != "mema(6,3,1)" {
		t.Errorf("mnemonic: expected 'mema(6,3,1)', got '%s'", meta.Mnemonic)
	}

	if len(meta.Outputs) != 1 {
		t.Errorf("outputs: expected 1, got %d", len(meta.Outputs))
	}
}

func TestModifiedExponentialMovingAverageInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params *Params
	}{
		{"period too small", &Params{Period: 1, Degree: 3, Skip: 1}},
		{"degree too small", &Params{Period: 6, Degree: 1, Skip: 1}},
		{"skip too small", &Params{Period: 6, Degree: 3, Skip: -1}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := NewModifiedExponentialMovingAverage(tt.params); err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}
