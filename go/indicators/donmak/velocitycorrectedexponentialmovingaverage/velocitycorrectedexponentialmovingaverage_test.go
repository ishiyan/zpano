//nolint:testpackage
package velocitycorrectedexponentialmovingaverage

import (
	"math"
	"testing"

	"zpano/indicators/core"
)

const tolerance = 1e-9

type vcemaCombo struct {
	period   int
	degree   int
	expected []float64
}

func vcemaCombos() []vcemaCombo {
	return []vcemaCombo{
		{3, 2, expectedP3D2}, {3, 3, expectedP3D3}, {3, 4, expectedP3D4}, {3, 5, expectedP3D5},
		{6, 2, expectedP6D2}, {6, 3, expectedP6D3}, {6, 4, expectedP6D4}, {6, 5, expectedP6D5},
		{12, 2, expectedP12D2}, {12, 3, expectedP12D3}, {12, 4, expectedP12D4}, {12, 5, expectedP12D5},
	}
}

func checkSeries(t *testing.T, name string, params *Params, inputs, expected []float64) {
	t.Helper()

	vcema, err := NewVelocityCorrectedExponentialMovingAverage(params)
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", name, err)
	}

	if len(inputs) != len(expected) {
		t.Fatalf("%s: length mismatch", name)
	}

	for i := 0; i < len(inputs); i++ {
		value := vcema.Update(inputs[i])
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

func TestVelocityCorrectedExponentialMovingAverageData(t *testing.T) {
	t.Parallel()

	for _, combo := range vcemaCombos() {
		combo := combo
		params := &Params{Period: combo.period, Degree: combo.degree}
		vcema, _ := NewVelocityCorrectedExponentialMovingAverage(params)
		name := vcema.LineIndicator.Mnemonic

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			checkSeries(t, name, params, inputClose, combo.expected)
		})
	}

	checkSeries(t, "TEST1", &Params{Period: 6, Degree: 3}, test1InputLinear, test1ExpectedP6D3)
}

func TestVelocityCorrectedExponentialMovingAverageMnemonic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		params *Params
		want   string
	}{
		{DefaultParams(), "vcema(6,3)"},
		{&Params{Period: 12, Degree: 5}, "vcema(12,5)"},
	}

	for _, tt := range tests {
		vcema, err := NewVelocityCorrectedExponentialMovingAverage(tt.params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if vcema.LineIndicator.Mnemonic != tt.want {
			t.Errorf("mnemonic: expected '%s', got '%s'", tt.want, vcema.LineIndicator.Mnemonic)
		}
	}
}

func TestVelocityCorrectedExponentialMovingAverageMetadata(t *testing.T) {
	t.Parallel()

	vcema, err := NewVelocityCorrectedExponentialMovingAverage(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	meta := vcema.Metadata()

	if meta.Identifier != core.VelocityCorrectedExponentialMovingAverage {
		t.Errorf("identifier: expected VelocityCorrectedExponentialMovingAverage, got %v", meta.Identifier)
	}

	if meta.Mnemonic != "vcema(6,3)" {
		t.Errorf("mnemonic: expected 'vcema(6,3)', got '%s'", meta.Mnemonic)
	}

	if len(meta.Outputs) != 1 {
		t.Errorf("outputs: expected 1, got %d", len(meta.Outputs))
	}
}

func TestVelocityCorrectedExponentialMovingAverageInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params *Params
	}{
		{"period too small", &Params{Period: 1, Degree: 3}},
		{"degree too small", &Params{Period: 6, Degree: 1}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := NewVelocityCorrectedExponentialMovingAverage(tt.params); err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}
