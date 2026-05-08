//nolint:testpackage
package rateofchangeratio

//nolint: gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func testRateOfChangeRatioTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

func TestRateOfChangeRatioUpdate(t *testing.T) { //nolint: funlen
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

	input := testRateOfChangeRatioInput()

	t.Run("ROCR length = 14", func(t *testing.T) {
		const ( // Values from index=0 to index=13 are NaN.
			i14value  = 0.994536 // Index=14 value (output index 0).
			i15value  = 0.978906 // Index=15 value (output index 1).
			i16value  = 0.944689 // Index=16 value (output index 2).
			i251value = 0.989633 // Index=251 (last) value (output index 237).
		)

		t.Parallel()
		rocr := testRateOfChangeRatioCreate(14, false)

		for i := 0; i < 13; i++ {
			checkNaN(i, rocr.Update(input[i]))
		}

		for i := 13; i < len(input); i++ {
			act := rocr.Update(input[i])

			switch i {
			case 14:
				check(i, i14value, act)
			case 15:
				check(i, i15value, act)
			case 16:
				check(i, i16value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, rocr.Update(math.NaN()))
	})

	t.Run("ROCR100 length = 14", func(t *testing.T) {
		const ( // Values from index=0 to index=13 are NaN.
			i14value  = 99.4536 // Index=14 value (output index 0).
			i15value  = 97.8906 // Index=15 value (output index 1).
			i16value  = 94.4689 // Index=16 value (output index 2).
			i251value = 98.9633 // Index=251 (last) value (output index 237).
		)

		t.Parallel()
		rocr100 := testRateOfChangeRatioCreate(14, true)

		for i := 0; i < 13; i++ {
			checkNaN(i, rocr100.Update(input[i]))
		}

		for i := 13; i < len(input); i++ {
			act := rocr100.Update(input[i])

			switch i {
			case 14:
				check(i, i14value, act)
			case 15:
				check(i, i15value, act)
			case 16:
				check(i, i16value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, rocr100.Update(math.NaN()))
	})

	t.Run("ROCR middle of data", func(t *testing.T) {
		// { 0, TA_ROCR_TEST, 20, 21, 14, TA_SUCCESS, 0, 0.955096, 20, 2 },
		// { 0, TA_ROCR_TEST, 20, 21, 14, TA_SUCCESS, 1, 0.944744, 20, 2 },
		// Output index 0 corresponds to input index 20, output index 1 to input index 21.
		t.Parallel()
		rocr := testRateOfChangeRatioCreate(14, false)

		for i := 0; i < len(input); i++ {
			act := rocr.Update(input[i])

			switch i {
			case 20:
				check(i, 0.955096, act)
			case 21:
				check(i, 0.944744, act)
			}
		}
	})

	t.Run("ROCR100 middle of data", func(t *testing.T) {
		// { 0, TA_ROCR100_TEST, 20, 21, 14, TA_SUCCESS, 0, 95.5096, 20, 2 },
		// { 0, TA_ROCR100_TEST, 20, 21, 14, TA_SUCCESS, 1, 94.4744, 20, 2 },
		t.Parallel()
		rocr100 := testRateOfChangeRatioCreate(14, true)

		for i := 0; i < len(input); i++ {
			act := rocr100.Update(input[i])

			switch i {
			case 20:
				check(i, 95.5096, act)
			case 21:
				check(i, 94.4744, act)
			}
		}
	})
}

func TestRateOfChangeRatioUpdateEntity(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		l   = 2
		inp = 3.
		exp = 1.
	)

	time := testRateOfChangeRatioTime()
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

		if math.Abs(s.Value-exp) > 1e-13 {
			t.Errorf("value is incorrect: expected %v, actual %v", exp, s.Value)
		}
	}

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		s := entities.Scalar{Time: time, Value: inp}
		rocr := testRateOfChangeRatioCreate(l, false)
		rocr.Update(inp)
		rocr.Update(inp)
		check(rocr.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: time, Close: inp}
		rocr := testRateOfChangeRatioCreate(l, false)
		rocr.Update(inp)
		rocr.Update(inp)
		check(rocr.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: time, Bid: inp, Ask: inp}
		rocr := testRateOfChangeRatioCreate(l, false)
		rocr.Update(inp)
		rocr.Update(inp)
		check(rocr.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: time, Price: inp}
		rocr := testRateOfChangeRatioCreate(l, false)
		rocr.Update(inp)
		rocr.Update(inp)
		check(rocr.UpdateTrade(&r))
	})
}

func TestRateOfChangeRatioIsPrimed(t *testing.T) { //nolint:funlen
	t.Parallel()

	input := testRateOfChangeRatioInput()
	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	t.Run("length = 1", func(t *testing.T) {
		t.Parallel()
		rocr := testRateOfChangeRatioCreate(1, false)

		check(-1, false, rocr.IsPrimed())

		for i := 0; i < 1; i++ {
			rocr.Update(input[i])
			check(i, false, rocr.IsPrimed())
		}

		for i := 1; i < len(input); i++ {
			rocr.Update(input[i])
			check(i, true, rocr.IsPrimed())
		}
	})

	t.Run("length = 2", func(t *testing.T) {
		t.Parallel()
		rocr := testRateOfChangeRatioCreate(2, false)

		check(-1, false, rocr.IsPrimed())

		for i := 0; i < 2; i++ {
			rocr.Update(input[i])
			check(i, false, rocr.IsPrimed())
		}

		for i := 2; i < len(input); i++ {
			rocr.Update(input[i])
			check(i, true, rocr.IsPrimed())
		}
	})

	t.Run("length = 5", func(t *testing.T) {
		t.Parallel()
		rocr := testRateOfChangeRatioCreate(5, false)

		check(-1, false, rocr.IsPrimed())

		for i := 0; i < 5; i++ {
			rocr.Update(input[i])
			check(i, false, rocr.IsPrimed())
		}

		for i := 5; i < len(input); i++ {
			rocr.Update(input[i])
			check(i, true, rocr.IsPrimed())
		}
	})

	t.Run("length = 10", func(t *testing.T) {
		t.Parallel()
		rocr := testRateOfChangeRatioCreate(10, false)

		check(-1, false, rocr.IsPrimed())

		for i := 0; i < 10; i++ {
			rocr.Update(input[i])
			check(i, false, rocr.IsPrimed())
		}

		for i := 10; i < len(input); i++ {
			rocr.Update(input[i])
			check(i, true, rocr.IsPrimed())
		}
	})
}

func TestRateOfChangeRatioMetadata(t *testing.T) {
	t.Parallel()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	t.Run("ROCR", func(t *testing.T) {
		t.Parallel()
		rocr := testRateOfChangeRatioCreate(5, false)
		act := rocr.Metadata()

		check("Identifier", core.RateOfChangeRatio, act.Identifier)
		check("Mnemonic", "rocr(5)", act.Mnemonic)
		check("Description", "Rate of Change Ratio rocr(5)", act.Description)
		check("len(Outputs)", 1, len(act.Outputs))
		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
		check("Outputs[0].Mnemonic", "rocr(5)", act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", "Rate of Change Ratio rocr(5)", act.Outputs[0].Description)
	})

	t.Run("ROCR100", func(t *testing.T) {
		t.Parallel()
		rocr100 := testRateOfChangeRatioCreate(5, true)
		act := rocr100.Metadata()

		check("Identifier", core.RateOfChangeRatio, act.Identifier)
		check("Mnemonic", "rocr100(5)", act.Mnemonic)
		check("Description", "Rate of Change Ratio 100 Scale rocr100(5)", act.Description)
		check("len(Outputs)", 1, len(act.Outputs))
		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
		check("Outputs[0].Mnemonic", "rocr100(5)", act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", "Rate of Change Ratio 100 Scale rocr100(5)", act.Outputs[0].Description)
	})
}

func TestNewRateOfChangeRatio(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		bc     entities.BarComponent   = entities.BarMedianPrice
		qc     entities.QuoteComponent = entities.QuoteMidPrice
		tc     entities.TradeComponent = entities.TradePrice
		length                         = 5
		errlen                         = "invalid rate of change ratio parameters: length should be positive"
		errbc                          = "invalid rate of change ratio parameters: 9999: unknown bar component"
		errqc                          = "invalid rate of change ratio parameters: 9999: unknown quote component"
		errtc                          = "invalid rate of change ratio parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("ROCR length > 0", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		rocr, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "rocr(5, hl/2)", rocr.LineIndicator.Mnemonic)
		check("description", "Rate of Change Ratio rocr(5, hl/2)", rocr.LineIndicator.Description)
		check("primed", false, rocr.primed)
		check("hundredScale", false, rocr.hundredScale)
		check("lastIndex", length, rocr.lastIndex)
		check("len(window)", length+1, len(rocr.window))
		check("windowLength", length+1, rocr.windowLength)
		check("windowCount", 0, rocr.windowCount)
	})

	t.Run("ROCR100 length > 0", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, HundredScale: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		rocr, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "rocr100(5, hl/2)", rocr.LineIndicator.Mnemonic)
		check("description", "Rate of Change Ratio 100 Scale rocr100(5, hl/2)", rocr.LineIndicator.Description)
		check("primed", false, rocr.primed)
		check("hundredScale", true, rocr.hundredScale)
	})

	t.Run("length = 1", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: 1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		rocr, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "rocr(1, hl/2)", rocr.LineIndicator.Mnemonic)
		check("description", "Rate of Change Ratio rocr(1, hl/2)", rocr.LineIndicator.Description)
		check("primed", false, rocr.primed)
		check("lastIndex", 1, rocr.lastIndex)
		check("len(window)", 2, len(rocr.window))
		check("windowLength", 2, rocr.windowLength)
		check("windowCount", 0, rocr.windowCount)
	})

	t.Run("length = 0", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: 0, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		rocr, err := New(&params)
		check("rocr == nil", true, rocr == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length < 0", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: -1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		rocr, err := New(&params)
		check("rocr == nil", true, rocr == nil)
		check("err", errlen, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		rocr, err := New(&params)
		check("rocr == nil", true, rocr == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		rocr, err := New(&params)
		check("rocr == nil", true, rocr == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		rocr, err := New(&params)
		check("rocr == nil", true, rocr == nil)
		check("err", errtc, err.Error())
	})

	// Zero-value component mnemonic tests.
	t.Run("all components zero", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length}

		rocr, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "rocr(5)", rocr.LineIndicator.Mnemonic)
		check("description", "Rate of Change Ratio rocr(5)", rocr.LineIndicator.Description)
	})

	t.Run("only bar component set", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length, BarComponent: entities.BarMedianPrice}

		rocr, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "rocr(5, hl/2)", rocr.LineIndicator.Mnemonic)
		check("description", "Rate of Change Ratio rocr(5, hl/2)", rocr.LineIndicator.Description)
	})

	t.Run("only quote component set", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length, QuoteComponent: entities.QuoteBidPrice}

		rocr, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "rocr(5, b)", rocr.LineIndicator.Mnemonic)
		check("description", "Rate of Change Ratio rocr(5, b)", rocr.LineIndicator.Description)
	})

	t.Run("only trade component set", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length, TradeComponent: entities.TradeVolume}

		rocr, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "rocr(5, v)", rocr.LineIndicator.Mnemonic)
		check("description", "Rate of Change Ratio rocr(5, v)", rocr.LineIndicator.Description)
	})
}

func testRateOfChangeRatioCreate(length int, hundredScale bool) *RateOfChangeRatio {
	params := Params{Length: length, HundredScale: hundredScale}

	rocr, _ := New(&params)

	return rocr
}
