//nolint:testpackage
package centerofgravityoscillator

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func TestCenterOfGravityOscillatorUpdate(t *testing.T) { //nolint: funlen
	t.Parallel()

	check := func(index int, exp, act float64) {
		t.Helper()

		if math.Abs(exp-act) > 1e-8 {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	checkNaN := func(index int, act float64) {
		t.Helper()

		if !math.IsNaN(act) {
			t.Errorf("[%v] is incorrect: expected NaN, actual %v", index, act)
		}
	}

	input := testCenterOfGravityOscillatorInput()

	const (
		l       = 10
		lprimed = 10 // First 10 values are NaN.
	)

	t.Run("reference implementation: COG value from test_Cog.xls", func(t *testing.T) {
		t.Parallel()

		cog := testCenterOfGravityOscillatorCreate(l)
		exp := testCenterOfGravityOscillatorExpectedCog()

		for i := range lprimed {
			checkNaN(i, cog.Update(input[i]))
		}

		for i := lprimed; i < len(input); i++ {
			act := cog.Update(input[i])
			check(i, exp[i], act)
		}

		checkNaN(0, cog.Update(math.NaN()))
	})

	t.Run("reference implementation: trigger from test_Cog.xls", func(t *testing.T) {
		t.Parallel()

		cog := testCenterOfGravityOscillatorCreate(l)
		expTrig := testCenterOfGravityOscillatorExpectedTrigger()

		for i := range lprimed {
			cog.Update(input[i])
		}

		for i := lprimed; i < len(input); i++ {
			cog.Update(input[i])
			act := cog.valuePrevious
			check(i, expTrig[i], act)
		}
	})
}

func TestCenterOfGravityOscillatorUpdateEntity(t *testing.T) { //nolint: funlen,cyclop
	t.Parallel()

	const (
		l       = 10
		lprimed = 10
	)

	time := testCenterOfGravityOscillatorTime()

	input := testCenterOfGravityOscillatorInput()
	inputHigh := testCenterOfGravityOscillatorInputHigh()
	inputLow := testCenterOfGravityOscillatorInputLow()
	expCog := testCenterOfGravityOscillatorExpectedCog()
	expTrig := testCenterOfGravityOscillatorExpectedTrigger()

	check := func(index int, expValue, expTrigger float64, act core.Output) {
		t.Helper()

		const outputLen = 2

		if len(act) != outputLen {
			t.Errorf("[%v] len(output) is incorrect: expected %v, actual %v", index, outputLen, len(act))
		}

		s0, ok := act[0].(entities.Scalar)
		if !ok {
			t.Errorf("[%v] output[0] is not a scalar", index)
		}

		s1, ok := act[1].(entities.Scalar)
		if !ok {
			t.Errorf("[%v] output[1] is not a scalar", index)
		}

		if s0.Time != time {
			t.Errorf("[%v] output[0] time is incorrect: expected %v, actual %v", index, time, s0.Time)
		}

		if s1.Time != time {
			t.Errorf("[%v] output[1] time is incorrect: expected %v, actual %v", index, time, s1.Time)
		}

		if math.IsNaN(expValue) {
			if !math.IsNaN(s0.Value) {
				t.Errorf("[%v] output[0] value: expected NaN, actual %v", index, s0.Value)
			}

			if !math.IsNaN(s1.Value) {
				t.Errorf("[%v] output[1] value: expected NaN, actual %v", index, s1.Value)
			}

			return
		}

		if math.Abs(expValue-s0.Value) > 1e-8 {
			t.Errorf("[%v] output[0] value: expected %v, actual %v", index, expValue, s0.Value)
		}

		if math.Abs(expTrigger-s1.Value) > 1e-8 {
			t.Errorf("[%v] output[1] value: expected %v, actual %v", index, expTrigger, s1.Value)
		}
	}

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		cog := testCenterOfGravityOscillatorCreate(l)

		for i := 0; i < len(input); i++ {
			s := entities.Scalar{Time: time, Value: input[i]}
			act := cog.UpdateScalar(&s)
			check(i, expCog[i], expTrig[i], act)
		}
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		// Default bar component for CoG is BarMedianPrice = (High+Low)/2.
		cog := testCenterOfGravityOscillatorCreate(l)

		for i := 0; i < len(input); i++ {
			b := entities.Bar{Time: time, High: inputHigh[i], Low: inputLow[i]}
			act := cog.UpdateBar(&b)
			check(i, expCog[i], expTrig[i], act)
		}
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		// Use QuoteMidPrice = (Ask+Bid)/2, feeding high/low as ask/bid.
		cog := testCenterOfGravityOscillatorCreate(l)

		for i := range lprimed {
			q := entities.Quote{Time: time, Ask: inputHigh[i], Bid: inputLow[i]}
			act := cog.UpdateQuote(&q)
			check(i, expCog[i], expTrig[i], act)
		}

		for i := lprimed; i < len(input); i++ {
			q := entities.Quote{Time: time, Ask: inputHigh[i], Bid: inputLow[i]}
			act := cog.UpdateQuote(&q)
			check(i, expCog[i], expTrig[i], act)
		}
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		cog := testCenterOfGravityOscillatorCreate(l)

		for i := 0; i < len(input); i++ {
			r := entities.Trade{Time: time, Price: input[i]}
			act := cog.UpdateTrade(&r)
			check(i, expCog[i], expTrig[i], act)
		}
	})
}

func TestCenterOfGravityOscillatorIsPrimed(t *testing.T) {
	t.Parallel()

	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	const (
		l       = 10
		lprimed = 10
	)

	t.Run("length = 10 (default)", func(t *testing.T) {
		t.Parallel()

		cog := testCenterOfGravityOscillatorCreate(l)
		input := testCenterOfGravityOscillatorInput()

		check(0, false, cog.IsPrimed())

		for i := range lprimed {
			cog.Update(input[i])
			check(i+1, false, cog.IsPrimed())
		}

		for i := lprimed; i < len(input); i++ {
			cog.Update(input[i])
			check(i+1, true, cog.IsPrimed())
		}
	})
}

func TestCenterOfGravityOscillatorMetadata(t *testing.T) {
	t.Parallel()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	t.Run("length = 10, default components", func(t *testing.T) {
		t.Parallel()

		const l = 10

		cog := testCenterOfGravityOscillatorCreate(l)
		act := cog.Metadata()

		mn := "cog(10, hl/2)"
		mnTrig := "cogTrig(10, hl/2)"
		descr := "Center of Gravity oscillator "
		descrTrig := "Center of Gravity trigger "

		check("Identifier", core.CenterOfGravityOscillator, act.Identifier)
		check("Mnemonic", mn, act.Mnemonic)
		check("Description", descr+mn, act.Description)
		check("len(Outputs)", 2, len(act.Outputs))

		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
		check("Outputs[0].Mnemonic", mn, act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", descr+mn, act.Outputs[0].Description)

		check("Outputs[1].Kind", int(Trigger), act.Outputs[1].Kind)
		check("Outputs[1].Shape", shape.Scalar, act.Outputs[1].Shape)
		check("Outputs[1].Mnemonic", mnTrig, act.Outputs[1].Mnemonic)
		check("Outputs[1].Description", descrTrig+mnTrig, act.Outputs[1].Description)
	})

	t.Run("with non-default trade component", func(t *testing.T) {
		t.Parallel()

		params := Params{
			Length:         10,
			TradeComponent: entities.TradeVolume,
		}

		cog, _ := NewCenterOfGravityOscillator(&params)
		act := cog.Metadata()

		check("Mnemonic", "cog(10, hl/2, v)", act.Mnemonic)
		check("Description", "Center of Gravity oscillator cog(10, hl/2, v)", act.Description)
	})
}

func TestNewCenterOfGravityOscillator(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		errl  = "invalid center of gravity oscillator parameters: length should be a positive integer"
		errbc = "invalid center of gravity oscillator parameters: 9999: unknown bar component"
		errqc = "invalid center of gravity oscillator parameters: 9999: unknown quote component"
		errtc = "invalid center of gravity oscillator parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	checkInstance := func(cog *CenterOfGravityOscillator,
		mnemonic string, length int,
	) {
		mnemonicTrig := "cogTrig" + mnemonic[3:]

		check("mnemonic", mnemonic, cog.mnemonic)
		check("mnemonicTrig", mnemonicTrig, cog.mnemonicTrig)
		check("length", length, cog.length)
		check("lengthMinOne", length-1, cog.lengthMinOne)
		check("windowCount", 0, cog.windowCount)
		check("window != nil", true, cog.window != nil)
		check("len(window)", length, len(cog.window))
		check("primed", false, cog.primed)
		check("barFunc == nil", false, cog.barFunc == nil)
		check("quoteFunc == nil", false, cog.quoteFunc == nil)
		check("tradeFunc == nil", false, cog.tradeFunc == nil)
	}

	t.Run("default", func(t *testing.T) {
		t.Parallel()

		const l = 10

		params := Params{Length: l}
		cog, err := NewCenterOfGravityOscillator(&params)

		check("err == nil", true, err == nil)
		checkInstance(cog, "cog(10, hl/2)", l)
	})

	t.Run("length = 1", func(t *testing.T) {
		t.Parallel()

		const l = 1

		params := Params{Length: l}
		cog, err := NewCenterOfGravityOscillator(&params)

		check("err == nil", true, err == nil)
		checkInstance(cog, "cog(1, hl/2)", l)
	})

	t.Run("length = 0, error", func(t *testing.T) {
		t.Parallel()

		params := Params{Length: 0}
		cog, err := NewCenterOfGravityOscillator(&params)

		check("cog == nil", true, cog == nil)
		check("err", errl, err.Error())
	})

	t.Run("length < 0, error", func(t *testing.T) {
		t.Parallel()

		params := Params{Length: -8}
		cog, err := NewCenterOfGravityOscillator(&params)

		check("cog == nil", true, cog == nil)
		check("err", errl, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()

		params := Params{
			Length:       10,
			BarComponent: entities.BarComponent(9999),
		}

		cog, err := NewCenterOfGravityOscillator(&params)
		check("cog == nil", true, cog == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()

		params := Params{
			Length:         10,
			QuoteComponent: entities.QuoteComponent(9999),
		}

		cog, err := NewCenterOfGravityOscillator(&params)
		check("cog == nil", true, cog == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()

		params := Params{
			Length:         10,
			TradeComponent: entities.TradeComponent(9999),
		}

		cog, err := NewCenterOfGravityOscillator(&params)
		check("cog == nil", true, cog == nil)
		check("err", errtc, err.Error())
	})
}
