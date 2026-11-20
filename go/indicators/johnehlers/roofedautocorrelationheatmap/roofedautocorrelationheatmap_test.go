//nolint:testpackage
package roofedautocorrelationheatmap

import (
	"math"
	"slices"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
	"zpano/indicators/core/outputs/shape"
)

const (
	fixtureTolerance   = 2e-11
	a2FixtureTolerance = 1e-10
	oracleTolerance    = 1e-13
)

type referenceCase struct {
	name            string
	params          Params
	expectedHeatmap [][]float64
	expectedMin     []float64
	expectedMax     []float64
}

func referenceCases() []referenceCase {
	return []referenceCase{
		{
			name:            "L03_L148_A0_HP48_LP10",
			params:          Params{MinLag: 3, MaxLag: 48, AveragingLength: 0, HighpassPeriod: 48.0, SmoothingPeriod: 10.0},
			expectedHeatmap: expectedL03_L148_A0_HP48_LP10_HEATMAP,
			expectedMin:     expectedL03_L148_A0_HP48_LP10_MIN,
			expectedMax:     expectedL03_L148_A0_HP48_LP10_MAX,
		},
		{
			name:            "L03_L148_A2_HP48_LP10",
			params:          Params{MinLag: 3, MaxLag: 48, AveragingLength: 2, HighpassPeriod: 48.0, SmoothingPeriod: 10.0},
			expectedHeatmap: expectedL03_L148_A2_HP48_LP10_HEATMAP,
			expectedMin:     expectedL03_L148_A2_HP48_LP10_MIN,
			expectedMax:     expectedL03_L148_A2_HP48_LP10_MAX,
		},
		{
			name:            "L03_L148_A3_HP48_LP10",
			params:          Params{MinLag: 3, MaxLag: 48, AveragingLength: 3, HighpassPeriod: 48.0, SmoothingPeriod: 10.0},
			expectedHeatmap: expectedL03_L148_A3_HP48_LP10_HEATMAP,
			expectedMin:     expectedL03_L148_A3_HP48_LP10_MIN,
			expectedMax:     expectedL03_L148_A3_HP48_LP10_MAX,
		},
		{
			name:            "L03_L148_A20_HP48_LP10",
			params:          Params{MinLag: 3, MaxLag: 48, AveragingLength: 20, HighpassPeriod: 48.0, SmoothingPeriod: 10.0},
			expectedHeatmap: expectedL03_L148_A20_HP48_LP10_HEATMAP,
			expectedMin:     expectedL03_L148_A20_HP48_LP10_MIN,
			expectedMax:     expectedL03_L148_A20_HP48_LP10_MAX,
		},
		{
			name:            "L03_L148_A100_HP48_LP10",
			params:          Params{MinLag: 3, MaxLag: 48, AveragingLength: 100, HighpassPeriod: 48.0, SmoothingPeriod: 10.0},
			expectedHeatmap: expectedL03_L148_A100_HP48_LP10_HEATMAP,
			expectedMin:     expectedL03_L148_A100_HP48_LP10_MIN,
			expectedMax:     expectedL03_L148_A100_HP48_LP10_MAX,
		},
		{
			name:            "L03_L148_A1_HP48_LP10",
			params:          Params{MinLag: 3, MaxLag: 48, AveragingLength: 1, HighpassPeriod: 48.0, SmoothingPeriod: 10.0},
			expectedHeatmap: expectedL03_L148_A1_HP48_LP10_HEATMAP,
			expectedMin:     expectedL03_L148_A1_HP48_LP10_MIN,
			expectedMax:     expectedL03_L148_A1_HP48_LP10_MAX,
		},
		{
			name:            "L03_L148_A0_HP48_LP10_5",
			params:          Params{MinLag: 3, MaxLag: 48, AveragingLength: 0, HighpassPeriod: 48.0, SmoothingPeriod: 10.5},
			expectedHeatmap: expectedL03_L148_A0_HP48_LP10_5_HEATMAP,
			expectedMin:     expectedL03_L148_A0_HP48_LP10_5_MIN,
			expectedMax:     expectedL03_L148_A0_HP48_LP10_5_MAX,
		},
		{
			name:            "L03_L148_A0_HP48_LP20",
			params:          Params{MinLag: 3, MaxLag: 48, AveragingLength: 0, HighpassPeriod: 48.0, SmoothingPeriod: 20.0},
			expectedHeatmap: expectedL03_L148_A0_HP48_LP20_HEATMAP,
			expectedMin:     expectedL03_L148_A0_HP48_LP20_MIN,
			expectedMax:     expectedL03_L148_A0_HP48_LP20_MAX,
		},
		{
			name:            "L03_L148_A0_HP80_LP10",
			params:          Params{MinLag: 3, MaxLag: 48, AveragingLength: 0, HighpassPeriod: 80.0, SmoothingPeriod: 10.0},
			expectedHeatmap: expectedL03_L148_A0_HP80_LP10_HEATMAP,
			expectedMin:     expectedL03_L148_A0_HP80_LP10_MIN,
			expectedMax:     expectedL03_L148_A0_HP80_LP10_MAX,
		},
		{
			name:            "L03_L124_A0_HP24_LP10",
			params:          Params{MinLag: 3, MaxLag: 24, AveragingLength: 0, HighpassPeriod: 24.0, SmoothingPeriod: 10.0},
			expectedHeatmap: expectedL03_L124_A0_HP24_LP10_HEATMAP,
			expectedMin:     expectedL03_L124_A0_HP24_LP10_MIN,
			expectedMax:     expectedL03_L124_A0_HP24_LP10_MAX,
		},
		{
			name:            "L012_L112_A12_HP48_LP10",
			params:          Params{MinLag: 12, MaxLag: 12, AveragingLength: 12, HighpassPeriod: 48.0, SmoothingPeriod: 10.0},
			expectedHeatmap: expectedL012_L112_A12_HP48_LP10_HEATMAP,
			expectedMin:     expectedL012_L112_A12_HP48_LP10_MIN,
			expectedMax:     expectedL012_L112_A12_HP48_LP10_MAX,
		},
		{
			name:            "L00_L10_A3_HP48_LP10",
			params:          Params{MinLag: 0, MaxLag: 0, AveragingLength: 3, HighpassPeriod: 48.0, SmoothingPeriod: 10.0},
			expectedHeatmap: expectedL00_L10_A3_HP48_LP10_HEATMAP,
			expectedMin:     expectedL00_L10_A3_HP48_LP10_MIN,
			expectedMax:     expectedL00_L10_A3_HP48_LP10_MAX,
		},
	}
}

type oracleRow struct {
	highPass float64
	filtered float64
	bins     []float64
}

func batchOracle(closes []float64, p Params) []oracleRow {
	angle := 2.0 * math.Pi / (math.Sqrt2 * p.HighpassPeriod)
	cosine := math.Cos(angle)
	alpha := (cosine + math.Sin(angle) - 1.0) / cosine
	pole := 1.0 - alpha
	attenuation := math.Exp(-math.Sqrt2 * math.Pi / p.SmoothingPeriod)
	hp0 := (1.0 - alpha/2.0) * (1.0 - alpha/2.0)
	hp1 := 2.0 * pole
	hp2 := pole * pole
	ss2 := 2.0 * attenuation * math.Cos(math.Sqrt2*math.Pi/p.SmoothingPeriod)
	ss3 := -attenuation * attenuation
	ss1 := 1.0 - ss2 - ss3

	effectiveWindow := p.AveragingLength
	if effectiveWindow == 0 {
		effectiveWindow = p.MaxLag
	}
	historyLength := p.MaxLag + effectiveWindow
	if historyLength < 1 {
		historyLength = 1
	}

	highSeries := make([]float64, len(closes))
	filteredSeries := make([]float64, len(closes))
	correlations := make([]float64, p.MaxLag+1)
	rows := make([]oracleRow, len(closes))

	for barIndex, price := range closes {
		var x1, x2, h1, h2, f1, f2 float64
		if barIndex >= 1 {
			x1 = closes[barIndex-1]
			h1 = highSeries[barIndex-1]
			f1 = filteredSeries[barIndex-1]
		}
		if barIndex >= 2 {
			x2 = closes[barIndex-2]
			h2 = highSeries[barIndex-2]
			f2 = filteredSeries[barIndex-2]
		}
		high := float64(float64(hp0*(price-float64(2.0*x1)+x2))+float64(hp1*h1)) - float64(hp2*h2)
		filtered := float64(float64(ss1*(high+h1)/2.0)+float64(ss2*f1)) + float64(ss3*f2)
		highSeries[barIndex] = high
		filteredSeries[barIndex] = filtered

		for lag := range p.MaxLag + 1 {
			window := p.AveragingLength
			if window == 0 {
				window = lag
			}
			windowFloat := float64(window)
			var sumX, sumY, sumXX, sumXY, sumYY float64
			for offset := range window {
				var sampleX, sampleY float64
				if barIndex-offset >= 0 {
					sampleX = filteredSeries[barIndex-offset]
				}
				if barIndex-lag-offset >= 0 {
					sampleY = filteredSeries[barIndex-lag-offset]
				}
				sumX += sampleX
				sumY += sampleY
				sumXX += float64(sampleX * sampleX)
				sumXY += float64(sampleX * sampleY)
				sumYY += float64(sampleY * sampleY)
			}
			varX := float64(windowFloat*sumXX) - float64(sumX*sumX)
			varY := float64(windowFloat*sumYY) - float64(sumY*sumY)
			denominator := float64(varX * varY)
			if denominator > 0.0 {
				numerator := float64(windowFloat*sumXY) - float64(sumX*sumY)
				correlations[lag] = numerator / math.Sqrt(denominator)
			}
			correlations[lag] = float64(0.5 * float64(correlations[lag]+1.0))
		}

		var bins []float64
		if barIndex+1 >= historyLength {
			bins = slices.Clone(correlations[p.MinLag : p.MaxLag+1])
		}
		rows[barIndex] = oracleRow{
			highPass: high,
			filtered: filtered,
			bins:     bins,
		}
	}

	return rows
}

func almostEqual(left, right, tolerance float64) bool {
	if math.IsNaN(left) && math.IsNaN(right) {
		return true
	}
	if math.IsNaN(left) || math.IsNaN(right) {
		return false
	}
	bound := math.Max(tolerance, math.Abs(right)*tolerance)
	return math.Abs(left-right) <= bound
}

func TestRoofedAutoCorrelationHeatmapReferenceCases(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	for _, tc := range referenceCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			paramsCopy := tc.params
			indicator, err := NewRoofedAutoCorrelationHeatmap(&paramsCopy)
			if err != nil {
				t.Fatalf("NewRoofedAutoCorrelationHeatmap failed: %v", err)
			}
			oracle := batchOracle(testInput, tc.params)
			expectedBins := tc.params.MaxLag - tc.params.MinLag + 1
			caseTolerance := fixtureTolerance
			if tc.params.AveragingLength == 2 {
				caseTolerance = a2FixtureTolerance
			}

			for barIndex, closePrice := range testInput {
				stamp := baseTime.AddDate(0, 0, barIndex)
				column := indicator.Update(closePrice, stamp)
				state := indicator.State()

				if !almostEqual(state.HighPass, oracle[barIndex].highPass, oracleTolerance) {
					t.Fatalf("bar %d HighPass = %.16g, want %.16g", barIndex, state.HighPass, oracle[barIndex].highPass)
				}
				if !almostEqual(state.Filtered, oracle[barIndex].filtered, oracleTolerance) {
					t.Fatalf("bar %d Filtered = %.16g, want %.16g", barIndex, state.Filtered, oracle[barIndex].filtered)
				}
				if column.ParameterFirst != float64(tc.params.MinLag) ||
					column.ParameterLast != float64(tc.params.MaxLag) ||
					column.ParameterResolution != 1.0 {
					t.Fatalf("bar %d axis mismatch: (%v, %v, %v)", barIndex, column.ParameterFirst, column.ParameterLast, column.ParameterResolution)
				}

				expectedRow := tc.expectedHeatmap[barIndex]
				expectedMin := tc.expectedMin[barIndex]
				expectedMax := tc.expectedMax[barIndex]
				if math.IsNaN(expectedMin) {
					if !column.IsEmpty() || !math.IsNaN(column.ValueMin) || !math.IsNaN(column.ValueMax) {
						t.Fatalf("bar %d expected empty warmup column, got %+v", barIndex, column)
					}
					for _, item := range expectedRow {
						if !math.IsNaN(item) {
							t.Fatalf("bar %d expected NaN row in fixture", barIndex)
						}
					}
					continue
				}

				if len(column.Values) != expectedBins {
					t.Fatalf("bar %d len(Values) = %d, want %d", barIndex, len(column.Values), expectedBins)
				}
				if !almostEqual(column.ValueMin, expectedMin, caseTolerance) {
					t.Fatalf("bar %d ValueMin = %.16g, want %.16g", barIndex, column.ValueMin, expectedMin)
				}
				if !almostEqual(column.ValueMax, expectedMax, caseTolerance) {
					t.Fatalf("bar %d ValueMax = %.16g, want %.16g", barIndex, column.ValueMax, expectedMax)
				}
				for binIndex, actual := range column.Values {
					if !almostEqual(actual, expectedRow[binIndex], caseTolerance) {
						t.Fatalf("bar %d bin %d = %.16g, want fixture %.16g", barIndex, binIndex, actual, expectedRow[binIndex])
					}
					if !almostEqual(actual, oracle[barIndex].bins[binIndex], oracleTolerance) {
						t.Fatalf("bar %d bin %d = %.16g, want oracle %.16g", barIndex, binIndex, actual, oracle[barIndex].bins[binIndex])
					}
				}
			}
		})
	}
}

func TestRoofedAutoCorrelationHeatmapMnemonicAndMetadata(t *testing.T) {
	t.Parallel()

	mnemonicTests := []struct {
		name   string
		params Params
		want   string
	}{
		{
			name:   "default components omitted",
			params: *DefaultParams(),
			want:   "rach(3, 48, 0, 48, 10)",
		},
		{
			name: "bar component only",
			params: Params{
				MinLag:          3,
				MaxLag:          48,
				AveragingLength: 0,
				HighpassPeriod:  48.0,
				SmoothingPeriod: 10.0,
				BarComponent:    entities.BarMedianPrice,
			},
			want: "rach(3, 48, 0, 48, 10, hl/2)",
		},
		{
			name: "quote component only",
			params: Params{
				MinLag:          3,
				MaxLag:          48,
				AveragingLength: 0,
				HighpassPeriod:  48.0,
				SmoothingPeriod: 10.0,
				QuoteComponent:  entities.QuoteWeightedPrice,
			},
			want: "rach(3, 48, 0, 48, 10, (bbs+aas)/(bs+as))",
		},
		{
			name: "trade component only",
			params: Params{
				MinLag:          3,
				MaxLag:          48,
				AveragingLength: 0,
				HighpassPeriod:  48.0,
				SmoothingPeriod: 10.0,
				TradeComponent:  entities.TradeVolume,
			},
			want: "rach(3, 48, 0, 48, 10, v)",
		},
		{
			name: "bar and quote components",
			params: Params{
				MinLag:          0,
				MaxLag:          0,
				AveragingLength: 3,
				HighpassPeriod:  48.0,
				SmoothingPeriod: 10.5,
				BarComponent:    entities.BarTypicalPrice,
				QuoteComponent:  entities.QuoteBidPrice,
			},
			want: "rach(0, 0, 3, 48, 10.5, hlc/3, b)",
		},
		{
			name: "quote and trade components",
			params: Params{
				MinLag:          12,
				MaxLag:          24,
				AveragingLength: 12,
				HighpassPeriod:  24.0,
				SmoothingPeriod: 10.0,
				QuoteComponent:  entities.QuoteAskPrice,
				TradeComponent:  entities.TradeVolume,
			},
			want: "rach(12, 24, 12, 24, 10, a, v)",
		},
	}

	for _, tc := range mnemonicTests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.params.Mnemonic(); got != tc.want {
				t.Fatalf("Mnemonic() = %q, want %q", got, tc.want)
			}
		})
	}

	defaultIndicator, err := NewRoofedAutoCorrelationHeatmapDefault()
	if err != nil {
		t.Fatalf("NewRoofedAutoCorrelationHeatmapDefault failed: %v", err)
	}
	metadata := defaultIndicator.Metadata()
	if metadata.Identifier != core.RoofedAutoCorrelationHeatmap {
		t.Fatalf("Metadata().Identifier = %v, want %v", metadata.Identifier, core.RoofedAutoCorrelationHeatmap)
	}
	if len(metadata.Outputs) != 1 {
		t.Fatalf("len(Metadata().Outputs) = %d, want 1", len(metadata.Outputs))
	}
	if metadata.Outputs[0].Shape != shape.Heatmap {
		t.Fatalf("Outputs[0].Shape = %v, want %v", metadata.Outputs[0].Shape, shape.Heatmap)
	}
	if metadata.Outputs[0].Mnemonic != "rach(3, 48, 0, 48, 10)" {
		t.Fatalf("Outputs[0].Mnemonic = %q, want %q", metadata.Outputs[0].Mnemonic, "rach(3, 48, 0, 48, 10)")
	}
	descriptor, found := core.DescriptorOf(core.RoofedAutoCorrelationHeatmap)
	if !found || len(descriptor.Outputs) != 1 || descriptor.Outputs[0].Role != core.Correlation || descriptor.Outputs[0].Pane != core.Own {
		t.Fatalf("unexpected descriptor = %+v (found=%v)", descriptor, found)
	}
}

func TestRoofedAutoCorrelationHeatmapValidationAndEdgeBehavior(t *testing.T) {
	t.Parallel()

	t.Run("parameter validation", func(t *testing.T) {
		t.Parallel()
		invalidParams := []Params{
			{MinLag: -1, MaxLag: 48, AveragingLength: 0, HighpassPeriod: 48.0, SmoothingPeriod: 10.0},
			{MinLag: 10, MaxLag: 9, AveragingLength: 0, HighpassPeriod: 48.0, SmoothingPeriod: 10.0},
			{MinLag: 3, MaxLag: 48, AveragingLength: -1, HighpassPeriod: 48.0, SmoothingPeriod: 10.0},
			{MinLag: 3, MaxLag: 48, AveragingLength: 0, HighpassPeriod: 4.9, SmoothingPeriod: 10.0},
			{MinLag: 3, MaxLag: 48, AveragingLength: 0, HighpassPeriod: math.NaN(), SmoothingPeriod: 10.0},
			{MinLag: 3, MaxLag: 48, AveragingLength: 0, HighpassPeriod: math.Inf(1), SmoothingPeriod: 10.0},
			{MinLag: 3, MaxLag: 48, AveragingLength: 0, HighpassPeriod: 48.0, SmoothingPeriod: 1.9},
			{MinLag: 3, MaxLag: 48, AveragingLength: 0, HighpassPeriod: 48.0, SmoothingPeriod: math.NaN()},
			{MinLag: 3, MaxLag: 48, AveragingLength: 0, HighpassPeriod: 48.0, SmoothingPeriod: math.Inf(1)},
			{MinLag: 3, MaxLag: 48, AveragingLength: 0, HighpassPeriod: 48.0, SmoothingPeriod: 10.0, BarComponent: entities.BarComponent(99)},
		}
		for index, bad := range invalidParams {
			badCopy := bad
			if _, err := NewRoofedAutoCorrelationHeatmap(&badCopy); err == nil {
				t.Fatalf("case %d: expected error for invalid params %+v", index, bad)
			}
		}

		zeroZero, err := NewRoofedAutoCorrelationHeatmap(&Params{
			MinLag: 0, MaxLag: 0, AveragingLength: 0, HighpassPeriod: 48.0, SmoothingPeriod: 10.0,
		})
		if err != nil {
			t.Fatalf("zero-lag zero-averaging should be valid: %v", err)
		}
		if zeroZero.WarmupBars() != 0 {
			t.Fatalf("WarmupBars() = %d, want 0", zeroZero.WarmupBars())
		}
		col := zeroZero.Update(100.0, time.Time{})
		if !zeroZero.IsPrimed() || !slices.Equal(col.Values, []float64{0.5}) {
			t.Fatalf("expected single-sample mapped seed [0.5], got %v", col.Values)
		}
	})

	t.Run("missing samples, rejected overflow, and reset", func(t *testing.T) {
		t.Parallel()
		subject, _ := NewRoofedAutoCorrelationHeatmap(&Params{
			MinLag: 3, MaxLag: 12, AveragingLength: 5, HighpassPeriod: 24.0, SmoothingPeriod: 10.0,
		})
		clean, _ := NewRoofedAutoCorrelationHeatmap(&Params{
			MinLag: 3, MaxLag: 12, AveragingLength: 5, HighpassPeriod: 24.0, SmoothingPeriod: 10.0,
		})
		stamp := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
		for barIndex, price := range testInput[:50] {
			before := subject.State()
			beforeSeen := subject.SamplesSeen()
			skipped, err := subject.TryUpdate(math.NaN(), stamp)
			if err != nil || !skipped.IsEmpty() {
				t.Fatalf("expected empty heatmap on NaN, got err=%v", err)
			}
			for _, badSample := range []float64{math.Inf(1), math.Inf(-1), 1e308} {
				if _, err := subject.TryUpdate(badSample, stamp); err == nil {
					t.Fatalf("expected error for %v", badSample)
				}
				after := subject.State()
				if subject.SamplesSeen() != beforeSeen ||
					after.HighPass != before.HighPass ||
					after.Filtered != before.Filtered ||
					!slices.Equal(after.Correlations, before.Correlations) {
					t.Fatalf("state mutated after rejected sample %v", badSample)
				}
			}
			actual := subject.Update(price, stamp)
			expected := clean.Update(price, stamp)
			if subject.IsPrimed() != (barIndex+1 >= 17) {
				t.Fatalf("IsPrimed mismatch at bar %d", barIndex)
			}
			if !slices.Equal(actual.Values, expected.Values) {
				t.Fatalf("values diverged at bar %d", barIndex)
			}
		}

		subject.Reset()
		clean.Reset()
		if subject.SamplesSeen() != 0 || subject.IsPrimed() {
			t.Fatalf("Reset did not clear state")
		}
		for _, price := range testInput[:50] {
			actual := subject.Update(price, stamp)
			expected := clean.Update(price, stamp)
			if !slices.Equal(actual.Values, expected.Values) {
				t.Fatalf("post-reset values diverged")
			}
		}
	})

	t.Run("zero-variance Listing 8-2 rescaling quirk and sine periodicity", func(t *testing.T) {
		t.Parallel()
		flatInd, _ := NewRoofedAutoCorrelationHeatmap(&Params{
			MinLag: 0, MaxLag: 2, AveragingLength: 3, HighpassPeriod: 48.0, SmoothingPeriod: 10.0,
		})
		for step := 1; step <= 6; step++ {
			col := flatInd.Update(0.0, time.Time{})
			expected := 1.0 - math.Pow(0.5, float64(step))
			state := flatInd.State()
			for lag := range 3 {
				if !almostEqual(state.Correlations[lag], expected, 1e-15) {
					t.Fatalf("step %d lag %d state = %.16g, want %.16g", step, lag, state.Correlations[lag], expected)
				}
			}
			if step >= 5 {
				for lag := range 3 {
					if !almostEqual(col.Values[lag], expected, 1e-15) {
						t.Fatalf("step %d lag %d output = %.16g, want %.16g", step, lag, col.Values[lag], expected)
					}
				}
			}
		}

		sineInd, _ := NewRoofedAutoCorrelationHeatmap(&Params{
			MinLag: 3, MaxLag: 40, AveragingLength: 20, HighpassPeriod: 48.0, SmoothingPeriod: 10.0,
		})
		var column *outputs.Heatmap
		for step := range 800 {
			price := 100.0 + 5.0*math.Sin(2.0*math.Pi*float64(step)/20.0)
			column = sineInd.Update(price, time.Time{})
		}
		if !almostEqual(column.Values[20-3], 1.0, 1e-10) {
			t.Fatalf("lag 20 mapped correlation = %.15g, want 1.0", column.Values[20-3])
		}
		if !almostEqual(column.Values[10-3], 0.0, 1e-10) {
			t.Fatalf("lag 10 mapped correlation = %.15g, want 0.0", column.Values[10-3])
		}
	})

	t.Run("entity updates and snapshot ownership", func(t *testing.T) {
		t.Parallel()
		scalarInd, _ := NewRoofedAutoCorrelationHeatmapDefault()
		barInd, _ := NewRoofedAutoCorrelationHeatmapDefault()
		quoteInd, _ := NewRoofedAutoCorrelationHeatmapDefault()
		tradeInd, _ := NewRoofedAutoCorrelationHeatmapDefault()
		openInd, _ := NewRoofedAutoCorrelationHeatmap(&Params{
			MinLag:          3,
			MaxLag:          48,
			AveragingLength: 0,
			HighpassPeriod:  48.0,
			SmoothingPeriod: 10.0,
			BarComponent:    entities.BarOpenPrice,
		})
		openRef, _ := NewRoofedAutoCorrelationHeatmapDefault()

		baseTime := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
		for barIndex := range testInput {
			stamp := baseTime.AddDate(0, 0, barIndex)
			closePrice := testInput[barIndex]
			openPrice := testOpen[barIndex]
			highPrice := testHigh[barIndex]
			lowPrice := testLow[barIndex]

			outScalar := scalarInd.UpdateScalar(&entities.Scalar{Time: stamp, Value: closePrice})[0].(*outputs.Heatmap)
			outBar := barInd.UpdateBar(&entities.Bar{
				Time: stamp, Open: openPrice, High: highPrice, Low: lowPrice, Close: closePrice, Volume: 1000,
			})[0].(*outputs.Heatmap)
			outQuote := quoteInd.UpdateQuote(&entities.Quote{
				Time: stamp, Bid: closePrice, Ask: closePrice, BidSize: 10, AskSize: 10,
			})[0].(*outputs.Heatmap)
			outTrade := tradeInd.UpdateTrade(&entities.Trade{
				Time: stamp, Price: closePrice, Volume: 10,
			})[0].(*outputs.Heatmap)

			if !slices.Equal(outScalar.Values, outBar.Values) ||
				!slices.Equal(outScalar.Values, outQuote.Values) ||
				!slices.Equal(outScalar.Values, outTrade.Values) {
				t.Fatalf("entity outputs mismatch at bar %d", barIndex)
			}

			outOpen := openInd.UpdateBar(&entities.Bar{
				Time: stamp, Open: openPrice, High: highPrice, Low: lowPrice, Close: closePrice, Volume: 1000,
			})[0].(*outputs.Heatmap)
			outOpenScalar := openRef.UpdateScalar(&entities.Scalar{Time: stamp, Value: openPrice})[0].(*outputs.Heatmap)
			if !slices.Equal(outOpen.Values, outOpenScalar.Values) {
				t.Fatalf("open bar component mismatch at bar %d", barIndex)
			}

			if !outScalar.IsEmpty() {
				stateBefore := scalarInd.State()
				outputBefore := scalarInd.Output()
				outScalar.Values[0] = 9999.0
				if !slices.Equal(scalarInd.State().Correlations, stateBefore.Correlations) {
					t.Fatalf("mutating returned Heatmap.Values mutated internal state")
				}
				if !slices.Equal(scalarInd.Output().Values, outputBefore.Values) {
					t.Fatalf("mutating returned Heatmap.Values mutated Output() snapshot")
				}
			}
		}
	})
}
