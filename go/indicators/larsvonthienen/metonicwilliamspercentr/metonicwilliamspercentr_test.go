// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package metonicwilliamspercentr

import (
	"math"
	"testing"
	"time"
)

func TestPreparedReferenceCases(t *testing.T) {
	maxError, maxScaledError := 0.0, 0.0
	defer func() { t.Logf("precision max absolute %.17g scaled %.17g", maxError, maxScaledError) }()
	t.Run("case0", func(t *testing.T) {
		p := DefaultParams()
		p.Length = 13
		p.Delay1 = 6939
		p.Delay2 = 4016
		p.Delay3 = 2922
		indicator, err := NewMetonicWilliamsPercentR(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedL13_D16939_D24016_D32922_VALUE, "LINEAR_PERCENT_R": expectedL13_D16939_D24016_D32922_LINEAR_PERCENT_R}
		for index, price := range testInput {
			result := indicator.Update(price, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, index), testHigh[index], testLow[index])
			actual := map[string]float64{"VALUE": result.Value, "LINEAR_PERCENT_R": result.LinearPercentR}
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
		p.Length = 13
		p.Delay1 = 19
		p.Delay2 = 11
		p.Delay3 = 8
		indicator, err := NewMetonicWilliamsPercentR(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedL13_D119_D211_D38_VALUE, "LINEAR_PERCENT_R": expectedL13_D119_D211_D38_LINEAR_PERCENT_R}
		for index, price := range testInput {
			result := indicator.Update(price, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, index), testHigh[index], testLow[index])
			actual := map[string]float64{"VALUE": result.Value, "LINEAR_PERCENT_R": result.LinearPercentR}
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
		p.Length = 1
		p.Delay1 = 19
		p.Delay2 = 11
		p.Delay3 = 8
		indicator, err := NewMetonicWilliamsPercentR(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedL1_D119_D211_D38_VALUE, "LINEAR_PERCENT_R": expectedL1_D119_D211_D38_LINEAR_PERCENT_R}
		for index, price := range testInput {
			result := indicator.Update(price, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, index), testHigh[index], testLow[index])
			actual := map[string]float64{"VALUE": result.Value, "LINEAR_PERCENT_R": result.LinearPercentR}
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
		p.Length = 14
		p.Delay1 = 19
		p.Delay2 = 11
		p.Delay3 = 8
		indicator, err := NewMetonicWilliamsPercentR(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedL14_D119_D211_D38_VALUE, "LINEAR_PERCENT_R": expectedL14_D119_D211_D38_LINEAR_PERCENT_R}
		for index, price := range testInput {
			result := indicator.Update(price, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, index), testHigh[index], testLow[index])
			actual := map[string]float64{"VALUE": result.Value, "LINEAR_PERCENT_R": result.LinearPercentR}
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
		p.Length = 50
		p.Delay1 = 100
		p.Delay2 = 60
		p.Delay3 = 40
		indicator, err := NewMetonicWilliamsPercentR(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedL50_D1100_D260_D340_VALUE, "LINEAR_PERCENT_R": expectedL50_D1100_D260_D340_LINEAR_PERCENT_R}
		for index, price := range testInput {
			result := indicator.Update(price, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, index), testHigh[index], testLow[index])
			actual := map[string]float64{"VALUE": result.Value, "LINEAR_PERCENT_R": result.LinearPercentR}
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
		p.Length = 13
		p.Delay1 = 8
		p.Delay2 = 19
		p.Delay3 = 11
		indicator, err := NewMetonicWilliamsPercentR(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedL13_D18_D219_D311_VALUE, "LINEAR_PERCENT_R": expectedL13_D18_D219_D311_LINEAR_PERCENT_R}
		for index, price := range testInput {
			result := indicator.Update(price, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, index), testHigh[index], testLow[index])
			actual := map[string]float64{"VALUE": result.Value, "LINEAR_PERCENT_R": result.LinearPercentR}
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
		p.Length = 13
		p.Delay1 = 239
		p.Delay2 = 100
		p.Delay3 = 50
		indicator, err := NewMetonicWilliamsPercentR(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedL13_D1239_D2100_D350_VALUE, "LINEAR_PERCENT_R": expectedL13_D1239_D2100_D350_LINEAR_PERCENT_R}
		for index, price := range testInput {
			result := indicator.Update(price, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, index), testHigh[index], testLow[index])
			actual := map[string]float64{"VALUE": result.Value, "LINEAR_PERCENT_R": result.LinearPercentR}
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
		p.Length = 13
		p.Delay1 = 240
		p.Delay2 = 100
		p.Delay3 = 50
		indicator, err := NewMetonicWilliamsPercentR(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedL13_D1240_D2100_D350_VALUE, "LINEAR_PERCENT_R": expectedL13_D1240_D2100_D350_LINEAR_PERCENT_R}
		for index, price := range testInput {
			result := indicator.Update(price, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, index), testHigh[index], testLow[index])
			actual := map[string]float64{"VALUE": result.Value, "LINEAR_PERCENT_R": result.LinearPercentR}
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
