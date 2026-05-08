//nolint:testpackage
package simplemovingaverage

//nolint: gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func testSimpleMovingAverageTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

func TestSimpleMovingAverageUpdate(t *testing.T) { //nolint: funlen
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

	input := testSimpleMovingAverageInput()

	t.Run("length = 3", func(t *testing.T) {
		t.Parallel()
		sma := testSimpleMovingAverageCreate(3)
		expected := testSimpleMovingAverageExpected3()

		for i := 0; i < 2; i++ {
			checkNaN(i, sma.Update(input[i]))
		}

		for i := 2; i < len(input); i++ {
			exp := expected[i]
			act := sma.Update(input[i])
			check(i, exp, act)
		}

		checkNaN(0, sma.Update(math.NaN()))
	})

	t.Run("length = 5", func(t *testing.T) {
		t.Parallel()
		sma := testSimpleMovingAverageCreate(5)
		expected := testSimpleMovingAverageExpected5()

		for i := 0; i < 4; i++ {
			checkNaN(i, sma.Update(input[i]))
		}

		for i := 4; i < len(input); i++ {
			exp := expected[i]
			act := sma.Update(input[i])
			check(i, exp, act)
		}

		checkNaN(0, sma.Update(math.NaN()))
	})

	t.Run("length = 10", func(t *testing.T) {
		t.Parallel()
		sma := testSimpleMovingAverageCreate(10)
		expected := testSimpleMovingAverageExpected10()

		for i := 0; i < 9; i++ {
			checkNaN(i, sma.Update(input[i]))
		}

		for i := 9; i < len(input); i++ {
			exp := expected[i]
			act := sma.Update(input[i])
			check(i, exp, act)
		}

		checkNaN(0, sma.Update(math.NaN()))
	})
}

func TestSimpleMovingAverageUpdateEntity(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		l   = 2
		inp = 3.
		exp = inp / float64(l)
	)

	time := testSimpleMovingAverageTime()
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
		sma := testSimpleMovingAverageCreate(l)
		sma.Update(0.)
		check(sma.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: time, Close: inp}
		sma := testSimpleMovingAverageCreate(l)
		sma.Update(0.)
		check(sma.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: time, Bid: inp, Ask: inp}
		sma := testSimpleMovingAverageCreate(l)
		sma.Update(0.)
		check(sma.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: time, Price: inp}
		sma := testSimpleMovingAverageCreate(l)
		sma.Update(0.)
		check(sma.UpdateTrade(&r))
	})
}

func TestSimpleMovingAverageIsPrimed(t *testing.T) { //nolint:funlen
	t.Parallel()

	input := testSimpleMovingAverageInput()
	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	t.Run("length = 3", func(t *testing.T) {
		t.Parallel()
		sma := testSimpleMovingAverageCreate(3)

		check(-1, false, sma.IsPrimed())

		for i := 0; i < 2; i++ {
			sma.Update(input[i])
			check(i, false, sma.IsPrimed())
		}

		for i := 2; i < len(input); i++ {
			sma.Update(input[i])
			check(i, true, sma.IsPrimed())
		}
	})

	t.Run("length = 5", func(t *testing.T) {
		t.Parallel()
		sma := testSimpleMovingAverageCreate(5)

		check(-1, false, sma.IsPrimed())

		for i := 0; i < 4; i++ {
			sma.Update(input[i])
			check(i, false, sma.IsPrimed())
		}

		for i := 4; i < len(input); i++ {
			sma.Update(input[i])
			check(i, true, sma.IsPrimed())
		}
	})

	t.Run("length = 10", func(t *testing.T) {
		t.Parallel()
		sma := testSimpleMovingAverageCreate(10)

		check(-1, false, sma.IsPrimed())

		for i := 0; i < 9; i++ {
			sma.Update(input[i])
			check(i, false, sma.IsPrimed())
		}

		for i := 9; i < len(input); i++ {
			sma.Update(input[i])
			check(i, true, sma.IsPrimed())
		}
	})
}

func TestSimpleMovingAverageMetadata(t *testing.T) {
	t.Parallel()

	sma := testSimpleMovingAverageCreate(5)
	act := sma.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	check("Identifier", core.SimpleMovingAverage, act.Identifier)
	check("len(Outputs)", 1, len(act.Outputs))
	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", "sma(5)", act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", "Simple moving average sma(5)", act.Outputs[0].Description)
}

func TestNewSimpleMovingAverage(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		bc entities.BarComponent   = entities.BarMedianPrice
		qc entities.QuoteComponent = entities.QuoteMidPrice
		tc entities.TradeComponent = entities.TradePrice

		length = 5
		errlen = "invalid simple moving average parameters: length should be greater than 1"
		errbc  = "invalid simple moving average parameters: 9999: unknown bar component"
		errqc  = "invalid simple moving average parameters: 9999: unknown quote component"
		errtc  = "invalid simple moving average parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("length > 1", func(t *testing.T) {
		t.Parallel()
		params := SimpleMovingAverageParams{
			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		sma, err := NewSimpleMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "sma(5, hl/2)", sma.LineIndicator.Mnemonic)
		check("description", "Simple moving average sma(5, hl/2)", sma.LineIndicator.Description)
		check("primed", false, sma.primed)
		check("lastIndex", length-1, sma.lastIndex)
		check("len(window)", length, len(sma.window))
		check("windowLength", length, sma.windowLength)
		check("windowCount", 0, sma.windowCount)
		check("windowSum", 0., sma.windowSum)
	})

	t.Run("length = 1", func(t *testing.T) {
		t.Parallel()
		params := SimpleMovingAverageParams{
			Length: 1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		sma, err := NewSimpleMovingAverage(&params)
		check("sma == nil", true, sma == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length = 0", func(t *testing.T) {
		t.Parallel()
		params := SimpleMovingAverageParams{
			Length: 0, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		sma, err := NewSimpleMovingAverage(&params)
		check("sma == nil", true, sma == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length < 0", func(t *testing.T) {
		t.Parallel()
		params := SimpleMovingAverageParams{
			Length: -1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		sma, err := NewSimpleMovingAverage(&params)
		check("sma == nil", true, sma == nil)
		check("err", errlen, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()
		params := SimpleMovingAverageParams{
			Length: length, BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		sma, err := NewSimpleMovingAverage(&params)
		check("sma == nil", true, sma == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()
		params := SimpleMovingAverageParams{
			Length: length, BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		sma, err := NewSimpleMovingAverage(&params)
		check("sma == nil", true, sma == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()
		params := SimpleMovingAverageParams{
			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		sma, err := NewSimpleMovingAverage(&params)
		check("sma == nil", true, sma == nil)
		check("err", errtc, err.Error())
	})

	// Zero-value component mnemonic tests.
	// A zero value means "use default, don't show in mnemonic".

	t.Run("all components zero", func(t *testing.T) {
		t.Parallel()
		params := SimpleMovingAverageParams{Length: length}

		sma, err := NewSimpleMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "sma(5)", sma.LineIndicator.Mnemonic)
		check("description", "Simple moving average sma(5)", sma.LineIndicator.Description)
	})

	t.Run("only bar component set", func(t *testing.T) {
		t.Parallel()
		params := SimpleMovingAverageParams{Length: length, BarComponent: entities.BarMedianPrice}

		sma, err := NewSimpleMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "sma(5, hl/2)", sma.LineIndicator.Mnemonic)
		check("description", "Simple moving average sma(5, hl/2)", sma.LineIndicator.Description)
	})

	t.Run("only quote component set", func(t *testing.T) {
		t.Parallel()
		params := SimpleMovingAverageParams{Length: length, QuoteComponent: entities.QuoteBidPrice}

		sma, err := NewSimpleMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "sma(5, b)", sma.LineIndicator.Mnemonic)
		check("description", "Simple moving average sma(5, b)", sma.LineIndicator.Description)
	})

	t.Run("only trade component set", func(t *testing.T) {
		t.Parallel()
		params := SimpleMovingAverageParams{Length: length, TradeComponent: entities.TradeVolume}

		sma, err := NewSimpleMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "sma(5, v)", sma.LineIndicator.Mnemonic)
		check("description", "Simple moving average sma(5, v)", sma.LineIndicator.Description)
	})

	t.Run("bar and quote components set", func(t *testing.T) {
		t.Parallel()
		params := SimpleMovingAverageParams{
			Length: length, BarComponent: entities.BarOpenPrice, QuoteComponent: entities.QuoteBidPrice,
		}

		sma, err := NewSimpleMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "sma(5, o, b)", sma.LineIndicator.Mnemonic)
		check("description", "Simple moving average sma(5, o, b)", sma.LineIndicator.Description)
	})

	t.Run("bar and trade components set", func(t *testing.T) {
		t.Parallel()
		params := SimpleMovingAverageParams{
			Length: length, BarComponent: entities.BarHighPrice, TradeComponent: entities.TradeVolume,
		}

		sma, err := NewSimpleMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "sma(5, h, v)", sma.LineIndicator.Mnemonic)
		check("description", "Simple moving average sma(5, h, v)", sma.LineIndicator.Description)
	})

	t.Run("quote and trade components set", func(t *testing.T) {
		t.Parallel()
		params := SimpleMovingAverageParams{
			Length: length, QuoteComponent: entities.QuoteAskPrice, TradeComponent: entities.TradeVolume,
		}

		sma, err := NewSimpleMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "sma(5, a, v)", sma.LineIndicator.Mnemonic)
		check("description", "Simple moving average sma(5, a, v)", sma.LineIndicator.Description)
	})
}

func testSimpleMovingAverageCreate(length int) *SimpleMovingAverage {
	params := SimpleMovingAverageParams{
		Length: length,
	}

	sma, _ := NewSimpleMovingAverage(&params)

	return sma
}
