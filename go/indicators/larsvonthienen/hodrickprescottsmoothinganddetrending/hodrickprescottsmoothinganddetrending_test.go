// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package hodrickprescottsmoothinganddetrending

import (
	"math"
	"testing"
)

func TestPreparedReferenceCases(t *testing.T) {
	maxError, maxScaledError := 0.0, 0.0
	defer func() { t.Logf("precision max absolute %.17g scaled %.17g", maxError, maxScaledError) }()
	t.Run("case0", func(t *testing.T) {
		p := DefaultParams()
		p.Window = 50
		p.Smoothing = 1600.0
		indicator, err := NewHodrickPrescottSmoothingAndDetrending(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"TREND": expectedW50_L1600_TREND, "CYCLE": expectedW50_L1600_CYCLE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"TREND": result.Trend, "CYCLE": result.Cycle}
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
				} else if math.IsNaN(value) || math.Abs(value-values[index]) > max(1e-13, math.Abs(values[index])*5e-15) {
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
		p.Window = 6
		p.Smoothing = 1600.0
		indicator, err := NewHodrickPrescottSmoothingAndDetrending(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"TREND": expectedW6_L1600_TREND, "CYCLE": expectedW6_L1600_CYCLE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"TREND": result.Trend, "CYCLE": result.Cycle}
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
				} else if math.IsNaN(value) || math.Abs(value-values[index]) > max(1e-13, math.Abs(values[index])*5e-15) {
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
		p.Window = 20
		p.Smoothing = 0.0
		indicator, err := NewHodrickPrescottSmoothingAndDetrending(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"TREND": expectedW20_L0_TREND, "CYCLE": expectedW20_L0_CYCLE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"TREND": result.Trend, "CYCLE": result.Cycle}
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
				} else if math.IsNaN(value) || math.Abs(value-values[index]) > max(1e-13, math.Abs(values[index])*5e-15) {
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
		p.Window = 20
		p.Smoothing = 1.0
		indicator, err := NewHodrickPrescottSmoothingAndDetrending(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"TREND": expectedW20_L1_TREND, "CYCLE": expectedW20_L1_CYCLE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"TREND": result.Trend, "CYCLE": result.Cycle}
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
				} else if math.IsNaN(value) || math.Abs(value-values[index]) > max(1e-13, math.Abs(values[index])*5e-15) {
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
		p.Window = 50
		p.Smoothing = 160.0
		indicator, err := NewHodrickPrescottSmoothingAndDetrending(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"TREND": expectedW50_L160_TREND, "CYCLE": expectedW50_L160_CYCLE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"TREND": result.Trend, "CYCLE": result.Cycle}
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
				} else if math.IsNaN(value) || math.Abs(value-values[index]) > max(1e-13, math.Abs(values[index])*5e-15) {
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
		p.Window = 100
		p.Smoothing = 1600.0
		indicator, err := NewHodrickPrescottSmoothingAndDetrending(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"TREND": expectedW100_L1600_TREND, "CYCLE": expectedW100_L1600_CYCLE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"TREND": result.Trend, "CYCLE": result.Cycle}
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
				} else if math.IsNaN(value) || math.Abs(value-values[index]) > max(1e-13, math.Abs(values[index])*5e-15) {
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
		p.Window = 100
		p.Smoothing = 100000.0
		indicator, err := NewHodrickPrescottSmoothingAndDetrending(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"TREND": expectedW100_L100000_TREND, "CYCLE": expectedW100_L100000_CYCLE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"TREND": result.Trend, "CYCLE": result.Cycle}
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
				} else if math.IsNaN(value) || math.Abs(value-values[index]) > max(1e-13, math.Abs(values[index])*5e-15) {
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
		p.Window = 252
		p.Smoothing = 1600.0
		indicator, err := NewHodrickPrescottSmoothingAndDetrending(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"TREND": expectedW252_L1600_TREND, "CYCLE": expectedW252_L1600_CYCLE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"TREND": result.Trend, "CYCLE": result.Cycle}
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
				} else if math.IsNaN(value) || math.Abs(value-values[index]) > max(1e-13, math.Abs(values[index])*5e-15) {
					t.Fatalf("bar %d %s expected %.17g got %.17g", index, field, values[index], value)
				}
			}
		}
		indicator.Reset()
		if indicator.IsPrimed() {
			t.Fatal("reset must clear priming")
		}
	})
	t.Run("case8", func(t *testing.T) {
		p := DefaultParams()
		p.Window = 253
		p.Smoothing = 1600.0
		indicator, err := NewHodrickPrescottSmoothingAndDetrending(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"TREND": expectedW253_L1600_TREND, "CYCLE": expectedW253_L1600_CYCLE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"TREND": result.Trend, "CYCLE": result.Cycle}
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
				} else if math.IsNaN(value) || math.Abs(value-values[index]) > max(1e-13, math.Abs(values[index])*5e-15) {
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
