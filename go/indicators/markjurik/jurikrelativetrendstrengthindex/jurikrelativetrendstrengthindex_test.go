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

func testJRSX(t *testing.T, length int, expected []float64) {
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

func TestJRSXLength2(t *testing.T)  { t.Parallel(); testJRSX(t, 2, expectedLength2) }
func TestJRSXLength3(t *testing.T)  { t.Parallel(); testJRSX(t, 3, expectedLength3) }
func TestJRSXLength4(t *testing.T)  { t.Parallel(); testJRSX(t, 4, expectedLength4) }
func TestJRSXLength5(t *testing.T)  { t.Parallel(); testJRSX(t, 5, expectedLength5) }
func TestJRSXLength6(t *testing.T)  { t.Parallel(); testJRSX(t, 6, expectedLength6) }
func TestJRSXLength7(t *testing.T)  { t.Parallel(); testJRSX(t, 7, expectedLength7) }
func TestJRSXLength8(t *testing.T)  { t.Parallel(); testJRSX(t, 8, expectedLength8) }
func TestJRSXLength9(t *testing.T)  { t.Parallel(); testJRSX(t, 9, expectedLength9) }
func TestJRSXLength10(t *testing.T) { t.Parallel(); testJRSX(t, 10, expectedLength10) }
func TestJRSXLength11(t *testing.T) { t.Parallel(); testJRSX(t, 11, expectedLength11) }
func TestJRSXLength12(t *testing.T) { t.Parallel(); testJRSX(t, 12, expectedLength12) }
func TestJRSXLength13(t *testing.T) { t.Parallel(); testJRSX(t, 13, expectedLength13) }
func TestJRSXLength14(t *testing.T) { t.Parallel(); testJRSX(t, 14, expectedLength14) }
func TestJRSXLength15(t *testing.T) { t.Parallel(); testJRSX(t, 15, expectedLength15) }
