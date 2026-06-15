//nolint:testpackage
package parabolicvertex

import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
)

const tolerance = 1e-9

func checkSeries(t *testing.T, name string, inputs, expected []float64) {
	t.Helper()

	pvtx, err := NewParabolicVertex(DefaultParams())
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", name, err)
	}

	if len(inputs) != len(expected) {
		t.Fatalf("%s: length mismatch", name)
	}

	for i := 0; i < len(inputs); i++ {
		value := pvtx.Update(inputs[i])
		exp := expected[i]

		if math.IsNaN(exp) {
			if !math.IsNaN(value) {
				t.Errorf("%s[%d]: expected NaN, got %v", name, i, value)
			}

			continue
		}

		// Combined absolute + relative tolerance. Near collinear points the vertex
		// location is ill-conditioned (denom -> 0); a relative tolerance preserves
		// 13+ significant-digit agreement.
		delta := tolerance * math.Max(1.0, math.Abs(exp))
		if math.Abs(value-exp) > delta {
			t.Errorf("%s[%d]: expected %v, got %v", name, i, exp, value)
		}
	}
}

func TestParabolicVertexData(t *testing.T) {
	t.Parallel()

	checkSeries(t, "RAW", iNPUT_CLOSE, expectedRAW)
	checkSeries(t, "EMA6", iNPUT_EMA6, expectedEMA6)
	checkSeries(t, "EMA20", iNPUT_EMA20, expectedEMA20)
	checkSeries(t, "TEST1", tEST1_INPUT_PARABOLA, tEST1_EXPECTED)
}

func TestParabolicVertexMnemonic(t *testing.T) {
	t.Parallel()

	pvtx, err := NewParabolicVertex(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pvtx.LineIndicator.Mnemonic != "pvtx" {
		t.Errorf("mnemonic: expected 'pvtx', got '%s'", pvtx.LineIndicator.Mnemonic)
	}
}

func TestParabolicVertexMnemonicWithComponent(t *testing.T) {
	t.Parallel()

	pvtx, err := NewParabolicVertex(&Params{BarComponent: entities.BarMedianPrice})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pvtx.LineIndicator.Mnemonic != "pvtx(hl/2)" {
		t.Errorf("mnemonic: expected 'pvtx(hl/2)', got '%s'", pvtx.LineIndicator.Mnemonic)
	}
}

func TestParabolicVertexMetadata(t *testing.T) {
	t.Parallel()

	pvtx, err := NewParabolicVertex(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	meta := pvtx.Metadata()

	if meta.Identifier != core.ParabolicVertex {
		t.Errorf("identifier: expected ParabolicVertex, got %v", meta.Identifier)
	}

	if meta.Mnemonic != "pvtx" {
		t.Errorf("mnemonic: expected 'pvtx', got '%s'", meta.Mnemonic)
	}

	if len(meta.Outputs) != 1 {
		t.Errorf("outputs: expected 1, got %d", len(meta.Outputs))
	}
}

func TestParabolicVertexPriming(t *testing.T) {
	t.Parallel()

	pvtx, err := NewParabolicVertex(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if v := pvtx.Update(1.0); !math.IsNaN(v) {
		t.Errorf("bar 0: expected NaN, got %v", v)
	}

	if pvtx.IsPrimed() {
		t.Errorf("bar 0: expected not primed")
	}

	if v := pvtx.Update(2.0); !math.IsNaN(v) {
		t.Errorf("bar 1: expected NaN, got %v", v)
	}

	// Three collinear points -> zero curvature -> NaN, but primed.
	if v := pvtx.Update(3.0); !math.IsNaN(v) {
		t.Errorf("bar 2: expected NaN (collinear), got %v", v)
	}

	if !pvtx.IsPrimed() {
		t.Errorf("bar 2: expected primed")
	}
}
