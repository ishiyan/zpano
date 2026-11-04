// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package cyclestrength

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
		p.Period = 20.0
		p.Smoothing = 1600.0
		indicator, err := NewCycleStrength(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"STRENGTH": expectedW100_P20_L1600_STRENGTH, "AMPLITUDE": expectedW100_P20_L1600_AMPLITUDE, "PHASE": expectedW100_P20_L1600_PHASE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"STRENGTH": result.Strength, "AMPLITUDE": result.Amplitude, "PHASE": result.Phase}
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
		p.Period = 20.0
		p.Smoothing = 0.0
		indicator, err := NewCycleStrength(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"STRENGTH": expectedW100_P20_L0_STRENGTH, "AMPLITUDE": expectedW100_P20_L0_AMPLITUDE, "PHASE": expectedW100_P20_L0_PHASE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"STRENGTH": result.Strength, "AMPLITUDE": result.Amplitude, "PHASE": result.Phase}
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
		p.Window = 60
		p.Period = 10.0
		p.Smoothing = 1600.0
		indicator, err := NewCycleStrength(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"STRENGTH": expectedW60_P10_L1600_STRENGTH, "AMPLITUDE": expectedW60_P10_L1600_AMPLITUDE, "PHASE": expectedW60_P10_L1600_PHASE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"STRENGTH": result.Strength, "AMPLITUDE": result.Amplitude, "PHASE": result.Phase}
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
		p.Window = 200
		p.Period = 40.0
		p.Smoothing = 1600.0
		indicator, err := NewCycleStrength(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"STRENGTH": expectedW200_P40_L1600_STRENGTH, "AMPLITUDE": expectedW200_P40_L1600_AMPLITUDE, "PHASE": expectedW200_P40_L1600_PHASE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"STRENGTH": result.Strength, "AMPLITUDE": result.Amplitude, "PHASE": result.Phase}
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
		p.Period = 20.5
		p.Smoothing = 1600.0
		indicator, err := NewCycleStrength(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"STRENGTH": expectedW100_P20_5_L1600_STRENGTH, "AMPLITUDE": expectedW100_P20_5_L1600_AMPLITUDE, "PHASE": expectedW100_P20_5_L1600_PHASE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"STRENGTH": result.Strength, "AMPLITUDE": result.Amplitude, "PHASE": result.Phase}
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
		p.Window = 100
		p.Period = 20.0
		p.Smoothing = 16000.0
		indicator, err := NewCycleStrength(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"STRENGTH": expectedW100_P20_L16000_STRENGTH, "AMPLITUDE": expectedW100_P20_L16000_AMPLITUDE, "PHASE": expectedW100_P20_L16000_PHASE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"STRENGTH": result.Strength, "AMPLITUDE": result.Amplitude, "PHASE": result.Phase}
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
		p.Window = 20
		p.Period = 40.0
		p.Smoothing = 1600.0
		indicator, err := NewCycleStrength(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"STRENGTH": expectedW20_P40_L1600_STRENGTH, "AMPLITUDE": expectedW20_P40_L1600_AMPLITUDE, "PHASE": expectedW20_P40_L1600_PHASE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"STRENGTH": result.Strength, "AMPLITUDE": result.Amplitude, "PHASE": result.Phase}
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
		p.Window = 252
		p.Period = 20.0
		p.Smoothing = 1600.0
		indicator, err := NewCycleStrength(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"STRENGTH": expectedW252_P20_L1600_STRENGTH, "AMPLITUDE": expectedW252_P20_L1600_AMPLITUDE, "PHASE": expectedW252_P20_L1600_PHASE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"STRENGTH": result.Strength, "AMPLITUDE": result.Amplitude, "PHASE": result.Phase}
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
