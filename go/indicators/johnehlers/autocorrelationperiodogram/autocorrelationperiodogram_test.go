//nolint:testpackage
package autocorrelationperiodogram

import (
	"encoding/json"
	"math"
	"slices"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

const (
	fixtureTolerance = 1e-11
	oracleTolerance  = 2e-10
	exactTolerance   = 1e-13
)

type referenceCase struct {
	name            string
	params          Params
	expectedHeatmap [][]float64
	expectedMin     []float64
	expectedMax     []float64
	expectedDC      []float64
}

func allReferenceCases() []referenceCase {
	return []referenceCase{
		{
			name:            "V0_L8_H24_A0_E0",
			params:          Params{Source: Article, MinPeriod: 8, MaxPeriod: 24, AveragingLength: 0, EnhanceResolution: false},
			expectedHeatmap: expectedV0_L8_H24_A0_E0_HEATMAP,
			expectedMin:     expectedV0_L8_H24_A0_E0_MIN,
			expectedMax:     expectedV0_L8_H24_A0_E0_MAX,
			expectedDC:      expectedV0_L8_H24_A0_E0_DC,
		},
		{
			name:            "V0_L8_H24_A0_E1",
			params:          Params{Source: Article, MinPeriod: 8, MaxPeriod: 24, AveragingLength: 0, EnhanceResolution: true},
			expectedHeatmap: expectedV0_L8_H24_A0_E1_HEATMAP,
			expectedMin:     expectedV0_L8_H24_A0_E1_MIN,
			expectedMax:     expectedV0_L8_H24_A0_E1_MAX,
			expectedDC:      expectedV0_L8_H24_A0_E1_DC,
		},
		{
			name:            "V0_L8_H24_A3_E0",
			params:          Params{Source: Article, MinPeriod: 8, MaxPeriod: 24, AveragingLength: 3, EnhanceResolution: false},
			expectedHeatmap: expectedV0_L8_H24_A3_E0_HEATMAP,
			expectedMin:     expectedV0_L8_H24_A3_E0_MIN,
			expectedMax:     expectedV0_L8_H24_A3_E0_MAX,
			expectedDC:      expectedV0_L8_H24_A3_E0_DC,
		},
		{
			name:            "V0_L8_H24_A3_E1",
			params:          Params{Source: Article, MinPeriod: 8, MaxPeriod: 24, AveragingLength: 3, EnhanceResolution: true},
			expectedHeatmap: expectedV0_L8_H24_A3_E1_HEATMAP,
			expectedMin:     expectedV0_L8_H24_A3_E1_MIN,
			expectedMax:     expectedV0_L8_H24_A3_E1_MAX,
			expectedDC:      expectedV0_L8_H24_A3_E1_DC,
		},
		{
			name:            "V0_L8_H24_A5_E0",
			params:          Params{Source: Article, MinPeriod: 8, MaxPeriod: 24, AveragingLength: 5, EnhanceResolution: false},
			expectedHeatmap: expectedV0_L8_H24_A5_E0_HEATMAP,
			expectedMin:     expectedV0_L8_H24_A5_E0_MIN,
			expectedMax:     expectedV0_L8_H24_A5_E0_MAX,
			expectedDC:      expectedV0_L8_H24_A5_E0_DC,
		},
		{
			name:            "V0_L8_H24_A5_E1",
			params:          Params{Source: Article, MinPeriod: 8, MaxPeriod: 24, AveragingLength: 5, EnhanceResolution: true},
			expectedHeatmap: expectedV0_L8_H24_A5_E1_HEATMAP,
			expectedMin:     expectedV0_L8_H24_A5_E1_MIN,
			expectedMax:     expectedV0_L8_H24_A5_E1_MAX,
			expectedDC:      expectedV0_L8_H24_A5_E1_DC,
		},
		{
			name:            "V0_L8_H24_A10_E0",
			params:          Params{Source: Article, MinPeriod: 8, MaxPeriod: 24, AveragingLength: 10, EnhanceResolution: false},
			expectedHeatmap: expectedV0_L8_H24_A10_E0_HEATMAP,
			expectedMin:     expectedV0_L8_H24_A10_E0_MIN,
			expectedMax:     expectedV0_L8_H24_A10_E0_MAX,
			expectedDC:      expectedV0_L8_H24_A10_E0_DC,
		},
		{
			name:            "V0_L8_H24_A10_E1",
			params:          Params{Source: Article, MinPeriod: 8, MaxPeriod: 24, AveragingLength: 10, EnhanceResolution: true},
			expectedHeatmap: expectedV0_L8_H24_A10_E1_HEATMAP,
			expectedMin:     expectedV0_L8_H24_A10_E1_MIN,
			expectedMax:     expectedV0_L8_H24_A10_E1_MAX,
			expectedDC:      expectedV0_L8_H24_A10_E1_DC,
		},
		{
			name:            "V0_L8_H48_A0_E0",
			params:          Params{Source: Article, MinPeriod: 8, MaxPeriod: 48, AveragingLength: 0, EnhanceResolution: false},
			expectedHeatmap: expectedV0_L8_H48_A0_E0_HEATMAP,
			expectedMin:     expectedV0_L8_H48_A0_E0_MIN,
			expectedMax:     expectedV0_L8_H48_A0_E0_MAX,
			expectedDC:      expectedV0_L8_H48_A0_E0_DC,
		},
		{
			name:            "V0_L8_H48_A0_E1",
			params:          Params{Source: Article, MinPeriod: 8, MaxPeriod: 48, AveragingLength: 0, EnhanceResolution: true},
			expectedHeatmap: expectedV0_L8_H48_A0_E1_HEATMAP,
			expectedMin:     expectedV0_L8_H48_A0_E1_MIN,
			expectedMax:     expectedV0_L8_H48_A0_E1_MAX,
			expectedDC:      expectedV0_L8_H48_A0_E1_DC,
		},
		{
			name:            "V0_L8_H48_A3_E0",
			params:          Params{Source: Article, MinPeriod: 8, MaxPeriod: 48, AveragingLength: 3, EnhanceResolution: false},
			expectedHeatmap: expectedV0_L8_H48_A3_E0_HEATMAP,
			expectedMin:     expectedV0_L8_H48_A3_E0_MIN,
			expectedMax:     expectedV0_L8_H48_A3_E0_MAX,
			expectedDC:      expectedV0_L8_H48_A3_E0_DC,
		},
		{
			name:            "V0_L8_H48_A3_E1",
			params:          Params{Source: Article, MinPeriod: 8, MaxPeriod: 48, AveragingLength: 3, EnhanceResolution: true},
			expectedHeatmap: expectedV0_L8_H48_A3_E1_HEATMAP,
			expectedMin:     expectedV0_L8_H48_A3_E1_MIN,
			expectedMax:     expectedV0_L8_H48_A3_E1_MAX,
			expectedDC:      expectedV0_L8_H48_A3_E1_DC,
		},
		{
			name:            "V0_L8_H48_A5_E0",
			params:          Params{Source: Article, MinPeriod: 8, MaxPeriod: 48, AveragingLength: 5, EnhanceResolution: false},
			expectedHeatmap: expectedV0_L8_H48_A5_E0_HEATMAP,
			expectedMin:     expectedV0_L8_H48_A5_E0_MIN,
			expectedMax:     expectedV0_L8_H48_A5_E0_MAX,
			expectedDC:      expectedV0_L8_H48_A5_E0_DC,
		},
		{
			name:            "V0_L8_H48_A5_E1",
			params:          Params{Source: Article, MinPeriod: 8, MaxPeriod: 48, AveragingLength: 5, EnhanceResolution: true},
			expectedHeatmap: expectedV0_L8_H48_A5_E1_HEATMAP,
			expectedMin:     expectedV0_L8_H48_A5_E1_MIN,
			expectedMax:     expectedV0_L8_H48_A5_E1_MAX,
			expectedDC:      expectedV0_L8_H48_A5_E1_DC,
		},
		{
			name:            "V0_L8_H48_A10_E0",
			params:          Params{Source: Article, MinPeriod: 8, MaxPeriod: 48, AveragingLength: 10, EnhanceResolution: false},
			expectedHeatmap: expectedV0_L8_H48_A10_E0_HEATMAP,
			expectedMin:     expectedV0_L8_H48_A10_E0_MIN,
			expectedMax:     expectedV0_L8_H48_A10_E0_MAX,
			expectedDC:      expectedV0_L8_H48_A10_E0_DC,
		},
		{
			name:            "V0_L8_H48_A10_E1",
			params:          Params{Source: Article, MinPeriod: 8, MaxPeriod: 48, AveragingLength: 10, EnhanceResolution: true},
			expectedHeatmap: expectedV0_L8_H48_A10_E1_HEATMAP,
			expectedMin:     expectedV0_L8_H48_A10_E1_MIN,
			expectedMax:     expectedV0_L8_H48_A10_E1_MAX,
			expectedDC:      expectedV0_L8_H48_A10_E1_DC,
		},
		{
			name:            "V0_L10_H48_A0_E0",
			params:          Params{Source: Article, MinPeriod: 10, MaxPeriod: 48, AveragingLength: 0, EnhanceResolution: false},
			expectedHeatmap: expectedV0_L10_H48_A0_E0_HEATMAP,
			expectedMin:     expectedV0_L10_H48_A0_E0_MIN,
			expectedMax:     expectedV0_L10_H48_A0_E0_MAX,
			expectedDC:      expectedV0_L10_H48_A0_E0_DC,
		},
		{
			name:            "V0_L10_H48_A0_E1",
			params:          Params{Source: Article, MinPeriod: 10, MaxPeriod: 48, AveragingLength: 0, EnhanceResolution: true},
			expectedHeatmap: expectedV0_L10_H48_A0_E1_HEATMAP,
			expectedMin:     expectedV0_L10_H48_A0_E1_MIN,
			expectedMax:     expectedV0_L10_H48_A0_E1_MAX,
			expectedDC:      expectedV0_L10_H48_A0_E1_DC,
		},
		{
			name:            "V0_L10_H48_A3_E0",
			params:          Params{Source: Article, MinPeriod: 10, MaxPeriod: 48, AveragingLength: 3, EnhanceResolution: false},
			expectedHeatmap: expectedV0_L10_H48_A3_E0_HEATMAP,
			expectedMin:     expectedV0_L10_H48_A3_E0_MIN,
			expectedMax:     expectedV0_L10_H48_A3_E0_MAX,
			expectedDC:      expectedV0_L10_H48_A3_E0_DC,
		},
		{
			name:            "V0_L10_H48_A3_E1",
			params:          Params{Source: Article, MinPeriod: 10, MaxPeriod: 48, AveragingLength: 3, EnhanceResolution: true},
			expectedHeatmap: expectedV0_L10_H48_A3_E1_HEATMAP,
			expectedMin:     expectedV0_L10_H48_A3_E1_MIN,
			expectedMax:     expectedV0_L10_H48_A3_E1_MAX,
			expectedDC:      expectedV0_L10_H48_A3_E1_DC,
		},
		{
			name:            "V0_L10_H48_A5_E0",
			params:          Params{Source: Article, MinPeriod: 10, MaxPeriod: 48, AveragingLength: 5, EnhanceResolution: false},
			expectedHeatmap: expectedV0_L10_H48_A5_E0_HEATMAP,
			expectedMin:     expectedV0_L10_H48_A5_E0_MIN,
			expectedMax:     expectedV0_L10_H48_A5_E0_MAX,
			expectedDC:      expectedV0_L10_H48_A5_E0_DC,
		},
		{
			name:            "V0_L10_H48_A5_E1",
			params:          Params{Source: Article, MinPeriod: 10, MaxPeriod: 48, AveragingLength: 5, EnhanceResolution: true},
			expectedHeatmap: expectedV0_L10_H48_A5_E1_HEATMAP,
			expectedMin:     expectedV0_L10_H48_A5_E1_MIN,
			expectedMax:     expectedV0_L10_H48_A5_E1_MAX,
			expectedDC:      expectedV0_L10_H48_A5_E1_DC,
		},
		{
			name:            "V0_L10_H48_A10_E0",
			params:          Params{Source: Article, MinPeriod: 10, MaxPeriod: 48, AveragingLength: 10, EnhanceResolution: false},
			expectedHeatmap: expectedV0_L10_H48_A10_E0_HEATMAP,
			expectedMin:     expectedV0_L10_H48_A10_E0_MIN,
			expectedMax:     expectedV0_L10_H48_A10_E0_MAX,
			expectedDC:      expectedV0_L10_H48_A10_E0_DC,
		},
		{
			name:            "V0_L10_H48_A10_E1",
			params:          Params{Source: Article, MinPeriod: 10, MaxPeriod: 48, AveragingLength: 10, EnhanceResolution: true},
			expectedHeatmap: expectedV0_L10_H48_A10_E1_HEATMAP,
			expectedMin:     expectedV0_L10_H48_A10_E1_MIN,
			expectedMax:     expectedV0_L10_H48_A10_E1_MAX,
			expectedDC:      expectedV0_L10_H48_A10_E1_DC,
		},
		{
			name:            "V0_L16_H64_A0_E0",
			params:          Params{Source: Article, MinPeriod: 16, MaxPeriod: 64, AveragingLength: 0, EnhanceResolution: false},
			expectedHeatmap: expectedV0_L16_H64_A0_E0_HEATMAP,
			expectedMin:     expectedV0_L16_H64_A0_E0_MIN,
			expectedMax:     expectedV0_L16_H64_A0_E0_MAX,
			expectedDC:      expectedV0_L16_H64_A0_E0_DC,
		},
		{
			name:            "V0_L16_H64_A0_E1",
			params:          Params{Source: Article, MinPeriod: 16, MaxPeriod: 64, AveragingLength: 0, EnhanceResolution: true},
			expectedHeatmap: expectedV0_L16_H64_A0_E1_HEATMAP,
			expectedMin:     expectedV0_L16_H64_A0_E1_MIN,
			expectedMax:     expectedV0_L16_H64_A0_E1_MAX,
			expectedDC:      expectedV0_L16_H64_A0_E1_DC,
		},
		{
			name:            "V0_L16_H64_A3_E0",
			params:          Params{Source: Article, MinPeriod: 16, MaxPeriod: 64, AveragingLength: 3, EnhanceResolution: false},
			expectedHeatmap: expectedV0_L16_H64_A3_E0_HEATMAP,
			expectedMin:     expectedV0_L16_H64_A3_E0_MIN,
			expectedMax:     expectedV0_L16_H64_A3_E0_MAX,
			expectedDC:      expectedV0_L16_H64_A3_E0_DC,
		},
		{
			name:            "V0_L16_H64_A3_E1",
			params:          Params{Source: Article, MinPeriod: 16, MaxPeriod: 64, AveragingLength: 3, EnhanceResolution: true},
			expectedHeatmap: expectedV0_L16_H64_A3_E1_HEATMAP,
			expectedMin:     expectedV0_L16_H64_A3_E1_MIN,
			expectedMax:     expectedV0_L16_H64_A3_E1_MAX,
			expectedDC:      expectedV0_L16_H64_A3_E1_DC,
		},
		{
			name:            "V0_L16_H64_A5_E0",
			params:          Params{Source: Article, MinPeriod: 16, MaxPeriod: 64, AveragingLength: 5, EnhanceResolution: false},
			expectedHeatmap: expectedV0_L16_H64_A5_E0_HEATMAP,
			expectedMin:     expectedV0_L16_H64_A5_E0_MIN,
			expectedMax:     expectedV0_L16_H64_A5_E0_MAX,
			expectedDC:      expectedV0_L16_H64_A5_E0_DC,
		},
		{
			name:            "V0_L16_H64_A5_E1",
			params:          Params{Source: Article, MinPeriod: 16, MaxPeriod: 64, AveragingLength: 5, EnhanceResolution: true},
			expectedHeatmap: expectedV0_L16_H64_A5_E1_HEATMAP,
			expectedMin:     expectedV0_L16_H64_A5_E1_MIN,
			expectedMax:     expectedV0_L16_H64_A5_E1_MAX,
			expectedDC:      expectedV0_L16_H64_A5_E1_DC,
		},
		{
			name:            "V0_L16_H64_A10_E0",
			params:          Params{Source: Article, MinPeriod: 16, MaxPeriod: 64, AveragingLength: 10, EnhanceResolution: false},
			expectedHeatmap: expectedV0_L16_H64_A10_E0_HEATMAP,
			expectedMin:     expectedV0_L16_H64_A10_E0_MIN,
			expectedMax:     expectedV0_L16_H64_A10_E0_MAX,
			expectedDC:      expectedV0_L16_H64_A10_E0_DC,
		},
		{
			name:            "V0_L16_H64_A10_E1",
			params:          Params{Source: Article, MinPeriod: 16, MaxPeriod: 64, AveragingLength: 10, EnhanceResolution: true},
			expectedHeatmap: expectedV0_L16_H64_A10_E1_HEATMAP,
			expectedMin:     expectedV0_L16_H64_A10_E1_MIN,
			expectedMax:     expectedV0_L16_H64_A10_E1_MAX,
			expectedDC:      expectedV0_L16_H64_A10_E1_DC,
		},
		{
			name:            "V1_L8_H24_A0_E0",
			params:          Params{Source: Book, MinPeriod: 8, MaxPeriod: 24, AveragingLength: 0, EnhanceResolution: false},
			expectedHeatmap: expectedV1_L8_H24_A0_E0_HEATMAP,
			expectedMin:     expectedV1_L8_H24_A0_E0_MIN,
			expectedMax:     expectedV1_L8_H24_A0_E0_MAX,
			expectedDC:      expectedV1_L8_H24_A0_E0_DC,
		},
		{
			name:            "V1_L8_H24_A0_E1",
			params:          Params{Source: Book, MinPeriod: 8, MaxPeriod: 24, AveragingLength: 0, EnhanceResolution: true},
			expectedHeatmap: expectedV1_L8_H24_A0_E1_HEATMAP,
			expectedMin:     expectedV1_L8_H24_A0_E1_MIN,
			expectedMax:     expectedV1_L8_H24_A0_E1_MAX,
			expectedDC:      expectedV1_L8_H24_A0_E1_DC,
		},
		{
			name:            "V1_L8_H24_A3_E0",
			params:          Params{Source: Book, MinPeriod: 8, MaxPeriod: 24, AveragingLength: 3, EnhanceResolution: false},
			expectedHeatmap: expectedV1_L8_H24_A3_E0_HEATMAP,
			expectedMin:     expectedV1_L8_H24_A3_E0_MIN,
			expectedMax:     expectedV1_L8_H24_A3_E0_MAX,
			expectedDC:      expectedV1_L8_H24_A3_E0_DC,
		},
		{
			name:            "V1_L8_H24_A3_E1",
			params:          Params{Source: Book, MinPeriod: 8, MaxPeriod: 24, AveragingLength: 3, EnhanceResolution: true},
			expectedHeatmap: expectedV1_L8_H24_A3_E1_HEATMAP,
			expectedMin:     expectedV1_L8_H24_A3_E1_MIN,
			expectedMax:     expectedV1_L8_H24_A3_E1_MAX,
			expectedDC:      expectedV1_L8_H24_A3_E1_DC,
		},
		{
			name:            "V1_L8_H24_A5_E0",
			params:          Params{Source: Book, MinPeriod: 8, MaxPeriod: 24, AveragingLength: 5, EnhanceResolution: false},
			expectedHeatmap: expectedV1_L8_H24_A5_E0_HEATMAP,
			expectedMin:     expectedV1_L8_H24_A5_E0_MIN,
			expectedMax:     expectedV1_L8_H24_A5_E0_MAX,
			expectedDC:      expectedV1_L8_H24_A5_E0_DC,
		},
		{
			name:            "V1_L8_H24_A5_E1",
			params:          Params{Source: Book, MinPeriod: 8, MaxPeriod: 24, AveragingLength: 5, EnhanceResolution: true},
			expectedHeatmap: expectedV1_L8_H24_A5_E1_HEATMAP,
			expectedMin:     expectedV1_L8_H24_A5_E1_MIN,
			expectedMax:     expectedV1_L8_H24_A5_E1_MAX,
			expectedDC:      expectedV1_L8_H24_A5_E1_DC,
		},
		{
			name:            "V1_L8_H24_A10_E0",
			params:          Params{Source: Book, MinPeriod: 8, MaxPeriod: 24, AveragingLength: 10, EnhanceResolution: false},
			expectedHeatmap: expectedV1_L8_H24_A10_E0_HEATMAP,
			expectedMin:     expectedV1_L8_H24_A10_E0_MIN,
			expectedMax:     expectedV1_L8_H24_A10_E0_MAX,
			expectedDC:      expectedV1_L8_H24_A10_E0_DC,
		},
		{
			name:            "V1_L8_H24_A10_E1",
			params:          Params{Source: Book, MinPeriod: 8, MaxPeriod: 24, AveragingLength: 10, EnhanceResolution: true},
			expectedHeatmap: expectedV1_L8_H24_A10_E1_HEATMAP,
			expectedMin:     expectedV1_L8_H24_A10_E1_MIN,
			expectedMax:     expectedV1_L8_H24_A10_E1_MAX,
			expectedDC:      expectedV1_L8_H24_A10_E1_DC,
		},
		{
			name:            "V1_L8_H48_A0_E0",
			params:          Params{Source: Book, MinPeriod: 8, MaxPeriod: 48, AveragingLength: 0, EnhanceResolution: false},
			expectedHeatmap: expectedV1_L8_H48_A0_E0_HEATMAP,
			expectedMin:     expectedV1_L8_H48_A0_E0_MIN,
			expectedMax:     expectedV1_L8_H48_A0_E0_MAX,
			expectedDC:      expectedV1_L8_H48_A0_E0_DC,
		},
		{
			name:            "V1_L8_H48_A0_E1",
			params:          Params{Source: Book, MinPeriod: 8, MaxPeriod: 48, AveragingLength: 0, EnhanceResolution: true},
			expectedHeatmap: expectedV1_L8_H48_A0_E1_HEATMAP,
			expectedMin:     expectedV1_L8_H48_A0_E1_MIN,
			expectedMax:     expectedV1_L8_H48_A0_E1_MAX,
			expectedDC:      expectedV1_L8_H48_A0_E1_DC,
		},
		{
			name:            "V1_L8_H48_A3_E0",
			params:          Params{Source: Book, MinPeriod: 8, MaxPeriod: 48, AveragingLength: 3, EnhanceResolution: false},
			expectedHeatmap: expectedV1_L8_H48_A3_E0_HEATMAP,
			expectedMin:     expectedV1_L8_H48_A3_E0_MIN,
			expectedMax:     expectedV1_L8_H48_A3_E0_MAX,
			expectedDC:      expectedV1_L8_H48_A3_E0_DC,
		},
		{
			name:            "V1_L8_H48_A3_E1",
			params:          Params{Source: Book, MinPeriod: 8, MaxPeriod: 48, AveragingLength: 3, EnhanceResolution: true},
			expectedHeatmap: expectedV1_L8_H48_A3_E1_HEATMAP,
			expectedMin:     expectedV1_L8_H48_A3_E1_MIN,
			expectedMax:     expectedV1_L8_H48_A3_E1_MAX,
			expectedDC:      expectedV1_L8_H48_A3_E1_DC,
		},
		{
			name:            "V1_L8_H48_A5_E0",
			params:          Params{Source: Book, MinPeriod: 8, MaxPeriod: 48, AveragingLength: 5, EnhanceResolution: false},
			expectedHeatmap: expectedV1_L8_H48_A5_E0_HEATMAP,
			expectedMin:     expectedV1_L8_H48_A5_E0_MIN,
			expectedMax:     expectedV1_L8_H48_A5_E0_MAX,
			expectedDC:      expectedV1_L8_H48_A5_E0_DC,
		},
		{
			name:            "V1_L8_H48_A5_E1",
			params:          Params{Source: Book, MinPeriod: 8, MaxPeriod: 48, AveragingLength: 5, EnhanceResolution: true},
			expectedHeatmap: expectedV1_L8_H48_A5_E1_HEATMAP,
			expectedMin:     expectedV1_L8_H48_A5_E1_MIN,
			expectedMax:     expectedV1_L8_H48_A5_E1_MAX,
			expectedDC:      expectedV1_L8_H48_A5_E1_DC,
		},
		{
			name:            "V1_L8_H48_A10_E0",
			params:          Params{Source: Book, MinPeriod: 8, MaxPeriod: 48, AveragingLength: 10, EnhanceResolution: false},
			expectedHeatmap: expectedV1_L8_H48_A10_E0_HEATMAP,
			expectedMin:     expectedV1_L8_H48_A10_E0_MIN,
			expectedMax:     expectedV1_L8_H48_A10_E0_MAX,
			expectedDC:      expectedV1_L8_H48_A10_E0_DC,
		},
		{
			name:            "V1_L8_H48_A10_E1",
			params:          Params{Source: Book, MinPeriod: 8, MaxPeriod: 48, AveragingLength: 10, EnhanceResolution: true},
			expectedHeatmap: expectedV1_L8_H48_A10_E1_HEATMAP,
			expectedMin:     expectedV1_L8_H48_A10_E1_MIN,
			expectedMax:     expectedV1_L8_H48_A10_E1_MAX,
			expectedDC:      expectedV1_L8_H48_A10_E1_DC,
		},
		{
			name:            "V1_L10_H48_A0_E0",
			params:          Params{Source: Book, MinPeriod: 10, MaxPeriod: 48, AveragingLength: 0, EnhanceResolution: false},
			expectedHeatmap: expectedV1_L10_H48_A0_E0_HEATMAP,
			expectedMin:     expectedV1_L10_H48_A0_E0_MIN,
			expectedMax:     expectedV1_L10_H48_A0_E0_MAX,
			expectedDC:      expectedV1_L10_H48_A0_E0_DC,
		},
		{
			name:            "V1_L10_H48_A0_E1",
			params:          Params{Source: Book, MinPeriod: 10, MaxPeriod: 48, AveragingLength: 0, EnhanceResolution: true},
			expectedHeatmap: expectedV1_L10_H48_A0_E1_HEATMAP,
			expectedMin:     expectedV1_L10_H48_A0_E1_MIN,
			expectedMax:     expectedV1_L10_H48_A0_E1_MAX,
			expectedDC:      expectedV1_L10_H48_A0_E1_DC,
		},
		{
			name:            "V1_L10_H48_A3_E0",
			params:          Params{Source: Book, MinPeriod: 10, MaxPeriod: 48, AveragingLength: 3, EnhanceResolution: false},
			expectedHeatmap: expectedV1_L10_H48_A3_E0_HEATMAP,
			expectedMin:     expectedV1_L10_H48_A3_E0_MIN,
			expectedMax:     expectedV1_L10_H48_A3_E0_MAX,
			expectedDC:      expectedV1_L10_H48_A3_E0_DC,
		},
		{
			name:            "V1_L10_H48_A3_E1",
			params:          Params{Source: Book, MinPeriod: 10, MaxPeriod: 48, AveragingLength: 3, EnhanceResolution: true},
			expectedHeatmap: expectedV1_L10_H48_A3_E1_HEATMAP,
			expectedMin:     expectedV1_L10_H48_A3_E1_MIN,
			expectedMax:     expectedV1_L10_H48_A3_E1_MAX,
			expectedDC:      expectedV1_L10_H48_A3_E1_DC,
		},
		{
			name:            "V1_L10_H48_A5_E0",
			params:          Params{Source: Book, MinPeriod: 10, MaxPeriod: 48, AveragingLength: 5, EnhanceResolution: false},
			expectedHeatmap: expectedV1_L10_H48_A5_E0_HEATMAP,
			expectedMin:     expectedV1_L10_H48_A5_E0_MIN,
			expectedMax:     expectedV1_L10_H48_A5_E0_MAX,
			expectedDC:      expectedV1_L10_H48_A5_E0_DC,
		},
		{
			name:            "V1_L10_H48_A5_E1",
			params:          Params{Source: Book, MinPeriod: 10, MaxPeriod: 48, AveragingLength: 5, EnhanceResolution: true},
			expectedHeatmap: expectedV1_L10_H48_A5_E1_HEATMAP,
			expectedMin:     expectedV1_L10_H48_A5_E1_MIN,
			expectedMax:     expectedV1_L10_H48_A5_E1_MAX,
			expectedDC:      expectedV1_L10_H48_A5_E1_DC,
		},
		{
			name:            "V1_L10_H48_A10_E0",
			params:          Params{Source: Book, MinPeriod: 10, MaxPeriod: 48, AveragingLength: 10, EnhanceResolution: false},
			expectedHeatmap: expectedV1_L10_H48_A10_E0_HEATMAP,
			expectedMin:     expectedV1_L10_H48_A10_E0_MIN,
			expectedMax:     expectedV1_L10_H48_A10_E0_MAX,
			expectedDC:      expectedV1_L10_H48_A10_E0_DC,
		},
		{
			name:            "V1_L10_H48_A10_E1",
			params:          Params{Source: Book, MinPeriod: 10, MaxPeriod: 48, AveragingLength: 10, EnhanceResolution: true},
			expectedHeatmap: expectedV1_L10_H48_A10_E1_HEATMAP,
			expectedMin:     expectedV1_L10_H48_A10_E1_MIN,
			expectedMax:     expectedV1_L10_H48_A10_E1_MAX,
			expectedDC:      expectedV1_L10_H48_A10_E1_DC,
		},
		{
			name:            "V1_L16_H64_A0_E0",
			params:          Params{Source: Book, MinPeriod: 16, MaxPeriod: 64, AveragingLength: 0, EnhanceResolution: false},
			expectedHeatmap: expectedV1_L16_H64_A0_E0_HEATMAP,
			expectedMin:     expectedV1_L16_H64_A0_E0_MIN,
			expectedMax:     expectedV1_L16_H64_A0_E0_MAX,
			expectedDC:      expectedV1_L16_H64_A0_E0_DC,
		},
		{
			name:            "V1_L16_H64_A0_E1",
			params:          Params{Source: Book, MinPeriod: 16, MaxPeriod: 64, AveragingLength: 0, EnhanceResolution: true},
			expectedHeatmap: expectedV1_L16_H64_A0_E1_HEATMAP,
			expectedMin:     expectedV1_L16_H64_A0_E1_MIN,
			expectedMax:     expectedV1_L16_H64_A0_E1_MAX,
			expectedDC:      expectedV1_L16_H64_A0_E1_DC,
		},
		{
			name:            "V1_L16_H64_A3_E0",
			params:          Params{Source: Book, MinPeriod: 16, MaxPeriod: 64, AveragingLength: 3, EnhanceResolution: false},
			expectedHeatmap: expectedV1_L16_H64_A3_E0_HEATMAP,
			expectedMin:     expectedV1_L16_H64_A3_E0_MIN,
			expectedMax:     expectedV1_L16_H64_A3_E0_MAX,
			expectedDC:      expectedV1_L16_H64_A3_E0_DC,
		},
		{
			name:            "V1_L16_H64_A3_E1",
			params:          Params{Source: Book, MinPeriod: 16, MaxPeriod: 64, AveragingLength: 3, EnhanceResolution: true},
			expectedHeatmap: expectedV1_L16_H64_A3_E1_HEATMAP,
			expectedMin:     expectedV1_L16_H64_A3_E1_MIN,
			expectedMax:     expectedV1_L16_H64_A3_E1_MAX,
			expectedDC:      expectedV1_L16_H64_A3_E1_DC,
		},
		{
			name:            "V1_L16_H64_A5_E0",
			params:          Params{Source: Book, MinPeriod: 16, MaxPeriod: 64, AveragingLength: 5, EnhanceResolution: false},
			expectedHeatmap: expectedV1_L16_H64_A5_E0_HEATMAP,
			expectedMin:     expectedV1_L16_H64_A5_E0_MIN,
			expectedMax:     expectedV1_L16_H64_A5_E0_MAX,
			expectedDC:      expectedV1_L16_H64_A5_E0_DC,
		},
		{
			name:            "V1_L16_H64_A5_E1",
			params:          Params{Source: Book, MinPeriod: 16, MaxPeriod: 64, AveragingLength: 5, EnhanceResolution: true},
			expectedHeatmap: expectedV1_L16_H64_A5_E1_HEATMAP,
			expectedMin:     expectedV1_L16_H64_A5_E1_MIN,
			expectedMax:     expectedV1_L16_H64_A5_E1_MAX,
			expectedDC:      expectedV1_L16_H64_A5_E1_DC,
		},
		{
			name:            "V1_L16_H64_A10_E0",
			params:          Params{Source: Book, MinPeriod: 16, MaxPeriod: 64, AveragingLength: 10, EnhanceResolution: false},
			expectedHeatmap: expectedV1_L16_H64_A10_E0_HEATMAP,
			expectedMin:     expectedV1_L16_H64_A10_E0_MIN,
			expectedMax:     expectedV1_L16_H64_A10_E0_MAX,
			expectedDC:      expectedV1_L16_H64_A10_E0_DC,
		},
		{
			name:            "V1_L16_H64_A10_E1",
			params:          Params{Source: Book, MinPeriod: 16, MaxPeriod: 64, AveragingLength: 10, EnhanceResolution: true},
			expectedHeatmap: expectedV1_L16_H64_A10_E1_HEATMAP,
			expectedMin:     expectedV1_L16_H64_A10_E1_MIN,
			expectedMax:     expectedV1_L16_H64_A10_E1_MAX,
			expectedDC:      expectedV1_L16_H64_A10_E1_DC,
		},
	}
}

func assertClose(t *testing.T, actual, expected, tolerance float64, label string) {
	t.Helper()
	if math.IsNaN(expected) {
		if !math.IsNaN(actual) {
			t.Fatalf("%s: expected NaN, got %.17g", label, actual)
		}
		return
	}
	if math.IsNaN(actual) {
		t.Fatalf("%s: expected %.17g, got NaN", label, expected)
	}
	bound := math.Max(tolerance, math.Abs(expected)*tolerance)
	if math.Abs(actual-expected) > bound {
		t.Fatalf("%s: expected %.17g, got %.17g (diff=%.17g, bound=%.17g)", label, expected, actual, math.Abs(actual-expected), bound)
	}
}

type oracleStep struct {
	highPass     float64
	filtered     float64
	correlations []float64
	powers       []float64
	maximum      float64
	dominant     float64
	display      []float64
	hasDominant  bool
}

func degreeToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180.0
}

func oldAt(series []float64, index int) float64 {
	if index >= 0 {
		return series[index]
	}
	return 0.0
}

func sourceReference(prices []float64, configuration Params) []oracleStep {
	book := configuration.Source == Book
	hpHistory := make([]float64, 0, len(prices))
	filteredHistory := make([]float64, 0, len(prices))
	correlations := make([]float64, configuration.MaxPeriod+1)
	binTotal := configuration.MaxPeriod - configuration.MinPeriod + 1
	powers := make([]float64, binTotal)
	maximum := 0.0
	dominant := 0.0
	hasDominant := false
	steps := make([]oracleStep, 0, len(prices))

	ssAngle := degreeToRadians(math.Sqrt2 * 180.0 / float64(configuration.MinPeriod))
	decay := math.Exp(-math.Sqrt2 * math.Pi / float64(configuration.MinPeriod))
	c2 := 2.0 * decay * math.Cos(ssAngle)
	c3 := -decay * decay
	c1 := 1.0 - c2 - c3

	articleHpAngle := degreeToRadians(360.0 / float64(configuration.MaxPeriod))
	articlePole := (1.0 - math.Sin(articleHpAngle)) / math.Cos(articleHpAngle)
	bookHpAngle := degreeToRadians((1.0 / math.Sqrt2) * 360.0 / float64(configuration.MaxPeriod))
	alpha := (math.Cos(bookHpAngle) + math.Sin(bookHpAngle) - 1.0) / math.Cos(bookHpAngle)

	for stepIndex, price := range prices {
		var highValue float64
		if book {
			oneMinusHalfAlpha := 1.0 - alpha/2.0
			oneMinusAlpha := 1.0 - alpha
			highValue = oneMinusHalfAlpha*oneMinusHalfAlpha*(price-2.0*oldAt(prices, stepIndex-1)+oldAt(prices, stepIndex-2)) +
				2.0*oneMinusAlpha*oldAt(hpHistory, stepIndex-1) -
				oneMinusAlpha*oneMinusAlpha*oldAt(hpHistory, stepIndex-2)
		} else {
			highValue = 0.5*(1.0+articlePole)*(price-oldAt(prices, stepIndex-1)) +
				articlePole*oldAt(hpHistory, stepIndex-1)
		}
		hpHistory = append(hpHistory, highValue)

		filteredValue := c1*(highValue+oldAt(hpHistory, stepIndex-1))/2.0 +
			c2*oldAt(filteredHistory, stepIndex-1) +
			c3*oldAt(filteredHistory, stepIndex-2)
		filteredHistory = append(filteredHistory, filteredValue)

		for lag := 0; lag <= configuration.MaxPeriod; lag++ {
			windowSize := configuration.AveragingLength
			if windowSize == 0 {
				windowSize = lag
			}
			var sx, sy, sxx, syy, sxy float64
			for offset := range windowSize {
				xSample := oldAt(filteredHistory, stepIndex-offset)
				ySample := oldAt(filteredHistory, stepIndex-lag-offset)
				sx += xSample
				sy += ySample
				sxx += xSample * xSample
				sxy += xSample * ySample
				syy += ySample * ySample
			}
			windowFloat := float64(windowSize)
			denominator := (windowFloat*sxx - sx*sx) * (windowFloat*syy - sy*sy)
			if denominator > 0.0 {
				correlations[lag] = (windowFloat*sxy - sx*sy) / math.Sqrt(denominator)
			}
		}

		for binIndex := range binTotal {
			period := configuration.MinPeriod + binIndex
			var cosinePart, sinePart float64
			for lag := 3; lag <= configuration.MaxPeriod; lag++ {
				angle := degreeToRadians(360.0 * float64(lag) / float64(period))
				cosinePart += correlations[lag] * math.Cos(angle)
				sinePart += correlations[lag] * math.Sin(angle)
			}
			squaredSum := cosinePart*cosinePart + sinePart*sinePart
			powers[binIndex] = 0.2*squaredSum*squaredSum + 0.8*powers[binIndex]
		}

		if book {
			maximum = 0.995 * maximum
		} else {
			maximum = 0.0
		}
		for _, powerValue := range powers {
			if powerValue > maximum {
				maximum = powerValue
			}
		}

		normalized := make([]float64, binTotal)
		if maximum > 0.0 {
			for binIndex, powerValue := range powers {
				normalized[binIndex] = powerValue / maximum
			}
		}

		weights := slices.Clone(normalized)
		if !book && configuration.EnhanceResolution {
			for binIndex, normValue := range normalized {
				weights[binIndex] = normValue * normValue * normValue
			}
		}

		peak := slices.Max(weights)
		var spx, sp float64
		for binIndex, weight := range weights {
			if (book && weight >= 0.5) || (!book && peak >= 0.25 && weight >= 0.25) {
				spx += float64(configuration.MinPeriod+binIndex) * weight
				sp += weight
			}
		}

		candidate := 0.0
		if sp != 0.0 {
			candidate = spx / sp
		}
		if (book && sp == 0.0) || (!book && sp < 0.25) {
			candidate = dominant
		} else if sp != 0.0 {
			hasDominant = true
		}
		dominant = candidate

		display := slices.Clone(weights)
		if book && configuration.EnhanceResolution {
			for binIndex, normValue := range normalized {
				display[binIndex] = normValue * normValue
			}
		}

		steps = append(steps, oracleStep{
			highPass:     highValue,
			filtered:     filteredValue,
			correlations: slices.Clone(correlations),
			powers:       slices.Clone(powers),
			maximum:      maximum,
			dominant:     dominant,
			display:      display,
			hasDominant:  hasDominant,
		})
	}

	return steps
}

func TestReferenceFixturesAll64Cases(t *testing.T) {
	t.Parallel()

	cases := allReferenceCases()
	if len(cases) != 64 {
		t.Fatalf("len(cases) = %d, want 64", len(cases))
	}

	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			paramsCopy := tc.params
			indicator, err := NewAutoCorrelationPeriodogramParams(&paramsCopy)
			if err != nil {
				t.Fatalf("unexpected constructor error: %v", err)
			}

			effectiveWindow := tc.params.AveragingLength
			if effectiveWindow == 0 {
				effectiveWindow = tc.params.MaxPeriod
			}
			expectedWarmup := max(3, tc.params.MaxPeriod+effectiveWindow) - 1
			if indicator.WarmupBars() != expectedWarmup {
				t.Fatalf("WarmupBars() = %d, want %d", indicator.WarmupBars(), expectedWarmup)
			}

			bins := tc.params.MaxPeriod - tc.params.MinPeriod + 1
			for barIndex, sample := range testInput {
				sampleTime := baseTime.Add(time.Duration(barIndex) * time.Minute)
				result := indicator.Update(sample, sampleTime)
				if !result.Heatmap.Time.Equal(sampleTime) {
					t.Fatalf("[%d] heatmap time mismatch", barIndex)
				}
				if result.Heatmap.ParameterFirst != float64(tc.params.MinPeriod) ||
					result.Heatmap.ParameterLast != float64(tc.params.MaxPeriod) ||
					result.Heatmap.ParameterResolution != 1.0 {
					t.Fatalf("[%d] unexpected heatmap axis: %+v", barIndex, result.Heatmap)
				}

				primed := barIndex >= indicator.WarmupBars()
				if indicator.IsPrimed() != primed {
					t.Fatalf("[%d] IsPrimed() = %v, want %v", barIndex, indicator.IsPrimed(), primed)
				}

				assertClose(t, result.DominantCycle, tc.expectedDC[barIndex], fixtureTolerance, "dc")
				assertClose(t, result.Heatmap.ValueMin, tc.expectedMin[barIndex], fixtureTolerance, "min")
				assertClose(t, result.Heatmap.ValueMax, tc.expectedMax[barIndex], fixtureTolerance, "max")

				if !primed {
					if !result.Heatmap.IsEmpty() {
						t.Fatalf("[%d] expected empty heatmap before warmup, got len=%d", barIndex, len(result.Heatmap.Values))
					}
					for _, expectedCell := range tc.expectedHeatmap[barIndex] {
						if !math.IsNaN(expectedCell) {
							t.Fatalf("[%d] expected NaN fixture before warmup", barIndex)
						}
					}
					continue
				}

				if len(result.Heatmap.Values) != bins {
					t.Fatalf("[%d] len(Values) = %d, want %d", barIndex, len(result.Heatmap.Values), bins)
				}
				for binIndex := range bins {
					assertClose(t, result.Heatmap.Values[binIndex], tc.expectedHeatmap[barIndex][binIndex], fixtureTolerance, "heatmap")
				}
			}
		})
	}
}

func TestIndependentDegreeBasedSourceTranscription(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, tc := range allReferenceCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			paramsCopy := tc.params
			indicator, err := NewAutoCorrelationPeriodogramParams(&paramsCopy)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			reference := sourceReference(testInput, tc.params)
			for barIndex, sample := range testInput {
				result := indicator.Update(sample, baseTime.Add(time.Duration(barIndex)*time.Minute))
				expected := reference[barIndex]
				state := indicator.State()

				assertClose(t, state.HighPass, expected.highPass, oracleTolerance, "state.hp")
				assertClose(t, state.Filtered, expected.filtered, oracleTolerance, "state.filtered")
				assertClose(t, state.NormalizationMaximum, expected.maximum, oracleTolerance, "state.maximum")
				assertClose(t, state.InternalDominantCycle, expected.dominant, oracleTolerance, "state.dominant")
				for lag := range expected.correlations {
					assertClose(t, state.Correlations[lag], expected.correlations[lag], oracleTolerance, "state.corr")
				}
				for binIndex := range expected.powers {
					assertClose(t, state.SmoothedPower[binIndex], expected.powers[binIndex], oracleTolerance, "state.power")
				}

				primed := barIndex >= indicator.WarmupBars()
				if primed {
					for binIndex := range expected.display {
						assertClose(t, result.Heatmap.Values[binIndex], expected.display[binIndex], oracleTolerance, "display")
					}
				} else if !result.Heatmap.IsEmpty() {
					t.Fatalf("[%d] expected empty heatmap", barIndex)
				}

				expectedDC := math.NaN()
				if primed && expected.hasDominant {
					expectedDC = expected.dominant
				}
				assertClose(t, result.DominantCycle, expectedDC, oracleTolerance, "dc")
			}
		})
	}
}

func TestEnhancementOrderAndBookHold(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	runSeries := func(source Source, enhanced bool) []PeriodogramOutput {
		params := DefaultParams()
		params.Source = source
		params.EnhanceResolution = enhanced
		indicator, err := NewAutoCorrelationPeriodogramParams(params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		series := make([]PeriodogramOutput, len(testInput))
		for barIndex, sample := range testInput {
			series[barIndex] = indicator.Update(sample, baseTime.Add(time.Duration(barIndex)*time.Minute))
		}
		return series
	}

	articlePlain := runSeries(Article, false)
	articleEnhanced := runSeries(Article, true)
	bookPlain := runSeries(Book, false)
	bookEnhanced := runSeries(Book, true)

	for barIndex := 50; barIndex < len(testInput); barIndex++ {
		assertClose(t, bookPlain[barIndex].DominantCycle, bookEnhanced[barIndex].DominantCycle, exactTolerance, "book_dc_unchanged")
		for binIndex := range bookPlain[barIndex].Heatmap.Values {
			plainCell := bookPlain[barIndex].Heatmap.Values[binIndex]
			assertClose(t, bookEnhanced[barIndex].Heatmap.Values[binIndex], plainCell*plainCell, exactTolerance, "book_square")
		}
		assertClose(t, articlePlain[barIndex].Heatmap.ValueMax, 1.0, exactTolerance, "article_max_one")
		for binIndex := range articlePlain[barIndex].Heatmap.Values {
			plainCell := articlePlain[barIndex].Heatmap.Values[binIndex]
			assertClose(t, articleEnhanced[barIndex].Heatmap.Values[binIndex], plainCell*plainCell*plainCell, exactTolerance, "article_cube")
		}
	}

	if math.Abs(articlePlain[len(testInput)-1].DominantCycle-articleEnhanced[len(testInput)-1].DominantCycle) < 1e-9 {
		t.Fatalf("expected article enhancement to alter dominant cycle")
	}

	holdCount := 0
	for barIndex := 51; barIndex < len(testInput); barIndex++ {
		if bookPlain[barIndex].Heatmap.ValueMax < 0.5 {
			holdCount++
			assertClose(t, bookPlain[barIndex].DominantCycle, bookPlain[barIndex-1].DominantCycle, exactTolerance, "book_hold")
		}
	}
	if holdCount == 0 {
		t.Fatalf("expected at least one book hold bar on testInput")
	}
}

func TestZeroPowerAndRetainedUndefinedCorrelation(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, source := range []Source{Article, Book} {
		params := DefaultParams()
		params.Source = source
		indicator, err := NewAutoCorrelationPeriodogramParams(params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var result PeriodogramOutput
		for stepIndex := range 100 {
			result = indicator.Update(0.0, baseTime.Add(time.Duration(stepIndex)*time.Minute))
		}
		if !indicator.IsPrimed() {
			t.Fatalf("expected indicator to be primed")
		}
		if !math.IsNaN(result.DominantCycle) {
			t.Fatalf("expected NaN dominant cycle on all-zero input, got %v", result.DominantCycle)
		}
		for binIndex, cell := range result.Heatmap.Values {
			if cell != 0.0 {
				t.Fatalf("expected zero cell at %d, got %v", binIndex, cell)
			}
		}

		seededIndicator, err := NewAutoCorrelationPeriodogramParams(params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for stepIndex := range 60 {
			seededIndicator.Update(testInput[stepIndex], baseTime.Add(time.Duration(stepIndex)*time.Minute))
		}
		beforeZeroRun := seededIndicator.State().Correlations[12]
		if beforeZeroRun == 0.0 {
			t.Fatalf("expected non-zero correlation at lag 12 after varying input")
		}
		seededIndicator.correlations[12] = 0.75
		seededIndicator.filtered = make([]float64, seededIndicator.historyLength)
		seededIndicator.close1 = 0.0
		seededIndicator.close2 = 0.0
		seededIndicator.high1 = 0.0
		seededIndicator.high2 = 0.0
		seededIndicator.Update(0.0, baseTime.Add(61*time.Minute))
		if got := seededIndicator.State().Correlations[12]; got != 0.75 {
			t.Fatalf("expected retained correlation 0.75 on zero denominator, got %v", got)
		}
	}
}

func TestSinePeakAt20(t *testing.T) {
	t.Parallel()

	indicator, err := NewAutoCorrelationPeriodogramDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	var result PeriodogramOutput
	for stepIndex := range 1000 {
		sample := 100.0 + math.Sin(2.0*math.Pi*float64(stepIndex)/20.0)
		result = indicator.Update(sample, baseTime.Add(time.Duration(stepIndex)*time.Minute))
	}
	bestBin := 0
	for binIndex, cell := range result.Heatmap.Values {
		if cell > result.Heatmap.Values[bestBin] {
			bestBin = binIndex
		}
	}
	peakPeriod := result.Heatmap.ParameterFirst + float64(bestBin)
	if peakPeriod != 20.0 {
		t.Fatalf("peakPeriod = %v, want 20.0", peakPeriod)
	}
	if !(result.DominantCycle > 18.0 && result.DominantCycle < 22.0) {
		t.Fatalf("DominantCycle = %v, expected in (18, 22)", result.DominantCycle)
	}
}

func TestSkippedNaNOverflowAtomicityResetAndImmutability(t *testing.T) {
	t.Parallel()

	indicator, err := NewAutoCorrelationPeriodogramDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	firstRun := make([]PeriodogramOutput, len(testInput))
	for barIndex, sample := range testInput {
		firstRun[barIndex] = indicator.Update(sample, baseTime.Add(time.Duration(barIndex)*time.Minute))
	}

	snapshotState := indicator.State()
	snapshotOutput := indicator.Output()
	snapshotCount := indicator.SamplesSeen()

	nanResult := indicator.Update(math.NaN(), baseTime.Add(300*time.Minute))
	if !nanResult.Heatmap.IsEmpty() || !math.IsNaN(nanResult.DominantCycle) {
		t.Fatalf("expected empty/NaN output on skipped NaN")
	}
	if indicator.SamplesSeen() != snapshotCount {
		t.Fatalf("SamplesSeen changed on NaN: got %d, want %d", indicator.SamplesSeen(), snapshotCount)
	}
	afterNaNState := indicator.State()
	assertClose(t, afterNaNState.HighPass, snapshotState.HighPass, 0, "nan.hp")
	assertClose(t, afterNaNState.Filtered, snapshotState.Filtered, 0, "nan.filtered")
	assertClose(t, indicator.Output().DominantCycle, snapshotOutput.DominantCycle, 0, "nan.output_dc")

	for _, badSample := range []float64{math.Inf(1), math.Inf(-1), 1e308} {
		if _, err := indicator.TryUpdate(badSample, baseTime.Add(301*time.Minute)); err == nil {
			t.Fatalf("expected TryUpdate error for %v", badSample)
		}
		didPanic := false
		func() {
			defer func() {
				if recover() != nil {
					didPanic = true
				}
			}()
			indicator.Update(badSample, baseTime.Add(301*time.Minute))
		}()
		if !didPanic {
			t.Fatalf("expected Update panic for %v", badSample)
		}
		if indicator.SamplesSeen() != snapshotCount {
			t.Fatalf("SamplesSeen mutated after failed update")
		}
		currentState := indicator.State()
		assertClose(t, currentState.HighPass, snapshotState.HighPass, 0, "overflow.hp")
		assertClose(t, currentState.Filtered, snapshotState.Filtered, 0, "overflow.filtered")
		assertClose(t, currentState.NormalizationMaximum, snapshotState.NormalizationMaximum, 0, "overflow.max")
		assertClose(t, currentState.InternalDominantCycle, snapshotState.InternalDominantCycle, 0, "overflow.dc")
	}

	indicator.Reset()
	if indicator.SamplesSeen() != 0 || indicator.IsPrimed() {
		t.Fatalf("Reset did not clear sample count / primed flag")
	}
	for barIndex, sample := range testInput {
		replayed := indicator.Update(sample, baseTime.Add(time.Duration(barIndex)*time.Minute))
		assertClose(t, replayed.DominantCycle, firstRun[barIndex].DominantCycle, 0, "replay.dc")
		if len(replayed.Heatmap.Values) != len(firstRun[barIndex].Heatmap.Values) {
			t.Fatalf("[%d] replay heatmap length mismatch", barIndex)
		}
		for binIndex := range replayed.Heatmap.Values {
			assertClose(t, replayed.Heatmap.Values[binIndex], firstRun[barIndex].Heatmap.Values[binIndex], 0, "replay.heatmap")
		}
	}

	latest := indicator.Output()
	originalZero := latest.Heatmap.Values[0]
	latest.Heatmap.Values[0] = -999.0
	if indicator.Output().Heatmap.Values[0] != originalZero {
		t.Fatalf("mutating Output() slice leaked into internal state")
	}
}

func TestSourceEnumJSONAndMethods(t *testing.T) {
	t.Parallel()

	if Article.String() != "article" || !Article.IsKnown() {
		t.Fatalf("unexpected Article String/IsKnown")
	}
	if Book.String() != "book" || !Book.IsKnown() {
		t.Fatalf("unexpected Book String/IsKnown")
	}
	if Source(99).String() != "unknown" || Source(99).IsKnown() || Source(-1).IsKnown() {
		t.Fatalf("unexpected unknown Source behavior")
	}

	encodedArticle, err := json.Marshal(Article)
	if err != nil || string(encodedArticle) != `"article"` {
		t.Fatalf("Marshal(Article) = %s, err=%v", string(encodedArticle), err)
	}
	encodedBook, err := json.Marshal(Book)
	if err != nil || string(encodedBook) != `"book"` {
		t.Fatalf("Marshal(Book) = %s, err=%v", string(encodedBook), err)
	}
	if _, err := json.Marshal(Source(99)); err == nil {
		t.Fatalf("expected Marshal error for unknown Source")
	}

	var decoded Source
	if err := json.Unmarshal([]byte(`"book"`), &decoded); err != nil || decoded != Book {
		t.Fatalf("Unmarshal book failed: %v, got %v", err, decoded)
	}
	if err := json.Unmarshal([]byte(`"article"`), &decoded); err != nil || decoded != Article {
		t.Fatalf("Unmarshal article failed: %v, got %v", err, decoded)
	}
	for _, invalidPayload := range []string{`"literal370"`, `123`, `true`} {
		if err := json.Unmarshal([]byte(invalidPayload), &decoded); err == nil {
			t.Fatalf("expected Unmarshal error for %s", invalidPayload)
		}
	}
}

func TestMetadataMnemonicsEntitiesAndValidation(t *testing.T) {
	t.Parallel()

	defaultIndicator, err := NewAutoCorrelationPeriodogramDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	meta := defaultIndicator.Metadata()
	if meta.Identifier != core.AutoCorrelationPeriodogram {
		t.Fatalf("Identifier = %v, want %v", meta.Identifier, core.AutoCorrelationPeriodogram)
	}
	if meta.Mnemonic != "acp(8, 48, 3)" {
		t.Fatalf("Mnemonic = %q, want %q", meta.Mnemonic, "acp(8, 48, 3)")
	}
	if len(meta.Outputs) != 2 {
		t.Fatalf("len(Outputs) = %d, want 2", len(meta.Outputs))
	}
	if meta.Outputs[0].Shape != shape.Heatmap || meta.Outputs[0].Mnemonic != "acp(8, 48, 3)" {
		t.Fatalf("unexpected primary output metadata: %+v", meta.Outputs[0])
	}
	if meta.Outputs[1].Shape != shape.Scalar || meta.Outputs[1].Mnemonic != "acp-dc(8, 48, 3)" {
		t.Fatalf("unexpected secondary output metadata: %+v", meta.Outputs[1])
	}

	bookEnhancedParams := DefaultParams()
	bookEnhancedParams.Source = Book
	bookEnhancedParams.EnhanceResolution = true
	bookEnhanced, err := NewAutoCorrelationPeriodogram(bookEnhancedParams)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bookEnhanced.Parameters().MinPeriod != 10 {
		t.Fatalf("resolved book MinPeriod = %d, want 10", bookEnhanced.Parameters().MinPeriod)
	}
	if bookEnhanced.Metadata().Mnemonic != "acp(10, 48, 3, book, enh)" {
		t.Fatalf("book enhanced mnemonic = %q", bookEnhanced.Metadata().Mnemonic)
	}
	if bookEnhanced.Metadata().Outputs[1].Mnemonic != "acp-dc(10, 48, 3, book)" {
		t.Fatalf("book enhanced dc mnemonic = %q", bookEnhanced.Metadata().Outputs[1].Mnemonic)
	}

	articleEnhancedParams := DefaultParams()
	articleEnhancedParams.EnhanceResolution = true
	articleEnhanced, err := NewAutoCorrelationPeriodogram(articleEnhancedParams)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if articleEnhanced.Metadata().Mnemonic != "acp(8, 48, 3, enh)" ||
		articleEnhanced.Metadata().Outputs[1].Mnemonic != "acp-dc(8, 48, 3, enh)" {
		t.Fatalf("unexpected article enhanced mnemonics: %+v", articleEnhanced.Metadata())
	}

	combinations := []struct {
		barComp   entities.BarComponent
		quoteComp entities.QuoteComponent
		tradeComp entities.TradeComponent
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
	for _, combo := range combinations {
		params := DefaultParams()
		params.BarComponent = combo.barComp
		params.QuoteComponent = combo.quoteComp
		params.TradeComponent = combo.tradeComp
		instance, err := NewAutoCorrelationPeriodogramParams(params)
		if err != nil {
			t.Fatalf("unexpected error for combo %+v: %v", combo, err)
		}
		resolvedBar := combo.barComp
		if resolvedBar == 0 {
			resolvedBar = entities.DefaultBarComponent
		}
		resolvedQuote := combo.quoteComp
		if resolvedQuote == 0 {
			resolvedQuote = entities.DefaultQuoteComponent
		}
		resolvedTrade := combo.tradeComp
		if resolvedTrade == 0 {
			resolvedTrade = entities.DefaultTradeComponent
		}
		suffix := core.ComponentTripleMnemonic(resolvedBar, resolvedQuote, resolvedTrade)
		wantPrimary := "acp(8, 48, 3" + suffix + ")"
		wantDC := "acp-dc(8, 48, 3" + suffix + ")"
		if instance.Metadata().Mnemonic != wantPrimary || params.Mnemonic() != wantPrimary {
			t.Fatalf("got primary mnemonic %q, want %q", instance.Metadata().Mnemonic, wantPrimary)
		}
		if instance.Metadata().Outputs[1].Mnemonic != wantDC || params.DominantCycleMnemonic() != wantDC {
			t.Fatalf("got dc mnemonic %q, want %q", instance.Metadata().Outputs[1].Mnemonic, wantDC)
		}
	}

	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	barInd, _ := NewAutoCorrelationPeriodogramDefault()
	scalarInd, _ := NewAutoCorrelationPeriodogramDefault()
	quoteInd, _ := NewAutoCorrelationPeriodogramDefault()
	tradeInd, _ := NewAutoCorrelationPeriodogramDefault()
	for barIndex := range 60 {
		sampleTime := baseTime.AddDate(0, 0, barIndex)
		price := testInput[barIndex]
		outBar := barInd.UpdateBar(&entities.Bar{
			Time:   sampleTime,
			Open:   testOpen[barIndex],
			High:   testHigh[barIndex],
			Low:    testLow[barIndex],
			Close:  price,
			Volume: 1000,
		})
		outScalar := scalarInd.UpdateScalar(&entities.Scalar{Time: sampleTime, Value: price})
		outQuote := quoteInd.UpdateQuote(&entities.Quote{Time: sampleTime, Bid: price - 0.5, Ask: price + 0.5, BidSize: 10, AskSize: 10})
		outTrade := tradeInd.UpdateTrade(&entities.Trade{Time: sampleTime, Price: price, Volume: 10})
		if len(outBar) != 2 {
			t.Fatalf("len(outBar) = %d, want 2", len(outBar))
		}
		dcBar := outBar[1].(entities.Scalar).Value
		dcScalar := outScalar[1].(entities.Scalar).Value
		dcQuote := outQuote[1].(entities.Scalar).Value
		dcTrade := outTrade[1].(entities.Scalar).Value
		assertClose(t, dcBar, dcScalar, exactTolerance, "bar_vs_scalar")
		assertClose(t, dcBar, dcQuote, exactTolerance, "bar_vs_quote")
		assertClose(t, dcBar, dcTrade, exactTolerance, "bar_vs_trade")
	}

	invalidParams := []Params{
		{Source: Source(99), MaxPeriod: 48, AveragingLength: 3},
		{Source: Article, MinPeriod: 2, MaxPeriod: 48, AveragingLength: 3},
		{Source: Article, MinPeriod: -1, MaxPeriod: 48, AveragingLength: 3},
		{Source: Article, MinPeriod: 3, MaxPeriod: 4, AveragingLength: 3},
		{Source: Article, MinPeriod: 8, MaxPeriod: 8, AveragingLength: 3},
		{Source: Book, MinPeriod: 0, MaxPeriod: 10, AveragingLength: 3},
		{Source: Article, MinPeriod: 8, MaxPeriod: 48, AveragingLength: 1},
		{Source: Article, MinPeriod: 8, MaxPeriod: 48, AveragingLength: -1},
		{Source: Article, MinPeriod: 8, MaxPeriod: 48, AveragingLength: 3, BarComponent: entities.BarComponent(999)},
		{Source: Article, MinPeriod: 8, MaxPeriod: 48, AveragingLength: 3, QuoteComponent: entities.QuoteComponent(999)},
		{Source: Article, MinPeriod: 8, MaxPeriod: 48, AveragingLength: 3, TradeComponent: entities.TradeComponent(999)},
	}
	for caseIndex, bad := range invalidParams {
		badCopy := bad
		if _, err := NewAutoCorrelationPeriodogramParams(&badCopy); err == nil {
			t.Fatalf("expected error for invalidParams[%d]: %+v", caseIndex, bad)
		}
	}
}
