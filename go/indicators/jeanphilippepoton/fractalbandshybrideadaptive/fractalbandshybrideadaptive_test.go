//nolint:testpackage
package fractalbandshybrideadaptive

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/indicators/core"
)

func testCreate(period, normalSpeedFallback int, alpha, nyquist, alphaHP float64) *FractalBandsHybrideAdaptive {
	ind, _ := NewFractalBandsHybrideAdaptive(&Params{
		Period:              period,
		NormalSpeedFallback: normalSpeedFallback,
		Alpha:              alpha,
		Nyquist:            nyquist,
		AlphaHP:            alphaHP,
	})

	return ind
}

func TestFractalBandsHybrideAdaptiveUpdate(t *testing.T) {
	t.Parallel()

	const epsilon = 2e-13

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
		name                string
		period              int
		normalSpeedFallback int
		alpha               float64
		nyquist             float64
		alphaHP             float64
		expFrasma           []float64
		expUpper            []float64
		expLower            []float64
	}{
		{"P10_NY05_AHP007", 10, 30, 2.0, 0.5, 0.07, expectedFrasmaP10NY05AHP007, expectedUpperP10NY05AHP007, expectedLowerP10NY05AHP007},
		{"P10_NY05_AHP015", 10, 30, 2.0, 0.5, 0.15, expectedFrasmaP10NY05AHP015, expectedUpperP10NY05AHP015, expectedLowerP10NY05AHP015},
		{"P10_NY10_AHP007", 10, 30, 2.0, 1.0, 0.07, expectedFrasmaP10NY10AHP007, expectedUpperP10NY10AHP007, expectedLowerP10NY10AHP007},
		{"P10_NY10_AHP015", 10, 30, 2.0, 1.0, 0.15, expectedFrasmaP10NY10AHP015, expectedUpperP10NY10AHP015, expectedLowerP10NY10AHP015},
		{"P20_NY05_AHP007", 20, 30, 2.0, 0.5, 0.07, expectedFrasmaP20NY05AHP007, expectedUpperP20NY05AHP007, expectedLowerP20NY05AHP007},
		{"P20_NY05_AHP015", 20, 30, 2.0, 0.5, 0.15, expectedFrasmaP20NY05AHP015, expectedUpperP20NY05AHP015, expectedLowerP20NY05AHP015},
		{"P20_NY10_AHP007", 20, 30, 2.0, 1.0, 0.07, expectedFrasmaP20NY10AHP007, expectedUpperP20NY10AHP007, expectedLowerP20NY10AHP007},
		{"P20_NY10_AHP015", 20, 30, 2.0, 1.0, 0.15, expectedFrasmaP20NY10AHP015, expectedUpperP20NY10AHP015, expectedLowerP20NY10AHP015},
		{"P30_NY05_AHP007", 30, 30, 2.0, 0.5, 0.07, expectedFrasmaP30NY05AHP007, expectedUpperP30NY05AHP007, expectedLowerP30NY05AHP007},
		{"P30_NY05_AHP015", 30, 30, 2.0, 0.5, 0.15, expectedFrasmaP30NY05AHP015, expectedUpperP30NY05AHP015, expectedLowerP30NY05AHP015},
		{"P30_NY10_AHP007", 30, 30, 2.0, 1.0, 0.07, expectedFrasmaP30NY10AHP007, expectedUpperP30NY10AHP007, expectedLowerP30NY10AHP007},
		{"P30_NY10_AHP015", 30, 30, 2.0, 1.0, 0.15, expectedFrasmaP30NY10AHP015, expectedUpperP30NY10AHP015, expectedLowerP30NY10AHP015},
		{"P50_NY05_AHP007", 50, 30, 2.0, 0.5, 0.07, expectedFrasmaP50NY05AHP007, expectedUpperP50NY05AHP007, expectedLowerP50NY05AHP007},
		{"P50_NY05_AHP015", 50, 30, 2.0, 0.5, 0.15, expectedFrasmaP50NY05AHP015, expectedUpperP50NY05AHP015, expectedLowerP50NY05AHP015},
		{"P50_NY10_AHP007", 50, 30, 2.0, 1.0, 0.07, expectedFrasmaP50NY10AHP007, expectedUpperP50NY10AHP007, expectedLowerP50NY10AHP007},
		{"P50_NY10_AHP015", 50, 30, 2.0, 1.0, 0.15, expectedFrasmaP50NY10AHP015, expectedUpperP50NY10AHP015, expectedLowerP50NY10AHP015},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ind := testCreate(tt.period, tt.normalSpeedFallback, tt.alpha, tt.nyquist, tt.alphaHP)

			for i, v := range input {
				frasma2, upper, lower := ind.UpdateAll(v)
				check(t, "frasma2", i, tt.expFrasma[i], frasma2)
				check(t, "upper", i, tt.expUpper[i], upper)
				check(t, "lower", i, tt.expLower[i], lower)
			}
		})
	}
}

func TestFractalBandsHybrideAdaptiveIsPrimed(t *testing.T) {
	t.Parallel()

	ind := testCreate(30, 30, 2.0, 0.5, 0.07)

	for i := 0; i < 30; i++ {
		ind.Update(testInput[i])

		if ind.IsPrimed() {
			t.Errorf("expected not primed at index %d", i)
		}
	}

	ind.Update(testInput[30])

	if !ind.IsPrimed() {
		t.Error("expected primed after 31 updates")
	}
}

func TestFractalBandsHybrideAdaptiveNaNPassthrough(t *testing.T) {
	t.Parallel()

	ind := testCreate(5, 30, 2.0, 0.5, 0.07)
	frasma2, upper, lower := ind.UpdateAll(math.NaN())

	if !math.IsNaN(frasma2) || !math.IsNaN(upper) || !math.IsNaN(lower) {
		t.Error("expected NaN passthrough")
	}
}

func TestFractalBandsHybrideAdaptiveInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    Params
	}{
		{"period<2", Params{Period: 1, NormalSpeedFallback: 30, Alpha: 2.0, Nyquist: 0.5, AlphaHP: 0.07}},
		{"nsf<1", Params{Period: 30, NormalSpeedFallback: 0, Alpha: 2.0, Nyquist: 0.5, AlphaHP: 0.07}},
		{"alpha<=0", Params{Period: 30, NormalSpeedFallback: 30, Alpha: 0.0, Nyquist: 0.5, AlphaHP: 0.07}},
		{"nyquist<=0", Params{Period: 30, NormalSpeedFallback: 30, Alpha: 2.0, Nyquist: 0.0, AlphaHP: 0.07}},
		{"alphaHP<=0", Params{Period: 30, NormalSpeedFallback: 30, Alpha: 2.0, Nyquist: 0.5, AlphaHP: 0.0}},
		{"alphaHP>=1", Params{Period: 30, NormalSpeedFallback: 30, Alpha: 2.0, Nyquist: 0.5, AlphaHP: 1.0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewFractalBandsHybrideAdaptive(&tt.p)
			if err == nil {
				t.Errorf("expected error for %s", tt.name)
			}
		})
	}
}

func TestFractalBandsHybrideAdaptiveMetadata(t *testing.T) {
	t.Parallel()

	ind := testCreate(30, 30, 2.0, 0.5, 0.07)
	meta := ind.Metadata()

	if meta.Identifier != core.FractalBandsHybrideAdaptive {
		t.Errorf("expected identifier %v, got %v", core.FractalBandsHybrideAdaptive, meta.Identifier)
	}
}
