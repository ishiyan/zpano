// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package dynamiccycleexplorer

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
		p.MaxPeriod = 80
		p.MaxSegments = 5
		p.MinPeriod = 5
		p.PeakDistance = 2
		p.PhaseCycles = 2.0
		p.Smoothing = 1600.0
		p.TopK = 3
		p.Window = 200
		indicator, err := NewDynamicCycleExplorer(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_VALUE, "PERIOD": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_PERIOD, "PHASE": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_PHASE, "ETA_TOP": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_ETA_TOP, "ETA_BOTTOM": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_ETA_BOTTOM, "HIGHLIGHTER": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_HIGHLIGHTER, "HIGHLIGHTER_INDEX": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_HIGHLIGHTER_INDEX, "FITTED_AT_HIGHLIGHTER": expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_FITTED_AT_HIGHLIGHTER}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "PERIOD": result.Period, "PHASE": result.Phase, "ETA_TOP": result.EtaTop, "ETA_BOTTOM": result.EtaBottom, "HIGHLIGHTER": result.Highlighter, "HIGHLIGHTER_INDEX": result.HighlighterIndex, "FITTED_AT_HIGHLIGHTER": result.FittedAtHighlighter}
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
		p.MaxPeriod = 12
		p.MaxSegments = 5
		p.MinPeriod = 12
		p.PeakDistance = 2
		p.PhaseCycles = 0.0
		p.Smoothing = 1600.0
		p.TopK = 1
		p.Window = 60
		indicator, err := NewDynamicCycleExplorer(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_VALUE, "PERIOD": expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_PERIOD, "PHASE": expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_PHASE, "ETA_TOP": expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_ETA_TOP, "ETA_BOTTOM": expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_ETA_BOTTOM, "HIGHLIGHTER": expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_HIGHLIGHTER, "HIGHLIGHTER_INDEX": expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_HIGHLIGHTER_INDEX, "FITTED_AT_HIGHLIGHTER": expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_FITTED_AT_HIGHLIGHTER}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "PERIOD": result.Period, "PHASE": result.Phase, "ETA_TOP": result.EtaTop, "ETA_BOTTOM": result.EtaBottom, "HIGHLIGHTER": result.Highlighter, "HIGHLIGHTER_INDEX": result.HighlighterIndex, "FITTED_AT_HIGHLIGHTER": result.FittedAtHighlighter}
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
		p.MaxPeriod = 30
		p.MaxSegments = 5
		p.MinPeriod = 5
		p.PeakDistance = 2
		p.PhaseCycles = 2.0
		p.Smoothing = 0.0
		p.TopK = 2
		p.Window = 100
		indicator, err := NewDynamicCycleExplorer(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_VALUE, "PERIOD": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_PERIOD, "PHASE": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_PHASE, "ETA_TOP": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_ETA_TOP, "ETA_BOTTOM": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_ETA_BOTTOM, "HIGHLIGHTER": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_HIGHLIGHTER, "HIGHLIGHTER_INDEX": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_HIGHLIGHTER_INDEX, "FITTED_AT_HIGHLIGHTER": expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_FITTED_AT_HIGHLIGHTER}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "PERIOD": result.Period, "PHASE": result.Phase, "ETA_TOP": result.EtaTop, "ETA_BOTTOM": result.EtaBottom, "HIGHLIGHTER": result.Highlighter, "HIGHLIGHTER_INDEX": result.HighlighterIndex, "FITTED_AT_HIGHLIGHTER": result.FittedAtHighlighter}
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
		p.MaxPeriod = 30
		p.MaxSegments = 5
		p.MinPeriod = 5
		p.PeakDistance = 2
		p.PhaseCycles = 2.0
		p.Smoothing = 1600.0
		p.TopK = 2
		p.Window = 100
		indicator, err := NewDynamicCycleExplorer(p)
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string][]float64{"VALUE": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_VALUE, "PERIOD": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_PERIOD, "PHASE": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_PHASE, "ETA_TOP": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_ETA_TOP, "ETA_BOTTOM": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_ETA_BOTTOM, "HIGHLIGHTER": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_HIGHLIGHTER, "HIGHLIGHTER_INDEX": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_HIGHLIGHTER_INDEX, "FITTED_AT_HIGHLIGHTER": expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_FITTED_AT_HIGHLIGHTER}
		for index, price := range testInput {
			result := indicator.Update(price)
			actual := map[string]float64{"VALUE": result.Value, "PERIOD": result.Period, "PHASE": result.Phase, "ETA_TOP": result.EtaTop, "ETA_BOTTOM": result.EtaBottom, "HIGHLIGHTER": result.Highlighter, "HIGHLIGHTER_INDEX": result.HighlighterIndex, "FITTED_AT_HIGHLIGHTER": result.FittedAtHighlighter}
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
