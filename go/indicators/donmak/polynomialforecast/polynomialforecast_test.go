//nolint:testpackage
package polynomialforecast

import (
	"math"
	"testing"

	"zpano/indicators/core"
)

const tolerance = 1e-9

type pofCombo struct {
	degree    int
	order     int
	smoothing int
	expected  []float64
}

func pofCombos() []pofCombo {
	return []pofCombo{
		{2, 1, 0, expectedD2O1S0}, {2, 1, 3, expectedD2O1S3}, {2, 1, 6, expectedD2O1S6},
		{2, 2, 0, expectedD2O2S0}, {2, 2, 3, expectedD2O2S3}, {2, 2, 6, expectedD2O2S6},
		{3, 1, 0, expectedD3O1S0}, {3, 1, 3, expectedD3O1S3}, {3, 1, 6, expectedD3O1S6},
		{3, 2, 0, expectedD3O2S0}, {3, 2, 3, expectedD3O2S3}, {3, 2, 6, expectedD3O2S6},
		{4, 1, 0, expectedD4O1S0}, {4, 1, 3, expectedD4O1S3}, {4, 1, 6, expectedD4O1S6},
		{4, 2, 0, expectedD4O2S0}, {4, 2, 3, expectedD4O2S3}, {4, 2, 6, expectedD4O2S6},
		{5, 1, 0, expectedD5O1S0}, {5, 1, 3, expectedD5O1S3}, {5, 1, 6, expectedD5O1S6},
		{5, 2, 0, expectedD5O2S0}, {5, 2, 3, expectedD5O2S3}, {5, 2, 6, expectedD5O2S6},
	}
}

func checkSeries(t *testing.T, name string, params *Params, inputs, expected []float64) {
	t.Helper()

	pof, err := NewPolynomialForecast(params)
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", name, err)
	}

	if len(inputs) != len(expected) {
		t.Fatalf("%s: length mismatch", name)
	}

	for i := 0; i < len(inputs); i++ {
		value := pof.Update(inputs[i])
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

func TestPolynomialForecastData(t *testing.T) {
	t.Parallel()

	for _, combo := range pofCombos() {
		combo := combo
		params := &Params{Degree: combo.degree, Order: combo.order, Smoothing: combo.smoothing}
		pof, _ := NewPolynomialForecast(params)
		name := pof.LineIndicator.Mnemonic

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			checkSeries(t, name, params, inputClose, combo.expected)
		})
	}

	checkSeries(t, "TEST1_O1", &Params{Degree: 3, Order: 1, Smoothing: 0}, test1InputLinear, test1ExpectedD3O1S0)
	checkSeries(t, "TEST1_O2", &Params{Degree: 3, Order: 2, Smoothing: 0}, test1InputLinear, test1ExpectedD3O2S0)
}

func TestPolynomialForecastMnemonic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		params *Params
		want   string
	}{
		{DefaultParams(), "pof(3,1,0)"},
		{&Params{Degree: 5, Order: 2, Smoothing: 6}, "pof(5,2,6)"},
	}

	for _, tt := range tests {
		pof, err := NewPolynomialForecast(tt.params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if pof.LineIndicator.Mnemonic != tt.want {
			t.Errorf("mnemonic: expected '%s', got '%s'", tt.want, pof.LineIndicator.Mnemonic)
		}
	}
}

func TestPolynomialForecastMetadata(t *testing.T) {
	t.Parallel()

	pof, err := NewPolynomialForecast(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	meta := pof.Metadata()

	if meta.Identifier != core.PolynomialForecast {
		t.Errorf("identifier: expected PolynomialForecast, got %v", meta.Identifier)
	}

	if meta.Mnemonic != "pof(3,1,0)" {
		t.Errorf("mnemonic: expected 'pof(3,1,0)', got '%s'", meta.Mnemonic)
	}

	if len(meta.Outputs) != 1 {
		t.Errorf("outputs: expected 1, got %d", len(meta.Outputs))
	}
}

func TestPolynomialForecastInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params *Params
	}{
		{"degree too small", &Params{Degree: 1, Order: 1}},
		{"order too small", &Params{Degree: 3, Order: -1}},
		{"order too large", &Params{Degree: 3, Order: 3}},
		{"smoothing negative", &Params{Degree: 3, Order: 1, Smoothing: -1}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := NewPolynomialForecast(tt.params); err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}
