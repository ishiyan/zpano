//nolint:testpackage
package directionalmovementplus

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func TestDirectionalMovementPlusConstructor(t *testing.T) {
	t.Parallel()

	dmp, err := NewDirectionalMovementPlus(14)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dmp.Length() != 14 {
		t.Errorf("expected length 14, got %d", dmp.Length())
	}

	if dmp.IsPrimed() {
		t.Error("should not be primed initially")
	}

	_, err = NewDirectionalMovementPlus(0)
	if err == nil {
		t.Error("expected error for length 0")
	}

	_, err = NewDirectionalMovementPlus(-8)
	if err == nil {
		t.Error("expected error for negative length")
	}
}

func TestDirectionalMovementPlusIsPrimed(t *testing.T) {
	t.Parallel()

	high := testInputHigh()
	low := testInputLow()

	t.Run("length=1", func(t *testing.T) {
		t.Parallel()

		dmp, _ := NewDirectionalMovementPlus(1)

		if dmp.IsPrimed() {
			t.Error("should not be primed before updates")
		}

		// First update: not primed yet.
		dmp.Update(high[0], low[0])
		if dmp.IsPrimed() {
			t.Error("[0] should not be primed yet")
		}

		// Second update: should be primed.
		dmp.Update(high[1], low[1])
		if !dmp.IsPrimed() {
			t.Error("[1] should be primed")
		}
	})

	t.Run("length=14", func(t *testing.T) {
		t.Parallel()

		dmp, _ := NewDirectionalMovementPlus(14)

		for i := 0; i < 14; i++ {
			dmp.Update(high[i], low[i])

			if dmp.IsPrimed() {
				t.Errorf("[%d] should not be primed yet", i)
			}
		}

		dmp.Update(high[14], low[14])
		if !dmp.IsPrimed() {
			t.Error("[14] should be primed")
		}
	})
}

func TestDirectionalMovementPlusUpdate(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-8

	high := testInputHigh()
	low := testInputLow()
	expected := testExpectedDMP14()
	dmp, _ := NewDirectionalMovementPlus(14)

	for i := range high {
		act := dmp.Update(high[i], low[i])

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

func TestDirectionalMovementPlusLength1(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-8

	high := testInputHigh()
	low := testInputLow()
	expected := testExpectedDMP1()
	dmp, _ := NewDirectionalMovementPlus(1)

	for i := range high {
		act := dmp.Update(high[i], low[i])

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

func TestDirectionalMovementPlusNaNPassthrough(t *testing.T) {
	t.Parallel()

	dmp, _ := NewDirectionalMovementPlus(14)

	if !math.IsNaN(dmp.Update(math.NaN(), 1)) {
		t.Error("expected NaN passthrough for NaN high")
	}

	if !math.IsNaN(dmp.Update(1, math.NaN())) {
		t.Error("expected NaN passthrough for NaN low")
	}

	if !math.IsNaN(dmp.Update(math.NaN(), math.NaN())) {
		t.Error("expected NaN passthrough for NaN high and low")
	}

	if !math.IsNaN(dmp.UpdateSample(math.NaN())) {
		t.Error("expected NaN passthrough for NaN sample")
	}
}

func TestDirectionalMovementPlusHighLowSwap(t *testing.T) {
	t.Parallel()

	// When high < low, they should be swapped internally.
	dmp1, _ := NewDirectionalMovementPlus(1)
	dmp2, _ := NewDirectionalMovementPlus(1)

	// Prime both.
	dmp1.Update(10, 5)
	dmp2.Update(5, 10) // Swapped.

	// Update both with same effective values.
	v1 := dmp1.Update(12, 6)
	v2 := dmp2.Update(6, 12) // Swapped.

	if v1 != v2 {
		t.Errorf("high/low swap should produce same result: %v vs %v", v1, v2)
	}
}

func TestDirectionalMovementPlusZeroInputs(t *testing.T) {
	t.Parallel()

	// Updating with same value repeatedly should produce 0.
	dmp, _ := NewDirectionalMovementPlus(10)

	for i := 0; i < 20; i++ {
		dmp.UpdateSample(0)
	}

	if !dmp.IsPrimed() {
		t.Error("should be primed after 20 updates with length 10")
	}
}

func TestDirectionalMovementPlusUpdateEntity(t *testing.T) {
	t.Parallel()

	tm := testTime()
	high := testInputHigh()
	low := testInputLow()

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

		dmp, _ := NewDirectionalMovementPlus(14)
		for i := 0; i < 14; i++ {
			dmp.Update(high[i], low[i])
		}

		b := entities.Bar{Time: tm, High: high[14], Low: low[14]}
		check(t, dmp.UpdateBar(&b))
	})

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		dmp, _ := NewDirectionalMovementPlus(14)
		for i := 0; i < 14; i++ {
			dmp.Update(high[i], low[i])
		}

		s := entities.Scalar{Time: tm, Value: high[14]}
		check(t, dmp.UpdateScalar(&s))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		dmp, _ := NewDirectionalMovementPlus(14)
		for i := 0; i < 14; i++ {
			dmp.Update(high[i], low[i])
		}

		q := entities.Quote{Time: tm, Bid: high[14] - 0.5, Ask: high[14] + 0.5}
		check(t, dmp.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		dmp, _ := NewDirectionalMovementPlus(14)
		for i := 0; i < 14; i++ {
			dmp.Update(high[i], low[i])
		}

		r := entities.Trade{Time: tm, Price: high[14]}
		check(t, dmp.UpdateTrade(&r))
	})
}

func TestDirectionalMovementPlusMetadata(t *testing.T) {
	t.Parallel()

	dmp, _ := NewDirectionalMovementPlus(14)
	act := dmp.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	check("Identifier", core.DirectionalMovementPlus, act.Identifier)
	check("Mnemonic", "+dm", act.Mnemonic)
	check("Description", "Directional Movement Plus", act.Description)
	check("len(Outputs)", 1, len(act.Outputs))
	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", "+dm", act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", "Directional Movement Plus", act.Outputs[0].Description)
}
