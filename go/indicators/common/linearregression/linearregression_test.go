//nolint:testpackage
package linearregression

//nolint: gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func testLinearRegressionTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

func TestLinearRegressionUpdate(t *testing.T) { //nolint: funlen
	t.Parallel()

	const tolerance = 1e-4

	check := func(name string, index int, exp, act float64) {
		t.Helper()

		if math.Abs(exp-act) > tolerance {
			t.Errorf("[%d] %s: expected %v, actual %v", index, name, exp, act)
		}
	}

	checkNaN := func(index int, act float64) {
		t.Helper()

		if !math.IsNaN(act) {
			t.Errorf("[%v] is incorrect: expected NaN, actual %v", index, act)
		}
	}

	input := testLinearRegressionInput()
	expValue := testLinearRegressionExpectedValue()
	expForecast := testLinearRegressionExpectedForecast()
	expIntercept := testLinearRegressionExpectedIntercept()
	expSlopeRad := testLinearRegressionExpectedSlopeRad()
	expSlopeDeg := testLinearRegressionExpectedSlopeDeg()

	t.Run("period 14 Value output all 252 rows", func(t *testing.T) {
		t.Parallel()

		linreg := testLinearRegressionCreate(14)

		// Feed first 12 samples (indices 0-12), all should be NaN.
		for i := 0; i < 13; i++ {
			checkNaN(i, linreg.Update(input[i]))
		}

		for i := 13; i < len(input); i++ {
			value := linreg.Update(input[i])
			check("Value", i, expValue[i], value)
		}

		checkNaN(0, linreg.Update(math.NaN()))
	})

	t.Run("period 14 all 5 outputs via updateEntity all 252 rows", func(t *testing.T) {
		t.Parallel()

		linreg := testLinearRegressionCreate(14)
		tm := testLinearRegressionTime()

		// Feed first 12 samples.
		for i := 0; i < 12; i++ {
			linreg.Update(input[i])
		}

		// Feed index 12 via UpdateScalar to get NaN outputs.
		out := linreg.UpdateScalar(&entities.Scalar{Time: tm, Value: input[12]})
		if len(out) != 5 {
			t.Fatalf("expected 5 outputs, got %d", len(out))
		}

		for j := 0; j < 5; j++ {
			s, ok := out[j].(entities.Scalar)
			if !ok {
				t.Fatalf("output[%d] is not Scalar", j)
			}

			if !math.IsNaN(s.Value) {
				t.Errorf("output[%d] expected NaN, got %v", j, s.Value)
			}
		}

		// Feed indices 13-251 via UpdateScalar and verify all 5 outputs.
		for i := 13; i < len(input); i++ {
			out = linreg.UpdateScalar(&entities.Scalar{Time: tm, Value: input[i]})
			if len(out) != 5 {
				t.Fatalf("[%d] expected 5 outputs, got %d", i, len(out))
			}

			s0 := out[0].(entities.Scalar)
			s1 := out[1].(entities.Scalar)
			s2 := out[2].(entities.Scalar)
			s3 := out[3].(entities.Scalar)
			s4 := out[4].(entities.Scalar)

			check("Value", i, expValue[i], s0.Value)
			check("Forecast", i, expForecast[i], s1.Value)
			check("Intercept", i, expIntercept[i], s2.Value)
			check("SlopeRad", i, expSlopeRad[i], s3.Value)
			check("SlopeDeg", i, expSlopeDeg[i], s4.Value)
		}
	})
}

func TestLinearRegressionUpdateEntity(t *testing.T) { //nolint: funlen
	t.Parallel()

	const numOutputs = 5

	tm := testLinearRegressionTime()
	input := testLinearRegressionInput()

	// Prime an indicator with 14 samples, then check entity updates.
	setup := func() *LinearRegression {
		lr := testLinearRegressionCreate(14)
		for i := 0; i < 14; i++ {
			lr.Update(input[i])
		}

		return lr
	}

	checkLen := func(out core.Output) {
		t.Helper()

		if len(out) != numOutputs {
			t.Fatalf("expected %d outputs, got %d", numOutputs, len(out))
		}
	}

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()
		lr := setup()
		b := entities.Bar{Time: tm, Close: input[14]}
		out := lr.UpdateBar(&b)
		checkLen(out)

		s, ok := out[0].(entities.Scalar)
		if !ok {
			t.Fatal("output is not scalar")
		}

		if s.Time != tm {
			t.Errorf("time mismatch: expected %v, got %v", tm, s.Time)
		}
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()
		lr := setup()
		q := entities.Quote{Time: tm, Bid: input[14], Ask: input[14]}
		out := lr.UpdateQuote(&q)
		checkLen(out)
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()
		lr := setup()
		r := entities.Trade{Time: tm, Price: input[14]}
		out := lr.UpdateTrade(&r)
		checkLen(out)
	})
}

func TestLinearRegressionIsPrimed(t *testing.T) {
	t.Parallel()

	input := testLinearRegressionInput()

	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	t.Run("length = 14", func(t *testing.T) {
		t.Parallel()
		lr := testLinearRegressionCreate(14)

		check(-1, false, lr.IsPrimed())

		for i := 0; i < 13; i++ {
			lr.Update(input[i])
			check(i, false, lr.IsPrimed())
		}

		for i := 13; i < len(input); i++ {
			lr.Update(input[i])
			check(i, true, lr.IsPrimed())
		}
	})

	t.Run("length = 2", func(t *testing.T) {
		t.Parallel()
		lr := testLinearRegressionCreate(2)

		check(-1, false, lr.IsPrimed())
		lr.Update(input[0])
		check(0, false, lr.IsPrimed())
		lr.Update(input[1])
		check(1, true, lr.IsPrimed())
	})
}

func TestLinearRegressionMetadata(t *testing.T) {
	t.Parallel()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	lr := testLinearRegressionCreate(14)
	act := lr.Metadata()

	check("Identifier", core.LinearRegression, act.Identifier)
	check("Mnemonic", "linreg(14)", act.Mnemonic)
	check("Description", "Linear Regression linreg(14)", act.Description)
	check("len(Outputs)", 5, len(act.Outputs))
	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
	check("Outputs[1].Kind", int(Forecast), act.Outputs[1].Kind)
	check("Outputs[2].Kind", int(Intercept), act.Outputs[2].Kind)
	check("Outputs[3].Kind", int(SlopeRad), act.Outputs[3].Kind)
	check("Outputs[4].Kind", int(SlopeDeg), act.Outputs[4].Kind)
}

func TestNewLinearRegression(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		bc     entities.BarComponent   = entities.BarMedianPrice
		qc     entities.QuoteComponent = entities.QuoteMidPrice
		tc     entities.TradeComponent = entities.TradePrice
		length                         = 14
		errlen                         = "invalid linear regression parameters: length should be greater than 1"
		errbc                          = "invalid linear regression parameters: 9999: unknown bar component"
		errqc                          = "invalid linear regression parameters: 9999: unknown quote component"
		errtc                          = "invalid linear regression parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("valid params", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		lr, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "linreg(14, hl/2)", lr.mnemonic)
		check("description", "Linear Regression linreg(14, hl/2)", lr.description)
		check("primed", false, lr.primed)
		check("length", length, lr.length)
	})

	t.Run("length = 1", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: 1}

		lr, err := New(&params)
		check("lr == nil", true, lr == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length = 0", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: 0}

		lr, err := New(&params)
		check("lr == nil", true, lr == nil)
		check("err", errlen, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		lr, err := New(&params)
		check("lr == nil", true, lr == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		lr, err := New(&params)
		check("lr == nil", true, lr == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		lr, err := New(&params)
		check("lr == nil", true, lr == nil)
		check("err", errtc, err.Error())
	})

	t.Run("all components zero", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length}

		lr, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "linreg(14)", lr.mnemonic)
	})
}

func testLinearRegressionCreate(length int) *LinearRegression {
	params := Params{Length: length}

	lr, _ := New(&params)

	return lr
}
