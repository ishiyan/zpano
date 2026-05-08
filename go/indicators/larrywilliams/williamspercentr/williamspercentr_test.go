//nolint:testpackage
package williamspercentr

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func TestWilliamsPercentRUpdate14(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-6

	high := testInputHigh()
	low := testInputLow()
	close := testInputClose()
	expected := testExpected14()
	w := NewWilliamsPercentR(14)

	for i := range close {
		act := w.Update(close[i], high[i], low[i])

		if i < 13 {
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

func TestWilliamsPercentRUpdate2(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-6

	high := testInputHigh()
	low := testInputLow()
	close := testInputClose()
	expected := testExpected2()
	w := NewWilliamsPercentR(2)

	for i := range close {
		act := w.Update(close[i], high[i], low[i])

		if i < 1 {
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

func TestWilliamsPercentRNaNPassthrough(t *testing.T) {
	t.Parallel()

	w := NewWilliamsPercentR(14)

	if !math.IsNaN(w.Update(math.NaN(), 1, 1)) {
		t.Error("expected NaN passthrough for NaN close")
	}

	if !math.IsNaN(w.Update(1, math.NaN(), 1)) {
		t.Error("expected NaN passthrough for NaN high")
	}

	if !math.IsNaN(w.Update(1, 1, math.NaN())) {
		t.Error("expected NaN passthrough for NaN low")
	}
}

func TestWilliamsPercentRIsPrimed(t *testing.T) {
	t.Parallel()

	high := testInputHigh()
	low := testInputLow()
	close := testInputClose()
	w := NewWilliamsPercentR(14)

	if w.IsPrimed() {
		t.Error("should not be primed before any updates")
	}

	for i := 0; i < 13; i++ {
		w.Update(close[i], high[i], low[i])

		if w.IsPrimed() {
			t.Errorf("[%d] should not be primed yet", i)
		}
	}

	w.Update(close[13], high[13], low[13])

	if !w.IsPrimed() {
		t.Error("[13] should be primed after 14th update")
	}

	w.Update(close[14], high[14], low[14])

	if !w.IsPrimed() {
		t.Error("[14] should remain primed")
	}
}

func TestWilliamsPercentRUpdateSample(t *testing.T) {
	t.Parallel()

	w := NewWilliamsPercentR(14)

	// When H=L=C are all the same, after priming all windows have same value, so %R = 0.
	for i := 0; i < 13; i++ {
		v := w.UpdateSample(9.0)
		if !math.IsNaN(v) {
			t.Errorf("[%d] expected NaN, got %v", i, v)
		}
	}

	v := w.UpdateSample(9.0)
	if v != 0.0 {
		t.Errorf("expected 0, got %v", v)
	}
}

func TestWilliamsPercentRUpdateEntity(t *testing.T) {
	t.Parallel()

	tm := testTime()
	high := testInputHigh()
	low := testInputLow()
	close := testInputClose()

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
	}

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		w := NewWilliamsPercentR(14)
		for i := 0; i < 14; i++ {
			w.Update(close[i], high[i], low[i])
		}

		b := entities.Bar{Time: tm, High: high[14], Low: low[14], Close: close[14]}
		check(w.UpdateBar(&b))
	})

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		w := NewWilliamsPercentR(14)
		for i := 0; i < 14; i++ {
			w.Update(close[i], high[i], low[i])
		}

		s := entities.Scalar{Time: tm, Value: 100}
		check(w.UpdateScalar(&s))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		w := NewWilliamsPercentR(14)
		for i := 0; i < 14; i++ {
			w.Update(close[i], high[i], low[i])
		}

		q := entities.Quote{Time: tm, Bid: 99, Ask: 101}
		check(w.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		w := NewWilliamsPercentR(14)
		for i := 0; i < 14; i++ {
			w.Update(close[i], high[i], low[i])
		}

		r := entities.Trade{Time: tm, Price: 100}
		check(w.UpdateTrade(&r))
	})
}

func TestWilliamsPercentRMetadata(t *testing.T) {
	t.Parallel()

	w := NewWilliamsPercentR(14)
	act := w.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	check("Identifier", core.WilliamsPercentR, act.Identifier)
	check("Mnemonic", "willr", act.Mnemonic)
	check("Description", "Williams %R", act.Description)
	check("len(Outputs)", 1, len(act.Outputs))
	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", "willr", act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", "Williams %R", act.Outputs[0].Description)
}
