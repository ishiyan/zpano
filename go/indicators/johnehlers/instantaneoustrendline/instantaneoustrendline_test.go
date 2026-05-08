//nolint:testpackage
package instantaneoustrendline

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func TestITLUpdate(t *testing.T) { //nolint: funlen
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

	input := testITLInput()

	const lprimed = 4 // First 4 values are NaN (primed on sample 5).

	t.Run("reference implementation: trend line from test_iTrend.xls", func(t *testing.T) {
		t.Parallel()

		itl := testITLCreateDefault()
		exp := testITLExpectedTrendLine()

		for i := range lprimed {
			checkNaN(i, itl.Update(input[i]))
		}

		for i := lprimed; i < len(input); i++ {
			act := itl.Update(input[i])
			check(i, exp[i], act)
		}

		checkNaN(0, itl.Update(math.NaN()))
	})

	t.Run("reference implementation: trigger line from test_iTrend.xls", func(t *testing.T) {
		t.Parallel()

		itl := testITLCreateDefault()
		expTrig := testITLExpectedTrigger()

		for i := range lprimed {
			itl.Update(input[i])
		}

		for i := lprimed; i < len(input); i++ {
			itl.Update(input[i])
			act := itl.triggerLine
			check(i, expTrig[i], act)
		}
	})
}

func TestITLUpdateEntity(t *testing.T) { //nolint: funlen,cyclop
	t.Parallel()

	const lprimed = 4

	time := testITLTime()

	input := testITLInput()
	inputHigh := testITLInputHigh()
	inputLow := testITLInputLow()
	expTrend := testITLExpectedTrendLine()
	expTrigger := testITLExpectedTrigger()

	check := func(index int, expValue, expTrig float64, act core.Output) {
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

		if math.Abs(expTrig-s1.Value) > 1e-8 {
			t.Errorf("[%v] output[1] value: expected %v, actual %v", index, expTrig, s1.Value)
		}
	}

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		itl := testITLCreateDefault()

		for i := 0; i < len(input); i++ {
			s := entities.Scalar{Time: time, Value: input[i]}
			act := itl.UpdateScalar(&s)
			check(i, expTrend[i], expTrigger[i], act)
		}
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		// Default bar component for ITL is BarMedianPrice = (High+Low)/2.
		itl := testITLCreateDefault()

		for i := 0; i < len(input); i++ {
			b := entities.Bar{Time: time, High: inputHigh[i], Low: inputLow[i]}
			act := itl.UpdateBar(&b)
			check(i, expTrend[i], expTrigger[i], act)
		}
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		// Use QuoteMidPrice = (Ask+Bid)/2, feeding high/low as ask/bid.
		itl := testITLCreateDefault()

		for i := range lprimed {
			q := entities.Quote{Time: time, Ask: inputHigh[i], Bid: inputLow[i]}
			act := itl.UpdateQuote(&q)
			check(i, expTrend[i], expTrigger[i], act)
		}

		for i := lprimed; i < len(input); i++ {
			q := entities.Quote{Time: time, Ask: inputHigh[i], Bid: inputLow[i]}
			act := itl.UpdateQuote(&q)
			check(i, expTrend[i], expTrigger[i], act)
		}
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		itl := testITLCreateDefault()

		for i := 0; i < len(input); i++ {
			r := entities.Trade{Time: time, Price: input[i]}
			act := itl.UpdateTrade(&r)
			check(i, expTrend[i], expTrigger[i], act)
		}
	})
}

func TestITLIsPrimed(t *testing.T) {
	t.Parallel()

	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	const lprimed = 4

	t.Run("default params", func(t *testing.T) {
		t.Parallel()

		itl := testITLCreateDefault()
		input := testITLInput()

		check(0, false, itl.IsPrimed())

		for i := range lprimed {
			itl.Update(input[i])
			check(i+1, false, itl.IsPrimed())
		}

		for i := lprimed; i < len(input); i++ {
			itl.Update(input[i])
			check(i+1, true, itl.IsPrimed())
		}
	})
}

func TestITLMetadata(t *testing.T) {
	t.Parallel()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	t.Run("default params (smoothing factor 0.07)", func(t *testing.T) {
		t.Parallel()

		itl := testITLCreateDefault()
		act := itl.Metadata()

		mn := "iTrend(28, hl/2)"
		mnTrig := "iTrendTrigger(28, hl/2)"
		descr := "Instantaneous Trend Line "
		descrTr := "Instantaneous Trend Line trigger "

		check("Identifier", core.InstantaneousTrendLine, act.Identifier)
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
		check("Outputs[1].Description", descrTr+mnTrig, act.Outputs[1].Description)
	})

	t.Run("length-based with non-default trade component", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{
			Length:         3,
			TradeComponent: entities.TradeVolume,
		}

		itl, _ := NewInstantaneousTrendLineLength(&params)
		act := itl.Metadata()

		check("Mnemonic", "iTrend(3, hl/2, v)", act.Mnemonic)
		check("Description", "Instantaneous Trend Line iTrend(3, hl/2, v)", act.Description)
	})
}

func TestNewITLLength(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		errLength = "invalid instantaneous trend line parameters: length should be a positive integer"
		errbc     = "invalid instantaneous trend line parameters: 9999: unknown bar component"
		errqc     = "invalid instantaneous trend line parameters: 9999: unknown quote component"
		errtc     = "invalid instantaneous trend line parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	checkInstance := func(itl *InstantaneousTrendLine, length int) {
		t.Helper()

		check("length", length, itl.length)
		check("primed", false, itl.primed)
		check("trendLine is NaN", true, math.IsNaN(itl.trendLine))
		check("triggerLine is NaN", true, math.IsNaN(itl.triggerLine))
		check("barFunc == nil", false, itl.barFunc == nil)
		check("quoteFunc == nil", false, itl.quoteFunc == nil)
		check("tradeFunc == nil", false, itl.tradeFunc == nil)
	}

	t.Run("length=28", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{Length: 28}
		itl, err := NewInstantaneousTrendLineLength(&params)

		check("err == nil", true, err == nil)
		checkInstance(itl, 28)
	})

	t.Run("length=1", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{Length: 1}
		itl, err := NewInstantaneousTrendLineLength(&params)

		check("err == nil", true, err == nil)
		checkInstance(itl, 1)
	})

	t.Run("length=0, error", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{Length: 0}
		itl, err := NewInstantaneousTrendLineLength(&params)

		check("itl == nil", true, itl == nil)
		check("err", errLength, err.Error())
	})

	t.Run("length=-8, error", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{Length: -8}
		itl, err := NewInstantaneousTrendLineLength(&params)

		check("itl == nil", true, itl == nil)
		check("err", errLength, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{
			Length:       10,
			BarComponent: entities.BarComponent(9999),
		}

		itl, err := NewInstantaneousTrendLineLength(&params)
		check("itl == nil", true, itl == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{
			Length:         10,
			QuoteComponent: entities.QuoteComponent(9999),
		}

		itl, err := NewInstantaneousTrendLineLength(&params)
		check("itl == nil", true, itl == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{
			Length:         10,
			TradeComponent: entities.TradeComponent(9999),
		}

		itl, err := NewInstantaneousTrendLineLength(&params)
		check("itl == nil", true, itl == nil)
		check("err", errtc, err.Error())
	})
}

func TestNewITLSmoothingFactor(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		errAlpha = "invalid instantaneous trend line parameters: smoothing factor should be in range [0, 1]"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("default (alpha=0.07)", func(t *testing.T) {
		t.Parallel()

		params := SmoothingFactorParams{SmoothingFactor: 0.07}
		itl, err := NewInstantaneousTrendLineSmoothingFactor(&params)

		check("err == nil", true, err == nil)
		check("smoothingFactor", 0.07, itl.smoothingFactor)
		check("length", 28, itl.length)
	})

	t.Run("alpha=0.06", func(t *testing.T) {
		t.Parallel()

		params := SmoothingFactorParams{SmoothingFactor: 0.06}
		itl, err := NewInstantaneousTrendLineSmoothingFactor(&params)

		check("err == nil", true, err == nil)
		check("smoothingFactor", 0.06, itl.smoothingFactor)
		check("length", 32, itl.length)
	})

	t.Run("near-zero alpha (epsilon case)", func(t *testing.T) {
		t.Parallel()

		params := SmoothingFactorParams{SmoothingFactor: 0.000000001}
		itl, err := NewInstantaneousTrendLineSmoothingFactor(&params)

		check("err == nil", true, err == nil)
		check("smoothingFactor", 0.000000001, itl.smoothingFactor)
		check("length", math.MaxInt, itl.length)
	})

	t.Run("alpha=0 (boundary)", func(t *testing.T) {
		t.Parallel()

		params := SmoothingFactorParams{SmoothingFactor: 0}
		itl, err := NewInstantaneousTrendLineSmoothingFactor(&params)

		check("err == nil", true, err == nil)
		check("smoothingFactor", 0.0, itl.smoothingFactor)
		check("length", math.MaxInt, itl.length)
	})

	t.Run("alpha=1 (boundary)", func(t *testing.T) {
		t.Parallel()

		params := SmoothingFactorParams{SmoothingFactor: 1}
		itl, err := NewInstantaneousTrendLineSmoothingFactor(&params)

		check("err == nil", true, err == nil)
		check("smoothingFactor", 1.0, itl.smoothingFactor)
		check("length", 1, itl.length)
	})

	t.Run("alpha=-0.0001, error", func(t *testing.T) {
		t.Parallel()

		params := SmoothingFactorParams{SmoothingFactor: -0.0001}
		itl, err := NewInstantaneousTrendLineSmoothingFactor(&params)

		check("itl == nil", true, itl == nil)
		check("err", errAlpha, err.Error())
	})

	t.Run("alpha=1.0001, error", func(t *testing.T) {
		t.Parallel()

		params := SmoothingFactorParams{SmoothingFactor: 1.0001}
		itl, err := NewInstantaneousTrendLineSmoothingFactor(&params)

		check("itl == nil", true, itl == nil)
		check("err", errAlpha, err.Error())
	})
}
