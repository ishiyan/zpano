//nolint:testpackage
package exponentialmovingaverage

//nolint: gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func testExponentialMovingAverageTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

func TestExponentialMovingAverageUpdate(t *testing.T) { //nolint: funlen
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

	input := testExponentialMovingAverageInput()

	t.Run("length = 2, firstIsAverage = true", func(t *testing.T) {
		const (
			i1value   = 93.15  // Index=1 value.
			i2value   = 93.96  // Index=2 value.
			i3value   = 94.71  // Index=3 value.
			i251value = 108.21 // Index=251 (last) value.
		)

		t.Parallel()
		ema := testExponentialMovingAverageCreateLength(2, true)

		for i := 0; i < len(input); i++ {
			act := ema.Update(input[i])

			switch i {
			case 0:
				checkNaN(i, act)
				check(i, input[1], act)
			case 1:
				check(i, i1value, act)
			case 2:
				check(i, i2value, act)
			case 3:
				check(i, i3value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, ema.Update(math.NaN()))
	})

	t.Run("length = 10, firstIsAverage = true", func(t *testing.T) {
		const (
			i9value   = 93.22  // Index=9 value.
			i10value  = 93.75  // Index=10 value.
			i29value  = 86.46  // Index=29 value.
			i251value = 108.97 // Index=251 (last) value.
		)

		t.Parallel()
		ema := testExponentialMovingAverageCreateLength(10, true)

		for i := 0; i < 9; i++ {
			checkNaN(i, ema.Update(input[i]))
		}

		for i := 9; i < len(input); i++ {
			act := ema.Update(input[i])

			switch i {
			case 9:
				check(i, i9value, act)
			case 10:
				check(i, i10value, act)
			case 29:
				check(i, i29value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, ema.Update(math.NaN()))
	})

	t.Run("length = 2, firstIsAverage = false (Metastock)", func(t *testing.T) {
		const (
			// The very first value is the input value.
			i1value   = 93.71  // Index=1 value.
			i2value   = 94.15  // Index=2 value.
			i3value   = 94.78  // Index=3 value.
			i251value = 108.21 // Index=251 (last) value.
		)

		t.Parallel()
		ema := testExponentialMovingAverageCreateLength(2, false)

		for i := 0; i < len(input); i++ {
			act := ema.Update(input[i])

			switch i {
			case 0:
				checkNaN(i, act)
				check(i, input[1], act)
			case 1:
				check(i, i1value, act)
			case 2:
				check(i, i2value, act)
			case 3:
				check(i, i3value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, ema.Update(math.NaN()))
	})

	t.Run("length = 10, firstIsAverage = false (Metastock)", func(t *testing.T) {
		const (
			// The very first value is the input value.
			i9value   = 92.60  // Index=9 value.
			i10value  = 93.24  // Index=10 value.
			i11value  = 93.97  // Index=11 value.
			i30value  = 86.23  // Index=30 value.
			i251value = 108.97 // Index=251 (last) value.
		)

		t.Parallel()
		ema := testExponentialMovingAverageCreateLength(10, false)

		for i := 0; i < 9; i++ {
			checkNaN(i, ema.Update(input[i]))
		}

		for i := 9; i < len(input); i++ {
			act := ema.Update(input[i])

			switch i {
			case 9:
				check(i, i9value, act)
			case 10:
				check(i, i10value, act)
			case 11:
				check(i, i11value, act)
			case 30:
				check(i, i30value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, ema.Update(math.NaN()))
	})
}

func TestExponentialMovingAverageUpdateEntity(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		l     = 2
		alpha = 2. / float64(l+1)
		inp   = 3.
		exp   = alpha * inp
	)

	time := testExponentialMovingAverageTime()
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

		if s.Value != exp {
			t.Errorf("value is incorrect: expected %v, actual %v", exp, s.Value)
		}
	}

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		s := entities.Scalar{Time: time, Value: inp}
		ema := testExponentialMovingAverageCreateLength(l, false)
		ema.Update(0.)
		ema.Update(0.)
		check(ema.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: time, Close: inp}
		ema := testExponentialMovingAverageCreateLength(l, false)
		ema.Update(0.)
		ema.Update(0.)
		check(ema.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: time, Bid: inp, Ask: inp}
		ema := testExponentialMovingAverageCreateLength(l, false)
		ema.Update(0.)
		ema.Update(0.)
		check(ema.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: time, Price: inp}
		ema := testExponentialMovingAverageCreateLength(l, false)
		ema.Update(0.)
		ema.Update(0.)
		check(ema.UpdateTrade(&r))
	})
}

func TestExponentialMovingAverageIsPrimed(t *testing.T) {
	t.Parallel()

	input := testExponentialMovingAverageInput()
	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	t.Run("length = 10, firstIsAverage = true", func(t *testing.T) {
		t.Parallel()
		ema := testExponentialMovingAverageCreateLength(10, true)

		check(0, false, ema.IsPrimed())

		for i := 0; i < 9; i++ {
			ema.Update(input[i])
			check(i+1, false, ema.IsPrimed())
		}

		for i := 9; i < len(input); i++ {
			ema.Update(input[i])
			check(i+1, true, ema.IsPrimed())
		}
	})

	t.Run("length = 10, firstIsAverage = false (Metastock)", func(t *testing.T) {
		t.Parallel()
		ema := testExponentialMovingAverageCreateLength(10, false)

		check(0, false, ema.IsPrimed())

		for i := 0; i < 9; i++ {
			ema.Update(input[i])
			check(i+1, false, ema.IsPrimed())
		}

		for i := 9; i < len(input); i++ {
			ema.Update(input[i])
			check(i+1, true, ema.IsPrimed())
		}
	})
}

func TestExponentialMovingAverageMetadata(t *testing.T) {
	t.Parallel()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	t.Run("length = 10, firstIsAverage = true", func(t *testing.T) {
		t.Parallel()
		ema := testExponentialMovingAverageCreateLength(10, true)
		act := ema.Metadata()

		check("Identifier", core.ExponentialMovingAverage, act.Identifier)
		check("Mnemonic", "ema(10)", act.Mnemonic)
		check("Description", "Exponential moving average ema(10)", act.Description)
		check("len(Outputs)", 1, len(act.Outputs))
		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
		check("Outputs[0].Mnemonic", "ema(10)", act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", "Exponential moving average ema(10)", act.Outputs[0].Description)
	})

	t.Run("alpha = 2/11 = 0.18181818..., firstIsAverage = false", func(t *testing.T) {
		t.Parallel()

		// α = 2 / (ℓ + 1) = 2/11 = 0.18181818...
		const alpha = 2. / 11.

		ema := testExponentialMovingAverageCreateAlpha(alpha, false)
		act := ema.Metadata()

		check("Identifier", core.ExponentialMovingAverage, act.Identifier)
		check("Mnemonic", "ema(10, 0.18181818)", act.Mnemonic)
		check("Description", "Exponential moving average ema(10, 0.18181818)", act.Description)
		check("len(Outputs)", 1, len(act.Outputs))
		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
		check("Outputs[0].Mnemonic", "ema(10, 0.18181818)", act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", "Exponential moving average ema(10, 0.18181818)", act.Outputs[0].Description)
	})

	t.Run("length with non-default bar component", func(t *testing.T) {
		t.Parallel()
		params := ExponentialMovingAverageLengthParams{
			Length: 10, FirstIsAverage: true, BarComponent: entities.BarMedianPrice,
		}

		ema, _ := NewExponentialMovingAverageLength(&params)
		act := ema.Metadata()

		check("Mnemonic", "ema(10, hl/2)", act.Mnemonic)
		check("Description", "Exponential moving average ema(10, hl/2)", act.Description)
	})

	t.Run("alpha with non-default quote component", func(t *testing.T) {
		t.Parallel()
		params := ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 2. / 11., FirstIsAverage: false, QuoteComponent: entities.QuoteBidPrice,
		}

		ema, _ := NewExponentialMovingAverageSmoothingFactor(&params)
		act := ema.Metadata()

		check("Mnemonic", "ema(10, 0.18181818, b)", act.Mnemonic)
		check("Description", "Exponential moving average ema(10, 0.18181818, b)", act.Description)
	})
}

func TestNewExponentialMovingAverage(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		bc     entities.BarComponent   = entities.BarMedianPrice
		qc     entities.QuoteComponent = entities.QuoteMidPrice
		tc     entities.TradeComponent = entities.TradePrice
		length                         = 10
		alpha                          = 2. / 11.

		errlen   = "invalid exponential moving average parameters: length should be positive"
		erralpha = "invalid exponential moving average parameters: smoothing factor should be in range [0, 1]"
		errbc    = "invalid exponential moving average parameters: 9999: unknown bar component"
		errqc    = "invalid exponential moving average parameters: 9999: unknown quote component"
		errtc    = "invalid exponential moving average parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("length > 1, firstIsAverage = false", func(t *testing.T) { //nolint:dupl
		t.Parallel()
		params := ExponentialMovingAverageLengthParams{
			Length: length, FirstIsAverage: false, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		ema, err := NewExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "ema(10, hl/2)", ema.LineIndicator.Mnemonic)
		check("description", "Exponential moving average ema(10, hl/2)", ema.LineIndicator.Description)
		check("firstIsAverage", false, ema.firstIsAverage)
		check("primed", false, ema.primed)
		check("length", length, ema.length)
		check("smoothingFactor", alpha, ema.smoothingFactor)
		check("count", 0, ema.count)
		check("sum", 0., ema.sum)
		check("value", 0., ema.value)
	})

	t.Run("length = 1, firstIsAverage = true", func(t *testing.T) { //nolint:dupl
		t.Parallel()
		params := ExponentialMovingAverageLengthParams{
			Length: 1, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		ema, err := NewExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "ema(1, hl/2)", ema.LineIndicator.Mnemonic)
		check("description", "Exponential moving average ema(1, hl/2)", ema.LineIndicator.Description)
		check("firstIsAverage", true, ema.firstIsAverage)
		check("primed", false, ema.primed)
		check("length", 1, ema.length)
		check("smoothingFactor", 1., ema.smoothingFactor)
		check("count", 0, ema.count)
		check("sum", 0., ema.sum)
		check("value", 0., ema.value)
	})

	t.Run("length = 0", func(t *testing.T) {
		t.Parallel()
		params := ExponentialMovingAverageLengthParams{
			Length: 0, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		ema, err := NewExponentialMovingAverageLength(&params)
		check("ema == nil", true, ema == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length < 0", func(t *testing.T) {
		t.Parallel()
		params := ExponentialMovingAverageLengthParams{
			Length: -1, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		ema, err := NewExponentialMovingAverageLength(&params)
		check("ema == nil", true, ema == nil)
		check("err", errlen, err.Error())
	})

	t.Run("epsilon < α ≤ 1", func(t *testing.T) { //nolint:dupl
		t.Parallel()
		params := ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: alpha, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		ema, err := NewExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "ema(10, 0.18181818, hl/2)", ema.LineIndicator.Mnemonic)
		check("description", "Exponential moving average ema(10, 0.18181818, hl/2)", ema.LineIndicator.Description)
		check("firstIsAverage", true, ema.firstIsAverage)
		check("primed", false, ema.primed)
		check("length", length, ema.length)
		check("smoothingFactor", alpha, ema.smoothingFactor)
		check("count", 0, ema.count)
		check("sum", 0., ema.sum)
		check("value", 0., ema.value)
	})

	t.Run("0 < α < epsilon", func(t *testing.T) { //nolint:dupl
		t.Parallel()
		params := ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 0.000000001, FirstIsAverage: false, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		ema, err := NewExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "ema(199999999, 0.00000001, hl/2)", ema.LineIndicator.Mnemonic)
		check("description", "Exponential moving average ema(199999999, 0.00000001, hl/2)", ema.LineIndicator.Description)
		check("firstIsAverage", false, ema.firstIsAverage)
		check("primed", false, ema.primed)
		check("length", 199999999, ema.length) // 2./0.00000001 - 1.
		check("smoothingFactor", 0.00000001, ema.smoothingFactor)
		check("count", 0, ema.count)
		check("sum", 0., ema.sum)
		check("value", 0., ema.value)
	})

	t.Run("α = 0", func(t *testing.T) { //nolint:dupl
		t.Parallel()
		params := ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 0, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		ema, err := NewExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "ema(199999999, 0.00000001, hl/2)", ema.LineIndicator.Mnemonic)
		check("description", "Exponential moving average ema(199999999, 0.00000001, hl/2)", ema.LineIndicator.Description)
		check("firstIsAverage", true, ema.firstIsAverage)
		check("primed", false, ema.primed)
		check("length", 199999999, ema.length) // 2./0.00000001 - 1.
		check("smoothingFactor", 0.00000001, ema.smoothingFactor)
		check("count", 0, ema.count)
		check("sum", 0., ema.sum)
		check("value", 0., ema.value)
	})

	t.Run("α = 1", func(t *testing.T) { //nolint:dupl
		t.Parallel()
		params := ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 1, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		ema, err := NewExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "ema(1, 1.00000000, hl/2)", ema.LineIndicator.Mnemonic)
		check("description", "Exponential moving average ema(1, 1.00000000, hl/2)", ema.LineIndicator.Description)
		check("firstIsAverage", true, ema.firstIsAverage)
		check("primed", false, ema.primed)
		check("length", 1, ema.length) // 2./1 - 1.
		check("smoothingFactor", 1., ema.smoothingFactor)
		check("count", 0, ema.count)
		check("sum", 0., ema.sum)
		check("value", 0., ema.value)
	})

	t.Run("α < 0", func(t *testing.T) {
		t.Parallel()
		params := ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: -1, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		ema, err := NewExponentialMovingAverageSmoothingFactor(&params)
		check("ema == nil", true, ema == nil)
		check("err", erralpha, err.Error())
	})

	t.Run("α > 1", func(t *testing.T) {
		t.Parallel()
		params := ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 2, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		ema, err := NewExponentialMovingAverageSmoothingFactor(&params)
		check("ema == nil", true, ema == nil)
		check("err", erralpha, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()
		params := ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 1, FirstIsAverage: true,
			BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		ema, err := NewExponentialMovingAverageSmoothingFactor(&params)
		check("ema == nil", true, ema == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()
		params := ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 1, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		ema, err := NewExponentialMovingAverageSmoothingFactor(&params)
		check("ema == nil", true, ema == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()
		params := ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 1, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		ema, err := NewExponentialMovingAverageSmoothingFactor(&params)
		check("ema == nil", true, ema == nil)
		check("err", errtc, err.Error())
	})

	// Zero-value component mnemonic tests.
	// A zero value means "use default, don't show in mnemonic".

	t.Run("all components zero, length", func(t *testing.T) {
		t.Parallel()
		params := ExponentialMovingAverageLengthParams{Length: length}

		ema, err := NewExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "ema(10)", ema.LineIndicator.Mnemonic)
		check("description", "Exponential moving average ema(10)", ema.LineIndicator.Description)
	})

	t.Run("all components zero, alpha", func(t *testing.T) {
		t.Parallel()
		params := ExponentialMovingAverageSmoothingFactorParams{SmoothingFactor: alpha}

		ema, err := NewExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "ema(10, 0.18181818)", ema.LineIndicator.Mnemonic)
		check("description", "Exponential moving average ema(10, 0.18181818)", ema.LineIndicator.Description)
	})

	t.Run("only bar component set", func(t *testing.T) {
		t.Parallel()
		params := ExponentialMovingAverageLengthParams{Length: length, BarComponent: entities.BarMedianPrice}

		ema, err := NewExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "ema(10, hl/2)", ema.LineIndicator.Mnemonic)
		check("description", "Exponential moving average ema(10, hl/2)", ema.LineIndicator.Description)
	})

	t.Run("only quote component set", func(t *testing.T) {
		t.Parallel()
		params := ExponentialMovingAverageLengthParams{Length: length, QuoteComponent: entities.QuoteBidPrice}

		ema, err := NewExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "ema(10, b)", ema.LineIndicator.Mnemonic)
		check("description", "Exponential moving average ema(10, b)", ema.LineIndicator.Description)
	})

	t.Run("only trade component set", func(t *testing.T) {
		t.Parallel()
		params := ExponentialMovingAverageLengthParams{Length: length, TradeComponent: entities.TradeVolume}

		ema, err := NewExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "ema(10, v)", ema.LineIndicator.Mnemonic)
		check("description", "Exponential moving average ema(10, v)", ema.LineIndicator.Description)
	})
}

func testExponentialMovingAverageCreateLength(length int, firstIsAverage bool) *ExponentialMovingAverage {
	params := ExponentialMovingAverageLengthParams{
		Length: length, FirstIsAverage: firstIsAverage,
	}

	ema, _ := NewExponentialMovingAverageLength(&params)

	return ema
}

func testExponentialMovingAverageCreateAlpha(alpha float64, firstIsAverage bool) *ExponentialMovingAverage {
	params := ExponentialMovingAverageSmoothingFactorParams{
		SmoothingFactor: alpha, FirstIsAverage: firstIsAverage,
	}

	ema, _ := NewExponentialMovingAverageSmoothingFactor(&params)

	return ema
}
