//nolint:testpackage
package weightedmovingaverage

//nolint: gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func testWeightedMovingAverageTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

func TestWeightedMovingAverageUpdate(t *testing.T) { //nolint: funlen
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

	input := testWeightedMovingAverageInput()

	t.Run("length = 2", func(t *testing.T) {
		const (
			i1value   = 93.71  // Index=1 value.
			i2value   = 94.52  // Index=2 value.
			i3value   = 94.855 // Index=3 value.
			i251value = 108.16 // Index=251 (last) value.
		)

		t.Parallel()
		wma := testWeightedMovingAverageCreate(2)

		for i := 0; i < len(input); i++ {
			act := wma.Update(input[i])

			switch i {
			case 0:
				checkNaN(i, act)
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

		checkNaN(0, wma.Update(math.NaN()))
	})

	t.Run("length = 30", func(t *testing.T) {
		const ( // The first 29 values are NaN.
			i29value  = 88.5677  // Index=29 value.
			i30value  = 88.2337  // Index=30 value.
			i31value  = 88.034   // Index=31 value.
			i58value  = 87.191   // Index=58 value.
			i250value = 109.3466 // Index=250 value.
			i251value = 109.3413 // Index=251 (last) value.
		)

		t.Parallel()
		wma := testWeightedMovingAverageCreate(30)

		for i := 0; i < 29; i++ {
			checkNaN(i, wma.Update(input[i]))
		}

		for i := 29; i < len(input); i++ {
			act := wma.Update(input[i])

			switch i {
			case 29:
				check(i, i29value, act)
			case 30:
				check(i, i30value, act)
			case 31:
				check(i, i31value, act)
			case 58:
				check(i, i58value, act)
			case 250:
				check(i, i250value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, wma.Update(math.NaN()))
	})
}

func TestWeightedMovingAverageUpdateEntity(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		l   = 2
		exp = 93.71
	)

	input := testWeightedMovingAverageInput()
	time := testWeightedMovingAverageTime()
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

		s := entities.Scalar{Time: time, Value: input[1]}
		wma := testWeightedMovingAverageCreate(l)
		wma.Update(input[0])
		check(wma.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: time, Close: input[1]}
		wma := testWeightedMovingAverageCreate(l)
		wma.Update(input[0])
		check(wma.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: time, Bid: input[1], Ask: input[1]}
		wma := testWeightedMovingAverageCreate(l)
		wma.Update(input[0])
		check(wma.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: time, Price: input[1]}
		wma := testWeightedMovingAverageCreate(l)
		wma.Update(input[0])
		check(wma.UpdateTrade(&r))
	})
}

func TestWeightedMovingAverageIsPrimed(t *testing.T) { //nolint:dupl
	t.Parallel()

	input := testWeightedMovingAverageInput()
	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	t.Run("length = 2", func(t *testing.T) {
		t.Parallel()
		wma := testWeightedMovingAverageCreate(2)

		check(-1, false, wma.IsPrimed())

		for i := 0; i < 1; i++ {
			wma.Update(input[i])
			check(i, false, wma.IsPrimed())
		}

		for i := 1; i < len(input); i++ {
			wma.Update(input[i])
			check(i, true, wma.IsPrimed())
		}
	})

	t.Run("length = 30", func(t *testing.T) {
		t.Parallel()
		wma := testWeightedMovingAverageCreate(30)

		check(-1, false, wma.IsPrimed())

		for i := 0; i < 29; i++ {
			wma.Update(input[i])
			check(i, false, wma.IsPrimed())
		}

		for i := 29; i < len(input); i++ {
			wma.Update(input[i])
			check(i, true, wma.IsPrimed())
		}
	})
}

func TestWeightedMovingAverageMetadata(t *testing.T) {
	t.Parallel()

	wma := testWeightedMovingAverageCreate(5)
	act := wma.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	check("Identifier", core.WeightedMovingAverage, act.Identifier)
	check("Mnemonic", "wma(5)", act.Mnemonic)
	check("Description", "Weighted moving average wma(5)", act.Description)
	check("len(Outputs)", 1, len(act.Outputs))
	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", "wma(5)", act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", "Weighted moving average wma(5)", act.Outputs[0].Description)
}

func TestNewWeightedMovingAverage(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		bc entities.BarComponent   = entities.BarMedianPrice
		qc entities.QuoteComponent = entities.QuoteMidPrice
		tc entities.TradeComponent = entities.TradePrice

		length  = 5
		divider = float64(length) * float64(length+1) / 2.
		errlen  = "invalid weighted moving average parameters: length should be greater than 1"
		errbc   = "invalid weighted moving average parameters: 9999: unknown bar component"
		errqc   = "invalid weighted moving average parameters: 9999: unknown quote component"
		errtc   = "invalid weighted moving average parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("length > 1", func(t *testing.T) {
		t.Parallel()
		params := WeightedMovingAverageParams{
			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		wma, err := NewWeightedMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "wma(5, hl/2)", wma.LineIndicator.Mnemonic)
		check("description", "Weighted moving average wma(5, hl/2)", wma.LineIndicator.Description)
		check("primed", false, wma.primed)
		check("lastIndex", length-1, wma.lastIndex)
		check("len(window)", length, len(wma.window))
		check("windowLength", length, wma.windowLength)
		check("divider", divider, wma.divider)
		check("windowCount", 0, wma.windowCount)
		check("windowSum", 0., wma.windowSum)
		check("windowSub", 0., wma.windowSub)
	})

	t.Run("length = 1", func(t *testing.T) {
		t.Parallel()
		params := WeightedMovingAverageParams{
			Length: 1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		wma, err := NewWeightedMovingAverage(&params)
		check("wma == nil", true, wma == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length = 0", func(t *testing.T) {
		t.Parallel()
		params := WeightedMovingAverageParams{
			Length: 0, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		wma, err := NewWeightedMovingAverage(&params)
		check("wma == nil", true, wma == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length < 0", func(t *testing.T) {
		t.Parallel()
		params := WeightedMovingAverageParams{
			Length: -1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		wma, err := NewWeightedMovingAverage(&params)
		check("wma == nil", true, wma == nil)
		check("err", errlen, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()
		params := WeightedMovingAverageParams{
			Length: length, BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		wma, err := NewWeightedMovingAverage(&params)
		check("wma == nil", true, wma == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()
		params := WeightedMovingAverageParams{
			Length: length, BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		wma, err := NewWeightedMovingAverage(&params)
		check("wma == nil", true, wma == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()
		params := WeightedMovingAverageParams{
			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		wma, err := NewWeightedMovingAverage(&params)
		check("wma == nil", true, wma == nil)
		check("err", errtc, err.Error())
	})

	// Zero-value component mnemonic tests.
	// A zero value means "use default, don't show in mnemonic".

	t.Run("all components zero", func(t *testing.T) {
		t.Parallel()
		params := WeightedMovingAverageParams{Length: length}

		wma, err := NewWeightedMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "wma(5)", wma.LineIndicator.Mnemonic)
		check("description", "Weighted moving average wma(5)", wma.LineIndicator.Description)
	})

	t.Run("only bar component set", func(t *testing.T) {
		t.Parallel()
		params := WeightedMovingAverageParams{Length: length, BarComponent: entities.BarMedianPrice}

		wma, err := NewWeightedMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "wma(5, hl/2)", wma.LineIndicator.Mnemonic)
		check("description", "Weighted moving average wma(5, hl/2)", wma.LineIndicator.Description)
	})

	t.Run("only quote component set", func(t *testing.T) {
		t.Parallel()
		params := WeightedMovingAverageParams{Length: length, QuoteComponent: entities.QuoteBidPrice}

		wma, err := NewWeightedMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "wma(5, b)", wma.LineIndicator.Mnemonic)
		check("description", "Weighted moving average wma(5, b)", wma.LineIndicator.Description)
	})

	t.Run("only trade component set", func(t *testing.T) {
		t.Parallel()
		params := WeightedMovingAverageParams{Length: length, TradeComponent: entities.TradeVolume}

		wma, err := NewWeightedMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "wma(5, v)", wma.LineIndicator.Mnemonic)
		check("description", "Weighted moving average wma(5, v)", wma.LineIndicator.Description)
	})

	t.Run("bar and quote components set", func(t *testing.T) {
		t.Parallel()
		params := WeightedMovingAverageParams{
			Length: length, BarComponent: entities.BarOpenPrice, QuoteComponent: entities.QuoteBidPrice,
		}

		wma, err := NewWeightedMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "wma(5, o, b)", wma.LineIndicator.Mnemonic)
		check("description", "Weighted moving average wma(5, o, b)", wma.LineIndicator.Description)
	})

	t.Run("bar and trade components set", func(t *testing.T) {
		t.Parallel()
		params := WeightedMovingAverageParams{
			Length: length, BarComponent: entities.BarHighPrice, TradeComponent: entities.TradeVolume,
		}

		wma, err := NewWeightedMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "wma(5, h, v)", wma.LineIndicator.Mnemonic)
		check("description", "Weighted moving average wma(5, h, v)", wma.LineIndicator.Description)
	})

	t.Run("quote and trade components set", func(t *testing.T) {
		t.Parallel()
		params := WeightedMovingAverageParams{
			Length: length, QuoteComponent: entities.QuoteAskPrice, TradeComponent: entities.TradeVolume,
		}

		wma, err := NewWeightedMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "wma(5, a, v)", wma.LineIndicator.Mnemonic)
		check("description", "Weighted moving average wma(5, a, v)", wma.LineIndicator.Description)
	})
}

func testWeightedMovingAverageCreate(length int) *WeightedMovingAverage {
	params := WeightedMovingAverageParams{
		Length: length,
	}

	wma, _ := NewWeightedMovingAverage(&params)

	return wma
}
