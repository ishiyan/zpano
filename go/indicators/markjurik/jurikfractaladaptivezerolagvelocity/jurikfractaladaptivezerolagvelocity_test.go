//nolint:testpackage
package jurikfractaladaptivezerolagvelocity

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

func testFAZLV(t *testing.T, loDepth, hiDepth, fractalType, smooth int, expected []float64) {
	t.Helper()

	ind, err := NewJurikFractalAdaptiveZeroLagVelocity(&JurikFractalAdaptiveZeroLagVelocityParams{
		LoDepth:     loDepth,
		HiDepth:     hiDepth,
		FractalType: fractalType,
		Smooth:      smooth,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, input := range testInput {
		result := ind.Update(input)
		exp := expected[i]

		if !almostEqual(result, exp, epsilon) {
			t.Errorf("bar %d: got %.15f, want %.15f", i, result, exp)
			if i > 35 {
				break
			}
		}
	}
}

func TestFAZLVLo2Hi15(t *testing.T)  { t.Parallel(); testFAZLV(t, 2, 15, 1, 10, expectedLo2Hi15) }
func TestFAZLVLo2Hi30(t *testing.T)  { t.Parallel(); testFAZLV(t, 2, 30, 1, 10, expectedLo2Hi30) }
func TestFAZLVLo2Hi60(t *testing.T)  { t.Parallel(); testFAZLV(t, 2, 60, 1, 10, expectedLo2Hi60) }
func TestFAZLVLo5Hi15(t *testing.T)  { t.Parallel(); testFAZLV(t, 5, 15, 1, 10, expectedLo5Hi15) }
func TestFAZLVLo5Hi30(t *testing.T)  { t.Parallel(); testFAZLV(t, 5, 30, 1, 10, expectedLo5Hi30) }
func TestFAZLVLo5Hi60(t *testing.T)  { t.Parallel(); testFAZLV(t, 5, 60, 1, 10, expectedLo5Hi60) }
func TestFAZLVLo10Hi15(t *testing.T) { t.Parallel(); testFAZLV(t, 10, 15, 1, 10, expectedLo10Hi15) }
func TestFAZLVLo10Hi30(t *testing.T) { t.Parallel(); testFAZLV(t, 10, 30, 1, 10, expectedLo10Hi30) }
func TestFAZLVLo10Hi60(t *testing.T) { t.Parallel(); testFAZLV(t, 10, 60, 1, 10, expectedLo10Hi60) }
func TestFAZLVFtype2(t *testing.T)   { t.Parallel(); testFAZLV(t, 5, 30, 2, 10, expectedFtype2) }
func TestFAZLVFtype3(t *testing.T)   { t.Parallel(); testFAZLV(t, 5, 30, 3, 10, expectedFtype3) }
func TestFAZLVFtype4(t *testing.T)   { t.Parallel(); testFAZLV(t, 5, 30, 4, 10, expectedFtype4) }
func TestFAZLVSmooth5(t *testing.T)  { t.Parallel(); testFAZLV(t, 5, 30, 1, 5, expectedSmooth5) }
func TestFAZLVSmooth20(t *testing.T) { t.Parallel(); testFAZLV(t, 5, 30, 1, 20, expectedSmooth20) }
func TestFAZLVSmooth40(t *testing.T) { t.Parallel(); testFAZLV(t, 5, 30, 1, 40, expectedSmooth40) }
