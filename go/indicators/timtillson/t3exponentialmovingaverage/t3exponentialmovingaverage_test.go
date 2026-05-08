//nolint:testpackage
package t3exponentialmovingaverage

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func TestT3ExponentialMovingAverageUpdate(t *testing.T) { //nolint: funlen, cyclop
	t.Parallel()

	check := func(index int, exp, act float64) {
		t.Helper()

		if math.Abs(exp-act) > 1e-3 {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	checkNaN := func(index int, act float64) {
		t.Helper()

		if !math.IsNaN(act) {
			t.Errorf("[%v] is incorrect: expected NaN, actual %v", index, act)
		}
	}

	input := testT3ExponentialMovingAverageInput()

	const (
		l       = 5
		lprimed = 6*l - 6
	)

	t.Run("length = 5, firstIsAverage = true", func(t *testing.T) {
		t.Parallel()

		const (
			i250value = 109.032 // Index=250 value.
			i251value = 108.88  // Index=251 (last) value.
		)

		t3 := testT3ExponentialMovingAverageCreateLength(l, true, 0.7)

		for i := 0; i < lprimed; i++ {
			checkNaN(i, t3.Update(input[i]))
		}

		for i := lprimed; i < len(input); i++ {
			act := t3.Update(input[i])

			switch i {
			case 250:
				check(i, i250value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, t3.Update(math.NaN()))
	})

	t.Run("length = 5, firstIsAverage = false (Metastock)", func(t *testing.T) {
		t.Parallel()

		const (
			i24value  = 85.749  // Index=24 value.
			i25value  = 84.380  // Index=25 value.
			i250value = 109.032 // Index=250 value.
			i251value = 108.88  // Index=251 (last) value.
		)

		t3 := testT3ExponentialMovingAverageCreateLength(l, false, 0.7)

		for i := 0; i < lprimed; i++ {
			checkNaN(i, t3.Update(input[i]))
		}

		for i := lprimed; i < len(input); i++ {
			act := t3.Update(input[i])

			switch i {
			case 24:
				check(i, i24value, act)
			case 25:
				check(i, i25value, act)
			case 250:
				check(i, i250value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, t3.Update(math.NaN()))
	})

	t.Run("length = 5, firstIsAverage = true (t3.xls)", func(t *testing.T) {
		t.Parallel()

		const (
			firstCheck = lprimed
		)

		t3 := testT3ExponentialMovingAverageCreateLength(l, true, 0.7)

		exp := testT3ExponentialMovingAverageExpected()

		for i := 0; i < lprimed; i++ {
			checkNaN(i, t3.Update(input[i]))
		}

		for i := lprimed; i < len(input); i++ {
			act := t3.Update(input[i])

			if i >= firstCheck {
				check(i, exp[i], act)
			}
		}

		checkNaN(0, t3.Update(math.NaN()))
	})
}

func TestT3ExponentialMovingAverageUpdateEntity(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		l        = 2
		lprimed  = 6*l - 6
		inp      = 3.
		expFalse = 1.6675884773662544
		expTrue  = 1.6901728395061721
	)

	time := testT3ExponentialMovingAverageTime()
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
		t3 := testT3ExponentialMovingAverageCreateLength(l, false, 0.7)

		for i := 0; i < lprimed; i++ {
			t3.Update(0.)
		}

		check(expFalse, t3.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: time, Close: inp}
		t3 := testT3ExponentialMovingAverageCreateLength(l, true, 0.7)

		for i := 0; i < lprimed; i++ {
			t3.Update(0.)
		}

		check(expTrue, t3.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: time, Bid: inp, Ask: inp}
		t3 := testT3ExponentialMovingAverageCreateLength(l, false, 0.7)

		for i := 0; i < lprimed; i++ {
			t3.Update(0.)
		}

		check(expFalse, t3.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: time, Price: inp}
		t3 := testT3ExponentialMovingAverageCreateLength(l, true, 0.7)

		for i := 0; i < lprimed; i++ {
			t3.Update(0.)
		}

		check(expTrue, t3.UpdateTrade(&r))
	})
}

func TestT3ExponentialMovingAverageIsPrimed(t *testing.T) {
	t.Parallel()

	input := testT3ExponentialMovingAverageInput()
	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	const (
		l       = 5
		lprimed = 6*l - 6
	)

	t.Run("length = 5, firstIsAverage = true", func(t *testing.T) {
		t.Parallel()

		t3 := testT3ExponentialMovingAverageCreateLength(l, true, 0.7)

		check(0, false, t3.IsPrimed())

		for i := 0; i < lprimed; i++ {
			t3.Update(input[i])
			check(i+1, false, t3.IsPrimed())
		}

		for i := lprimed; i < len(input); i++ {
			t3.Update(input[i])
			check(i+1, true, t3.IsPrimed())
		}
	})

	t.Run("length = 5, firstIsAverage = false (Metastock)", func(t *testing.T) {
		t.Parallel()

		t3 := testT3ExponentialMovingAverageCreateLength(l, false, 0.7)

		check(0, false, t3.IsPrimed())

		for i := 0; i < lprimed; i++ {
			t3.Update(input[i])
			check(i+1, false, t3.IsPrimed())
		}

		for i := lprimed; i < len(input); i++ {
			t3.Update(input[i])
			check(i+1, true, t3.IsPrimed())
		}
	})
}

func TestT3ExponentialMovingAverageMetadata(t *testing.T) {
	t.Parallel()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	t.Run("length = 10, v=0.3333, firstIsAverage = true", func(t *testing.T) {
		t.Parallel()

		t3 := testT3ExponentialMovingAverageCreateLength(10, true, 0.3333)
		act := t3.Metadata()

		check("Identifier", core.T3ExponentialMovingAverage, act.Identifier)
		check("Mnemonic", "t3(10, 0.33330000)", act.Mnemonic)
		check("Description", "T3 exponential moving average t3(10, 0.33330000)", act.Description)
		check("len(Outputs)", 1, len(act.Outputs))
		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
		check("Outputs[0].Mnemonic", "t3(10, 0.33330000)", act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", "T3 exponential moving average t3(10, 0.33330000)", act.Outputs[0].Description)
	})

	t.Run("alpha = 2/11 = 0.18181818..., v=0.3333333, firstIsAverage = false", func(t *testing.T) {
		t.Parallel()

		// α = 2 / (ℓ + 1) = 2/11 = 0.18181818...
		const alpha = 2. / 11.

		t3 := testT3ExponentialMovingAverageCreateAlpha(alpha, false, 0.3333333)
		act := t3.Metadata()

		check("Identifier", core.T3ExponentialMovingAverage, act.Identifier)
		check("Mnemonic", "t3(10, 0.18181818, 0.33333330)", act.Mnemonic)
		check("Description", "T3 exponential moving average t3(10, 0.18181818, 0.33333330)", act.Description)
		check("len(Outputs)", 1, len(act.Outputs))
		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
		check("Outputs[0].Mnemonic", "t3(10, 0.18181818, 0.33333330)", act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", "T3 exponential moving average t3(10, 0.18181818, 0.33333330)", act.Outputs[0].Description)
	})

	t.Run("length with non-default bar component", func(t *testing.T) {
		t.Parallel()
		params := T3ExponentialMovingAverageLengthParams{
			Length: 10, VolumeFactor: 0.7, FirstIsAverage: true, BarComponent: entities.BarMedianPrice,
		}

		t3, _ := NewT3ExponentialMovingAverageLength(&params)
		act := t3.Metadata()

		check("Mnemonic", "t3(10, 0.70000000, hl/2)", act.Mnemonic)
		check("Description", "T3 exponential moving average t3(10, 0.70000000, hl/2)", act.Description)
	})

	t.Run("alpha with non-default quote component", func(t *testing.T) {
		t.Parallel()
		params := T3ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 2. / 11., VolumeFactor: 0.7, FirstIsAverage: false, QuoteComponent: entities.QuoteBidPrice,
		}

		t3, _ := NewT3ExponentialMovingAverageSmoothingFactor(&params)
		act := t3.Metadata()

		check("Mnemonic", "t3(10, 0.18181818, 0.70000000, b)", act.Mnemonic)
		check("Description", "T3 exponential moving average t3(10, 0.18181818, 0.70000000, b)", act.Description)
	})
}

func TestNewT3ExponentialMovingAverage(t *testing.T) { //nolint: funlen, maintidx
	t.Parallel()

	const (
		bc     entities.BarComponent   = entities.BarMedianPrice
		qc     entities.QuoteComponent = entities.QuoteMidPrice
		tc     entities.TradeComponent = entities.TradePrice
		length                         = 10
		alpha                          = 2. / 11.

		errlen   = "invalid t3 exponential moving average parameters: length should be greater than 1"
		erralpha = "invalid t3 exponential moving average parameters: smoothing factor should be in range [0, 1]"
		errvol   = "invalid t3 exponential moving average parameters: volume factor should be in range [0, 1]"
		errbc    = "invalid t3 exponential moving average parameters: 9999: unknown bar component"
		errqc    = "invalid t3 exponential moving average parameters: 9999: unknown quote component"
		errtc    = "invalid t3 exponential moving average parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	checkInstance := func(
		t3 *T3ExponentialMovingAverage, mnemonic string, length int, alpha float64, firstIsAverage bool,
	) {
		check("mnemonic", mnemonic, t3.LineIndicator.Mnemonic)
		check("description", "T3 exponential moving average "+mnemonic, t3.LineIndicator.Description)
		check("firstIsAverage", firstIsAverage, t3.firstIsAverage)
		check("primed", false, t3.primed)
		check("length", length, t3.length)
		check("length2", length+length-1, t3.length2)
		check("length3", length+length+length-2, t3.length3)
		check("length4", length+length+length+length-3, t3.length4)
		check("length5", length+length+length+length+length-4, t3.length5)
		check("length6", length+length+length+length+length+length-5, t3.length6)
		check("smoothingFactor", alpha, t3.smoothingFactor)
		check("count", 0, t3.count)
		check("sum", 0., t3.sum)
		check("ema1", 0., t3.ema1)
		check("ema2", 0., t3.ema2)
		check("ema3", 0., t3.ema3)
		check("ema4", 0., t3.ema4)
		check("ema5", 0., t3.ema5)
		check("ema6", 0., t3.ema6)
	}

	t.Run("length > 1, firstIsAverage = false", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageLengthParams{
			Length: length, VolumeFactor: 0.7, FirstIsAverage: false,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		checkInstance(t3, "t3(10, 0.70000000, hl/2)", length, alpha, false)
	})

	t.Run("length = 1, firstIsAverage = true", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageLengthParams{
			Length: 1, VolumeFactor: 0.7, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageLength(&params)
		check("t3 == nil", true, t3 == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length = 0", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageLengthParams{
			Length: 0, VolumeFactor: 0.7, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageLength(&params)
		check("t3 == nil", true, t3 == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length < 0", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageLengthParams{
			Length: -1, VolumeFactor: 0.7, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageLength(&params)
		check("t3 == nil", true, t3 == nil)
		check("err", errlen, err.Error())
	})

	t.Run("epsilon < α ≤ 1", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: alpha, VolumeFactor: 0.7, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		checkInstance(t3, "t3(10, 0.18181818, 0.70000000, hl/2)", length, alpha, true)
	})

	t.Run("0 < α < epsilon", func(t *testing.T) {
		t.Parallel()

		const (
			alpha  = 0.00000001
			length = 199999999 // 2./0.00000001 - 1.
		)

		params := T3ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: alpha, VolumeFactor: 0.7, FirstIsAverage: false,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		checkInstance(t3, "t3(199999999, 0.00000001, 0.70000000, hl/2)", length, alpha, false)
	})

	t.Run("α = 0", func(t *testing.T) {
		t.Parallel()

		const (
			alpha  = 0.00000001
			length = 199999999 // 2./0.00000001 - 1.
		)

		params := T3ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 0, VolumeFactor: 0.7, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		checkInstance(t3, "t3(199999999, 0.00000001, 0.70000000, hl/2)", length, alpha, true)
	})

	t.Run("α = 1", func(t *testing.T) {
		t.Parallel()

		const (
			alpha  = 1
			length = 1 // 2./1 - 1.
		)

		params := T3ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: alpha, VolumeFactor: 0.7, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		checkInstance(t3, "t3(1, 1.00000000, 0.70000000, hl/2)", length, alpha, true)
	})

	t.Run("α < 0", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: -1, VolumeFactor: 0.7, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageSmoothingFactor(&params)
		check("t3 == nil", true, t3 == nil)
		check("err", erralpha, err.Error())
	})

	t.Run("α > 1", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 2, VolumeFactor: 0.7, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageSmoothingFactor(&params)
		check("t3 == nil", true, t3 == nil)
		check("err", erralpha, err.Error())
	})

	t.Run("volume factor = 0.5", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageLengthParams{
			Length: length, VolumeFactor: 0.5, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		checkInstance(t3, "t3(10, 0.50000000, hl/2)", length, alpha, true)
	})

	t.Run("volume factor = 0", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageLengthParams{
			Length: length, VolumeFactor: 0, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		checkInstance(t3, "t3(10, 0.00000000, hl/2)", length, alpha, true)
	})

	t.Run("volume factor = 1", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageLengthParams{
			Length: length, VolumeFactor: 1, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		checkInstance(t3, "t3(10, 1.00000000, hl/2)", length, alpha, true)
	})

	t.Run("volume factor < 0", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageLengthParams{
			Length: 3, VolumeFactor: -0.7, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageLength(&params)
		check("t3 == nil", true, t3 == nil)
		check("err", errvol, err.Error())
	})

	t.Run("volume factor > 1", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageLengthParams{
			Length: 3, VolumeFactor: 1.7, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageLength(&params)
		check("t3 == nil", true, t3 == nil)
		check("err", errvol, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 1, VolumeFactor: 0.5, FirstIsAverage: true,
			BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageSmoothingFactor(&params)
		check("t3 == nil", true, t3 == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 1, VolumeFactor: 0.5, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageSmoothingFactor(&params)
		check("t3 == nil", true, t3 == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 1, VolumeFactor: 0.5, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		t3, err := NewT3ExponentialMovingAverageSmoothingFactor(&params)
		check("t3 == nil", true, t3 == nil)
		check("err", errtc, err.Error())
	})

	// Zero-value component mnemonic tests.
	// A zero value means "use default, don't show in mnemonic".

	t.Run("all components zero, length", func(t *testing.T) {
		t.Parallel()
		params := T3ExponentialMovingAverageLengthParams{Length: length, VolumeFactor: 0.7, FirstIsAverage: true}

		t3, err := NewT3ExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "t3(10, 0.70000000)", t3.LineIndicator.Mnemonic)
		check("description", "T3 exponential moving average t3(10, 0.70000000)", t3.LineIndicator.Description)
	})

	t.Run("all components zero, alpha", func(t *testing.T) {
		t.Parallel()
		params := T3ExponentialMovingAverageSmoothingFactorParams{SmoothingFactor: alpha, VolumeFactor: 0.7}

		t3, err := NewT3ExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "t3(10, 0.18181818, 0.70000000)", t3.LineIndicator.Mnemonic)
		check("description", "T3 exponential moving average t3(10, 0.18181818, 0.70000000)", t3.LineIndicator.Description)
	})

	t.Run("only bar component set", func(t *testing.T) {
		t.Parallel()
		params := T3ExponentialMovingAverageLengthParams{
			Length: length, VolumeFactor: 0.7, FirstIsAverage: true, BarComponent: entities.BarMedianPrice,
		}

		t3, err := NewT3ExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "t3(10, 0.70000000, hl/2)", t3.LineIndicator.Mnemonic)
		check("description", "T3 exponential moving average t3(10, 0.70000000, hl/2)", t3.LineIndicator.Description)
	})

	t.Run("only quote component set", func(t *testing.T) {
		t.Parallel()
		params := T3ExponentialMovingAverageLengthParams{
			Length: length, VolumeFactor: 0.7, FirstIsAverage: true, QuoteComponent: entities.QuoteBidPrice,
		}

		t3, err := NewT3ExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "t3(10, 0.70000000, b)", t3.LineIndicator.Mnemonic)
		check("description", "T3 exponential moving average t3(10, 0.70000000, b)", t3.LineIndicator.Description)
	})

	t.Run("only trade component set", func(t *testing.T) {
		t.Parallel()
		params := T3ExponentialMovingAverageLengthParams{
			Length: length, VolumeFactor: 0.7, FirstIsAverage: true, TradeComponent: entities.TradeVolume,
		}

		t3, err := NewT3ExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "t3(10, 0.70000000, v)", t3.LineIndicator.Mnemonic)
		check("description", "T3 exponential moving average t3(10, 0.70000000, v)", t3.LineIndicator.Description)
	})
}
