//nolint:testpackage
package cubicvertex

import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
)

const tolerance = 1e-9

func runSeries(inputs []float64) ([]float64, []float64) {
	cvtx, _ := NewCubicVertex(DefaultParams())

	near := make([]float64, len(inputs))
	far := make([]float64, len(inputs))

	for i := 0; i < len(inputs); i++ {
		near[i], far[i] = cvtx.Update(inputs[i])
	}

	return near, far
}

func checkSeries(t *testing.T, name string, actual, expected []float64) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Fatalf("%s: length mismatch", name)
	}

	for i := 0; i < len(expected); i++ {
		exp := expected[i]

		if math.IsNaN(exp) {
			if !math.IsNaN(actual[i]) {
				t.Errorf("%s[%d]: expected NaN, got %v", name, i, actual[i])
			}

			continue
		}

		// Combined absolute + relative tolerance (ill-conditioned near degenerate points).
		delta := tolerance * math.Max(1.0, math.Abs(exp))
		if math.Abs(actual[i]-exp) > delta {
			t.Errorf("%s[%d]: expected %v, got %v", name, i, exp, actual[i])
		}
	}
}

func TestCubicVertexData(t *testing.T) {
	t.Parallel()

	rawNear, rawFar := runSeries(iNPUT_CLOSE)
	checkSeries(t, "RAW_NEAR", rawNear, expectedRAW_NEAR)
	checkSeries(t, "RAW_FAR", rawFar, expectedRAW_FAR)

	ema6Near, ema6Far := runSeries(iNPUT_EMA6)
	checkSeries(t, "EMA6_NEAR", ema6Near, expectedEMA6_NEAR)
	checkSeries(t, "EMA6_FAR", ema6Far, expectedEMA6_FAR)

	ema20Near, ema20Far := runSeries(iNPUT_EMA20)
	checkSeries(t, "EMA20_NEAR", ema20Near, expectedEMA20_NEAR)
	checkSeries(t, "EMA20_FAR", ema20Far, expectedEMA20_FAR)

	test1Near, test1Far := runSeries(tEST1_INPUT_CUBIC)
	checkSeries(t, "TEST1_NEAR", test1Near, tEST1_EXPECTED_NEAR)
	checkSeries(t, "TEST1_FAR", test1Far, tEST1_EXPECTED_FAR)
}

func TestCubicVertexMnemonic(t *testing.T) {
	t.Parallel()

	cvtx, err := NewCubicVertex(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cvtx.mnemonic != "cvtx" {
		t.Errorf("mnemonic: expected 'cvtx', got '%s'", cvtx.mnemonic)
	}
}

func TestCubicVertexMnemonicWithComponent(t *testing.T) {
	t.Parallel()

	cvtx, err := NewCubicVertex(&Params{BarComponent: entities.BarMedianPrice})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cvtx.mnemonic != "cvtx(hl/2)" {
		t.Errorf("mnemonic: expected 'cvtx(hl/2)', got '%s'", cvtx.mnemonic)
	}
}

func TestCubicVertexMetadata(t *testing.T) {
	t.Parallel()

	cvtx, err := NewCubicVertex(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	meta := cvtx.Metadata()

	if meta.Identifier != core.CubicVertex {
		t.Errorf("identifier: expected CubicVertex, got %v", meta.Identifier)
	}

	if meta.Mnemonic != "cvtx" {
		t.Errorf("mnemonic: expected 'cvtx', got '%s'", meta.Mnemonic)
	}

	if len(meta.Outputs) != 2 {
		t.Errorf("outputs: expected 2, got %d", len(meta.Outputs))
	}
}

func TestCubicVertexPriming(t *testing.T) {
	t.Parallel()

	cvtx, err := NewCubicVertex(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < 3; i++ {
		near, far := cvtx.Update(1.0)
		if !math.IsNaN(near) || !math.IsNaN(far) {
			t.Errorf("bar %d: expected (NaN, NaN), got (%v, %v)", i, near, far)
		}

		if cvtx.IsPrimed() {
			t.Errorf("bar %d: expected not primed", i)
		}
	}

	// Four collinear points -> c == 0 and d == 0 -> both NaN, but primed.
	near, far := cvtx.Update(1.0)
	if !math.IsNaN(near) || !math.IsNaN(far) {
		t.Errorf("bar 3: expected (NaN, NaN) for collinear, got (%v, %v)", near, far)
	}

	if !cvtx.IsPrimed() {
		t.Errorf("bar 3: expected primed")
	}
}

func TestCubicVertexUpdateScalar(t *testing.T) {
	t.Parallel()

	cvtx, err := NewCubicVertex(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out core.Output
	for _, c := range iNPUT_CLOSE {
		out = cvtx.UpdateScalar(&entities.Scalar{Value: c})
	}

	if len(out) != 2 {
		t.Errorf("outputs: expected 2, got %d", len(out))
	}
}
