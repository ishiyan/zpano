package supersmoother

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
		{Params{ShortestCyclePeriod: 2}, expectedP2, 2},
		{Params{ShortestCyclePeriod: 3}, expectedP3, 2},
		{Params{ShortestCyclePeriod: 5}, expectedP5, 2},
		{Params{ShortestCyclePeriod: 8}, expectedP8, 2},
		{Params{ShortestCyclePeriod: 10}, expectedP10, 2},
		{Params{ShortestCyclePeriod: 10.5}, expectedP10_5, 2},
		{Params{ShortestCyclePeriod: 20}, expectedP20, 2},
		{Params{ShortestCyclePeriod: 40}, expectedP40, 2},
		{Params{ShortestCyclePeriod: 80}, expectedP80, 2},
		{Params{ShortestCyclePeriod: 252}, expectedP252, 2},
	}
}

func TestSuperSmootherReferenceCases(t *testing.T) {
	t.Parallel()

	if len(testInput) != 252 {
		t.Fatalf("expected 252 input bars, got %d", len(testInput))
	}

	for _, item := range allReferenceCases() {
		caseItem := item
		t.Run(caseItem.params.Mnemonic(), func(t *testing.T) {
			t.Parallel()

			filter, err := NewSuperSmoother(&caseItem.params)
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
				bound := math.Max(1e-11, math.Abs(target)*1e-11)
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

func TestSuperSmootherNaNAndReset(t *testing.T) {
	t.Parallel()

	for _, item := range allReferenceCases() {
		caseItem := item
		t.Run(caseItem.params.Mnemonic(), func(t *testing.T) {
			t.Parallel()

			first, err := NewSuperSmoother(&caseItem.params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			second, err := NewSuperSmoother(&caseItem.params)
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
			third, err := NewSuperSmoother(&caseItem.params)
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

func TestSuperSmootherInvariants(t *testing.T) {
	t.Parallel()

	constantFilter, err := NewSuperSmootherDefault()
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
	if math.Abs(lastConstant-7.0) > 1e-12 {
		t.Fatalf("expected constant input convergence to 7, got %v", lastConstant)
	}

	nyquistFilter, err := NewSuperSmootherDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var lastNyquist float64
	for index := range 500 {
		sample := -1.0
		if index%2 == 0 {
			sample = 1.0
		}
		lastNyquist = nyquistFilter.Update(sample)
	}
	if math.Abs(lastNyquist) > 1e-12 {
		t.Fatalf("expected Nyquist alternating input suppression to 0, got %v", lastNyquist)
	}
}

func TestSuperSmootherValidationAndInf(t *testing.T) {
	t.Parallel()

	invalidCases := []Params{
		{ShortestCyclePeriod: 1},
		{ShortestCyclePeriod: 1.999},
		{ShortestCyclePeriod: 0},
		{ShortestCyclePeriod: -2},
		{ShortestCyclePeriod: math.NaN()},
		{ShortestCyclePeriod: math.Inf(1)},
		{ShortestCyclePeriod: math.Inf(-1)},
		{ShortestCyclePeriod: 10, BarComponent: entities.BarComponent(9999)},
		{ShortestCyclePeriod: 10, QuoteComponent: entities.QuoteComponent(9999)},
		{ShortestCyclePeriod: 10, TradeComponent: entities.TradeComponent(9999)},
	}

	for index, params := range invalidCases {
		filter, err := NewSuperSmoother(&params)
		if err == nil || filter != nil {
			t.Fatalf("case %d: expected validation error, got filter=%v err=%v", index, filter, err)
		}
	}

	filter, err := NewSuperSmoother(nil)
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

func TestSuperSmootherMetadataMnemonicsAndEntities(t *testing.T) {
	t.Parallel()

	filter, err := NewSuperSmootherDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	meta := filter.Metadata()
	if meta.Identifier != core.SuperSmoother {
		t.Fatalf("expected Identifier %v, got %v", core.SuperSmoother, meta.Identifier)
	}
	if meta.Mnemonic != "ss(10)" {
		t.Fatalf("expected mnemonic ss(10), got %s", meta.Mnemonic)
	}
	if meta.Description != "Super Smoother ss(10)" {
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
	if meta.Outputs[0].Mnemonic != "ss(10)" {
		t.Fatalf("unexpected output mnemonic: %s", meta.Outputs[0].Mnemonic)
	}
	if meta.Outputs[0].Description != "Super Smoother ss(10)" {
		t.Fatalf("unexpected output description: %s", meta.Outputs[0].Description)
	}

	mnemonicCases := []struct {
		params   Params
		expected string
	}{
		{*DefaultParams(), "ss(10)"},
		{Params{ShortestCyclePeriod: 10.5}, "ss(10.5)"},
		{Params{ShortestCyclePeriod: 10, BarComponent: entities.BarMedianPrice}, "ss(10, hl/2)"},
		{Params{ShortestCyclePeriod: 10, QuoteComponent: entities.QuoteBidPrice}, "ss(10, b)"},
		{Params{ShortestCyclePeriod: 10, TradeComponent: entities.TradeVolume}, "ss(10, v)"},
		{Params{ShortestCyclePeriod: 10, BarComponent: entities.BarOpenPrice, QuoteComponent: entities.QuoteAskPrice}, "ss(10, o, a)"},
		{Params{ShortestCyclePeriod: 10, BarComponent: entities.BarTypicalPrice, TradeComponent: entities.TradeVolume}, "ss(10, hlc/3, v)"},
		{Params{ShortestCyclePeriod: 10, QuoteComponent: entities.QuoteWeightedPrice, TradeComponent: entities.TradeVolume}, "ss(10, (bbs+aas)/(bs+as), v)"},
	}

	for _, mnemonicCase := range mnemonicCases {
		instance, createErr := NewSuperSmoother(&mnemonicCase.params)
		if createErr != nil {
			t.Fatalf("unexpected error for %s: %v", mnemonicCase.expected, createErr)
		}
		if instance.LineIndicator.Mnemonic != mnemonicCase.expected {
			t.Fatalf("expected %s, got %s", mnemonicCase.expected, instance.LineIndicator.Mnemonic)
		}
	}

	barFilter, _ := NewSuperSmootherDefault()
	scalarFilter, _ := NewSuperSmootherDefault()
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

	quoteFilter, _ := NewSuperSmootherDefault()
	var quoteOut core.Output
	for range 5 {
		quoteOut = quoteFilter.UpdateQuote(&entities.Quote{
			Time: timestamp, Bid: 99, Ask: 101, BidSize: 1, AskSize: 1,
		})
	}
	if math.IsNaN(quoteOut[0].(entities.Scalar).Value) {
		t.Fatal("expected primed quote output")
	}

	tradeFilter, _ := NewSuperSmootherDefault()
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
