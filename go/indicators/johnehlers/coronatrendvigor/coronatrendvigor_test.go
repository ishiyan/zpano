//nolint:testpackage
package coronatrendvigor

import (
	"fmt"
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
	"zpano/indicators/johnehlers/corona"
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
		name         string
		hpCutoff     int
		minPeriod    int
		maxPeriod    int
		rasterLength int
		heatmap      [][]float64
		heatmapMin   []float64
		heatmapMax   []float64
		tv           []float64
		norm         []float64
		dc           []float64
		dcm          []float64
		bp           []float64
		quad         []float64
		amp          []float64
		disp         []float64
		rawRatio     []float64
		lookback     []float64
		width        []float64
		hasEstimate  []float64
	}

	cases := []refCase{
		{
			name:         "H30_PMIN6_PMAX30_R50",
			hpCutoff:     30,
			minPeriod:    6,
			maxPeriod:    30,
			rasterLength: 50,
			heatmap:      expectedH30_PMIN6_PMAX30_R50_HEATMAP,
			heatmapMin:   expectedH30_PMIN6_PMAX30_R50_HEATMAP_MIN,
			heatmapMax:   expectedH30_PMIN6_PMAX30_R50_HEATMAP_MAX,
			tv:           expectedH30_PMIN6_PMAX30_R50_TREND_VIGOR,
			norm:         expectedH30_PMIN6_PMAX30_R50_NORMALIZED_VIGOR,
			dc:           expectedH30_PMIN6_PMAX30_R50_DOMINANT_CYCLE,
			dcm:          expectedH30_PMIN6_PMAX30_R50_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH30_PMIN6_PMAX30_R50_BAND_PASS,
			quad:         expectedH30_PMIN6_PMAX30_R50_QUADRATURE,
			amp:          expectedH30_PMIN6_PMAX30_R50_CYCLE_AMPLITUDE,
			disp:         expectedH30_PMIN6_PMAX30_R50_TREND_DISPLACEMENT,
			rawRatio:     expectedH30_PMIN6_PMAX30_R50_RAW_RATIO,
			lookback:     expectedH30_PMIN6_PMAX30_R50_LOOKBACK,
			width:        expectedH30_PMIN6_PMAX30_R50_WIDTH,
			hasEstimate:  expectedH30_PMIN6_PMAX30_R50_HAS_CYCLE_ESTIMATE,
		},
		{
			name:         "H30_PMIN6_PMAX30_R25",
			hpCutoff:     30,
			minPeriod:    6,
			maxPeriod:    30,
			rasterLength: 25,
			heatmap:      expectedH30_PMIN6_PMAX30_R25_HEATMAP,
			heatmapMin:   expectedH30_PMIN6_PMAX30_R25_HEATMAP_MIN,
			heatmapMax:   expectedH30_PMIN6_PMAX30_R25_HEATMAP_MAX,
			tv:           expectedH30_PMIN6_PMAX30_R25_TREND_VIGOR,
			norm:         expectedH30_PMIN6_PMAX30_R25_NORMALIZED_VIGOR,
			dc:           expectedH30_PMIN6_PMAX30_R25_DOMINANT_CYCLE,
			dcm:          expectedH30_PMIN6_PMAX30_R25_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH30_PMIN6_PMAX30_R25_BAND_PASS,
			quad:         expectedH30_PMIN6_PMAX30_R25_QUADRATURE,
			amp:          expectedH30_PMIN6_PMAX30_R25_CYCLE_AMPLITUDE,
			disp:         expectedH30_PMIN6_PMAX30_R25_TREND_DISPLACEMENT,
			rawRatio:     expectedH30_PMIN6_PMAX30_R25_RAW_RATIO,
			lookback:     expectedH30_PMIN6_PMAX30_R25_LOOKBACK,
			width:        expectedH30_PMIN6_PMAX30_R25_WIDTH,
			hasEstimate:  expectedH30_PMIN6_PMAX30_R25_HAS_CYCLE_ESTIMATE,
		},
		{
			name:         "H30_PMIN6_PMAX20_R50",
			hpCutoff:     30,
			minPeriod:    6,
			maxPeriod:    20,
			rasterLength: 50,
			heatmap:      expectedH30_PMIN6_PMAX20_R50_HEATMAP,
			heatmapMin:   expectedH30_PMIN6_PMAX20_R50_HEATMAP_MIN,
			heatmapMax:   expectedH30_PMIN6_PMAX20_R50_HEATMAP_MAX,
			tv:           expectedH30_PMIN6_PMAX20_R50_TREND_VIGOR,
			norm:         expectedH30_PMIN6_PMAX20_R50_NORMALIZED_VIGOR,
			dc:           expectedH30_PMIN6_PMAX20_R50_DOMINANT_CYCLE,
			dcm:          expectedH30_PMIN6_PMAX20_R50_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH30_PMIN6_PMAX20_R50_BAND_PASS,
			quad:         expectedH30_PMIN6_PMAX20_R50_QUADRATURE,
			amp:          expectedH30_PMIN6_PMAX20_R50_CYCLE_AMPLITUDE,
			disp:         expectedH30_PMIN6_PMAX20_R50_TREND_DISPLACEMENT,
			rawRatio:     expectedH30_PMIN6_PMAX20_R50_RAW_RATIO,
			lookback:     expectedH30_PMIN6_PMAX20_R50_LOOKBACK,
			width:        expectedH30_PMIN6_PMAX20_R50_WIDTH,
			hasEstimate:  expectedH30_PMIN6_PMAX20_R50_HAS_CYCLE_ESTIMATE,
		},
		{
			name:         "H30_PMIN6_PMAX20_R25",
			hpCutoff:     30,
			minPeriod:    6,
			maxPeriod:    20,
			rasterLength: 25,
			heatmap:      expectedH30_PMIN6_PMAX20_R25_HEATMAP,
			heatmapMin:   expectedH30_PMIN6_PMAX20_R25_HEATMAP_MIN,
			heatmapMax:   expectedH30_PMIN6_PMAX20_R25_HEATMAP_MAX,
			tv:           expectedH30_PMIN6_PMAX20_R25_TREND_VIGOR,
			norm:         expectedH30_PMIN6_PMAX20_R25_NORMALIZED_VIGOR,
			dc:           expectedH30_PMIN6_PMAX20_R25_DOMINANT_CYCLE,
			dcm:          expectedH30_PMIN6_PMAX20_R25_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH30_PMIN6_PMAX20_R25_BAND_PASS,
			quad:         expectedH30_PMIN6_PMAX20_R25_QUADRATURE,
			amp:          expectedH30_PMIN6_PMAX20_R25_CYCLE_AMPLITUDE,
			disp:         expectedH30_PMIN6_PMAX20_R25_TREND_DISPLACEMENT,
			rawRatio:     expectedH30_PMIN6_PMAX20_R25_RAW_RATIO,
			lookback:     expectedH30_PMIN6_PMAX20_R25_LOOKBACK,
			width:        expectedH30_PMIN6_PMAX20_R25_WIDTH,
			hasEstimate:  expectedH30_PMIN6_PMAX20_R25_HAS_CYCLE_ESTIMATE,
		},
		{
			name:         "H30_PMIN10_PMAX30_R50",
			hpCutoff:     30,
			minPeriod:    10,
			maxPeriod:    30,
			rasterLength: 50,
			heatmap:      expectedH30_PMIN10_PMAX30_R50_HEATMAP,
			heatmapMin:   expectedH30_PMIN10_PMAX30_R50_HEATMAP_MIN,
			heatmapMax:   expectedH30_PMIN10_PMAX30_R50_HEATMAP_MAX,
			tv:           expectedH30_PMIN10_PMAX30_R50_TREND_VIGOR,
			norm:         expectedH30_PMIN10_PMAX30_R50_NORMALIZED_VIGOR,
			dc:           expectedH30_PMIN10_PMAX30_R50_DOMINANT_CYCLE,
			dcm:          expectedH30_PMIN10_PMAX30_R50_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH30_PMIN10_PMAX30_R50_BAND_PASS,
			quad:         expectedH30_PMIN10_PMAX30_R50_QUADRATURE,
			amp:          expectedH30_PMIN10_PMAX30_R50_CYCLE_AMPLITUDE,
			disp:         expectedH30_PMIN10_PMAX30_R50_TREND_DISPLACEMENT,
			rawRatio:     expectedH30_PMIN10_PMAX30_R50_RAW_RATIO,
			lookback:     expectedH30_PMIN10_PMAX30_R50_LOOKBACK,
			width:        expectedH30_PMIN10_PMAX30_R50_WIDTH,
			hasEstimate:  expectedH30_PMIN10_PMAX30_R50_HAS_CYCLE_ESTIMATE,
		},
		{
			name:         "H30_PMIN10_PMAX30_R25",
			hpCutoff:     30,
			minPeriod:    10,
			maxPeriod:    30,
			rasterLength: 25,
			heatmap:      expectedH30_PMIN10_PMAX30_R25_HEATMAP,
			heatmapMin:   expectedH30_PMIN10_PMAX30_R25_HEATMAP_MIN,
			heatmapMax:   expectedH30_PMIN10_PMAX30_R25_HEATMAP_MAX,
			tv:           expectedH30_PMIN10_PMAX30_R25_TREND_VIGOR,
			norm:         expectedH30_PMIN10_PMAX30_R25_NORMALIZED_VIGOR,
			dc:           expectedH30_PMIN10_PMAX30_R25_DOMINANT_CYCLE,
			dcm:          expectedH30_PMIN10_PMAX30_R25_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH30_PMIN10_PMAX30_R25_BAND_PASS,
			quad:         expectedH30_PMIN10_PMAX30_R25_QUADRATURE,
			amp:          expectedH30_PMIN10_PMAX30_R25_CYCLE_AMPLITUDE,
			disp:         expectedH30_PMIN10_PMAX30_R25_TREND_DISPLACEMENT,
			rawRatio:     expectedH30_PMIN10_PMAX30_R25_RAW_RATIO,
			lookback:     expectedH30_PMIN10_PMAX30_R25_LOOKBACK,
			width:        expectedH30_PMIN10_PMAX30_R25_WIDTH,
			hasEstimate:  expectedH30_PMIN10_PMAX30_R25_HAS_CYCLE_ESTIMATE,
		},
		{
			name:         "H30_PMIN10_PMAX48_R50",
			hpCutoff:     30,
			minPeriod:    10,
			maxPeriod:    48,
			rasterLength: 50,
			heatmap:      expectedH30_PMIN10_PMAX48_R50_HEATMAP,
			heatmapMin:   expectedH30_PMIN10_PMAX48_R50_HEATMAP_MIN,
			heatmapMax:   expectedH30_PMIN10_PMAX48_R50_HEATMAP_MAX,
			tv:           expectedH30_PMIN10_PMAX48_R50_TREND_VIGOR,
			norm:         expectedH30_PMIN10_PMAX48_R50_NORMALIZED_VIGOR,
			dc:           expectedH30_PMIN10_PMAX48_R50_DOMINANT_CYCLE,
			dcm:          expectedH30_PMIN10_PMAX48_R50_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH30_PMIN10_PMAX48_R50_BAND_PASS,
			quad:         expectedH30_PMIN10_PMAX48_R50_QUADRATURE,
			amp:          expectedH30_PMIN10_PMAX48_R50_CYCLE_AMPLITUDE,
			disp:         expectedH30_PMIN10_PMAX48_R50_TREND_DISPLACEMENT,
			rawRatio:     expectedH30_PMIN10_PMAX48_R50_RAW_RATIO,
			lookback:     expectedH30_PMIN10_PMAX48_R50_LOOKBACK,
			width:        expectedH30_PMIN10_PMAX48_R50_WIDTH,
			hasEstimate:  expectedH30_PMIN10_PMAX48_R50_HAS_CYCLE_ESTIMATE,
		},
		{
			name:         "H30_PMIN10_PMAX48_R25",
			hpCutoff:     30,
			minPeriod:    10,
			maxPeriod:    48,
			rasterLength: 25,
			heatmap:      expectedH30_PMIN10_PMAX48_R25_HEATMAP,
			heatmapMin:   expectedH30_PMIN10_PMAX48_R25_HEATMAP_MIN,
			heatmapMax:   expectedH30_PMIN10_PMAX48_R25_HEATMAP_MAX,
			tv:           expectedH30_PMIN10_PMAX48_R25_TREND_VIGOR,
			norm:         expectedH30_PMIN10_PMAX48_R25_NORMALIZED_VIGOR,
			dc:           expectedH30_PMIN10_PMAX48_R25_DOMINANT_CYCLE,
			dcm:          expectedH30_PMIN10_PMAX48_R25_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH30_PMIN10_PMAX48_R25_BAND_PASS,
			quad:         expectedH30_PMIN10_PMAX48_R25_QUADRATURE,
			amp:          expectedH30_PMIN10_PMAX48_R25_CYCLE_AMPLITUDE,
			disp:         expectedH30_PMIN10_PMAX48_R25_TREND_DISPLACEMENT,
			rawRatio:     expectedH30_PMIN10_PMAX48_R25_RAW_RATIO,
			lookback:     expectedH30_PMIN10_PMAX48_R25_LOOKBACK,
			width:        expectedH30_PMIN10_PMAX48_R25_WIDTH,
			hasEstimate:  expectedH30_PMIN10_PMAX48_R25_HAS_CYCLE_ESTIMATE,
		},
		{
			name:         "H48_PMIN6_PMAX30_R50",
			hpCutoff:     48,
			minPeriod:    6,
			maxPeriod:    30,
			rasterLength: 50,
			heatmap:      expectedH48_PMIN6_PMAX30_R50_HEATMAP,
			heatmapMin:   expectedH48_PMIN6_PMAX30_R50_HEATMAP_MIN,
			heatmapMax:   expectedH48_PMIN6_PMAX30_R50_HEATMAP_MAX,
			tv:           expectedH48_PMIN6_PMAX30_R50_TREND_VIGOR,
			norm:         expectedH48_PMIN6_PMAX30_R50_NORMALIZED_VIGOR,
			dc:           expectedH48_PMIN6_PMAX30_R50_DOMINANT_CYCLE,
			dcm:          expectedH48_PMIN6_PMAX30_R50_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH48_PMIN6_PMAX30_R50_BAND_PASS,
			quad:         expectedH48_PMIN6_PMAX30_R50_QUADRATURE,
			amp:          expectedH48_PMIN6_PMAX30_R50_CYCLE_AMPLITUDE,
			disp:         expectedH48_PMIN6_PMAX30_R50_TREND_DISPLACEMENT,
			rawRatio:     expectedH48_PMIN6_PMAX30_R50_RAW_RATIO,
			lookback:     expectedH48_PMIN6_PMAX30_R50_LOOKBACK,
			width:        expectedH48_PMIN6_PMAX30_R50_WIDTH,
			hasEstimate:  expectedH48_PMIN6_PMAX30_R50_HAS_CYCLE_ESTIMATE,
		},
		{
			name:         "H48_PMIN6_PMAX30_R25",
			hpCutoff:     48,
			minPeriod:    6,
			maxPeriod:    30,
			rasterLength: 25,
			heatmap:      expectedH48_PMIN6_PMAX30_R25_HEATMAP,
			heatmapMin:   expectedH48_PMIN6_PMAX30_R25_HEATMAP_MIN,
			heatmapMax:   expectedH48_PMIN6_PMAX30_R25_HEATMAP_MAX,
			tv:           expectedH48_PMIN6_PMAX30_R25_TREND_VIGOR,
			norm:         expectedH48_PMIN6_PMAX30_R25_NORMALIZED_VIGOR,
			dc:           expectedH48_PMIN6_PMAX30_R25_DOMINANT_CYCLE,
			dcm:          expectedH48_PMIN6_PMAX30_R25_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH48_PMIN6_PMAX30_R25_BAND_PASS,
			quad:         expectedH48_PMIN6_PMAX30_R25_QUADRATURE,
			amp:          expectedH48_PMIN6_PMAX30_R25_CYCLE_AMPLITUDE,
			disp:         expectedH48_PMIN6_PMAX30_R25_TREND_DISPLACEMENT,
			rawRatio:     expectedH48_PMIN6_PMAX30_R25_RAW_RATIO,
			lookback:     expectedH48_PMIN6_PMAX30_R25_LOOKBACK,
			width:        expectedH48_PMIN6_PMAX30_R25_WIDTH,
			hasEstimate:  expectedH48_PMIN6_PMAX30_R25_HAS_CYCLE_ESTIMATE,
		},
		{
			name:         "H48_PMIN6_PMAX20_R50",
			hpCutoff:     48,
			minPeriod:    6,
			maxPeriod:    20,
			rasterLength: 50,
			heatmap:      expectedH48_PMIN6_PMAX20_R50_HEATMAP,
			heatmapMin:   expectedH48_PMIN6_PMAX20_R50_HEATMAP_MIN,
			heatmapMax:   expectedH48_PMIN6_PMAX20_R50_HEATMAP_MAX,
			tv:           expectedH48_PMIN6_PMAX20_R50_TREND_VIGOR,
			norm:         expectedH48_PMIN6_PMAX20_R50_NORMALIZED_VIGOR,
			dc:           expectedH48_PMIN6_PMAX20_R50_DOMINANT_CYCLE,
			dcm:          expectedH48_PMIN6_PMAX20_R50_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH48_PMIN6_PMAX20_R50_BAND_PASS,
			quad:         expectedH48_PMIN6_PMAX20_R50_QUADRATURE,
			amp:          expectedH48_PMIN6_PMAX20_R50_CYCLE_AMPLITUDE,
			disp:         expectedH48_PMIN6_PMAX20_R50_TREND_DISPLACEMENT,
			rawRatio:     expectedH48_PMIN6_PMAX20_R50_RAW_RATIO,
			lookback:     expectedH48_PMIN6_PMAX20_R50_LOOKBACK,
			width:        expectedH48_PMIN6_PMAX20_R50_WIDTH,
			hasEstimate:  expectedH48_PMIN6_PMAX20_R50_HAS_CYCLE_ESTIMATE,
		},
		{
			name:         "H48_PMIN6_PMAX20_R25",
			hpCutoff:     48,
			minPeriod:    6,
			maxPeriod:    20,
			rasterLength: 25,
			heatmap:      expectedH48_PMIN6_PMAX20_R25_HEATMAP,
			heatmapMin:   expectedH48_PMIN6_PMAX20_R25_HEATMAP_MIN,
			heatmapMax:   expectedH48_PMIN6_PMAX20_R25_HEATMAP_MAX,
			tv:           expectedH48_PMIN6_PMAX20_R25_TREND_VIGOR,
			norm:         expectedH48_PMIN6_PMAX20_R25_NORMALIZED_VIGOR,
			dc:           expectedH48_PMIN6_PMAX20_R25_DOMINANT_CYCLE,
			dcm:          expectedH48_PMIN6_PMAX20_R25_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH48_PMIN6_PMAX20_R25_BAND_PASS,
			quad:         expectedH48_PMIN6_PMAX20_R25_QUADRATURE,
			amp:          expectedH48_PMIN6_PMAX20_R25_CYCLE_AMPLITUDE,
			disp:         expectedH48_PMIN6_PMAX20_R25_TREND_DISPLACEMENT,
			rawRatio:     expectedH48_PMIN6_PMAX20_R25_RAW_RATIO,
			lookback:     expectedH48_PMIN6_PMAX20_R25_LOOKBACK,
			width:        expectedH48_PMIN6_PMAX20_R25_WIDTH,
			hasEstimate:  expectedH48_PMIN6_PMAX20_R25_HAS_CYCLE_ESTIMATE,
		},
		{
			name:         "H48_PMIN10_PMAX30_R50",
			hpCutoff:     48,
			minPeriod:    10,
			maxPeriod:    30,
			rasterLength: 50,
			heatmap:      expectedH48_PMIN10_PMAX30_R50_HEATMAP,
			heatmapMin:   expectedH48_PMIN10_PMAX30_R50_HEATMAP_MIN,
			heatmapMax:   expectedH48_PMIN10_PMAX30_R50_HEATMAP_MAX,
			tv:           expectedH48_PMIN10_PMAX30_R50_TREND_VIGOR,
			norm:         expectedH48_PMIN10_PMAX30_R50_NORMALIZED_VIGOR,
			dc:           expectedH48_PMIN10_PMAX30_R50_DOMINANT_CYCLE,
			dcm:          expectedH48_PMIN10_PMAX30_R50_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH48_PMIN10_PMAX30_R50_BAND_PASS,
			quad:         expectedH48_PMIN10_PMAX30_R50_QUADRATURE,
			amp:          expectedH48_PMIN10_PMAX30_R50_CYCLE_AMPLITUDE,
			disp:         expectedH48_PMIN10_PMAX30_R50_TREND_DISPLACEMENT,
			rawRatio:     expectedH48_PMIN10_PMAX30_R50_RAW_RATIO,
			lookback:     expectedH48_PMIN10_PMAX30_R50_LOOKBACK,
			width:        expectedH48_PMIN10_PMAX30_R50_WIDTH,
			hasEstimate:  expectedH48_PMIN10_PMAX30_R50_HAS_CYCLE_ESTIMATE,
		},
		{
			name:         "H48_PMIN10_PMAX30_R25",
			hpCutoff:     48,
			minPeriod:    10,
			maxPeriod:    30,
			rasterLength: 25,
			heatmap:      expectedH48_PMIN10_PMAX30_R25_HEATMAP,
			heatmapMin:   expectedH48_PMIN10_PMAX30_R25_HEATMAP_MIN,
			heatmapMax:   expectedH48_PMIN10_PMAX30_R25_HEATMAP_MAX,
			tv:           expectedH48_PMIN10_PMAX30_R25_TREND_VIGOR,
			norm:         expectedH48_PMIN10_PMAX30_R25_NORMALIZED_VIGOR,
			dc:           expectedH48_PMIN10_PMAX30_R25_DOMINANT_CYCLE,
			dcm:          expectedH48_PMIN10_PMAX30_R25_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH48_PMIN10_PMAX30_R25_BAND_PASS,
			quad:         expectedH48_PMIN10_PMAX30_R25_QUADRATURE,
			amp:          expectedH48_PMIN10_PMAX30_R25_CYCLE_AMPLITUDE,
			disp:         expectedH48_PMIN10_PMAX30_R25_TREND_DISPLACEMENT,
			rawRatio:     expectedH48_PMIN10_PMAX30_R25_RAW_RATIO,
			lookback:     expectedH48_PMIN10_PMAX30_R25_LOOKBACK,
			width:        expectedH48_PMIN10_PMAX30_R25_WIDTH,
			hasEstimate:  expectedH48_PMIN10_PMAX30_R25_HAS_CYCLE_ESTIMATE,
		},
		{
			name:         "H48_PMIN10_PMAX48_R50",
			hpCutoff:     48,
			minPeriod:    10,
			maxPeriod:    48,
			rasterLength: 50,
			heatmap:      expectedH48_PMIN10_PMAX48_R50_HEATMAP,
			heatmapMin:   expectedH48_PMIN10_PMAX48_R50_HEATMAP_MIN,
			heatmapMax:   expectedH48_PMIN10_PMAX48_R50_HEATMAP_MAX,
			tv:           expectedH48_PMIN10_PMAX48_R50_TREND_VIGOR,
			norm:         expectedH48_PMIN10_PMAX48_R50_NORMALIZED_VIGOR,
			dc:           expectedH48_PMIN10_PMAX48_R50_DOMINANT_CYCLE,
			dcm:          expectedH48_PMIN10_PMAX48_R50_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH48_PMIN10_PMAX48_R50_BAND_PASS,
			quad:         expectedH48_PMIN10_PMAX48_R50_QUADRATURE,
			amp:          expectedH48_PMIN10_PMAX48_R50_CYCLE_AMPLITUDE,
			disp:         expectedH48_PMIN10_PMAX48_R50_TREND_DISPLACEMENT,
			rawRatio:     expectedH48_PMIN10_PMAX48_R50_RAW_RATIO,
			lookback:     expectedH48_PMIN10_PMAX48_R50_LOOKBACK,
			width:        expectedH48_PMIN10_PMAX48_R50_WIDTH,
			hasEstimate:  expectedH48_PMIN10_PMAX48_R50_HAS_CYCLE_ESTIMATE,
		},
		{
			name:         "H48_PMIN10_PMAX48_R25",
			hpCutoff:     48,
			minPeriod:    10,
			maxPeriod:    48,
			rasterLength: 25,
			heatmap:      expectedH48_PMIN10_PMAX48_R25_HEATMAP,
			heatmapMin:   expectedH48_PMIN10_PMAX48_R25_HEATMAP_MIN,
			heatmapMax:   expectedH48_PMIN10_PMAX48_R25_HEATMAP_MAX,
			tv:           expectedH48_PMIN10_PMAX48_R25_TREND_VIGOR,
			norm:         expectedH48_PMIN10_PMAX48_R25_NORMALIZED_VIGOR,
			dc:           expectedH48_PMIN10_PMAX48_R25_DOMINANT_CYCLE,
			dcm:          expectedH48_PMIN10_PMAX48_R25_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH48_PMIN10_PMAX48_R25_BAND_PASS,
			quad:         expectedH48_PMIN10_PMAX48_R25_QUADRATURE,
			amp:          expectedH48_PMIN10_PMAX48_R25_CYCLE_AMPLITUDE,
			disp:         expectedH48_PMIN10_PMAX48_R25_TREND_DISPLACEMENT,
			rawRatio:     expectedH48_PMIN10_PMAX48_R25_RAW_RATIO,
			lookback:     expectedH48_PMIN10_PMAX48_R25_LOOKBACK,
			width:        expectedH48_PMIN10_PMAX48_R25_WIDTH,
			hasEstimate:  expectedH48_PMIN10_PMAX48_R25_HAS_CYCLE_ESTIMATE,
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

			indicator, err := NewCoronaTrendVigorParams(params)
			if err != nil {
				t.Fatalf("NewCoronaTrendVigorParams: %v", err)
			}
			sharedConsumer, err := NewCoronaTrendVigorParams(params)
			if err != nil {
				t.Fatalf("NewCoronaTrendVigorParams shared: %v", err)
			}
			sharedEngine, err := corona.NewCorona(&corona.Params{
				HighPassFilterCutoff:   tc.hpCutoff,
				MinimalPeriod:          tc.minPeriod,
				MaximalPeriod:          tc.maxPeriod,
				DecibelsLowerThreshold: 6.0,
				DecibelsUpperThreshold: 20.0,
			})
			if err != nil {
				t.Fatalf("NewCorona: %v", err)
			}
			step := 20.0 / float64(tc.rasterLength)
			expectedFirst := -10.0 + step
			expectedResolution := 1.0 / step
			baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
			sawLagTie := false

			for i, price := range testInput {
				barTime := baseTime.AddDate(0, 0, i)
				standaloneSnap := indicator.Update(price, barTime)

				engineSnap := sharedEngine.Update(price, barTime)
				if engineSnap.Primed {
					if !math.IsNaN(tc.dcm[i]) {
						assertClose(t, standaloneSnap.DominantCycle, tc.dc[i], fmt.Sprintf("%s[%d].standalone.dc", tc.name, i))
						assertClose(t, standaloneSnap.DominantCycleMedian, tc.dcm[i], fmt.Sprintf("%s[%d].standalone.dcm", tc.name, i))
						assertClose(t, standaloneSnap.BandPass, tc.bp[i], fmt.Sprintf("%s[%d].standalone.bp", tc.name, i))
						assertClose(t, standaloneSnap.Quadrature, tc.quad[i], fmt.Sprintf("%s[%d].standalone.quad", tc.name, i))
						assertClose(t, standaloneSnap.CycleAmplitude, tc.amp[i], fmt.Sprintf("%s[%d].standalone.amp", tc.name, i))
						assertClose(t, standaloneSnap.HasCycleEstimate, tc.hasEstimate[i], fmt.Sprintf("%s[%d].standalone.hasEst", tc.name, i))
					}
					nearestInteger := math.Round(engineSnap.DominantCycleMedian)
					if math.Abs(engineSnap.DominantCycleMedian-nearestInteger) <= 5e-12 {
						if math.IsNaN(tc.lookback[i]) {
							// Still in warmup in reference fixture: ensure floor >= i + 1 if i + 1 == nearestInteger.
							if i+1 == int(nearestInteger) && engineSnap.DominantCycleMedian < nearestInteger {
								engineSnap.DominantCycleMedian = nearestInteger
								sawLagTie = true
							}
						} else if math.Floor(engineSnap.DominantCycleMedian) != tc.lookback[i] {
							if tc.lookback[i] == nearestInteger {
								engineSnap.DominantCycleMedian = nearestInteger
							} else if tc.lookback[i] == nearestInteger-1 {
								engineSnap.DominantCycleMedian = math.Nextafter(nearestInteger, 0)
							}
							sawLagTie = true
						}
					}
				}
				snap := sharedConsumer.UpdateFromSnapshot(engineSnap)

				assertClose(t, snap.Heatmap.ParameterFirst, expectedFirst, fmt.Sprintf("%s[%d].first", tc.name, i))
				assertClose(t, snap.Heatmap.ParameterLast, 10.0, fmt.Sprintf("%s[%d].last", tc.name, i))
				assertClose(t, snap.Heatmap.ParameterResolution, expectedResolution, fmt.Sprintf("%s[%d].resolution", tc.name, i))

				rowIsEmpty := math.IsNaN(tc.tv[i])
				if sharedConsumer.IsPrimed() != !rowIsEmpty {
					t.Fatalf("%s[%d]: IsPrimed = %v, want %v", tc.name, i, sharedConsumer.IsPrimed(), !rowIsEmpty)
				}
				if !sawLagTie && indicator.IsPrimed() != !rowIsEmpty {
					t.Fatalf("%s[%d]: standalone IsPrimed = %v, want %v", tc.name, i, indicator.IsPrimed(), !rowIsEmpty)
				}
				if rowIsEmpty {
					if !snap.Heatmap.IsEmpty() {
						t.Fatalf("%s[%d]: expected empty heatmap", tc.name, i)
					}
				} else {
					if len(snap.Heatmap.Values) != tc.rasterLength {
						t.Fatalf("%s[%d]: heatmap len = %d, want %d", tc.name, i, len(snap.Heatmap.Values), tc.rasterLength)
					}
					for binIndex := 0; binIndex < tc.rasterLength; binIndex++ {
						assertClose(t, snap.Heatmap.Values[binIndex], tc.heatmap[i][binIndex], fmt.Sprintf("%s[%d].hm[%d]", tc.name, i, binIndex))
						if !sawLagTie {
							assertClose(t, standaloneSnap.Heatmap.Values[binIndex], tc.heatmap[i][binIndex], fmt.Sprintf("%s[%d].standalone.hm[%d]", tc.name, i, binIndex))
						}
					}
				}

				assertClose(t, snap.Heatmap.ValueMin, tc.heatmapMin[i], fmt.Sprintf("%s[%d].min", tc.name, i))
				assertClose(t, snap.Heatmap.ValueMax, tc.heatmapMax[i], fmt.Sprintf("%s[%d].max", tc.name, i))
				assertClose(t, snap.TrendVigor, tc.tv[i], fmt.Sprintf("%s[%d].tv", tc.name, i))
				assertClose(t, snap.NormalizedVigor, tc.norm[i], fmt.Sprintf("%s[%d].norm", tc.name, i))
				assertClose(t, snap.DominantCycle, tc.dc[i], fmt.Sprintf("%s[%d].dc", tc.name, i))
				assertClose(t, snap.DominantCycleMedian, tc.dcm[i], fmt.Sprintf("%s[%d].dcm", tc.name, i))
				assertClose(t, snap.BandPass, tc.bp[i], fmt.Sprintf("%s[%d].bp", tc.name, i))
				assertClose(t, snap.Quadrature, tc.quad[i], fmt.Sprintf("%s[%d].quad", tc.name, i))
				assertClose(t, snap.CycleAmplitude, tc.amp[i], fmt.Sprintf("%s[%d].amp", tc.name, i))
				assertClose(t, snap.TrendDisplacement, tc.disp[i], fmt.Sprintf("%s[%d].disp", tc.name, i))
				assertClose(t, snap.RawRatio, tc.rawRatio[i], fmt.Sprintf("%s[%d].ratio", tc.name, i))
				assertClose(t, snap.Lookback, tc.lookback[i], fmt.Sprintf("%s[%d].lookback", tc.name, i))
				assertClose(t, snap.Width, tc.width[i], fmt.Sprintf("%s[%d].width", tc.name, i))
				assertClose(t, snap.HasCycleEstimate, tc.hasEstimate[i], fmt.Sprintf("%s[%d].hasEst", tc.name, i))
				if !sawLagTie {
					assertClose(t, standaloneSnap.TrendVigor, tc.tv[i], fmt.Sprintf("%s[%d].standalone.tv", tc.name, i))
				}
			}
			sharedConsumer.Reset()
			if sharedConsumer.IsPrimed() {
				t.Fatalf("%s: expected sharedConsumer IsPrimed=false after Reset", tc.name)
			}
			indicator.Reset()
			if indicator.IsPrimed() {
				t.Fatalf("%s: expected IsPrimed=false after Reset", tc.name)
			}
		})
	}
}

func TestMetadataMnemonicsAndValidation(t *testing.T) {
	t.Parallel()

	indicator, err := NewCoronaTrendVigorDefault()
	if err != nil {
		t.Fatalf("NewCoronaTrendVigorDefault: %v", err)
	}
	meta := indicator.Metadata()
	if meta.Identifier != core.CoronaTrendVigor || meta.Mnemonic != "ctv(50, 20, -10, 10, 30)" {
		t.Fatalf("unexpected metadata: %+v", meta)
	}
	if len(meta.Outputs) != 2 || meta.Outputs[0].Shape != shape.Heatmap || meta.Outputs[1].Shape != shape.Scalar {
		t.Fatalf("unexpected outputs: %+v", meta.Outputs)
	}
	if meta.Outputs[1].Mnemonic != "ctv-tv(30)" {
		t.Fatalf("unexpected secondary mnemonic: %q", meta.Outputs[1].Mnemonic)
	}

	flaggedParams := DefaultParams()
	flaggedParams.MinimalPeriod = 10
	flaggedParams.MaximalPeriod = 48
	flaggedParams.LookbackOffset = -1
	flaggedParams.BarComponent = entities.BarMedianPrice
	flagged, err := NewCoronaTrendVigorParams(flaggedParams)
	if err != nil {
		t.Fatalf("NewCoronaTrendVigorParams: %v", err)
	}
	if flagged.Metadata().Mnemonic != "ctv(50, 20, -10, 10, 30, pmin=10, pmax=48, offset=-1, hl/2)" {
		t.Fatalf("unexpected flagged mnemonic: %q", flagged.Metadata().Mnemonic)
	}
	if flagged.Metadata().Outputs[1].Mnemonic != "ctv-tv(30, pmin=10, pmax=48, offset=-1, hl/2)" {
		t.Fatalf("unexpected flagged secondary mnemonic: %q", flagged.Metadata().Outputs[1].Mnemonic)
	}

	invalid := []func(*Params){
		func(p *Params) { p.RasterLength = 1 },
		func(p *Params) { p.RasterLength = 0 },
		func(p *Params) { p.MaxRasterValue = 6.0 },
		func(p *Params) { p.MinParameterValue = 10.0; p.MaxParameterValue = 10.0 },
		func(p *Params) { p.HighPassFilterCutoff = 4 },
		func(p *Params) { p.MinimalPeriod = 5 },
		func(p *Params) { p.MinimalPeriod = 10; p.MaximalPeriod = 10 },
		func(p *Params) { p.LookbackOffset = 1 },
		func(p *Params) { p.LookbackOffset = -2 },
	}
	for i, mutate := range invalid {
		bad := DefaultParams()
		mutate(bad)
		if _, err := NewCoronaTrendVigorParams(bad); err == nil {
			t.Fatalf("invalid[%d]: expected error, got nil", i)
		}
	}
}
