//nolint:testpackage
package jurikrelativetrendstrengthindex

import (
	"math"
	"testing"
)

func almostEqual(a, b, epsilon float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}

	if math.IsNaN(a) || math.IsNaN(b) {
		return false
	}

	return math.Abs(a-b) < epsilon
}

const epsilon = 1e-13

func testRSX(t *testing.T, length int, expected []float64) {
	t.Helper()

	ind, err := NewJurikRelativeTrendStrengthIndex(&JurikRelativeTrendStrengthIndexParams{Length: length})
	if err != nil {
		t.Fatal(err)
	}

	input := testInput
	for i, sample := range input {
		result := ind.Update(sample)
		if !almostEqual(result, expected[i], epsilon) {
			t.Errorf("bar %d: got %.15f, want %.15f", i, result, expected[i])
		}
	}
}

func TestRSXLength2(t *testing.T)  { t.Parallel(); testRSX(t, 2, expectedLength2) }
func TestRSXLength3(t *testing.T)  { t.Parallel(); testRSX(t, 3, expectedLength3) }
func TestRSXLength4(t *testing.T)  { t.Parallel(); testRSX(t, 4, expectedLength4) }
func TestRSXLength5(t *testing.T)  { t.Parallel(); testRSX(t, 5, expectedLength5) }
func TestRSXLength6(t *testing.T)  { t.Parallel(); testRSX(t, 6, expectedLength6) }
func TestRSXLength7(t *testing.T)  { t.Parallel(); testRSX(t, 7, expectedLength7) }
func TestRSXLength8(t *testing.T)  { t.Parallel(); testRSX(t, 8, expectedLength8) }
func TestRSXLength9(t *testing.T)  { t.Parallel(); testRSX(t, 9, expectedLength9) }
func TestRSXLength10(t *testing.T) { t.Parallel(); testRSX(t, 10, expectedLength10) }
func TestRSXLength11(t *testing.T) { t.Parallel(); testRSX(t, 11, expectedLength11) }
func TestRSXLength12(t *testing.T) { t.Parallel(); testRSX(t, 12, expectedLength12) }
func TestRSXLength13(t *testing.T) { t.Parallel(); testRSX(t, 13, expectedLength13) }
func TestRSXLength14(t *testing.T) { t.Parallel(); testRSX(t, 14, expectedLength14) }
func TestRSXLength15(t *testing.T) { t.Parallel(); testRSX(t, 15, expectedLength15) }
