//nolint:testpackage
package rateofchange

//nolint: gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func testRateOfChangeTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

func TestRateOfChangeUpdate(t *testing.T) { //nolint: funlen
	t.Parallel()

	check := func(index int, exp, act float64) {
		t.Helper()

		if math.Abs(exp-act) > 1e-2 {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	checkNaN := func(index int, act float64) {
		t.Helper()

		if !math.IsNaN(act) {
			t.Errorf("[%v] is incorrect: expected NaN, actual %v", index, act)
		}
	}

	input := testRateOfChangeInput()

	t.Run("length = 14", func(t *testing.T) {
		const ( // Values from index=0 to index=13 are NaN.
			i14value  = -0.546  // Index=14 value.
			i15value  = -2.109  // Index=15 value.
			i16value  = -5.53   // Index=16 value.
			i251value = -1.0367 // Index=251 (last) value.
		)

		t.Parallel()
		roc := testRateOfChangeCreate(14)

		for i := 0; i < 13; i++ {
			checkNaN(i, roc.Update(input[i]))
		}

		for i := 13; i < len(input); i++ {
			act := roc.Update(input[i])

			switch i {
			case 14:
				check(i, i14value, act)
			case 15:
				check(i, i15value, act)
			case 16:
				check(i, i16value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, roc.Update(math.NaN()))
	})
}

func TestRateOfChangeUpdateEntity(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		l   = 2
		inp = 3.
		exp = 0.
	)

	time := testRateOfChangeTime()
	check := func(act core.Output) {
		t.Helper()

		if len(act) != 1 {
			t.Errorf("len(output) is incorrect: expected 1, actual %v", len(act))
		}

		s, ok := act[0].(entities.Scalar)
		if !ok {
			t.Error("output is not scalar")
		}

		if s.Time != time {
			t.Errorf("time is incorrect: expected %v, actual %v", time, s.Time)
		}

		if math.Abs(s.Value-exp) > 1e-13 {
			t.Errorf("value is incorrect: expected %v, actual %v", exp, s.Value)
		}
	}

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		s := entities.Scalar{Time: time, Value: inp}
		roc := testRateOfChangeCreate(l)
		roc.Update(inp)
		roc.Update(inp)
		check(roc.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: time, Close: inp}
		roc := testRateOfChangeCreate(l)
		roc.Update(inp)
		roc.Update(inp)
		check(roc.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: time, Bid: inp, Ask: inp}
		roc := testRateOfChangeCreate(l)
		roc.Update(inp)
		roc.Update(inp)
		check(roc.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: time, Price: inp}
		roc := testRateOfChangeCreate(l)
		roc.Update(inp)
		roc.Update(inp)
		check(roc.UpdateTrade(&r))
	})
}

func TestRateOfChangeIsPrimed(t *testing.T) { //nolint:funlen
	t.Parallel()

	input := testRateOfChangeInput()
	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	t.Run("length = 1", func(t *testing.T) {
		t.Parallel()
		roc := testRateOfChangeCreate(1)

		check(-1, false, roc.IsPrimed())

		for i := 0; i < 1; i++ {
			roc.Update(input[i])
			check(i, false, roc.IsPrimed())
		}

		for i := 1; i < len(input); i++ {
			roc.Update(input[i])
			check(i, true, roc.IsPrimed())
		}
	})

	t.Run("length = 2", func(t *testing.T) {
		t.Parallel()
		roc := testRateOfChangeCreate(2)

		check(-1, false, roc.IsPrimed())

		for i := 0; i < 2; i++ {
			roc.Update(input[i])
			check(i, false, roc.IsPrimed())
		}

		for i := 2; i < len(input); i++ {
			roc.Update(input[i])
			check(i, true, roc.IsPrimed())
		}
	})

	t.Run("length = 5", func(t *testing.T) {
		t.Parallel()
		roc := testRateOfChangeCreate(5)

		check(-1, false, roc.IsPrimed())

		for i := 0; i < 5; i++ {
			roc.Update(input[i])
			check(i, false, roc.IsPrimed())
		}

		for i := 5; i < len(input); i++ {
			roc.Update(input[i])
			check(i, true, roc.IsPrimed())
		}
	})

	t.Run("length = 10", func(t *testing.T) {
		t.Parallel()
		roc := testRateOfChangeCreate(10)

		check(-1, false, roc.IsPrimed())

		for i := 0; i < 10; i++ {
			roc.Update(input[i])
			check(i, false, roc.IsPrimed())
		}

		for i := 10; i < len(input); i++ {
			roc.Update(input[i])
			check(i, true, roc.IsPrimed())
		}
	})
}

func TestRateOfChangeMetadata(t *testing.T) {
	t.Parallel()

	roc := testRateOfChangeCreate(5)
	act := roc.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	check("Identifier", core.RateOfChange, act.Identifier)
	check("Mnemonic", "roc(5)", act.Mnemonic)
	check("Description", "Rate of Change roc(5)", act.Description)
	check("len(Outputs)", 1, len(act.Outputs))
	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", "roc(5)", act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", "Rate of Change roc(5)", act.Outputs[0].Description)
}

func TestNewRateOfChange(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		bc     entities.BarComponent   = entities.BarMedianPrice
		qc     entities.QuoteComponent = entities.QuoteMidPrice
		tc     entities.TradeComponent = entities.TradePrice
		length                         = 5
		errlen                         = "invalid rate of change parameters: length should be positive"
		errbc                          = "invalid rate of change parameters: 9999: unknown bar component"
		errqc                          = "invalid rate of change parameters: 9999: unknown quote component"
		errtc                          = "invalid rate of change parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("length > 0", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		roc, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "roc(5, hl/2)", roc.LineIndicator.Mnemonic)
		check("description", "Rate of Change roc(5, hl/2)", roc.LineIndicator.Description)
		check("primed", false, roc.primed)
		check("lastIndex", length, roc.lastIndex)
		check("len(window)", length+1, len(roc.window))
		check("windowLength", length+1, roc.windowLength)
		check("windowCount", 0, roc.windowCount)
	})

	t.Run("length = 1", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: 1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		roc, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "roc(1, hl/2)", roc.LineIndicator.Mnemonic)
		check("description", "Rate of Change roc(1, hl/2)", roc.LineIndicator.Description)
		check("primed", false, roc.primed)
		check("lastIndex", 1, roc.lastIndex)
		check("len(window)", 2, len(roc.window))
		check("windowLength", 2, roc.windowLength)
		check("windowCount", 0, roc.windowCount)
	})

	t.Run("length = 0", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: 0, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		roc, err := New(&params)
		check("roc == nil", true, roc == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length < 0", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: -1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		roc, err := New(&params)
		check("roc == nil", true, roc == nil)
		check("err", errlen, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		roc, err := New(&params)
		check("roc == nil", true, roc == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		roc, err := New(&params)
		check("roc == nil", true, roc == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		roc, err := New(&params)
		check("roc == nil", true, roc == nil)
		check("err", errtc, err.Error())
	})

	// Zero-value component mnemonic tests.
	// A zero value means "use default, don't show in mnemonic".

	t.Run("all components zero", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length}

		roc, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "roc(5)", roc.LineIndicator.Mnemonic)
		check("description", "Rate of Change roc(5)", roc.LineIndicator.Description)
	})

	t.Run("only bar component set", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length, BarComponent: entities.BarMedianPrice}

		roc, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "roc(5, hl/2)", roc.LineIndicator.Mnemonic)
		check("description", "Rate of Change roc(5, hl/2)", roc.LineIndicator.Description)
	})

	t.Run("only quote component set", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length, QuoteComponent: entities.QuoteBidPrice}

		roc, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "roc(5, b)", roc.LineIndicator.Mnemonic)
		check("description", "Rate of Change roc(5, b)", roc.LineIndicator.Description)
	})

	t.Run("only trade component set", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length, TradeComponent: entities.TradeVolume}

		roc, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "roc(5, v)", roc.LineIndicator.Mnemonic)
		check("description", "Rate of Change roc(5, v)", roc.LineIndicator.Description)
	})
}

func testRateOfChangeCreate(length int) *RateOfChange {
	params := Params{Length: length}

	roc, _ := New(&params)

	return roc
}
