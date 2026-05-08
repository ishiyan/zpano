//nolint:testpackage
package pearsonscorrelationcoefficient

//nolint: gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func testTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

func TestPearsonsCorrelationCoefficientUpdatePair(t *testing.T) { //nolint: funlen
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

	high := testHighInput()
	low := testLowInput()

	t.Run("TaLib spot checks period=20", func(t *testing.T) {
		t.Parallel()
		c := testCreate(20)

		for i := 0; i < 18; i++ {
			checkNaN(i, c.UpdatePair(high[i], low[i]))
		}

		for i := 18; i < len(high); i++ {
			act := c.UpdatePair(high[i], low[i])

			// Output starts at input index 19 (lookback = 19).
			// Output index = input index - 19.
			switch i {
			case 19: // Output index 0.
				check(i, 0.9401569, act)
			case 20: // Output index 1.
				check(i, 0.9471812, act)
			case 251: // Output index 232 (252-20 = last).
				check(i, 0.8866901, act)
			}
		}

		checkNaN(0, c.UpdatePair(math.NaN(), 1.0))
		checkNaN(0, c.UpdatePair(1.0, math.NaN()))
	})

	t.Run("Excel verification period=20", func(t *testing.T) {
		t.Parallel()
		c := testCreate(20)
		expected := testExcelExpected()

		const eps = 1e-10

		for i := 0; i < 18; i++ {
			checkNaN(i, c.UpdatePair(high[i], low[i]))
		}

		for i := 18; i < len(high); i++ {
			act := c.UpdatePair(high[i], low[i])

			if i >= 19 {
				if math.Abs(expected[i]-act) > eps {
					t.Errorf("input %v, expected %.16f, actual %.16f", i, expected[i], act)
				}
			}
		}
	})
}

func TestPearsonsCorrelationCoefficientUpdateEntity(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		l   = 2
		inp = 3.
	)

	tm := testTime()
	check := func(act core.Output) {
		t.Helper()

		if len(act) != 1 {
			t.Errorf("len(output) is incorrect: expected 1, actual %v", len(act))
		}

		s, ok := act[0].(entities.Scalar)
		if !ok {
			t.Error("output is not scalar")
		}

		if s.Time != tm {
			t.Errorf("time is incorrect: expected %v, actual %v", tm, s.Time)
		}

		// correl(x,x) with constant value returns 0 (zero variance).
		if math.Abs(s.Value-0.0) > 1e-10 {
			t.Errorf("value is incorrect: expected 0.0, actual %v", s.Value)
		}
	}

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		s := entities.Scalar{Time: tm, Value: inp}
		c := testCreate(l)
		c.Update(inp)
		c.Update(inp)
		check(c.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: tm, High: 10, Low: 5}
		c := testCreate(l)
		c.UpdatePair(10, 5)
		c.UpdatePair(20, 10)
		out := c.UpdateBar(&b)

		if len(out) != 1 {
			t.Errorf("len(output) is incorrect: expected 1, actual %v", len(out))
		}

		s, ok := out[0].(entities.Scalar)
		if !ok {
			t.Error("output is not scalar")
		}

		if s.Time != tm {
			t.Errorf("time is incorrect: expected %v, actual %v", tm, s.Time)
		}

		if math.IsNaN(s.Value) {
			t.Error("value should not be NaN")
		}
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: tm, Bid: inp, Ask: inp}
		c := testCreate(l)
		c.Update(inp)
		c.Update(inp)
		check(c.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: tm, Price: inp}
		c := testCreate(l)
		c.Update(inp)
		c.Update(inp)
		check(c.UpdateTrade(&r))
	})
}

func TestPearsonsCorrelationCoefficientIsPrimed(t *testing.T) { //nolint:funlen
	t.Parallel()

	high := testHighInput()
	low := testLowInput()
	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	t.Run("length = 1", func(t *testing.T) {
		t.Parallel()
		c := testCreate(1)

		check(-1, false, c.IsPrimed())

		c.UpdatePair(high[0], low[0])
		check(0, true, c.IsPrimed())
	})

	t.Run("length = 2", func(t *testing.T) {
		t.Parallel()
		c := testCreate(2)

		check(-1, false, c.IsPrimed())

		c.UpdatePair(high[0], low[0])
		check(0, false, c.IsPrimed())

		c.UpdatePair(high[1], low[1])
		check(1, true, c.IsPrimed())
	})

	t.Run("length = 20", func(t *testing.T) {
		t.Parallel()
		c := testCreate(20)

		check(-1, false, c.IsPrimed())

		for i := 0; i < 19; i++ {
			c.UpdatePair(high[i], low[i])
			check(i, false, c.IsPrimed())
		}

		c.UpdatePair(high[19], low[19])
		check(19, true, c.IsPrimed())
	})
}

func TestPearsonsCorrelationCoefficientMetadata(t *testing.T) {
	t.Parallel()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	c := testCreate(20)
	act := c.Metadata()

	check("Identifier", core.PearsonsCorrelationCoefficient, act.Identifier)
	check("Mnemonic", "correl(20)", act.Mnemonic)
	check("Description", "Pearsons Correlation Coefficient correl(20)", act.Description)
	check("len(Outputs)", 1, len(act.Outputs))
	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", "correl(20)", act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", "Pearsons Correlation Coefficient correl(20)", act.Outputs[0].Description)
}

func TestNewPearsonsCorrelationCoefficient(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		bc     entities.BarComponent   = entities.BarMedianPrice
		qc     entities.QuoteComponent = entities.QuoteMidPrice
		tc     entities.TradeComponent = entities.TradePrice
		length                         = 20
		errlen                         = "invalid pearsons correlation coefficient parameters: length should be positive"
		errbc                          = "invalid pearsons correlation coefficient parameters: 9999: unknown bar component"
		errqc                          = "invalid pearsons correlation coefficient parameters: 9999: unknown quote component"
		errtc                          = "invalid pearsons correlation coefficient parameters: 9999: unknown trade component"
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

		c, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "correl(20, hl/2)", c.LineIndicator.Mnemonic)
		check("description", "Pearsons Correlation Coefficient correl(20, hl/2)", c.LineIndicator.Description)
		check("primed", false, c.primed)
		check("length", length, c.length)
		check("len(windowX)", length, len(c.windowX))
		check("len(windowY)", length, len(c.windowY))
		check("count", 0, c.count)
	})

	t.Run("length = 1", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: 1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		c, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "correl(1, hl/2)", c.LineIndicator.Mnemonic)
		check("primed", false, c.primed)
		check("length", 1, c.length)
	})

	t.Run("length = 0", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: 0, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		c, err := New(&params)
		check("c == nil", true, c == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length < 0", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: -1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		c, err := New(&params)
		check("c == nil", true, c == nil)
		check("err", errlen, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		c, err := New(&params)
		check("c == nil", true, c == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		c, err := New(&params)
		check("c == nil", true, c == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		c, err := New(&params)
		check("c == nil", true, c == nil)
		check("err", errtc, err.Error())
	})

	t.Run("all components zero", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length}

		c, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "correl(20)", c.LineIndicator.Mnemonic)
		check("description", "Pearsons Correlation Coefficient correl(20)", c.LineIndicator.Description)
	})
}

func testCreate(length int) *PearsonsCorrelationCoefficient {
	params := Params{Length: length}

	c, _ := New(&params)

	return c
}
