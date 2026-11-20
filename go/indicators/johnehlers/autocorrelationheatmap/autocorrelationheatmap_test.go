//nolint:testpackage
package autocorrelationheatmap

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
	fixtureTolerance = 1e-13
	oracleTolerance  = 1e-13
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
			name:            "L20_P20_H99",
			params:          Params{Length: 20, SmoothingPeriod: 20.0, MaxLag: 99},
			expectedHeatmap: expectedL20_P20_H99_HEATMAP,
			expectedMin:     expectedL20_P20_H99_MIN,
			expectedMax:     expectedL20_P20_H99_MAX,
		},
		{
			name:            "L2_P20_H99",
			params:          Params{Length: 2, SmoothingPeriod: 20.0, MaxLag: 99},
			expectedHeatmap: expectedL2_P20_H99_HEATMAP,
			expectedMin:     expectedL2_P20_H99_MIN,
			expectedMax:     expectedL2_P20_H99_MAX,
		},
		{
			name:            "L3_P20_H99",
			params:          Params{Length: 3, SmoothingPeriod: 20.0, MaxLag: 99},
			expectedHeatmap: expectedL3_P20_H99_HEATMAP,
			expectedMin:     expectedL3_P20_H99_MIN,
			expectedMax:     expectedL3_P20_H99_MAX,
		},
		{
			name:            "L5_P20_H99",
			params:          Params{Length: 5, SmoothingPeriod: 20.0, MaxLag: 99},
			expectedHeatmap: expectedL5_P20_H99_HEATMAP,
			expectedMin:     expectedL5_P20_H99_MIN,
			expectedMax:     expectedL5_P20_H99_MAX,
		},
		{
			name:            "L40_P20_H99",
			params:          Params{Length: 40, SmoothingPeriod: 20.0, MaxLag: 99},
			expectedHeatmap: expectedL40_P20_H99_HEATMAP,
			expectedMin:     expectedL40_P20_H99_MIN,
			expectedMax:     expectedL40_P20_H99_MAX,
		},
		{
			name:            "L100_P20_H99",
			params:          Params{Length: 100, SmoothingPeriod: 20.0, MaxLag: 99},
			expectedHeatmap: expectedL100_P20_H99_HEATMAP,
			expectedMin:     expectedL100_P20_H99_MIN,
			expectedMax:     expectedL100_P20_H99_MAX,
		},
		{
			name:            "L20_P10_H99",
			params:          Params{Length: 20, SmoothingPeriod: 10.0, MaxLag: 99},
			expectedHeatmap: expectedL20_P10_H99_HEATMAP,
			expectedMin:     expectedL20_P10_H99_MIN,
			expectedMax:     expectedL20_P10_H99_MAX,
		},
		{
			name:            "L20_P10_5_H99",
			params:          Params{Length: 20, SmoothingPeriod: 10.5, MaxLag: 99},
			expectedHeatmap: expectedL20_P10_5_H99_HEATMAP,
			expectedMin:     expectedL20_P10_5_H99_MIN,
			expectedMax:     expectedL20_P10_5_H99_MAX,
		},
		{
			name:            "L20_P40_H99",
			params:          Params{Length: 20, SmoothingPeriod: 40.0, MaxLag: 99},
			expectedHeatmap: expectedL20_P40_H99_HEATMAP,
			expectedMin:     expectedL20_P40_H99_MIN,
			expectedMax:     expectedL20_P40_H99_MAX,
		},
		{
			name:            "L20_P20_H32",
			params:          Params{Length: 20, SmoothingPeriod: 20.0, MaxLag: 32},
			expectedHeatmap: expectedL20_P20_H32_HEATMAP,
			expectedMin:     expectedL20_P20_H32_MIN,
			expectedMax:     expectedL20_P20_H32_MAX,
		},
		{
			name:            "L2_P20_H32",
			params:          Params{Length: 2, SmoothingPeriod: 20.0, MaxLag: 32},
			expectedHeatmap: expectedL2_P20_H32_HEATMAP,
			expectedMin:     expectedL2_P20_H32_MIN,
			expectedMax:     expectedL2_P20_H32_MAX,
		},
		{
			name:            "L20_P20_H0",
			params:          Params{Length: 20, SmoothingPeriod: 20.0, MaxLag: 0},
			expectedHeatmap: expectedL20_P20_H0_HEATMAP,
			expectedMin:     expectedL20_P20_H0_MIN,
			expectedMax:     expectedL20_P20_H0_MAX,
		},
	}
}

type oracleRow struct {
	filtered float64
	bins     []float64
}

func batchOracle(closes []float64, length int, smoothingPeriod float64, maxLag int) []oracleRow {
	filteredSeries := make([]float64, len(closes))
	var correlations [101]float64
	angle := math.Sqrt2 * math.Pi / smoothingPeriod
	attenuation := math.Exp(-angle)
	coeffC2 := 2.0 * attenuation * math.Cos(angle)
	coeffC3 := -attenuation * attenuation
	coeffC1 := (1.0 + coeffC2 - coeffC3) / 4.0
	primeCount := 4
	if length+maxLag > primeCount {
		primeCount = length + maxLag
	}

	rows := make([]oracleRow, len(closes))
	lengthFloat := float64(length)
	for barIndex, price := range closes {
		if barIndex < 3 {
			filteredSeries[barIndex] = price
		} else {
			filteredSeries[barIndex] = (1.0-coeffC1)*price +
				(2.0*coeffC1-coeffC2)*closes[barIndex-1] -
				(coeffC1+coeffC3)*closes[barIndex-2] +
				coeffC2*filteredSeries[barIndex-1] +
				coeffC3*filteredSeries[barIndex-2]
		}
		for lag := range maxLag + 1 {
			var sumX, sumY, sumXX, sumXY, sumYY float64
			for offset := range length {
				var sampleX, sampleY float64
				if barIndex-offset >= 0 {
					sampleX = filteredSeries[barIndex-offset]
				}
				if barIndex-lag-offset >= 0 {
					sampleY = filteredSeries[barIndex-lag-offset]
				}
				sumX += sampleX
				sumY += sampleY
				sumXX += sampleX * sampleX
				sumXY += sampleX * sampleY
				sumYY += sampleY * sampleY
			}
			varX := lengthFloat*sumXX - sumX*sumX
			varY := lengthFloat*sumYY - sumY*sumY
			if varX > 0.0 && varY > 0.0 {
				correlations[lag+1] = (lengthFloat*sumXY - sumX*sumY) / math.Sqrt(varX*varY)
			}
		}
		if barIndex+1 < primeCount {
			rows[barIndex] = oracleRow{filtered: filteredSeries[barIndex]}
		} else {
			rows[barIndex] = oracleRow{
				filtered: filteredSeries[barIndex],
				bins:     slices.Clone(correlations[1 : maxLag+2]),
			}
		}
	}
	return rows
}

func almostEqual(first, second, tolerance float64) bool {
	if math.IsNaN(first) && math.IsNaN(second) {
		return true
	}
	if math.IsNaN(first) || math.IsNaN(second) {
		return false
	}
	return math.Abs(first-second) <= tolerance
}

func TestAutoCorrelationHeatmapReferenceCases(t *testing.T) {
	t.Parallel()

	if len(testInput) != 252 {
		t.Fatalf("len(testInput) = %d, want 252", len(testInput))
	}

	for _, tc := range referenceCases() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			indicator, err := NewAutoCorrelationHeatmap(&tc.params)
			if err != nil {
				t.Fatalf("NewAutoCorrelationHeatmap failed: %v", err)
			}

			warmup := tc.params.Length + tc.params.MaxLag - 1
			if warmup < 3 {
				warmup = 3
			}
			if indicator.WarmupBars() != warmup {
				t.Fatalf("WarmupBars = %d, want %d", indicator.WarmupBars(), warmup)
			}

			oracle := batchOracle(testInput, tc.params.Length, tc.params.SmoothingPeriod, tc.params.MaxLag)
			binCount := tc.params.MaxLag + 1
			baseTime := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

			for barIndex, price := range testInput {
				timestamp := baseTime.AddDate(0, 0, barIndex)
				result := indicator.Update(price, timestamp)

				if result.ParameterFirst != 0.0 || result.ParameterLast != float64(tc.params.MaxLag) || result.ParameterResolution != 1.0 {
					t.Fatalf("bar %d axis = (%v, %v, %v), want (0, %v, 1)",
						barIndex, result.ParameterFirst, result.ParameterLast, result.ParameterResolution, float64(tc.params.MaxLag))
				}
				if indicator.IsPrimed() != (barIndex >= warmup) {
					t.Fatalf("bar %d IsPrimed = %v, want %v", barIndex, indicator.IsPrimed(), barIndex >= warmup)
				}
				if !almostEqual(indicator.State().Filtered, oracle[barIndex].filtered, oracleTolerance) {
					t.Fatalf("bar %d filtered = %.17g, oracle = %.17g", barIndex, indicator.State().Filtered, oracle[barIndex].filtered)
				}

				expectedRow := tc.expectedHeatmap[barIndex]
				if barIndex < warmup {
					if !result.IsEmpty() || !math.IsNaN(result.ValueMin) || !math.IsNaN(result.ValueMax) {
						t.Fatalf("bar %d expected empty heatmap before warmup", barIndex)
					}
					if !math.IsNaN(tc.expectedMin[barIndex]) || !math.IsNaN(tc.expectedMax[barIndex]) {
						t.Fatalf("bar %d expected NaN min/max in fixture", barIndex)
					}
					continue
				}

				if len(result.Values) != binCount {
					t.Fatalf("bar %d len(Values) = %d, want %d", barIndex, len(result.Values), binCount)
				}
				if !almostEqual(result.ValueMin, tc.expectedMin[barIndex], fixtureTolerance) {
					t.Fatalf("bar %d ValueMin = %.17g, want %.17g", barIndex, result.ValueMin, tc.expectedMin[barIndex])
				}
				if !almostEqual(result.ValueMax, tc.expectedMax[barIndex], fixtureTolerance) {
					t.Fatalf("bar %d ValueMax = %.17g, want %.17g", barIndex, result.ValueMax, tc.expectedMax[barIndex])
				}
				for lag := range binCount {
					if !almostEqual(result.Values[lag], expectedRow[lag], fixtureTolerance) {
						t.Fatalf("bar %d lag %d = %.17g, fixture = %.17g", barIndex, lag, result.Values[lag], expectedRow[lag])
					}
					if !almostEqual(result.Values[lag], oracle[barIndex].bins[lag], oracleTolerance) {
						t.Fatalf("bar %d lag %d = %.17g, oracle = %.17g", barIndex, lag, result.Values[lag], oracle[barIndex].bins[lag])
					}
				}
			}
		})
	}
}

func TestAutoCorrelationHeatmapMnemonicsAndMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		params       Params
		wantMnemonic string
	}{
		{
			name:         "all components zero",
			params:       *DefaultParams(),
			wantMnemonic: "ach(20, 20, 99)",
		},
		{
			name:         "fractional period and zero lag",
			params:       Params{Length: 20, SmoothingPeriod: 10.5, MaxLag: 0},
			wantMnemonic: "ach(20, 10.5, 0)",
		},
		{
			name:         "only bar component set",
			params:       Params{Length: 20, SmoothingPeriod: 20, MaxLag: 99, BarComponent: entities.BarMedianPrice},
			wantMnemonic: "ach(20, 20, 99, hl/2)",
		},
		{
			name:         "only quote component set",
			params:       Params{Length: 20, SmoothingPeriod: 20, MaxLag: 99, QuoteComponent: entities.QuoteBidPrice},
			wantMnemonic: "ach(20, 20, 99, b)",
		},
		{
			name:         "only trade component set",
			params:       Params{Length: 20, SmoothingPeriod: 20, MaxLag: 99, TradeComponent: entities.TradeVolume},
			wantMnemonic: "ach(20, 20, 99, v)",
		},
		{
			name: "bar and quote components set",
			params: Params{
				Length:          20,
				SmoothingPeriod: 20,
				MaxLag:          99,
				BarComponent:    entities.BarTypicalPrice,
				QuoteComponent:  entities.QuoteAskPrice,
			},
			wantMnemonic: "ach(20, 20, 99, hlc/3, a)",
		},
		{
			name: "all three components set",
			params: Params{
				Length:          20,
				SmoothingPeriod: 20,
				MaxLag:          99,
				BarComponent:    entities.BarOpenPrice,
				QuoteComponent:  entities.QuoteWeightedPrice,
				TradeComponent:  entities.TradeVolume,
			},
			wantMnemonic: "ach(20, 20, 99, o, (bbs+aas)/(bs+as), v)",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			indicator, err := NewAutoCorrelationHeatmap(&tc.params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			meta := indicator.Metadata()
			if meta.Identifier != core.AutoCorrelationHeatmap {
				t.Fatalf("Identifier = %v, want %v", meta.Identifier, core.AutoCorrelationHeatmap)
			}
			if meta.Mnemonic != tc.wantMnemonic {
				t.Fatalf("Mnemonic = %q, want %q", meta.Mnemonic, tc.wantMnemonic)
			}
			wantDescription := "Autocorrelation heatmap " + tc.wantMnemonic
			if meta.Description != wantDescription {
				t.Fatalf("Description = %q, want %q", meta.Description, wantDescription)
			}
			if len(meta.Outputs) != 1 {
				t.Fatalf("len(Outputs) = %d, want 1", len(meta.Outputs))
			}
			if meta.Outputs[0].Shape != shape.Heatmap {
				t.Fatalf("Outputs[0].Shape = %v, want %v", meta.Outputs[0].Shape, shape.Heatmap)
			}
			if meta.Outputs[0].Mnemonic != tc.wantMnemonic {
				t.Fatalf("Outputs[0].Mnemonic = %q, want %q", meta.Outputs[0].Mnemonic, tc.wantMnemonic)
			}
		})
	}
}

func TestAutoCorrelationHeatmapBehaviorAndEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("validation errors", func(t *testing.T) {
		t.Parallel()
		invalid := []Params{
			{Length: 1, SmoothingPeriod: 20, MaxLag: 99},
			{Length: 20, SmoothingPeriod: 0.5, MaxLag: 99},
			{Length: 20, SmoothingPeriod: math.NaN(), MaxLag: 99},
			{Length: 20, SmoothingPeriod: math.Inf(1), MaxLag: 99},
			{Length: 20, SmoothingPeriod: 20, MaxLag: -1},
			{Length: 20, SmoothingPeriod: 20, MaxLag: 100},
			{Length: 20, SmoothingPeriod: 20, MaxLag: 99, BarComponent: entities.BarComponent(9999)},
		}
		for _, bad := range invalid {
			if _, err := NewAutoCorrelationHeatmap(&bad); err == nil {
				t.Fatalf("expected error for %+v", bad)
			}
		}
	})

	t.Run("nan skip overflow rollback and reset", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: 5, SmoothingPeriod: 10.0, MaxLag: 12}
		subject, err := NewAutoCorrelationHeatmap(&params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		clean, err := NewAutoCorrelationHeatmap(&params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stamp := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
		for barIndex, price := range testInput[:50] {
			before := subject.State()
			skipped, err := subject.TryUpdate(math.NaN(), stamp)
			if err != nil || !skipped.IsEmpty() {
				t.Fatalf("expected empty heatmap on NaN, got err=%v", err)
			}
			for _, badSample := range []float64{math.Inf(1), math.Inf(-1), 1e308} {
				if _, err := subject.TryUpdate(badSample, stamp); err == nil {
					t.Fatalf("expected error for %v", badSample)
				}
				after := subject.State()
				if after.AcceptedSamples != before.AcceptedSamples || after.Filtered != before.Filtered || !slices.Equal(after.Correlations, before.Correlations) {
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

	t.Run("analytical sine and degenerate retention", func(t *testing.T) {
		t.Parallel()
		sineInd, err := NewAutoCorrelationHeatmap(&Params{Length: 20, SmoothingPeriod: 20.0, MaxLag: 40})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var column *outputs.Heatmap
		for step := range 1000 {
			price := math.Sin(2.0 * math.Pi * float64(step) / 20.0)
			column = sineInd.Update(price, time.Time{})
		}
		for lag := range 41 {
			ideal := math.Cos(2.0 * math.Pi * float64(lag) / 20.0)
			if !almostEqual(column.Values[lag], ideal, 1e-11) {
				t.Fatalf("lag %d = %.15g, want %.15g", lag, column.Values[lag], ideal)
			}
		}

		flatInd, err := NewAutoCorrelationHeatmap(&Params{Length: 3, SmoothingPeriod: 20.0, MaxLag: 2})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var flatOut *outputs.Heatmap
		for range 30 {
			flatOut = flatInd.Update(0.0, time.Time{})
		}
		if !slices.Equal(flatOut.Values, []float64{0.0, 0.0, 0.0}) {
			t.Fatalf("flat zero stream values = %v, want [0, 0, 0]", flatOut.Values)
		}
		flatInd.correlations = []float64{0.25, -0.5, 0.75}
		retained := flatInd.Update(0.0, time.Time{})
		if !slices.Equal(retained.Values, []float64{0.25, -0.5, 0.75}) || retained.ValueMin != -0.5 || retained.ValueMax != 0.75 {
			t.Fatalf("retained = %+v", retained)
		}
	})

	t.Run("entity updates and snapshot ownership", func(t *testing.T) {
		t.Parallel()
		scalarInd, _ := NewAutoCorrelationHeatmapDefault()
		barInd, _ := NewAutoCorrelationHeatmapDefault()
		quoteInd, _ := NewAutoCorrelationHeatmapDefault()
		tradeInd, _ := NewAutoCorrelationHeatmapDefault()
		openInd, _ := NewAutoCorrelationHeatmap(&Params{
			Length:          20,
			SmoothingPeriod: 20,
			MaxLag:          99,
			BarComponent:    entities.BarOpenPrice,
		})
		openRef, _ := NewAutoCorrelationHeatmapDefault()

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
				outScalar.Values[0] = 9999.0
				if !slices.Equal(scalarInd.State().Correlations, stateBefore.Correlations) {
					t.Fatalf("mutating returned Heatmap.Values mutated internal state")
				}
			}
		}
	})
}
