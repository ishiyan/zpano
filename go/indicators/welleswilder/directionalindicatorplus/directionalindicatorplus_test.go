//nolint:testpackage
package directionalindicatorplus

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func TestDirectionalIndicatorPlusConstructor(t *testing.T) {
	t.Parallel()

	dip, err := NewDirectionalIndicatorPlus(14)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dip.Length() != 14 {
		t.Errorf("expected length 14, got %d", dip.Length())
	}

	if dip.IsPrimed() {
		t.Error("should not be primed initially")
	}

	_, err = NewDirectionalIndicatorPlus(0)
	if err == nil {
		t.Error("expected error for length 0")
	}

	_, err = NewDirectionalIndicatorPlus(-8)
	if err == nil {
		t.Error("expected error for negative length")
	}
}

func TestDirectionalIndicatorPlusIsPrimed(t *testing.T) {
	t.Parallel()

	high := testInputHigh()
	low := testInputLow()
	close_ := testInputClose()

	t.Run("length=14", func(t *testing.T) {
		t.Parallel()

		dip, _ := NewDirectionalIndicatorPlus(14)

		for i := 0; i < 14; i++ {
			dip.Update(close_[i], high[i], low[i])

			if dip.IsPrimed() {
				t.Errorf("[%d] should not be primed yet", i)
			}
		}

		dip.Update(close_[14], high[14], low[14])
		if !dip.IsPrimed() {
			t.Error("[14] should be primed")
		}
	})
}

func TestDirectionalIndicatorPlusUpdate(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-8

	high := testInputHigh()
	low := testInputLow()
	close_ := testInputClose()
	expected := testExpectedDI14()
	dip, _ := NewDirectionalIndicatorPlus(14)

	for i := range high {
		act := dip.Update(close_[i], high[i], low[i])

		if math.IsNaN(expected[i]) {
			if !math.IsNaN(act) {
				t.Errorf("[%d] expected NaN, got %v", i, act)
			}

			continue
		}

		if math.IsNaN(act) {
			t.Errorf("[%d] expected %v, got NaN", i, expected[i])
			continue
		}

		if math.Abs(act-expected[i]) > tolerance {
			t.Errorf("[%d] expected %v, got %v", i, expected[i], act)
		}
	}
}

func TestDirectionalIndicatorPlusNaNPassthrough(t *testing.T) {
	t.Parallel()

	dip, _ := NewDirectionalIndicatorPlus(14)

	if !math.IsNaN(dip.Update(math.NaN(), 1, 1)) {
		t.Error("expected NaN passthrough for NaN close")
	}

	if !math.IsNaN(dip.Update(1, math.NaN(), 1)) {
		t.Error("expected NaN passthrough for NaN high")
	}

	if !math.IsNaN(dip.Update(1, 1, math.NaN())) {
		t.Error("expected NaN passthrough for NaN low")
	}

	if !math.IsNaN(dip.UpdateSample(math.NaN())) {
		t.Error("expected NaN passthrough for NaN sample")
	}
}

func TestDirectionalIndicatorPlusUpdateEntity(t *testing.T) {
	t.Parallel()

	tm := testTime()
	high := testInputHigh()
	low := testInputLow()
	close_ := testInputClose()

	check := func(t *testing.T, act core.Output) {
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
	}

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		dip, _ := NewDirectionalIndicatorPlus(14)
		for i := 0; i < 14; i++ {
			dip.Update(close_[i], high[i], low[i])
		}

		b := entities.Bar{Time: tm, Close: close_[14], High: high[14], Low: low[14]}
		check(t, dip.UpdateBar(&b))
	})

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		dip, _ := NewDirectionalIndicatorPlus(14)
		for i := 0; i < 14; i++ {
			dip.Update(close_[i], high[i], low[i])
		}

		s := entities.Scalar{Time: tm, Value: high[14]}
		check(t, dip.UpdateScalar(&s))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		dip, _ := NewDirectionalIndicatorPlus(14)
		for i := 0; i < 14; i++ {
			dip.Update(close_[i], high[i], low[i])
		}

		q := entities.Quote{Time: tm, Bid: high[14] - 0.5, Ask: high[14] + 0.5}
		check(t, dip.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		dip, _ := NewDirectionalIndicatorPlus(14)
		for i := 0; i < 14; i++ {
			dip.Update(close_[i], high[i], low[i])
		}

		r := entities.Trade{Time: tm, Price: high[14]}
		check(t, dip.UpdateTrade(&r))
	})
}

func TestDirectionalIndicatorPlusMetadata(t *testing.T) {
	t.Parallel()

	dip, _ := NewDirectionalIndicatorPlus(14)
	act := dip.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	check("Identifier", core.DirectionalIndicatorPlus, act.Identifier)
	check("Mnemonic", "+di", act.Mnemonic)
	check("Description", "Directional Indicator Plus", act.Description)
	check("len(Outputs)", 4, len(act.Outputs))
	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", "+di", act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", "Directional Indicator Plus", act.Outputs[0].Description)
	check("Outputs[1].Kind", int(DirectionalMovementPlus), act.Outputs[1].Kind)
	check("Outputs[1].Shape", shape.Scalar, act.Outputs[1].Shape)
	check("Outputs[2].Kind", int(AverageTrueRange), act.Outputs[2].Kind)
	check("Outputs[2].Shape", shape.Scalar, act.Outputs[2].Shape)
	check("Outputs[3].Kind", int(TrueRange), act.Outputs[3].Kind)
	check("Outputs[3].Shape", shape.Scalar, act.Outputs[3].Shape)
}
