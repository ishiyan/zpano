//nolint:testpackage
package fractalgeneralizeddimensionindex

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/indicators/core"
)

func testFGDICreate(period int) *FractalGraphDimensionIndex {
	ind, _ := NewFractalGraphDimensionIndex(&Params{Period: period})
	return ind
}

func TestFractalGraphDimensionIndexUpdate(t *testing.T) {
	t.Parallel()

	const epsilon = 1e-13

	check := func(t *testing.T, label string, index int, exp, act float64) {
		t.Helper()

		if math.IsNaN(exp) {
			if !math.IsNaN(act) {
				t.Errorf("[%v] %s: expected NaN, got %v", index, label, act)
			}

			return
		}

		if math.Abs(exp-act) > epsilon {
			t.Errorf("[%v] %s: expected %v, got %v", index, label, exp, act)
		}
	}

	input := testInput

	tests := []struct {
		name      string
		period    int
		expFdi    []float64
		expUpper  []float64
		expLower  []float64
		expStddev []float64
	}{
		{"period = 5", 5, expectedFdiP5, expectedUpperP5, expectedLowerP5, expectedStddevP5},
		{"period = 10", 10, expectedFdiP10, expectedUpperP10, expectedLowerP10, expectedStddevP10},
		{"period = 15", 15, expectedFdiP15, expectedUpperP15, expectedLowerP15, expectedStddevP15},
		{"period = 20", 20, expectedFdiP20, expectedUpperP20, expectedLowerP20, expectedStddevP20},
		{"period = 30", 30, expectedFdiP30, expectedUpperP30, expectedLowerP30, expectedStddevP30},
		{"period = 50", 50, expectedFdiP50, expectedUpperP50, expectedLowerP50, expectedStddevP50},
		{"period = 80", 80, expectedFdiP80, expectedUpperP80, expectedLowerP80, expectedStddevP80},
		{"period = 120", 120, expectedFdiP120, expectedUpperP120, expectedLowerP120, expectedStddevP120},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ind := testFGDICreate(tt.period)

			for i, v := range input {
				fgdi, upper, lower, stddev := ind.UpdateAll(v)
				check(t, "fgdi", i, tt.expFdi[i], fgdi)
				check(t, "upper", i, tt.expUpper[i], upper)
				check(t, "lower", i, tt.expLower[i], lower)
				check(t, "stddev", i, tt.expStddev[i], stddev)
			}
		})
	}
}

func TestFractalGraphDimensionIndexIsPrimed(t *testing.T) {
	t.Parallel()

	ind := testFGDICreate(30)

	for i := 0; i < 29; i++ {
		ind.Update(testInput[i])

		if ind.IsPrimed() {
			t.Errorf("expected not primed at index %d", i)
		}
	}

	ind.Update(testInput[29])

	if !ind.IsPrimed() {
		t.Error("expected primed after 30 samples")
	}
}

func TestFractalGraphDimensionIndexNaN(t *testing.T) {
	t.Parallel()

	ind := testFGDICreate(5)
	fgdi, upper, lower, stddev := ind.UpdateAll(math.NaN())

	if !math.IsNaN(fgdi) {
		t.Errorf("expected NaN for fgdi, got %v", fgdi)
	}

	if !math.IsNaN(upper) {
		t.Errorf("expected NaN for upper, got %v", upper)
	}

	if !math.IsNaN(lower) {
		t.Errorf("expected NaN for lower, got %v", lower)
	}

	if !math.IsNaN(stddev) {
		t.Errorf("expected NaN for stddev, got %v", stddev)
	}
}

func TestFractalGraphDimensionIndexInvalidPeriod(t *testing.T) {
	t.Parallel()

	_, err := NewFractalGraphDimensionIndex(&Params{Period: 1})
	if err == nil {
		t.Error("expected error for period < 2")
	}
}

func TestFractalGraphDimensionIndexMetadata(t *testing.T) {
	t.Parallel()

	ind := testFGDICreate(30)
	meta := ind.Metadata()

	if meta.Identifier != core.FractalGraphDimensionIndex {
		t.Errorf("expected identifier %v, got %v", core.FractalGraphDimensionIndex, meta.Identifier)
	}
}
