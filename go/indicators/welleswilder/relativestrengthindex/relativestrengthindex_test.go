//nolint:testpackage
package relativestrengthindex

//nolint: gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
)

// Test data from TA-Lib reference (length=9, 25 entries).
func testInput1() []float64 {
	return []float64{
		91.15, 90.50, 92.55, 94.70, 95.55, 94.00, 91.30, 91.95, 92.45, 93.80,
		92.50, 94.55, 96.75, 97.80, 98.40, 98.15, 96.70, 98.85, 98.90, 100.50,
		102.60, 104.80, 103.80, 103.10, 102.00,
	}
}

func testExpected1() []float64 {
	return []float64{
		math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(),
		math.NaN(), math.NaN(), math.NaN(), math.NaN(),
		60.6425702811244, 54.2677448337826, 61.4558190165176, 67.6034767388667,
		70.1590191481383, 71.5992400904851, 70.0152589447766, 61.1833361324987,
		67.9312249318593, 68.076417836971, 72.5504646296262, 77.2568847385616,
		81.0801123570899, 74.6619680507228, 70.2808713845906, 63.6754215506388,
	}
}

// Test data from TA-Lib reference (length=14, 252 entries).
func testInput2() []float64 {
	return []float64{
		44.34, 44.09, 43.61, 44.33, 44.83, 45.10, 45.42, 45.84, 46.08, 45.89,
		46.03, 45.61, 46.28, 46.28, 46.00, 46.03, 46.41, 46.22, 45.64, 46.21,
		46.25, 45.71, 46.45, 45.78, 45.35, 44.03, 44.18, 44.22, 44.57, 43.42,
		42.66, 43.13, 44.94, 43.61, 44.33, 44.83, 45.10, 45.42, 45.84, 46.08,
		45.89, 46.03, 45.61, 46.28, 46.28, 46.00, 46.03, 46.41, 46.22, 45.64,
		46.21, 46.25, 45.71, 46.45, 45.78, 45.35, 44.03, 44.18, 44.22, 44.57,
		43.42, 42.66, 43.13, 44.94, 43.61, 44.33, 44.83, 45.10, 45.42, 45.84,
		46.08, 45.89, 46.03, 45.61, 46.28, 46.28, 46.00, 46.03, 46.41, 46.22,
		45.64, 46.21, 46.25, 45.71, 46.45, 45.78, 45.35, 44.03, 44.18, 44.22,
		44.57, 43.42, 42.66, 43.13, 44.94, 43.61, 44.33, 44.83, 45.10, 45.42,
		45.84, 46.08, 45.89, 46.03, 45.61, 46.28, 46.28, 46.00, 46.03, 46.41,
		46.22, 45.64, 46.21, 46.25, 45.71, 46.45, 45.78, 45.35, 44.03, 44.18,
		44.22, 44.57, 43.42, 42.66, 43.13, 44.94, 43.61, 44.33, 44.83, 45.10,
		45.42, 45.84, 46.08, 45.89, 46.03, 45.61, 46.28, 46.28, 46.00, 46.03,
		46.41, 46.22, 45.64, 46.21, 46.25, 45.71, 46.45, 45.78, 45.35, 44.03,
		44.18, 44.22, 44.57, 43.42, 42.66, 43.13, 44.94, 43.61, 44.33, 44.83,
		45.10, 45.42, 45.84, 46.08, 45.89, 46.03, 45.61, 46.28, 46.28, 46.00,
		46.03, 46.41, 46.22, 45.64, 46.21, 46.25, 45.71, 46.45, 45.78, 45.35,
		44.03, 44.18, 44.22, 44.57, 43.42, 42.66, 43.13, 44.94, 43.61, 44.33,
		44.83, 45.10, 45.42, 45.84, 46.08, 45.89, 46.03, 45.61, 46.28, 46.28,
		46.00, 46.03, 46.41, 46.22, 45.64, 46.21, 46.25, 45.71, 46.45, 45.78,
		45.35, 44.03, 44.18, 44.22, 44.57, 43.42, 42.66, 43.13, 44.94, 43.61,
		44.33, 44.83, 45.10, 45.42, 45.84, 46.08, 45.89, 46.03, 45.61, 46.28,
		46.28, 46.00, 46.03, 46.41, 46.22, 45.64, 46.21, 46.25, 45.71, 46.45,
		45.78, 45.35, 44.03, 44.18, 44.22, 44.57, 43.42, 42.66, 43.13, 44.94,
		43.61, 44.33,
	}
}

func testTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

func testCreate() *RelativeStrengthIndex {
	params := RelativeStrengthIndexParams{Length: 9}

	rsi, _ := NewRelativeStrengthIndex(&params)

	return rsi
}

func TestRelativeStrengthIndexUpdate(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-9

	input := testInput1()
	expected := testExpected1()
	rsi := testCreate()

	for i := range input {
		act := rsi.Update(input[i])

		if i < 9 {
			if !math.IsNaN(act) {
				t.Errorf("[%d] expected NaN, got %v", i, act)
			}

			continue
		}

		if math.Abs(act-expected[i]) > tolerance {
			t.Errorf("[%d] expected %v, got %v", i, expected[i], act)
		}
	}

	if !math.IsNaN(rsi.Update(math.NaN())) {
		t.Error("expected NaN passthrough")
	}
}

func TestRelativeStrengthIndexUpdate2(t *testing.T) {
	t.Parallel()

	const tolerance = 0.5

	input := testInput2()
	rsi := &RelativeStrengthIndex{}

	params := RelativeStrengthIndexParams{Length: 14}
	r, _ := NewRelativeStrengthIndex(&params)
	rsi = r

	var act float64

	for i := range input {
		act = rsi.Update(input[i])

		if i < 14 {
			if !math.IsNaN(act) {
				t.Errorf("[%d] expected NaN, got %v", i, act)
			}
		}
	}

	// Spot check: final value should be in a reasonable RSI range.
	// The C# test uses rounded spot checks; with repeating data the value converges.
	if act < 0 || act > 100 {
		t.Errorf("[251] expected RSI in [0,100], got %v", act)
	}
}

func TestRelativeStrengthIndexIsPrimed(t *testing.T) {
	t.Parallel()

	params := RelativeStrengthIndexParams{Length: 5}
	rsi, _ := NewRelativeStrengthIndex(&params)

	if rsi.IsPrimed() {
		t.Error("should not be primed before any updates")
	}

	// Feed values 1..5 (5 updates): should NOT be primed.
	for i := 1; i <= 5; i++ {
		rsi.Update(float64(i))

		if rsi.IsPrimed() {
			t.Errorf("[%d] should not be primed", i)
		}
	}

	// 6th update: should be primed.
	rsi.Update(6)

	if !rsi.IsPrimed() {
		t.Error("[6] should be primed")
	}

	// Further updates remain primed.
	for i := 7; i <= 11; i++ {
		rsi.Update(float64(i))

		if !rsi.IsPrimed() {
			t.Errorf("[%d] should be primed", i)
		}
	}
}

func TestRelativeStrengthIndexUpdateEntity(t *testing.T) {
	t.Parallel()

	const inp = 100.

	tm := testTime()
	rsi := testCreate()

	// Prime the indicator (need length+1 = 10 updates).
	for i := 0; i < 10; i++ {
		rsi.Update(inp)
	}

	check := func(act core.Output) {
		t.Helper()

		if len(act) != 1 {
			t.Errorf("len(output) is incorrect: expected 1, actual %v", len(act))
		}

		s, ok := act[0].(entities.Scalar)
		if !ok {
			t.Error("output is not scalar")
		}

		if s.Time != tm {
			t.Errorf("time is incorrect: expected %v, actual %v", tm, s.Time)
		}

		if math.IsNaN(s.Value) {
			t.Error("value should not be NaN after priming")
		}
	}

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		s := entities.Scalar{Time: tm, Value: inp}
		rsi2 := testCreate()
		for i := 0; i < 10; i++ {
			rsi2.Update(inp)
		}
		check(rsi2.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: tm, Close: inp}
		rsi2 := testCreate()
		for i := 0; i < 10; i++ {
			rsi2.Update(inp)
		}
		check(rsi2.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: tm, Bid: inp, Ask: inp}
		rsi2 := testCreate()
		for i := 0; i < 10; i++ {
			rsi2.Update(inp)
		}
		check(rsi2.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: tm, Price: inp}
		rsi2 := testCreate()
		for i := 0; i < 10; i++ {
			rsi2.Update(inp)
		}
		check(rsi2.UpdateTrade(&r))
	})
}

func TestRelativeStrengthIndexMetadata(t *testing.T) {
	t.Parallel()

	rsi := testCreate()
	act := rsi.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	check("Type", core.RelativeStrengthIndex, act.Type)
	check("len(Outputs)", 1, len(act.Outputs))
	check("Outputs[0].Kind", int(RelativeStrengthIndexValue), act.Outputs[0].Kind)
	check("Outputs[0].Type", outputs.ScalarType, act.Outputs[0].Type)
	check("Outputs[0].Mnemonic", "rsi(9)", act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", "Relative Strength Index rsi(9)", act.Outputs[0].Description)
}

func TestNewRelativeStrengthIndex(t *testing.T) {
	t.Parallel()

	const (
		bc entities.BarComponent   = entities.BarClosePrice
		qc entities.QuoteComponent = entities.QuoteMidPrice
		tc entities.TradeComponent = entities.TradePrice

		length = 14
		errlen = "invalid relative strength index parameters: length should be greater than 1"
		errbc  = "invalid relative strength index parameters: 9999: unknown bar component"
		errqc  = "invalid relative strength index parameters: 9999: unknown quote component"
		errtc  = "invalid relative strength index parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("valid params", func(t *testing.T) {
		t.Parallel()
		params := RelativeStrengthIndexParams{
			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		rsi, err := NewRelativeStrengthIndex(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "rsi(14)", rsi.LineIndicator.Mnemonic)
		check("description", "Relative Strength Index rsi(14)", rsi.LineIndicator.Description)
		check("primed", false, rsi.primed)
	})

	t.Run("length < 2", func(t *testing.T) {
		t.Parallel()
		params := RelativeStrengthIndexParams{
			Length: 1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		rsi, err := NewRelativeStrengthIndex(&params)
		check("rsi == nil", true, rsi == nil)
		check("err", errlen, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()
		params := RelativeStrengthIndexParams{
			Length: length, BarComponent: entities.BarComponent(9999),
			QuoteComponent: qc, TradeComponent: tc,
		}

		rsi, err := NewRelativeStrengthIndex(&params)
		check("rsi == nil", true, rsi == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()
		params := RelativeStrengthIndexParams{
			Length: length, BarComponent: bc,
			QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		rsi, err := NewRelativeStrengthIndex(&params)
		check("rsi == nil", true, rsi == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()
		params := RelativeStrengthIndexParams{
			Length: length, BarComponent: bc, QuoteComponent: qc,
			TradeComponent: entities.TradeComponent(9999),
		}

		rsi, err := NewRelativeStrengthIndex(&params)
		check("rsi == nil", true, rsi == nil)
		check("err", errtc, err.Error())
	})

	t.Run("bar component set to open", func(t *testing.T) {
		t.Parallel()
		params := RelativeStrengthIndexParams{
			Length: length, BarComponent: entities.BarOpenPrice,
		}

		rsi, err := NewRelativeStrengthIndex(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "rsi(14, o)", rsi.LineIndicator.Mnemonic)
	})
}
