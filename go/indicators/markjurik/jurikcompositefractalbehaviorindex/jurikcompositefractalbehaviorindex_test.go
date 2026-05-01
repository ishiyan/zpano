//nolint:testpackage
package jurikcompositefractalbehaviorindex

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

func testCFB(t *testing.T, fractalType, smooth int, expected []float64) {
	t.Helper()

	cfb, err := NewJurikCompositeFractalBehaviorIndex(&JurikCompositeFractalBehaviorIndexParams{
		FractalType: fractalType,
		Smooth:      smooth,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, input := range testInput {
		result := cfb.Update(input)
		exp := expected[i]

		// Skip last bar: reference aux loop stops at len-2, so last bar's aux values
		// are 0 in reference but computed in streaming. This causes expected divergence.
		if i == len(testInput)-1 {
			continue
		}

		if !almostEqual(result, exp, epsilon) {
			t.Errorf("bar %d: got %.15f, want %.15f", i, result, exp)
		}
	}
}

func TestCFBType1Smooth2(t *testing.T)  { t.Parallel(); testCFB(t, 1, 2, expectedType1Smooth2) }
func TestCFBType1Smooth10(t *testing.T) { t.Parallel(); testCFB(t, 1, 10, expectedType1Smooth10) }
func TestCFBType1Smooth50(t *testing.T) { t.Parallel(); testCFB(t, 1, 50, expectedType1Smooth50) }
func TestCFBType2Smooth2(t *testing.T)  { t.Parallel(); testCFB(t, 2, 2, expectedType2Smooth2) }
func TestCFBType2Smooth10(t *testing.T) { t.Parallel(); testCFB(t, 2, 10, expectedType2Smooth10) }
func TestCFBType2Smooth50(t *testing.T) { t.Parallel(); testCFB(t, 2, 50, expectedType2Smooth50) }
func TestCFBType3Smooth2(t *testing.T)  { t.Parallel(); testCFB(t, 3, 2, expectedType3Smooth2) }
func TestCFBType3Smooth10(t *testing.T) { t.Parallel(); testCFB(t, 3, 10, expectedType3Smooth10) }
func TestCFBType3Smooth50(t *testing.T) { t.Parallel(); testCFB(t, 3, 50, expectedType3Smooth50) }
func TestCFBType4Smooth2(t *testing.T)  { t.Parallel(); testCFB(t, 4, 2, expectedType4Smooth2) }
func TestCFBType4Smooth10(t *testing.T) { t.Parallel(); testCFB(t, 4, 10, expectedType4Smooth10) }
func TestCFBType4Smooth50(t *testing.T) { t.Parallel(); testCFB(t, 4, 50, expectedType4Smooth50) }
