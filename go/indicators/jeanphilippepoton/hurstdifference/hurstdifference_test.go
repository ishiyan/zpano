//nolint:testpackage
package hurstdifference

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/indicators/core"
)

func testHurstDifferenceCreate(period int) *HurstDifference {
	ind, _ := NewHurstDifference(&Params{Period: period})
	return ind
}

func TestHurstDifferenceUpdate(t *testing.T) {
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
		name     string
		period   int
		expFdi   []float64
		expHdiff []float64
	}{
		{"period = 5", 5, expectedFDIP5, expectedHDIFFP5},
		{"period = 10", 10, expectedFDIP10, expectedHDIFFP10},
		{"period = 15", 15, expectedFDIP15, expectedHDIFFP15},
		{"period = 20", 20, expectedFDIP20, expectedHDIFFP20},
		{"period = 30", 30, expectedFDIP30, expectedHDIFFP30},
		{"period = 50", 50, expectedFDIP50, expectedHDIFFP50},
		{"period = 80", 80, expectedFDIP80, expectedHDIFFP80},
		{"period = 120", 120, expectedFDIP120, expectedHDIFFP120},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ind := testHurstDifferenceCreate(tt.period)

			for i, v := range input {
				hdiff, fdi := ind.UpdateAll(v)
				check(t, "fdi", i, tt.expFdi[i], fdi)
				check(t, "hdiff", i, tt.expHdiff[i], hdiff)
			}
		})
	}
}

func TestHurstDifferenceIsPrimed(t *testing.T) {
	t.Parallel()

	ind := testHurstDifferenceCreate(30)

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

func TestHurstDifferenceNaN(t *testing.T) {
	t.Parallel()

	ind := testHurstDifferenceCreate(5)
	hdiff, fdi := ind.UpdateAll(math.NaN())

	if !math.IsNaN(hdiff) {
		t.Errorf("expected NaN for hdiff, got %v", hdiff)
	}

	if !math.IsNaN(fdi) {
		t.Errorf("expected NaN for fdi, got %v", fdi)
	}
}

func TestHurstDifferenceInvalidPeriod(t *testing.T) {
	t.Parallel()

	_, err := NewHurstDifference(&Params{Period: 1})
	if err == nil {
		t.Error("expected error for period < 2")
	}
}

func TestHurstDifferenceMetadata(t *testing.T) {
	t.Parallel()

	ind := testHurstDifferenceCreate(30)
	meta := ind.Metadata()

	if meta.Identifier != core.HurstDifference {
		t.Errorf("expected identifier %v, got %v", core.HurstDifference, meta.Identifier)
	}
}
