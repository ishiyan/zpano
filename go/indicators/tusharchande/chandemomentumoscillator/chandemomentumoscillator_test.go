//nolint:testpackage
package chandemomentumoscillator

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

//nolint:lll
// Input data is taken from the TA-Lib (http://ta-lib.org/) tests,
//    test_data.c, TA_SREF_close_daily_ref_0_PRIV[252].
//
// TA Lib uses incorrect CMO calculation which is incompatible with Chande's book.
// TA Lib smooths the CMO values in the same way as RSI does, but Chande didn't smooth the values.
// We don't use TA Lib test data here.

func TestChandeMomentumOscillatorUpdate(t *testing.T) { //nolint: funlen
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

	t.Run("length = 10 (book)", func(t *testing.T) {
		t.Parallel()
		input := testChandeMomentumOscillatorBookLength10Input()
		output := testChandeMomentumOscillatorBookLength10Output()
		cmo := testChandeMomentumOscillatorCreate(10)

		for i := 0; i < 10; i++ {
			checkNaN(i, cmo.Update(input[i]))
		}

		for i := 10; i < len(input); i++ {
			act := cmo.Update(input[i])
			check(i, output[i], act)
		}

		checkNaN(0, cmo.Update(math.NaN()))
	})
}

func TestChandeMomentumOscillatorUpdateEntity(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		l   = 2
		inp = 3.
		exp = 100.
	)

	time := testChandeMomentumOscillatorTime()
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
		cmo := testChandeMomentumOscillatorCreate(l)
		cmo.Update(0.)
		cmo.Update(0.)
		check(cmo.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: time, Close: inp}
		cmo := testChandeMomentumOscillatorCreate(l)
		cmo.Update(0.)
		cmo.Update(0.)
		check(cmo.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: time, Bid: inp, Ask: inp}
		cmo := testChandeMomentumOscillatorCreate(l)
		cmo.Update(0.)
		cmo.Update(0.)
		check(cmo.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: time, Price: inp}
		cmo := testChandeMomentumOscillatorCreate(l)
		cmo.Update(0.)
		cmo.Update(0.)
		check(cmo.UpdateTrade(&r))
	})
}

func TestChandeMomentumOscillatorIsPrimed(t *testing.T) { //nolint:funlen
	t.Parallel()

	input := testChandeMomentumOscillatorInput()
	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	t.Run("length = 1", func(t *testing.T) {
		t.Parallel()
		cmo := testChandeMomentumOscillatorCreate(1)

		check(-1, false, cmo.IsPrimed())

		for i := 0; i < 1; i++ {
			cmo.Update(input[i])
			check(i, false, cmo.IsPrimed())
		}

		for i := 1; i < len(input); i++ {
			cmo.Update(input[i])
			check(i, true, cmo.IsPrimed())
		}
	})

	t.Run("length = 2", func(t *testing.T) {
		t.Parallel()
		cmo := testChandeMomentumOscillatorCreate(2)

		check(-1, false, cmo.IsPrimed())

		for i := 0; i < 2; i++ {
			cmo.Update(input[i])
			check(i, false, cmo.IsPrimed())
		}

		for i := 2; i < len(input); i++ {
			cmo.Update(input[i])
			check(i, true, cmo.IsPrimed())
		}
	})

	t.Run("length = 3", func(t *testing.T) {
		t.Parallel()
		cmo := testChandeMomentumOscillatorCreate(3)

		check(-1, false, cmo.IsPrimed())

		for i := 0; i < 3; i++ {
			cmo.Update(input[i])
			check(i, false, cmo.IsPrimed())
		}

		for i := 3; i < len(input); i++ {
			cmo.Update(input[i])
			check(i, true, cmo.IsPrimed())
		}
	})

	t.Run("length = 5", func(t *testing.T) {
		t.Parallel()
		cmo := testChandeMomentumOscillatorCreate(5)

		check(-1, false, cmo.IsPrimed())

		for i := 0; i < 5; i++ {
			cmo.Update(input[i])
			check(i, false, cmo.IsPrimed())
		}

		for i := 5; i < len(input); i++ {
			cmo.Update(input[i])
			check(i, true, cmo.IsPrimed())
		}
	})

	t.Run("length = 10", func(t *testing.T) {
		t.Parallel()
		cmo := testChandeMomentumOscillatorCreate(10)

		check(-1, false, cmo.IsPrimed())

		for i := 0; i < 10; i++ {
			cmo.Update(input[i])
			check(i, false, cmo.IsPrimed())
		}

		for i := 10; i < len(input); i++ {
			cmo.Update(input[i])
			check(i, true, cmo.IsPrimed())
		}
	})
}

func TestChandeMomentumOscillatorMetadata(t *testing.T) {
	t.Parallel()

	cmo := testChandeMomentumOscillatorCreate(5)
	act := cmo.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	check("Identifier", core.ChandeMomentumOscillator, act.Identifier)
	check("Mnemonic", "cmo(5)", act.Mnemonic)
	check("Description", "Chande Momentum Oscillator cmo(5)", act.Description)
	check("len(Outputs)", 1, len(act.Outputs))
	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", "cmo(5)", act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", "Chande Momentum Oscillator cmo(5)", act.Outputs[0].Description)
}

func TestNewChandeMomentumOscillator(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		bc     entities.BarComponent   = entities.BarMedianPrice
		qc     entities.QuoteComponent = entities.QuoteMidPrice
		tc     entities.TradeComponent = entities.TradePrice
		length                         = 5
		errlen                         = "invalid Chande momentum oscillator parameters: length should be positive"
		errbc                          = "invalid Chande momentum oscillator parameters: 9999: unknown bar component"
		errqc                          = "invalid Chande momentum oscillator parameters: 9999: unknown quote component"
		errtc                          = "invalid Chande momentum oscillator parameters: 9999: unknown trade component"
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

		cmo, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "cmo(5, hl/2)", cmo.LineIndicator.Mnemonic)
		check("description", "Chande Momentum Oscillator cmo(5, hl/2)", cmo.LineIndicator.Description)
		check("primed", false, cmo.primed)
		check("length", length, cmo.length)
		check("len(ringBuffer)", length, len(cmo.ringBuffer))
		check("ringHead", 0, cmo.ringHead)
		check("count", 0, cmo.count)
		check("previousSample", 0., cmo.previousSample)
		check("gainSum", 0., cmo.gainSum)
		check("lossSum", 0., cmo.lossSum)
	})

	t.Run("length = 1", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: 1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		cmo, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "cmo(1, hl/2)", cmo.LineIndicator.Mnemonic)
		check("description", "Chande Momentum Oscillator cmo(1, hl/2)", cmo.LineIndicator.Description)
		check("primed", false, cmo.primed)
		check("length", 1, cmo.length)
		check("len(ringBuffer)", 1, len(cmo.ringBuffer))
		check("ringHead", 0, cmo.ringHead)
		check("count", 0, cmo.count)
		check("previousSample", 0., cmo.previousSample)
		check("gainSum", 0., cmo.gainSum)
		check("lossSum", 0., cmo.lossSum)
	})

	t.Run("length = 0", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: 0, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		cmo, err := New(&params)
		check("cmo == nil", true, cmo == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length < 0", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: -1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		cmo, err := New(&params)
		check("cmo == nil", true, cmo == nil)
		check("err", errlen, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		cmo, err := New(&params)
		check("cmo == nil", true, cmo == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		cmo, err := New(&params)
		check("cmo == nil", true, cmo == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		cmo, err := New(&params)
		check("cmo == nil", true, cmo == nil)
		check("err", errtc, err.Error())
	})

	// Zero-value component mnemonic tests.
	// A zero value means "use default, don't show in mnemonic".

	t.Run("all components zero", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length}

		cmo, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "cmo(5)", cmo.LineIndicator.Mnemonic)
		check("description", "Chande Momentum Oscillator cmo(5)", cmo.LineIndicator.Description)
	})

	t.Run("only bar component set", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length, BarComponent: entities.BarMedianPrice}

		cmo, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "cmo(5, hl/2)", cmo.LineIndicator.Mnemonic)
		check("description", "Chande Momentum Oscillator cmo(5, hl/2)", cmo.LineIndicator.Description)
	})

	t.Run("only quote component set", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length, QuoteComponent: entities.QuoteBidPrice}

		cmo, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "cmo(5, b)", cmo.LineIndicator.Mnemonic)
		check("description", "Chande Momentum Oscillator cmo(5, b)", cmo.LineIndicator.Description)
	})

	t.Run("only trade component set", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length, TradeComponent: entities.TradeVolume}

		cmo, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "cmo(5, v)", cmo.LineIndicator.Mnemonic)
		check("description", "Chande Momentum Oscillator cmo(5, v)", cmo.LineIndicator.Description)
	})
}
