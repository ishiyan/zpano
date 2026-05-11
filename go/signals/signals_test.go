package signals

import (
	"math"
	"testing"

	"zpano/fuzzy"
)

func almostEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

// -- Threshold --

func TestMuAbove(t *testing.T) {
	t.Run("well_above", func(t *testing.T) {
		if !almostEqual(MuAbove(80.0, 70.0, 5.0, fuzzy.Sigmoid), 1.0, 0.01) {
			t.Error("expected ~1.0")
		}
	})
	t.Run("well_below", func(t *testing.T) {
		if !almostEqual(MuAbove(60.0, 70.0, 5.0, fuzzy.Sigmoid), 0.0, 0.01) {
			t.Error("expected ~0.0")
		}
	})
	t.Run("at_threshold", func(t *testing.T) {
		if !almostEqual(MuAbove(70.0, 70.0, 5.0, fuzzy.Sigmoid), 0.5, 1e-10) {
			t.Error("expected 0.5")
		}
	})
	t.Run("zero_width_above", func(t *testing.T) {
		if MuAbove(70.1, 70.0, 0.0, fuzzy.Sigmoid) != 1.0 {
			t.Error("expected 1.0")
		}
	})
	t.Run("zero_width_below", func(t *testing.T) {
		if MuAbove(69.9, 70.0, 0.0, fuzzy.Sigmoid) != 0.0 {
			t.Error("expected 0.0")
		}
	})
	t.Run("zero_width_equal", func(t *testing.T) {
		if !almostEqual(MuAbove(70.0, 70.0, 0.0, fuzzy.Sigmoid), 0.5, 1e-10) {
			t.Error("expected 0.5")
		}
	})
	t.Run("monotonic", func(t *testing.T) {
		m1 := MuAbove(68.0, 70.0, 5.0, fuzzy.Sigmoid)
		m2 := MuAbove(70.0, 70.0, 5.0, fuzzy.Sigmoid)
		m3 := MuAbove(72.0, 70.0, 5.0, fuzzy.Sigmoid)
		if m1 >= m2 || m2 >= m3 {
			t.Error("expected monotonically increasing")
		}
	})
	t.Run("linear_shape", func(t *testing.T) {
		if !almostEqual(MuAbove(70.0, 70.0, 10.0, fuzzy.Linear), 0.5, 1e-10) {
			t.Error("expected 0.5")
		}
		if !almostEqual(MuAbove(65.0, 70.0, 10.0, fuzzy.Linear), 0.0, 1e-10) {
			t.Error("expected 0.0")
		}
		if !almostEqual(MuAbove(75.0, 70.0, 10.0, fuzzy.Linear), 1.0, 1e-10) {
			t.Error("expected 1.0")
		}
	})
}

func TestMuBelow(t *testing.T) {
	t.Run("well_below", func(t *testing.T) {
		if !almostEqual(MuBelow(20.0, 30.0, 5.0, fuzzy.Sigmoid), 1.0, 0.01) {
			t.Error("expected ~1.0")
		}
	})
	t.Run("well_above", func(t *testing.T) {
		if !almostEqual(MuBelow(40.0, 30.0, 5.0, fuzzy.Sigmoid), 0.0, 0.01) {
			t.Error("expected ~0.0")
		}
	})
	t.Run("at_threshold", func(t *testing.T) {
		if !almostEqual(MuBelow(30.0, 30.0, 5.0, fuzzy.Sigmoid), 0.5, 1e-10) {
			t.Error("expected 0.5")
		}
	})
	t.Run("complement_of_above", func(t *testing.T) {
		for _, v := range []float64{25.0, 30.0, 35.0, 50.0} {
			total := MuBelow(v, 30.0, 5.0, fuzzy.Sigmoid) + MuAbove(v, 30.0, 5.0, fuzzy.Sigmoid)
			if !almostEqual(total, 1.0, 1e-10) {
				t.Errorf("expected complement at v=%f", v)
			}
		}
	})
}

func TestOverboughtOversold(t *testing.T) {
	t.Run("overbought_high_rsi", func(t *testing.T) {
		if MuOverbought(85.0, 70.0, 5.0, fuzzy.Sigmoid) <= 0.95 {
			t.Error("expected > 0.95")
		}
	})
	t.Run("overbought_low_rsi", func(t *testing.T) {
		if MuOverbought(50.0, 70.0, 5.0, fuzzy.Sigmoid) >= 0.01 {
			t.Error("expected < 0.01")
		}
	})
	t.Run("oversold_low_rsi", func(t *testing.T) {
		if MuOversold(15.0, 30.0, 5.0, fuzzy.Sigmoid) <= 0.95 {
			t.Error("expected > 0.95")
		}
	})
	t.Run("oversold_high_rsi", func(t *testing.T) {
		if MuOversold(50.0, 30.0, 5.0, fuzzy.Sigmoid) >= 0.01 {
			t.Error("expected < 0.01")
		}
	})
	t.Run("overbought_custom_level", func(t *testing.T) {
		if !almostEqual(MuOverbought(80.0, 80.0, 5.0, fuzzy.Sigmoid), 0.5, 1e-10) {
			t.Error("expected 0.5")
		}
	})
	t.Run("oversold_custom_level", func(t *testing.T) {
		if !almostEqual(MuOversold(20.0, 20.0, 5.0, fuzzy.Sigmoid), 0.5, 1e-10) {
			t.Error("expected 0.5")
		}
	})
}

// -- Crossover --

func TestCrossesAbove(t *testing.T) {
	t.Run("clear_cross_above", func(t *testing.T) {
		if !almostEqual(MuCrossesAbove(25.0, 35.0, 30.0, 0.0, fuzzy.Sigmoid), 1.0, 1e-10) {
			t.Error("expected 1.0")
		}
	})
	t.Run("no_cross_both_above", func(t *testing.T) {
		if !almostEqual(MuCrossesAbove(35.0, 40.0, 30.0, 0.0, fuzzy.Sigmoid), 0.0, 1e-10) {
			t.Error("expected 0.0")
		}
	})
	t.Run("no_cross_both_below", func(t *testing.T) {
		if !almostEqual(MuCrossesAbove(25.0, 28.0, 30.0, 0.0, fuzzy.Sigmoid), 0.0, 1e-10) {
			t.Error("expected 0.0")
		}
	})
	t.Run("cross_down_not_up", func(t *testing.T) {
		if !almostEqual(MuCrossesAbove(35.0, 25.0, 30.0, 0.0, fuzzy.Sigmoid), 0.0, 1e-10) {
			t.Error("expected 0.0")
		}
	})
	t.Run("fuzzy_near_threshold", func(t *testing.T) {
		result := MuCrossesAbove(29.0, 31.0, 30.0, 5.0, fuzzy.Sigmoid)
		if result <= 0.1 || result >= 0.9 {
			t.Errorf("expected moderate, got %f", result)
		}
	})
	t.Run("at_threshold", func(t *testing.T) {
		if !almostEqual(MuCrossesAbove(30.0, 30.0, 30.0, 0.0, fuzzy.Sigmoid), 0.25, 1e-10) {
			t.Error("expected 0.25")
		}
	})
}

func TestCrossesBelow(t *testing.T) {
	t.Run("clear_cross_below", func(t *testing.T) {
		if !almostEqual(MuCrossesBelow(35.0, 25.0, 30.0, 0.0, fuzzy.Sigmoid), 1.0, 1e-10) {
			t.Error("expected 1.0")
		}
	})
	t.Run("no_cross_both_below", func(t *testing.T) {
		if !almostEqual(MuCrossesBelow(25.0, 20.0, 30.0, 0.0, fuzzy.Sigmoid), 0.0, 1e-10) {
			t.Error("expected 0.0")
		}
	})
	t.Run("symmetry", func(t *testing.T) {
		cb := MuCrossesBelow(35.0, 25.0, 30.0, 2.0, fuzzy.Sigmoid)
		ca := MuCrossesAbove(25.0, 35.0, 30.0, 2.0, fuzzy.Sigmoid)
		if !almostEqual(cb, ca, 1e-10) {
			t.Error("expected symmetry")
		}
	})
}

func TestLineCrossesAbove(t *testing.T) {
	t.Run("golden_cross", func(t *testing.T) {
		if !almostEqual(MuLineCrossesAbove(49.0, 51.0, 50.0, 50.0, 0.0, fuzzy.Sigmoid), 1.0, 1e-10) {
			t.Error("expected 1.0")
		}
	})
	t.Run("no_cross", func(t *testing.T) {
		if !almostEqual(MuLineCrossesAbove(52.0, 53.0, 50.0, 50.0, 0.0, fuzzy.Sigmoid), 0.0, 1e-10) {
			t.Error("expected 0.0")
		}
	})
	t.Run("fuzzy_near_cross", func(t *testing.T) {
		result := MuLineCrossesAbove(49.5, 50.5, 50.0, 50.0, 2.0, fuzzy.Sigmoid)
		if result <= 0.0 || result >= 1.0 {
			t.Errorf("expected moderate, got %f", result)
		}
	})
}

func TestLineCrossesBelow(t *testing.T) {
	t.Run("death_cross", func(t *testing.T) {
		if !almostEqual(MuLineCrossesBelow(51.0, 49.0, 50.0, 50.0, 0.0, fuzzy.Sigmoid), 1.0, 1e-10) {
			t.Error("expected 1.0")
		}
	})
	t.Run("no_cross", func(t *testing.T) {
		if !almostEqual(MuLineCrossesBelow(48.0, 47.0, 50.0, 50.0, 0.0, fuzzy.Sigmoid), 0.0, 1e-10) {
			t.Error("expected 0.0")
		}
	})
}

// -- Band --

func TestAboveBand(t *testing.T) {
	t.Run("well_above", func(t *testing.T) {
		if !almostEqual(MuAboveBand(110.0, 100.0, 5.0, fuzzy.Sigmoid), 1.0, 0.01) {
			t.Error("expected ~1.0")
		}
	})
	t.Run("well_below", func(t *testing.T) {
		if !almostEqual(MuAboveBand(90.0, 100.0, 5.0, fuzzy.Sigmoid), 0.0, 0.01) {
			t.Error("expected ~0.0")
		}
	})
	t.Run("at_band", func(t *testing.T) {
		if !almostEqual(MuAboveBand(100.0, 100.0, 5.0, fuzzy.Sigmoid), 0.5, 1e-10) {
			t.Error("expected 0.5")
		}
	})
	t.Run("crisp", func(t *testing.T) {
		if MuAboveBand(100.1, 100.0, 0.0, fuzzy.Sigmoid) != 1.0 {
			t.Error("expected 1.0")
		}
		if MuAboveBand(99.9, 100.0, 0.0, fuzzy.Sigmoid) != 0.0 {
			t.Error("expected 0.0")
		}
	})
}

func TestBelowBand(t *testing.T) {
	t.Run("well_below", func(t *testing.T) {
		if !almostEqual(MuBelowBand(85.0, 90.0, 5.0, fuzzy.Sigmoid), 1.0, 0.01) {
			t.Error("expected ~1.0")
		}
	})
	t.Run("well_above", func(t *testing.T) {
		if !almostEqual(MuBelowBand(100.0, 90.0, 5.0, fuzzy.Sigmoid), 0.0, 0.01) {
			t.Error("expected ~0.0")
		}
	})
	t.Run("at_band", func(t *testing.T) {
		if !almostEqual(MuBelowBand(90.0, 90.0, 5.0, fuzzy.Sigmoid), 0.5, 1e-10) {
			t.Error("expected 0.5")
		}
	})
}

func TestBetweenBands(t *testing.T) {
	t.Run("centered", func(t *testing.T) {
		if MuBetweenBands(100.0, 90.0, 110.0, fuzzy.Sigmoid) <= 0.8 {
			t.Error("expected > 0.8")
		}
	})
	t.Run("at_upper_band", func(t *testing.T) {
		if MuBetweenBands(110.0, 90.0, 110.0, fuzzy.Sigmoid) >= 0.6 {
			t.Error("expected < 0.6")
		}
	})
	t.Run("at_lower_band", func(t *testing.T) {
		if MuBetweenBands(90.0, 90.0, 110.0, fuzzy.Sigmoid) >= 0.6 {
			t.Error("expected < 0.6")
		}
	})
	t.Run("outside_above", func(t *testing.T) {
		if MuBetweenBands(130.0, 90.0, 110.0, fuzzy.Sigmoid) >= 0.1 {
			t.Error("expected < 0.1")
		}
	})
	t.Run("outside_below", func(t *testing.T) {
		if MuBetweenBands(70.0, 90.0, 110.0, fuzzy.Sigmoid) >= 0.1 {
			t.Error("expected < 0.1")
		}
	})
	t.Run("degenerate_bands", func(t *testing.T) {
		if MuBetweenBands(100.0, 110.0, 90.0, fuzzy.Sigmoid) != 0.0 {
			t.Error("expected 0.0")
		}
		if MuBetweenBands(100.0, 100.0, 100.0, fuzzy.Sigmoid) != 0.0 {
			t.Error("expected 0.0")
		}
	})
	t.Run("monotonic_from_center", func(t *testing.T) {
		center := MuBetweenBands(100.0, 90.0, 110.0, fuzzy.Sigmoid)
		edge := MuBetweenBands(108.0, 90.0, 110.0, fuzzy.Sigmoid)
		outside := MuBetweenBands(115.0, 90.0, 110.0, fuzzy.Sigmoid)
		if center <= edge || edge <= outside {
			t.Error("expected monotonically decreasing from center")
		}
	})
}

// -- Histogram --

func TestTurnsPositive(t *testing.T) {
	t.Run("clear_turn_positive", func(t *testing.T) {
		if !almostEqual(MuTurnsPositive(-5.0, 5.0, 0.0, fuzzy.Sigmoid), 1.0, 1e-10) {
			t.Error("expected 1.0")
		}
	})
	t.Run("stays_positive", func(t *testing.T) {
		if !almostEqual(MuTurnsPositive(3.0, 5.0, 0.0, fuzzy.Sigmoid), 0.0, 1e-10) {
			t.Error("expected 0.0")
		}
	})
	t.Run("stays_negative", func(t *testing.T) {
		if !almostEqual(MuTurnsPositive(-5.0, -3.0, 0.0, fuzzy.Sigmoid), 0.0, 1e-10) {
			t.Error("expected 0.0")
		}
	})
	t.Run("turns_more_negative", func(t *testing.T) {
		if !almostEqual(MuTurnsPositive(5.0, -5.0, 0.0, fuzzy.Sigmoid), 0.0, 1e-10) {
			t.Error("expected 0.0")
		}
	})
	t.Run("from_zero", func(t *testing.T) {
		if !almostEqual(MuTurnsPositive(0.0, 5.0, 0.0, fuzzy.Sigmoid), 0.5, 1e-10) {
			t.Error("expected 0.5")
		}
	})
	t.Run("fuzzy_near_zero", func(t *testing.T) {
		result := MuTurnsPositive(-0.5, 0.5, 2.0, fuzzy.Sigmoid)
		if result <= 0.1 || result >= 0.95 {
			t.Errorf("expected moderate, got %f", result)
		}
	})
	t.Run("fuzzy_width_makes_softer", func(t *testing.T) {
		narrow := MuTurnsPositive(-1.0, 1.0, 0.5, fuzzy.Sigmoid)
		wide := MuTurnsPositive(-1.0, 1.0, 10.0, fuzzy.Sigmoid)
		if narrow <= wide {
			t.Error("expected narrow > wide")
		}
	})
}

func TestTurnsNegative(t *testing.T) {
	t.Run("clear_turn_negative", func(t *testing.T) {
		if !almostEqual(MuTurnsNegative(5.0, -5.0, 0.0, fuzzy.Sigmoid), 1.0, 1e-10) {
			t.Error("expected 1.0")
		}
	})
	t.Run("stays_negative", func(t *testing.T) {
		if !almostEqual(MuTurnsNegative(-5.0, -3.0, 0.0, fuzzy.Sigmoid), 0.0, 1e-10) {
			t.Error("expected 0.0")
		}
	})
	t.Run("stays_positive", func(t *testing.T) {
		if !almostEqual(MuTurnsNegative(3.0, 5.0, 0.0, fuzzy.Sigmoid), 0.0, 1e-10) {
			t.Error("expected 0.0")
		}
	})
	t.Run("symmetry", func(t *testing.T) {
		tn := MuTurnsNegative(3.0, -3.0, 1.0, fuzzy.Sigmoid)
		tp := MuTurnsPositive(-3.0, 3.0, 1.0, fuzzy.Sigmoid)
		if !almostEqual(tn, tp, 1e-10) {
			t.Error("expected symmetry")
		}
	})
}

// -- Compose --

func TestSignalAnd(t *testing.T) {
	t.Run("all_high", func(t *testing.T) {
		if !almostEqual(SignalAnd(0.9, 0.8, 0.95), 0.9*0.8*0.95, 1e-10) {
			t.Error("expected product")
		}
	})
	t.Run("one_zero", func(t *testing.T) {
		if !almostEqual(SignalAnd(0.9, 0.0, 0.8), 0.0, 1e-10) {
			t.Error("expected 0.0")
		}
	})
	t.Run("all_one", func(t *testing.T) {
		if !almostEqual(SignalAnd(1.0, 1.0, 1.0), 1.0, 1e-10) {
			t.Error("expected 1.0")
		}
	})
	t.Run("two_args", func(t *testing.T) {
		if !almostEqual(SignalAnd(0.6, 0.7), 0.42, 1e-10) {
			t.Error("expected 0.42")
		}
	})
}

func TestSignalOr(t *testing.T) {
	t.Run("both_high", func(t *testing.T) {
		if !almostEqual(SignalOr(0.8, 0.9), 0.8+0.9-0.8*0.9, 1e-10) {
			t.Error("expected probabilistic sum")
		}
	})
	t.Run("one_zero", func(t *testing.T) {
		if !almostEqual(SignalOr(0.0, 0.7), 0.7, 1e-10) {
			t.Error("expected 0.7")
		}
	})
	t.Run("both_zero", func(t *testing.T) {
		if !almostEqual(SignalOr(0.0, 0.0), 0.0, 1e-10) {
			t.Error("expected 0.0")
		}
	})
	t.Run("both_one", func(t *testing.T) {
		if !almostEqual(SignalOr(1.0, 1.0), 1.0, 1e-10) {
			t.Error("expected 1.0")
		}
	})
	t.Run("greater_than_either", func(t *testing.T) {
		a, b := 0.6, 0.7
		result := SignalOr(a, b)
		if result < math.Max(a, b) {
			t.Error("expected >= max")
		}
	})
}

func TestSignalNot(t *testing.T) {
	t.Run("zero", func(t *testing.T) {
		if !almostEqual(SignalNot(0.0), 1.0, 1e-10) {
			t.Error("expected 1.0")
		}
	})
	t.Run("one", func(t *testing.T) {
		if !almostEqual(SignalNot(1.0), 0.0, 1e-10) {
			t.Error("expected 0.0")
		}
	})
	t.Run("half", func(t *testing.T) {
		if !almostEqual(SignalNot(0.5), 0.5, 1e-10) {
			t.Error("expected 0.5")
		}
	})
	t.Run("complement", func(t *testing.T) {
		for _, v := range []float64{0.0, 0.3, 0.5, 0.7, 1.0} {
			if !almostEqual(SignalNot(v), 1.0-v, 1e-10) {
				t.Errorf("expected complement at v=%f", v)
			}
		}
	})
}

func TestSignalStrength(t *testing.T) {
	t.Run("above_threshold", func(t *testing.T) {
		if SignalStrength(0.8, 0.5) != 0.8 {
			t.Error("expected 0.8")
		}
	})
	t.Run("below_threshold", func(t *testing.T) {
		if SignalStrength(0.3, 0.5) != 0.0 {
			t.Error("expected 0.0")
		}
	})
	t.Run("at_threshold", func(t *testing.T) {
		if SignalStrength(0.5, 0.5) != 0.5 {
			t.Error("expected 0.5")
		}
	})
	t.Run("just_below", func(t *testing.T) {
		if SignalStrength(0.499, 0.5) != 0.0 {
			t.Error("expected 0.0")
		}
	})
	t.Run("default_threshold", func(t *testing.T) {
		if SignalStrength(0.6, 0.5) != 0.6 {
			t.Error("expected 0.6")
		}
		if SignalStrength(0.4, 0.5) != 0.0 {
			t.Error("expected 0.0")
		}
	})
}
