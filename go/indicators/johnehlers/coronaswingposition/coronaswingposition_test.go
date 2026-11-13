//nolint:testpackage
package coronaswingposition

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
		name         string
		hpCutoff     int
		minPeriod    int
		maxPeriod    int
		rasterLength int
		heatmap      [][]float64
		heatmapMin   []float64
		heatmapMax   []float64
		sp           []float64
		norm         []float64
		dc           []float64
		dcm          []float64
		bp           []float64
		quad         []float64
		amp          []float64
		lead60       []float64
		span         []float64
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
			sp:           expectedH30_PMIN6_PMAX30_R50_SWING_POSITION,
			norm:         expectedH30_PMIN6_PMAX30_R50_NORMALIZED_POSITION,
			dc:           expectedH30_PMIN6_PMAX30_R50_DOMINANT_CYCLE,
			dcm:          expectedH30_PMIN6_PMAX30_R50_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH30_PMIN6_PMAX30_R50_BAND_PASS,
			quad:         expectedH30_PMIN6_PMAX30_R50_QUADRATURE,
			amp:          expectedH30_PMIN6_PMAX30_R50_CYCLE_AMPLITUDE,
			lead60:       expectedH30_PMIN6_PMAX30_R50_LEAD60,
			span:         expectedH30_PMIN6_PMAX30_R50_POSITION_SPAN,
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
			sp:           expectedH30_PMIN6_PMAX30_R25_SWING_POSITION,
			norm:         expectedH30_PMIN6_PMAX30_R25_NORMALIZED_POSITION,
			dc:           expectedH30_PMIN6_PMAX30_R25_DOMINANT_CYCLE,
			dcm:          expectedH30_PMIN6_PMAX30_R25_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH30_PMIN6_PMAX30_R25_BAND_PASS,
			quad:         expectedH30_PMIN6_PMAX30_R25_QUADRATURE,
			amp:          expectedH30_PMIN6_PMAX30_R25_CYCLE_AMPLITUDE,
			lead60:       expectedH30_PMIN6_PMAX30_R25_LEAD60,
			span:         expectedH30_PMIN6_PMAX30_R25_POSITION_SPAN,
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
			sp:           expectedH30_PMIN6_PMAX20_R50_SWING_POSITION,
			norm:         expectedH30_PMIN6_PMAX20_R50_NORMALIZED_POSITION,
			dc:           expectedH30_PMIN6_PMAX20_R50_DOMINANT_CYCLE,
			dcm:          expectedH30_PMIN6_PMAX20_R50_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH30_PMIN6_PMAX20_R50_BAND_PASS,
			quad:         expectedH30_PMIN6_PMAX20_R50_QUADRATURE,
			amp:          expectedH30_PMIN6_PMAX20_R50_CYCLE_AMPLITUDE,
			lead60:       expectedH30_PMIN6_PMAX20_R50_LEAD60,
			span:         expectedH30_PMIN6_PMAX20_R50_POSITION_SPAN,
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
			sp:           expectedH30_PMIN6_PMAX20_R25_SWING_POSITION,
			norm:         expectedH30_PMIN6_PMAX20_R25_NORMALIZED_POSITION,
			dc:           expectedH30_PMIN6_PMAX20_R25_DOMINANT_CYCLE,
			dcm:          expectedH30_PMIN6_PMAX20_R25_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH30_PMIN6_PMAX20_R25_BAND_PASS,
			quad:         expectedH30_PMIN6_PMAX20_R25_QUADRATURE,
			amp:          expectedH30_PMIN6_PMAX20_R25_CYCLE_AMPLITUDE,
			lead60:       expectedH30_PMIN6_PMAX20_R25_LEAD60,
			span:         expectedH30_PMIN6_PMAX20_R25_POSITION_SPAN,
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
			sp:           expectedH30_PMIN10_PMAX30_R50_SWING_POSITION,
			norm:         expectedH30_PMIN10_PMAX30_R50_NORMALIZED_POSITION,
			dc:           expectedH30_PMIN10_PMAX30_R50_DOMINANT_CYCLE,
			dcm:          expectedH30_PMIN10_PMAX30_R50_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH30_PMIN10_PMAX30_R50_BAND_PASS,
			quad:         expectedH30_PMIN10_PMAX30_R50_QUADRATURE,
			amp:          expectedH30_PMIN10_PMAX30_R50_CYCLE_AMPLITUDE,
			lead60:       expectedH30_PMIN10_PMAX30_R50_LEAD60,
			span:         expectedH30_PMIN10_PMAX30_R50_POSITION_SPAN,
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
			sp:           expectedH30_PMIN10_PMAX30_R25_SWING_POSITION,
			norm:         expectedH30_PMIN10_PMAX30_R25_NORMALIZED_POSITION,
			dc:           expectedH30_PMIN10_PMAX30_R25_DOMINANT_CYCLE,
			dcm:          expectedH30_PMIN10_PMAX30_R25_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH30_PMIN10_PMAX30_R25_BAND_PASS,
			quad:         expectedH30_PMIN10_PMAX30_R25_QUADRATURE,
			amp:          expectedH30_PMIN10_PMAX30_R25_CYCLE_AMPLITUDE,
			lead60:       expectedH30_PMIN10_PMAX30_R25_LEAD60,
			span:         expectedH30_PMIN10_PMAX30_R25_POSITION_SPAN,
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
			sp:           expectedH30_PMIN10_PMAX48_R50_SWING_POSITION,
			norm:         expectedH30_PMIN10_PMAX48_R50_NORMALIZED_POSITION,
			dc:           expectedH30_PMIN10_PMAX48_R50_DOMINANT_CYCLE,
			dcm:          expectedH30_PMIN10_PMAX48_R50_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH30_PMIN10_PMAX48_R50_BAND_PASS,
			quad:         expectedH30_PMIN10_PMAX48_R50_QUADRATURE,
			amp:          expectedH30_PMIN10_PMAX48_R50_CYCLE_AMPLITUDE,
			lead60:       expectedH30_PMIN10_PMAX48_R50_LEAD60,
			span:         expectedH30_PMIN10_PMAX48_R50_POSITION_SPAN,
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
			sp:           expectedH30_PMIN10_PMAX48_R25_SWING_POSITION,
			norm:         expectedH30_PMIN10_PMAX48_R25_NORMALIZED_POSITION,
			dc:           expectedH30_PMIN10_PMAX48_R25_DOMINANT_CYCLE,
			dcm:          expectedH30_PMIN10_PMAX48_R25_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH30_PMIN10_PMAX48_R25_BAND_PASS,
			quad:         expectedH30_PMIN10_PMAX48_R25_QUADRATURE,
			amp:          expectedH30_PMIN10_PMAX48_R25_CYCLE_AMPLITUDE,
			lead60:       expectedH30_PMIN10_PMAX48_R25_LEAD60,
			span:         expectedH30_PMIN10_PMAX48_R25_POSITION_SPAN,
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
			sp:           expectedH48_PMIN6_PMAX30_R50_SWING_POSITION,
			norm:         expectedH48_PMIN6_PMAX30_R50_NORMALIZED_POSITION,
			dc:           expectedH48_PMIN6_PMAX30_R50_DOMINANT_CYCLE,
			dcm:          expectedH48_PMIN6_PMAX30_R50_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH48_PMIN6_PMAX30_R50_BAND_PASS,
			quad:         expectedH48_PMIN6_PMAX30_R50_QUADRATURE,
			amp:          expectedH48_PMIN6_PMAX30_R50_CYCLE_AMPLITUDE,
			lead60:       expectedH48_PMIN6_PMAX30_R50_LEAD60,
			span:         expectedH48_PMIN6_PMAX30_R50_POSITION_SPAN,
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
			sp:           expectedH48_PMIN6_PMAX30_R25_SWING_POSITION,
			norm:         expectedH48_PMIN6_PMAX30_R25_NORMALIZED_POSITION,
			dc:           expectedH48_PMIN6_PMAX30_R25_DOMINANT_CYCLE,
			dcm:          expectedH48_PMIN6_PMAX30_R25_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH48_PMIN6_PMAX30_R25_BAND_PASS,
			quad:         expectedH48_PMIN6_PMAX30_R25_QUADRATURE,
			amp:          expectedH48_PMIN6_PMAX30_R25_CYCLE_AMPLITUDE,
			lead60:       expectedH48_PMIN6_PMAX30_R25_LEAD60,
			span:         expectedH48_PMIN6_PMAX30_R25_POSITION_SPAN,
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
			sp:           expectedH48_PMIN6_PMAX20_R50_SWING_POSITION,
			norm:         expectedH48_PMIN6_PMAX20_R50_NORMALIZED_POSITION,
			dc:           expectedH48_PMIN6_PMAX20_R50_DOMINANT_CYCLE,
			dcm:          expectedH48_PMIN6_PMAX20_R50_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH48_PMIN6_PMAX20_R50_BAND_PASS,
			quad:         expectedH48_PMIN6_PMAX20_R50_QUADRATURE,
			amp:          expectedH48_PMIN6_PMAX20_R50_CYCLE_AMPLITUDE,
			lead60:       expectedH48_PMIN6_PMAX20_R50_LEAD60,
			span:         expectedH48_PMIN6_PMAX20_R50_POSITION_SPAN,
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
			sp:           expectedH48_PMIN6_PMAX20_R25_SWING_POSITION,
			norm:         expectedH48_PMIN6_PMAX20_R25_NORMALIZED_POSITION,
			dc:           expectedH48_PMIN6_PMAX20_R25_DOMINANT_CYCLE,
			dcm:          expectedH48_PMIN6_PMAX20_R25_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH48_PMIN6_PMAX20_R25_BAND_PASS,
			quad:         expectedH48_PMIN6_PMAX20_R25_QUADRATURE,
			amp:          expectedH48_PMIN6_PMAX20_R25_CYCLE_AMPLITUDE,
			lead60:       expectedH48_PMIN6_PMAX20_R25_LEAD60,
			span:         expectedH48_PMIN6_PMAX20_R25_POSITION_SPAN,
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
			sp:           expectedH48_PMIN10_PMAX30_R50_SWING_POSITION,
			norm:         expectedH48_PMIN10_PMAX30_R50_NORMALIZED_POSITION,
			dc:           expectedH48_PMIN10_PMAX30_R50_DOMINANT_CYCLE,
			dcm:          expectedH48_PMIN10_PMAX30_R50_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH48_PMIN10_PMAX30_R50_BAND_PASS,
			quad:         expectedH48_PMIN10_PMAX30_R50_QUADRATURE,
			amp:          expectedH48_PMIN10_PMAX30_R50_CYCLE_AMPLITUDE,
			lead60:       expectedH48_PMIN10_PMAX30_R50_LEAD60,
			span:         expectedH48_PMIN10_PMAX30_R50_POSITION_SPAN,
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
			sp:           expectedH48_PMIN10_PMAX30_R25_SWING_POSITION,
			norm:         expectedH48_PMIN10_PMAX30_R25_NORMALIZED_POSITION,
			dc:           expectedH48_PMIN10_PMAX30_R25_DOMINANT_CYCLE,
			dcm:          expectedH48_PMIN10_PMAX30_R25_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH48_PMIN10_PMAX30_R25_BAND_PASS,
			quad:         expectedH48_PMIN10_PMAX30_R25_QUADRATURE,
			amp:          expectedH48_PMIN10_PMAX30_R25_CYCLE_AMPLITUDE,
			lead60:       expectedH48_PMIN10_PMAX30_R25_LEAD60,
			span:         expectedH48_PMIN10_PMAX30_R25_POSITION_SPAN,
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
			sp:           expectedH48_PMIN10_PMAX48_R50_SWING_POSITION,
			norm:         expectedH48_PMIN10_PMAX48_R50_NORMALIZED_POSITION,
			dc:           expectedH48_PMIN10_PMAX48_R50_DOMINANT_CYCLE,
			dcm:          expectedH48_PMIN10_PMAX48_R50_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH48_PMIN10_PMAX48_R50_BAND_PASS,
			quad:         expectedH48_PMIN10_PMAX48_R50_QUADRATURE,
			amp:          expectedH48_PMIN10_PMAX48_R50_CYCLE_AMPLITUDE,
			lead60:       expectedH48_PMIN10_PMAX48_R50_LEAD60,
			span:         expectedH48_PMIN10_PMAX48_R50_POSITION_SPAN,
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
			sp:           expectedH48_PMIN10_PMAX48_R25_SWING_POSITION,
			norm:         expectedH48_PMIN10_PMAX48_R25_NORMALIZED_POSITION,
			dc:           expectedH48_PMIN10_PMAX48_R25_DOMINANT_CYCLE,
			dcm:          expectedH48_PMIN10_PMAX48_R25_DOMINANT_CYCLE_MEDIAN,
			bp:           expectedH48_PMIN10_PMAX48_R25_BAND_PASS,
			quad:         expectedH48_PMIN10_PMAX48_R25_QUADRATURE,
			amp:          expectedH48_PMIN10_PMAX48_R25_CYCLE_AMPLITUDE,
			lead60:       expectedH48_PMIN10_PMAX48_R25_LEAD60,
			span:         expectedH48_PMIN10_PMAX48_R25_POSITION_SPAN,
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

			indicator, err := NewCoronaSwingPositionParams(params)
			if err != nil {
				t.Fatalf("NewCoronaSwingPositionParams: %v", err)
			}
			step := 10.0 / float64(tc.rasterLength)
			expectedFirst := -5.0 + step
			expectedResolution := 1.0 / step
			baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

			for i, price := range testInput {
				barTime := baseTime.AddDate(0, 0, i)
				snap := indicator.Update(price, barTime)
				assertClose(t, snap.Heatmap.ParameterFirst, expectedFirst, fmt.Sprintf("%s[%d].first", tc.name, i))
				assertClose(t, snap.Heatmap.ParameterLast, 5.0, fmt.Sprintf("%s[%d].last", tc.name, i))
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
				assertClose(t, snap.SwingPosition, tc.sp[i], fmt.Sprintf("%s[%d].sp", tc.name, i))
				assertClose(t, snap.NormalizedPosition, tc.norm[i], fmt.Sprintf("%s[%d].norm", tc.name, i))
				assertClose(t, snap.DominantCycle, tc.dc[i], fmt.Sprintf("%s[%d].dc", tc.name, i))
				assertClose(t, snap.DominantCycleMedian, tc.dcm[i], fmt.Sprintf("%s[%d].dcm", tc.name, i))
				assertClose(t, snap.BandPass, tc.bp[i], fmt.Sprintf("%s[%d].bp", tc.name, i))
				assertClose(t, snap.Quadrature, tc.quad[i], fmt.Sprintf("%s[%d].quad", tc.name, i))
				assertClose(t, snap.CycleAmplitude, tc.amp[i], fmt.Sprintf("%s[%d].amp", tc.name, i))
				assertClose(t, snap.Lead60, tc.lead60[i], fmt.Sprintf("%s[%d].lead60", tc.name, i))
				assertClose(t, snap.PositionSpan, tc.span[i], fmt.Sprintf("%s[%d].span", tc.name, i))
				assertClose(t, snap.Width, tc.width[i], fmt.Sprintf("%s[%d].width", tc.name, i))
				assertClose(t, snap.HasCycleEstimate, tc.hasEstimate[i], fmt.Sprintf("%s[%d].hasEst", tc.name, i))
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

	indicator, err := NewCoronaSwingPositionDefault()
	if err != nil {
		t.Fatalf("NewCoronaSwingPositionDefault: %v", err)
	}
	meta := indicator.Metadata()
	if meta.Identifier != core.CoronaSwingPosition || meta.Mnemonic != "cswing(50, 20, -5, 5, 30)" {
		t.Fatalf("unexpected metadata: %+v", meta)
	}
	if len(meta.Outputs) != 2 || meta.Outputs[0].Shape != shape.Heatmap || meta.Outputs[1].Shape != shape.Scalar {
		t.Fatalf("unexpected outputs: %+v", meta.Outputs)
	}
	if meta.Outputs[1].Mnemonic != "cswing-sp(30)" {
		t.Fatalf("unexpected secondary mnemonic: %q", meta.Outputs[1].Mnemonic)
	}

	flaggedParams := DefaultParams()
	flaggedParams.MinimalPeriod = 10
	flaggedParams.MaximalPeriod = 48
	flaggedParams.LeadWindow = 31
	flaggedParams.PositionWindow = 15
	flaggedParams.BarComponent = entities.BarMedianPrice
	flagged, err := NewCoronaSwingPositionParams(flaggedParams)
	if err != nil {
		t.Fatalf("NewCoronaSwingPositionParams: %v", err)
	}
	if flagged.Metadata().Mnemonic != "cswing(50, 20, -5, 5, 30, pmin=10, pmax=48, lead=31, pos=15, hl/2)" {
		t.Fatalf("unexpected flagged mnemonic: %q", flagged.Metadata().Mnemonic)
	}
	if flagged.Metadata().Outputs[1].Mnemonic != "cswing-sp(30, pmin=10, pmax=48, lead=31, pos=15, hl/2)" {
		t.Fatalf("unexpected flagged secondary mnemonic: %q", flagged.Metadata().Outputs[1].Mnemonic)
	}

	invalid := []func(*Params){
		func(p *Params) { p.RasterLength = 1 },
		func(p *Params) { p.RasterLength = 0 },
		func(p *Params) { p.MaxRasterValue = 6.0 },
		func(p *Params) { p.MinParameterValue = 5.0; p.MaxParameterValue = 5.0 },
		func(p *Params) { p.HighPassFilterCutoff = 4 },
		func(p *Params) { p.MinimalPeriod = 5 },
		func(p *Params) { p.MinimalPeriod = 10; p.MaximalPeriod = 10 },
		func(p *Params) { p.LeadWindow = 1 },
		func(p *Params) { p.PositionWindow = 1 },
	}
	for i, mutate := range invalid {
		bad := DefaultParams()
		mutate(bad)
		if _, err := NewCoronaSwingPositionParams(bad); err == nil {
			t.Fatalf("invalid[%d]: expected error, got nil", i)
		}
	}
}
