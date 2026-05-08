//nolint:testpackage
package tripleexponentialmovingaverage

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
// Output data is taken from TA-Lib (http://ta-lib.org/) tests,
//    test_ma.c.
//
//   /*******************************/
//   /*  TEMA TEST - Metastock      */
//   /*******************************/
//   /* No output value. */
//   { 0, TA_ANY_MA_TEST, 0, 1, 1,  14, TA_MAType_TEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS, 0, 0, 0, 0},
//#ifndef TA_FUNC_NO_RANGE_CHECK
//   { 0, TA_ANY_MA_TEST, 0, 0, 251,  0, TA_MAType_TEMA, TA_COMPATIBILITY_METASTOCK, TA_BAD_PARAM, 0, 0, 0, 0 },
//#endif
//
//   /* Test with period 14 */
//   { 1, TA_ANY_MA_TEST, 0, 0, 251, 14, TA_MAType_TEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   0,  84.721, 39, 252-39 }, /* First Value */
//   { 0, TA_ANY_MA_TEST, 0, 0, 251, 14, TA_MAType_TEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   1,  84.089, 39, 252-39 },
//   { 0, TA_ANY_MA_TEST, 0, 0, 251, 14, TA_MAType_TEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS, 252-40, 108.418, 39, 252-39 }, /* Last Value */

func TestTripleExponentialMovingAverageUpdate(t *testing.T) { //nolint: funlen, cyclop
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

	input := testTripleExponentialMovingAverageInput()

	const (
		l       = 14
		lprimed = 3*l - 3
	)

	t.Run("length = 14, firstIsAverage = true", func(t *testing.T) {
		t.Parallel()

		const (
			i39value  = 84.8629 // Index=39 value.
			i40value  = 84.2246 // Index=40 value.
			i251value = 108.418 // Index=251 (last) value.
		)

		tema := testTripleExponentialMovingAverageCreateLength(l, true)

		for i := 0; i < lprimed; i++ {
			checkNaN(i, tema.Update(input[i]))
		}

		for i := lprimed; i < len(input); i++ {
			act := tema.Update(input[i])

			switch i {
			case 39:
				check(i, i39value, act)
			case 40:
				check(i, i40value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, tema.Update(math.NaN()))
	})

	t.Run("length = 14, firstIsAverage = false (Metastock)", func(t *testing.T) {
		t.Parallel()

		const (
			i39value  = 84.721  // Index=39 value.
			i40value  = 84.089  // Index=40 value.
			i251value = 108.418 // Index=251 (last) value.
		)

		tema := testTripleExponentialMovingAverageCreateLength(l, false)

		for i := 0; i < lprimed; i++ {
			checkNaN(i, tema.Update(input[i]))
		}

		for i := lprimed; i < len(input); i++ {
			act := tema.Update(input[i])

			switch i {
			case 39:
				check(i, i39value, act)
			case 40:
				check(i, i40value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, tema.Update(math.NaN()))
	})

	t.Run("length = 26, firstIsAverage = false (Metastock)", func(t *testing.T) {
		t.Parallel()

		const (
			l          = 26
			lprimed    = 3*l - 3
			firstCheck = 216
		)

		tema := testTripleExponentialMovingAverageCreateLength(l, false)

		in := testTripleExponentialMovingAverageTascInput()
		exp := testTripleExponentialMovingAverageTascExpected()
		inlen := len(in)

		for i := 0; i < lprimed; i++ {
			checkNaN(i, tema.Update(in[i]))
		}

		for i := lprimed; i < inlen; i++ {
			act := tema.Update(in[i])

			if i >= firstCheck {
				check(i, exp[i], act)
			}
		}

		checkNaN(0, tema.Update(math.NaN()))
	})
}

func TestTripleExponentialMovingAverageUpdateEntity(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		l        = 2
		lprimed  = 3*l - 3
		inp      = 3.
		expFalse = 2.888888888888889
		expTrue  = 2.6666666666666665
	)

	time := testTripleExponentialMovingAverageTime()
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
		tema := testTripleExponentialMovingAverageCreateLength(l, false)

		for i := 0; i < lprimed; i++ {
			tema.Update(0.)
		}

		check(expFalse, tema.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: time, Close: inp}
		tema := testTripleExponentialMovingAverageCreateLength(l, true)

		for i := 0; i < lprimed; i++ {
			tema.Update(0.)
		}

		check(expTrue, tema.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: time, Bid: inp, Ask: inp}
		tema := testTripleExponentialMovingAverageCreateLength(l, false)

		for i := 0; i < lprimed; i++ {
			tema.Update(0.)
		}

		check(expFalse, tema.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: time, Price: inp}
		tema := testTripleExponentialMovingAverageCreateLength(l, true)

		for i := 0; i < lprimed; i++ {
			tema.Update(0.)
		}

		check(expTrue, tema.UpdateTrade(&r))
	})
}

func TestTripleExponentialMovingAverageIsPrimed(t *testing.T) { //nolint:dupl
	t.Parallel()

	input := testTripleExponentialMovingAverageInput()
	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	const (
		l       = 14
		lprimed = 3*l - 3
	)

	t.Run("length = 14, firstIsAverage = true", func(t *testing.T) {
		t.Parallel()

		tema := testTripleExponentialMovingAverageCreateLength(l, true)

		check(0, false, tema.IsPrimed())

		for i := 0; i < lprimed; i++ {
			tema.Update(input[i])
			check(i+1, false, tema.IsPrimed())
		}

		for i := lprimed; i < len(input); i++ {
			tema.Update(input[i])
			check(i+1, true, tema.IsPrimed())
		}
	})

	t.Run("length = 14, firstIsAverage = false (Metastock)", func(t *testing.T) {
		t.Parallel()

		tema := testTripleExponentialMovingAverageCreateLength(l, false)

		check(0, false, tema.IsPrimed())

		for i := 0; i < lprimed; i++ {
			tema.Update(input[i])
			check(i+1, false, tema.IsPrimed())
		}

		for i := lprimed; i < len(input); i++ {
			tema.Update(input[i])
			check(i+1, true, tema.IsPrimed())
		}
	})
}

func TestTripleExponentialMovingAverageMetadata(t *testing.T) {
	t.Parallel()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	t.Run("length = 10, firstIsAverage = true", func(t *testing.T) {
		t.Parallel()

		tema := testTripleExponentialMovingAverageCreateLength(10, true)
		act := tema.Metadata()

		check("Identifier", core.TripleExponentialMovingAverage, act.Identifier)
		check("Mnemonic", "tema(10)", act.Mnemonic)
		check("Description", "Triple exponential moving average tema(10)", act.Description)
		check("len(Outputs)", 1, len(act.Outputs))
		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
		check("Outputs[0].Mnemonic", "tema(10)", act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", "Triple exponential moving average tema(10)", act.Outputs[0].Description)
	})

	t.Run("alpha = 2/11 = 0.18181818..., firstIsAverage = false", func(t *testing.T) {
		t.Parallel()

		// α = 2 / (ℓ + 1) = 2/11 = 0.18181818...
		const alpha = 2. / 11.

		tema := testTripleExponentialMovingAverageCreateAlpha(alpha, false)
		act := tema.Metadata()

		check("Identifier", core.TripleExponentialMovingAverage, act.Identifier)
		check("Mnemonic", "tema(10, 0.18181818)", act.Mnemonic)
		check("Description", "Triple exponential moving average tema(10, 0.18181818)", act.Description)
		check("len(Outputs)", 1, len(act.Outputs))
		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
		check("Outputs[0].Mnemonic", "tema(10, 0.18181818)", act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", "Triple exponential moving average tema(10, 0.18181818)", act.Outputs[0].Description)
	})

	t.Run("length with non-default bar component", func(t *testing.T) {
		t.Parallel()
		params := TripleExponentialMovingAverageLengthParams{
			Length: 10, FirstIsAverage: true, BarComponent: entities.BarMedianPrice,
		}

		tema, _ := NewTripleExponentialMovingAverageLength(&params)
		act := tema.Metadata()

		check("Mnemonic", "tema(10, hl/2)", act.Mnemonic)
		check("Description", "Triple exponential moving average tema(10, hl/2)", act.Description)
	})

	t.Run("alpha with non-default quote component", func(t *testing.T) {
		t.Parallel()
		params := TripleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 2. / 11., FirstIsAverage: false, QuoteComponent: entities.QuoteBidPrice,
		}

		tema, _ := NewTripleExponentialMovingAverageSmoothingFactor(&params)
		act := tema.Metadata()

		check("Mnemonic", "tema(10, 0.18181818, b)", act.Mnemonic)
		check("Description", "Triple exponential moving average tema(10, 0.18181818, b)", act.Description)
	})
}

func TestNewTripleExponentialMovingAverage(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		bc     entities.BarComponent   = entities.BarMedianPrice
		qc     entities.QuoteComponent = entities.QuoteMidPrice
		tc     entities.TradeComponent = entities.TradePrice
		length                         = 10
		alpha                          = 2. / 11.

		errlen   = "invalid triple exponential moving average parameters: length should be greater than 1"
		erralpha = "invalid triple exponential moving average parameters: smoothing factor should be in range [0, 1]"
		errbc    = "invalid triple exponential moving average parameters: 9999: unknown bar component"
		errqc    = "invalid triple exponential moving average parameters: 9999: unknown quote component"
		errtc    = "invalid triple exponential moving average parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("length > 1, firstIsAverage = false", func(t *testing.T) { //nolint:dupl
		t.Parallel()
		params := TripleExponentialMovingAverageLengthParams{
			Length: length, FirstIsAverage: false, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		tema, err := NewTripleExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "tema(10, hl/2)", tema.LineIndicator.Mnemonic)
		check("description", "Triple exponential moving average tema(10, hl/2)", tema.LineIndicator.Description)
		check("firstIsAverage", false, tema.firstIsAverage)
		check("primed", false, tema.primed)
		check("length", length, tema.length)
		check("length2", length+length-1, tema.length2)
		check("length3", length+length+length-2, tema.length3)
		check("smoothingFactor", alpha, tema.smoothingFactor)
		check("count", 0, tema.count)
		check("sum", 0., tema.sum)
		check("ema1", 0., tema.ema1)
		check("ema2", 0., tema.ema2)
		check("ema3", 0., tema.ema3)
	})

	t.Run("length = 1, firstIsAverage = true", func(t *testing.T) {
		t.Parallel()
		params := TripleExponentialMovingAverageLengthParams{
			Length: 1, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		tema, err := NewTripleExponentialMovingAverageLength(&params)
		check("tema == nil", true, tema == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length = 0", func(t *testing.T) {
		t.Parallel()
		params := TripleExponentialMovingAverageLengthParams{
			Length: 0, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		tema, err := NewTripleExponentialMovingAverageLength(&params)
		check("tema == nil", true, tema == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length < 0", func(t *testing.T) {
		t.Parallel()
		params := TripleExponentialMovingAverageLengthParams{
			Length: -1, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		tema, err := NewTripleExponentialMovingAverageLength(&params)
		check("tema == nil", true, tema == nil)
		check("err", errlen, err.Error())
	})

	t.Run("epsilon < α ≤ 1", func(t *testing.T) { //nolint:dupl
		t.Parallel()
		params := TripleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: alpha, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		tema, err := NewTripleExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "tema(10, 0.18181818, hl/2)", tema.LineIndicator.Mnemonic)
		check("description", "Triple exponential moving average tema(10, 0.18181818, hl/2)", tema.LineIndicator.Description)
		check("firstIsAverage", true, tema.firstIsAverage)
		check("primed", false, tema.primed)
		check("length", length, tema.length)
		check("length2", length+length-1, tema.length2)
		check("length3", length+length+length-2, tema.length3)
		check("smoothingFactor", alpha, tema.smoothingFactor)
		check("count", 0, tema.count)
		check("sum", 0., tema.sum)
		check("ema1", 0., tema.ema1)
		check("ema2", 0., tema.ema2)
		check("ema3", 0., tema.ema3)
	})

	t.Run("0 < α < epsilon", func(t *testing.T) { //nolint:dupl
		t.Parallel()

		const (
			alpha  = 0.00000001
			length = 199999999 // 2./0.00000001 - 1.
		)

		params := TripleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: alpha, FirstIsAverage: false, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		tema, err := NewTripleExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "tema(199999999, 0.00000001, hl/2)", tema.LineIndicator.Mnemonic)
		check("description", "Triple exponential moving average tema(199999999, 0.00000001, hl/2)", tema.LineIndicator.Description)
		check("firstIsAverage", false, tema.firstIsAverage)
		check("primed", false, tema.primed)
		check("length", length, tema.length)
		check("smoothingFactor", alpha, tema.smoothingFactor)
		check("count", 0, tema.count)
		check("sum", 0., tema.sum)
		check("ema1", 0., tema.ema1)
		check("ema2", 0., tema.ema2)
		check("ema3", 0., tema.ema3)
	})

	t.Run("α = 0", func(t *testing.T) { //nolint:dupl
		t.Parallel()

		const (
			alpha  = 0.00000001
			length = 199999999 // 2./0.00000001 - 1.
		)

		params := TripleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 0, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		tema, err := NewTripleExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "tema(199999999, 0.00000001, hl/2)", tema.LineIndicator.Mnemonic)
		check("description", "Triple exponential moving average tema(199999999, 0.00000001, hl/2)", tema.LineIndicator.Description)
		check("firstIsAverage", true, tema.firstIsAverage)
		check("primed", false, tema.primed)
		check("length", length, tema.length)
		check("smoothingFactor", alpha, tema.smoothingFactor)
		check("count", 0, tema.count)
		check("sum", 0., tema.sum)
		check("ema1", 0., tema.ema1)
		check("ema2", 0., tema.ema2)
		check("ema3", 0., tema.ema3)
	})

	t.Run("α = 1", func(t *testing.T) { //nolint:dupl
		t.Parallel()

		const (
			alpha  = 1
			length = 1 // 2./1 - 1.
		)

		params := TripleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: alpha, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		tema, err := NewTripleExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "tema(1, 1.00000000, hl/2)", tema.LineIndicator.Mnemonic)
		check("description", "Triple exponential moving average tema(1, 1.00000000, hl/2)", tema.LineIndicator.Description)
		check("firstIsAverage", true, tema.firstIsAverage)
		check("primed", false, tema.primed)
		check("length", length, tema.length)
		check("smoothingFactor", 1., tema.smoothingFactor)
		check("count", 0, tema.count)
		check("sum", 0., tema.sum)
		check("ema1", 0., tema.ema1)
		check("ema2", 0., tema.ema2)
		check("ema3", 0., tema.ema3)
	})

	t.Run("α < 0", func(t *testing.T) {
		t.Parallel()
		params := TripleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: -1, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		tema, err := NewTripleExponentialMovingAverageSmoothingFactor(&params)
		check("tema == nil", true, tema == nil)
		check("err", erralpha, err.Error())
	})

	t.Run("α > 1", func(t *testing.T) {
		t.Parallel()
		params := TripleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 2, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		tema, err := NewTripleExponentialMovingAverageSmoothingFactor(&params)
		check("tema == nil", true, tema == nil)
		check("err", erralpha, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()
		params := TripleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 1, FirstIsAverage: true,
			BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		tema, err := NewTripleExponentialMovingAverageSmoothingFactor(&params)
		check("tema == nil", true, tema == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()
		params := TripleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 1, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		tema, err := NewTripleExponentialMovingAverageSmoothingFactor(&params)
		check("tema == nil", true, tema == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()
		params := TripleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 1, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		tema, err := NewTripleExponentialMovingAverageSmoothingFactor(&params)
		check("tema == nil", true, tema == nil)
		check("err", errtc, err.Error())
	})

	// Zero-value component mnemonic tests.
	// A zero value means "use default, don't show in mnemonic".

	t.Run("all components zero, length", func(t *testing.T) {
		t.Parallel()
		params := TripleExponentialMovingAverageLengthParams{Length: length, FirstIsAverage: true}

		tema, err := NewTripleExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "tema(10)", tema.LineIndicator.Mnemonic)
		check("description", "Triple exponential moving average tema(10)", tema.LineIndicator.Description)
	})

	t.Run("all components zero, alpha", func(t *testing.T) {
		t.Parallel()
		params := TripleExponentialMovingAverageSmoothingFactorParams{SmoothingFactor: alpha}

		tema, err := NewTripleExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "tema(10, 0.18181818)", tema.LineIndicator.Mnemonic)
		check("description", "Triple exponential moving average tema(10, 0.18181818)", tema.LineIndicator.Description)
	})

	t.Run("only bar component set", func(t *testing.T) {
		t.Parallel()
		params := TripleExponentialMovingAverageLengthParams{Length: length, FirstIsAverage: true, BarComponent: entities.BarMedianPrice}

		tema, err := NewTripleExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "tema(10, hl/2)", tema.LineIndicator.Mnemonic)
		check("description", "Triple exponential moving average tema(10, hl/2)", tema.LineIndicator.Description)
	})

	t.Run("only quote component set", func(t *testing.T) {
		t.Parallel()
		params := TripleExponentialMovingAverageLengthParams{Length: length, FirstIsAverage: true, QuoteComponent: entities.QuoteBidPrice}

		tema, err := NewTripleExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "tema(10, b)", tema.LineIndicator.Mnemonic)
		check("description", "Triple exponential moving average tema(10, b)", tema.LineIndicator.Description)
	})

	t.Run("only trade component set", func(t *testing.T) {
		t.Parallel()
		params := TripleExponentialMovingAverageLengthParams{Length: length, FirstIsAverage: true, TradeComponent: entities.TradeVolume}

		tema, err := NewTripleExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "tema(10, v)", tema.LineIndicator.Mnemonic)
		check("description", "Triple exponential moving average tema(10, v)", tema.LineIndicator.Description)
	})
}
