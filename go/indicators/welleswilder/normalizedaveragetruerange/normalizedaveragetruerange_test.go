//nolint:testpackage
package normalizedaveragetruerange

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func TestNormalizedAverageTrueRangeConstructor(t *testing.T) {
	t.Parallel()

	natr, err := NewNormalizedAverageTrueRange(14)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if natr.Length() != 14 {
		t.Errorf("expected length 14, got %d", natr.Length())
	}

	if natr.IsPrimed() {
		t.Error("should not be primed initially")
	}

	_, err = NewNormalizedAverageTrueRange(0)
	if err == nil {
		t.Error("expected error for length 0")
	}

	_, err = NewNormalizedAverageTrueRange(-8)
	if err == nil {
		t.Error("expected error for negative length")
	}
}

func TestNormalizedAverageTrueRangeIsPrimed(t *testing.T) {
	t.Parallel()

	high := testInputHigh()
	low := testInputLow()
	cls := testInputClose()
	natr, _ := NewNormalizedAverageTrueRange(5)

	if natr.IsPrimed() {
		t.Error("should not be primed before updates")
	}

	for i := 0; i < 5; i++ {
		natr.Update(cls[i], high[i], low[i])

		if natr.IsPrimed() {
			t.Errorf("[%d] should not be primed yet", i)
		}
	}

	for i := 5; i < 10; i++ {
		natr.Update(cls[i], high[i], low[i])

		if !natr.IsPrimed() {
			t.Errorf("[%d] should be primed", i)
		}
	}
}

func TestNormalizedAverageTrueRangeUpdate(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-11

	high := testInputHigh()
	low := testInputLow()
	cls := testInputClose()
	expected := testExpectedNATR()
	natr, _ := NewNormalizedAverageTrueRange(14)

	for i := range cls {
		act := natr.Update(cls[i], high[i], low[i])

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

func TestNormalizedAverageTrueRangeLength1(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-11

	high := testInputHigh()
	low := testInputLow()
	cls := testInputClose()
	expected := testExpectedNATR1()
	natr, _ := NewNormalizedAverageTrueRange(1)

	for i := range cls {
		act := natr.Update(cls[i], high[i], low[i])

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

func TestNormalizedAverageTrueRangeCloseZero(t *testing.T) {
	t.Parallel()

	natr, _ := NewNormalizedAverageTrueRange(14)

	high := testInputHigh()
	low := testInputLow()
	cls := testInputClose()

	// Prime the indicator.
	for i := 0; i < 15; i++ {
		natr.Update(cls[i], high[i], low[i])
	}

	// close=0 should return 0, not panic or NaN.
	result := natr.Update(0, 3.3, 2.2)
	if result != 0 {
		t.Errorf("expected 0 for close=0, got %v", result)
	}
}

func TestNormalizedAverageTrueRangeNaNPassthrough(t *testing.T) {
	t.Parallel()

	natr, _ := NewNormalizedAverageTrueRange(14)

	if !math.IsNaN(natr.Update(math.NaN(), 1, 1)) {
		t.Error("expected NaN passthrough for NaN close")
	}

	if !math.IsNaN(natr.Update(1, math.NaN(), 1)) {
		t.Error("expected NaN passthrough for NaN high")
	}

	if !math.IsNaN(natr.Update(1, 1, math.NaN())) {
		t.Error("expected NaN passthrough for NaN low")
	}

	if !math.IsNaN(natr.UpdateSample(math.NaN())) {
		t.Error("expected NaN passthrough for NaN sample")
	}
}

func TestNormalizedAverageTrueRangeUpdateEntity(t *testing.T) {
	t.Parallel()

	tm := testTime()
	natr, _ := NewNormalizedAverageTrueRange(14)

	// Prime with enough bars.
	high := testInputHigh()
	low := testInputLow()
	cls := testInputClose()

	for i := 0; i < 15; i++ {
		natr.Update(cls[i], high[i], low[i])
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

		n2, _ := NewNormalizedAverageTrueRange(14)
		for i := 0; i < 15; i++ {
			n2.Update(cls[i], high[i], low[i])
		}

		b := entities.Bar{Time: tm, High: high[15], Low: low[15], Close: cls[15]}
		check(n2.UpdateBar(&b))
	})

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		n2, _ := NewNormalizedAverageTrueRange(14)
		for i := 0; i < 15; i++ {
			n2.Update(cls[i], high[i], low[i])
		}

		s := entities.Scalar{Time: tm, Value: cls[15]}
		check(n2.UpdateScalar(&s))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		n2, _ := NewNormalizedAverageTrueRange(14)
		for i := 0; i < 15; i++ {
			n2.Update(cls[i], high[i], low[i])
		}

		q := entities.Quote{Time: tm, Bid: cls[15] - 0.5, Ask: cls[15] + 0.5}
		check(n2.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		n2, _ := NewNormalizedAverageTrueRange(14)
		for i := 0; i < 15; i++ {
			n2.Update(cls[i], high[i], low[i])
		}

		r := entities.Trade{Time: tm, Price: cls[15]}
		check(n2.UpdateTrade(&r))
	})
}

func TestNormalizedAverageTrueRangeMetadata(t *testing.T) {
	t.Parallel()

	natr, _ := NewNormalizedAverageTrueRange(14)
	act := natr.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	check("Identifier", core.NormalizedAverageTrueRange, act.Identifier)
	check("Mnemonic", "natr", act.Mnemonic)
	check("Description", "Normalized Average True Range", act.Description)
	check("len(Outputs)", 1, len(act.Outputs))
	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", "natr", act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", "Normalized Average True Range", act.Outputs[0].Description)
}
