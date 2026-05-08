//nolint:testpackage
package cybercycle

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func TestCyberCycleUpdate(t *testing.T) { //nolint: funlen
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

	input := testCyberCycleInput()

	const lprimed = 7 // First 7 values are NaN (primed on sample 8).

	t.Run("reference implementation: cycle value from test_iTrend.xls", func(t *testing.T) {
		t.Parallel()

		cc := testCyberCycleCreateDefault()
		exp := testCyberCycleExpectedCycle()

		for i := range lprimed {
			checkNaN(i, cc.Update(input[i]))
		}

		for i := lprimed; i < len(input); i++ {
			act := cc.Update(input[i])
			check(i, exp[i], act)
		}

		checkNaN(0, cc.Update(math.NaN()))
	})

	t.Run("reference implementation: signal from test_iTrend.xls", func(t *testing.T) {
		t.Parallel()

		cc := testCyberCycleCreateDefault()
		expSignal := testCyberCycleExpectedSignal()

		for i := range lprimed {
			cc.Update(input[i])
		}

		for i := lprimed; i < len(input); i++ {
			cc.Update(input[i])
			act := cc.signal
			check(i, expSignal[i], act)
		}
	})
}

func TestCyberCycleUpdateEntity(t *testing.T) { //nolint: funlen,cyclop
	t.Parallel()

	const lprimed = 7

	time := testCyberCycleTime()

	input := testCyberCycleInput()
	inputHigh := testCyberCycleInputHigh()
	inputLow := testCyberCycleInputLow()
	expCycle := testCyberCycleExpectedCycle()
	expSignal := testCyberCycleExpectedSignal()

	check := func(index int, expValue, expSignal float64, act core.Output) {
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

		if math.Abs(expSignal-s1.Value) > 1e-8 {
			t.Errorf("[%v] output[1] value: expected %v, actual %v", index, expSignal, s1.Value)
		}
	}

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		cc := testCyberCycleCreateDefault()

		for i := 0; i < len(input); i++ {
			s := entities.Scalar{Time: time, Value: input[i]}
			act := cc.UpdateScalar(&s)
			check(i, expCycle[i], expSignal[i], act)
		}
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		// Default bar component for CyberCycle is BarMedianPrice = (High+Low)/2.
		cc := testCyberCycleCreateDefault()

		for i := 0; i < len(input); i++ {
			b := entities.Bar{Time: time, High: inputHigh[i], Low: inputLow[i]}
			act := cc.UpdateBar(&b)
			check(i, expCycle[i], expSignal[i], act)
		}
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		// Use QuoteMidPrice = (Ask+Bid)/2, feeding high/low as ask/bid.
		cc := testCyberCycleCreateDefault()

		for i := range lprimed {
			q := entities.Quote{Time: time, Ask: inputHigh[i], Bid: inputLow[i]}
			act := cc.UpdateQuote(&q)
			check(i, expCycle[i], expSignal[i], act)
		}

		for i := lprimed; i < len(input); i++ {
			q := entities.Quote{Time: time, Ask: inputHigh[i], Bid: inputLow[i]}
			act := cc.UpdateQuote(&q)
			check(i, expCycle[i], expSignal[i], act)
		}
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		cc := testCyberCycleCreateDefault()

		for i := 0; i < len(input); i++ {
			r := entities.Trade{Time: time, Price: input[i]}
			act := cc.UpdateTrade(&r)
			check(i, expCycle[i], expSignal[i], act)
		}
	})
}

func TestCyberCycleIsPrimed(t *testing.T) {
	t.Parallel()

	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	const lprimed = 7

	t.Run("default params", func(t *testing.T) {
		t.Parallel()

		cc := testCyberCycleCreateDefault()
		input := testCyberCycleInput()

		check(0, false, cc.IsPrimed())

		for i := range lprimed {
			cc.Update(input[i])
			check(i+1, false, cc.IsPrimed())
		}

		for i := lprimed; i < len(input); i++ {
			cc.Update(input[i])
			check(i+1, true, cc.IsPrimed())
		}
	})
}

func TestCyberCycleMetadata(t *testing.T) {
	t.Parallel()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	t.Run("default params (smoothing factor 0.07)", func(t *testing.T) {
		t.Parallel()

		cc := testCyberCycleCreateDefault()
		act := cc.Metadata()

		mn := "cc(28, hl/2)"
		mnSignal := "ccSignal(28, hl/2)"
		descr := "Cyber Cycle "
		descrSignal := "Cyber Cycle signal "

		check("Identifier", core.CyberCycle, act.Identifier)
		check("Mnemonic", mn, act.Mnemonic)
		check("Description", descr+mn, act.Description)
		check("len(Outputs)", 2, len(act.Outputs))

		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
		check("Outputs[0].Mnemonic", mn, act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", descr+mn, act.Outputs[0].Description)

		check("Outputs[1].Kind", int(Signal), act.Outputs[1].Kind)
		check("Outputs[1].Shape", shape.Scalar, act.Outputs[1].Shape)
		check("Outputs[1].Mnemonic", mnSignal, act.Outputs[1].Mnemonic)
		check("Outputs[1].Description", descrSignal+mnSignal, act.Outputs[1].Description)
	})

	t.Run("length-based with non-default trade component", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{
			Length:         3,
			SignalLag:      2,
			TradeComponent: entities.TradeVolume,
		}

		cc, _ := NewCyberCycleLength(&params)
		act := cc.Metadata()

		check("Mnemonic", "cc(3, hl/2, v)", act.Mnemonic)
		check("Description", "Cyber Cycle cc(3, hl/2, v)", act.Description)
	})
}

func TestNewCyberCycleLength(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		errLength    = "invalid cyber cycle parameters: length should be a positive integer"
		errSignalLag = "invalid cyber cycle parameters: signal lag should be a positive integer"
		errbc        = "invalid cyber cycle parameters: 9999: unknown bar component"
		errqc        = "invalid cyber cycle parameters: 9999: unknown quote component"
		errtc        = "invalid cyber cycle parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	checkInstance := func(cc *CyberCycle,
		length int, signalLag int,
	) {
		t.Helper()

		check("length", length, cc.length)
		check("signalLag", signalLag, cc.signalLag)
		check("primed", false, cc.primed)
		check("value is NaN", true, math.IsNaN(cc.value))
		check("signal is NaN", true, math.IsNaN(cc.signal))
		check("barFunc == nil", false, cc.barFunc == nil)
		check("quoteFunc == nil", false, cc.quoteFunc == nil)
		check("tradeFunc == nil", false, cc.tradeFunc == nil)
	}

	t.Run("length=28, signalLag=14", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{Length: 28, SignalLag: 14}
		cc, err := NewCyberCycleLength(&params)

		check("err == nil", true, err == nil)
		checkInstance(cc, 28, 14)
	})

	t.Run("length=1, signalLag=1", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{Length: 1, SignalLag: 1}
		cc, err := NewCyberCycleLength(&params)

		check("err == nil", true, err == nil)
		checkInstance(cc, 1, 1)
	})

	t.Run("length=0, error", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{Length: 0, SignalLag: 1}
		cc, err := NewCyberCycleLength(&params)

		check("cc == nil", true, cc == nil)
		check("err", errLength, err.Error())
	})

	t.Run("length=-8, error", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{Length: -8, SignalLag: 1}
		cc, err := NewCyberCycleLength(&params)

		check("cc == nil", true, cc == nil)
		check("err", errLength, err.Error())
	})

	t.Run("signalLag=0, error", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{Length: 1, SignalLag: 0}
		cc, err := NewCyberCycleLength(&params)

		check("cc == nil", true, cc == nil)
		check("err", errSignalLag, err.Error())
	})

	t.Run("signalLag=-8, error", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{Length: 1, SignalLag: -8}
		cc, err := NewCyberCycleLength(&params)

		check("cc == nil", true, cc == nil)
		check("err", errSignalLag, err.Error())
	})

	t.Run("length=-8, signalLag=-9, error", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{Length: -8, SignalLag: -9}
		cc, err := NewCyberCycleLength(&params)

		check("cc == nil", true, cc == nil)
		check("err", errLength, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{
			Length:       10,
			SignalLag:    9,
			BarComponent: entities.BarComponent(9999),
		}

		cc, err := NewCyberCycleLength(&params)
		check("cc == nil", true, cc == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{
			Length:         10,
			SignalLag:      9,
			QuoteComponent: entities.QuoteComponent(9999),
		}

		cc, err := NewCyberCycleLength(&params)
		check("cc == nil", true, cc == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{
			Length:         10,
			SignalLag:      9,
			TradeComponent: entities.TradeComponent(9999),
		}

		cc, err := NewCyberCycleLength(&params)
		check("cc == nil", true, cc == nil)
		check("err", errtc, err.Error())
	})
}

func TestNewCyberCycleSmoothingFactor(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		errAlpha     = "invalid cyber cycle parameters: smoothing factor should be in range [0, 1]"
		errSignalLag = "invalid cyber cycle parameters: signal lag should be a positive integer"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("default (alpha=0.07, signalLag=9)", func(t *testing.T) {
		t.Parallel()

		params := SmoothingFactorParams{SmoothingFactor: 0.07, SignalLag: 9}
		cc, err := NewCyberCycleSmoothingFactor(&params)

		check("err == nil", true, err == nil)
		check("smoothingFactor", 0.07, cc.smoothingFactor)
		check("length", 28, cc.length)
		check("signalLag", 9, cc.signalLag)
	})

	t.Run("alpha=0.06, signalLag=11", func(t *testing.T) {
		t.Parallel()

		params := SmoothingFactorParams{SmoothingFactor: 0.06, SignalLag: 11}
		cc, err := NewCyberCycleSmoothingFactor(&params)

		check("err == nil", true, err == nil)
		check("smoothingFactor", 0.06, cc.smoothingFactor)
		check("length", 32, cc.length)
		check("signalLag", 11, cc.signalLag)
	})

	t.Run("near-zero alpha (epsilon case)", func(t *testing.T) {
		t.Parallel()

		params := SmoothingFactorParams{SmoothingFactor: 0.000000001, SignalLag: 9}
		cc, err := NewCyberCycleSmoothingFactor(&params)

		check("err == nil", true, err == nil)
		check("smoothingFactor", 0.000000001, cc.smoothingFactor)
		check("length", math.MaxInt, cc.length)
	})

	t.Run("alpha=0 (boundary)", func(t *testing.T) {
		t.Parallel()

		params := SmoothingFactorParams{SmoothingFactor: 0, SignalLag: 9}
		cc, err := NewCyberCycleSmoothingFactor(&params)

		check("err == nil", true, err == nil)
		check("smoothingFactor", 0.0, cc.smoothingFactor)
		check("length", math.MaxInt, cc.length)
	})

	t.Run("alpha=1 (boundary)", func(t *testing.T) {
		t.Parallel()

		params := SmoothingFactorParams{SmoothingFactor: 1, SignalLag: 9}
		cc, err := NewCyberCycleSmoothingFactor(&params)

		check("err == nil", true, err == nil)
		check("smoothingFactor", 1.0, cc.smoothingFactor)
		check("length", 1, cc.length)
	})

	t.Run("alpha=-0.0001, error", func(t *testing.T) {
		t.Parallel()

		params := SmoothingFactorParams{SmoothingFactor: -0.0001, SignalLag: 8}
		cc, err := NewCyberCycleSmoothingFactor(&params)

		check("cc == nil", true, cc == nil)
		check("err", errAlpha, err.Error())
	})

	t.Run("alpha=1.0001, error", func(t *testing.T) {
		t.Parallel()

		params := SmoothingFactorParams{SmoothingFactor: 1.0001, SignalLag: 8}
		cc, err := NewCyberCycleSmoothingFactor(&params)

		check("cc == nil", true, cc == nil)
		check("err", errAlpha, err.Error())
	})

	t.Run("signalLag=0, error", func(t *testing.T) {
		t.Parallel()

		params := SmoothingFactorParams{SmoothingFactor: 0.07, SignalLag: 0}
		cc, err := NewCyberCycleSmoothingFactor(&params)

		check("cc == nil", true, cc == nil)
		check("err", errSignalLag, err.Error())
	})

	t.Run("signalLag=-8, error", func(t *testing.T) {
		t.Parallel()

		params := SmoothingFactorParams{SmoothingFactor: 0.07, SignalLag: -8}
		cc, err := NewCyberCycleSmoothingFactor(&params)

		check("cc == nil", true, cc == nil)
		check("err", errSignalLag, err.Error())
	})
}
