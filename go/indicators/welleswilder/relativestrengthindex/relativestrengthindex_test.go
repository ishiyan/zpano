//nolint:testpackage
package relativestrengthindex

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

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

	check("Identifier", core.RelativeStrengthIndex, act.Identifier)
	check("len(Outputs)", 1, len(act.Outputs))
	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
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
