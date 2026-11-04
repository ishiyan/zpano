// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package compositecycle

import (
	"math"
	"testing"
)

func TestPreparedReferenceCases(t *testing.T) {
	maxError, maxScaledError := 0.0, 0.0
	defer func() { t.Logf("precision max absolute %.17g scaled %.17g", maxError, maxScaledError) }()
	t.Run("case0", func(t *testing.T) {
		p := DefaultParams()
		p.Window = 100
		p.Periods = []float64{20.0, 40.0}
		p.Smoothing = 1600.0
		indicator, err := NewCompositeCycle(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedW100_P20_40_L1600_VALUE, "SLOPE": expectedW100_P20_40_L1600_SLOPE, "NEXT_VALUE": expectedW100_P20_40_L1600_NEXT_VALUE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "SLOPE": result.Slope, "NEXT_VALUE": result.NextValue}
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
		p.Window = 100
		p.Periods = []float64{20.0}
		p.Smoothing = 0.0
		indicator, err := NewCompositeCycle(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedW100_P20_L0_VALUE, "SLOPE": expectedW100_P20_L0_SLOPE, "NEXT_VALUE": expectedW100_P20_L0_NEXT_VALUE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "SLOPE": result.Slope, "NEXT_VALUE": result.NextValue}
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
		p.Window = 100
		p.Periods = []float64{20.0, 40.0}
		p.Smoothing = 0.0
		indicator, err := NewCompositeCycle(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedW100_P20_40_L0_VALUE, "SLOPE": expectedW100_P20_40_L0_SLOPE, "NEXT_VALUE": expectedW100_P20_40_L0_NEXT_VALUE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "SLOPE": result.Slope, "NEXT_VALUE": result.NextValue}
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
		p.Window = 120
		p.Periods = []float64{10.0, 20.0, 40.0}
		p.Smoothing = 1600.0
		indicator, err := NewCompositeCycle(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedW120_P10_20_40_L1600_VALUE, "SLOPE": expectedW120_P10_20_40_L1600_SLOPE, "NEXT_VALUE": expectedW120_P10_20_40_L1600_NEXT_VALUE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "SLOPE": result.Slope, "NEXT_VALUE": result.NextValue}
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
		p.Window = 100
		p.Periods = []float64{17.5, 31.0}
		p.Smoothing = 1600.0
		indicator, err := NewCompositeCycle(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedW100_P17_5_31_L1600_VALUE, "SLOPE": expectedW100_P17_5_31_L1600_SLOPE, "NEXT_VALUE": expectedW100_P17_5_31_L1600_NEXT_VALUE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "SLOPE": result.Slope, "NEXT_VALUE": result.NextValue}
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
		p.Window = 200
		p.Periods = []float64{20.0, 40.0}
		p.Smoothing = 1600.0
		indicator, err := NewCompositeCycle(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedW200_P20_40_L1600_VALUE, "SLOPE": expectedW200_P20_40_L1600_SLOPE, "NEXT_VALUE": expectedW200_P20_40_L1600_NEXT_VALUE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "SLOPE": result.Slope, "NEXT_VALUE": result.NextValue}
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
		p.Window = 100
		p.Periods = []float64{20.0, 40.0}
		p.Smoothing = 160.0
		indicator, err := NewCompositeCycle(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedW100_P20_40_L160_VALUE, "SLOPE": expectedW100_P20_40_L160_SLOPE, "NEXT_VALUE": expectedW100_P20_40_L160_NEXT_VALUE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "SLOPE": result.Slope, "NEXT_VALUE": result.NextValue}
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
		p.Window = 253
		p.Periods = []float64{20.0, 40.0}
		p.Smoothing = 1600.0
		indicator, err := NewCompositeCycle(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedW253_P20_40_L1600_VALUE, "SLOPE": expectedW253_P20_40_L1600_SLOPE, "NEXT_VALUE": expectedW253_P20_40_L1600_NEXT_VALUE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "SLOPE": result.Slope, "NEXT_VALUE": result.NextValue}
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
