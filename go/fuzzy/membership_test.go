package fuzzy

import (
	"math"
	"testing"
)

func almostEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

// -- MuLess / MuLessEqual --

func TestMuLess(t *testing.T) {
	t.Run("crossover_at_threshold", func(t *testing.T) {
		if !almostEqual(MuLess(10.0, 10.0, 2.0, Sigmoid), 0.5, 1e-10) {
			t.Error("expected 0.5 at threshold")
		}
	})
	t.Run("well_below_threshold", func(t *testing.T) {
		val := MuLess(8.0, 10.0, 2.0, Sigmoid)
		if val <= 0.99 {
			t.Errorf("expected > 0.99, got %f", val)
		}
	})
	t.Run("well_above_threshold", func(t *testing.T) {
		val := MuLess(12.0, 10.0, 2.0, Sigmoid)
		if val >= 0.01 {
			t.Errorf("expected < 0.01, got %f", val)
		}
	})
	t.Run("monotonically_decreasing", func(t *testing.T) {
		xs := []float64{8.0, 9.0, 10.0, 11.0, 12.0}
		for i := 0; i < len(xs)-1; i++ {
			if MuLess(xs[i], 10.0, 2.0, Sigmoid) <= MuLess(xs[i+1], 10.0, 2.0, Sigmoid) {
				t.Error("not monotonically decreasing")
			}
		}
	})
	t.Run("symmetry", func(t *testing.T) {
		below := MuLess(9.0, 10.0, 2.0, Sigmoid)
		above := MuLess(11.0, 10.0, 2.0, Sigmoid)
		if !almostEqual(below+above, 1.0, 1e-10) {
			t.Errorf("expected symmetry, got %f + %f", below, above)
		}
	})
	t.Run("linear_crossover", func(t *testing.T) {
		if !almostEqual(MuLess(10.0, 10.0, 4.0, Linear), 0.5, 1e-10) {
			t.Error("expected 0.5")
		}
	})
	t.Run("linear_below_range", func(t *testing.T) {
		if MuLess(7.0, 10.0, 4.0, Linear) != 1.0 {
			t.Error("expected 1.0")
		}
	})
	t.Run("linear_above_range", func(t *testing.T) {
		if MuLess(13.0, 10.0, 4.0, Linear) != 0.0 {
			t.Error("expected 0.0")
		}
	})
	t.Run("linear_midpoint", func(t *testing.T) {
		if !almostEqual(MuLess(9.0, 10.0, 4.0, Linear), 0.75, 1e-10) {
			t.Error("expected 0.75")
		}
	})
	t.Run("crisp_below", func(t *testing.T) {
		if MuLess(9.0, 10.0, 0.0, Sigmoid) != 1.0 {
			t.Error("expected 1.0")
		}
	})
	t.Run("crisp_above", func(t *testing.T) {
		if MuLess(11.0, 10.0, 0.0, Sigmoid) != 0.0 {
			t.Error("expected 0.0")
		}
	})
	t.Run("crisp_at_threshold", func(t *testing.T) {
		if MuLess(10.0, 10.0, 0.0, Sigmoid) != 0.5 {
			t.Error("expected 0.5")
		}
	})
	t.Run("less_equal_same_as_less", func(t *testing.T) {
		if MuLessEqual(9.5, 10.0, 2.0, Sigmoid) != MuLess(9.5, 10.0, 2.0, Sigmoid) {
			t.Error("expected identical")
		}
	})
}

// -- MuGreater / MuGreaterEqual --

func TestMuGreater(t *testing.T) {
	t.Run("complement_of_less", func(t *testing.T) {
		for _, x := range []float64{8.0, 9.0, 10.0, 11.0, 12.0} {
			sum := MuGreater(x, 10.0, 2.0, Sigmoid) + MuLess(x, 10.0, 2.0, Sigmoid)
			if !almostEqual(sum, 1.0, 1e-10) {
				t.Errorf("expected complement at x=%f, got sum=%f", x, sum)
			}
		}
	})
	t.Run("crossover", func(t *testing.T) {
		if !almostEqual(MuGreater(10.0, 10.0, 2.0, Sigmoid), 0.5, 1e-10) {
			t.Error("expected 0.5")
		}
	})
	t.Run("well_above", func(t *testing.T) {
		if MuGreater(12.0, 10.0, 2.0, Sigmoid) <= 0.99 {
			t.Error("expected > 0.99")
		}
	})
	t.Run("well_below", func(t *testing.T) {
		if MuGreater(8.0, 10.0, 2.0, Sigmoid) >= 0.01 {
			t.Error("expected < 0.01")
		}
	})
	t.Run("greater_equal_complement", func(t *testing.T) {
		sum := MuGreaterEqual(9.5, 10.0, 2.0, Sigmoid) + MuLessEqual(9.5, 10.0, 2.0, Sigmoid)
		if !almostEqual(sum, 1.0, 1e-10) {
			t.Error("expected complement")
		}
	})
}

// -- MuNear --

func TestMuNear(t *testing.T) {
	t.Run("peak_at_target", func(t *testing.T) {
		if !almostEqual(MuNear(10.0, 10.0, 2.0, Sigmoid), 1.0, 1e-10) {
			t.Error("expected 1.0")
		}
	})
	t.Run("falls_off", func(t *testing.T) {
		val := MuNear(12.0, 10.0, 2.0, Sigmoid)
		if val >= 0.05 {
			t.Errorf("expected < 0.05, got %f", val)
		}
	})
	t.Run("symmetric", func(t *testing.T) {
		below := MuNear(9.0, 10.0, 2.0, Sigmoid)
		above := MuNear(11.0, 10.0, 2.0, Sigmoid)
		if !almostEqual(below, above, 1e-10) {
			t.Error("expected symmetric")
		}
	})
	t.Run("monotonic_from_center", func(t *testing.T) {
		for _, d := range []float64{0, 0.5, 1.0, 1.5} {
			v1 := MuNear(10.0+d, 10.0, 2.0, Sigmoid)
			v2 := MuNear(10.0+d+0.5, 10.0, 2.0, Sigmoid)
			if v1 <= v2 {
				t.Error("not monotonically decreasing from center")
			}
		}
	})
	t.Run("linear_peak", func(t *testing.T) {
		if !almostEqual(MuNear(10.0, 10.0, 2.0, Linear), 1.0, 1e-10) {
			t.Error("expected 1.0")
		}
	})
	t.Run("linear_at_boundary", func(t *testing.T) {
		if MuNear(12.0, 10.0, 2.0, Linear) != 0.0 {
			t.Error("expected 0.0")
		}
	})
	t.Run("linear_midpoint", func(t *testing.T) {
		if !almostEqual(MuNear(11.0, 10.0, 2.0, Linear), 0.5, 1e-10) {
			t.Error("expected 0.5")
		}
	})
	t.Run("crisp_exact", func(t *testing.T) {
		if MuNear(10.0, 10.0, 0.0, Sigmoid) != 1.0 {
			t.Error("expected 1.0")
		}
	})
	t.Run("crisp_any_distance", func(t *testing.T) {
		if MuNear(10.1, 10.0, 0.0, Sigmoid) != 0.0 {
			t.Error("expected 0.0")
		}
	})
}

// -- MuDirection --

func TestMuDirection(t *testing.T) {
	t.Run("large_white_body", func(t *testing.T) {
		if MuDirection(100.0, 110.0, 5.0, 2.0) <= 0.95 {
			t.Error("expected > 0.95")
		}
	})
	t.Run("large_black_body", func(t *testing.T) {
		if MuDirection(110.0, 100.0, 5.0, 2.0) >= -0.95 {
			t.Error("expected < -0.95")
		}
	})
	t.Run("doji", func(t *testing.T) {
		if !almostEqual(MuDirection(100.0, 100.0, 5.0, 2.0), 0.0, 1e-10) {
			t.Error("expected 0.0")
		}
	})
	t.Run("tiny_white_body", func(t *testing.T) {
		d := MuDirection(100.0, 100.1, 5.0, 2.0)
		if d <= 0.0 || d >= 0.1 {
			t.Errorf("expected in (0, 0.1), got %f", d)
		}
	})
	t.Run("antisymmetric", func(t *testing.T) {
		d1 := MuDirection(100.0, 105.0, 5.0, 2.0)
		d2 := MuDirection(105.0, 100.0, 5.0, 2.0)
		if !almostEqual(d1, -d2, 1e-10) {
			t.Error("expected antisymmetric")
		}
	})
	t.Run("zero_body_avg_white", func(t *testing.T) {
		if MuDirection(100.0, 101.0, 0.0, 2.0) != 1.0 {
			t.Error("expected 1.0")
		}
	})
	t.Run("zero_body_avg_black", func(t *testing.T) {
		if MuDirection(101.0, 100.0, 0.0, 2.0) != -1.0 {
			t.Error("expected -1.0")
		}
	})
	t.Run("zero_body_avg_doji", func(t *testing.T) {
		if MuDirection(100.0, 100.0, 0.0, 2.0) != 1.0 {
			t.Error("expected 1.0")
		}
	})
	t.Run("range_bounded", func(t *testing.T) {
		cases := [][3]float64{{0, 1000, 1}, {1000, 0, 1}, {50, 50, 100}}
		for _, c := range cases {
			d := MuDirection(c[0], c[1], c[2], 2.0)
			if d < -1.0 || d > 1.0 {
				t.Errorf("out of range: %f", d)
			}
		}
	})
}

// -- Edge Cases --

func TestEdgeCases(t *testing.T) {
	t.Run("very_large_x", func(t *testing.T) {
		if MuLess(1e10, 0.0, 1.0, Sigmoid) != 0.0 {
			t.Error("expected 0.0")
		}
	})
	t.Run("very_small_x", func(t *testing.T) {
		if MuLess(-1e10, 0.0, 1.0, Sigmoid) != 1.0 {
			t.Error("expected 1.0")
		}
	})
	t.Run("tiny_width", func(t *testing.T) {
		if MuLess(9.999, 10.0, 0.001, Sigmoid) <= 0.99 {
			t.Error("expected > 0.99")
		}
	})
	t.Run("huge_width", func(t *testing.T) {
		val := MuLess(0.0, 10.0, 1000.0, Sigmoid)
		if val <= 0.49 || val >= 0.60 {
			t.Errorf("expected in (0.49, 0.60), got %f", val)
		}
	})
}
