//nolint:testpackage
package zerolagexponentialmovingaverage

//nolint: gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func testZemaTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

func testZemaCreate(sf float64, gf float64, ml int) *ZeroLagExponentialMovingAverage {
	params := ZeroLagExponentialMovingAverageParams{
		SmoothingFactor:        sf,
		VelocityGainFactor:     gf,
		VelocityMomentumLength: ml,
	}

	z, _ := NewZeroLagExponentialMovingAverage(&params)

	return z
}

func testZemaCreateDefault() *ZeroLagExponentialMovingAverage {
	return testZemaCreate(0.25, 0.5, 3)
}

func TestZemaIsPrimed(t *testing.T) {
	t.Parallel()

	z := testZemaCreateDefault()

	if z.IsPrimed() {
		t.Error("should not be primed before any updates")
	}

	// First 3 updates (momentumLength=3) should not prime.
	for i := 0; i < 3; i++ {
		z.Update(100)

		if z.IsPrimed() {
			t.Errorf("[%d] should not be primed", i)
		}
	}

	// 4th update should prime.
	z.Update(100)

	if !z.IsPrimed() {
		t.Error("[3] should be primed")
	}
}

func TestZemaUpdateNaN(t *testing.T) {
	t.Parallel()

	z := testZemaCreateDefault()

	if !math.IsNaN(z.Update(math.NaN())) {
		t.Error("expected NaN passthrough")
	}

	// NaN should not change state.
	if z.IsPrimed() {
		t.Error("should not be primed after NaN")
	}
}

func TestZemaUpdateConstant(t *testing.T) {
	t.Parallel()

	const value = 42.0

	z := testZemaCreateDefault()

	// Feed constant values. After priming, momentum=0 so output should equal input.
	for i := 0; i < 3; i++ {
		act := z.Update(value)
		if !math.IsNaN(act) {
			t.Errorf("[%d] expected NaN during priming, got %v", i, act)
		}
	}

	act := z.Update(value)
	if math.Abs(act-value) > 1e-10 {
		t.Errorf("expected %v after priming with constant input, got %v", value, act)
	}

	// Further updates with same constant should stay at value.
	for i := 0; i < 10; i++ {
		act = z.Update(value)
		if math.Abs(act-value) > 1e-10 {
			t.Errorf("[%d] expected %v, got %v", i, value, act)
		}
	}
}

func TestZemaUpdateEntity(t *testing.T) {
	t.Parallel()

	const inp = 100.

	tm := testZemaTime()
	z := testZemaCreateDefault()

	// Prime the indicator (4 updates for momentumLength=3).
	z.Update(inp)
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
		z2 := testZemaCreateDefault()
		z2.Update(inp)
		z2.Update(inp)
		z2.Update(inp)
		z2.Update(inp)
		check(z2.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: tm, Close: inp}
		z2 := testZemaCreateDefault()
		z2.Update(inp)
		z2.Update(inp)
		z2.Update(inp)
		z2.Update(inp)
		check(z2.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: tm, Bid: inp, Ask: inp}
		z2 := testZemaCreateDefault()
		z2.Update(inp)
		z2.Update(inp)
		z2.Update(inp)
		z2.Update(inp)
		check(z2.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: tm, Price: inp}
		z2 := testZemaCreateDefault()
		z2.Update(inp)
		z2.Update(inp)
		z2.Update(inp)
		z2.Update(inp)
		check(z2.UpdateTrade(&r))
	})
}

func TestZemaMetadata(t *testing.T) {
	t.Parallel()

	z := testZemaCreateDefault()
	act := z.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	check("Identifier", core.ZeroLagExponentialMovingAverage, act.Identifier)
	check("len(Outputs)", 1, len(act.Outputs))
	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", "zema(0.25, 0.5, 3)", act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", "Zero-lag Exponential Moving Average zema(0.25, 0.5, 3)", act.Outputs[0].Description)
}

func TestNewZeroLagExponentialMovingAverage(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		errSf = "invalid zero-lag exponential moving average parameters: smoothing factor should be in (0, 1]"
		errMl = "invalid zero-lag exponential moving average parameters: velocity momentum length should be positive"
		errBc = "invalid zero-lag exponential moving average parameters: 9999: unknown bar component"
		errQc = "invalid zero-lag exponential moving average parameters: 9999: unknown quote component"
		errTc = "invalid zero-lag exponential moving average parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("valid defaults", func(t *testing.T) {
		t.Parallel()
		params := ZeroLagExponentialMovingAverageParams{
			SmoothingFactor: 0.25, VelocityGainFactor: 0.5, VelocityMomentumLength: 3,
		}

		z, err := NewZeroLagExponentialMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "zema(0.25, 0.5, 3)", z.LineIndicator.Mnemonic)
		check("primed", false, z.primed)
	})

	t.Run("smoothing factor = 0", func(t *testing.T) {
		t.Parallel()
		params := ZeroLagExponentialMovingAverageParams{
			SmoothingFactor: 0, VelocityGainFactor: 0.5, VelocityMomentumLength: 3,
		}

		z, err := NewZeroLagExponentialMovingAverage(&params)
		check("z == nil", true, z == nil)
		check("err", errSf, err.Error())
	})

	t.Run("smoothing factor < 0", func(t *testing.T) {
		t.Parallel()
		params := ZeroLagExponentialMovingAverageParams{
			SmoothingFactor: -0.1, VelocityGainFactor: 0.5, VelocityMomentumLength: 3,
		}

		z, err := NewZeroLagExponentialMovingAverage(&params)
		check("z == nil", true, z == nil)
		check("err", errSf, err.Error())
	})

	t.Run("smoothing factor > 1", func(t *testing.T) {
		t.Parallel()
		params := ZeroLagExponentialMovingAverageParams{
			SmoothingFactor: 1.1, VelocityGainFactor: 0.5, VelocityMomentumLength: 3,
		}

		z, err := NewZeroLagExponentialMovingAverage(&params)
		check("z == nil", true, z == nil)
		check("err", errSf, err.Error())
	})

	t.Run("smoothing factor = 1", func(t *testing.T) {
		t.Parallel()
		params := ZeroLagExponentialMovingAverageParams{
			SmoothingFactor: 1, VelocityGainFactor: 0.5, VelocityMomentumLength: 3,
		}

		z, err := NewZeroLagExponentialMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("z != nil", true, z != nil)
	})

	t.Run("momentum length = 0", func(t *testing.T) {
		t.Parallel()
		params := ZeroLagExponentialMovingAverageParams{
			SmoothingFactor: 0.25, VelocityGainFactor: 0.5, VelocityMomentumLength: 0,
		}

		z, err := NewZeroLagExponentialMovingAverage(&params)
		check("z == nil", true, z == nil)
		check("err", errMl, err.Error())
	})

	t.Run("momentum length < 0", func(t *testing.T) {
		t.Parallel()
		params := ZeroLagExponentialMovingAverageParams{
			SmoothingFactor: 0.25, VelocityGainFactor: 0.5, VelocityMomentumLength: -1,
		}

		z, err := NewZeroLagExponentialMovingAverage(&params)
		check("z == nil", true, z == nil)
		check("err", errMl, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()
		params := ZeroLagExponentialMovingAverageParams{
			SmoothingFactor: 0.25, VelocityGainFactor: 0.5, VelocityMomentumLength: 3,
			BarComponent: entities.BarComponent(9999),
		}

		z, err := NewZeroLagExponentialMovingAverage(&params)
		check("z == nil", true, z == nil)
		check("err", errBc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()
		params := ZeroLagExponentialMovingAverageParams{
			SmoothingFactor: 0.25, VelocityGainFactor: 0.5, VelocityMomentumLength: 3,
			QuoteComponent: entities.QuoteComponent(9999),
		}

		z, err := NewZeroLagExponentialMovingAverage(&params)
		check("z == nil", true, z == nil)
		check("err", errQc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()
		params := ZeroLagExponentialMovingAverageParams{
			SmoothingFactor: 0.25, VelocityGainFactor: 0.5, VelocityMomentumLength: 3,
			TradeComponent: entities.TradeComponent(9999),
		}

		z, err := NewZeroLagExponentialMovingAverage(&params)
		check("z == nil", true, z == nil)
		check("err", errTc, err.Error())
	})

	t.Run("non-default bar component", func(t *testing.T) {
		t.Parallel()
		params := ZeroLagExponentialMovingAverageParams{
			SmoothingFactor: 0.25, VelocityGainFactor: 0.5, VelocityMomentumLength: 3,
			BarComponent: entities.BarOpenPrice,
		}

		z, err := NewZeroLagExponentialMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "zema(0.25, 0.5, 3, o)", z.LineIndicator.Mnemonic)
	})
}
