//nolint:testpackage
package rateofchangepercent

//nolint: gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func testRateOfChangePercentTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

func TestRateOfChangePercentUpdate(t *testing.T) { //nolint: funlen
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

	input := testRateOfChangePercentInput()

	t.Run("length = 14", func(t *testing.T) {
		const ( // Values from index=0 to index=13 are NaN.
			i14value  = -0.00546  // Index=14 value.
			i15value  = -0.02109  // Index=15 value.
			i16value  = -0.0553   // Index=16 value.
			i251value = -0.010367 // Index=251 (last) value.
		)

		t.Parallel()
		rocp := testRateOfChangePercentCreate(14)

		for i := 0; i < 13; i++ {
			checkNaN(i, rocp.Update(input[i]))
		}

		for i := 13; i < len(input); i++ {
			act := rocp.Update(input[i])

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

		checkNaN(0, rocp.Update(math.NaN()))
	})
}

func TestRateOfChangePercentUpdateEntity(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		l   = 2
		inp = 3.
		exp = 0.
	)

	time := testRateOfChangePercentTime()
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
		rocp := testRateOfChangePercentCreate(l)
		rocp.Update(inp)
		rocp.Update(inp)
		check(rocp.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: time, Close: inp}
		rocp := testRateOfChangePercentCreate(l)
		rocp.Update(inp)
		rocp.Update(inp)
		check(rocp.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: time, Bid: inp, Ask: inp}
		rocp := testRateOfChangePercentCreate(l)
		rocp.Update(inp)
		rocp.Update(inp)
		check(rocp.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: time, Price: inp}
		rocp := testRateOfChangePercentCreate(l)
		rocp.Update(inp)
		rocp.Update(inp)
		check(rocp.UpdateTrade(&r))
	})
}

func TestRateOfChangePercentIsPrimed(t *testing.T) { //nolint:funlen
	t.Parallel()

	input := testRateOfChangePercentInput()
	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	t.Run("length = 1", func(t *testing.T) {
		t.Parallel()
		rocp := testRateOfChangePercentCreate(1)

		check(-1, false, rocp.IsPrimed())

		for i := 0; i < 1; i++ {
			rocp.Update(input[i])
			check(i, false, rocp.IsPrimed())
		}

		for i := 1; i < len(input); i++ {
			rocp.Update(input[i])
			check(i, true, rocp.IsPrimed())
		}
	})

	t.Run("length = 2", func(t *testing.T) {
		t.Parallel()
		rocp := testRateOfChangePercentCreate(2)

		check(-1, false, rocp.IsPrimed())

		for i := 0; i < 2; i++ {
			rocp.Update(input[i])
			check(i, false, rocp.IsPrimed())
		}

		for i := 2; i < len(input); i++ {
			rocp.Update(input[i])
			check(i, true, rocp.IsPrimed())
		}
	})

	t.Run("length = 5", func(t *testing.T) {
		t.Parallel()
		rocp := testRateOfChangePercentCreate(5)

		check(-1, false, rocp.IsPrimed())

		for i := 0; i < 5; i++ {
			rocp.Update(input[i])
			check(i, false, rocp.IsPrimed())
		}

		for i := 5; i < len(input); i++ {
			rocp.Update(input[i])
			check(i, true, rocp.IsPrimed())
		}
	})

	t.Run("length = 10", func(t *testing.T) {
		t.Parallel()
		rocp := testRateOfChangePercentCreate(10)

		check(-1, false, rocp.IsPrimed())

		for i := 0; i < 10; i++ {
			rocp.Update(input[i])
			check(i, false, rocp.IsPrimed())
		}

		for i := 10; i < len(input); i++ {
			rocp.Update(input[i])
			check(i, true, rocp.IsPrimed())
		}
	})
}

func TestRateOfChangePercentMetadata(t *testing.T) {
	t.Parallel()

	rocp := testRateOfChangePercentCreate(5)
	act := rocp.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	check("Identifier", core.RateOfChangePercent, act.Identifier)
	check("Mnemonic", "rocp(5)", act.Mnemonic)
	check("Description", "Rate of Change Percent rocp(5)", act.Description)
	check("len(Outputs)", 1, len(act.Outputs))
	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", "rocp(5)", act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", "Rate of Change Percent rocp(5)", act.Outputs[0].Description)
}

func TestNewRateOfChangePercent(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		bc     entities.BarComponent   = entities.BarMedianPrice
		qc     entities.QuoteComponent = entities.QuoteMidPrice
		tc     entities.TradeComponent = entities.TradePrice
		length                         = 5
		errlen                         = "invalid rate of change percent parameters: length should be positive"
		errbc                          = "invalid rate of change percent parameters: 9999: unknown bar component"
		errqc                          = "invalid rate of change percent parameters: 9999: unknown quote component"
		errtc                          = "invalid rate of change percent parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("length > 0", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		rocp, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "rocp(5, hl/2)", rocp.LineIndicator.Mnemonic)
		check("description", "Rate of Change Percent rocp(5, hl/2)", rocp.LineIndicator.Description)
		check("primed", false, rocp.primed)
		check("lastIndex", length, rocp.lastIndex)
		check("len(window)", length+1, len(rocp.window))
		check("windowLength", length+1, rocp.windowLength)
		check("windowCount", 0, rocp.windowCount)
	})

	t.Run("length = 1", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: 1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		rocp, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "rocp(1, hl/2)", rocp.LineIndicator.Mnemonic)
		check("description", "Rate of Change Percent rocp(1, hl/2)", rocp.LineIndicator.Description)
		check("primed", false, rocp.primed)
		check("lastIndex", 1, rocp.lastIndex)
		check("len(window)", 2, len(rocp.window))
		check("windowLength", 2, rocp.windowLength)
		check("windowCount", 0, rocp.windowCount)
	})

	t.Run("length = 0", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: 0, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		rocp, err := New(&params)
		check("rocp == nil", true, rocp == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length < 0", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: -1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		rocp, err := New(&params)
		check("rocp == nil", true, rocp == nil)
		check("err", errlen, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		rocp, err := New(&params)
		check("rocp == nil", true, rocp == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		rocp, err := New(&params)
		check("rocp == nil", true, rocp == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		rocp, err := New(&params)
		check("rocp == nil", true, rocp == nil)
		check("err", errtc, err.Error())
	})

	// Zero-value component mnemonic tests.
	// A zero value means "use default, don't show in mnemonic".

	t.Run("all components zero", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length}

		rocp, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "rocp(5)", rocp.LineIndicator.Mnemonic)
		check("description", "Rate of Change Percent rocp(5)", rocp.LineIndicator.Description)
	})

	t.Run("only bar component set", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length, BarComponent: entities.BarMedianPrice}

		rocp, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "rocp(5, hl/2)", rocp.LineIndicator.Mnemonic)
		check("description", "Rate of Change Percent rocp(5, hl/2)", rocp.LineIndicator.Description)
	})

	t.Run("only quote component set", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length, QuoteComponent: entities.QuoteBidPrice}

		rocp, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "rocp(5, b)", rocp.LineIndicator.Mnemonic)
		check("description", "Rate of Change Percent rocp(5, b)", rocp.LineIndicator.Description)
	})

	t.Run("only trade component set", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length, TradeComponent: entities.TradeVolume}

		rocp, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "rocp(5, v)", rocp.LineIndicator.Mnemonic)
		check("description", "Rate of Change Percent rocp(5, v)", rocp.LineIndicator.Description)
	})
}

func testRateOfChangePercentCreate(length int) *RateOfChangePercent {
	params := Params{Length: length}

	rocp, _ := New(&params)

	return rocp
}
