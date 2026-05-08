//nolint:testpackage
package fractaladaptivemovingaverage

//nolint: gofumpt
import (
	"math"
	"strings"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func TestFractalAdaptiveMovingAverageUpdate(t *testing.T) { //nolint: funlen
	t.Parallel()

	check := func(index int, exp, act float64) {
		t.Helper()

		if math.Abs(exp-act) > 1e-9 {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	checkNaN := func(index int, act float64) {
		t.Helper()

		if !math.IsNaN(act) {
			t.Errorf("[%v] is incorrect: expected NaN, actual %v", index, act)
		}
	}

	inputHigh := testFractalAdaptiveMovingAverageInputHigh()
	inputLow := testFractalAdaptiveMovingAverageInputLow()
	inputMid := testFractalAdaptiveMovingAverageInputMid()

	const (
		lprimed = 15
		l       = 16
		a       = 0.01
	)

	t.Run("reference implementation: FRAMA from test_frama.xls", func(t *testing.T) {
		t.Parallel()

		frama := testFractalAdaptiveMovingAverageCreate(l, a)
		exp := testFractalAdaptiveMovingAverageExpected()

		for i := range lprimed {
			checkNaN(i, frama.Update(inputMid[i], inputHigh[i], inputLow[i]))
		}

		for i := lprimed; i < len(inputMid); i++ {
			act := frama.Update(inputMid[i], inputHigh[i], inputLow[i])
			check(i, exp[i], act)
		}

		checkNaN(0, frama.Update(math.NaN(), math.NaN(), math.NaN()))
	})

	t.Run("reference implementation: Fdim from test_frama.xls", func(t *testing.T) {
		t.Parallel()

		frama := testFractalAdaptiveMovingAverageCreate(l, a)
		exp := testFractalAdaptiveMovingAverageExpectedFdim()

		for i := range lprimed {
			frama.Update(inputMid[i], inputHigh[i], inputLow[i])
			checkNaN(i, frama.fractalDimension)
		}

		for i := lprimed; i < len(inputMid); i++ {
			frama.Update(inputMid[i], inputHigh[i], inputLow[i])
			act := frama.fractalDimension
			check(i, exp[i], act)
		}
	})
}

func TestFractalAdaptiveMovingAverageUpdateEntity(t *testing.T) { //nolint: funlen,cyclop
	t.Parallel()

	const (
		lprimed       = 15
		l             = 16
		a             = 0.01
		inp           = 3.
		expectedFrama = 2.999999999999997
		expectedFdim  = 1.0000000000000002
	)

	time := testFractalAdaptiveMovingAverageTime()
	check := func(expFrama, expFdim float64, act core.Output) {
		t.Helper()

		const outputLen = 2

		if len(act) != outputLen {
			t.Errorf("len(output) is incorrect: expected %v, actual %v", outputLen, len(act))
		}

		i := 0

		s0, ok := act[i].(entities.Scalar)
		if !ok {
			t.Error("output[0] is not a scalar")
		}

		i++

		s1, ok := act[i].(entities.Scalar)
		if !ok {
			t.Error("output[1] is not a scalar")
		}

		if s0.Time != time {
			t.Errorf("output[0] time is incorrect: expected %v, actual %v", time, s0.Time)
		}

		if s0.Value != expFrama {
			t.Errorf("output[0] value is incorrect: expected %v, actual %v", expFrama, s0.Value)
		}

		if s1.Time != time {
			t.Errorf("output[1] time is incorrect: expected %v, actual %v", time, s1.Time)
		}

		if s1.Value != expFdim {
			t.Errorf("output[1] value is incorrect: expected %v, actual %v", expFdim, s1.Value)
		}
	}

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		s := entities.Scalar{Time: time, Value: inp}
		frama := testFractalAdaptiveMovingAverageCreate(l, a)

		for range lprimed {
			frama.Update(0., 0., 0.)
		}

		check(expectedFrama, expectedFdim, frama.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: time, Close: inp, High: inp, Low: inp}
		frama := testFractalAdaptiveMovingAverageCreate(l, a)

		for range lprimed {
			frama.Update(0., 0., 0.)
		}

		check(expectedFrama, expectedFdim, frama.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: time, Bid: inp, Ask: inp}
		frama := testFractalAdaptiveMovingAverageCreate(l, a)

		for range lprimed {
			frama.Update(0., 0., 0.)
		}

		check(expectedFrama, expectedFdim, frama.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: time, Price: inp}
		frama := testFractalAdaptiveMovingAverageCreate(l, a)

		for range lprimed {
			frama.Update(0., 0., 0.)
		}

		check(expectedFrama, expectedFdim, frama.UpdateTrade(&r))
	})
}

func TestFractalAdaptiveMovingAverageIsPrimed(t *testing.T) {
	t.Parallel()

	inputHigh := testFractalAdaptiveMovingAverageInputHigh()
	inputLow := testFractalAdaptiveMovingAverageInputLow()
	inputMid := testFractalAdaptiveMovingAverageInputMid()

	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	const (
		lprimed = 15
		l       = 16
		a       = 0.01
	)

	t.Run("length = 16, slow alpha = 0.01 (test_frama.xls)", func(t *testing.T) {
		t.Parallel()

		frama := testFractalAdaptiveMovingAverageCreate(l, a)

		check(0, false, frama.IsPrimed())

		for i := range lprimed {
			frama.Update(inputMid[i], inputHigh[i], inputLow[i])
			check(i+1, false, frama.IsPrimed())
		}

		for i := lprimed; i < len(inputMid); i++ {
			frama.Update(inputMid[i], inputHigh[i], inputLow[i])
			check(i+1, true, frama.IsPrimed())
		}
	})
}

func TestFractalAdaptiveMovingAverageMetadata(t *testing.T) {
	t.Parallel()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	t.Run("length = 16, alpha = 0.01", func(t *testing.T) {
		t.Parallel()

		const (
			l = 16
			a = 0.01
		)

		frama := testFractalAdaptiveMovingAverageCreate(l, a)
		act := frama.Metadata()

		mn := "frama(16, 0.010)"
		mnFdim := "framaDim(16, 0.010)"
		descr := "Fractal adaptive moving average "

		check("Identifier", core.FractalAdaptiveMovingAverage, act.Identifier)
		check("Mnemonic", mn, act.Mnemonic)
		check("Description", descr+mn, act.Description)
		check("len(Outputs)", 2, len(act.Outputs))

		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
		check("Outputs[0].Mnemonic", mn, act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", descr+mn, act.Outputs[0].Description)

		check("Outputs[1].Kind", int(Fdim), act.Outputs[1].Kind)
		check("Outputs[1].Shape", shape.Scalar, act.Outputs[1].Shape)
		check("Outputs[1].Mnemonic", mnFdim, act.Outputs[1].Mnemonic)
		check("Outputs[1].Description", descr+mnFdim, act.Outputs[1].Description)
	})

	t.Run("with non-default bar component", func(t *testing.T) {
		t.Parallel()

		params := Params{
			Length: 16, SlowestSmoothingFactor: 0.01,
			BarComponent: entities.BarMedianPrice,
		}

		frama, _ := NewFractalAdaptiveMovingAverage(&params)
		act := frama.Metadata()

		check("Mnemonic", "frama(16, 0.010, hl/2)", act.Mnemonic)
		check("Description", "Fractal adaptive moving average frama(16, 0.010, hl/2)", act.Description)
	})
}

func TestNewFractalAdaptiveMovingAverage(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		bc entities.BarComponent   = entities.BarMedianPrice
		qc entities.QuoteComponent = entities.QuoteMidPrice
		tc entities.TradeComponent = entities.TradePrice

		errl  = "invalid fractal adaptive moving average parameters: length should be an even integer larger than 1"
		erra  = "invalid fractal adaptive moving average parameters: slowest smoothing factor should be in range [0, 1]"
		errbc = "invalid fractal adaptive moving average parameters: 9999: unknown bar component"
		errqc = "invalid fractal adaptive moving average parameters: 9999: unknown quote component"
		errtc = "invalid fractal adaptive moving average parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	checkInstance := func(frama *FractalAdaptiveMovingAverage,
		mnemonic string, length int, aSlow float64,
	) {
		const (
			descr = "Fractal adaptive moving average "
			two   = 2
		)

		mnemonicFdim := strings.ReplaceAll(mnemonic, "frama", "framaDim")

		check("mnemonic", mnemonic, frama.mnemonic)
		check("description", descr+mnemonic, frama.description)
		check("mnemonicFdim", mnemonicFdim, frama.mnemonicFdim)
		check("descriptionFdim", descr+mnemonicFdim, frama.descriptionFdim)
		check("length", length, frama.length)
		check("lengthMinOne", length-1, frama.lengthMinOne)
		check("halfLength", length/two, frama.halfLength)
		check("alphaSlowest", aSlow, frama.alphaSlowest)
		check("scalingFactor", math.Log(aSlow), frama.scalingFactor)
		check("windowCount", 0, frama.windowCount)
		check("windowHigh != nil", true, frama.windowHigh != nil)
		check("windowLow != nil", true, frama.windowLow != nil)
		check("len(windowHigh)", length, len(frama.windowHigh))
		check("len(windowLow)", length, len(frama.windowLow))
		check("primed", false, frama.primed)
		check("barFunc == nil", false, frama.barFunc == nil)
		check("quoteFunc == nil", false, frama.quoteFunc == nil)
		check("tradeFunc == nil", false, frama.tradeFunc == nil)
	}

	t.Run("default", func(t *testing.T) {
		t.Parallel()

		const (
			l = 16
			a = 0.01
		)

		params := Params{
			Length: l, SlowestSmoothingFactor: a,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		frama, err := NewFractalAdaptiveMovingAverage(&params)
		check("err == nil", true, err == nil)
		checkInstance(frama, "frama(16, 0.010, hl/2)", l, a)
	})

	t.Run("non-default lengths and alpha", func(t *testing.T) {
		t.Parallel()

		const (
			l = 18
			a = 0.05
		)

		params := Params{
			Length: l, SlowestSmoothingFactor: a,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		frama, err := NewFractalAdaptiveMovingAverage(&params)
		check("err == nil", true, err == nil)
		checkInstance(frama, "frama(18, 0.050, hl/2)", l, a)
	})

	t.Run("odd lengths", func(t *testing.T) {
		t.Parallel()

		const (
			l = 17
			a = 0.01
		)

		params := Params{
			Length: l, SlowestSmoothingFactor: a,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		frama, err := NewFractalAdaptiveMovingAverage(&params)
		check("err == nil", true, err == nil)
		checkInstance(frama, "frama(18, 0.010, hl/2)", l+1, a)
	})

	t.Run("length = 1, error", func(t *testing.T) {
		t.Parallel()

		const (
			l = 1
			a = 0.01
		)

		params := Params{
			Length: l, SlowestSmoothingFactor: a,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		frama, err := NewFractalAdaptiveMovingAverage(&params)
		check("frama == nil", true, frama == nil)
		check("err", errl, err.Error())
	})

	t.Run("length = 0, error", func(t *testing.T) {
		t.Parallel()

		const (
			l = 0
			a = 0.01
		)

		params := Params{
			Length: l, SlowestSmoothingFactor: a,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		frama, err := NewFractalAdaptiveMovingAverage(&params)
		check("frama == nil", true, frama == nil)
		check("err", errl, err.Error())
	})

	t.Run("length < 0, error", func(t *testing.T) {
		t.Parallel()

		const (
			l = -1
			a = 0.01
		)

		params := Params{
			Length: l, SlowestSmoothingFactor: a,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		frama, err := NewFractalAdaptiveMovingAverage(&params)
		check("frama == nil", true, frama == nil)
		check("err", errl, err.Error())
	})

	t.Run("αs < 0, error", func(t *testing.T) {
		t.Parallel()

		const (
			l = 16
			a = -0.01
		)

		params := Params{
			Length: l, SlowestSmoothingFactor: a,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		frama, err := NewFractalAdaptiveMovingAverage(&params)
		check("frama == nil", true, frama == nil)
		check("err", erra, err.Error())
	})

	t.Run("αs > 1, error", func(t *testing.T) {
		t.Parallel()

		const (
			l = 16
			a = 1.01
		)

		params := Params{
			Length: l, SlowestSmoothingFactor: a,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		frama, err := NewFractalAdaptiveMovingAverage(&params)
		check("frama == nil", true, frama == nil)
		check("err", erra, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()

		const (
			l = 16
			a = 0.01
		)

		params := Params{
			Length: l, SlowestSmoothingFactor: a,
			BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		frama, err := NewFractalAdaptiveMovingAverage(&params)
		check("frama == nil", true, frama == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()

		const (
			l = 16
			a = 0.01
		)

		params := Params{
			Length: l, SlowestSmoothingFactor: a,
			BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		frama, err := NewFractalAdaptiveMovingAverage(&params)
		check("frama == nil", true, frama == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()

		const (
			l = 16
			a = 0.01
		)

		params := Params{
			Length: l, SlowestSmoothingFactor: a,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		frama, err := NewFractalAdaptiveMovingAverage(&params)
		check("frama == nil", true, frama == nil)
		check("err", errtc, err.Error())
	})

	// Zero-value component mnemonic tests.
	// A zero value means "use default, don't show in mnemonic".

	t.Run("all components zero", func(t *testing.T) {
		t.Parallel()

		params := Params{Length: 16, SlowestSmoothingFactor: 0.01}

		frama, err := NewFractalAdaptiveMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "frama(16, 0.010)", frama.mnemonic)
		check("description", "Fractal adaptive moving average frama(16, 0.010)", frama.description)
	})

	t.Run("only bar component set", func(t *testing.T) {
		t.Parallel()

		params := Params{
			Length: 16, SlowestSmoothingFactor: 0.01,
			BarComponent: entities.BarMedianPrice,
		}

		frama, err := NewFractalAdaptiveMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "frama(16, 0.010, hl/2)", frama.mnemonic)
		check("description", "Fractal adaptive moving average frama(16, 0.010, hl/2)", frama.description)
	})

	t.Run("only quote component set", func(t *testing.T) {
		t.Parallel()

		params := Params{
			Length: 16, SlowestSmoothingFactor: 0.01,
			QuoteComponent: entities.QuoteBidPrice,
		}

		frama, err := NewFractalAdaptiveMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "frama(16, 0.010, b)", frama.mnemonic)
		check("description", "Fractal adaptive moving average frama(16, 0.010, b)", frama.description)
	})

	t.Run("only trade component set", func(t *testing.T) {
		t.Parallel()

		params := Params{
			Length: 16, SlowestSmoothingFactor: 0.01,
			TradeComponent: entities.TradeVolume,
		}

		frama, err := NewFractalAdaptiveMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "frama(16, 0.010, v)", frama.mnemonic)
		check("description", "Fractal adaptive moving average frama(16, 0.010, v)", frama.description)
	})
}
