//nolint:testpackage
package triangularmovingaverage

//nolint: gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func testTriangularMovingAverageTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

func TestTriangularMovingAverageUpdate(t *testing.T) { //nolint: funlen
	t.Parallel()

	check := func(index int, exp, act float64) {
		t.Helper()

		if math.Abs(exp-act) > 1e-4 {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	checkNaN := func(index int, act float64) {
		t.Helper()

		if !math.IsNaN(act) {
			t.Errorf("[%v] is incorrect: expected NaN, actual %v", index, act)
		}
	}

	input := testTriangularMovingAverageInput()

	t.Run("length = 9", func(t *testing.T) {
		const ( // Values from index=0 to index=7 are NaN.
			i8value   = 93.8176  // Index=8 value.
			i251value = 109.1312 // Index=251 (last) value.
		)

		t.Parallel()
		trima := testTriangularMovingAverageCreate(9)

		for i := 0; i < 8; i++ {
			checkNaN(i, trima.Update(input[i]))
		}

		for i := 8; i < len(input); i++ {
			act := trima.Update(input[i])

			switch i {
			case 8:
				check(i, i8value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, trima.Update(math.NaN()))
	})

	t.Run("length = 10", func(t *testing.T) {
		const ( // Values from index=0 to index=8 are NaN.
			i9value   = 93.6043  // Index=9 value.
			i10value  = 93.4252  // Index=10 value.
			i250value = 109.1850 // Index=250 value.
			i251value = 109.1407 // Index=251 (last) value.
		)

		t.Parallel()
		trima := testTriangularMovingAverageCreate(10)

		for i := 0; i < 9; i++ {
			checkNaN(i, trima.Update(input[i]))
		}

		for i := 9; i < len(input); i++ {
			act := trima.Update(input[i])

			switch i {
			case 9:
				check(i, i9value, act)
			case 10:
				check(i, i10value, act)
			case 250:
				check(i, i250value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, trima.Update(math.NaN()))
	})

	t.Run("length = 12", func(t *testing.T) {
		const ( // Values from index=0 to index=10 are NaN.
			i11value  = 93.5329  // Index=11 value.
			i251value = 109.1157 // Index=251 (last) value.
		)

		t.Parallel()
		trima := testTriangularMovingAverageCreate(12)

		for i := 0; i < 10; i++ {
			checkNaN(i, trima.Update(input[i]))
		}

		for i := 10; i < len(input); i++ {
			act := trima.Update(input[i])

			switch i {
			case 11:
				check(i, i11value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, trima.Update(math.NaN()))
	})
}

func TestTriangularMovingAverageUpdateXls(t *testing.T) {
	t.Parallel()

	check := func(index int, exp, act float64) {
		t.Helper()

		if math.Abs(exp-act) > 1e-12 {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	checkNaN := func(index int, act float64) {
		t.Helper()

		if !math.IsNaN(act) {
			t.Errorf("[%v] is incorrect: expected NaN, actual %v", index, act)
		}
	}

	input := testTriangularMovingAverageInput()
	expected := testTriangularMovingAverageExpectedXls12()
	trima := testTriangularMovingAverageCreate(12)

	for i := 0; i < 11; i++ {
		checkNaN(i, trima.Update(input[i]))
	}

	for i := 11; i < len(input); i++ {
		act := trima.Update(input[i])
		check(i, expected[i], act)
	}
}

func TestTriangularMovingAverageUpdateEntity(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		l   = 12
		l1  = l - 1
		inp = 97.250000 // input[l1] = input[11]
		exp = 93.5329761904762
	)

	time := testTriangularMovingAverageTime()
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

		if math.Abs(exp-s.Value) > 1e-12 {
			t.Errorf("value is incorrect: expected %v, actual %v", exp, s.Value)
		}
	}

	input := testTriangularMovingAverageInput()

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		s := entities.Scalar{Time: time, Value: inp}
		trima := testTriangularMovingAverageCreate(l)

		for i := 0; i < l1; i++ {
			trima.Update(input[i])
		}

		check(trima.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: time, Close: inp}
		trima := testTriangularMovingAverageCreate(l)

		for i := 0; i < l1; i++ {
			trima.Update(input[i])
		}

		check(trima.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: time, Bid: inp, Ask: inp}
		trima := testTriangularMovingAverageCreate(l)

		for i := 0; i < l1; i++ {
			trima.Update(input[i])
		}

		check(trima.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: time, Price: inp}
		trima := testTriangularMovingAverageCreate(l)

		for i := 0; i < l1; i++ {
			trima.Update(input[i])
		}

		check(trima.UpdateTrade(&r))
	})
}

func TestTriangularMovingAverageIsPrimed(t *testing.T) { //nolint:dupl
	t.Parallel()

	input := testTriangularMovingAverageInput()
	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	t.Run("length = 9", func(t *testing.T) {
		t.Parallel()
		trima := testTriangularMovingAverageCreate(9)

		check(-1, false, trima.IsPrimed())

		for i := 0; i < 8; i++ {
			trima.Update(input[i])
			check(i, false, trima.IsPrimed())
		}

		for i := 8; i < len(input); i++ {
			trima.Update(input[i])
			check(i, true, trima.IsPrimed())
		}
	})

	t.Run("length = 12", func(t *testing.T) {
		t.Parallel()
		trima := testTriangularMovingAverageCreate(12)

		check(-1, false, trima.IsPrimed())

		for i := 0; i < 11; i++ {
			trima.Update(input[i])
			check(i, false, trima.IsPrimed())
		}

		for i := 11; i < len(input); i++ {
			trima.Update(input[i])
			check(i, true, trima.IsPrimed())
		}
	})
}

func TestTriangularMovingAverageMetadata(t *testing.T) {
	t.Parallel()

	trima := testTriangularMovingAverageCreate(5)
	act := trima.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	check("Identifier", core.TriangularMovingAverage, act.Identifier)
	check("Mnemonic", "trima(5)", act.Mnemonic)
	check("Description", "Triangular moving average trima(5)", act.Description)
	check("len(Outputs)", 1, len(act.Outputs))
	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", "trima(5)", act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", "Triangular moving average trima(5)", act.Outputs[0].Description)
}

func TestNewTriangularMovingAverage(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		bc entities.BarComponent   = entities.BarMedianPrice
		qc entities.QuoteComponent = entities.QuoteMidPrice
		tc entities.TradeComponent = entities.TradePrice

		length = 5
		errlen = "invalid triangular moving average parameters: length should be greater than 1"
		errbc  = "invalid triangular moving average parameters: 9999: unknown bar component"
		errqc  = "invalid triangular moving average parameters: 9999: unknown quote component"
		errtc  = "invalid triangular moving average parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("length > 1, even", func(t *testing.T) {
		t.Parallel()

		const (
			evenLength     = 6
			evenLengthHalf = 3
			factor         = 1. / 12.
		)

		params := TriangularMovingAverageParams{
			Length: evenLength, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		trima, err := NewTriangularMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "trima(6, hl/2)", trima.LineIndicator.Mnemonic)
		check("description", "Triangular moving average trima(6, hl/2)", trima.LineIndicator.Description)
		check("numerator", 0., trima.numerator)
		check("numeratorSub", 0., trima.numeratorSub)
		check("numeratorAdd", 0., trima.numeratorAdd)
		check("len(window)", evenLength, len(trima.window))
		check("windowLength", evenLength, trima.windowLength)
		check("windowLengthHalf", evenLengthHalf-1, trima.windowLengthHalf)
		check("windowCount", 0, trima.windowCount)
		check("factor", factor, trima.factor)
		check("isOdd", false, trima.isOdd)
		check("primed", false, trima.primed)
	})

	t.Run("length > 1, odd", func(t *testing.T) {
		t.Parallel()

		const (
			oddLength     = 5
			oddLengthHalf = 2
			factor        = 1. / 9.
		)

		params := TriangularMovingAverageParams{
			Length: oddLength, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		trima, err := NewTriangularMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "trima(5, hl/2)", trima.LineIndicator.Mnemonic)
		check("description", "Triangular moving average trima(5, hl/2)", trima.LineIndicator.Description)
		check("numerator", 0., trima.numerator)
		check("numeratorSub", 0., trima.numeratorSub)
		check("numeratorAdd", 0., trima.numeratorAdd)
		check("len(window)", oddLength, len(trima.window))
		check("windowLength", oddLength, trima.windowLength)
		check("windowLengthHalf", oddLengthHalf, trima.windowLengthHalf)
		check("windowCount", 0, trima.windowCount)
		check("factor", factor, trima.factor)
		check("isOdd", true, trima.isOdd)
		check("primed", false, trima.primed)
	})

	t.Run("length = 1", func(t *testing.T) {
		t.Parallel()
		params := TriangularMovingAverageParams{
			Length: 1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		trima, err := NewTriangularMovingAverage(&params)
		check("trima == nil", true, trima == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length = 0", func(t *testing.T) {
		t.Parallel()
		params := TriangularMovingAverageParams{
			Length: 0, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		trima, err := NewTriangularMovingAverage(&params)
		check("trima == nil", true, trima == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length < 0", func(t *testing.T) {
		t.Parallel()
		params := TriangularMovingAverageParams{
			Length: -1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		trima, err := NewTriangularMovingAverage(&params)
		check("trima == nil", true, trima == nil)
		check("err", errlen, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()
		params := TriangularMovingAverageParams{
			Length: length, BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		trima, err := NewTriangularMovingAverage(&params)
		check("trima == nil", true, trima == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()
		params := TriangularMovingAverageParams{
			Length: length, BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		trima, err := NewTriangularMovingAverage(&params)
		check("trima == nil", true, trima == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()
		params := TriangularMovingAverageParams{
			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		trima, err := NewTriangularMovingAverage(&params)
		check("trima == nil", true, trima == nil)
		check("err", errtc, err.Error())
	})

	// Zero-value component mnemonic tests.
	// A zero value means "use default, don't show in mnemonic".

	t.Run("all components zero", func(t *testing.T) {
		t.Parallel()
		params := TriangularMovingAverageParams{Length: length}

		trima, err := NewTriangularMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "trima(5)", trima.LineIndicator.Mnemonic)
		check("description", "Triangular moving average trima(5)", trima.LineIndicator.Description)
	})

	t.Run("only bar component set", func(t *testing.T) {
		t.Parallel()
		params := TriangularMovingAverageParams{Length: length, BarComponent: entities.BarMedianPrice}

		trima, err := NewTriangularMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "trima(5, hl/2)", trima.LineIndicator.Mnemonic)
		check("description", "Triangular moving average trima(5, hl/2)", trima.LineIndicator.Description)
	})

	t.Run("only quote component set", func(t *testing.T) {
		t.Parallel()
		params := TriangularMovingAverageParams{Length: length, QuoteComponent: entities.QuoteBidPrice}

		trima, err := NewTriangularMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "trima(5, b)", trima.LineIndicator.Mnemonic)
		check("description", "Triangular moving average trima(5, b)", trima.LineIndicator.Description)
	})

	t.Run("only trade component set", func(t *testing.T) {
		t.Parallel()
		params := TriangularMovingAverageParams{Length: length, TradeComponent: entities.TradeVolume}

		trima, err := NewTriangularMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "trima(5, v)", trima.LineIndicator.Mnemonic)
		check("description", "Triangular moving average trima(5, v)", trima.LineIndicator.Description)
	})

	t.Run("bar and quote components set", func(t *testing.T) {
		t.Parallel()
		params := TriangularMovingAverageParams{
			Length: length, BarComponent: entities.BarOpenPrice, QuoteComponent: entities.QuoteBidPrice,
		}

		trima, err := NewTriangularMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "trima(5, o, b)", trima.LineIndicator.Mnemonic)
		check("description", "Triangular moving average trima(5, o, b)", trima.LineIndicator.Description)
	})

	t.Run("bar and trade components set", func(t *testing.T) {
		t.Parallel()
		params := TriangularMovingAverageParams{
			Length: length, BarComponent: entities.BarHighPrice, TradeComponent: entities.TradeVolume,
		}

		trima, err := NewTriangularMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "trima(5, h, v)", trima.LineIndicator.Mnemonic)
		check("description", "Triangular moving average trima(5, h, v)", trima.LineIndicator.Description)
	})

	t.Run("quote and trade components set", func(t *testing.T) {
		t.Parallel()
		params := TriangularMovingAverageParams{
			Length: length, QuoteComponent: entities.QuoteAskPrice, TradeComponent: entities.TradeVolume,
		}

		trima, err := NewTriangularMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "trima(5, a, v)", trima.LineIndicator.Mnemonic)
		check("description", "Triangular moving average trima(5, a, v)", trima.LineIndicator.Description)
	})
}

func testTriangularMovingAverageCreate(length int) *TriangularMovingAverage {
	params := TriangularMovingAverageParams{
		Length: length,
	}

	trima, _ := NewTriangularMovingAverage(&params)

	return trima
}
