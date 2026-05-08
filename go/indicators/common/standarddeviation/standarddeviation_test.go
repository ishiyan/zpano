//nolint:testpackage
package standarddeviation

//nolint: gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func testStandardDeviationTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

func TestStandardDeviationUpdate(t *testing.T) { //nolint: funlen
	t.Parallel()

	check := func(index int, exp, act float64) {
		t.Helper()

		if math.Abs(exp-act) > 1e-10 {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	checkNaN := func(index int, act float64) {
		t.Helper()

		if !math.IsNaN(act) {
			t.Errorf("[%v] is incorrect: expected NaN, actual %v", index, act)
		}
	}

	input := testStandardDeviationInput()

	t.Run("population standard deviation", func(t *testing.T) {
		t.Parallel()
		sd := testStandardDeviationCreate(5, false)
		expected := testStandardDeviationLength5PopulationExpected()

		for i := 0; i < 4; i++ {
			checkNaN(i, sd.Update(input[i]))
		}

		for i := 4; i < len(input); i++ {
			exp := expected[i]
			act := sd.Update(input[i])
			check(i, exp, act)
		}

		checkNaN(0, sd.Update(math.NaN()))
	})

	t.Run("sample standard deviation", func(t *testing.T) {
		t.Parallel()
		sd := testStandardDeviationCreate(5, true)
		expected := testStandardDeviationLength5SampleExpected()

		for i := 0; i < 4; i++ {
			checkNaN(i, sd.Update(input[i]))
		}

		for i := 4; i < len(input); i++ {
			exp := expected[i]
			act := sd.Update(input[i])
			check(i, exp, act)
		}

		checkNaN(0, sd.Update(math.NaN()))
	})
}

func TestStandardDeviationUpdateEntity(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		l   = 3
		inp = 3.
	)

	exp := math.Sqrt(inp * inp / float64(l))
	time := testStandardDeviationTime()
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
		sd := testStandardDeviationCreate(l, true)
		sd.Update(0.)
		sd.Update(0.)
		check(sd.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: time, Close: inp}
		sd := testStandardDeviationCreate(l, true)
		sd.Update(0.)
		sd.Update(0.)
		check(sd.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: time, Bid: inp, Ask: inp}
		sd := testStandardDeviationCreate(l, true)
		sd.Update(0.)
		sd.Update(0.)
		check(sd.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: time, Price: inp}
		sd := testStandardDeviationCreate(l, true)
		sd.Update(0.)
		sd.Update(0.)
		check(sd.UpdateTrade(&r))
	})
}

func TestStandardDeviationIsPrimed(t *testing.T) {
	t.Parallel()

	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	input := testStandardDeviationInput()
	sd := testStandardDeviationCreate(5, true)

	check(0, false, sd.IsPrimed())

	for i := 0; i < 4; i++ {
		sd.Update(input[i])
		check(i+1, false, sd.IsPrimed())
	}

	for i := 4; i < len(input); i++ {
		sd.Update(input[i])
		check(i+1, true, sd.IsPrimed())
	}
}

func TestStandardDeviationMetadata(t *testing.T) {
	t.Parallel()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	t.Run("population standard deviation", func(t *testing.T) {
		t.Parallel()
		sd := testStandardDeviationCreate(7, false)
		act := sd.Metadata()

		check("Identifier", core.StandardDeviation, act.Identifier)
		check("Mnemonic", "stdev.p(7)", act.Mnemonic)
		check("Description", "Standard deviation based on estimation of the population variance stdev.p(7)", act.Description)
		check("len(Outputs)", 1, len(act.Outputs))
		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
		check("Outputs[0].Mnemonic", "stdev.p(7)", act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", "Standard deviation based on estimation of the population variance stdev.p(7)", act.Outputs[0].Description)
	})

	t.Run("sample standard deviation", func(t *testing.T) {
		t.Parallel()
		sd := testStandardDeviationCreate(7, true)
		act := sd.Metadata()

		check("Identifier", core.StandardDeviation, act.Identifier)
		check("Mnemonic", "stdev.s(7)", act.Mnemonic)
		check("Description", "Standard deviation based on unbiased estimation of the sample variance stdev.s(7)", act.Description)
		check("len(Outputs)", 1, len(act.Outputs))
		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
		check("Outputs[0].Mnemonic", "stdev.s(7)", act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", "Standard deviation based on unbiased estimation of the sample variance stdev.s(7)", act.Outputs[0].Description)
	})
}

func TestNewStandardDeviation(t *testing.T) { //nolint: funlen
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
		params := StandardDeviationParams{
			Length: length, IsUnbiased: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		sd, err := NewStandardDeviation(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "stdev.s(5, hl/2)", sd.LineIndicator.Mnemonic)
		check("description", "Standard deviation based on unbiased estimation of the sample variance stdev.s(5, hl/2)", sd.LineIndicator.Description)
		check("variance != nil", true, sd.variance != nil)
	})

	t.Run("length > 1, biased", func(t *testing.T) {
		t.Parallel()
		params := StandardDeviationParams{
			Length: length, IsUnbiased: false, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		sd, err := NewStandardDeviation(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "stdev.p(5, hl/2)", sd.LineIndicator.Mnemonic)
		check("description", "Standard deviation based on estimation of the population variance stdev.p(5, hl/2)", sd.LineIndicator.Description)
		check("variance != nil", true, sd.variance != nil)
	})

	t.Run("length = 1", func(t *testing.T) {
		t.Parallel()
		params := StandardDeviationParams{
			Length: 1, IsUnbiased: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		sd, err := NewStandardDeviation(&params)
		check("sd == nil", true, sd == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length = 0", func(t *testing.T) {
		t.Parallel()
		params := StandardDeviationParams{
			Length: 0, IsUnbiased: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		sd, err := NewStandardDeviation(&params)
		check("sd == nil", true, sd == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length = -1", func(t *testing.T) {
		t.Parallel()
		params := StandardDeviationParams{
			Length: -1, IsUnbiased: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		sd, err := NewStandardDeviation(&params)
		check("sd == nil", true, sd == nil)
		check("err", errlen, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()
		params := StandardDeviationParams{
			Length: length, IsUnbiased: true, BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		sd, err := NewStandardDeviation(&params)
		check("sd == nil", true, sd == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()
		params := StandardDeviationParams{
			Length: length, IsUnbiased: true, BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		sd, err := NewStandardDeviation(&params)
		check("sd == nil", true, sd == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()
		params := StandardDeviationParams{
			Length: length, IsUnbiased: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		sd, err := NewStandardDeviation(&params)
		check("sd == nil", true, sd == nil)
		check("err", errtc, err.Error())
	})

	// Zero-value component mnemonic tests.
	// A zero value means "use default, don't show in mnemonic".

	t.Run("all components zero", func(t *testing.T) {
		t.Parallel()
		params := StandardDeviationParams{Length: length, IsUnbiased: true}

		sd, err := NewStandardDeviation(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "stdev.s(5)", sd.LineIndicator.Mnemonic)
		check("description", "Standard deviation based on unbiased estimation of the sample variance stdev.s(5)", sd.LineIndicator.Description)
	})

	t.Run("only bar component set", func(t *testing.T) {
		t.Parallel()
		params := StandardDeviationParams{Length: length, IsUnbiased: true, BarComponent: entities.BarMedianPrice}

		sd, err := NewStandardDeviation(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "stdev.s(5, hl/2)", sd.LineIndicator.Mnemonic)
		check("description", "Standard deviation based on unbiased estimation of the sample variance stdev.s(5, hl/2)", sd.LineIndicator.Description)
	})

	t.Run("only quote component set", func(t *testing.T) {
		t.Parallel()
		params := StandardDeviationParams{Length: length, IsUnbiased: true, QuoteComponent: entities.QuoteBidPrice}

		sd, err := NewStandardDeviation(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "stdev.s(5, b)", sd.LineIndicator.Mnemonic)
		check("description", "Standard deviation based on unbiased estimation of the sample variance stdev.s(5, b)", sd.LineIndicator.Description)
	})

	t.Run("only trade component set", func(t *testing.T) {
		t.Parallel()
		params := StandardDeviationParams{Length: length, IsUnbiased: true, TradeComponent: entities.TradeVolume}

		sd, err := NewStandardDeviation(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "stdev.s(5, v)", sd.LineIndicator.Mnemonic)
		check("description", "Standard deviation based on unbiased estimation of the sample variance stdev.s(5, v)", sd.LineIndicator.Description)
	})
}

func testStandardDeviationCreate(length int, unbiased bool) *StandardDeviation {
	params := StandardDeviationParams{
		Length: length, IsUnbiased: unbiased,
	}

	sd, _ := NewStandardDeviation(&params)

	return sd
}
