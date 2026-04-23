//nolint:testpackage
package variance

//nolint: gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func testVarianceTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

// testVarianceInput is variance input test data.
func testVarianceInput() []float64 { return []float64{1, 2, 8, 4, 9, 6, 7, 13, 9, 10, 3, 12} }

// testVarianceExpectedLength3Population is the Excel (VAR.P) output of population variance of length 3.
func testVarianceExpectedLength3Population() []float64 {
	return []float64{
		math.NaN(), math.NaN(),
		9.55555555555556000, 6.22222222222222000, 4.66666666666667000, 4.22222222222222000, 1.55555555555556000,
		9.55555555555556000, 6.22222222222222000, 2.88888888888889000, 9.55555555555556000, 14.88888888888890000,
	}
}

// testVarianceExpectedLength5Population is the Excel (VAR.P) output of population variance of length 5.
func testVarianceExpectedLength5Population() []float64 {
	return []float64{
		math.NaN(), math.NaN(), math.NaN(), math.NaN(),
		10.16000, 6.56000, 2.96000, 9.36000, 5.76000, 6.00000, 11.04000, 12.24000,
	}
}

// testVarianceExpectedLength3Sample is the Excel (VAR.S) output of sample variance of length 3.
func testVarianceExpectedLength3Sample() []float64 {
	return []float64{
		math.NaN(), math.NaN(),
		14.3333333333333000, 9.3333333333333400, 7.0000000000000000, 6.3333333333333400, 2.3333333333333300,
		14.3333333333333000, 9.3333333333333400, 4.3333333333333400, 14.3333333333333000, 22.3333333333333000,
	}
}

// testVarianceExpectedLength5Sample is the Excel (VAR.S) output of sample variance of length 5.
func testVarianceExpectedLength5Sample() []float64 {
	return []float64{
		math.NaN(), math.NaN(), math.NaN(), math.NaN(),
		12.7000, 8.2000, 3.7000, 11.7000, 7.2000, 7.5000, 13.8000, 15.3000,
	}
}

func TestVarianceUpdate(t *testing.T) { //nolint: funlen
	t.Parallel()

	check := func(index int, exp, act float64) {
		t.Helper()

		if math.Abs(exp-act) > 1e-13 {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	checkNaN := func(index int, act float64) {
		t.Helper()

		if !math.IsNaN(act) {
			t.Errorf("[%v] is incorrect: expected NaN, actual %v", index, act)
		}
	}

	input := testVarianceInput()

	t.Run("population variance length of 3", func(t *testing.T) {
		t.Parallel()
		v := testVarianceCreate(3, false)
		expected := testVarianceExpectedLength3Population()

		for i := 0; i < 2; i++ {
			checkNaN(i, v.Update(input[i]))
		}

		for i := 2; i < len(input); i++ {
			exp := expected[i]
			act := v.Update(input[i])
			check(i, exp, act)
		}

		checkNaN(0, v.Update(math.NaN()))
	})

	t.Run("population variance length of 5", func(t *testing.T) {
		t.Parallel()
		v := testVarianceCreate(5, false)
		expected := testVarianceExpectedLength5Population()

		for i := 0; i < 4; i++ {
			checkNaN(i, v.Update(input[i]))
		}

		for i := 4; i < len(input); i++ {
			exp := expected[i]
			act := v.Update(input[i])
			check(i, exp, act)
		}
	})

	t.Run("sample variance length of 3", func(t *testing.T) {
		t.Parallel()
		v := testVarianceCreate(3, true)
		expected := testVarianceExpectedLength3Sample()

		for i := 0; i < 2; i++ {
			checkNaN(i, v.Update(input[i]))
		}

		for i := 2; i < len(input); i++ {
			exp := expected[i]
			act := v.Update(input[i])
			check(i, exp, act)
		}

		checkNaN(0, v.Update(math.NaN()))
	})
}

func TestVarianceUpdateEntity(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		l   = 3
		inp = 3.
		exp = inp * inp / float64(l)
	)

	time := testVarianceTime()
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
		v := testVarianceCreate(l, true)
		v.Update(0.)
		v.Update(0.)
		check(v.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: time, Close: inp}
		v := testVarianceCreate(l, true)
		v.Update(0.)
		v.Update(0.)
		check(v.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: time, Bid: inp, Ask: inp}
		v := testVarianceCreate(l, true)
		v.Update(0.)
		v.Update(0.)
		check(v.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: time, Price: inp}
		v := testVarianceCreate(l, true)
		v.Update(0.)
		v.Update(0.)
		check(v.UpdateTrade(&r))
	})
}

func TestVarianceIsPrimed(t *testing.T) {
	t.Parallel()

	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	input := testVarianceInput()
	v := testVarianceCreate(3, false)

	check(0, false, v.IsPrimed())

	for i := 0; i < 2; i++ {
		v.Update(input[i])
		check(i+1, false, v.IsPrimed())
	}

	for i := 2; i < len(input); i++ {
		v.Update(input[i])
		check(i+1, true, v.IsPrimed())
	}
}

func TestVarianceMetadata(t *testing.T) {
	t.Parallel()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	t.Run("population variance", func(t *testing.T) {
		t.Parallel()
		v := testVarianceCreate(7, false)
		act := v.Metadata()

		check("Identifier", core.Variance, act.Identifier)
		check("Mnemonic", "var.p(7)", act.Mnemonic)
		check("Description", "Estimation of the population variance var.p(7)", act.Description)
		check("len(Outputs)", 1, len(act.Outputs))
		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
		check("Outputs[0].Mnemonic", "var.p(7)", act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", "Estimation of the population variance var.p(7)", act.Outputs[0].Description)
	})

	t.Run("sample variance", func(t *testing.T) {
		t.Parallel()
		v := testVarianceCreate(7, true)
		act := v.Metadata()

		check("Identifier", core.Variance, act.Identifier)
		check("Mnemonic", "var.s(7)", act.Mnemonic)
		check("Description", "Unbiased estimation of the sample variance var.s(7)", act.Description)
		check("len(Outputs)", 1, len(act.Outputs))
		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
		check("Outputs[0].Mnemonic", "var.s(7)", act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", "Unbiased estimation of the sample variance var.s(7)", act.Outputs[0].Description)
	})
}

func TestNewVariance(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		bc entities.BarComponent   = entities.BarMedianPrice
		qc entities.QuoteComponent = entities.QuoteMidPrice
		tc entities.TradeComponent = entities.TradePrice

		length = 5
		errlen = "invalid variance parameters: length should be greater than 1"
		errbc  = "invalid variance parameters: 9999: unknown bar component"
		errqc  = "invalid variance parameters: 9999: unknown quote component"
		errtc  = "invalid variance parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("length > 1, unbiased", func(t *testing.T) {
		t.Parallel()
		params := VarianceParams{
			Length: length, IsUnbiased: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		v, err := NewVariance(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "var.s(5, hl/2)", v.LineIndicator.Mnemonic)
		check("description", "Unbiased estimation of the sample variance var.s(5, hl/2)", v.LineIndicator.Description)
		check("unbiased", true, v.unbiased)
		check("primed", false, v.primed)
		check("windowLength", length, v.windowLength)
		check("lastIndex", length-1, v.lastIndex)
		check("len(window)", length, len(v.window))
		check("windowSum", 0., v.windowSum)
		check("windowSquaredSum", 0., v.windowSquaredSum)
		check("windowCount", 0, v.windowCount)
	})

	t.Run("length > 1, biased", func(t *testing.T) {
		t.Parallel()
		params := VarianceParams{
			Length: length, IsUnbiased: false, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		v, err := NewVariance(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "var.p(5, hl/2)", v.LineIndicator.Mnemonic)
		check("description", "Estimation of the population variance var.p(5, hl/2)", v.LineIndicator.Description)
		check("unbiased", false, v.unbiased)
		check("primed", false, v.primed)
		check("windowLength", length, v.windowLength)
		check("lastIndex", length-1, v.lastIndex)
		check("len(window)", length, len(v.window))
		check("windowSum", 0., v.windowSum)
		check("windowSquaredSum", 0., v.windowSquaredSum)
		check("windowCount", 0, v.windowCount)
	})

	t.Run("length = 1", func(t *testing.T) {
		t.Parallel()
		params := VarianceParams{
			Length: 1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		v, err := NewVariance(&params)
		check("v == nil", true, v == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length = 0", func(t *testing.T) {
		t.Parallel()
		params := VarianceParams{
			Length: 0, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		v, err := NewVariance(&params)
		check("v == nil", true, v == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length = -1", func(t *testing.T) {
		t.Parallel()
		params := VarianceParams{
			Length: -1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		v, err := NewVariance(&params)
		check("v == nil", true, v == nil)
		check("err", errlen, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()
		params := VarianceParams{
			Length: length, BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		v, err := NewVariance(&params)
		check("v == nil", true, v == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()
		params := VarianceParams{
			Length: length, BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		v, err := NewVariance(&params)
		check("v == nil", true, v == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()
		params := VarianceParams{
			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		v, err := NewVariance(&params)
		check("v == nil", true, v == nil)
		check("err", errtc, err.Error())
	})

	// Zero-value component mnemonic tests.
	// A zero value means "use default, don't show in mnemonic".

	t.Run("all components zero", func(t *testing.T) {
		t.Parallel()
		params := VarianceParams{Length: length, IsUnbiased: true}

		v, err := NewVariance(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "var.s(5)", v.LineIndicator.Mnemonic)
		check("description", "Unbiased estimation of the sample variance var.s(5)", v.LineIndicator.Description)
	})

	t.Run("only bar component set", func(t *testing.T) {
		t.Parallel()
		params := VarianceParams{Length: length, IsUnbiased: true, BarComponent: entities.BarMedianPrice}

		v, err := NewVariance(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "var.s(5, hl/2)", v.LineIndicator.Mnemonic)
		check("description", "Unbiased estimation of the sample variance var.s(5, hl/2)", v.LineIndicator.Description)
	})

	t.Run("only quote component set", func(t *testing.T) {
		t.Parallel()
		params := VarianceParams{Length: length, IsUnbiased: true, QuoteComponent: entities.QuoteBidPrice}

		v, err := NewVariance(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "var.s(5, b)", v.LineIndicator.Mnemonic)
		check("description", "Unbiased estimation of the sample variance var.s(5, b)", v.LineIndicator.Description)
	})

	t.Run("only trade component set", func(t *testing.T) {
		t.Parallel()
		params := VarianceParams{Length: length, IsUnbiased: true, TradeComponent: entities.TradeVolume}

		v, err := NewVariance(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "var.s(5, v)", v.LineIndicator.Mnemonic)
		check("description", "Unbiased estimation of the sample variance var.s(5, v)", v.LineIndicator.Description)
	})
}

func testVarianceCreate(length int, unbiased bool) *Variance {
	params := VarianceParams{
		Length: length, IsUnbiased: unbiased,
	}

	v, _ := NewVariance(&params)

	return v
}
