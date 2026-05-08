//nolint:testpackage
package momentum

//nolint: gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func testMomentumTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

func TestMomentumUpdate(t *testing.T) { //nolint: funlen
	t.Parallel()

	check := func(index int, exp, act float64) {
		t.Helper()

		if math.Abs(exp-act) > 1e-13 {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	checkNaN := func(index int, act float64) {
		t.Helper()

		if !math.IsNaN(act) {
			t.Errorf("[%v] is incorrect: expected NaN, actual %v", index, act)
		}
	}

	input := testMomentumInput()

	t.Run("length = 14", func(t *testing.T) {
		const ( // Values from index=0 to index=13 are NaN.
			i14value  = -0.50 // Index=14 value.
			i15value  = -2.00 // Index=15 value.
			i16value  = -5.22 // Index=16 value.
			i251value = -1.13 // Index=251 (last) value.
		)

		t.Parallel()
		mom := testMomentumCreate(14)

		for i := 0; i < 13; i++ {
			checkNaN(i, mom.Update(input[i]))
		}

		for i := 13; i < len(input); i++ {
			act := mom.Update(input[i])

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

		checkNaN(0, mom.Update(math.NaN()))
	})
}

func TestMomentumUpdateEntity(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		l   = 2
		inp = 3.
		exp = 3.
	)

	time := testMomentumTime()
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
		mom := testMomentumCreate(l)
		mom.Update(0.)
		mom.Update(0.)
		check(mom.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: time, Close: inp}
		mom := testMomentumCreate(l)
		mom.Update(0.)
		mom.Update(0.)
		check(mom.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: time, Bid: inp, Ask: inp}
		mom := testMomentumCreate(l)
		mom.Update(0.)
		mom.Update(0.)
		check(mom.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: time, Price: inp}
		mom := testMomentumCreate(l)
		mom.Update(0.)
		mom.Update(0.)
		check(mom.UpdateTrade(&r))
	})
}

func TestMomentumIsPrimed(t *testing.T) { //nolint:funlen
	t.Parallel()

	input := testMomentumInput()
	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	t.Run("length = 1", func(t *testing.T) {
		t.Parallel()
		mom := testMomentumCreate(1)

		check(-1, false, mom.IsPrimed())

		for i := 0; i < 1; i++ {
			mom.Update(input[i])
			check(i, false, mom.IsPrimed())
		}

		for i := 1; i < len(input); i++ {
			mom.Update(input[i])
			check(i, true, mom.IsPrimed())
		}
	})

	t.Run("length = 2", func(t *testing.T) {
		t.Parallel()
		mom := testMomentumCreate(2)

		check(-1, false, mom.IsPrimed())

		for i := 0; i < 2; i++ {
			mom.Update(input[i])
			check(i, false, mom.IsPrimed())
		}

		for i := 2; i < len(input); i++ {
			mom.Update(input[i])
			check(i, true, mom.IsPrimed())
		}
	})

	t.Run("length = 3", func(t *testing.T) {
		t.Parallel()
		mom := testMomentumCreate(3)

		check(-1, false, mom.IsPrimed())

		for i := 0; i < 3; i++ {
			mom.Update(input[i])
			check(i, false, mom.IsPrimed())
		}

		for i := 3; i < len(input); i++ {
			mom.Update(input[i])
			check(i, true, mom.IsPrimed())
		}
	})

	t.Run("length = 5", func(t *testing.T) {
		t.Parallel()
		mom := testMomentumCreate(5)

		check(-1, false, mom.IsPrimed())

		for i := 0; i < 5; i++ {
			mom.Update(input[i])
			check(i, false, mom.IsPrimed())
		}

		for i := 5; i < len(input); i++ {
			mom.Update(input[i])
			check(i, true, mom.IsPrimed())
		}
	})

	t.Run("length = 10", func(t *testing.T) {
		t.Parallel()
		mom := testMomentumCreate(10)

		check(-1, false, mom.IsPrimed())

		for i := 0; i < 10; i++ {
			mom.Update(input[i])
			check(i, false, mom.IsPrimed())
		}

		for i := 10; i < len(input); i++ {
			mom.Update(input[i])
			check(i, true, mom.IsPrimed())
		}
	})
}

func TestMomentumMetadata(t *testing.T) {
	t.Parallel()

	mom := testMomentumCreate(5)
	act := mom.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	check("Identifier", core.Momentum, act.Identifier)
	check("Mnemonic", "mom(5)", act.Mnemonic)
	check("Description", "Momentum mom(5)", act.Description)
	check("len(Outputs)", 1, len(act.Outputs))
	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", "mom(5)", act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", "Momentum mom(5)", act.Outputs[0].Description)
}

func TestNewMomentum(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		bc     entities.BarComponent   = entities.BarMedianPrice
		qc     entities.QuoteComponent = entities.QuoteMidPrice
		tc     entities.TradeComponent = entities.TradePrice
		length                         = 5
		errlen                         = "invalid momentum parameters: length should be positive"
		errbc                          = "invalid momentum parameters: 9999: unknown bar component"
		errqc                          = "invalid momentum parameters: 9999: unknown quote component"
		errtc                          = "invalid momentum parameters: 9999: unknown trade component"
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

		mom, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "mom(5, hl/2)", mom.LineIndicator.Mnemonic)
		check("description", "Momentum mom(5, hl/2)", mom.LineIndicator.Description)
		check("primed", false, mom.primed)
		check("lastIndex", length, mom.lastIndex)
		check("len(window)", length+1, len(mom.window))
		check("windowLength", length+1, mom.windowLength)
		check("windowCount", 0, mom.windowCount)
	})

	t.Run("length = 1", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: 1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mom, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "mom(1, hl/2)", mom.LineIndicator.Mnemonic)
		check("description", "Momentum mom(1, hl/2)", mom.LineIndicator.Description)
		check("primed", false, mom.primed)
		check("lastIndex", 1, mom.lastIndex)
		check("len(window)", 2, len(mom.window))
		check("windowLength", 2, mom.windowLength)
		check("windowCount", 0, mom.windowCount)
	})

	t.Run("length = 0", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: 0, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mom, err := New(&params)
		check("mom == nil", true, mom == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length < 0", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: -1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mom, err := New(&params)
		check("mom == nil", true, mom == nil)
		check("err", errlen, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		mom, err := New(&params)
		check("mom == nil", true, mom == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		mom, err := New(&params)
		check("mom == nil", true, mom == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		mom, err := New(&params)
		check("mom == nil", true, mom == nil)
		check("err", errtc, err.Error())
	})

	// Zero-value component mnemonic tests.
	// A zero value means "use default, don't show in mnemonic".

	t.Run("all components zero", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length}

		mom, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "mom(5)", mom.LineIndicator.Mnemonic)
		check("description", "Momentum mom(5)", mom.LineIndicator.Description)
	})

	t.Run("only bar component set", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length, BarComponent: entities.BarMedianPrice}

		mom, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "mom(5, hl/2)", mom.LineIndicator.Mnemonic)
		check("description", "Momentum mom(5, hl/2)", mom.LineIndicator.Description)
	})

	t.Run("only quote component set", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length, QuoteComponent: entities.QuoteBidPrice}

		mom, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "mom(5, b)", mom.LineIndicator.Mnemonic)
		check("description", "Momentum mom(5, b)", mom.LineIndicator.Description)
	})

	t.Run("only trade component set", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length, TradeComponent: entities.TradeVolume}

		mom, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "mom(5, v)", mom.LineIndicator.Mnemonic)
		check("description", "Momentum mom(5, v)", mom.LineIndicator.Description)
	})
}

func testMomentumCreate(length int) *Momentum {
	params := Params{Length: length}

	mom, _ := New(&params)

	return mom
}
