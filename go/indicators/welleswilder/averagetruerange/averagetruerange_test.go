//nolint:testpackage
package averagetruerange

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func TestAverageTrueRangeConstructor(t *testing.T) {
	t.Parallel()

	atr, err := NewAverageTrueRange(14)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if atr.Length() != 14 {
		t.Errorf("expected length 14, got %d", atr.Length())
	}

	if atr.IsPrimed() {
		t.Error("should not be primed initially")
	}

	_, err = NewAverageTrueRange(0)
	if err == nil {
		t.Error("expected error for length 0")
	}

	_, err = NewAverageTrueRange(-8)
	if err == nil {
		t.Error("expected error for negative length")
	}
}

func TestAverageTrueRangeIsPrimed(t *testing.T) {
	t.Parallel()

	high := testInputHigh()
	low := testInputLow()
	cls := testInputClose()
	atr, _ := NewAverageTrueRange(5)

	if atr.IsPrimed() {
		t.Error("should not be primed before updates")
	}

	for i := 0; i < 5; i++ {
		atr.Update(cls[i], high[i], low[i])

		if atr.IsPrimed() {
			t.Errorf("[%d] should not be primed yet", i)
		}
	}

	for i := 5; i < 10; i++ {
		atr.Update(cls[i], high[i], low[i])

		if !atr.IsPrimed() {
			t.Errorf("[%d] should be primed", i)
		}
	}
}

func TestAverageTrueRangeUpdate(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-12

	high := testInputHigh()
	low := testInputLow()
	cls := testInputClose()
	expected := testExpectedATR()
	atr, _ := NewAverageTrueRange(14)

	for i := range cls {
		act := atr.Update(cls[i], high[i], low[i])

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

func TestAverageTrueRangeLength1(t *testing.T) {
	t.Parallel()

	// ATR with length=1 should equal TR values (after priming).
	const tolerance = 1e-3

	high := testInputHigh()
	low := testInputLow()
	cls := testInputClose()
	atr, _ := NewAverageTrueRange(1)

	// Expected TR values (same as TrueRange test).
	expectedTR := []float64{
		math.NaN(), 3.535, 2.125, 2.69, 3.185, 1.22, 3.0, 3.97, 3.31, 2.435,
		3.78, 3.5, 3.095, 9.685, 4.565, 2.31, 4.5, 1.875, 2.72, 2.5,
		2.845, 1.97, 3.625, 3.22, 2.875, 3.875, 3.19, 5.34, 3.655, 3.155,
		2.75, 2.155, 1.875, 3.44, 2.125, 3.28, 2.315, 3.565, 2.31, 2.03,
		1.94, 5.125, 3.97, 1.47, 3.16, 1.315, 2.22, 2.72, 2.59, 1.655,
		1.5, 2.56, 5.0, 1.935, 1.815, 2.56, 1.875, 2.66, 3.185, 2.185,
		2.5, 1.5, 3.47, 2.28, 4.285, 2.875, 2.03, 2.625, 2.03, 3.125,
		3.0, 3.81, 4.125, 3.375, 3.375, 13.435, 6.81, 5.5, 3.375, 4.25,
		2.78, 3.78, 2.97, 2.25, 2.595, 3.0, 4.125, 2.41, 2.19, 6.47,
		10.25, 5.0, 3.815, 1.815, 2.845, 1.94, 2.095, 4.47, 2.5, 7.72,
		5.505, 2.56, 4.81, 5.18, 3.75, 3.56, 5.69, 4.44, 4.31, 3.69,
		3.94, 2.0, 3.56, 5.63, 3.0, 1.69, 5.0, 4.75, 2.44, 2.88,
		3.19, 2.5, 2.75, 7.63, 3.31, 2.75, 3.07, 2.87, 3.37, 3.56,
		3.31, 3.44, 2.31, 3.63, 2.12, 5.19, 8.57, 2.88, 5.5, 2.82,
		2.37, 3.5, 2.88, 3.76, 2.32, 4.19, 4.93, 3.81, 6.44, 6.0,
		3.32, 3.56, 4.44, 2.81, 3.38, 4.5, 2.0, 6.13, 3.62, 3.31,
		3.12, 3.44, 3.19, 3.81, 2.44, 2.93, 3.13, 3.94, 3.0, 3.88,
		4.31, 3.25, 5.75, 3.31, 3.56, 1.62, 3.06, 2.82, 7.69, 5.51,
		3.88, 2.88, 6.37, 4.56, 5.44, 3.06, 3.5, 2.63, 5.56, 3.25,
		4.19, 3.69, 4.13, 5.31, 3.57, 4.87, 5.87, 2.56, 5.0, 4.31,
		3.06, 6.06, 18.0, 3.62, 3.0, 2.13, 2.75, 2.31, 4.06, 2.44,
		3.12, 2.44, 4.44, 2.75, 3.69, 3.38, 3.44, 2.63, 3.25, 2.5,
		2.0, 2.25, 4.69, 7.12, 4.5, 3.87, 4.25, 1.88, 1.63, 2.38,
		2.19, 2.94, 7.6, 4.63, 3.75, 5.5, 9.87, 5.81, 6.19, 3.32,
		4.75, 3.94, 2.44, 2.69, 2.06, 2.31, 2.44, 1.88, 1.69, 1.75,
		1.94, 2.88,
	}

	for i := range cls {
		act := atr.Update(cls[i], high[i], low[i])

		if math.IsNaN(expectedTR[i]) {
			if !math.IsNaN(act) {
				t.Errorf("[%d] expected NaN, got %v", i, act)
			}

			continue
		}

		if math.IsNaN(act) {
			t.Errorf("[%d] expected %v, got NaN", i, expectedTR[i])
			continue
		}

		if math.Abs(act-expectedTR[i]) > tolerance {
			t.Errorf("[%d] expected %v, got %v", i, expectedTR[i], act)
		}
	}
}

func TestAverageTrueRangeNaNPassthrough(t *testing.T) {
	t.Parallel()

	atr, _ := NewAverageTrueRange(14)

	if !math.IsNaN(atr.Update(math.NaN(), 1, 1)) {
		t.Error("expected NaN passthrough for NaN close")
	}

	if !math.IsNaN(atr.Update(1, math.NaN(), 1)) {
		t.Error("expected NaN passthrough for NaN high")
	}

	if !math.IsNaN(atr.Update(1, 1, math.NaN())) {
		t.Error("expected NaN passthrough for NaN low")
	}

	if !math.IsNaN(atr.UpdateSample(math.NaN())) {
		t.Error("expected NaN passthrough for NaN sample")
	}
}

func TestAverageTrueRangeUpdateEntity(t *testing.T) {
	t.Parallel()

	tm := testTime()
	atr, _ := NewAverageTrueRange(14)

	// Prime with enough bars.
	high := testInputHigh()
	low := testInputLow()
	cls := testInputClose()

	for i := 0; i < 14; i++ {
		atr.Update(cls[i], high[i], low[i])
	}

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

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		a2, _ := NewAverageTrueRange(14)
		for i := 0; i < 14; i++ {
			a2.Update(cls[i], high[i], low[i])
		}

		b := entities.Bar{Time: tm, High: high[14], Low: low[14], Close: cls[14]}
		check(a2.UpdateBar(&b))
	})

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		a2, _ := NewAverageTrueRange(14)
		for i := 0; i < 14; i++ {
			a2.Update(cls[i], high[i], low[i])
		}

		s := entities.Scalar{Time: tm, Value: cls[14]}
		check(a2.UpdateScalar(&s))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		a2, _ := NewAverageTrueRange(14)
		for i := 0; i < 14; i++ {
			a2.Update(cls[i], high[i], low[i])
		}

		q := entities.Quote{Time: tm, Bid: cls[14] - 0.5, Ask: cls[14] + 0.5}
		check(a2.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		a2, _ := NewAverageTrueRange(14)
		for i := 0; i < 14; i++ {
			a2.Update(cls[i], high[i], low[i])
		}

		r := entities.Trade{Time: tm, Price: cls[14]}
		check(a2.UpdateTrade(&r))
	})
}

func TestAverageTrueRangeMetadata(t *testing.T) {
	t.Parallel()

	atr, _ := NewAverageTrueRange(14)
	act := atr.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	check("Identifier", core.AverageTrueRange, act.Identifier)
	check("Mnemonic", "atr", act.Mnemonic)
	check("Description", "Average True Range", act.Description)
	check("len(Outputs)", 1, len(act.Outputs))
	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", "atr", act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", "Average True Range", act.Outputs[0].Description)
}
