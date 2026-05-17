//nolint:testpackage
package fractalbands

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/indicators/core"
)

func testCreate(period, normalSpeed int, alpha float64) *FractalBands {
	ind, _ := NewFractalBands(&Params{Period: period, NormalSpeed: normalSpeed, Alpha: alpha})
	return ind
}

func TestFractalBandsUpdate(t *testing.T) {
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
		name        string
		period      int
		normalSpeed int
		alpha       float64
		expFrasma2  []float64
		expUpper    []float64
		expLower    []float64
	}{
		{"P10_NS20_A2", 10, 20, 2.0, expectedFrasma2P10Ns20A2, expectedUpperP10Ns20A2, expectedLowerP10Ns20A2},
		{"P20_NS20_A2", 20, 20, 2.0, expectedFrasma2P20Ns20A2, expectedUpperP20Ns20A2, expectedLowerP20Ns20A2},
		{"P30_NS20_A2", 30, 20, 2.0, expectedFrasma2P30Ns20A2, expectedUpperP30Ns20A2, expectedLowerP30Ns20A2},
		{"P50_NS20_A2", 50, 20, 2.0, expectedFrasma2P50Ns20A2, expectedUpperP50Ns20A2, expectedLowerP50Ns20A2},
		{"P30_NS10_A2", 30, 10, 2.0, expectedFrasma2P30Ns10A2, expectedUpperP30Ns10A2, expectedLowerP30Ns10A2},
		{"P30_NS40_A2", 30, 40, 2.0, expectedFrasma2P30Ns40A2, expectedUpperP30Ns40A2, expectedLowerP30Ns40A2},
		{"P30_NS20_A1", 30, 20, 1.0, expectedFrasma2P30Ns20A1, expectedUpperP30Ns20A1, expectedLowerP30Ns20A1},
		{"P30_NS20_A3", 30, 20, 3.0, expectedFrasma2P30Ns20A3, expectedUpperP30Ns20A3, expectedLowerP30Ns20A3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ind := testCreate(tt.period, tt.normalSpeed, tt.alpha)

			for i, v := range input {
				frasma2, upper, lower := ind.UpdateAll(v)
				check(t, "frasma2", i, tt.expFrasma2[i], frasma2)
				check(t, "upper", i, tt.expUpper[i], upper)
				check(t, "lower", i, tt.expLower[i], lower)
			}
		})
	}
}

func TestFractalBandsIsPrimed(t *testing.T) {
	t.Parallel()

	ind := testCreate(30, 20, 2.0)

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

func TestFractalBandsNaN(t *testing.T) {
	t.Parallel()

	ind := testCreate(5, 20, 2.0)
	frasma2, upper, lower := ind.UpdateAll(math.NaN())

	if !math.IsNaN(frasma2) {
		t.Errorf("expected NaN for frasma2, got %v", frasma2)
	}

	if !math.IsNaN(upper) {
		t.Errorf("expected NaN for upper, got %v", upper)
	}

	if !math.IsNaN(lower) {
		t.Errorf("expected NaN for lower, got %v", lower)
	}
}

func TestFractalBandsInvalidParams(t *testing.T) {
	t.Parallel()

	_, err := NewFractalBands(&Params{Period: 1, NormalSpeed: 20, Alpha: 2.0})
	if err == nil {
		t.Error("expected error for period < 2")
	}

	_, err = NewFractalBands(&Params{Period: 30, NormalSpeed: 0, Alpha: 2.0})
	if err == nil {
		t.Error("expected error for normal_speed < 1")
	}

	_, err = NewFractalBands(&Params{Period: 30, NormalSpeed: 20, Alpha: 0.0})
	if err == nil {
		t.Error("expected error for alpha <= 0")
	}
}

func TestFractalBandsMetadata(t *testing.T) {
	t.Parallel()

	ind := testCreate(30, 20, 2.0)
	meta := ind.Metadata()

	if meta.Identifier != core.FractalBands {
		t.Errorf("expected identifier %v, got %v", core.FractalBands, meta.Identifier)
	}
}
