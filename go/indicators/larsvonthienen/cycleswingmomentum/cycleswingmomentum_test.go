// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package cycleswingmomentum

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
		p.FastSmoothing = 160.0
		p.SlowSmoothing = 1600.0
		p.BandMemory = 34
		p.Leveling = 10.0
		indicator, err := NewCycleSwingMomentum(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedW50_F160_S1600_M34_Q10_VALUE, "LOWER": expectedW50_F160_S1600_M34_Q10_LOWER, "UPPER": expectedW50_F160_S1600_M34_Q10_UPPER, "SCORE": expectedW50_F160_S1600_M34_Q10_SCORE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "LOWER": result.Lower, "UPPER": result.Upper, "SCORE": result.Score}
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
		p.FastSmoothing = 16.0
		p.SlowSmoothing = 160.0
		p.BandMemory = 5
		p.Leveling = 10.0
		indicator, err := NewCycleSwingMomentum(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedW6_F16_S160_M5_Q10_VALUE, "LOWER": expectedW6_F16_S160_M5_Q10_LOWER, "UPPER": expectedW6_F16_S160_M5_Q10_UPPER, "SCORE": expectedW6_F16_S160_M5_Q10_SCORE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "LOWER": result.Lower, "UPPER": result.Upper, "SCORE": result.Score}
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
		p.FastSmoothing = 160.0
		p.SlowSmoothing = 1600.0
		p.BandMemory = 20
		p.Leveling = 10.0
		indicator, err := NewCycleSwingMomentum(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedW20_F160_S1600_M20_Q10_VALUE, "LOWER": expectedW20_F160_S1600_M20_Q10_LOWER, "UPPER": expectedW20_F160_S1600_M20_Q10_UPPER, "SCORE": expectedW20_F160_S1600_M20_Q10_SCORE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "LOWER": result.Lower, "UPPER": result.Upper, "SCORE": result.Score}
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
		p.Window = 100
		p.FastSmoothing = 160.0
		p.SlowSmoothing = 1600.0
		p.BandMemory = 34
		p.Leveling = 10.0
		indicator, err := NewCycleSwingMomentum(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedW100_F160_S1600_M34_Q10_VALUE, "LOWER": expectedW100_F160_S1600_M34_Q10_LOWER, "UPPER": expectedW100_F160_S1600_M34_Q10_UPPER, "SCORE": expectedW100_F160_S1600_M34_Q10_SCORE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "LOWER": result.Lower, "UPPER": result.Upper, "SCORE": result.Score}
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
		p.FastSmoothing = 0.0
		p.SlowSmoothing = 1600.0
		p.BandMemory = 34
		p.Leveling = 10.0
		indicator, err := NewCycleSwingMomentum(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedW50_F0_S1600_M34_Q10_VALUE, "LOWER": expectedW50_F0_S1600_M34_Q10_LOWER, "UPPER": expectedW50_F0_S1600_M34_Q10_UPPER, "SCORE": expectedW50_F0_S1600_M34_Q10_SCORE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "LOWER": result.Lower, "UPPER": result.Upper, "SCORE": result.Score}
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
		p.Window = 50
		p.FastSmoothing = 1600.0
		p.SlowSmoothing = 16000.0
		p.BandMemory = 34
		p.Leveling = 10.0
		indicator, err := NewCycleSwingMomentum(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedW50_F1600_S16000_M34_Q10_VALUE, "LOWER": expectedW50_F1600_S16000_M34_Q10_LOWER, "UPPER": expectedW50_F1600_S16000_M34_Q10_UPPER, "SCORE": expectedW50_F1600_S16000_M34_Q10_SCORE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "LOWER": result.Lower, "UPPER": result.Upper, "SCORE": result.Score}
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
		p.Window = 50
		p.FastSmoothing = 160.0
		p.SlowSmoothing = 1600.0
		p.BandMemory = 34
		p.Leveling = 5.0
		indicator, err := NewCycleSwingMomentum(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedW50_F160_S1600_M34_Q5_VALUE, "LOWER": expectedW50_F160_S1600_M34_Q5_LOWER, "UPPER": expectedW50_F160_S1600_M34_Q5_UPPER, "SCORE": expectedW50_F160_S1600_M34_Q5_SCORE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "LOWER": result.Lower, "UPPER": result.Upper, "SCORE": result.Score}
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
		p.FastSmoothing = 160.0
		p.SlowSmoothing = 1600.0
		p.BandMemory = 34
		p.Leveling = 10.0
		indicator, err := NewCycleSwingMomentum(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedW252_F160_S1600_M34_Q10_VALUE, "LOWER": expectedW252_F160_S1600_M34_Q10_LOWER, "UPPER": expectedW252_F160_S1600_M34_Q10_UPPER, "SCORE": expectedW252_F160_S1600_M34_Q10_SCORE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "LOWER": result.Lower, "UPPER": result.Upper, "SCORE": result.Score}
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
