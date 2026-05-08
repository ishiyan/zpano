//nolint:testpackage
package jurikadaptiverelativetrendstrengthindex

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

func testARTSI(t *testing.T, loLength, hiLength int, expected []float64) {
	t.Helper()

	ind, err := NewJurikAdaptiveRelativeTrendStrengthIndex(&JurikAdaptiveRelativeTrendStrengthIndexParams{
		LoLength: loLength,
		HiLength: hiLength,
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

func TestARTSILo2Hi15(t *testing.T)  { t.Parallel(); testARTSI(t, 2, 15, expectedLo2Hi15) }
func TestARTSILo2Hi30(t *testing.T)  { t.Parallel(); testARTSI(t, 2, 30, expectedLo2Hi30) }
func TestARTSILo2Hi60(t *testing.T)  { t.Parallel(); testARTSI(t, 2, 60, expectedLo2Hi60) }
func TestARTSILo5Hi15(t *testing.T)  { t.Parallel(); testARTSI(t, 5, 15, expectedLo5Hi15) }
func TestARTSILo5Hi30(t *testing.T)  { t.Parallel(); testARTSI(t, 5, 30, expectedLo5Hi30) }
func TestARTSILo5Hi60(t *testing.T)  { t.Parallel(); testARTSI(t, 5, 60, expectedLo5Hi60) }
func TestARTSILo10Hi15(t *testing.T) { t.Parallel(); testARTSI(t, 10, 15, expectedLo10Hi15) }
func TestARTSILo10Hi30(t *testing.T) { t.Parallel(); testARTSI(t, 10, 30, expectedLo10Hi30) }
func TestARTSILo10Hi60(t *testing.T) { t.Parallel(); testARTSI(t, 10, 60, expectedLo10Hi60) }
