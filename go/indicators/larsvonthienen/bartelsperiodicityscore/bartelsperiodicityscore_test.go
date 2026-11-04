// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package bartelsperiodicityscore

import (
	"math"
	"testing"
)

func TestPreparedReferenceCases(t *testing.T) {
	maxError, maxScaledError := 0.0, 0.0
	defer func() { t.Logf("precision max absolute %.17g scaled %.17g", maxError, maxScaledError) }()
	t.Run("case0", func(t *testing.T) {
		p := DefaultParams()
		p.Period = 20
		p.Segments = 5
		p.Smoothing = 0.0
		indicator, err := NewBartelsPeriodicityScore(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"SCORE": expectedP20_M5_L0_SCORE, "GENUINE_PERCENT": expectedP20_M5_L0_GENUINE_PERCENT, "SEGMENTS": expectedP20_M5_L0_SEGMENTS}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"SCORE": result.Score, "GENUINE_PERCENT": result.GenuinePercent, "SEGMENTS": result.Segments}
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
		p.Period = 3
		p.Segments = 2
		p.Smoothing = 0.0
		indicator, err := NewBartelsPeriodicityScore(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"SCORE": expectedP3_M2_L0_SCORE, "GENUINE_PERCENT": expectedP3_M2_L0_GENUINE_PERCENT, "SEGMENTS": expectedP3_M2_L0_SEGMENTS}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"SCORE": result.Score, "GENUINE_PERCENT": result.GenuinePercent, "SEGMENTS": result.Segments}
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
		p.Period = 10
		p.Segments = 5
		p.Smoothing = 0.0
		indicator, err := NewBartelsPeriodicityScore(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"SCORE": expectedP10_M5_L0_SCORE, "GENUINE_PERCENT": expectedP10_M5_L0_GENUINE_PERCENT, "SEGMENTS": expectedP10_M5_L0_SEGMENTS}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"SCORE": result.Score, "GENUINE_PERCENT": result.GenuinePercent, "SEGMENTS": result.Segments}
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
		p.Period = 20
		p.Segments = 2
		p.Smoothing = 0.0
		indicator, err := NewBartelsPeriodicityScore(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"SCORE": expectedP20_M2_L0_SCORE, "GENUINE_PERCENT": expectedP20_M2_L0_GENUINE_PERCENT, "SEGMENTS": expectedP20_M2_L0_SEGMENTS}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"SCORE": result.Score, "GENUINE_PERCENT": result.GenuinePercent, "SEGMENTS": result.Segments}
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
	t.Run("case4", func(t *testing.T) {
		p := DefaultParams()
		p.Period = 20
		p.Segments = 5
		p.Smoothing = 1600.0
		indicator, err := NewBartelsPeriodicityScore(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"SCORE": expectedP20_M5_L1600_SCORE, "GENUINE_PERCENT": expectedP20_M5_L1600_GENUINE_PERCENT, "SEGMENTS": expectedP20_M5_L1600_SEGMENTS}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"SCORE": result.Score, "GENUINE_PERCENT": result.GenuinePercent, "SEGMENTS": result.Segments}
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
	t.Run("case5", func(t *testing.T) {
		p := DefaultParams()
		p.Period = 40
		p.Segments = 5
		p.Smoothing = 0.0
		indicator, err := NewBartelsPeriodicityScore(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"SCORE": expectedP40_M5_L0_SCORE, "GENUINE_PERCENT": expectedP40_M5_L0_GENUINE_PERCENT, "SEGMENTS": expectedP40_M5_L0_SEGMENTS}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"SCORE": result.Score, "GENUINE_PERCENT": result.GenuinePercent, "SEGMENTS": result.Segments}
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
	t.Run("case6", func(t *testing.T) {
		p := DefaultParams()
		p.Period = 42
		p.Segments = 6
		p.Smoothing = 0.0
		indicator, err := NewBartelsPeriodicityScore(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"SCORE": expectedP42_M6_L0_SCORE, "GENUINE_PERCENT": expectedP42_M6_L0_GENUINE_PERCENT, "SEGMENTS": expectedP42_M6_L0_SEGMENTS}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"SCORE": result.Score, "GENUINE_PERCENT": result.GenuinePercent, "SEGMENTS": result.Segments}
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
	t.Run("case7", func(t *testing.T) {
		p := DefaultParams()
		p.Period = 60
		p.Segments = 5
		p.Smoothing = 0.0
		indicator, err := NewBartelsPeriodicityScore(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"SCORE": expectedP60_M5_L0_SCORE, "GENUINE_PERCENT": expectedP60_M5_L0_GENUINE_PERCENT, "SEGMENTS": expectedP60_M5_L0_SEGMENTS}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"SCORE": result.Score, "GENUINE_PERCENT": result.GenuinePercent, "SEGMENTS": result.Segments}
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
