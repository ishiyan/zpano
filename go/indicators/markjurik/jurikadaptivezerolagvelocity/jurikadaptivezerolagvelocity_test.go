//nolint:testpackage
package jurikadaptivezerolagvelocity

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

func testAZLV(t *testing.T, loLength, hiLength int, sensitivity, period float64, expected []float64) {
	t.Helper()

	ind, err := NewJurikAdaptiveZeroLagVelocity(&JurikAdaptiveZeroLagVelocityParams{
		LoLength:    loLength,
		HiLength:    hiLength,
		Sensitivity: sensitivity,
		Period:      period,
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

func TestAZLVLo2Hi15(t *testing.T)   { t.Parallel(); testAZLV(t, 2, 15, 1.0, 3.0, expectedLo2Hi15) }
func TestAZLVLo2Hi30(t *testing.T)   { t.Parallel(); testAZLV(t, 2, 30, 1.0, 3.0, expectedLo2Hi30) }
func TestAZLVLo2Hi60(t *testing.T)   { t.Parallel(); testAZLV(t, 2, 60, 1.0, 3.0, expectedLo2Hi60) }
func TestAZLVLo5Hi15(t *testing.T)   { t.Parallel(); testAZLV(t, 5, 15, 1.0, 3.0, expectedLo5Hi15) }
func TestAZLVLo5Hi30(t *testing.T)   { t.Parallel(); testAZLV(t, 5, 30, 1.0, 3.0, expectedLo5Hi30) }
func TestAZLVLo5Hi60(t *testing.T)   { t.Parallel(); testAZLV(t, 5, 60, 1.0, 3.0, expectedLo5Hi60) }
func TestAZLVLo10Hi15(t *testing.T)  { t.Parallel(); testAZLV(t, 10, 15, 1.0, 3.0, expectedLo10Hi15) }
func TestAZLVLo10Hi30(t *testing.T)  { t.Parallel(); testAZLV(t, 10, 30, 1.0, 3.0, expectedLo10Hi30) }
func TestAZLVLo10Hi60(t *testing.T)  { t.Parallel(); testAZLV(t, 10, 60, 1.0, 3.0, expectedLo10Hi60) }
func TestAZLVSens05(t *testing.T)    { t.Parallel(); testAZLV(t, 5, 30, 0.5, 3.0, expectedSens05) }
func TestAZLVSens25(t *testing.T)    { t.Parallel(); testAZLV(t, 5, 30, 2.5, 3.0, expectedSens25) }
func TestAZLVSens50(t *testing.T)    { t.Parallel(); testAZLV(t, 5, 30, 5.0, 3.0, expectedSens50) }
func TestAZLVPeriod15(t *testing.T)  { t.Parallel(); testAZLV(t, 5, 30, 1.0, 1.5, expectedPeriod15) }
func TestAZLVPeriod100(t *testing.T) { t.Parallel(); testAZLV(t, 5, 30, 1.0, 10.0, expectedPeriod100) }
func TestAZLVPeriod300(t *testing.T) { t.Parallel(); testAZLV(t, 5, 30, 1.0, 30.0, expectedPeriod300) }
