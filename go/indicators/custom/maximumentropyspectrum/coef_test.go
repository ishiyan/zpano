//nolint:testpackage
package maximumentropyspectrum

import (
	"math"
	"testing"
)

// Expected AR coefficients from MBST's MaximumEntropySpectrumEstimatorTest.cs
// (originally from http://paulbourke.net/miscellaneous/ar/).
// All are compared at MBST's tolerance: rounded to `dec` decimals.
//
//nolint:gochecknoglobals
var coefCases = []struct {
	name   string
	input  []float64
	degree int
	dec    int
	want   []float64
}{
	{"sinusoids/1", testInputFourSinusoids(), 1, 1, []float64{0.941872}},
	{"sinusoids/2", testInputFourSinusoids(), 2, 1, []float64{1.826156, -0.938849}},
	{"sinusoids/3", testInputFourSinusoids(), 3, 1, []float64{2.753231, -2.740306, 0.985501}},
	{"sinusoids/4", testInputFourSinusoids(), 4, 1, []float64{3.736794, -5.474295, 3.731127, -0.996783}},
	{"test1/5", testInputTest1(), 5, 1, []float64{1.4, -0.7, 0.04, 0.7, -0.5}},
	{"test2/7", testInputTest2(), 7, 0, []float64{0.677, 0.175, 0.297, 0.006, -0.114, -0.083, -0.025}},
	{"test3/2", testInputTest3(), 2, 1, []float64{1.02, -0.53}},
}

func roundDec(v float64, dec int) float64 {
	p := math.Pow(10, float64(dec))

	return math.Round(v*p) / p
}

func TestBurgCoefficientsAgainstMbst(t *testing.T) {
	t.Parallel()

	for _, tc := range coefCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			length := len(tc.input)
			e := newEstimator(length, tc.degree, 2, float64(length*2), 1, false, 0.995)
			copy(e.inputSeries, tc.input)
			e.calculate()

			if len(e.coefficients) != tc.degree {
				t.Fatalf("coefficients len: expected %d, got %d", tc.degree, len(e.coefficients))
			}

			for i, w := range tc.want {
				got := roundDec(e.coefficients[i], tc.dec)
				exp := roundDec(w, tc.dec)

				if got != exp {
					t.Errorf("coef[%d]: expected %v, got %v (raw %.6f)",
						i, exp, got, e.coefficients[i])
				}
			}
		})
	}
}
