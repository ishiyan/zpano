package jurikdirectionalmovementindex //nolint:testpackage

import (
	"math"
	"testing"
)

const epsilon = 1e-10

func almostEqual(a, b, eps float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}

	if math.IsNaN(a) || math.IsNaN(b) {
		return false
	}

	return math.Abs(a-b) < eps
}

func testDMX(t *testing.T, length int, expectedBipolar, expectedPlus, expectedMinus []float64) {
	t.Helper()

	ind, err := newJurikDirectionalMovementIndex(length)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	closeData := testInputClose()
	highData := testInputHigh()
	lowData := testInputLow()

	if len(closeData) != len(expectedBipolar) {
		t.Fatalf("data length mismatch: close=%d, expected=%d", len(closeData), len(expectedBipolar))
	}

	for i := range closeData {
		bipolar, plus, minus := ind.Update(highData[i], lowData[i], closeData[i])

		// First 41 bars (indices 0-40) are warmup: reference outputs 0.0, our streaming outputs NaN.
		if i <= 40 { //nolint:mnd
			if !math.IsNaN(bipolar) {
				t.Errorf("bar %d: bipolar expected NaN during warmup, got %v", i, bipolar)
			}

			continue
		}

		if !almostEqual(bipolar, expectedBipolar[i], epsilon) {
			t.Errorf("bar %d: bipolar expected %.15f, got %.15f (diff=%.2e)",
				i, expectedBipolar[i], bipolar, math.Abs(bipolar-expectedBipolar[i]))
		}

		if !almostEqual(plus, expectedPlus[i], epsilon) {
			t.Errorf("bar %d: plus expected %.15f, got %.15f (diff=%.2e)",
				i, expectedPlus[i], plus, math.Abs(plus-expectedPlus[i]))
		}

		if len(expectedMinus) > 0 && !almostEqual(minus, expectedMinus[i], epsilon) {
			t.Errorf("bar %d: minus expected %.15f, got %.15f (diff=%.2e)",
				i, expectedMinus[i], minus, math.Abs(minus-expectedMinus[i]))
		}
	}
}

func TestDMXLength2(t *testing.T)  { testDMX(t, 2, dmxBipolarLen2, dmxPlusLen2, dmxMinusLen2) }
func TestDMXLength3(t *testing.T)  { testDMX(t, 3, dmxBipolarLen3, dmxPlusLen3, dmxMinusLen3) }
func TestDMXLength4(t *testing.T)  { testDMX(t, 4, dmxBipolarLen4, dmxPlusLen4, dmxMinusLen4) }
func TestDMXLength5(t *testing.T)  { testDMX(t, 5, dmxBipolarLen5, dmxPlusLen5, dmxMinusLen5) }
func TestDMXLength6(t *testing.T)  { testDMX(t, 6, dmxBipolarLen6, dmxPlusLen6, dmxMinusLen6) }
func TestDMXLength7(t *testing.T)  { testDMX(t, 7, dmxBipolarLen7, dmxPlusLen7, dmxMinusLen7) }
func TestDMXLength8(t *testing.T)  { testDMX(t, 8, dmxBipolarLen8, dmxPlusLen8, dmxMinusLen8) }
func TestDMXLength9(t *testing.T)  { testDMX(t, 9, dmxBipolarLen9, dmxPlusLen9, dmxMinusLen9) }
func TestDMXLength10(t *testing.T) { testDMX(t, 10, dmxBipolarLen10, dmxPlusLen10, dmxMinusLen10) }
func TestDMXLength11(t *testing.T) { testDMX(t, 11, dmxBipolarLen11, dmxPlusLen11, dmxMinusLen11) }
func TestDMXLength12(t *testing.T) { testDMX(t, 12, dmxBipolarLen12, dmxPlusLen12, dmxMinusLen12) }
func TestDMXLength13(t *testing.T) { testDMX(t, 13, dmxBipolarLen13, dmxPlusLen13, dmxMinusLen13) }
func TestDMXLength14(t *testing.T) { testDMX(t, 14, dmxBipolarLen14, dmxPlusLen14, dmxMinusLen14) }
func TestDMXLength15(t *testing.T) { testDMX(t, 15, dmxBipolarLen15, dmxPlusLen15, dmxMinusLen15) }
func TestDMXLength16(t *testing.T) { testDMX(t, 16, dmxBipolarLen16, dmxPlusLen16, dmxMinusLen16) }
func TestDMXLength17(t *testing.T) { testDMX(t, 17, dmxBipolarLen17, dmxPlusLen17, dmxMinusLen17) }
func TestDMXLength18(t *testing.T) { testDMX(t, 18, dmxBipolarLen18, dmxPlusLen18, dmxMinusLen18) }
func TestDMXLength19(t *testing.T) { testDMX(t, 19, dmxBipolarLen19, dmxPlusLen19, dmxMinusLen19) }
func TestDMXLength20(t *testing.T) { testDMX(t, 20, dmxBipolarLen20, dmxPlusLen20, dmxMinusLen20) }
