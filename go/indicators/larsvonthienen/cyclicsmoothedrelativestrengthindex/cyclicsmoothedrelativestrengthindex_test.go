// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package cyclicsmoothedrelativestrengthindex

import (
	"math"
	"testing"
)

func TestPreparedReferenceCases(t *testing.T) {
	maxError, maxScaledError := 0.0, 0.0
	defer func() { t.Logf("precision max absolute %.17g scaled %.17g", maxError, maxScaledError) }()
	t.Run("case0", func(t *testing.T) {
		p := DefaultParams()
		p.DominantCycle = 20
		p.Vibration = 10
		p.Leveling = 10.0
		indicator, err := NewCyclicSmoothedRelativeStrengthIndex(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedD20_V10_Q10_VALUE, "LOWER": expectedD20_V10_Q10_LOWER, "UPPER": expectedD20_V10_Q10_UPPER}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "LOWER": result.Lower, "UPPER": result.Upper}
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
		p.DominantCycle = 10
		p.Vibration = 10
		p.Leveling = 10.0
		indicator, err := NewCyclicSmoothedRelativeStrengthIndex(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedD10_V10_Q10_VALUE, "LOWER": expectedD10_V10_Q10_LOWER, "UPPER": expectedD10_V10_Q10_UPPER}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "LOWER": result.Lower, "UPPER": result.Upper}
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
		p.DominantCycle = 21
		p.Vibration = 10
		p.Leveling = 10.0
		indicator, err := NewCyclicSmoothedRelativeStrengthIndex(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedD21_V10_Q10_VALUE, "LOWER": expectedD21_V10_Q10_LOWER, "UPPER": expectedD21_V10_Q10_UPPER}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "LOWER": result.Lower, "UPPER": result.Upper}
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
		p.DominantCycle = 40
		p.Vibration = 10
		p.Leveling = 10.0
		indicator, err := NewCyclicSmoothedRelativeStrengthIndex(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedD40_V10_Q10_VALUE, "LOWER": expectedD40_V10_Q10_LOWER, "UPPER": expectedD40_V10_Q10_UPPER}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "LOWER": result.Lower, "UPPER": result.Upper}
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
		p.DominantCycle = 100
		p.Vibration = 10
		p.Leveling = 10.0
		indicator, err := NewCyclicSmoothedRelativeStrengthIndex(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedD100_V10_Q10_VALUE, "LOWER": expectedD100_V10_Q10_LOWER, "UPPER": expectedD100_V10_Q10_UPPER}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "LOWER": result.Lower, "UPPER": result.Upper}
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
		p.DominantCycle = 20
		p.Vibration = 3
		p.Leveling = 10.0
		indicator, err := NewCyclicSmoothedRelativeStrengthIndex(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedD20_V3_Q10_VALUE, "LOWER": expectedD20_V3_Q10_LOWER, "UPPER": expectedD20_V3_Q10_UPPER}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "LOWER": result.Lower, "UPPER": result.Upper}
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
		p.DominantCycle = 20
		p.Vibration = 6
		p.Leveling = 10.0
		indicator, err := NewCyclicSmoothedRelativeStrengthIndex(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedD20_V6_Q10_VALUE, "LOWER": expectedD20_V6_Q10_LOWER, "UPPER": expectedD20_V6_Q10_UPPER}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "LOWER": result.Lower, "UPPER": result.Upper}
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
		p.DominantCycle = 20
		p.Vibration = 10
		p.Leveling = 5.0
		indicator, err := NewCyclicSmoothedRelativeStrengthIndex(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedD20_V10_Q5_VALUE, "LOWER": expectedD20_V10_Q5_LOWER, "UPPER": expectedD20_V10_Q5_UPPER}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "LOWER": result.Lower, "UPPER": result.Upper}
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
		p.DominantCycle = 20
		p.Vibration = 10
		p.Leveling = 20.0
		indicator, err := NewCyclicSmoothedRelativeStrengthIndex(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedD20_V10_Q20_VALUE, "LOWER": expectedD20_V10_Q20_LOWER, "UPPER": expectedD20_V10_Q20_UPPER}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "LOWER": result.Lower, "UPPER": result.Upper}
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
