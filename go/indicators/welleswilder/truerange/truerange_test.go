//nolint:testpackage
package truerange

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func TestTrueRangeUpdate(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-3

	high := testInputHigh()
	low := testInputLow()
	close := testInputClose()
	expected := testExpectedTR()
	tr := NewTrueRange()

	for i := range close {
		act := tr.Update(close[i], high[i], low[i])

		if i == 0 {
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

func TestTrueRangeNaNPassthrough(t *testing.T) {
	t.Parallel()

	tr := NewTrueRange()

	if !math.IsNaN(tr.Update(math.NaN(), 1, 1)) {
		t.Error("expected NaN passthrough for NaN close")
	}

	if !math.IsNaN(tr.Update(1, math.NaN(), 1)) {
		t.Error("expected NaN passthrough for NaN high")
	}

	if !math.IsNaN(tr.Update(1, 1, math.NaN())) {
		t.Error("expected NaN passthrough for NaN low")
	}
}

func TestTrueRangeIsPrimed(t *testing.T) {
	t.Parallel()

	high := testInputHigh()
	low := testInputLow()
	close := testInputClose()
	tr := NewTrueRange()

	if tr.IsPrimed() {
		t.Error("should not be primed before any updates")
	}

	tr.Update(close[0], high[0], low[0])

	if tr.IsPrimed() {
		t.Error("[0] should not be primed after first update")
	}

	tr.Update(close[1], high[1], low[1])

	if !tr.IsPrimed() {
		t.Error("[1] should be primed after second update")
	}

	tr.Update(close[2], high[2], low[2])

	if !tr.IsPrimed() {
		t.Error("[2] should remain primed")
	}
}

func TestTrueRangeUpdateSample(t *testing.T) {
	t.Parallel()

	// When using a single value as H=L=C, TR = |current - previous| after priming.
	tr := NewTrueRange()

	v := tr.UpdateSample(100.0)
	if !math.IsNaN(v) {
		t.Errorf("expected NaN, got %v", v)
	}

	v = tr.UpdateSample(105.0)
	if math.Abs(v-5.0) > 1e-10 {
		t.Errorf("expected 5.0, got %v", v)
	}

	v = tr.UpdateSample(102.0)
	if math.Abs(v-3.0) > 1e-10 {
		t.Errorf("expected 3.0, got %v", v)
	}
}

func TestTrueRangeUpdateEntity(t *testing.T) {
	t.Parallel()

	tm := testTime()
	tr := NewTrueRange()

	// Prime with first bar.
	tr.Update(100, 105, 95)

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

		tr2 := NewTrueRange()
		tr2.Update(100, 105, 95)

		b := entities.Bar{Time: tm, High: 110, Low: 98, Close: 108}
		check(tr2.UpdateBar(&b))
	})

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		tr2 := NewTrueRange()
		tr2.Update(100, 105, 95)

		s := entities.Scalar{Time: tm, Value: 108}
		check(tr2.UpdateScalar(&s))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		tr2 := NewTrueRange()
		tr2.Update(100, 105, 95)

		q := entities.Quote{Time: tm, Bid: 107, Ask: 109}
		check(tr2.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		tr2 := NewTrueRange()
		tr2.Update(100, 105, 95)

		r := entities.Trade{Time: tm, Price: 108}
		check(tr2.UpdateTrade(&r))
	})
}

func TestTrueRangeMetadata(t *testing.T) {
	t.Parallel()

	tr := NewTrueRange()
	act := tr.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	check("Identifier", core.TrueRange, act.Identifier)
	check("Mnemonic", "tr", act.Mnemonic)
	check("Description", "True Range", act.Description)
	check("len(Outputs)", 1, len(act.Outputs))
	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", "tr", act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", "True Range", act.Outputs[0].Description)
}
