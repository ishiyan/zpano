//nolint:testpackage
package transformeddiscretefouriertransformspectrum

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

const fixtureTolerance = 1e-12

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
			name:            "P8_Q50_W50_HP40",
			params:          Params{MinPeriod: 8, MaxPeriod: 50, WindowLength: 50, HighpassPeriod: 40},
			expectedHeatmap: expectedP8_Q50_W50_HP40_HEATMAP,
			expectedMin:     expectedP8_Q50_W50_HP40_MIN,
			expectedMax:     expectedP8_Q50_W50_HP40_MAX,
			expectedDC:      expectedP8_Q50_W50_HP40_DOMINANT_CYCLE,
		},
		{
			name:            "P8_Q50_W1_HP40",
			params:          Params{MinPeriod: 8, MaxPeriod: 50, WindowLength: 1, HighpassPeriod: 40},
			expectedHeatmap: expectedP8_Q50_W1_HP40_HEATMAP,
			expectedMin:     expectedP8_Q50_W1_HP40_MIN,
			expectedMax:     expectedP8_Q50_W1_HP40_MAX,
			expectedDC:      expectedP8_Q50_W1_HP40_DOMINANT_CYCLE,
		},
		{
			name:            "P8_Q50_W20_HP40",
			params:          Params{MinPeriod: 8, MaxPeriod: 50, WindowLength: 20, HighpassPeriod: 40},
			expectedHeatmap: expectedP8_Q50_W20_HP40_HEATMAP,
			expectedMin:     expectedP8_Q50_W20_HP40_MIN,
			expectedMax:     expectedP8_Q50_W20_HP40_MAX,
			expectedDC:      expectedP8_Q50_W20_HP40_DOMINANT_CYCLE,
		},
		{
			name:            "P8_Q50_W48_HP40",
			params:          Params{MinPeriod: 8, MaxPeriod: 50, WindowLength: 48, HighpassPeriod: 40},
			expectedHeatmap: expectedP8_Q50_W48_HP40_HEATMAP,
			expectedMin:     expectedP8_Q50_W48_HP40_MIN,
			expectedMax:     expectedP8_Q50_W48_HP40_MAX,
			expectedDC:      expectedP8_Q50_W48_HP40_DOMINANT_CYCLE,
		},
		{
			name:            "P8_Q50_W64_HP40",
			params:          Params{MinPeriod: 8, MaxPeriod: 50, WindowLength: 64, HighpassPeriod: 40},
			expectedHeatmap: expectedP8_Q50_W64_HP40_HEATMAP,
			expectedMin:     expectedP8_Q50_W64_HP40_MIN,
			expectedMax:     expectedP8_Q50_W64_HP40_MAX,
			expectedDC:      expectedP8_Q50_W64_HP40_DOMINANT_CYCLE,
		},
		{
			name:            "P8_Q50_W252_HP40",
			params:          Params{MinPeriod: 8, MaxPeriod: 50, WindowLength: 252, HighpassPeriod: 40},
			expectedHeatmap: expectedP8_Q50_W252_HP40_HEATMAP,
			expectedMin:     expectedP8_Q50_W252_HP40_MIN,
			expectedMax:     expectedP8_Q50_W252_HP40_MAX,
			expectedDC:      expectedP8_Q50_W252_HP40_DOMINANT_CYCLE,
		},
		{
			name:            "P8_Q50_W50_HP20",
			params:          Params{MinPeriod: 8, MaxPeriod: 50, WindowLength: 50, HighpassPeriod: 20},
			expectedHeatmap: expectedP8_Q50_W50_HP20_HEATMAP,
			expectedMin:     expectedP8_Q50_W50_HP20_MIN,
			expectedMax:     expectedP8_Q50_W50_HP20_MAX,
			expectedDC:      expectedP8_Q50_W50_HP20_DOMINANT_CYCLE,
		},
		{
			name:            "P8_Q50_W50_HP80",
			params:          Params{MinPeriod: 8, MaxPeriod: 50, WindowLength: 50, HighpassPeriod: 80},
			expectedHeatmap: expectedP8_Q50_W50_HP80_HEATMAP,
			expectedMin:     expectedP8_Q50_W50_HP80_MIN,
			expectedMax:     expectedP8_Q50_W50_HP80_MAX,
			expectedDC:      expectedP8_Q50_W50_HP80_DOMINANT_CYCLE,
		},
		{
			name:            "P8_Q50_W50_HP40_5",
			params:          Params{MinPeriod: 8, MaxPeriod: 50, WindowLength: 50, HighpassPeriod: 40.5},
			expectedHeatmap: expectedP8_Q50_W50_HP40_5_HEATMAP,
			expectedMin:     expectedP8_Q50_W50_HP40_5_MIN,
			expectedMax:     expectedP8_Q50_W50_HP40_5_MAX,
			expectedDC:      expectedP8_Q50_W50_HP40_5_DOMINANT_CYCLE,
		},
		{
			name:            "P10_Q30_W50_HP40",
			params:          Params{MinPeriod: 10, MaxPeriod: 30, WindowLength: 50, HighpassPeriod: 40},
			expectedHeatmap: expectedP10_Q30_W50_HP40_HEATMAP,
			expectedMin:     expectedP10_Q30_W50_HP40_MIN,
			expectedMax:     expectedP10_Q30_W50_HP40_MAX,
			expectedDC:      expectedP10_Q30_W50_HP40_DOMINANT_CYCLE,
		},
		{
			name:            "P8_Q80_W50_HP40",
			params:          Params{MinPeriod: 8, MaxPeriod: 80, WindowLength: 50, HighpassPeriod: 40},
			expectedHeatmap: expectedP8_Q80_W50_HP40_HEATMAP,
			expectedMin:     expectedP8_Q80_W50_HP40_MIN,
			expectedMax:     expectedP8_Q80_W50_HP40_MAX,
			expectedDC:      expectedP8_Q80_W50_HP40_DOMINANT_CYCLE,
		},
		{
			name:            "P20_Q20_W50_HP40",
			params:          Params{MinPeriod: 20, MaxPeriod: 20, WindowLength: 50, HighpassPeriod: 40},
			expectedHeatmap: expectedP20_Q20_W50_HP40_HEATMAP,
			expectedMin:     expectedP20_Q20_W50_HP40_MIN,
			expectedMax:     expectedP20_Q20_W50_HP40_MAX,
			expectedDC:      expectedP20_Q20_W50_HP40_DOMINANT_CYCLE,
		},
	}
}

func assertClose(t *testing.T, label string, actual, expected, tolerance float64) {
	t.Helper()
	if math.IsNaN(expected) {
		if !math.IsNaN(actual) {
			t.Fatalf("%s: expected NaN, got %0.17g", label, actual)
		}
		return
	}
	if math.IsNaN(actual) {
		t.Fatalf("%s: expected %0.17g, got NaN", label, expected)
	}
	bound := math.Max(tolerance, math.Abs(expected)*tolerance)
	if math.Abs(actual-expected) > bound {
		t.Fatalf("%s: actual=%0.17g expected=%0.17g diff=%0.17g", label, actual, expected, math.Abs(actual-expected))
	}
}

func assertSameSpectrumOutput(t *testing.T, actual, expected SpectrumOutput) {
	t.Helper()
	if !actual.Heatmap.Time.Equal(expected.Heatmap.Time) {
		t.Fatalf("heatmap time mismatch: %v vs %v", actual.Heatmap.Time, expected.Heatmap.Time)
	}
	if actual.Heatmap.ParameterFirst != expected.Heatmap.ParameterFirst ||
		actual.Heatmap.ParameterLast != expected.Heatmap.ParameterLast ||
		actual.Heatmap.ParameterResolution != expected.Heatmap.ParameterResolution ||
		actual.Heatmap.IsEmpty() != expected.Heatmap.IsEmpty() {
		t.Fatalf("heatmap axis/empty mismatch")
	}
	assertClose(t, "valueMin", actual.Heatmap.ValueMin, expected.Heatmap.ValueMin, 0)
	assertClose(t, "valueMax", actual.Heatmap.ValueMax, expected.Heatmap.ValueMax, 0)
	if !slices.Equal(actual.Heatmap.Values, expected.Heatmap.Values) {
		t.Fatalf("heatmap values mismatch")
	}
	assertClose(t, "dominantCycle", actual.DominantCycle, expected.DominantCycle, 0)
}

func TestAllReferenceCases(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, tc := range allReferenceCases() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			indicator, err := NewTransformedDiscreteFourierTransformSpectrumParams(&tc.params)
			if err != nil {
				t.Fatalf("construct failed: %v", err)
			}
			channelCount := tc.params.MaxPeriod - tc.params.MinPeriod + 1
			for barIndex, closePrice := range testInput {
				timestamp := baseTime.AddDate(0, 0, barIndex)
				result := indicator.Update(closePrice, timestamp)
				if !result.Heatmap.Time.Equal(timestamp) {
					t.Fatalf("bar %d: unexpected heatmap time", barIndex)
				}
				if result.Heatmap.ParameterFirst != float64(tc.params.MinPeriod) ||
					result.Heatmap.ParameterLast != float64(tc.params.MaxPeriod) ||
					result.Heatmap.ParameterResolution != 1.0 {
					t.Fatalf("bar %d: unexpected heatmap grid metadata", barIndex)
				}

				row := tc.expectedHeatmap[barIndex]
				if math.IsNaN(tc.expectedDC[barIndex]) {
					if indicator.IsPrimed() {
						t.Fatalf("bar %d: expected unprimed", barIndex)
					}
					if !result.Heatmap.IsEmpty() {
						t.Fatalf("bar %d: expected empty heatmap", barIndex)
					}
					if !math.IsNaN(result.Heatmap.ValueMin) || !math.IsNaN(result.Heatmap.ValueMax) {
						t.Fatalf("bar %d: expected NaN extrema", barIndex)
					}
					if !math.IsNaN(result.DominantCycle) {
						t.Fatalf("bar %d: expected NaN dominant cycle", barIndex)
					}
					for _, cell := range row {
						if !math.IsNaN(cell) {
							t.Fatalf("bar %d: expected all NaN fixture row", barIndex)
						}
					}
				} else {
					if !indicator.IsPrimed() {
						t.Fatalf("bar %d: expected primed", barIndex)
					}
					if result.Heatmap.IsEmpty() {
						t.Fatalf("bar %d: expected non-empty heatmap", barIndex)
					}
					values := result.Heatmap.Values
					if len(values) != channelCount {
						t.Fatalf("bar %d: len(values)=%d want %d", barIndex, len(values), channelCount)
					}
					assertClose(t, "valueMin", result.Heatmap.ValueMin, tc.expectedMin[barIndex], fixtureTolerance)
					assertClose(t, "valueMax", result.Heatmap.ValueMax, tc.expectedMax[barIndex], fixtureTolerance)
					assertClose(t, "dominantCycle", result.DominantCycle, tc.expectedDC[barIndex], fixtureTolerance)
					for channelIndex, expectedCell := range row {
						assertClose(t, "cell", values[channelIndex], expectedCell, fixtureTolerance)
					}
				}
			}
		})
	}
}

func TestMissingInvalidOverflowAndReset(t *testing.T) {
	t.Parallel()

	stream, err := NewTransformedDiscreteFourierTransformSpectrumDefault()
	if err != nil {
		t.Fatalf("construct failed: %v", err)
	}
	control, err := NewTransformedDiscreteFourierTransformSpectrumDefault()
	if err != nil {
		t.Fatalf("construct failed: %v", err)
	}

	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for barIndex := range 80 {
		timestamp := baseTime.AddDate(0, 0, barIndex)
		stream.Update(testInput[barIndex], timestamp)
		control.Update(testInput[barIndex], timestamp)
	}

	beforeState := stream.State()
	beforeSamples := stream.SamplesSeen()
	beforeOutput := stream.Output()

	missingTime := baseTime.AddDate(1, 0, 0)
	missing := stream.Update(math.NaN(), missingTime)
	if !missing.Heatmap.IsEmpty() || !missing.Heatmap.Time.Equal(missingTime) || !math.IsNaN(missing.DominantCycle) {
		t.Fatalf("unexpected NaN sample output")
	}

	for _, invalidSample := range []float64{math.Inf(1), math.Inf(-1), 1e308} {
		if _, updateErr := stream.TryUpdate(invalidSample, baseTime); updateErr == nil {
			t.Fatalf("expected error for invalid sample %v", invalidSample)
		}
		afterState := stream.State()
		if stream.SamplesSeen() != beforeSamples ||
			afterState.HighPass != beforeState.HighPass ||
			afterState.CleanedData != beforeState.CleanedData ||
			afterState.MaximumPower != beforeState.MaximumPower ||
			afterState.DominantCycle != beforeState.DominantCycle ||
			!slices.Equal(afterState.Decibels, beforeState.Decibels) {
			t.Fatalf("state mutated on invalid sample %v", invalidSample)
		}
		assertSameSpectrumOutput(t, stream.Output(), beforeOutput)
	}

	t80 := baseTime.AddDate(0, 0, 80)
	assertSameSpectrumOutput(t, stream.Update(testInput[80], t80), control.Update(testInput[80], t80))

	t81 := baseTime.AddDate(0, 0, 81)
	snapshot := stream.Update(testInput[81], t81)
	control.Update(testInput[81], t81)
	mutated := snapshot.Heatmap.Values
	mutated[0] = -999.0
	t82 := baseTime.AddDate(0, 0, 82)
	assertSameSpectrumOutput(t, stream.Update(testInput[82], t82), control.Update(testInput[82], t82))

	stream.Reset()
	fresh, err := NewTransformedDiscreteFourierTransformSpectrumDefault()
	if err != nil {
		t.Fatalf("construct failed: %v", err)
	}
	for barIndex, closePrice := range testInput {
		timestamp := baseTime.AddDate(0, 0, barIndex)
		assertSameSpectrumOutput(t, stream.Update(closePrice, timestamp), fresh.Update(closePrice, timestamp))
	}
}

func TestParameterValidation(t *testing.T) {
	t.Parallel()

	invalidCases := []Params{
		{MinPeriod: 7, MaxPeriod: 50, WindowLength: 50, HighpassPeriod: 40},
		{MinPeriod: 20, MaxPeriod: 19, WindowLength: 50, HighpassPeriod: 40},
		{MinPeriod: 8, MaxPeriod: 50, WindowLength: 0, HighpassPeriod: 40},
		{MinPeriod: 8, MaxPeriod: 50, WindowLength: 50, HighpassPeriod: 4.0},
		{MinPeriod: 8, MaxPeriod: 50, WindowLength: 50, HighpassPeriod: -10.0},
		{MinPeriod: 8, MaxPeriod: 50, WindowLength: 50, HighpassPeriod: math.NaN()},
		{MinPeriod: 8, MaxPeriod: 50, WindowLength: 50, HighpassPeriod: math.Inf(1)},
	}

	for caseIndex, invalid := range invalidCases {
		if _, err := NewTransformedDiscreteFourierTransformSpectrumParams(&invalid); err == nil {
			t.Fatalf("case %d: expected validation error for %+v", caseIndex, invalid)
		}
	}
}

func TestMnemonicsMetadataAndDescriptor(t *testing.T) {
	t.Parallel()

	cases := []struct {
		params Params
		want   string
		wantDC string
	}{
		{*DefaultParams(), "tdfts(8, 50, 50, 40)", "tdfts-dc(8, 50, 50, 40)"},
		{
			Params{MinPeriod: 10, MaxPeriod: 30, WindowLength: 48, HighpassPeriod: 40.5, BarComponent: entities.BarMedianPrice},
			"tdfts(10, 30, 48, 40.5, hl/2)",
			"tdfts-dc(10, 30, 48, 40.5, hl/2)",
		},
		{
			Params{MinPeriod: 8, MaxPeriod: 50, WindowLength: 50, HighpassPeriod: 40, QuoteComponent: entities.QuoteAskPrice},
			"tdfts(8, 50, 50, 40, a)",
			"tdfts-dc(8, 50, 50, 40, a)",
		},
		{
			Params{MinPeriod: 8, MaxPeriod: 50, WindowLength: 50, HighpassPeriod: 40, TradeComponent: entities.TradeVolume},
			"tdfts(8, 50, 50, 40, v)",
			"tdfts-dc(8, 50, 50, 40, v)",
		},
		{
			Params{
				MinPeriod:      10,
				MaxPeriod:      30,
				WindowLength:   48,
				HighpassPeriod: 40.5,
				BarComponent:   entities.BarMedianPrice,
				QuoteComponent: entities.QuoteAskPrice,
			},
			"tdfts(10, 30, 48, 40.5, hl/2, a)",
			"tdfts-dc(10, 30, 48, 40.5, hl/2, a)",
		},
		{
			Params{
				MinPeriod:      8,
				MaxPeriod:      50,
				WindowLength:   50,
				HighpassPeriod: 40,
				QuoteComponent: entities.QuoteBidPrice,
				TradeComponent: entities.TradeVolume,
			},
			"tdfts(8, 50, 50, 40, b, v)",
			"tdfts-dc(8, 50, 50, 40, b, v)",
		},
	}

	for _, tc := range cases {
		if got := tc.params.Mnemonic(); got != tc.want {
			t.Fatalf("Mnemonic() = %q, want %q", got, tc.want)
		}
		if got := tc.params.DominantCycleMnemonic(); got != tc.wantDC {
			t.Fatalf("DominantCycleMnemonic() = %q, want %q", got, tc.wantDC)
		}
		indicator, err := NewTransformedDiscreteFourierTransformSpectrumParams(&tc.params)
		if err != nil {
			t.Fatalf("construct failed: %v", err)
		}
		meta := indicator.Metadata()
		if meta.Identifier != core.TransformedDiscreteFourierTransformSpectrum {
			t.Fatalf("Identifier = %v, want %v", meta.Identifier, core.TransformedDiscreteFourierTransformSpectrum)
		}
		if len(meta.Outputs) != 2 {
			t.Fatalf("len(Outputs) = %d, want 2", len(meta.Outputs))
		}
		if meta.Outputs[0].Kind != int(Value) || meta.Outputs[0].Shape != shape.Heatmap || meta.Outputs[0].Mnemonic != tc.want {
			t.Fatalf("unexpected primary output metadata: %+v", meta.Outputs[0])
		}
		if meta.Outputs[1].Kind != int(DominantCycle) || meta.Outputs[1].Shape != shape.Scalar || meta.Outputs[1].Mnemonic != tc.wantDC {
			t.Fatalf("unexpected secondary output metadata: %+v", meta.Outputs[1])
		}
	}

	desc, ok := core.DescriptorOf(core.TransformedDiscreteFourierTransformSpectrum)
	if !ok {
		t.Fatalf("descriptor missing")
	}
	if desc.Family != "John Ehlers" ||
		desc.Adaptivity != core.Static ||
		desc.InputRequirement != core.ScalarInput ||
		desc.VolumeUsage != core.NoVolume ||
		len(desc.Outputs) != 2 ||
		desc.Outputs[0].Role != core.Spectrum ||
		desc.Outputs[0].Pane != core.Own ||
		desc.Outputs[1].Role != core.CyclePeriod ||
		desc.Outputs[1].Pane != core.Own {
		t.Fatalf("unexpected descriptor: %+v", desc)
	}
}

func TestEntityUpdates(t *testing.T) {
	t.Parallel()

	scalarInd, err := NewTransformedDiscreteFourierTransformSpectrumDefault()
	if err != nil {
		t.Fatalf("construct failed: %v", err)
	}
	barInd, err := NewTransformedDiscreteFourierTransformSpectrumDefault()
	if err != nil {
		t.Fatalf("construct failed: %v", err)
	}
	quoteInd, err := NewTransformedDiscreteFourierTransformSpectrumDefault()
	if err != nil {
		t.Fatalf("construct failed: %v", err)
	}
	tradeInd, err := NewTransformedDiscreteFourierTransformSpectrumDefault()
	if err != nil {
		t.Fatalf("construct failed: %v", err)
	}
	openParams := *DefaultParams()
	openParams.BarComponent = entities.BarOpenPrice
	openInd, err := NewTransformedDiscreteFourierTransformSpectrumParams(&openParams)
	if err != nil {
		t.Fatalf("construct failed: %v", err)
	}
	openRef, err := NewTransformedDiscreteFourierTransformSpectrumDefault()
	if err != nil {
		t.Fatalf("construct failed: %v", err)
	}

	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for barIndex, closePrice := range testInput {
		timestamp := baseTime.AddDate(0, 0, barIndex)
		openPrice := testOpen[barIndex]
		highPrice := testHigh[barIndex]
		lowPrice := testLow[barIndex]

		outScalar := scalarInd.UpdateScalar(&entities.Scalar{Time: timestamp, Value: closePrice})
		outBar := barInd.UpdateBar(&entities.Bar{
			Time:   timestamp,
			Open:   openPrice,
			High:   highPrice,
			Low:    lowPrice,
			Close:  closePrice,
			Volume: 1000.0,
		})
		outQuote := quoteInd.UpdateQuote(&entities.Quote{
			Time: timestamp,
			Bid:  closePrice,
			Ask:  closePrice,
		})
		outTrade := tradeInd.UpdateTrade(&entities.Trade{
			Time:  timestamp,
			Price: closePrice,
		})

		hmScalar := outScalar[0].(*outputs.Heatmap)
		scScalar := outScalar[1].(entities.Scalar)
		for _, candidate := range []core.Output{outBar, outQuote, outTrade} {
			hmCandidate := candidate[0].(*outputs.Heatmap)
			scCandidate := candidate[1].(entities.Scalar)
			assertSameSpectrumOutput(
				t,
				SpectrumOutput{Heatmap: hmCandidate, DominantCycle: scCandidate.Value},
				SpectrumOutput{Heatmap: hmScalar, DominantCycle: scScalar.Value},
			)
		}

		outOpenBar := openInd.UpdateBar(&entities.Bar{
			Time:   timestamp,
			Open:   openPrice,
			High:   highPrice,
			Low:    lowPrice,
			Close:  closePrice,
			Volume: 1000.0,
		})
		outOpenRef := openRef.UpdateScalar(&entities.Scalar{Time: timestamp, Value: openPrice})
		assertSameSpectrumOutput(
			t,
			SpectrumOutput{Heatmap: outOpenBar[0].(*outputs.Heatmap), DominantCycle: outOpenBar[1].(entities.Scalar).Value},
			SpectrumOutput{Heatmap: outOpenRef[0].(*outputs.Heatmap), DominantCycle: outOpenRef[1].(entities.Scalar).Value},
		)
	}
}
