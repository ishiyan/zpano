//nolint:testpackage
package jurikcommoditychannelindex

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

func testJCCX(t *testing.T, length int, expected []float64) {
	t.Helper()

	ind, err := NewJurikCommodityChannelIndex(&JurikCommodityChannelIndexParams{
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

func TestJCCXLen10(t *testing.T)  { t.Parallel(); testJCCX(t, 10, expectedLen10) }
func TestJCCXLen14(t *testing.T)  { t.Parallel(); testJCCX(t, 14, expectedLen14) }
func TestJCCXLen20(t *testing.T)  { t.Parallel(); testJCCX(t, 20, expectedLen20) }
func TestJCCXLen30(t *testing.T)  { t.Parallel(); testJCCX(t, 30, expectedLen30) }
func TestJCCXLen40(t *testing.T)  { t.Parallel(); testJCCX(t, 40, expectedLen40) }
func TestJCCXLen50(t *testing.T)  { t.Parallel(); testJCCX(t, 50, expectedLen50) }
func TestJCCXLen60(t *testing.T)  { t.Parallel(); testJCCX(t, 60, expectedLen60) }
func TestJCCXLen80(t *testing.T)  { t.Parallel(); testJCCX(t, 80, expectedLen80) }
func TestJCCXLen100(t *testing.T) { t.Parallel(); testJCCX(t, 100, expectedLen100) }
