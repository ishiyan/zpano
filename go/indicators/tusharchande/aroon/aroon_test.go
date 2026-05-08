//nolint:testpackage
package aroon

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func TestAroon_Length14_FullData(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-6

	high := testInputHigh()
	low := testInputLow()
	expected := testExpected()

	ind, err := NewAroon(&AroonParams{Length: 14})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 252; i++ {
		up, down, osc := ind.Update(high[i], low[i])

		if math.IsNaN(expected[i].up) {
			if !math.IsNaN(up) {
				t.Errorf("[%d] Up: expected NaN, got %v", i, up)
			}

			continue
		}

		if math.Abs(up-expected[i].up) > tolerance {
			t.Errorf("[%d] Up: expected %v, got %v", i, expected[i].up, up)
		}

		if math.Abs(down-expected[i].down) > tolerance {
			t.Errorf("[%d] Down: expected %v, got %v", i, expected[i].down, down)
		}

		if math.Abs(osc-expected[i].osc) > tolerance {
			t.Errorf("[%d] Osc: expected %v, got %v", i, expected[i].osc, osc)
		}
	}
}

func TestAroonIsPrimed(t *testing.T) {
	t.Parallel()

	ind, err := NewAroon(&AroonParams{Length: 14})
	if err != nil {
		t.Fatal(err)
	}

	high := testInputHigh()
	low := testInputLow()

	if ind.IsPrimed() {
		t.Error("expected not primed initially")
	}

	// Feed first 14 bars (indices 0..13), should not be primed.
	for i := 0; i < 14; i++ {
		ind.Update(high[i], low[i])
		if ind.IsPrimed() {
			t.Errorf("[%d] expected not primed", i)
		}
	}

	// Index 14: first primed value.
	ind.Update(high[14], low[14])

	if !ind.IsPrimed() {
		t.Error("expected primed after index 14")
	}
}

func TestAroonNaN(t *testing.T) {
	t.Parallel()

	ind, err := NewAroon(&AroonParams{Length: 14})
	if err != nil {
		t.Fatal(err)
	}

	up, down, osc := ind.Update(math.NaN(), 1.0)
	if !math.IsNaN(up) {
		t.Errorf("expected NaN Up, got %v", up)
	}

	if !math.IsNaN(down) {
		t.Errorf("expected NaN Down, got %v", down)
	}

	if !math.IsNaN(osc) {
		t.Errorf("expected NaN Osc, got %v", osc)
	}
}

func TestAroonMetadata(t *testing.T) {
	t.Parallel()

	ind, err := NewAroon(&AroonParams{Length: 14})
	if err != nil {
		t.Fatal(err)
	}

	meta := ind.Metadata()

	if meta.Identifier != core.Aroon {
		t.Errorf("expected identifier Aroon, got %v", meta.Identifier)
	}

	exp := "aroon(14)"
	if meta.Mnemonic != exp {
		t.Errorf("expected mnemonic '%s', got '%s'", exp, meta.Mnemonic)
	}

	if len(meta.Outputs) != 3 {
		t.Fatalf("expected 3 outputs, got %d", len(meta.Outputs))
	}

	if meta.Outputs[0].Kind != int(Up) {
		t.Errorf("expected output 0 kind %d, got %d", Up, meta.Outputs[0].Kind)
	}

	if meta.Outputs[0].Shape != shape.Scalar {
		t.Errorf("expected scalar output type, got %v", meta.Outputs[0].Shape)
	}

	if meta.Outputs[1].Kind != int(Down) {
		t.Errorf("expected output 1 kind %d, got %d", Down, meta.Outputs[1].Kind)
	}

	if meta.Outputs[2].Kind != int(Osc) {
		t.Errorf("expected output 2 kind %d, got %d", Osc, meta.Outputs[2].Kind)
	}
}

func TestAroonUpdateBar(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-6

	high := testInputHigh()
	low := testInputLow()
	expected := testExpected()

	ind, err := NewAroon(&AroonParams{Length: 14})
	if err != nil {
		t.Fatal(err)
	}

	tm := testTime()

	for i := 0; i < 14; i++ {
		bar := &entities.Bar{Time: tm, High: high[i], Low: low[i], Close: 0}
		out := ind.UpdateBar(bar)

		v := out[0].(entities.Scalar).Value //nolint:forcetypeassert
		if !math.IsNaN(v) {
			t.Errorf("[%d] expected NaN Up, got %v", i, v)
		}
	}

	bar := &entities.Bar{Time: tm, High: high[14], Low: low[14], Close: 0}
	out := ind.UpdateBar(bar)

	up := out[0].(entities.Scalar).Value   //nolint:forcetypeassert
	down := out[1].(entities.Scalar).Value //nolint:forcetypeassert
	osc := out[2].(entities.Scalar).Value  //nolint:forcetypeassert

	if math.Abs(up-expected[14].up) > tolerance {
		t.Errorf("[14] Up: expected %v, got %v", expected[14].up, up)
	}

	if math.Abs(down-expected[14].down) > tolerance {
		t.Errorf("[14] Down: expected %v, got %v", expected[14].down, down)
	}

	if math.Abs(osc-expected[14].osc) > tolerance {
		t.Errorf("[14] Osc: expected %v, got %v", expected[14].osc, osc)
	}
}

func TestAroonInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		length int
	}{
		{"length too small", 1},
		{"length zero", 0},
		{"length negative", -1},
	}

	for _, tt := range tests {
		_, err := NewAroon(&AroonParams{Length: tt.length})
		if err == nil {
			t.Errorf("%s: expected error, got nil", tt.name)
		}
	}
}
