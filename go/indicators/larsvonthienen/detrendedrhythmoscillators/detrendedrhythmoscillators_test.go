// SPDX-License-Identifier: MPL-2.0
// Source: Detrended Rhythm Oscillator, © WhenToTrade / Lars von Thienen.
// Numerical source: Lars von Thienen / WhenToTrade (2017), MPL-2.0.
// Research adaptations are documented in description.md.
package detrendedrhythmoscillators

import (
	"math"
	"testing"
)

func TestPreparedReferenceCases(t *testing.T) {
	maxError, maxScaledError := 0.0, 0.0
	defer func() { t.Logf("precision max absolute %.17g scaled %.17g", maxError, maxScaledError) }()
	t.Run("case0", func(t *testing.T) {
		p := DefaultParams()
		p.FastLength = 3
		p.SlowLength = 50
		p.ZigPeriod = 30
		p.CycleMemory = 15
		p.Detrend = true
		indicator, err := NewDetrendedRhythmOscillators(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedF3_S50_Z30_M15_D1_VALUE, "DIRECTION": expectedF3_S50_Z30_M15_D1_DIRECTION, "CYCLE_LENGTH": expectedF3_S50_Z30_M15_D1_CYCLE_LENGTH, "PIVOT_VALUE": expectedF3_S50_Z30_M15_D1_PIVOT_VALUE, "PIVOT_OFFSET": expectedF3_S50_Z30_M15_D1_PIVOT_OFFSET}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "DIRECTION": result.Direction, "CYCLE_LENGTH": result.CycleLength, "PIVOT_VALUE": result.PivotValue, "PIVOT_OFFSET": result.PivotOffset}
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
		p.FastLength = 3
		p.SlowLength = 50
		p.ZigPeriod = 5
		p.CycleMemory = 15
		p.Detrend = true
		indicator, err := NewDetrendedRhythmOscillators(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedF3_S50_Z5_M15_D1_VALUE, "DIRECTION": expectedF3_S50_Z5_M15_D1_DIRECTION, "CYCLE_LENGTH": expectedF3_S50_Z5_M15_D1_CYCLE_LENGTH, "PIVOT_VALUE": expectedF3_S50_Z5_M15_D1_PIVOT_VALUE, "PIVOT_OFFSET": expectedF3_S50_Z5_M15_D1_PIVOT_OFFSET}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "DIRECTION": result.Direction, "CYCLE_LENGTH": result.CycleLength, "PIVOT_VALUE": result.PivotValue, "PIVOT_OFFSET": result.PivotOffset}
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
		p.FastLength = 3
		p.SlowLength = 50
		p.ZigPeriod = 10
		p.CycleMemory = 15
		p.Detrend = true
		indicator, err := NewDetrendedRhythmOscillators(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedF3_S50_Z10_M15_D1_VALUE, "DIRECTION": expectedF3_S50_Z10_M15_D1_DIRECTION, "CYCLE_LENGTH": expectedF3_S50_Z10_M15_D1_CYCLE_LENGTH, "PIVOT_VALUE": expectedF3_S50_Z10_M15_D1_PIVOT_VALUE, "PIVOT_OFFSET": expectedF3_S50_Z10_M15_D1_PIVOT_OFFSET}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "DIRECTION": result.Direction, "CYCLE_LENGTH": result.CycleLength, "PIVOT_VALUE": result.PivotValue, "PIVOT_OFFSET": result.PivotOffset}
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
		p.FastLength = 2
		p.SlowLength = 20
		p.ZigPeriod = 10
		p.CycleMemory = 15
		p.Detrend = true
		indicator, err := NewDetrendedRhythmOscillators(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedF2_S20_Z10_M15_D1_VALUE, "DIRECTION": expectedF2_S20_Z10_M15_D1_DIRECTION, "CYCLE_LENGTH": expectedF2_S20_Z10_M15_D1_CYCLE_LENGTH, "PIVOT_VALUE": expectedF2_S20_Z10_M15_D1_PIVOT_VALUE, "PIVOT_OFFSET": expectedF2_S20_Z10_M15_D1_PIVOT_OFFSET}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "DIRECTION": result.Direction, "CYCLE_LENGTH": result.CycleLength, "PIVOT_VALUE": result.PivotValue, "PIVOT_OFFSET": result.PivotOffset}
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
		p.FastLength = 5
		p.SlowLength = 100
		p.ZigPeriod = 30
		p.CycleMemory = 15
		p.Detrend = true
		indicator, err := NewDetrendedRhythmOscillators(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedF5_S100_Z30_M15_D1_VALUE, "DIRECTION": expectedF5_S100_Z30_M15_D1_DIRECTION, "CYCLE_LENGTH": expectedF5_S100_Z30_M15_D1_CYCLE_LENGTH, "PIVOT_VALUE": expectedF5_S100_Z30_M15_D1_PIVOT_VALUE, "PIVOT_OFFSET": expectedF5_S100_Z30_M15_D1_PIVOT_OFFSET}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "DIRECTION": result.Direction, "CYCLE_LENGTH": result.CycleLength, "PIVOT_VALUE": result.PivotValue, "PIVOT_OFFSET": result.PivotOffset}
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
		p.FastLength = 3
		p.SlowLength = 50
		p.ZigPeriod = 5
		p.CycleMemory = 3
		p.Detrend = true
		indicator, err := NewDetrendedRhythmOscillators(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedF3_S50_Z5_M3_D1_VALUE, "DIRECTION": expectedF3_S50_Z5_M3_D1_DIRECTION, "CYCLE_LENGTH": expectedF3_S50_Z5_M3_D1_CYCLE_LENGTH, "PIVOT_VALUE": expectedF3_S50_Z5_M3_D1_PIVOT_VALUE, "PIVOT_OFFSET": expectedF3_S50_Z5_M3_D1_PIVOT_OFFSET}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "DIRECTION": result.Direction, "CYCLE_LENGTH": result.CycleLength, "PIVOT_VALUE": result.PivotValue, "PIVOT_OFFSET": result.PivotOffset}
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
		p.FastLength = 3
		p.SlowLength = 50
		p.ZigPeriod = 10
		p.CycleMemory = 15
		p.Detrend = false
		indicator, err := NewDetrendedRhythmOscillators(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedF3_S50_Z10_M15_D0_VALUE, "DIRECTION": expectedF3_S50_Z10_M15_D0_DIRECTION, "CYCLE_LENGTH": expectedF3_S50_Z10_M15_D0_CYCLE_LENGTH, "PIVOT_VALUE": expectedF3_S50_Z10_M15_D0_PIVOT_VALUE, "PIVOT_OFFSET": expectedF3_S50_Z10_M15_D0_PIVOT_OFFSET}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "DIRECTION": result.Direction, "CYCLE_LENGTH": result.CycleLength, "PIVOT_VALUE": result.PivotValue, "PIVOT_OFFSET": result.PivotOffset}
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
		p.FastLength = 3
		p.SlowLength = 300
		p.ZigPeriod = 30
		p.CycleMemory = 15
		p.Detrend = true
		indicator, err := NewDetrendedRhythmOscillators(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedF3_S300_Z30_M15_D1_VALUE, "DIRECTION": expectedF3_S300_Z30_M15_D1_DIRECTION, "CYCLE_LENGTH": expectedF3_S300_Z30_M15_D1_CYCLE_LENGTH, "PIVOT_VALUE": expectedF3_S300_Z30_M15_D1_PIVOT_VALUE, "PIVOT_OFFSET": expectedF3_S300_Z30_M15_D1_PIVOT_OFFSET}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "DIRECTION": result.Direction, "CYCLE_LENGTH": result.CycleLength, "PIVOT_VALUE": result.PivotValue, "PIVOT_OFFSET": result.PivotOffset}
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
