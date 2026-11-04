// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package cyclepolarity

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
		p.ShortPeriod = 20.0
		p.LongPeriod = 40.0
		p.Smoothing = 1600.0
		p.Tolerance = 0.0
		indicator, err := NewCyclePolarity(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"POLARITY": expectedW100_S20_L40_HP1600_T0_POLARITY, "SHORT_VALUE": expectedW100_S20_L40_HP1600_T0_SHORT_VALUE, "LONG_VALUE": expectedW100_S20_L40_HP1600_T0_LONG_VALUE, "SHORT_SLOPE": expectedW100_S20_L40_HP1600_T0_SHORT_SLOPE, "LONG_SLOPE": expectedW100_S20_L40_HP1600_T0_LONG_SLOPE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"POLARITY": result.Polarity, "SHORT_VALUE": result.ShortValue, "LONG_VALUE": result.LongValue, "SHORT_SLOPE": result.ShortSlope, "LONG_SLOPE": result.LongSlope}
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
		p.ShortPeriod = 20.0
		p.LongPeriod = 40.0
		p.Smoothing = 0.0
		p.Tolerance = 0.0
		indicator, err := NewCyclePolarity(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"POLARITY": expectedW100_S20_L40_HP0_T0_POLARITY, "SHORT_VALUE": expectedW100_S20_L40_HP0_T0_SHORT_VALUE, "LONG_VALUE": expectedW100_S20_L40_HP0_T0_LONG_VALUE, "SHORT_SLOPE": expectedW100_S20_L40_HP0_T0_SHORT_SLOPE, "LONG_SLOPE": expectedW100_S20_L40_HP0_T0_LONG_SLOPE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"POLARITY": result.Polarity, "SHORT_VALUE": result.ShortValue, "LONG_VALUE": result.LongValue, "SHORT_SLOPE": result.ShortSlope, "LONG_SLOPE": result.LongSlope}
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
		p.ShortPeriod = 10.0
		p.LongPeriod = 20.0
		p.Smoothing = 1600.0
		p.Tolerance = 0.0
		indicator, err := NewCyclePolarity(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"POLARITY": expectedW60_S10_L20_HP1600_T0_POLARITY, "SHORT_VALUE": expectedW60_S10_L20_HP1600_T0_SHORT_VALUE, "LONG_VALUE": expectedW60_S10_L20_HP1600_T0_LONG_VALUE, "SHORT_SLOPE": expectedW60_S10_L20_HP1600_T0_SHORT_SLOPE, "LONG_SLOPE": expectedW60_S10_L20_HP1600_T0_LONG_SLOPE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"POLARITY": result.Polarity, "SHORT_VALUE": result.ShortValue, "LONG_VALUE": result.LongValue, "SHORT_SLOPE": result.ShortSlope, "LONG_SLOPE": result.LongSlope}
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
		p.ShortPeriod = 40.0
		p.LongPeriod = 80.0
		p.Smoothing = 1600.0
		p.Tolerance = 0.0
		indicator, err := NewCyclePolarity(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"POLARITY": expectedW200_S40_L80_HP1600_T0_POLARITY, "SHORT_VALUE": expectedW200_S40_L80_HP1600_T0_SHORT_VALUE, "LONG_VALUE": expectedW200_S40_L80_HP1600_T0_LONG_VALUE, "SHORT_SLOPE": expectedW200_S40_L80_HP1600_T0_SHORT_SLOPE, "LONG_SLOPE": expectedW200_S40_L80_HP1600_T0_LONG_SLOPE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"POLARITY": result.Polarity, "SHORT_VALUE": result.ShortValue, "LONG_VALUE": result.LongValue, "SHORT_SLOPE": result.ShortSlope, "LONG_SLOPE": result.LongSlope}
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
		p.ShortPeriod = 17.5
		p.LongPeriod = 31.0
		p.Smoothing = 1600.0
		p.Tolerance = 0.0
		indicator, err := NewCyclePolarity(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"POLARITY": expectedW100_S17_5_L31_HP1600_T0_POLARITY, "SHORT_VALUE": expectedW100_S17_5_L31_HP1600_T0_SHORT_VALUE, "LONG_VALUE": expectedW100_S17_5_L31_HP1600_T0_LONG_VALUE, "SHORT_SLOPE": expectedW100_S17_5_L31_HP1600_T0_SHORT_SLOPE, "LONG_SLOPE": expectedW100_S17_5_L31_HP1600_T0_LONG_SLOPE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"POLARITY": result.Polarity, "SHORT_VALUE": result.ShortValue, "LONG_VALUE": result.LongValue, "SHORT_SLOPE": result.ShortSlope, "LONG_SLOPE": result.LongSlope}
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
		p.ShortPeriod = 20.0
		p.LongPeriod = 40.0
		p.Smoothing = 1600.0
		p.Tolerance = 0.1
		indicator, err := NewCyclePolarity(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"POLARITY": expectedW100_S20_L40_HP1600_T0_1_POLARITY, "SHORT_VALUE": expectedW100_S20_L40_HP1600_T0_1_SHORT_VALUE, "LONG_VALUE": expectedW100_S20_L40_HP1600_T0_1_LONG_VALUE, "SHORT_SLOPE": expectedW100_S20_L40_HP1600_T0_1_SHORT_SLOPE, "LONG_SLOPE": expectedW100_S20_L40_HP1600_T0_1_LONG_SLOPE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"POLARITY": result.Polarity, "SHORT_VALUE": result.ShortValue, "LONG_VALUE": result.LongValue, "SHORT_SLOPE": result.ShortSlope, "LONG_SLOPE": result.LongSlope}
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
		p.ShortPeriod = 20.0
		p.LongPeriod = 40.0
		p.Smoothing = 1600.0
		p.Tolerance = 1000.0
		indicator, err := NewCyclePolarity(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"POLARITY": expectedW100_S20_L40_HP1600_T1000_POLARITY, "SHORT_VALUE": expectedW100_S20_L40_HP1600_T1000_SHORT_VALUE, "LONG_VALUE": expectedW100_S20_L40_HP1600_T1000_LONG_VALUE, "SHORT_SLOPE": expectedW100_S20_L40_HP1600_T1000_SHORT_SLOPE, "LONG_SLOPE": expectedW100_S20_L40_HP1600_T1000_LONG_SLOPE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"POLARITY": result.Polarity, "SHORT_VALUE": result.ShortValue, "LONG_VALUE": result.LongValue, "SHORT_SLOPE": result.ShortSlope, "LONG_SLOPE": result.LongSlope}
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
		p.ShortPeriod = 20.0
		p.LongPeriod = 40.0
		p.Smoothing = 1600.0
		p.Tolerance = 0.0
		indicator, err := NewCyclePolarity(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"POLARITY": expectedW253_S20_L40_HP1600_T0_POLARITY, "SHORT_VALUE": expectedW253_S20_L40_HP1600_T0_SHORT_VALUE, "LONG_VALUE": expectedW253_S20_L40_HP1600_T0_LONG_VALUE, "SHORT_SLOPE": expectedW253_S20_L40_HP1600_T0_SHORT_SLOPE, "LONG_SLOPE": expectedW253_S20_L40_HP1600_T0_LONG_SLOPE}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"POLARITY": result.Polarity, "SHORT_VALUE": result.ShortValue, "LONG_VALUE": result.LongValue, "SHORT_SLOPE": result.ShortSlope, "LONG_SLOPE": result.LongSlope}
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
