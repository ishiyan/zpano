//nolint:testpackage
package jurikturningpointoscillator

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

func testTPO(t *testing.T, length int, expected []float64) {
	t.Helper()

	ind, err := NewJurikTurningPointOscillator(&JurikTurningPointOscillatorParams{
		Length: length,
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

func TestTPOLen5(t *testing.T)  { t.Parallel(); testTPO(t, 5, expectedLen5) }
func TestTPOLen7(t *testing.T)  { t.Parallel(); testTPO(t, 7, expectedLen7) }
func TestTPOLen10(t *testing.T) { t.Parallel(); testTPO(t, 10, expectedLen10) }
func TestTPOLen14(t *testing.T) { t.Parallel(); testTPO(t, 14, expectedTPO) }
func TestTPOLen20(t *testing.T) { t.Parallel(); testTPO(t, 20, expectedLen20) }
func TestTPOLen28(t *testing.T) { t.Parallel(); testTPO(t, 28, expectedLen28) }
func TestTPOLen40(t *testing.T) { t.Parallel(); testTPO(t, 40, expectedLen40) }
func TestTPOLen60(t *testing.T) { t.Parallel(); testTPO(t, 60, expectedLen60) }
func TestTPOLen80(t *testing.T) { t.Parallel(); testTPO(t, 80, expectedLen80) }
