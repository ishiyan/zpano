//nolint:testpackage
package jurikzerolagvelocity

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

func testVEL(t *testing.T, depth int, expected []float64) {
	t.Helper()

	vel, err := NewJurikZeroLagVelocity(&JurikZeroLagVelocityParams{
		Depth: depth,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, input := range testInput {
		result := vel.Update(input)
		exp := expected[i]

		if !almostEqual(result, exp, epsilon) {
			t.Errorf("bar %d: got %.15f, want %.15f", i, result, exp)
			if i > 35 {
				break
			}
		}
	}
}

func TestVELDepth2(t *testing.T)  { t.Parallel(); testVEL(t, 2, expectedDepth2) }
func TestVELDepth3(t *testing.T)  { t.Parallel(); testVEL(t, 3, expectedDepth3) }
func TestVELDepth4(t *testing.T)  { t.Parallel(); testVEL(t, 4, expectedDepth4) }
func TestVELDepth5(t *testing.T)  { t.Parallel(); testVEL(t, 5, expectedDepth5) }
func TestVELDepth6(t *testing.T)  { t.Parallel(); testVEL(t, 6, expectedDepth6) }
func TestVELDepth7(t *testing.T)  { t.Parallel(); testVEL(t, 7, expectedDepth7) }
func TestVELDepth8(t *testing.T)  { t.Parallel(); testVEL(t, 8, expectedDepth8) }
func TestVELDepth9(t *testing.T)  { t.Parallel(); testVEL(t, 9, expectedDepth9) }
func TestVELDepth10(t *testing.T) { t.Parallel(); testVEL(t, 10, expectedDepth10) }
func TestVELDepth11(t *testing.T) { t.Parallel(); testVEL(t, 11, expectedDepth11) }
func TestVELDepth12(t *testing.T) { t.Parallel(); testVEL(t, 12, expectedDepth12) }
func TestVELDepth13(t *testing.T) { t.Parallel(); testVEL(t, 13, expectedDepth13) }
func TestVELDepth14(t *testing.T) { t.Parallel(); testVEL(t, 14, expectedDepth14) }
func TestVELDepth15(t *testing.T) { t.Parallel(); testVEL(t, 15, expectedDepth15) }
