//nolint:testpackage
package doubleexponentialmovingaverage

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func TestDoubleExponentialMovingAverageUpdate(t *testing.T) { //nolint:cyclop,funlen,gocognit,gocyclo,maintidx
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

	input := testDoubleExponentialMovingAverageInput()

	t.Run("length = 2, firstIsAverage = true", func(t *testing.T) {
		const (
			i4value   = 94.013 // Index=4 value.
			i5value   = 94.539 // Index=5 value.
			i251value = 107.94 // Index=251 (last) value.
			l         = 2
			lprimed   = 2*l - 2
		)

		t.Parallel()

		dema := testDoubleExponentialMovingAverageCreateLength(l, true)

		for i := 0; i < lprimed; i++ {
			checkNaN(i, dema.Update(input[i]))
		}

		for i := lprimed; i < len(input); i++ {
			act := dema.Update(input[i])

			switch i {
			case 4:
				check(i, i4value, act)
			case 5:
				check(i, i5value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, dema.Update(math.NaN()))
	})

	t.Run("length = 14, firstIsAverage = true", func(t *testing.T) { //nolint:dupl
		const (
			i28value  = 84.347   // Index=28 value.
			i29value  = 84.487   // Index=29 value.
			i30value  = 84.374   // Index=30 value.
			i31value  = 84.772   // Index=31 value.
			i48value  = 89.803   // Index=48 value.
			i251value = 109.4676 // Index=251 (last) value.
			l         = 14
			lprimed   = 2*l - 2
		)

		t.Parallel()

		dema := testDoubleExponentialMovingAverageCreateLength(l, true)

		for i := 0; i < lprimed; i++ {
			checkNaN(i, dema.Update(input[i]))
		}

		for i := lprimed; i < len(input); i++ {
			act := dema.Update(input[i])

			switch i {
			case 28:
				check(i, i28value, act)
			case 29:
				check(i, i29value, act)
			case 30:
				check(i, i30value, act)
			case 31:
				check(i, i31value, act)
			case 48:
				check(i, i48value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, dema.Update(math.NaN()))
	})

	t.Run("length = 2, firstIsAverage = false (Metastock)", func(t *testing.T) {
		const (
			// The very first value is the input value.
			i4value   = 93.977 /*93.960*/ // Index=4 value.
			i5value   = 94.522 // Index=5 value.
			i251value = 107.94 // Index=251 (last) value.
			l         = 2
			lprimed   = 2*l - 2
		)

		t.Parallel()

		dema := testDoubleExponentialMovingAverageCreateLength(l, false)

		for i := 0; i < lprimed; i++ {
			checkNaN(i, dema.Update(input[i]))
		}

		for i := lprimed; i < len(input); i++ {
			act := dema.Update(input[i])

			switch i {
			case 4:
				check(i, i4value, act)
			case 5:
				check(i, i5value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, dema.Update(math.NaN()))
	})

	t.Run("length = 14, firstIsAverage = false (Metastock)", func(t *testing.T) { //nolint:dupl
		const (
			// The very first value is the input value.
			i28value  = 84.87    // 84.91 // Index=28 value.
			i29value  = 84.94    // 84.97 // Index=29 value.
			i30value  = 84.77    // 84.80 // Index=30 value.
			i31value  = 85.12    // 85.14 // Index=31 value.
			i48value  = 89.83    // Index=48 value.
			i251value = 109.4676 // Index=251 (last) value.
			l         = 14
			lprimed   = 2*l - 2
		)

		t.Parallel()

		dema := testDoubleExponentialMovingAverageCreateLength(l, false)

		for i := 0; i < lprimed; i++ {
			checkNaN(i, dema.Update(input[i]))
		}

		for i := lprimed; i < len(input); i++ {
			act := dema.Update(input[i])

			switch i {
			case 28:
				check(i, i28value, act)
			case 29:
				check(i, i29value, act)
			case 30:
				check(i, i30value, act)
			case 31:
				check(i, i31value, act)
			case 48:
				check(i, i48value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, dema.Update(math.NaN()))
	})

	t.Run("length = 26, firstIsAverage = false (Metastock)", func(t *testing.T) {
		t.Parallel()

		const (
			l          = 26
			lprimed    = 2*l - 2
			firstCheck = 216
		)

		dema := testDoubleExponentialMovingAverageCreateLength(l, false)

		in := testDoubleExponentialMovingAverageTascInput()
		exp := testDoubleExponentialMovingAverageTascExpected()
		inlen := len(in)

		for i := 0; i < lprimed; i++ {
			checkNaN(i, dema.Update(in[i]))
		}

		for i := lprimed; i < inlen; i++ {
			act := dema.Update(in[i])

			if i >= firstCheck {
				check(i, exp[i], act)
			}
		}

		checkNaN(0, dema.Update(math.NaN()))
	})
}

func TestDoubleExponentialMovingAverageUpdateEntity(t *testing.T) { //nolint:funlen
	t.Parallel()

	const (
		l        = 2
		lprimed  = 2*l - 2
		inp      = 3.
		expFalse = 2.666666666666667
	)

	time := testDoubleExponentialMovingAverageTime()
	check := func(exp float64, act core.Output) {
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

		if s.Value != exp {
			t.Errorf("value is incorrect: expected %v, actual %v", exp, s.Value)
		}
	}

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		s := entities.Scalar{Time: time, Value: inp}
		dema := testDoubleExponentialMovingAverageCreateLength(l, false)

		for i := 0; i < lprimed; i++ {
			dema.Update(0.)
		}

		check(expFalse, dema.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: time, Close: inp}
		dema := testDoubleExponentialMovingAverageCreateLength(l, false)

		for i := 0; i < lprimed; i++ {
			dema.Update(0.)
		}

		check(expFalse, dema.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: time, Bid: inp, Ask: inp}
		dema := testDoubleExponentialMovingAverageCreateLength(l, false)

		for i := 0; i < lprimed; i++ {
			dema.Update(0.)
		}

		check(expFalse, dema.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: time, Price: inp}
		dema := testDoubleExponentialMovingAverageCreateLength(l, false)

		for i := 0; i < lprimed; i++ {
			dema.Update(0.)
		}

		check(expFalse, dema.UpdateTrade(&r))
	})
}

func TestDoubleExponentialMovingAverageIsPrimed(t *testing.T) { //nolint:dupl
	t.Parallel()

	input := testDoubleExponentialMovingAverageInput()
	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	const (
		l       = 14
		lprimed = 2*l - 2
	)

	t.Run("length = 14, firstIsAverage = true", func(t *testing.T) {
		t.Parallel()

		dema := testDoubleExponentialMovingAverageCreateLength(l, true)

		check(0, false, dema.IsPrimed())

		for i := 0; i < lprimed; i++ {
			dema.Update(input[i])
			check(i+1, false, dema.IsPrimed())
		}

		for i := lprimed; i < len(input); i++ {
			dema.Update(input[i])
			check(i+1, true, dema.IsPrimed())
		}
	})

	t.Run("length = 14, firstIsAverage = false (Metastock)", func(t *testing.T) {
		t.Parallel()

		dema := testDoubleExponentialMovingAverageCreateLength(l, false)

		check(0, false, dema.IsPrimed())

		for i := 0; i < lprimed; i++ {
			dema.Update(input[i])
			check(i+1, false, dema.IsPrimed())
		}

		for i := lprimed; i < len(input); i++ {
			dema.Update(input[i])
			check(i+1, true, dema.IsPrimed())
		}
	})
}

func TestDoubleExponentialMovingAverageMetadata(t *testing.T) {
	t.Parallel()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	t.Run("length = 10, firstIsAverage = true", func(t *testing.T) {
		t.Parallel()

		dema := testDoubleExponentialMovingAverageCreateLength(10, true)
		act := dema.Metadata()

		check("Identifier", core.DoubleExponentialMovingAverage, act.Identifier)
		check("Mnemonic", "dema(10)", act.Mnemonic)
		check("Description", "Double exponential moving average dema(10)", act.Description)
		check("len(Outputs)", 1, len(act.Outputs))
		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
		check("Outputs[0].Mnemonic", "dema(10)", act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", "Double exponential moving average dema(10)", act.Outputs[0].Description)
	})

	t.Run("alpha = 2/11 = 0.18181818..., firstIsAverage = false", func(t *testing.T) {
		t.Parallel()

		// α = 2 / (ℓ + 1) = 2/11 = 0.18181818...
		const alpha = 2. / 11.

		dema := testDoubleExponentialMovingAverageCreateAlpha(alpha, false)
		act := dema.Metadata()

		check("Identifier", core.DoubleExponentialMovingAverage, act.Identifier)
		check("Mnemonic", "dema(10, 0.18181818)", act.Mnemonic)
		check("Description", "Double exponential moving average dema(10, 0.18181818)", act.Description)
		check("len(Outputs)", 1, len(act.Outputs))
		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
		check("Outputs[0].Mnemonic", "dema(10, 0.18181818)", act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", "Double exponential moving average dema(10, 0.18181818)", act.Outputs[0].Description)
	})

	t.Run("length with non-default bar component", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageLengthParams{
			Length: 10, FirstIsAverage: true, BarComponent: entities.BarMedianPrice,
		}

		dema, _ := NewDoubleExponentialMovingAverageLength(&params)
		act := dema.Metadata()

		check("Mnemonic", "dema(10, hl/2)", act.Mnemonic)
		check("Description", "Double exponential moving average dema(10, hl/2)", act.Description)
	})

	t.Run("alpha with non-default quote component", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 2. / 11., FirstIsAverage: false, QuoteComponent: entities.QuoteBidPrice,
		}

		dema, _ := NewDoubleExponentialMovingAverageSmoothingFactor(&params)
		act := dema.Metadata()

		check("Mnemonic", "dema(10, 0.18181818, b)", act.Mnemonic)
		check("Description", "Double exponential moving average dema(10, 0.18181818, b)", act.Description)
	})
}

func TestNewDoubleExponentialMovingAverage(t *testing.T) { //nolint:funlen
	t.Parallel()

	const (
		bc     entities.BarComponent   = entities.BarMedianPrice
		qc     entities.QuoteComponent = entities.QuoteMidPrice
		tc     entities.TradeComponent = entities.TradePrice
		length                         = 10
		alpha                          = 2. / 11.

		errlen   = "invalid double exponential moving average parameters: length should be positive"
		erralpha = "invalid double exponential moving average parameters: smoothing factor should be in range [0, 1]"
		errbc    = "invalid double exponential moving average parameters: 9999: unknown bar component"
		errqc    = "invalid double exponential moving average parameters: 9999: unknown quote component"
		errtc    = "invalid double exponential moving average parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("length > 1, firstIsAverage = false", func(t *testing.T) { //nolint:dupl
		t.Parallel()
		params := DoubleExponentialMovingAverageLengthParams{
			Length: length, FirstIsAverage: false, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		dema, err := NewDoubleExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "dema(10, hl/2)", dema.LineIndicator.Mnemonic)
		check("description", "Double exponential moving average dema(10, hl/2)", dema.LineIndicator.Description)
		check("firstIsAverage", false, dema.firstIsAverage)
		check("primed", false, dema.primed)
		check("length", length, dema.length)
		check("length2", length+length-1, dema.length2)
		check("smoothingFactor", alpha, dema.smoothingFactor)
		check("count", 0, dema.count)
		check("sum", 0., dema.sum)
		check("ema1", 0., dema.ema1)
		check("ema2", 0., dema.ema2)
	})

	t.Run("length = 1, firstIsAverage = true", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageLengthParams{
			Length: 1, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		dema, err := NewDoubleExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "dema(1, hl/2)", dema.LineIndicator.Mnemonic)
		check("description", "Double exponential moving average dema(1, hl/2)", dema.LineIndicator.Description)
		check("firstIsAverage", true, dema.firstIsAverage)
		check("primed", false, dema.primed)
		check("length", 1, dema.length)
		check("length2", 1, dema.length2)
		check("smoothingFactor", 1., dema.smoothingFactor)
	})

	t.Run("length = 0", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageLengthParams{
			Length: 0, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		dema, err := NewDoubleExponentialMovingAverageLength(&params)
		check("dema == nil", true, dema == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length < 0", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageLengthParams{
			Length: -1, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		dema, err := NewDoubleExponentialMovingAverageLength(&params)
		check("dema == nil", true, dema == nil)
		check("err", errlen, err.Error())
	})

	t.Run("epsilon < α ≤ 1", func(t *testing.T) { //nolint:dupl
		t.Parallel()
		params := DoubleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: alpha, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		dema, err := NewDoubleExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "dema(10, 0.18181818, hl/2)", dema.LineIndicator.Mnemonic)
		check("description", "Double exponential moving average dema(10, 0.18181818, hl/2)", dema.LineIndicator.Description)
		check("firstIsAverage", true, dema.firstIsAverage)
		check("primed", false, dema.primed)
		check("length", length, dema.length)
		check("length2", length+length-1, dema.length2)
		check("smoothingFactor", alpha, dema.smoothingFactor)
		check("count", 0, dema.count)
		check("sum", 0., dema.sum)
		check("ema1", 0., dema.ema1)
		check("ema2", 0., dema.ema2)
	})

	t.Run("0 < α < epsilon", func(t *testing.T) { //nolint:dupl
		t.Parallel()

		const (
			alpha  = 0.00000001
			length = 199999999 // 2./0.00000001 - 1.
		)

		params := DoubleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: alpha, FirstIsAverage: false, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		dema, err := NewDoubleExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "dema(199999999, 0.00000001, hl/2)", dema.LineIndicator.Mnemonic)
		check("description", "Double exponential moving average dema(199999999, 0.00000001, hl/2)", dema.LineIndicator.Description)
		check("firstIsAverage", false, dema.firstIsAverage)
		check("primed", false, dema.primed)
		check("length", length, dema.length)
		check("smoothingFactor", alpha, dema.smoothingFactor)
		check("count", 0, dema.count)
		check("sum", 0., dema.sum)
		check("ema1", 0., dema.ema1)
		check("ema2", 0., dema.ema2)
	})

	t.Run("α = 0", func(t *testing.T) { //nolint:dupl
		t.Parallel()

		const (
			alpha  = 0.00000001
			length = 199999999 // 2./0.00000001 - 1.
		)

		params := DoubleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 0, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		dema, err := NewDoubleExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "dema(199999999, 0.00000001, hl/2)", dema.LineIndicator.Mnemonic)
		check("description", "Double exponential moving average dema(199999999, 0.00000001, hl/2)", dema.LineIndicator.Description)
		check("firstIsAverage", true, dema.firstIsAverage)
		check("primed", false, dema.primed)
		check("length", length, dema.length)
		check("smoothingFactor", alpha, dema.smoothingFactor)
		check("count", 0, dema.count)
		check("sum", 0., dema.sum)
		check("ema1", 0., dema.ema1)
		check("ema2", 0., dema.ema2)
	})

	t.Run("α = 1", func(t *testing.T) { //nolint:dupl
		t.Parallel()

		const (
			alpha  = 1
			length = 1 // 2./1 - 1.
		)

		params := DoubleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: alpha, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		dema, err := NewDoubleExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "dema(1, 1.00000000, hl/2)", dema.LineIndicator.Mnemonic)
		check("description", "Double exponential moving average dema(1, 1.00000000, hl/2)", dema.LineIndicator.Description)
		check("firstIsAverage", true, dema.firstIsAverage)
		check("primed", false, dema.primed)
		check("length", length, dema.length)
		check("smoothingFactor", 1., dema.smoothingFactor)
		check("count", 0, dema.count)
		check("sum", 0., dema.sum)
		check("ema1", 0., dema.ema1)
		check("ema2", 0., dema.ema2)
	})

	t.Run("α < 0", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: -1, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		dema, err := NewDoubleExponentialMovingAverageSmoothingFactor(&params)
		check("dema == nil", true, dema == nil)
		check("err", erralpha, err.Error())
	})

	t.Run("α > 1", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 2, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		dema, err := NewDoubleExponentialMovingAverageSmoothingFactor(&params)
		check("dema == nil", true, dema == nil)
		check("err", erralpha, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 1, FirstIsAverage: true,
			BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		dema, err := NewDoubleExponentialMovingAverageSmoothingFactor(&params)
		check("dema == nil", true, dema == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 1, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		dema, err := NewDoubleExponentialMovingAverageSmoothingFactor(&params)
		check("dema == nil", true, dema == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 1, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		dema, err := NewDoubleExponentialMovingAverageSmoothingFactor(&params)
		check("dema == nil", true, dema == nil)
		check("err", errtc, err.Error())
	})

	// Zero-value component mnemonic tests.
	// A zero value means "use default, don't show in mnemonic".

	t.Run("all components zero, length", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageLengthParams{Length: length, FirstIsAverage: true}

		dema, err := NewDoubleExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "dema(10)", dema.LineIndicator.Mnemonic)
		check("description", "Double exponential moving average dema(10)", dema.LineIndicator.Description)
	})

	t.Run("all components zero, alpha", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageSmoothingFactorParams{SmoothingFactor: alpha}

		dema, err := NewDoubleExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "dema(10, 0.18181818)", dema.LineIndicator.Mnemonic)
		check("description", "Double exponential moving average dema(10, 0.18181818)", dema.LineIndicator.Description)
	})

	t.Run("only bar component set", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageLengthParams{Length: length, FirstIsAverage: true, BarComponent: entities.BarMedianPrice}

		dema, err := NewDoubleExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "dema(10, hl/2)", dema.LineIndicator.Mnemonic)
		check("description", "Double exponential moving average dema(10, hl/2)", dema.LineIndicator.Description)
	})

	t.Run("only quote component set", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageLengthParams{Length: length, FirstIsAverage: true, QuoteComponent: entities.QuoteBidPrice}

		dema, err := NewDoubleExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "dema(10, b)", dema.LineIndicator.Mnemonic)
		check("description", "Double exponential moving average dema(10, b)", dema.LineIndicator.Description)
	})

	t.Run("only trade component set", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageLengthParams{Length: length, FirstIsAverage: true, TradeComponent: entities.TradeVolume}

		dema, err := NewDoubleExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "dema(10, v)", dema.LineIndicator.Mnemonic)
		check("description", "Double exponential moving average dema(10, v)", dema.LineIndicator.Description)
	})
}
