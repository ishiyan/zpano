//nolint:testpackage
package fractionalbands

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/indicators/core"
)

func testCreate(period int, priceScale float64) *FractionalBands {
	ind, _ := NewFractionalBands(&Params{Period: period, PriceScale: priceScale})
	return ind
}

func TestFractionalBandsUpdate(t *testing.T) {
	t.Parallel()

	const epsilon = 1e-11

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
		name       string
		period     int
		priceScale float64
		expFrasma2  []float64
		expUpper   []float64
		expLower   []float64
	}{
		{"P5_S1", 5, 1.0, expectedFrasma2P5S1, expectedUpperP5S1, expectedLowerP5S1},
		{"P10_S1", 10, 1.0, expectedFrasma2P10S1, expectedUpperP10S1, expectedLowerP10S1},
		{"P20_S1", 20, 1.0, expectedFrasma2P20S1, expectedUpperP20S1, expectedLowerP20S1},
		{"P30_S1", 30, 1.0, expectedFrasma2P30S1, expectedUpperP30S1, expectedLowerP30S1},
		{"P50_S1", 50, 1.0, expectedFrasma2P50S1, expectedUpperP50S1, expectedLowerP50S1},
		{"P80_S1", 80, 1.0, expectedFrasma2P80S1, expectedUpperP80S1, expectedLowerP80S1},
		{"P30_S100", 30, 100.0, expectedFrasma2P30S100, expectedUpperP30S100, expectedLowerP30S100},
		{"P30_S10000", 30, 10000.0, expectedFrasma2P30S10000, expectedUpperP30S10000, expectedLowerP30S10000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ind := testCreate(tt.period, tt.priceScale)

			for i, v := range input {
				frasma2, upper, lower := ind.UpdateAll(v)
				check(t, "frasma2", i, tt.expFrasma2[i], frasma2)
				check(t, "upper", i, tt.expUpper[i], upper)
				check(t, "lower", i, tt.expLower[i], lower)
			}
		})
	}
}

func TestFractionalBandsIsPrimed(t *testing.T) {
	t.Parallel()

	ind := testCreate(30, 1.0)

	for i := 0; i < 30; i++ {
		ind.Update(testInput[i])

		if ind.IsPrimed() {
			t.Errorf("expected not primed at index %d", i)
		}
	}

	ind.Update(testInput[30])

	if !ind.IsPrimed() {
		t.Error("expected primed after 31 samples")
	}
}

func TestFractionalBandsNaN(t *testing.T) {
	t.Parallel()

	ind := testCreate(5, 1.0)
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

func TestFractionalBandsInvalidParams(t *testing.T) {
	t.Parallel()

	_, err := NewFractionalBands(&Params{Period: 1, PriceScale: 1.0})
	if err == nil {
		t.Error("expected error for period < 2")
	}

	_, err = NewFractionalBands(&Params{Period: 30, PriceScale: 0.0})
	if err == nil {
		t.Error("expected error for price_scale <= 0")
	}

	_, err = NewFractionalBands(&Params{Period: 30, PriceScale: -1.0})
	if err == nil {
		t.Error("expected error for price_scale < 0")
	}
}

func TestFractionalBandsMetadata(t *testing.T) {
	t.Parallel()

	ind := testCreate(30, 1.0)
	meta := ind.Metadata()

	if meta.Identifier != core.FractionalBands {
		t.Errorf("expected identifier %v, got %v", core.FractionalBands, meta.Identifier)
	}
}
