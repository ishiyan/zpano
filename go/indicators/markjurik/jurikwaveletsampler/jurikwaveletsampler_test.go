//nolint:testpackage
package jurikwaveletsampler

import (
	"math"
	"testing"
)

const epsilon = 1e-13

func almostEqual(a, b, eps float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}

	return math.Abs(a-b) < eps
}

func TestWAVDefault(t *testing.T) {
	t.Parallel()

	ind, err := NewJurikWaveletSampler(&JurikWaveletSamplerParams{
		Index: 12,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedCols := [][]float64{
		expectedWAVCol0, expectedWAVCol1, expectedWAVCol2, expectedWAVCol3,
		expectedWAVCol4, expectedWAVCol5, expectedWAVCol6, expectedWAVCol7,
		expectedWAVCol8, expectedWAVCol9, expectedWAVCol10, expectedWAVCol11,
	}

	for i, input := range testInput {
		result := ind.Update(input)
		exp := expectedWAVCol0[i]

		if !almostEqual(result, exp, epsilon) {
			t.Errorf("bar %d col0: got %.15f, want %.15f", i, result, exp)
			if i > 35 {
				break
			}
		}

		cols := ind.Columns()
		for c := 0; c < 12; c++ {
			if !almostEqual(cols[c], expectedCols[c][i], epsilon) {
				t.Errorf("bar %d col %d: got %.15f, want %.15f", i, c, cols[c], expectedCols[c][i])
				if i > 35 {
					break
				}
			}
		}
	}
}

func TestWAVIndex6(t *testing.T) {
	t.Parallel()

	ind, err := NewJurikWaveletSampler(&JurikWaveletSamplerParams{Index: 6})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedCols := [][]float64{
		expectedIndex6Col0, expectedIndex6Col1, expectedIndex6Col2,
		expectedIndex6Col3, expectedIndex6Col4, expectedIndex6Col5,
	}

	for i, input := range testInput {
		ind.Update(input)

		cols := ind.Columns()
		for c := 0; c < 6; c++ {
			if !almostEqual(cols[c], expectedCols[c][i], epsilon) {
				t.Errorf("bar %d col %d: got %.15f, want %.15f", i, c, cols[c], expectedCols[c][i])
				if i > 35 {
					return
				}
			}
		}
	}
}

func TestWAVIndex16(t *testing.T) {
	t.Parallel()

	ind, err := NewJurikWaveletSampler(&JurikWaveletSamplerParams{Index: 16})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedCols := [][]float64{
		expectedIndex16Col0, expectedIndex16Col1, expectedIndex16Col2, expectedIndex16Col3,
		expectedIndex16Col4, expectedIndex16Col5, expectedIndex16Col6, expectedIndex16Col7,
		expectedIndex16Col8, expectedIndex16Col9, expectedIndex16Col10, expectedIndex16Col11,
		expectedIndex16Col12, expectedIndex16Col13, expectedIndex16Col14, expectedIndex16Col15,
	}

	for i, input := range testInput {
		ind.Update(input)

		cols := ind.Columns()
		for c := 0; c < 16; c++ {
			if !almostEqual(cols[c], expectedCols[c][i], epsilon) {
				t.Errorf("bar %d col %d: got %.15f, want %.15f", i, c, cols[c], expectedCols[c][i])
				if i > 35 {
					return
				}
			}
		}
	}
}
