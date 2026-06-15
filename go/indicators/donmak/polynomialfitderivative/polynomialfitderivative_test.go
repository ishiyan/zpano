//nolint:testpackage
package polynomialfitderivative

import (
	"math"
	"testing"

	"zpano/indicators/core"
)

const tolerance = 1e-9

type pfdCombo struct {
	degree    int
	order     int
	smoothing int
	expected  []float64
}

func pfdCombos() []pfdCombo {
	return []pfdCombo{
		{2, 1, 0, expectedD2_O1_S0}, {2, 1, 3, expectedD2_O1_S3}, {2, 1, 6, expectedD2_O1_S6},
		{2, 2, 0, expectedD2_O2_S0}, {2, 2, 3, expectedD2_O2_S3}, {2, 2, 6, expectedD2_O2_S6},
		{3, 1, 0, expectedD3_O1_S0}, {3, 1, 3, expectedD3_O1_S3}, {3, 1, 6, expectedD3_O1_S6},
		{3, 2, 0, expectedD3_O2_S0}, {3, 2, 3, expectedD3_O2_S3}, {3, 2, 6, expectedD3_O2_S6},
		{4, 1, 0, expectedD4_O1_S0}, {4, 1, 3, expectedD4_O1_S3}, {4, 1, 6, expectedD4_O1_S6},
		{4, 2, 0, expectedD4_O2_S0}, {4, 2, 3, expectedD4_O2_S3}, {4, 2, 6, expectedD4_O2_S6},
		{5, 1, 0, expectedD5_O1_S0}, {5, 1, 3, expectedD5_O1_S3}, {5, 1, 6, expectedD5_O1_S6},
		{5, 2, 0, expectedD5_O2_S0}, {5, 2, 3, expectedD5_O2_S3}, {5, 2, 6, expectedD5_O2_S6},
		{6, 1, 0, expectedD6_O1_S0}, {6, 1, 3, expectedD6_O1_S3}, {6, 1, 6, expectedD6_O1_S6},
		{6, 2, 0, expectedD6_O2_S0}, {6, 2, 3, expectedD6_O2_S3}, {6, 2, 6, expectedD6_O2_S6},
		{4, 3, 6, expectedD4_O3_S6}, {5, 3, 6, expectedD5_O3_S6}, {6, 3, 6, expectedD6_O3_S6},
		{6, 5, 6, expectedD6_O5_S6},
	}
}

func TestPolynomialFitDerivativeData(t *testing.T) {
	t.Parallel()

	for _, combo := range pfdCombos() {
		combo := combo
		name := mnemonicName(combo)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			pfd, err := NewPolynomialFitDerivative(&Params{
				Degree: combo.degree, Order: combo.order, Smoothing: combo.smoothing,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for i := 0; i < len(testInput); i++ {
				value := pfd.Update(testInput[i])
				exp := combo.expected[i]

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
		})
	}
}

func mnemonicName(c pfdCombo) string {
	pfd, _ := NewPolynomialFitDerivative(&Params{Degree: c.degree, Order: c.order, Smoothing: c.smoothing})

	return pfd.LineIndicator.Mnemonic
}

func TestPolynomialFitDerivativeMnemonic(t *testing.T) {
	t.Parallel()

	pfd, err := NewPolynomialFitDerivative(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pfd.LineIndicator.Mnemonic != "pfd(3,1,6)" {
		t.Errorf("mnemonic: expected 'pfd(3,1,6)', got '%s'", pfd.LineIndicator.Mnemonic)
	}
}

func TestPolynomialFitDerivativeMetadata(t *testing.T) {
	t.Parallel()

	pfd, err := NewPolynomialFitDerivative(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	meta := pfd.Metadata()

	if meta.Identifier != core.PolynomialFitDerivative {
		t.Errorf("identifier: expected PolynomialFitDerivative, got %v", meta.Identifier)
	}

	if meta.Mnemonic != "pfd(3,1,6)" {
		t.Errorf("mnemonic: expected 'pfd(3,1,6)', got '%s'", meta.Mnemonic)
	}

	if len(meta.Outputs) != 1 {
		t.Errorf("outputs: expected 1, got %d", len(meta.Outputs))
	}
}

func TestPolynomialFitDerivativeInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params *Params
	}{
		{"degree too small", &Params{Degree: 1, Order: 1, Smoothing: 6}},
		{"order too small", &Params{Degree: 3, Order: -1, Smoothing: 6}},
		{"order gt degree", &Params{Degree: 3, Order: 4, Smoothing: 6}},
		{"smoothing negative", &Params{Degree: 3, Order: 1, Smoothing: -1}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := NewPolynomialFitDerivative(tt.params); err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}
