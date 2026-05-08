//nolint:testpackage
package averagedirectionalmovementindex

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func TestAverageDirectionalMovementIndexConstructor(t *testing.T) {
	t.Parallel()

	adx, err := NewAverageDirectionalMovementIndex(14)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if adx.Length() != 14 {
		t.Errorf("expected length 14, got %d", adx.Length())
	}

	if adx.IsPrimed() {
		t.Error("should not be primed initially")
	}

	_, err = NewAverageDirectionalMovementIndex(0)
	if err == nil {
		t.Error("expected error for length 0")
	}

	_, err = NewAverageDirectionalMovementIndex(-8)
	if err == nil {
		t.Error("expected error for negative length")
	}
}

func TestAverageDirectionalMovementIndexIsPrimed(t *testing.T) {
	t.Parallel()

	high := testInputHigh()
	low := testInputLow()
	close_ := testInputClose()

	t.Run("length=14", func(t *testing.T) {
		t.Parallel()

		adx, _ := NewAverageDirectionalMovementIndex(14)

		// ADX primes after DX primes (at index 14) + length more DX values.
		// DX primes at index 14 (after 15 updates). Then ADX needs 14 DX values: indices 14..27.
		// So ADX primes at index 27 (after 28 updates).
		for i := 0; i < 27; i++ {
			adx.Update(close_[i], high[i], low[i])

			if adx.IsPrimed() {
				t.Errorf("[%d] should not be primed yet", i)
			}
		}

		adx.Update(close_[27], high[27], low[27])
		if !adx.IsPrimed() {
			t.Error("[27] should be primed")
		}
	})
}

func TestAverageDirectionalMovementIndexUpdate(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-8

	high := testInputHigh()
	low := testInputLow()
	close_ := testInputClose()
	expected := testExpectedADX14()
	adx, _ := NewAverageDirectionalMovementIndex(14)

	for i := range high {
		act := adx.Update(close_[i], high[i], low[i])

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

func TestAverageDirectionalMovementIndexNaNPassthrough(t *testing.T) {
	t.Parallel()

	adx, _ := NewAverageDirectionalMovementIndex(14)

	if !math.IsNaN(adx.Update(math.NaN(), 1, 1)) {
		t.Error("expected NaN passthrough for NaN close")
	}

	if !math.IsNaN(adx.Update(1, math.NaN(), 1)) {
		t.Error("expected NaN passthrough for NaN high")
	}

	if !math.IsNaN(adx.Update(1, 1, math.NaN())) {
		t.Error("expected NaN passthrough for NaN low")
	}

	if !math.IsNaN(adx.UpdateSample(math.NaN())) {
		t.Error("expected NaN passthrough for NaN sample")
	}
}

func TestAverageDirectionalMovementIndexUpdateEntity(t *testing.T) {
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

		adx, _ := NewAverageDirectionalMovementIndex(14)
		for i := 0; i < 27; i++ {
			adx.Update(close_[i], high[i], low[i])
		}

		b := entities.Bar{Time: tm, Close: close_[27], High: high[27], Low: low[27]}
		check(t, adx.UpdateBar(&b))
	})

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		adx, _ := NewAverageDirectionalMovementIndex(14)
		for i := 0; i < 27; i++ {
			adx.Update(close_[i], high[i], low[i])
		}

		s := entities.Scalar{Time: tm, Value: high[27]}
		check(t, adx.UpdateScalar(&s))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		adx, _ := NewAverageDirectionalMovementIndex(14)
		for i := 0; i < 27; i++ {
			adx.Update(close_[i], high[i], low[i])
		}

		q := entities.Quote{Time: tm, Bid: high[27] - 0.5, Ask: high[27] + 0.5}
		check(t, adx.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		adx, _ := NewAverageDirectionalMovementIndex(14)
		for i := 0; i < 27; i++ {
			adx.Update(close_[i], high[i], low[i])
		}

		r := entities.Trade{Time: tm, Price: high[27]}
		check(t, adx.UpdateTrade(&r))
	})
}

func TestAverageDirectionalMovementIndexMetadata(t *testing.T) {
	t.Parallel()

	adx, _ := NewAverageDirectionalMovementIndex(14)
	act := adx.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	check("Identifier", core.AverageDirectionalMovementIndex, act.Identifier)
	check("Mnemonic", "adx", act.Mnemonic)
	check("Description", "Average Directional Movement Index", act.Description)
	check("len(Outputs)", 8, len(act.Outputs))
	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", "adx", act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", "Average Directional Movement Index", act.Outputs[0].Description)
	check("Outputs[1].Kind", int(DirectionalMovementIndex), act.Outputs[1].Kind)
	check("Outputs[1].Shape", shape.Scalar, act.Outputs[1].Shape)
	check("Outputs[2].Kind", int(DirectionalIndicatorPlus), act.Outputs[2].Kind)
	check("Outputs[2].Shape", shape.Scalar, act.Outputs[2].Shape)
	check("Outputs[3].Kind", int(DirectionalIndicatorMinus), act.Outputs[3].Kind)
	check("Outputs[3].Shape", shape.Scalar, act.Outputs[3].Shape)
	check("Outputs[4].Kind", int(DirectionalMovementPlus), act.Outputs[4].Kind)
	check("Outputs[4].Shape", shape.Scalar, act.Outputs[4].Shape)
	check("Outputs[5].Kind", int(DirectionalMovementMinus), act.Outputs[5].Kind)
	check("Outputs[5].Shape", shape.Scalar, act.Outputs[5].Shape)
	check("Outputs[6].Kind", int(AverageTrueRange), act.Outputs[6].Kind)
	check("Outputs[6].Shape", shape.Scalar, act.Outputs[6].Shape)
	check("Outputs[7].Kind", int(TrueRange), act.Outputs[7].Kind)
	check("Outputs[7].Shape", shape.Scalar, act.Outputs[7].Shape)
}
