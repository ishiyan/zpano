//nolint:testpackage
package averagedirectionalmovementindexrating

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func TestAverageDirectionalMovementIndexRatingConstructor(t *testing.T) {
	t.Parallel()

	adxr, err := NewAverageDirectionalMovementIndexRating(14)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if adxr.Length() != 14 {
		t.Errorf("expected length 14, got %d", adxr.Length())
	}

	if adxr.IsPrimed() {
		t.Error("should not be primed initially")
	}

	_, err = NewAverageDirectionalMovementIndexRating(0)
	if err == nil {
		t.Error("expected error for length 0")
	}

	_, err = NewAverageDirectionalMovementIndexRating(-8)
	if err == nil {
		t.Error("expected error for negative length")
	}
}

func TestAverageDirectionalMovementIndexRatingIsPrimed(t *testing.T) {
	t.Parallel()

	high := testInputHigh()
	low := testInputLow()
	close_ := testInputClose()

	t.Run("length=14", func(t *testing.T) {
		t.Parallel()

		adxr, _ := NewAverageDirectionalMovementIndexRating(14)

		// ADX primes at index 27. ADXR needs (length-1)=13 more ADX values after that,
		// so ADXR primes at index 40.
		for i := 0; i < 40; i++ {
			adxr.Update(close_[i], high[i], low[i])

			if adxr.IsPrimed() {
				t.Errorf("[%d] should not be primed yet", i)
			}
		}

		adxr.Update(close_[40], high[40], low[40])
		if !adxr.IsPrimed() {
			t.Error("[40] should be primed")
		}
	})
}

func TestAverageDirectionalMovementIndexRatingUpdate(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-8

	high := testInputHigh()
	low := testInputLow()
	close_ := testInputClose()
	expected := testExpectedADXR14()
	adxr, _ := NewAverageDirectionalMovementIndexRating(14)

	for i := range high {
		act := adxr.Update(close_[i], high[i], low[i])

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

func TestAverageDirectionalMovementIndexRatingNaNPassthrough(t *testing.T) {
	t.Parallel()

	adxr, _ := NewAverageDirectionalMovementIndexRating(14)

	if !math.IsNaN(adxr.Update(math.NaN(), 1, 1)) {
		t.Error("expected NaN passthrough for NaN close")
	}

	if !math.IsNaN(adxr.Update(1, math.NaN(), 1)) {
		t.Error("expected NaN passthrough for NaN high")
	}

	if !math.IsNaN(adxr.Update(1, 1, math.NaN())) {
		t.Error("expected NaN passthrough for NaN low")
	}

	if !math.IsNaN(adxr.UpdateSample(math.NaN())) {
		t.Error("expected NaN passthrough for NaN sample")
	}
}

func TestAverageDirectionalMovementIndexRatingUpdateEntity(t *testing.T) {
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

		adxr, _ := NewAverageDirectionalMovementIndexRating(14)
		for i := 0; i < 40; i++ {
			adxr.Update(close_[i], high[i], low[i])
		}

		b := entities.Bar{Time: tm, Close: close_[40], High: high[40], Low: low[40]}
		check(t, adxr.UpdateBar(&b))
	})

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		adxr, _ := NewAverageDirectionalMovementIndexRating(14)
		for i := 0; i < 40; i++ {
			adxr.Update(close_[i], high[i], low[i])
		}

		s := entities.Scalar{Time: tm, Value: high[40]}
		check(t, adxr.UpdateScalar(&s))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		adxr, _ := NewAverageDirectionalMovementIndexRating(14)
		for i := 0; i < 40; i++ {
			adxr.Update(close_[i], high[i], low[i])
		}

		q := entities.Quote{Time: tm, Bid: high[40] - 0.5, Ask: high[40] + 0.5}
		check(t, adxr.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		adxr, _ := NewAverageDirectionalMovementIndexRating(14)
		for i := 0; i < 40; i++ {
			adxr.Update(close_[i], high[i], low[i])
		}

		r := entities.Trade{Time: tm, Price: high[40]}
		check(t, adxr.UpdateTrade(&r))
	})
}

func TestAverageDirectionalMovementIndexRatingMetadata(t *testing.T) {
	t.Parallel()

	adxr, _ := NewAverageDirectionalMovementIndexRating(14)
	act := adxr.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	check("Identifier", core.AverageDirectionalMovementIndexRating, act.Identifier)
	check("Mnemonic", "adxr", act.Mnemonic)
	check("Description", "Average Directional Movement Index Rating", act.Description)
	check("len(Outputs)", 9, len(act.Outputs))
	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", "adxr", act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", "Average Directional Movement Index Rating", act.Outputs[0].Description)
	check("Outputs[1].Kind", int(AverageDirectionalMovementIndex), act.Outputs[1].Kind)
	check("Outputs[1].Shape", shape.Scalar, act.Outputs[1].Shape)
	check("Outputs[2].Kind", int(DirectionalMovementIndex), act.Outputs[2].Kind)
	check("Outputs[2].Shape", shape.Scalar, act.Outputs[2].Shape)
	check("Outputs[3].Kind", int(DirectionalIndicatorPlus), act.Outputs[3].Kind)
	check("Outputs[3].Shape", shape.Scalar, act.Outputs[3].Shape)
	check("Outputs[4].Kind", int(DirectionalIndicatorMinus), act.Outputs[4].Kind)
	check("Outputs[4].Shape", shape.Scalar, act.Outputs[4].Shape)
	check("Outputs[5].Kind", int(DirectionalMovementPlus), act.Outputs[5].Kind)
	check("Outputs[5].Shape", shape.Scalar, act.Outputs[5].Shape)
	check("Outputs[6].Kind", int(DirectionalMovementMinus), act.Outputs[6].Kind)
	check("Outputs[6].Shape", shape.Scalar, act.Outputs[6].Shape)
	check("Outputs[7].Kind", int(AverageTrueRange), act.Outputs[7].Kind)
	check("Outputs[7].Shape", shape.Scalar, act.Outputs[7].Shape)
	check("Outputs[8].Kind", int(TrueRange), act.Outputs[8].Kind)
	check("Outputs[8].Shape", shape.Scalar, act.Outputs[8].Shape)
}
