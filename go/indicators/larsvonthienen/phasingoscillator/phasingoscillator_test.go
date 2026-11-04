// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package phasingoscillator

import (
	"math"
	"testing"
)

func TestPreparedReferenceCases(t *testing.T) {
	maxError, maxScaledError := 0.0, 0.0
	defer func() { t.Logf("precision max absolute %.17g scaled %.17g", maxError, maxScaledError) }()
	t.Run("case0", func(t *testing.T) {
		p := DefaultParams()
		p.GenuineThreshold = 49.0
		p.LongMax = 80
		p.LongMin = 31
		p.PhaseCycles = 2.0
		p.ShortMax = 30
		p.ShortMin = 5
		p.Smoothing = 1600.0
		p.Window = 200
		indicator, err := NewPhasingOscillator(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedW200_S05_S130_L031_L180_HP1600_G49_F2_VALUE, "SHORT_SCORE": expectedW200_S05_S130_L031_L180_HP1600_G49_F2_SHORT_SCORE, "LONG_SCORE": expectedW200_S05_S130_L031_L180_HP1600_G49_F2_LONG_SCORE, "SHORT_PERIOD": expectedW200_S05_S130_L031_L180_HP1600_G49_F2_SHORT_PERIOD, "LONG_PERIOD": expectedW200_S05_S130_L031_L180_HP1600_G49_F2_LONG_PERIOD}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "SHORT_SCORE": result.ShortScore, "LONG_SCORE": result.LongScore, "SHORT_PERIOD": result.ShortPeriod, "LONG_PERIOD": result.LongPeriod}
			for field, values := range expected {
				value := actual[field]
				if !math.IsNaN(values[index]) && !math.IsNaN(value) {
					diff := math.Abs(value - values[index])
					maxError = max(maxError, diff)
					maxScaledError = max(maxScaledError, diff/max(1, math.Abs(values[index])))
				}
				if math.IsNaN(values[index]) {
					if !math.IsNaN(value) {
						t.Fatalf("bar %d %s expected NaN, got %.17g", index, field, value)
					}
				} else if math.IsNaN(value) || math.Abs(value-values[index]) > max(1e-10, math.Abs(values[index])*5e-15) {
					t.Fatalf("bar %d %s expected %.17g got %.17g", index, field, values[index], value)
				}
			}
		}
		indicator.Reset()
		if indicator.IsPrimed() {
			t.Fatal("reset must clear priming")
		}
	})
	t.Run("case1", func(t *testing.T) {
		p := DefaultParams()
		p.GenuineThreshold = 49.0
		p.LongMax = 40
		p.LongMin = 16
		p.PhaseCycles = 0.0
		p.ShortMax = 15
		p.ShortMin = 5
		p.Smoothing = 1600.0
		p.Window = 100
		indicator, err := NewPhasingOscillator(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedW100_S05_S115_L016_L140_HP1600_G49_F0_VALUE, "SHORT_SCORE": expectedW100_S05_S115_L016_L140_HP1600_G49_F0_SHORT_SCORE, "LONG_SCORE": expectedW100_S05_S115_L016_L140_HP1600_G49_F0_LONG_SCORE, "SHORT_PERIOD": expectedW100_S05_S115_L016_L140_HP1600_G49_F0_SHORT_PERIOD, "LONG_PERIOD": expectedW100_S05_S115_L016_L140_HP1600_G49_F0_LONG_PERIOD}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "SHORT_SCORE": result.ShortScore, "LONG_SCORE": result.LongScore, "SHORT_PERIOD": result.ShortPeriod, "LONG_PERIOD": result.LongPeriod}
			for field, values := range expected {
				value := actual[field]
				if !math.IsNaN(values[index]) && !math.IsNaN(value) {
					diff := math.Abs(value - values[index])
					maxError = max(maxError, diff)
					maxScaledError = max(maxScaledError, diff/max(1, math.Abs(values[index])))
				}
				if math.IsNaN(values[index]) {
					if !math.IsNaN(value) {
						t.Fatalf("bar %d %s expected NaN, got %.17g", index, field, value)
					}
				} else if math.IsNaN(value) || math.Abs(value-values[index]) > max(1e-10, math.Abs(values[index])*5e-15) {
					t.Fatalf("bar %d %s expected %.17g got %.17g", index, field, values[index], value)
				}
			}
		}
		indicator.Reset()
		if indicator.IsPrimed() {
			t.Fatal("reset must clear priming")
		}
	})
	t.Run("case2", func(t *testing.T) {
		p := DefaultParams()
		p.GenuineThreshold = 49.0
		p.LongMax = 40
		p.LongMin = 16
		p.PhaseCycles = 2.0
		p.ShortMax = 15
		p.ShortMin = 5
		p.Smoothing = 0.0
		p.Window = 100
		indicator, err := NewPhasingOscillator(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedW100_S05_S115_L016_L140_HP0_G49_F2_VALUE, "SHORT_SCORE": expectedW100_S05_S115_L016_L140_HP0_G49_F2_SHORT_SCORE, "LONG_SCORE": expectedW100_S05_S115_L016_L140_HP0_G49_F2_LONG_SCORE, "SHORT_PERIOD": expectedW100_S05_S115_L016_L140_HP0_G49_F2_SHORT_PERIOD, "LONG_PERIOD": expectedW100_S05_S115_L016_L140_HP0_G49_F2_LONG_PERIOD}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "SHORT_SCORE": result.ShortScore, "LONG_SCORE": result.LongScore, "SHORT_PERIOD": result.ShortPeriod, "LONG_PERIOD": result.LongPeriod}
			for field, values := range expected {
				value := actual[field]
				if !math.IsNaN(values[index]) && !math.IsNaN(value) {
					diff := math.Abs(value - values[index])
					maxError = max(maxError, diff)
					maxScaledError = max(maxScaledError, diff/max(1, math.Abs(values[index])))
				}
				if math.IsNaN(values[index]) {
					if !math.IsNaN(value) {
						t.Fatalf("bar %d %s expected NaN, got %.17g", index, field, value)
					}
				} else if math.IsNaN(value) || math.Abs(value-values[index]) > max(1e-10, math.Abs(values[index])*5e-15) {
					t.Fatalf("bar %d %s expected %.17g got %.17g", index, field, values[index], value)
				}
			}
		}
		indicator.Reset()
		if indicator.IsPrimed() {
			t.Fatal("reset must clear priming")
		}
	})
	t.Run("case3", func(t *testing.T) {
		p := DefaultParams()
		p.GenuineThreshold = 100.0
		p.LongMax = 80
		p.LongMin = 31
		p.PhaseCycles = 2.0
		p.ShortMax = 30
		p.ShortMin = 5
		p.Smoothing = 1600.0
		p.Window = 200
		indicator, err := NewPhasingOscillator(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedW200_S05_S130_L031_L180_HP1600_G100_F2_VALUE, "SHORT_SCORE": expectedW200_S05_S130_L031_L180_HP1600_G100_F2_SHORT_SCORE, "LONG_SCORE": expectedW200_S05_S130_L031_L180_HP1600_G100_F2_LONG_SCORE, "SHORT_PERIOD": expectedW200_S05_S130_L031_L180_HP1600_G100_F2_SHORT_PERIOD, "LONG_PERIOD": expectedW200_S05_S130_L031_L180_HP1600_G100_F2_LONG_PERIOD}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "SHORT_SCORE": result.ShortScore, "LONG_SCORE": result.LongScore, "SHORT_PERIOD": result.ShortPeriod, "LONG_PERIOD": result.LongPeriod}
			for field, values := range expected {
				value := actual[field]
				if !math.IsNaN(values[index]) && !math.IsNaN(value) {
					diff := math.Abs(value - values[index])
					maxError = max(maxError, diff)
					maxScaledError = max(maxScaledError, diff/max(1, math.Abs(values[index])))
				}
				if math.IsNaN(values[index]) {
					if !math.IsNaN(value) {
						t.Fatalf("bar %d %s expected NaN, got %.17g", index, field, value)
					}
				} else if math.IsNaN(value) || math.Abs(value-values[index]) > max(1e-10, math.Abs(values[index])*5e-15) {
					t.Fatalf("bar %d %s expected %.17g got %.17g", index, field, values[index], value)
				}
			}
		}
		indicator.Reset()
		if indicator.IsPrimed() {
			t.Fatal("reset must clear priming")
		}
	})
}
