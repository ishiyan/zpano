//nolint:testpackage
package coronaspectrum

import (
	"fmt"
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

const testRelTolerance = 1e-10
const testAbsTolerance = 1e-10

func assertClose(t *testing.T, actual, expected float64, label string) {
	t.Helper()
	if math.IsNaN(expected) {
		if !math.IsNaN(actual) {
			t.Fatalf("%s: expected NaN, got %v", label, actual)
		}
		return
	}
	if math.IsInf(expected, 0) {
		if actual != expected {
			t.Fatalf("%s: expected %v, got %v", label, expected, actual)
		}
		return
	}
	if math.IsNaN(actual) || math.IsInf(actual, 0) {
		t.Fatalf("%s: expected %v, got %v", label, expected, actual)
	}
	bound := max(testAbsTolerance, math.Abs(expected)*testRelTolerance)
	if math.Abs(actual-expected) > bound {
		t.Fatalf("%s: expected %.15g, got %.15g (diff %.3e > %.3e)", label, expected, actual, math.Abs(actual-expected), bound)
	}
}

func TestReferenceCases(t *testing.T) {
	t.Parallel()

	type refCase struct {
		name          string
		hpCutoff      int
		minPeriod     int
		maxPeriod     int
		threshold     float64
		heatmap       [][]float64
		heatmapMin    []float64
		heatmapMax    []float64
		dc            []float64
		dcm           []float64
		maxPower      []float64
		normPower     []float64
		selectedBins  []float64
		hasEstimate   []float64
		power         [][]float64
		relativePower [][]float64
	}

	cases := []refCase{
		{
			name:          "H30_PMIN6_PMAX30_DB6",
			hpCutoff:      30,
			minPeriod:     6,
			maxPeriod:     30,
			threshold:     6,
			heatmap:       expectedH30_PMIN6_PMAX30_DB6_HEATMAP,
			heatmapMin:    expectedH30_PMIN6_PMAX30_DB6_HEATMAP_MIN,
			heatmapMax:    expectedH30_PMIN6_PMAX30_DB6_HEATMAP_MAX,
			dc:            expectedH30_PMIN6_PMAX30_DB6_DOMINANT_CYCLE,
			dcm:           expectedH30_PMIN6_PMAX30_DB6_DOMINANT_CYCLE_MEDIAN,
			maxPower:      expectedH30_PMIN6_PMAX30_DB6_MAXIMUM_POWER,
			normPower:     expectedH30_PMIN6_PMAX30_DB6_NORMALIZATION_POWER,
			selectedBins:  expectedH30_PMIN6_PMAX30_DB6_SELECTED_BINS,
			hasEstimate:   expectedH30_PMIN6_PMAX30_DB6_HAS_CYCLE_ESTIMATE,
			power:         expectedH30_PMIN6_PMAX30_DB6_POWER,
			relativePower: expectedH30_PMIN6_PMAX30_DB6_RELATIVE_POWER,
		},
		{
			name:          "H30_PMIN6_PMAX30_DB3",
			hpCutoff:      30,
			minPeriod:     6,
			maxPeriod:     30,
			threshold:     3,
			heatmap:       expectedH30_PMIN6_PMAX30_DB3_HEATMAP,
			heatmapMin:    expectedH30_PMIN6_PMAX30_DB3_HEATMAP_MIN,
			heatmapMax:    expectedH30_PMIN6_PMAX30_DB3_HEATMAP_MAX,
			dc:            expectedH30_PMIN6_PMAX30_DB3_DOMINANT_CYCLE,
			dcm:           expectedH30_PMIN6_PMAX30_DB3_DOMINANT_CYCLE_MEDIAN,
			maxPower:      expectedH30_PMIN6_PMAX30_DB3_MAXIMUM_POWER,
			normPower:     expectedH30_PMIN6_PMAX30_DB3_NORMALIZATION_POWER,
			selectedBins:  expectedH30_PMIN6_PMAX30_DB3_SELECTED_BINS,
			hasEstimate:   expectedH30_PMIN6_PMAX30_DB3_HAS_CYCLE_ESTIMATE,
			power:         expectedH30_PMIN6_PMAX30_DB3_POWER,
			relativePower: expectedH30_PMIN6_PMAX30_DB3_RELATIVE_POWER,
		},
		{
			name:          "H30_PMIN6_PMAX20_DB6",
			hpCutoff:      30,
			minPeriod:     6,
			maxPeriod:     20,
			threshold:     6,
			heatmap:       expectedH30_PMIN6_PMAX20_DB6_HEATMAP,
			heatmapMin:    expectedH30_PMIN6_PMAX20_DB6_HEATMAP_MIN,
			heatmapMax:    expectedH30_PMIN6_PMAX20_DB6_HEATMAP_MAX,
			dc:            expectedH30_PMIN6_PMAX20_DB6_DOMINANT_CYCLE,
			dcm:           expectedH30_PMIN6_PMAX20_DB6_DOMINANT_CYCLE_MEDIAN,
			maxPower:      expectedH30_PMIN6_PMAX20_DB6_MAXIMUM_POWER,
			normPower:     expectedH30_PMIN6_PMAX20_DB6_NORMALIZATION_POWER,
			selectedBins:  expectedH30_PMIN6_PMAX20_DB6_SELECTED_BINS,
			hasEstimate:   expectedH30_PMIN6_PMAX20_DB6_HAS_CYCLE_ESTIMATE,
			power:         expectedH30_PMIN6_PMAX20_DB6_POWER,
			relativePower: expectedH30_PMIN6_PMAX20_DB6_RELATIVE_POWER,
		},
		{
			name:          "H30_PMIN6_PMAX20_DB3",
			hpCutoff:      30,
			minPeriod:     6,
			maxPeriod:     20,
			threshold:     3,
			heatmap:       expectedH30_PMIN6_PMAX20_DB3_HEATMAP,
			heatmapMin:    expectedH30_PMIN6_PMAX20_DB3_HEATMAP_MIN,
			heatmapMax:    expectedH30_PMIN6_PMAX20_DB3_HEATMAP_MAX,
			dc:            expectedH30_PMIN6_PMAX20_DB3_DOMINANT_CYCLE,
			dcm:           expectedH30_PMIN6_PMAX20_DB3_DOMINANT_CYCLE_MEDIAN,
			maxPower:      expectedH30_PMIN6_PMAX20_DB3_MAXIMUM_POWER,
			normPower:     expectedH30_PMIN6_PMAX20_DB3_NORMALIZATION_POWER,
			selectedBins:  expectedH30_PMIN6_PMAX20_DB3_SELECTED_BINS,
			hasEstimate:   expectedH30_PMIN6_PMAX20_DB3_HAS_CYCLE_ESTIMATE,
			power:         expectedH30_PMIN6_PMAX20_DB3_POWER,
			relativePower: expectedH30_PMIN6_PMAX20_DB3_RELATIVE_POWER,
		},
		{
			name:          "H30_PMIN10_PMAX30_DB6",
			hpCutoff:      30,
			minPeriod:     10,
			maxPeriod:     30,
			threshold:     6,
			heatmap:       expectedH30_PMIN10_PMAX30_DB6_HEATMAP,
			heatmapMin:    expectedH30_PMIN10_PMAX30_DB6_HEATMAP_MIN,
			heatmapMax:    expectedH30_PMIN10_PMAX30_DB6_HEATMAP_MAX,
			dc:            expectedH30_PMIN10_PMAX30_DB6_DOMINANT_CYCLE,
			dcm:           expectedH30_PMIN10_PMAX30_DB6_DOMINANT_CYCLE_MEDIAN,
			maxPower:      expectedH30_PMIN10_PMAX30_DB6_MAXIMUM_POWER,
			normPower:     expectedH30_PMIN10_PMAX30_DB6_NORMALIZATION_POWER,
			selectedBins:  expectedH30_PMIN10_PMAX30_DB6_SELECTED_BINS,
			hasEstimate:   expectedH30_PMIN10_PMAX30_DB6_HAS_CYCLE_ESTIMATE,
			power:         expectedH30_PMIN10_PMAX30_DB6_POWER,
			relativePower: expectedH30_PMIN10_PMAX30_DB6_RELATIVE_POWER,
		},
		{
			name:          "H30_PMIN10_PMAX30_DB3",
			hpCutoff:      30,
			minPeriod:     10,
			maxPeriod:     30,
			threshold:     3,
			heatmap:       expectedH30_PMIN10_PMAX30_DB3_HEATMAP,
			heatmapMin:    expectedH30_PMIN10_PMAX30_DB3_HEATMAP_MIN,
			heatmapMax:    expectedH30_PMIN10_PMAX30_DB3_HEATMAP_MAX,
			dc:            expectedH30_PMIN10_PMAX30_DB3_DOMINANT_CYCLE,
			dcm:           expectedH30_PMIN10_PMAX30_DB3_DOMINANT_CYCLE_MEDIAN,
			maxPower:      expectedH30_PMIN10_PMAX30_DB3_MAXIMUM_POWER,
			normPower:     expectedH30_PMIN10_PMAX30_DB3_NORMALIZATION_POWER,
			selectedBins:  expectedH30_PMIN10_PMAX30_DB3_SELECTED_BINS,
			hasEstimate:   expectedH30_PMIN10_PMAX30_DB3_HAS_CYCLE_ESTIMATE,
			power:         expectedH30_PMIN10_PMAX30_DB3_POWER,
			relativePower: expectedH30_PMIN10_PMAX30_DB3_RELATIVE_POWER,
		},
		{
			name:          "H30_PMIN10_PMAX48_DB6",
			hpCutoff:      30,
			minPeriod:     10,
			maxPeriod:     48,
			threshold:     6,
			heatmap:       expectedH30_PMIN10_PMAX48_DB6_HEATMAP,
			heatmapMin:    expectedH30_PMIN10_PMAX48_DB6_HEATMAP_MIN,
			heatmapMax:    expectedH30_PMIN10_PMAX48_DB6_HEATMAP_MAX,
			dc:            expectedH30_PMIN10_PMAX48_DB6_DOMINANT_CYCLE,
			dcm:           expectedH30_PMIN10_PMAX48_DB6_DOMINANT_CYCLE_MEDIAN,
			maxPower:      expectedH30_PMIN10_PMAX48_DB6_MAXIMUM_POWER,
			normPower:     expectedH30_PMIN10_PMAX48_DB6_NORMALIZATION_POWER,
			selectedBins:  expectedH30_PMIN10_PMAX48_DB6_SELECTED_BINS,
			hasEstimate:   expectedH30_PMIN10_PMAX48_DB6_HAS_CYCLE_ESTIMATE,
			power:         expectedH30_PMIN10_PMAX48_DB6_POWER,
			relativePower: expectedH30_PMIN10_PMAX48_DB6_RELATIVE_POWER,
		},
		{
			name:          "H30_PMIN10_PMAX48_DB3",
			hpCutoff:      30,
			minPeriod:     10,
			maxPeriod:     48,
			threshold:     3,
			heatmap:       expectedH30_PMIN10_PMAX48_DB3_HEATMAP,
			heatmapMin:    expectedH30_PMIN10_PMAX48_DB3_HEATMAP_MIN,
			heatmapMax:    expectedH30_PMIN10_PMAX48_DB3_HEATMAP_MAX,
			dc:            expectedH30_PMIN10_PMAX48_DB3_DOMINANT_CYCLE,
			dcm:           expectedH30_PMIN10_PMAX48_DB3_DOMINANT_CYCLE_MEDIAN,
			maxPower:      expectedH30_PMIN10_PMAX48_DB3_MAXIMUM_POWER,
			normPower:     expectedH30_PMIN10_PMAX48_DB3_NORMALIZATION_POWER,
			selectedBins:  expectedH30_PMIN10_PMAX48_DB3_SELECTED_BINS,
			hasEstimate:   expectedH30_PMIN10_PMAX48_DB3_HAS_CYCLE_ESTIMATE,
			power:         expectedH30_PMIN10_PMAX48_DB3_POWER,
			relativePower: expectedH30_PMIN10_PMAX48_DB3_RELATIVE_POWER,
		},
		{
			name:          "H48_PMIN6_PMAX30_DB6",
			hpCutoff:      48,
			minPeriod:     6,
			maxPeriod:     30,
			threshold:     6,
			heatmap:       expectedH48_PMIN6_PMAX30_DB6_HEATMAP,
			heatmapMin:    expectedH48_PMIN6_PMAX30_DB6_HEATMAP_MIN,
			heatmapMax:    expectedH48_PMIN6_PMAX30_DB6_HEATMAP_MAX,
			dc:            expectedH48_PMIN6_PMAX30_DB6_DOMINANT_CYCLE,
			dcm:           expectedH48_PMIN6_PMAX30_DB6_DOMINANT_CYCLE_MEDIAN,
			maxPower:      expectedH48_PMIN6_PMAX30_DB6_MAXIMUM_POWER,
			normPower:     expectedH48_PMIN6_PMAX30_DB6_NORMALIZATION_POWER,
			selectedBins:  expectedH48_PMIN6_PMAX30_DB6_SELECTED_BINS,
			hasEstimate:   expectedH48_PMIN6_PMAX30_DB6_HAS_CYCLE_ESTIMATE,
			power:         expectedH48_PMIN6_PMAX30_DB6_POWER,
			relativePower: expectedH48_PMIN6_PMAX30_DB6_RELATIVE_POWER,
		},
		{
			name:          "H48_PMIN6_PMAX30_DB3",
			hpCutoff:      48,
			minPeriod:     6,
			maxPeriod:     30,
			threshold:     3,
			heatmap:       expectedH48_PMIN6_PMAX30_DB3_HEATMAP,
			heatmapMin:    expectedH48_PMIN6_PMAX30_DB3_HEATMAP_MIN,
			heatmapMax:    expectedH48_PMIN6_PMAX30_DB3_HEATMAP_MAX,
			dc:            expectedH48_PMIN6_PMAX30_DB3_DOMINANT_CYCLE,
			dcm:           expectedH48_PMIN6_PMAX30_DB3_DOMINANT_CYCLE_MEDIAN,
			maxPower:      expectedH48_PMIN6_PMAX30_DB3_MAXIMUM_POWER,
			normPower:     expectedH48_PMIN6_PMAX30_DB3_NORMALIZATION_POWER,
			selectedBins:  expectedH48_PMIN6_PMAX30_DB3_SELECTED_BINS,
			hasEstimate:   expectedH48_PMIN6_PMAX30_DB3_HAS_CYCLE_ESTIMATE,
			power:         expectedH48_PMIN6_PMAX30_DB3_POWER,
			relativePower: expectedH48_PMIN6_PMAX30_DB3_RELATIVE_POWER,
		},
		{
			name:          "H48_PMIN6_PMAX20_DB6",
			hpCutoff:      48,
			minPeriod:     6,
			maxPeriod:     20,
			threshold:     6,
			heatmap:       expectedH48_PMIN6_PMAX20_DB6_HEATMAP,
			heatmapMin:    expectedH48_PMIN6_PMAX20_DB6_HEATMAP_MIN,
			heatmapMax:    expectedH48_PMIN6_PMAX20_DB6_HEATMAP_MAX,
			dc:            expectedH48_PMIN6_PMAX20_DB6_DOMINANT_CYCLE,
			dcm:           expectedH48_PMIN6_PMAX20_DB6_DOMINANT_CYCLE_MEDIAN,
			maxPower:      expectedH48_PMIN6_PMAX20_DB6_MAXIMUM_POWER,
			normPower:     expectedH48_PMIN6_PMAX20_DB6_NORMALIZATION_POWER,
			selectedBins:  expectedH48_PMIN6_PMAX20_DB6_SELECTED_BINS,
			hasEstimate:   expectedH48_PMIN6_PMAX20_DB6_HAS_CYCLE_ESTIMATE,
			power:         expectedH48_PMIN6_PMAX20_DB6_POWER,
			relativePower: expectedH48_PMIN6_PMAX20_DB6_RELATIVE_POWER,
		},
		{
			name:          "H48_PMIN6_PMAX20_DB3",
			hpCutoff:      48,
			minPeriod:     6,
			maxPeriod:     20,
			threshold:     3,
			heatmap:       expectedH48_PMIN6_PMAX20_DB3_HEATMAP,
			heatmapMin:    expectedH48_PMIN6_PMAX20_DB3_HEATMAP_MIN,
			heatmapMax:    expectedH48_PMIN6_PMAX20_DB3_HEATMAP_MAX,
			dc:            expectedH48_PMIN6_PMAX20_DB3_DOMINANT_CYCLE,
			dcm:           expectedH48_PMIN6_PMAX20_DB3_DOMINANT_CYCLE_MEDIAN,
			maxPower:      expectedH48_PMIN6_PMAX20_DB3_MAXIMUM_POWER,
			normPower:     expectedH48_PMIN6_PMAX20_DB3_NORMALIZATION_POWER,
			selectedBins:  expectedH48_PMIN6_PMAX20_DB3_SELECTED_BINS,
			hasEstimate:   expectedH48_PMIN6_PMAX20_DB3_HAS_CYCLE_ESTIMATE,
			power:         expectedH48_PMIN6_PMAX20_DB3_POWER,
			relativePower: expectedH48_PMIN6_PMAX20_DB3_RELATIVE_POWER,
		},
		{
			name:          "H48_PMIN10_PMAX30_DB6",
			hpCutoff:      48,
			minPeriod:     10,
			maxPeriod:     30,
			threshold:     6,
			heatmap:       expectedH48_PMIN10_PMAX30_DB6_HEATMAP,
			heatmapMin:    expectedH48_PMIN10_PMAX30_DB6_HEATMAP_MIN,
			heatmapMax:    expectedH48_PMIN10_PMAX30_DB6_HEATMAP_MAX,
			dc:            expectedH48_PMIN10_PMAX30_DB6_DOMINANT_CYCLE,
			dcm:           expectedH48_PMIN10_PMAX30_DB6_DOMINANT_CYCLE_MEDIAN,
			maxPower:      expectedH48_PMIN10_PMAX30_DB6_MAXIMUM_POWER,
			normPower:     expectedH48_PMIN10_PMAX30_DB6_NORMALIZATION_POWER,
			selectedBins:  expectedH48_PMIN10_PMAX30_DB6_SELECTED_BINS,
			hasEstimate:   expectedH48_PMIN10_PMAX30_DB6_HAS_CYCLE_ESTIMATE,
			power:         expectedH48_PMIN10_PMAX30_DB6_POWER,
			relativePower: expectedH48_PMIN10_PMAX30_DB6_RELATIVE_POWER,
		},
		{
			name:          "H48_PMIN10_PMAX30_DB3",
			hpCutoff:      48,
			minPeriod:     10,
			maxPeriod:     30,
			threshold:     3,
			heatmap:       expectedH48_PMIN10_PMAX30_DB3_HEATMAP,
			heatmapMin:    expectedH48_PMIN10_PMAX30_DB3_HEATMAP_MIN,
			heatmapMax:    expectedH48_PMIN10_PMAX30_DB3_HEATMAP_MAX,
			dc:            expectedH48_PMIN10_PMAX30_DB3_DOMINANT_CYCLE,
			dcm:           expectedH48_PMIN10_PMAX30_DB3_DOMINANT_CYCLE_MEDIAN,
			maxPower:      expectedH48_PMIN10_PMAX30_DB3_MAXIMUM_POWER,
			normPower:     expectedH48_PMIN10_PMAX30_DB3_NORMALIZATION_POWER,
			selectedBins:  expectedH48_PMIN10_PMAX30_DB3_SELECTED_BINS,
			hasEstimate:   expectedH48_PMIN10_PMAX30_DB3_HAS_CYCLE_ESTIMATE,
			power:         expectedH48_PMIN10_PMAX30_DB3_POWER,
			relativePower: expectedH48_PMIN10_PMAX30_DB3_RELATIVE_POWER,
		},
		{
			name:          "H48_PMIN10_PMAX48_DB6",
			hpCutoff:      48,
			minPeriod:     10,
			maxPeriod:     48,
			threshold:     6,
			heatmap:       expectedH48_PMIN10_PMAX48_DB6_HEATMAP,
			heatmapMin:    expectedH48_PMIN10_PMAX48_DB6_HEATMAP_MIN,
			heatmapMax:    expectedH48_PMIN10_PMAX48_DB6_HEATMAP_MAX,
			dc:            expectedH48_PMIN10_PMAX48_DB6_DOMINANT_CYCLE,
			dcm:           expectedH48_PMIN10_PMAX48_DB6_DOMINANT_CYCLE_MEDIAN,
			maxPower:      expectedH48_PMIN10_PMAX48_DB6_MAXIMUM_POWER,
			normPower:     expectedH48_PMIN10_PMAX48_DB6_NORMALIZATION_POWER,
			selectedBins:  expectedH48_PMIN10_PMAX48_DB6_SELECTED_BINS,
			hasEstimate:   expectedH48_PMIN10_PMAX48_DB6_HAS_CYCLE_ESTIMATE,
			power:         expectedH48_PMIN10_PMAX48_DB6_POWER,
			relativePower: expectedH48_PMIN10_PMAX48_DB6_RELATIVE_POWER,
		},
		{
			name:          "H48_PMIN10_PMAX48_DB3",
			hpCutoff:      48,
			minPeriod:     10,
			maxPeriod:     48,
			threshold:     3,
			heatmap:       expectedH48_PMIN10_PMAX48_DB3_HEATMAP,
			heatmapMin:    expectedH48_PMIN10_PMAX48_DB3_HEATMAP_MIN,
			heatmapMax:    expectedH48_PMIN10_PMAX48_DB3_HEATMAP_MAX,
			dc:            expectedH48_PMIN10_PMAX48_DB3_DOMINANT_CYCLE,
			dcm:           expectedH48_PMIN10_PMAX48_DB3_DOMINANT_CYCLE_MEDIAN,
			maxPower:      expectedH48_PMIN10_PMAX48_DB3_MAXIMUM_POWER,
			normPower:     expectedH48_PMIN10_PMAX48_DB3_NORMALIZATION_POWER,
			selectedBins:  expectedH48_PMIN10_PMAX48_DB3_SELECTED_BINS,
			hasEstimate:   expectedH48_PMIN10_PMAX48_DB3_HAS_CYCLE_ESTIMATE,
			power:         expectedH48_PMIN10_PMAX48_DB3_POWER,
			relativePower: expectedH48_PMIN10_PMAX48_DB3_RELATIVE_POWER,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			params := &Params{
				MinRasterValue:       tc.threshold,
				MaxRasterValue:       20.0,
				MinParameterValue:    float64(tc.minPeriod),
				MaxParameterValue:    float64(tc.maxPeriod),
				HighPassFilterCutoff: tc.hpCutoff,
			}
			indicator, err := NewCoronaSpectrumParams(params)
			if err != nil {
				t.Fatalf("NewCoronaSpectrumParams: %v", err)
			}
			cells := 2*(tc.maxPeriod-tc.minPeriod) + 1
			baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

			for i, price := range testInput {
				barTime := baseTime.AddDate(0, 0, i)
				snap := indicator.Update(price, barTime)
				assertClose(t, snap.Heatmap.ParameterFirst, float64(tc.minPeriod), fmt.Sprintf("%s[%d].first", tc.name, i))
				assertClose(t, snap.Heatmap.ParameterLast, float64(tc.maxPeriod), fmt.Sprintf("%s[%d].last", tc.name, i))
				assertClose(t, snap.Heatmap.ParameterResolution, 2.0, fmt.Sprintf("%s[%d].resolution", tc.name, i))

				if i < 12 {
					if indicator.IsPrimed() {
						t.Fatalf("%s[%d]: expected IsPrimed=false", tc.name, i)
					}
					if !snap.Heatmap.IsEmpty() {
						t.Fatalf("%s[%d]: expected empty heatmap", tc.name, i)
					}
					if len(snap.Power) != cells || len(snap.RelativePower) != cells {
						t.Fatalf("%s[%d]: expected NaN slices of length %d", tc.name, i, cells)
					}
					for binIndex := 0; binIndex < cells; binIndex++ {
						if !math.IsNaN(snap.Power[binIndex]) || !math.IsNaN(snap.RelativePower[binIndex]) {
							t.Fatalf("%s[%d][%d]: expected NaN power/relativePower", tc.name, i, binIndex)
						}
					}
				} else {
					if !indicator.IsPrimed() {
						t.Fatalf("%s[%d]: expected IsPrimed=true", tc.name, i)
					}
					if len(snap.Heatmap.Values) != cells {
						t.Fatalf("%s[%d]: heatmap len = %d, want %d", tc.name, i, len(snap.Heatmap.Values), cells)
					}
					for binIndex := 0; binIndex < cells; binIndex++ {
						assertClose(t, snap.Heatmap.Values[binIndex], tc.heatmap[i][binIndex], fmt.Sprintf("%s[%d].hm[%d]", tc.name, i, binIndex))
						assertClose(t, snap.Power[binIndex], tc.power[i][binIndex], fmt.Sprintf("%s[%d].pow[%d]", tc.name, i, binIndex))
						assertClose(t, snap.RelativePower[binIndex], tc.relativePower[i][binIndex], fmt.Sprintf("%s[%d].rel[%d]", tc.name, i, binIndex))
					}
				}

				assertClose(t, snap.Heatmap.ValueMin, tc.heatmapMin[i], fmt.Sprintf("%s[%d].min", tc.name, i))
				assertClose(t, snap.Heatmap.ValueMax, tc.heatmapMax[i], fmt.Sprintf("%s[%d].max", tc.name, i))
				assertClose(t, snap.DominantCycle, tc.dc[i], fmt.Sprintf("%s[%d].dc", tc.name, i))
				assertClose(t, snap.DominantCycleMedian, tc.dcm[i], fmt.Sprintf("%s[%d].dcm", tc.name, i))
				assertClose(t, snap.MaximumPower, tc.maxPower[i], fmt.Sprintf("%s[%d].maxPower", tc.name, i))
				assertClose(t, snap.NormalizationPower, tc.normPower[i], fmt.Sprintf("%s[%d].normPower", tc.name, i))
				assertClose(t, snap.SelectedBins, tc.selectedBins[i], fmt.Sprintf("%s[%d].selected", tc.name, i))
				assertClose(t, snap.HasCycleEstimate, tc.hasEstimate[i], fmt.Sprintf("%s[%d].hasEst", tc.name, i))
			}

			indicator.Reset()
			if indicator.IsPrimed() {
				t.Fatalf("%s: expected IsPrimed=false after Reset", tc.name)
			}
		})
	}
}

func TestMetadataMnemonicsAndEntities(t *testing.T) {
	t.Parallel()

	indicator, err := NewCoronaSpectrumDefault()
	if err != nil {
		t.Fatalf("NewCoronaSpectrumDefault: %v", err)
	}
	meta := indicator.Metadata()
	if meta.Identifier != core.CoronaSpectrum {
		t.Fatalf("Identifier = %v, want %v", meta.Identifier, core.CoronaSpectrum)
	}
	if meta.Mnemonic != "cspect(6, 20, 6, 30, 30)" {
		t.Fatalf("Mnemonic = %q", meta.Mnemonic)
	}
	if len(meta.Outputs) != 3 {
		t.Fatalf("Outputs len = %d, want 3", len(meta.Outputs))
	}
	if meta.Outputs[0].Shape != shape.Heatmap || meta.Outputs[1].Shape != shape.Scalar || meta.Outputs[2].Shape != shape.Scalar {
		t.Fatalf("unexpected output shapes: %+v", meta.Outputs)
	}
	if meta.Outputs[1].Mnemonic != "cspect-dc(30)" || meta.Outputs[2].Mnemonic != "cspect-dcm(30)" {
		t.Fatalf("unexpected sub-output mnemonics: %+v", meta.Outputs)
	}

	combos := []struct {
		bc entities.BarComponent
		qc entities.QuoteComponent
		tc entities.TradeComponent
	}{
		{0, 0, 0},
		{entities.BarOpenPrice, 0, 0},
		{0, entities.QuoteBidPrice, 0},
		{0, 0, entities.TradeVolume},
		{entities.BarMedianPrice, entities.QuoteAskPrice, 0},
		{entities.BarMedianPrice, 0, entities.TradeVolume},
		{0, entities.QuoteAskPrice, entities.TradeVolume},
		{entities.BarOpenPrice, entities.QuoteBidPrice, entities.TradeVolume},
	}
	for _, combo := range combos {
		params := DefaultParams()
		params.BarComponent = combo.bc
		params.QuoteComponent = combo.qc
		params.TradeComponent = combo.tc
		instance, err := NewCoronaSpectrumParams(params)
		if err != nil {
			t.Fatalf("NewCoronaSpectrumParams: %v", err)
		}
		resolvedBC, resolvedQC, resolvedTC := combo.bc, combo.qc, combo.tc
		if resolvedBC == 0 {
			resolvedBC = entities.DefaultBarComponent
		}
		if resolvedQC == 0 {
			resolvedQC = entities.DefaultQuoteComponent
		}
		if resolvedTC == 0 {
			resolvedTC = entities.DefaultTradeComponent
		}
		expected := "cspect(6, 20, 6, 30, 30" + core.ComponentTripleMnemonic(resolvedBC, resolvedQC, resolvedTC) + ")"
		if instance.Metadata().Mnemonic != expected || params.Mnemonic() != expected {
			t.Fatalf("mnemonic mismatch: got %q / %q, want %q", instance.Metadata().Mnemonic, params.Mnemonic(), expected)
		}
	}

	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	barInd, _ := NewCoronaSpectrumDefault()
	scalarInd, _ := NewCoronaSpectrumDefault()
	quoteInd, _ := NewCoronaSpectrumDefault()
	tradeInd, _ := NewCoronaSpectrumDefault()
	for i := 0; i < 15; i++ {
		barTime := baseTime.AddDate(0, 0, i)
		price := testInput[i]
		outBar := barInd.UpdateBar(&entities.Bar{Time: barTime, Open: testOpen[i], High: testHigh[i], Low: testLow[i], Close: price, Volume: 1000})
		outScalar := scalarInd.UpdateScalar(&entities.Scalar{Time: barTime, Value: price})
		outQuote := quoteInd.UpdateQuote(&entities.Quote{Time: barTime, Bid: price - 0.5, Ask: price + 0.5, BidSize: 10, AskSize: 10})
		outTrade := tradeInd.UpdateTrade(&entities.Trade{Time: barTime, Price: price, Volume: 10})
		if len(outBar) != 3 {
			t.Fatalf("len(outBar) = %d, want 3", len(outBar))
		}
		assertClose(t, outBar[1].(entities.Scalar).Value, outScalar[1].(entities.Scalar).Value, "bar vs scalar")
		assertClose(t, outBar[1].(entities.Scalar).Value, outQuote[1].(entities.Scalar).Value, "bar vs quote")
		assertClose(t, outBar[1].(entities.Scalar).Value, outTrade[1].(entities.Scalar).Value, "bar vs trade")
	}
}

func TestValidation(t *testing.T) {
	t.Parallel()

	invalid := []*Params{
		{MinRasterValue: -0.1, MaxRasterValue: 20, MinParameterValue: 6, MaxParameterValue: 30, HighPassFilterCutoff: 30},
		{MinRasterValue: 20, MaxRasterValue: 20, MinParameterValue: 6, MaxParameterValue: 30, HighPassFilterCutoff: 30},
		{MinRasterValue: 6, MaxRasterValue: 20, MinParameterValue: 5, MaxParameterValue: 30, HighPassFilterCutoff: 30},
		{MinRasterValue: 6, MaxRasterValue: 20, MinParameterValue: 0, MaxParameterValue: 30, HighPassFilterCutoff: 30},
		{MinRasterValue: 6, MaxRasterValue: 20, MinParameterValue: 10, MaxParameterValue: 10, HighPassFilterCutoff: 30},
		{MinRasterValue: 6, MaxRasterValue: 20, MinParameterValue: 6, MaxParameterValue: 30, HighPassFilterCutoff: 4},
		{MinRasterValue: 6, MaxRasterValue: 20, MinParameterValue: 6, MaxParameterValue: 30, HighPassFilterCutoff: 0},
	}
	for i, bad := range invalid {
		if _, err := NewCoronaSpectrumParams(bad); err == nil {
			t.Fatalf("invalid[%d]: expected error, got nil", i)
		}
	}
}
