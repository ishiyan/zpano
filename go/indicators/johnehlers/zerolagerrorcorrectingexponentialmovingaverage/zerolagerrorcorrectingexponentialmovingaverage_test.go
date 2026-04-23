//nolint:testpackage
package zerolagerrorcorrectingexponentialmovingaverage

//nolint: gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func testZecemaTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

func testZecemaCreate(sf float64, gl float64, gs float64) *ZeroLagErrorCorrectingExponentialMovingAverage {
	params := ZeroLagErrorCorrectingExponentialMovingAverageParams{
		SmoothingFactor: sf,
		GainLimit:       gl,
		GainStep:        gs,
	}

	z, _ := NewZeroLagErrorCorrectingExponentialMovingAverage(&params)

	return z
}

func testZecemaCreateDefault() *ZeroLagErrorCorrectingExponentialMovingAverage {
	return testZecemaCreate(0.095, 5, 0.1)
}

func TestZecemaIsPrimed(t *testing.T) {
	t.Parallel()

	z := testZecemaCreateDefault()

	if z.IsPrimed() {
		t.Error("should not be primed before any updates")
	}

	// First 2 updates should not prime.
	for i := 0; i < 2; i++ {
		z.Update(100)

		if z.IsPrimed() {
			t.Errorf("[%d] should not be primed", i)
		}
	}

	// 3rd update should prime.
	z.Update(100)

	if !z.IsPrimed() {
		t.Error("[2] should be primed")
	}
}

func TestZecemaUpdateNaN(t *testing.T) {
	t.Parallel()

	z := testZecemaCreateDefault()

	if !math.IsNaN(z.Update(math.NaN())) {
		t.Error("expected NaN passthrough")
	}

	// NaN should not change state.
	if z.IsPrimed() {
		t.Error("should not be primed after NaN")
	}
}

func TestZecemaUpdateConstant(t *testing.T) {
	t.Parallel()

	const value = 42.0

	z := testZecemaCreateDefault()

	// Feed constant values. First 2 should return NaN.
	for i := 0; i < 2; i++ {
		act := z.Update(value)
		if !math.IsNaN(act) {
			t.Errorf("[%d] expected NaN during priming, got %v", i, act)
		}
	}

	// 3rd update primes.
	act := z.Update(value)
	if math.IsNaN(act) {
		t.Error("expected non-NaN after priming")
	}

	if math.Abs(act-value) > 1e-6 {
		t.Errorf("expected close to %v after priming with constant input, got %v", value, act)
	}

	// Further updates with same constant should stay close to value.
	for i := 0; i < 10; i++ {
		act = z.Update(value)
		if math.Abs(act-value) > 1e-6 {
			t.Errorf("[%d] expected close to %v, got %v", i, value, act)
		}
	}
}

func TestZecemaUpdateEntity(t *testing.T) {
	t.Parallel()

	const inp = 100.

	tm := testZecemaTime()
	z := testZecemaCreateDefault()

	// Prime the indicator (3 updates).
	z.Update(inp)
	z.Update(inp)
	z.Update(inp)

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

		if math.IsNaN(s.Value) {
			t.Error("value should not be NaN after priming")
		}
	}

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		s := entities.Scalar{Time: tm, Value: inp}
		z2 := testZecemaCreateDefault()
		z2.Update(inp)
		z2.Update(inp)
		z2.Update(inp)
		check(z2.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: tm, Close: inp}
		z2 := testZecemaCreateDefault()
		z2.Update(inp)
		z2.Update(inp)
		z2.Update(inp)
		check(z2.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: tm, Bid: inp, Ask: inp}
		z2 := testZecemaCreateDefault()
		z2.Update(inp)
		z2.Update(inp)
		z2.Update(inp)
		check(z2.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: tm, Price: inp}
		z2 := testZecemaCreateDefault()
		z2.Update(inp)
		z2.Update(inp)
		z2.Update(inp)
		check(z2.UpdateTrade(&r))
	})
}

func TestZecemaMetadata(t *testing.T) {
	t.Parallel()

	z := testZecemaCreateDefault()
	act := z.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	check("Identifier", core.ZeroLagErrorCorrectingExponentialMovingAverage, act.Identifier)
	check("len(Outputs)", 1, len(act.Outputs))
	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", "zecema(0.095, 5, 0.1)", act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", "Zero-lag Error-Correcting Exponential Moving Average zecema(0.095, 5, 0.1)", act.Outputs[0].Description)
}

func TestNewZeroLagErrorCorrectingExponentialMovingAverage(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		errSf = "invalid zero-lag error-correcting exponential moving average parameters: smoothing factor should be in (0, 1]"
		errGl = "invalid zero-lag error-correcting exponential moving average parameters: gain limit should be positive"
		errGs = "invalid zero-lag error-correcting exponential moving average parameters: gain step should be positive"
		errBc = "invalid zero-lag error-correcting exponential moving average parameters: 9999: unknown bar component"
		errQc = "invalid zero-lag error-correcting exponential moving average parameters: 9999: unknown quote component"
		errTc = "invalid zero-lag error-correcting exponential moving average parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("valid defaults", func(t *testing.T) {
		t.Parallel()
		params := ZeroLagErrorCorrectingExponentialMovingAverageParams{
			SmoothingFactor: 0.095, GainLimit: 5, GainStep: 0.1,
		}

		z, err := NewZeroLagErrorCorrectingExponentialMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "zecema(0.095, 5, 0.1)", z.LineIndicator.Mnemonic)
		check("primed", false, z.primed)
	})

	t.Run("smoothing factor = 0", func(t *testing.T) {
		t.Parallel()
		params := ZeroLagErrorCorrectingExponentialMovingAverageParams{
			SmoothingFactor: 0, GainLimit: 5, GainStep: 0.1,
		}

		z, err := NewZeroLagErrorCorrectingExponentialMovingAverage(&params)
		check("z == nil", true, z == nil)
		check("err", errSf, err.Error())
	})

	t.Run("smoothing factor < 0", func(t *testing.T) {
		t.Parallel()
		params := ZeroLagErrorCorrectingExponentialMovingAverageParams{
			SmoothingFactor: -0.1, GainLimit: 5, GainStep: 0.1,
		}

		z, err := NewZeroLagErrorCorrectingExponentialMovingAverage(&params)
		check("z == nil", true, z == nil)
		check("err", errSf, err.Error())
	})

	t.Run("smoothing factor > 1", func(t *testing.T) {
		t.Parallel()
		params := ZeroLagErrorCorrectingExponentialMovingAverageParams{
			SmoothingFactor: 1.1, GainLimit: 5, GainStep: 0.1,
		}

		z, err := NewZeroLagErrorCorrectingExponentialMovingAverage(&params)
		check("z == nil", true, z == nil)
		check("err", errSf, err.Error())
	})

	t.Run("smoothing factor = 1", func(t *testing.T) {
		t.Parallel()
		params := ZeroLagErrorCorrectingExponentialMovingAverageParams{
			SmoothingFactor: 1, GainLimit: 5, GainStep: 0.1,
		}

		z, err := NewZeroLagErrorCorrectingExponentialMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("z != nil", true, z != nil)
	})

	t.Run("gain limit = 0", func(t *testing.T) {
		t.Parallel()
		params := ZeroLagErrorCorrectingExponentialMovingAverageParams{
			SmoothingFactor: 0.095, GainLimit: 0, GainStep: 0.1,
		}

		z, err := NewZeroLagErrorCorrectingExponentialMovingAverage(&params)
		check("z == nil", true, z == nil)
		check("err", errGl, err.Error())
	})

	t.Run("gain limit < 0", func(t *testing.T) {
		t.Parallel()
		params := ZeroLagErrorCorrectingExponentialMovingAverageParams{
			SmoothingFactor: 0.095, GainLimit: -1, GainStep: 0.1,
		}

		z, err := NewZeroLagErrorCorrectingExponentialMovingAverage(&params)
		check("z == nil", true, z == nil)
		check("err", errGl, err.Error())
	})

	t.Run("gain step = 0", func(t *testing.T) {
		t.Parallel()
		params := ZeroLagErrorCorrectingExponentialMovingAverageParams{
			SmoothingFactor: 0.095, GainLimit: 5, GainStep: 0,
		}

		z, err := NewZeroLagErrorCorrectingExponentialMovingAverage(&params)
		check("z == nil", true, z == nil)
		check("err", errGs, err.Error())
	})

	t.Run("gain step < 0", func(t *testing.T) {
		t.Parallel()
		params := ZeroLagErrorCorrectingExponentialMovingAverageParams{
			SmoothingFactor: 0.095, GainLimit: 5, GainStep: -0.1,
		}

		z, err := NewZeroLagErrorCorrectingExponentialMovingAverage(&params)
		check("z == nil", true, z == nil)
		check("err", errGs, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()
		params := ZeroLagErrorCorrectingExponentialMovingAverageParams{
			SmoothingFactor: 0.095, GainLimit: 5, GainStep: 0.1,
			BarComponent: entities.BarComponent(9999),
		}

		z, err := NewZeroLagErrorCorrectingExponentialMovingAverage(&params)
		check("z == nil", true, z == nil)
		check("err", errBc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()
		params := ZeroLagErrorCorrectingExponentialMovingAverageParams{
			SmoothingFactor: 0.095, GainLimit: 5, GainStep: 0.1,
			QuoteComponent: entities.QuoteComponent(9999),
		}

		z, err := NewZeroLagErrorCorrectingExponentialMovingAverage(&params)
		check("z == nil", true, z == nil)
		check("err", errQc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()
		params := ZeroLagErrorCorrectingExponentialMovingAverageParams{
			SmoothingFactor: 0.095, GainLimit: 5, GainStep: 0.1,
			TradeComponent: entities.TradeComponent(9999),
		}

		z, err := NewZeroLagErrorCorrectingExponentialMovingAverage(&params)
		check("z == nil", true, z == nil)
		check("err", errTc, err.Error())
	})

	t.Run("non-default bar component", func(t *testing.T) {
		t.Parallel()
		params := ZeroLagErrorCorrectingExponentialMovingAverageParams{
			SmoothingFactor: 0.095, GainLimit: 5, GainStep: 0.1,
			BarComponent: entities.BarOpenPrice,
		}

		z, err := NewZeroLagErrorCorrectingExponentialMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "zecema(0.095, 5, 0.1, o)", z.LineIndicator.Mnemonic)
	})
}
