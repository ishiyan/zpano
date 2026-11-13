//nolint:testpackage
package coronasignaltonoiseratio

import (
	"fmt"
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

const testTolerance = 5e-12

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
	bound := max(testTolerance, math.Abs(expected)*testTolerance)
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
		rasterLength  int
		heatmap       [][]float64
		heatmapMin    []float64
		heatmapMax    []float64
		snr           []float64
		physDB        []float64
		calDB         []float64
		norm          []float64
		signal        []float64
		noise         []float64
		medianRange   []float64
		width         []float64
		dc            []float64
		dcm           []float64
		hasEstimate   []float64
		ratioIsFinite []float64
	}

	cases := []refCase{
		{
			name:          "H30_PMIN6_PMAX30_R50",
			hpCutoff:      30,
			minPeriod:     6,
			maxPeriod:     30,
			rasterLength:  50,
			heatmap:       expectedH30_PMIN6_PMAX30_R50_HEATMAP,
			heatmapMin:    expectedH30_PMIN6_PMAX30_R50_HEATMAP_MIN,
			heatmapMax:    expectedH30_PMIN6_PMAX30_R50_HEATMAP_MAX,
			snr:           expectedH30_PMIN6_PMAX30_R50_SIGNAL_TO_NOISE_RATIO,
			physDB:        expectedH30_PMIN6_PMAX30_R50_PHYSICAL_DECIBELS,
			calDB:         expectedH30_PMIN6_PMAX30_R50_CALIBRATED_DECIBELS,
			norm:          expectedH30_PMIN6_PMAX30_R50_NORMALIZED_POSITION,
			signal:        expectedH30_PMIN6_PMAX30_R50_SIGNAL,
			noise:         expectedH30_PMIN6_PMAX30_R50_NOISE,
			medianRange:   expectedH30_PMIN6_PMAX30_R50_MEDIAN_BAR_RANGE,
			width:         expectedH30_PMIN6_PMAX30_R50_WIDTH,
			dc:            expectedH30_PMIN6_PMAX30_R50_DOMINANT_CYCLE,
			dcm:           expectedH30_PMIN6_PMAX30_R50_DOMINANT_CYCLE_MEDIAN,
			hasEstimate:   expectedH30_PMIN6_PMAX30_R50_HAS_CYCLE_ESTIMATE,
			ratioIsFinite: expectedH30_PMIN6_PMAX30_R50_RATIO_IS_FINITE,
		},
		{
			name:          "H30_PMIN6_PMAX30_R25",
			hpCutoff:      30,
			minPeriod:     6,
			maxPeriod:     30,
			rasterLength:  25,
			heatmap:       expectedH30_PMIN6_PMAX30_R25_HEATMAP,
			heatmapMin:    expectedH30_PMIN6_PMAX30_R25_HEATMAP_MIN,
			heatmapMax:    expectedH30_PMIN6_PMAX30_R25_HEATMAP_MAX,
			snr:           expectedH30_PMIN6_PMAX30_R25_SIGNAL_TO_NOISE_RATIO,
			physDB:        expectedH30_PMIN6_PMAX30_R25_PHYSICAL_DECIBELS,
			calDB:         expectedH30_PMIN6_PMAX30_R25_CALIBRATED_DECIBELS,
			norm:          expectedH30_PMIN6_PMAX30_R25_NORMALIZED_POSITION,
			signal:        expectedH30_PMIN6_PMAX30_R25_SIGNAL,
			noise:         expectedH30_PMIN6_PMAX30_R25_NOISE,
			medianRange:   expectedH30_PMIN6_PMAX30_R25_MEDIAN_BAR_RANGE,
			width:         expectedH30_PMIN6_PMAX30_R25_WIDTH,
			dc:            expectedH30_PMIN6_PMAX30_R25_DOMINANT_CYCLE,
			dcm:           expectedH30_PMIN6_PMAX30_R25_DOMINANT_CYCLE_MEDIAN,
			hasEstimate:   expectedH30_PMIN6_PMAX30_R25_HAS_CYCLE_ESTIMATE,
			ratioIsFinite: expectedH30_PMIN6_PMAX30_R25_RATIO_IS_FINITE,
		},
		{
			name:          "H30_PMIN6_PMAX20_R50",
			hpCutoff:      30,
			minPeriod:     6,
			maxPeriod:     20,
			rasterLength:  50,
			heatmap:       expectedH30_PMIN6_PMAX20_R50_HEATMAP,
			heatmapMin:    expectedH30_PMIN6_PMAX20_R50_HEATMAP_MIN,
			heatmapMax:    expectedH30_PMIN6_PMAX20_R50_HEATMAP_MAX,
			snr:           expectedH30_PMIN6_PMAX20_R50_SIGNAL_TO_NOISE_RATIO,
			physDB:        expectedH30_PMIN6_PMAX20_R50_PHYSICAL_DECIBELS,
			calDB:         expectedH30_PMIN6_PMAX20_R50_CALIBRATED_DECIBELS,
			norm:          expectedH30_PMIN6_PMAX20_R50_NORMALIZED_POSITION,
			signal:        expectedH30_PMIN6_PMAX20_R50_SIGNAL,
			noise:         expectedH30_PMIN6_PMAX20_R50_NOISE,
			medianRange:   expectedH30_PMIN6_PMAX20_R50_MEDIAN_BAR_RANGE,
			width:         expectedH30_PMIN6_PMAX20_R50_WIDTH,
			dc:            expectedH30_PMIN6_PMAX20_R50_DOMINANT_CYCLE,
			dcm:           expectedH30_PMIN6_PMAX20_R50_DOMINANT_CYCLE_MEDIAN,
			hasEstimate:   expectedH30_PMIN6_PMAX20_R50_HAS_CYCLE_ESTIMATE,
			ratioIsFinite: expectedH30_PMIN6_PMAX20_R50_RATIO_IS_FINITE,
		},
		{
			name:          "H30_PMIN6_PMAX20_R25",
			hpCutoff:      30,
			minPeriod:     6,
			maxPeriod:     20,
			rasterLength:  25,
			heatmap:       expectedH30_PMIN6_PMAX20_R25_HEATMAP,
			heatmapMin:    expectedH30_PMIN6_PMAX20_R25_HEATMAP_MIN,
			heatmapMax:    expectedH30_PMIN6_PMAX20_R25_HEATMAP_MAX,
			snr:           expectedH30_PMIN6_PMAX20_R25_SIGNAL_TO_NOISE_RATIO,
			physDB:        expectedH30_PMIN6_PMAX20_R25_PHYSICAL_DECIBELS,
			calDB:         expectedH30_PMIN6_PMAX20_R25_CALIBRATED_DECIBELS,
			norm:          expectedH30_PMIN6_PMAX20_R25_NORMALIZED_POSITION,
			signal:        expectedH30_PMIN6_PMAX20_R25_SIGNAL,
			noise:         expectedH30_PMIN6_PMAX20_R25_NOISE,
			medianRange:   expectedH30_PMIN6_PMAX20_R25_MEDIAN_BAR_RANGE,
			width:         expectedH30_PMIN6_PMAX20_R25_WIDTH,
			dc:            expectedH30_PMIN6_PMAX20_R25_DOMINANT_CYCLE,
			dcm:           expectedH30_PMIN6_PMAX20_R25_DOMINANT_CYCLE_MEDIAN,
			hasEstimate:   expectedH30_PMIN6_PMAX20_R25_HAS_CYCLE_ESTIMATE,
			ratioIsFinite: expectedH30_PMIN6_PMAX20_R25_RATIO_IS_FINITE,
		},
		{
			name:          "H30_PMIN10_PMAX30_R50",
			hpCutoff:      30,
			minPeriod:     10,
			maxPeriod:     30,
			rasterLength:  50,
			heatmap:       expectedH30_PMIN10_PMAX30_R50_HEATMAP,
			heatmapMin:    expectedH30_PMIN10_PMAX30_R50_HEATMAP_MIN,
			heatmapMax:    expectedH30_PMIN10_PMAX30_R50_HEATMAP_MAX,
			snr:           expectedH30_PMIN10_PMAX30_R50_SIGNAL_TO_NOISE_RATIO,
			physDB:        expectedH30_PMIN10_PMAX30_R50_PHYSICAL_DECIBELS,
			calDB:         expectedH30_PMIN10_PMAX30_R50_CALIBRATED_DECIBELS,
			norm:          expectedH30_PMIN10_PMAX30_R50_NORMALIZED_POSITION,
			signal:        expectedH30_PMIN10_PMAX30_R50_SIGNAL,
			noise:         expectedH30_PMIN10_PMAX30_R50_NOISE,
			medianRange:   expectedH30_PMIN10_PMAX30_R50_MEDIAN_BAR_RANGE,
			width:         expectedH30_PMIN10_PMAX30_R50_WIDTH,
			dc:            expectedH30_PMIN10_PMAX30_R50_DOMINANT_CYCLE,
			dcm:           expectedH30_PMIN10_PMAX30_R50_DOMINANT_CYCLE_MEDIAN,
			hasEstimate:   expectedH30_PMIN10_PMAX30_R50_HAS_CYCLE_ESTIMATE,
			ratioIsFinite: expectedH30_PMIN10_PMAX30_R50_RATIO_IS_FINITE,
		},
		{
			name:          "H30_PMIN10_PMAX30_R25",
			hpCutoff:      30,
			minPeriod:     10,
			maxPeriod:     30,
			rasterLength:  25,
			heatmap:       expectedH30_PMIN10_PMAX30_R25_HEATMAP,
			heatmapMin:    expectedH30_PMIN10_PMAX30_R25_HEATMAP_MIN,
			heatmapMax:    expectedH30_PMIN10_PMAX30_R25_HEATMAP_MAX,
			snr:           expectedH30_PMIN10_PMAX30_R25_SIGNAL_TO_NOISE_RATIO,
			physDB:        expectedH30_PMIN10_PMAX30_R25_PHYSICAL_DECIBELS,
			calDB:         expectedH30_PMIN10_PMAX30_R25_CALIBRATED_DECIBELS,
			norm:          expectedH30_PMIN10_PMAX30_R25_NORMALIZED_POSITION,
			signal:        expectedH30_PMIN10_PMAX30_R25_SIGNAL,
			noise:         expectedH30_PMIN10_PMAX30_R25_NOISE,
			medianRange:   expectedH30_PMIN10_PMAX30_R25_MEDIAN_BAR_RANGE,
			width:         expectedH30_PMIN10_PMAX30_R25_WIDTH,
			dc:            expectedH30_PMIN10_PMAX30_R25_DOMINANT_CYCLE,
			dcm:           expectedH30_PMIN10_PMAX30_R25_DOMINANT_CYCLE_MEDIAN,
			hasEstimate:   expectedH30_PMIN10_PMAX30_R25_HAS_CYCLE_ESTIMATE,
			ratioIsFinite: expectedH30_PMIN10_PMAX30_R25_RATIO_IS_FINITE,
		},
		{
			name:          "H30_PMIN10_PMAX48_R50",
			hpCutoff:      30,
			minPeriod:     10,
			maxPeriod:     48,
			rasterLength:  50,
			heatmap:       expectedH30_PMIN10_PMAX48_R50_HEATMAP,
			heatmapMin:    expectedH30_PMIN10_PMAX48_R50_HEATMAP_MIN,
			heatmapMax:    expectedH30_PMIN10_PMAX48_R50_HEATMAP_MAX,
			snr:           expectedH30_PMIN10_PMAX48_R50_SIGNAL_TO_NOISE_RATIO,
			physDB:        expectedH30_PMIN10_PMAX48_R50_PHYSICAL_DECIBELS,
			calDB:         expectedH30_PMIN10_PMAX48_R50_CALIBRATED_DECIBELS,
			norm:          expectedH30_PMIN10_PMAX48_R50_NORMALIZED_POSITION,
			signal:        expectedH30_PMIN10_PMAX48_R50_SIGNAL,
			noise:         expectedH30_PMIN10_PMAX48_R50_NOISE,
			medianRange:   expectedH30_PMIN10_PMAX48_R50_MEDIAN_BAR_RANGE,
			width:         expectedH30_PMIN10_PMAX48_R50_WIDTH,
			dc:            expectedH30_PMIN10_PMAX48_R50_DOMINANT_CYCLE,
			dcm:           expectedH30_PMIN10_PMAX48_R50_DOMINANT_CYCLE_MEDIAN,
			hasEstimate:   expectedH30_PMIN10_PMAX48_R50_HAS_CYCLE_ESTIMATE,
			ratioIsFinite: expectedH30_PMIN10_PMAX48_R50_RATIO_IS_FINITE,
		},
		{
			name:          "H30_PMIN10_PMAX48_R25",
			hpCutoff:      30,
			minPeriod:     10,
			maxPeriod:     48,
			rasterLength:  25,
			heatmap:       expectedH30_PMIN10_PMAX48_R25_HEATMAP,
			heatmapMin:    expectedH30_PMIN10_PMAX48_R25_HEATMAP_MIN,
			heatmapMax:    expectedH30_PMIN10_PMAX48_R25_HEATMAP_MAX,
			snr:           expectedH30_PMIN10_PMAX48_R25_SIGNAL_TO_NOISE_RATIO,
			physDB:        expectedH30_PMIN10_PMAX48_R25_PHYSICAL_DECIBELS,
			calDB:         expectedH30_PMIN10_PMAX48_R25_CALIBRATED_DECIBELS,
			norm:          expectedH30_PMIN10_PMAX48_R25_NORMALIZED_POSITION,
			signal:        expectedH30_PMIN10_PMAX48_R25_SIGNAL,
			noise:         expectedH30_PMIN10_PMAX48_R25_NOISE,
			medianRange:   expectedH30_PMIN10_PMAX48_R25_MEDIAN_BAR_RANGE,
			width:         expectedH30_PMIN10_PMAX48_R25_WIDTH,
			dc:            expectedH30_PMIN10_PMAX48_R25_DOMINANT_CYCLE,
			dcm:           expectedH30_PMIN10_PMAX48_R25_DOMINANT_CYCLE_MEDIAN,
			hasEstimate:   expectedH30_PMIN10_PMAX48_R25_HAS_CYCLE_ESTIMATE,
			ratioIsFinite: expectedH30_PMIN10_PMAX48_R25_RATIO_IS_FINITE,
		},
		{
			name:          "H48_PMIN6_PMAX30_R50",
			hpCutoff:      48,
			minPeriod:     6,
			maxPeriod:     30,
			rasterLength:  50,
			heatmap:       expectedH48_PMIN6_PMAX30_R50_HEATMAP,
			heatmapMin:    expectedH48_PMIN6_PMAX30_R50_HEATMAP_MIN,
			heatmapMax:    expectedH48_PMIN6_PMAX30_R50_HEATMAP_MAX,
			snr:           expectedH48_PMIN6_PMAX30_R50_SIGNAL_TO_NOISE_RATIO,
			physDB:        expectedH48_PMIN6_PMAX30_R50_PHYSICAL_DECIBELS,
			calDB:         expectedH48_PMIN6_PMAX30_R50_CALIBRATED_DECIBELS,
			norm:          expectedH48_PMIN6_PMAX30_R50_NORMALIZED_POSITION,
			signal:        expectedH48_PMIN6_PMAX30_R50_SIGNAL,
			noise:         expectedH48_PMIN6_PMAX30_R50_NOISE,
			medianRange:   expectedH48_PMIN6_PMAX30_R50_MEDIAN_BAR_RANGE,
			width:         expectedH48_PMIN6_PMAX30_R50_WIDTH,
			dc:            expectedH48_PMIN6_PMAX30_R50_DOMINANT_CYCLE,
			dcm:           expectedH48_PMIN6_PMAX30_R50_DOMINANT_CYCLE_MEDIAN,
			hasEstimate:   expectedH48_PMIN6_PMAX30_R50_HAS_CYCLE_ESTIMATE,
			ratioIsFinite: expectedH48_PMIN6_PMAX30_R50_RATIO_IS_FINITE,
		},
		{
			name:          "H48_PMIN6_PMAX30_R25",
			hpCutoff:      48,
			minPeriod:     6,
			maxPeriod:     30,
			rasterLength:  25,
			heatmap:       expectedH48_PMIN6_PMAX30_R25_HEATMAP,
			heatmapMin:    expectedH48_PMIN6_PMAX30_R25_HEATMAP_MIN,
			heatmapMax:    expectedH48_PMIN6_PMAX30_R25_HEATMAP_MAX,
			snr:           expectedH48_PMIN6_PMAX30_R25_SIGNAL_TO_NOISE_RATIO,
			physDB:        expectedH48_PMIN6_PMAX30_R25_PHYSICAL_DECIBELS,
			calDB:         expectedH48_PMIN6_PMAX30_R25_CALIBRATED_DECIBELS,
			norm:          expectedH48_PMIN6_PMAX30_R25_NORMALIZED_POSITION,
			signal:        expectedH48_PMIN6_PMAX30_R25_SIGNAL,
			noise:         expectedH48_PMIN6_PMAX30_R25_NOISE,
			medianRange:   expectedH48_PMIN6_PMAX30_R25_MEDIAN_BAR_RANGE,
			width:         expectedH48_PMIN6_PMAX30_R25_WIDTH,
			dc:            expectedH48_PMIN6_PMAX30_R25_DOMINANT_CYCLE,
			dcm:           expectedH48_PMIN6_PMAX30_R25_DOMINANT_CYCLE_MEDIAN,
			hasEstimate:   expectedH48_PMIN6_PMAX30_R25_HAS_CYCLE_ESTIMATE,
			ratioIsFinite: expectedH48_PMIN6_PMAX30_R25_RATIO_IS_FINITE,
		},
		{
			name:          "H48_PMIN6_PMAX20_R50",
			hpCutoff:      48,
			minPeriod:     6,
			maxPeriod:     20,
			rasterLength:  50,
			heatmap:       expectedH48_PMIN6_PMAX20_R50_HEATMAP,
			heatmapMin:    expectedH48_PMIN6_PMAX20_R50_HEATMAP_MIN,
			heatmapMax:    expectedH48_PMIN6_PMAX20_R50_HEATMAP_MAX,
			snr:           expectedH48_PMIN6_PMAX20_R50_SIGNAL_TO_NOISE_RATIO,
			physDB:        expectedH48_PMIN6_PMAX20_R50_PHYSICAL_DECIBELS,
			calDB:         expectedH48_PMIN6_PMAX20_R50_CALIBRATED_DECIBELS,
			norm:          expectedH48_PMIN6_PMAX20_R50_NORMALIZED_POSITION,
			signal:        expectedH48_PMIN6_PMAX20_R50_SIGNAL,
			noise:         expectedH48_PMIN6_PMAX20_R50_NOISE,
			medianRange:   expectedH48_PMIN6_PMAX20_R50_MEDIAN_BAR_RANGE,
			width:         expectedH48_PMIN6_PMAX20_R50_WIDTH,
			dc:            expectedH48_PMIN6_PMAX20_R50_DOMINANT_CYCLE,
			dcm:           expectedH48_PMIN6_PMAX20_R50_DOMINANT_CYCLE_MEDIAN,
			hasEstimate:   expectedH48_PMIN6_PMAX20_R50_HAS_CYCLE_ESTIMATE,
			ratioIsFinite: expectedH48_PMIN6_PMAX20_R50_RATIO_IS_FINITE,
		},
		{
			name:          "H48_PMIN6_PMAX20_R25",
			hpCutoff:      48,
			minPeriod:     6,
			maxPeriod:     20,
			rasterLength:  25,
			heatmap:       expectedH48_PMIN6_PMAX20_R25_HEATMAP,
			heatmapMin:    expectedH48_PMIN6_PMAX20_R25_HEATMAP_MIN,
			heatmapMax:    expectedH48_PMIN6_PMAX20_R25_HEATMAP_MAX,
			snr:           expectedH48_PMIN6_PMAX20_R25_SIGNAL_TO_NOISE_RATIO,
			physDB:        expectedH48_PMIN6_PMAX20_R25_PHYSICAL_DECIBELS,
			calDB:         expectedH48_PMIN6_PMAX20_R25_CALIBRATED_DECIBELS,
			norm:          expectedH48_PMIN6_PMAX20_R25_NORMALIZED_POSITION,
			signal:        expectedH48_PMIN6_PMAX20_R25_SIGNAL,
			noise:         expectedH48_PMIN6_PMAX20_R25_NOISE,
			medianRange:   expectedH48_PMIN6_PMAX20_R25_MEDIAN_BAR_RANGE,
			width:         expectedH48_PMIN6_PMAX20_R25_WIDTH,
			dc:            expectedH48_PMIN6_PMAX20_R25_DOMINANT_CYCLE,
			dcm:           expectedH48_PMIN6_PMAX20_R25_DOMINANT_CYCLE_MEDIAN,
			hasEstimate:   expectedH48_PMIN6_PMAX20_R25_HAS_CYCLE_ESTIMATE,
			ratioIsFinite: expectedH48_PMIN6_PMAX20_R25_RATIO_IS_FINITE,
		},
		{
			name:          "H48_PMIN10_PMAX30_R50",
			hpCutoff:      48,
			minPeriod:     10,
			maxPeriod:     30,
			rasterLength:  50,
			heatmap:       expectedH48_PMIN10_PMAX30_R50_HEATMAP,
			heatmapMin:    expectedH48_PMIN10_PMAX30_R50_HEATMAP_MIN,
			heatmapMax:    expectedH48_PMIN10_PMAX30_R50_HEATMAP_MAX,
			snr:           expectedH48_PMIN10_PMAX30_R50_SIGNAL_TO_NOISE_RATIO,
			physDB:        expectedH48_PMIN10_PMAX30_R50_PHYSICAL_DECIBELS,
			calDB:         expectedH48_PMIN10_PMAX30_R50_CALIBRATED_DECIBELS,
			norm:          expectedH48_PMIN10_PMAX30_R50_NORMALIZED_POSITION,
			signal:        expectedH48_PMIN10_PMAX30_R50_SIGNAL,
			noise:         expectedH48_PMIN10_PMAX30_R50_NOISE,
			medianRange:   expectedH48_PMIN10_PMAX30_R50_MEDIAN_BAR_RANGE,
			width:         expectedH48_PMIN10_PMAX30_R50_WIDTH,
			dc:            expectedH48_PMIN10_PMAX30_R50_DOMINANT_CYCLE,
			dcm:           expectedH48_PMIN10_PMAX30_R50_DOMINANT_CYCLE_MEDIAN,
			hasEstimate:   expectedH48_PMIN10_PMAX30_R50_HAS_CYCLE_ESTIMATE,
			ratioIsFinite: expectedH48_PMIN10_PMAX30_R50_RATIO_IS_FINITE,
		},
		{
			name:          "H48_PMIN10_PMAX30_R25",
			hpCutoff:      48,
			minPeriod:     10,
			maxPeriod:     30,
			rasterLength:  25,
			heatmap:       expectedH48_PMIN10_PMAX30_R25_HEATMAP,
			heatmapMin:    expectedH48_PMIN10_PMAX30_R25_HEATMAP_MIN,
			heatmapMax:    expectedH48_PMIN10_PMAX30_R25_HEATMAP_MAX,
			snr:           expectedH48_PMIN10_PMAX30_R25_SIGNAL_TO_NOISE_RATIO,
			physDB:        expectedH48_PMIN10_PMAX30_R25_PHYSICAL_DECIBELS,
			calDB:         expectedH48_PMIN10_PMAX30_R25_CALIBRATED_DECIBELS,
			norm:          expectedH48_PMIN10_PMAX30_R25_NORMALIZED_POSITION,
			signal:        expectedH48_PMIN10_PMAX30_R25_SIGNAL,
			noise:         expectedH48_PMIN10_PMAX30_R25_NOISE,
			medianRange:   expectedH48_PMIN10_PMAX30_R25_MEDIAN_BAR_RANGE,
			width:         expectedH48_PMIN10_PMAX30_R25_WIDTH,
			dc:            expectedH48_PMIN10_PMAX30_R25_DOMINANT_CYCLE,
			dcm:           expectedH48_PMIN10_PMAX30_R25_DOMINANT_CYCLE_MEDIAN,
			hasEstimate:   expectedH48_PMIN10_PMAX30_R25_HAS_CYCLE_ESTIMATE,
			ratioIsFinite: expectedH48_PMIN10_PMAX30_R25_RATIO_IS_FINITE,
		},
		{
			name:          "H48_PMIN10_PMAX48_R50",
			hpCutoff:      48,
			minPeriod:     10,
			maxPeriod:     48,
			rasterLength:  50,
			heatmap:       expectedH48_PMIN10_PMAX48_R50_HEATMAP,
			heatmapMin:    expectedH48_PMIN10_PMAX48_R50_HEATMAP_MIN,
			heatmapMax:    expectedH48_PMIN10_PMAX48_R50_HEATMAP_MAX,
			snr:           expectedH48_PMIN10_PMAX48_R50_SIGNAL_TO_NOISE_RATIO,
			physDB:        expectedH48_PMIN10_PMAX48_R50_PHYSICAL_DECIBELS,
			calDB:         expectedH48_PMIN10_PMAX48_R50_CALIBRATED_DECIBELS,
			norm:          expectedH48_PMIN10_PMAX48_R50_NORMALIZED_POSITION,
			signal:        expectedH48_PMIN10_PMAX48_R50_SIGNAL,
			noise:         expectedH48_PMIN10_PMAX48_R50_NOISE,
			medianRange:   expectedH48_PMIN10_PMAX48_R50_MEDIAN_BAR_RANGE,
			width:         expectedH48_PMIN10_PMAX48_R50_WIDTH,
			dc:            expectedH48_PMIN10_PMAX48_R50_DOMINANT_CYCLE,
			dcm:           expectedH48_PMIN10_PMAX48_R50_DOMINANT_CYCLE_MEDIAN,
			hasEstimate:   expectedH48_PMIN10_PMAX48_R50_HAS_CYCLE_ESTIMATE,
			ratioIsFinite: expectedH48_PMIN10_PMAX48_R50_RATIO_IS_FINITE,
		},
		{
			name:          "H48_PMIN10_PMAX48_R25",
			hpCutoff:      48,
			minPeriod:     10,
			maxPeriod:     48,
			rasterLength:  25,
			heatmap:       expectedH48_PMIN10_PMAX48_R25_HEATMAP,
			heatmapMin:    expectedH48_PMIN10_PMAX48_R25_HEATMAP_MIN,
			heatmapMax:    expectedH48_PMIN10_PMAX48_R25_HEATMAP_MAX,
			snr:           expectedH48_PMIN10_PMAX48_R25_SIGNAL_TO_NOISE_RATIO,
			physDB:        expectedH48_PMIN10_PMAX48_R25_PHYSICAL_DECIBELS,
			calDB:         expectedH48_PMIN10_PMAX48_R25_CALIBRATED_DECIBELS,
			norm:          expectedH48_PMIN10_PMAX48_R25_NORMALIZED_POSITION,
			signal:        expectedH48_PMIN10_PMAX48_R25_SIGNAL,
			noise:         expectedH48_PMIN10_PMAX48_R25_NOISE,
			medianRange:   expectedH48_PMIN10_PMAX48_R25_MEDIAN_BAR_RANGE,
			width:         expectedH48_PMIN10_PMAX48_R25_WIDTH,
			dc:            expectedH48_PMIN10_PMAX48_R25_DOMINANT_CYCLE,
			dcm:           expectedH48_PMIN10_PMAX48_R25_DOMINANT_CYCLE_MEDIAN,
			hasEstimate:   expectedH48_PMIN10_PMAX48_R25_HAS_CYCLE_ESTIMATE,
			ratioIsFinite: expectedH48_PMIN10_PMAX48_R25_RATIO_IS_FINITE,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			params := DefaultParams()
			params.RasterLength = tc.rasterLength
			params.HighPassFilterCutoff = tc.hpCutoff
			params.MinimalPeriod = tc.minPeriod
			params.MaximalPeriod = tc.maxPeriod

			indicator, err := NewCoronaSignalToNoiseRatioParams(params)
			if err != nil {
				t.Fatalf("NewCoronaSignalToNoiseRatioParams: %v", err)
			}
			step := 10.0 / float64(tc.rasterLength)
			expectedFirst := 1.0 + step
			expectedResolution := 1.0 / step
			baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

			for i, price := range testInput {
				barTime := baseTime.AddDate(0, 0, i)
				snap := indicator.Update(price, testLow[i], testHigh[i], barTime)
				assertClose(t, snap.Heatmap.ParameterFirst, expectedFirst, fmt.Sprintf("%s[%d].first", tc.name, i))
				assertClose(t, snap.Heatmap.ParameterLast, 11.0, fmt.Sprintf("%s[%d].last", tc.name, i))
				assertClose(t, snap.Heatmap.ParameterResolution, expectedResolution, fmt.Sprintf("%s[%d].resolution", tc.name, i))

				if i < 12 {
					if indicator.IsPrimed() || !snap.Heatmap.IsEmpty() {
						t.Fatalf("%s[%d]: expected unprimed empty heatmap", tc.name, i)
					}
				} else {
					if !indicator.IsPrimed() || len(snap.Heatmap.Values) != tc.rasterLength {
						t.Fatalf("%s[%d]: expected primed heatmap of len %d", tc.name, i, tc.rasterLength)
					}
					for binIndex := 0; binIndex < tc.rasterLength; binIndex++ {
						assertClose(t, snap.Heatmap.Values[binIndex], tc.heatmap[i][binIndex], fmt.Sprintf("%s[%d].hm[%d]", tc.name, i, binIndex))
					}
				}

				assertClose(t, snap.Heatmap.ValueMin, tc.heatmapMin[i], fmt.Sprintf("%s[%d].min", tc.name, i))
				assertClose(t, snap.Heatmap.ValueMax, tc.heatmapMax[i], fmt.Sprintf("%s[%d].max", tc.name, i))
				assertClose(t, snap.SignalToNoiseRatio, tc.snr[i], fmt.Sprintf("%s[%d].snr", tc.name, i))
				assertClose(t, snap.PhysicalDecibels, tc.physDB[i], fmt.Sprintf("%s[%d].physDB", tc.name, i))
				assertClose(t, snap.CalibratedDecibels, tc.calDB[i], fmt.Sprintf("%s[%d].calDB", tc.name, i))
				assertClose(t, snap.NormalizedPosition, tc.norm[i], fmt.Sprintf("%s[%d].norm", tc.name, i))
				assertClose(t, snap.Signal, tc.signal[i], fmt.Sprintf("%s[%d].signal", tc.name, i))
				assertClose(t, snap.Noise, tc.noise[i], fmt.Sprintf("%s[%d].noise", tc.name, i))
				assertClose(t, snap.MedianBarRange, tc.medianRange[i], fmt.Sprintf("%s[%d].medianRange", tc.name, i))
				assertClose(t, snap.Width, tc.width[i], fmt.Sprintf("%s[%d].width", tc.name, i))
				assertClose(t, snap.DominantCycle, tc.dc[i], fmt.Sprintf("%s[%d].dc", tc.name, i))
				assertClose(t, snap.DominantCycleMedian, tc.dcm[i], fmt.Sprintf("%s[%d].dcm", tc.name, i))
				assertClose(t, snap.HasCycleEstimate, tc.hasEstimate[i], fmt.Sprintf("%s[%d].hasEst", tc.name, i))
				assertClose(t, snap.RatioIsFinite, tc.ratioIsFinite[i], fmt.Sprintf("%s[%d].finite", tc.name, i))
			}

			indicator.Reset()
			if indicator.IsPrimed() {
				t.Fatalf("%s: expected IsPrimed=false after Reset", tc.name)
			}
		})
	}
}

func TestMetadataMnemonicsEntitiesAndValidation(t *testing.T) {
	t.Parallel()

	indicator, err := NewCoronaSignalToNoiseRatioDefault()
	if err != nil {
		t.Fatalf("NewCoronaSignalToNoiseRatioDefault: %v", err)
	}
	meta := indicator.Metadata()
	if meta.Identifier != core.CoronaSignalToNoiseRatio || meta.Mnemonic != "csnr(50, 20, 1, 11, 30)" {
		t.Fatalf("unexpected metadata: %+v", meta)
	}
	if len(meta.Outputs) != 2 || meta.Outputs[0].Shape != shape.Heatmap || meta.Outputs[1].Shape != shape.Scalar {
		t.Fatalf("unexpected outputs: %+v", meta.Outputs)
	}
	if meta.Outputs[1].Mnemonic != "csnr-snr(30)" {
		t.Fatalf("unexpected secondary mnemonic: %q", meta.Outputs[1].Mnemonic)
	}

	negativeParams := DefaultParams()
	negativeParams.MinParameterValue = -5.0
	negativeParams.MaxParameterValue = 5.0
	negativeParams.MinimalPeriod = 10
	negativeParams.MaximalPeriod = 48
	negativeParams.BarComponent = entities.BarMedianPrice
	negativeBounds, err := NewCoronaSignalToNoiseRatioParams(negativeParams)
	if err != nil {
		t.Fatalf("NewCoronaSignalToNoiseRatioParams negative bounds: %v", err)
	}
	if negativeBounds.Metadata().Mnemonic != "csnr(50, 20, -5, 5, 30, pmin=10, pmax=48, hl/2)" {
		t.Fatalf("unexpected negative-bounds mnemonic: %q", negativeBounds.Metadata().Mnemonic)
	}

	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	barInd, _ := NewCoronaSignalToNoiseRatioDefault()
	directInd, _ := NewCoronaSignalToNoiseRatioDefault()
	scalarInd, _ := NewCoronaSignalToNoiseRatioDefault()
	quoteInd, _ := NewCoronaSignalToNoiseRatioDefault()
	tradeInd, _ := NewCoronaSignalToNoiseRatioDefault()
	for i := 0; i < 20; i++ {
		barTime := baseTime.AddDate(0, 0, i)
		price := testInput[i]
		outBar := barInd.UpdateBar(&entities.Bar{Time: barTime, Open: testOpen[i], High: testHigh[i], Low: testLow[i], Close: price, Volume: 1000})
		directSnap := directInd.Update(price, testLow[i], testHigh[i], barTime)
		assertClose(t, outBar[1].(entities.Scalar).Value, directSnap.SignalToNoiseRatio, "bar vs direct")

		outScalar := scalarInd.UpdateScalar(&entities.Scalar{Time: barTime, Value: price})
		outQuote := quoteInd.UpdateQuote(&entities.Quote{Time: barTime, Bid: price - 0.5, Ask: price + 0.5, BidSize: 10, AskSize: 10})
		outTrade := tradeInd.UpdateTrade(&entities.Trade{Time: barTime, Price: price, Volume: 10})
		assertClose(t, outScalar[1].(entities.Scalar).Value, outQuote[1].(entities.Scalar).Value, "scalar vs quote")
		assertClose(t, outScalar[1].(entities.Scalar).Value, outTrade[1].(entities.Scalar).Value, "scalar vs trade")
	}

	invalid := []func(*Params){
		func(p *Params) { p.RasterLength = 1 },
		func(p *Params) { p.RasterLength = 0 },
		func(p *Params) { p.MaxRasterValue = 6.0 },
		func(p *Params) { p.MinParameterValue = 5.0; p.MaxParameterValue = 5.0 },
		func(p *Params) { p.HighPassFilterCutoff = 4 },
		func(p *Params) { p.MinimalPeriod = 5 },
		func(p *Params) { p.MinimalPeriod = 10; p.MaximalPeriod = 10 },
	}
	for i, mutate := range invalid {
		bad := DefaultParams()
		mutate(bad)
		if _, err := NewCoronaSignalToNoiseRatioParams(bad); err == nil {
			t.Fatalf("invalid[%d]: expected error, got nil", i)
		}
	}
}
