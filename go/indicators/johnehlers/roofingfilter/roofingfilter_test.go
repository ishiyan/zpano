package roofingfilter

import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

type referenceCase struct {
	params         Params
	expected       []float64
	expectedWarmup int
}

func allReferenceCases() []referenceCase {
	return []referenceCase{
		{Params{ShortestCyclePeriod: 10, LongestCyclePeriod: 48, HighpassPoles: 2, ZeroMean: false}, expectedLP10_HP48_PO2_ZM0, 4},
		{Params{ShortestCyclePeriod: 40, LongestCyclePeriod: 80, HighpassPoles: 2, ZeroMean: false}, expectedLP40_HP80_PO2_ZM0, 4},
		{Params{ShortestCyclePeriod: 2, LongestCyclePeriod: 3, HighpassPoles: 2, ZeroMean: false}, expectedLP2_HP3_PO2_ZM0, 4},
		{Params{ShortestCyclePeriod: 5, LongestCyclePeriod: 24, HighpassPoles: 2, ZeroMean: false}, expectedLP5_HP24_PO2_ZM0, 4},
		{Params{ShortestCyclePeriod: 10.5, LongestCyclePeriod: 48, HighpassPoles: 2, ZeroMean: false}, expectedLP10_5_HP48_PO2_ZM0, 4},
		{Params{ShortestCyclePeriod: 10, LongestCyclePeriod: 11, HighpassPoles: 2, ZeroMean: false}, expectedLP10_HP11_PO2_ZM0, 4},
		{Params{ShortestCyclePeriod: 20, LongestCyclePeriod: 80, HighpassPoles: 2, ZeroMean: false}, expectedLP20_HP80_PO2_ZM0, 4},
		{Params{ShortestCyclePeriod: 40, LongestCyclePeriod: 252, HighpassPoles: 2, ZeroMean: false}, expectedLP40_HP252_PO2_ZM0, 4},
		{Params{ShortestCyclePeriod: 10, LongestCyclePeriod: 48, HighpassPoles: 1, ZeroMean: false}, expectedLP10_HP48_PO1_ZM0, 3},
		{Params{ShortestCyclePeriod: 2, LongestCyclePeriod: 4, HighpassPoles: 1, ZeroMean: false}, expectedLP2_HP4_PO1_ZM0, 3},
		{Params{ShortestCyclePeriod: 10, LongestCyclePeriod: 48, HighpassPoles: 1, ZeroMean: true}, expectedLP10_HP48_PO1_ZM1, 4},
		{Params{ShortestCyclePeriod: 2, LongestCyclePeriod: 4, HighpassPoles: 1, ZeroMean: true}, expectedLP2_HP4_PO1_ZM1, 4},
	}
}

func TestRoofingFilterReferenceCases(t *testing.T) {
	t.Parallel()

	if len(testInput) != 252 {
		t.Fatalf("expected 252 input bars, got %d", len(testInput))
	}

	for _, tc := range allReferenceCases() {
		caseItem := tc
		t.Run(caseItem.params.Mnemonic(), func(t *testing.T) {
			t.Parallel()

			filter, err := NewRoofingFilter(&caseItem.params)
			if err != nil {
				t.Fatalf("unexpected constructor error: %v", err)
			}
			if filter.WarmupBars() != caseItem.expectedWarmup {
				t.Fatalf("expected warmup %d, got %d", caseItem.expectedWarmup, filter.WarmupBars())
			}
			if filter.IsPrimed() {
				t.Fatal("expected unprimed before updates")
			}
			if !math.IsNaN(filter.Value()) {
				t.Fatalf("expected NaN initial value, got %v", filter.Value())
			}

			for index, sample := range testInput {
				actual := filter.Update(sample)
				target := caseItem.expected[index]

				if index < caseItem.expectedWarmup {
					if filter.IsPrimed() {
						t.Fatalf("bar %d: unexpectedly primed", index)
					}
					if !math.IsNaN(actual) {
						t.Fatalf("bar %d: expected NaN, got %v", index, actual)
					}
					if !math.IsNaN(filter.Value()) {
						t.Fatalf("bar %d: expected NaN Value(), got %v", index, filter.Value())
					}
					continue
				}

				if !filter.IsPrimed() {
					t.Fatalf("bar %d: expected primed", index)
				}
				if math.IsNaN(target) {
					t.Fatalf("bar %d: fixture target unexpectedly NaN", index)
				}
				bound := math.Max(1e-12, math.Abs(target)*1e-12)
				if math.Abs(actual-target) > bound {
					t.Fatalf("bar %d: expected %v, got %v", index, target, actual)
				}
				if math.Abs(filter.Value()-actual) > 1e-15 {
					t.Fatalf("bar %d: Value() %v != actual %v", index, filter.Value(), actual)
				}
			}
		})
	}
}

func TestRoofingFilterNaNAndReset(t *testing.T) {
	t.Parallel()

	for _, tc := range allReferenceCases() {
		caseItem := tc
		t.Run(caseItem.params.Mnemonic(), func(t *testing.T) {
			t.Parallel()

			first, err := NewRoofingFilter(&caseItem.params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			second, err := NewRoofingFilter(&caseItem.params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for index, sample := range testInput {
				if !math.IsNaN(first.Update(math.NaN())) {
					t.Fatalf("bar %d: expected NaN for NaN input", index)
				}
				left := first.Update(sample)
				right := second.Update(sample)
				if first.IsPrimed() != second.IsPrimed() {
					t.Fatalf("bar %d: primed mismatch", index)
				}
				if math.IsNaN(right) {
					if !math.IsNaN(left) {
						t.Fatalf("bar %d: expected NaN, got %v", index, left)
					}
				} else if math.Abs(left-right) > 1e-15 {
					t.Fatalf("bar %d: %v != %v", index, left, right)
				}
			}

			first.Reset()
			if first.IsPrimed() || !math.IsNaN(first.Value()) {
				t.Fatal("expected unprimed NaN state after Reset()")
			}
			third, err := NewRoofingFilter(&caseItem.params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for index, sample := range testInput {
				left := first.Update(sample)
				right := third.Update(sample)
				if math.IsNaN(right) {
					if !math.IsNaN(left) {
						t.Fatalf("bar %d after reset: expected NaN, got %v", index, left)
					}
				} else if math.Abs(left-right) > 1e-15 {
					t.Fatalf("bar %d after reset: %v != %v", index, left, right)
				}
			}
		})
	}
}

func TestRoofingFilterInvariants(t *testing.T) {
	t.Parallel()

	constantFilter, err := NewRoofingFilterDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	firstFinite := -1
	var lastConstant float64
	for index := range 500 {
		out := constantFilter.Update(7.0)
		if firstFinite == -1 && !math.IsNaN(out) {
			firstFinite = index
		}
		lastConstant = out
	}
	if firstFinite != constantFilter.WarmupBars() {
		t.Fatalf("expected first finite at %d, got %d", constantFilter.WarmupBars(), firstFinite)
	}
	if math.Abs(lastConstant) > 1e-12 {
		t.Fatalf("expected constant input decay to 0, got %v", lastConstant)
	}

	rampFilter, err := NewRoofingFilterDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var lastRamp float64
	for index := range 1000 {
		lastRamp = rampFilter.Update(float64(index))
	}
	if math.Abs(lastRamp) > 1e-12 {
		t.Fatalf("expected 2-pole linear ramp rejection, got %v", lastRamp)
	}

	boundary, err := NewRoofingFilter(&Params{
		ShortestCyclePeriod: 2,
		LongestCyclePeriod:  4,
		HighpassPoles:       1,
	})
	if err != nil {
		t.Fatalf("unexpected boundary error: %v", err)
	}
	if math.Abs(boundary.gain-0.5) > 1e-15 {
		t.Fatalf("expected gain 0.5, got %v", boundary.gain)
	}
	if math.Abs(boundary.decay) > 1e-15 {
		t.Fatalf("expected decay 0, got %v", boundary.decay)
	}
}

func TestRoofingFilterValidationAndInf(t *testing.T) {
	t.Parallel()

	invalidCases := []Params{
		{ShortestCyclePeriod: 1, LongestCyclePeriod: 48, HighpassPoles: 2},
		{ShortestCyclePeriod: 0, LongestCyclePeriod: 48, HighpassPoles: 2},
		{ShortestCyclePeriod: -2, LongestCyclePeriod: 48, HighpassPoles: 2},
		{ShortestCyclePeriod: math.NaN(), LongestCyclePeriod: 48, HighpassPoles: 2},
		{ShortestCyclePeriod: math.Inf(1), LongestCyclePeriod: 48, HighpassPoles: 2},
		{ShortestCyclePeriod: 10, LongestCyclePeriod: 10, HighpassPoles: 2},
		{ShortestCyclePeriod: 10, LongestCyclePeriod: 5, HighpassPoles: 2},
		{ShortestCyclePeriod: 10, LongestCyclePeriod: math.Inf(1), HighpassPoles: 2},
		{ShortestCyclePeriod: 10, LongestCyclePeriod: math.NaN(), HighpassPoles: 2},
		{ShortestCyclePeriod: 10, LongestCyclePeriod: 48, HighpassPoles: 0},
		{ShortestCyclePeriod: 10, LongestCyclePeriod: 48, HighpassPoles: 3},
		{ShortestCyclePeriod: 10, LongestCyclePeriod: 48, HighpassPoles: 2, ZeroMean: true},
		{ShortestCyclePeriod: 10, LongestCyclePeriod: 48, HighpassPoles: 2, BarComponent: entities.BarComponent(9999)},
		{ShortestCyclePeriod: 10, LongestCyclePeriod: 48, HighpassPoles: 2, QuoteComponent: entities.QuoteComponent(9999)},
		{ShortestCyclePeriod: 10, LongestCyclePeriod: 48, HighpassPoles: 2, TradeComponent: entities.TradeComponent(9999)},
	}

	for index, params := range invalidCases {
		filter, err := NewRoofingFilter(&params)
		if err == nil || filter != nil {
			t.Fatalf("case %d: expected validation error, got filter=%v err=%v", index, filter, err)
		}
	}

	filter, err := NewRoofingFilter(nil)
	if err != nil {
		t.Fatalf("expected nil params to use DefaultParams, got %v", err)
	}
	if _, infErr := filter.TryUpdate(math.Inf(1)); infErr == nil {
		t.Fatal("expected error on +Inf TryUpdate")
	}
	if _, negInfErr := filter.TryUpdate(math.Inf(-1)); negInfErr == nil {
		t.Fatal("expected error on -Inf TryUpdate")
	}
	if filter.IsPrimed() {
		t.Fatal("expected filter to remain unprimed after rejected infinite inputs")
	}

	didPanic := false
	func() {
		defer func() {
			if recover() != nil {
				didPanic = true
			}
		}()
		filter.Update(math.Inf(1))
	}()
	if !didPanic {
		t.Fatal("expected Update(+Inf) to panic")
	}
}

func TestRoofingFilterMetadataMnemonicsAndEntities(t *testing.T) {
	t.Parallel()

	filter, err := NewRoofingFilterDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	meta := filter.Metadata()
	if meta.Identifier != core.RoofingFilter {
		t.Fatalf("expected Identifier %v, got %v", core.RoofingFilter, meta.Identifier)
	}
	if meta.Mnemonic != "roof2hp(10, 48)" {
		t.Fatalf("expected mnemonic roof2hp(10, 48), got %s", meta.Mnemonic)
	}
	if meta.Description != "Roofing Filter roof2hp(10, 48)" {
		t.Fatalf("unexpected description: %s", meta.Description)
	}
	if len(meta.Outputs) != 1 {
		t.Fatalf("expected 1 output, got %d", len(meta.Outputs))
	}
	if meta.Outputs[0].Kind != int(Value) {
		t.Fatalf("expected Kind %d, got %d", int(Value), meta.Outputs[0].Kind)
	}
	if meta.Outputs[0].Shape != shape.Scalar {
		t.Fatalf("expected Shape %v, got %v", shape.Scalar, meta.Outputs[0].Shape)
	}
	if meta.Outputs[0].Mnemonic != "roof2hp(10, 48)" {
		t.Fatalf("unexpected output mnemonic: %s", meta.Outputs[0].Mnemonic)
	}

	mnemonicCases := []struct {
		params   Params
		expected string
	}{
		{*DefaultParams(), "roof2hp(10, 48)"},
		{Params{ShortestCyclePeriod: 10.5, LongestCyclePeriod: 48, HighpassPoles: 2}, "roof2hp(10.5, 48)"},
		{Params{ShortestCyclePeriod: 10, LongestCyclePeriod: 48, HighpassPoles: 1}, "roof1hp(10, 48)"},
		{Params{ShortestCyclePeriod: 10, LongestCyclePeriod: 48, HighpassPoles: 1, ZeroMean: true}, "roof1hpzm(10, 48)"},
		{Params{ShortestCyclePeriod: 10, LongestCyclePeriod: 48, HighpassPoles: 2, BarComponent: entities.BarMedianPrice}, "roof2hp(10, 48, hl/2)"},
		{Params{ShortestCyclePeriod: 10, LongestCyclePeriod: 48, HighpassPoles: 2, QuoteComponent: entities.QuoteBidPrice}, "roof2hp(10, 48, b)"},
		{Params{ShortestCyclePeriod: 10, LongestCyclePeriod: 48, HighpassPoles: 2, TradeComponent: entities.TradeVolume}, "roof2hp(10, 48, v)"},
		{Params{ShortestCyclePeriod: 10, LongestCyclePeriod: 48, HighpassPoles: 2, BarComponent: entities.BarOpenPrice, QuoteComponent: entities.QuoteAskPrice}, "roof2hp(10, 48, o, a)"},
		{Params{ShortestCyclePeriod: 10, LongestCyclePeriod: 48, HighpassPoles: 2, BarComponent: entities.BarTypicalPrice, TradeComponent: entities.TradeVolume}, "roof2hp(10, 48, hlc/3, v)"},
		{Params{ShortestCyclePeriod: 10, LongestCyclePeriod: 48, HighpassPoles: 2, QuoteComponent: entities.QuoteWeightedPrice, TradeComponent: entities.TradeVolume}, "roof2hp(10, 48, (bbs+aas)/(bs+as), v)"},
	}

	for _, mc := range mnemonicCases {
		inst, createErr := NewRoofingFilter(&mc.params)
		if createErr != nil {
			t.Fatalf("unexpected error for %s: %v", mc.expected, createErr)
		}
		if inst.LineIndicator.Mnemonic != mc.expected {
			t.Fatalf("expected %s, got %s", mc.expected, inst.LineIndicator.Mnemonic)
		}
	}

	barFilter, _ := NewRoofingFilterDefault()
	scalarFilter, _ := NewRoofingFilterDefault()
	timestamp := time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)
	for index := range 10 {
		bar := entities.Bar{
			Time:   timestamp,
			Open:   testOpen[index],
			High:   testHigh[index],
			Low:    testLow[index],
			Close:  testInput[index],
			Volume: 1000,
		}
		barOut := barFilter.UpdateBar(&bar)
		scalarOut := scalarFilter.UpdateScalar(&entities.Scalar{Time: timestamp, Value: testInput[index]})
		barScalar := barOut[0].(entities.Scalar)
		directScalar := scalarOut[0].(entities.Scalar)
		if math.IsNaN(directScalar.Value) {
			if !math.IsNaN(barScalar.Value) {
				t.Fatalf("bar %d: expected NaN bar output", index)
			}
		} else if math.Abs(barScalar.Value-directScalar.Value) > 1e-15 {
			t.Fatalf("bar %d: bar Close output %v != scalar %v", index, barScalar.Value, directScalar.Value)
		}
	}

	quoteFilter, _ := NewRoofingFilterDefault()
	var quoteOut core.Output
	for range 5 {
		quoteOut = quoteFilter.UpdateQuote(&entities.Quote{
			Time: timestamp, Bid: 99, Ask: 101, BidSize: 1, AskSize: 1,
		})
	}
	if math.IsNaN(quoteOut[0].(entities.Scalar).Value) {
		t.Fatal("expected primed quote output")
	}

	tradeFilter, _ := NewRoofingFilterDefault()
	var tradeOut core.Output
	for range 5 {
		tradeOut = tradeFilter.UpdateTrade(&entities.Trade{
			Time: timestamp, Price: 100, Volume: 50,
		})
	}
	if math.IsNaN(tradeOut[0].(entities.Scalar).Value) {
		t.Fatal("expected primed trade output")
	}
}
