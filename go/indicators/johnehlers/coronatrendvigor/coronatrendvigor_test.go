//nolint:testpackage
package coronatrendvigor

import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
	"zpano/indicators/core/outputs/shape"
)

func TestCoronaTrendVigorUpdate(t *testing.T) {
	t.Parallel()

	input := testCTVInput()
	t0 := testCTVTime()

	type snap struct {
		i   int
		tv  float64
		vmn float64
		vmx float64
	}
	snapshots := []snap{
		{11, 5.6512200755, 20.0000000000, 20.0000000000},
		{12, 6.8379492897, 20.0000000000, 20.0000000000},
		{50, 2.6145116709, 2.3773561485, 20.0000000000},
		{100, 2.7536803664, 2.4892742850, 20.0000000000},
		{150, -6.4606404251, 20.0000000000, 20.0000000000},
		{200, -10.0000000000, 20.0000000000, 20.0000000000},
		{251, -0.1894989954, 0.5847573715, 20.0000000000},
	}

	x, err := NewCoronaTrendVigorDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	si := 0
	for i := range input {
		h, tv := x.Update(input[i], t0.Add(time.Duration(i)*time.Minute))

		if h == nil {
			t.Fatalf("[%d] heatmap must not be nil", i)
		}

		if h.ParameterFirst != -10 || h.ParameterLast != 10 || math.Abs(h.ParameterResolution-2.45) > 1e-9 {
			t.Errorf("[%d] heatmap axis incorrect: first=%v last=%v result=%v",
				i, h.ParameterFirst, h.ParameterLast, h.ParameterResolution)
		}

		if !x.IsPrimed() {
			if !h.IsEmpty() {
				t.Errorf("[%d] expected empty heatmap before priming, got len=%d", i, len(h.Values))
			}

			if !math.IsNaN(tv) {
				t.Errorf("[%d] expected NaN tv before priming, got %v", i, tv)
			}

			continue
		}

		if len(h.Values) != 50 {
			t.Errorf("[%d] heatmap values length: expected 50, got %d", i, len(h.Values))
		}

		if si < len(snapshots) && snapshots[si].i == i {
			if math.Abs(snapshots[si].tv-tv) > testCTVTolerance {
				t.Errorf("[%d] tv: expected %v, got %v", i, snapshots[si].tv, tv)
			}

			if math.Abs(snapshots[si].vmn-h.ValueMin) > testCTVTolerance {
				t.Errorf("[%d] vmin: expected %v, got %v", i, snapshots[si].vmn, h.ValueMin)
			}

			if math.Abs(snapshots[si].vmx-h.ValueMax) > testCTVTolerance {
				t.Errorf("[%d] vmax: expected %v, got %v", i, snapshots[si].vmx, h.ValueMax)
			}

			si++
		}
	}

	if si != len(snapshots) {
		t.Errorf("did not hit all %d snapshots, reached %d", len(snapshots), si)
	}
}

func TestCoronaTrendVigorPrimesAtBar11(t *testing.T) {
	t.Parallel()

	x, _ := NewCoronaTrendVigorDefault()

	if x.IsPrimed() {
		t.Error("expected not primed at start")
	}

	input := testCTVInput()
	t0 := testCTVTime()
	primedAt := -1

	for i := range input {
		x.Update(input[i], t0.Add(time.Duration(i)*time.Minute))

		if x.IsPrimed() && primedAt < 0 {
			primedAt = i
		}
	}

	if primedAt != 11 {
		t.Errorf("expected priming at index 11, got %d", primedAt)
	}
}

func TestCoronaTrendVigorNaNInput(t *testing.T) {
	t.Parallel()

	x, _ := NewCoronaTrendVigorDefault()

	h, tv := x.Update(math.NaN(), testCTVTime())

	if h == nil || !h.IsEmpty() {
		t.Errorf("expected empty heatmap for NaN input, got %v", h)
	}

	if !math.IsNaN(tv) {
		t.Errorf("expected NaN tv for NaN input, got %v", tv)
	}

	if x.IsPrimed() {
		t.Error("NaN input must not prime the indicator")
	}
}

func TestCoronaTrendVigorMetadata(t *testing.T) {
	t.Parallel()

	x, _ := NewCoronaTrendVigorDefault()
	md := x.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s: expected %v, actual %v", what, exp, act)
		}
	}

	mnValue := "ctv(50, 20, -10, 10, 30, hl/2)"
	mnTV := "ctv-tv(30, hl/2)"

	check("Identifier", core.CoronaTrendVigor, md.Identifier)
	check("Mnemonic", mnValue, md.Mnemonic)
	check("Description", "Corona trend vigor "+mnValue, md.Description)
	check("len(Outputs)", 2, len(md.Outputs))

	check("Outputs[0].Kind", int(Value), md.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Heatmap, md.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", mnValue, md.Outputs[0].Mnemonic)

	check("Outputs[1].Kind", int(TrendVigor), md.Outputs[1].Kind)
	check("Outputs[1].Shape", shape.Scalar, md.Outputs[1].Shape)
	check("Outputs[1].Mnemonic", mnTV, md.Outputs[1].Mnemonic)
}

//nolint:funlen
func TestCoronaTrendVigorUpdateEntity(t *testing.T) {
	t.Parallel()

	const (
		primeCount = 50
		inp        = 100.
		outputLen  = 2
	)

	tm := testCTVTime()
	input := testCTVInput()

	check := func(act core.Output) {
		t.Helper()

		if len(act) != outputLen {
			t.Errorf("len(output): expected %v, actual %v", outputLen, len(act))

			return
		}

		h, ok := act[0].(*outputs.Heatmap)
		if !ok {
			t.Errorf("output[0] is not a heatmap: %T", act[0])
		} else if h.Time != tm {
			t.Errorf("output[0].Time: expected %v, actual %v", tm, h.Time)
		}

		s, ok := act[1].(entities.Scalar)
		if !ok {
			t.Errorf("output[1] is not a scalar")
		} else if s.Time != tm {
			t.Errorf("output[1].Time: expected %v, actual %v", tm, s.Time)
		}
	}

	prime := func(x *CoronaTrendVigor) {
		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}
	}

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		s := entities.Scalar{Time: tm, Value: inp}
		x, _ := NewCoronaTrendVigorDefault()
		prime(x)
		check(x.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: tm, High: inp * 1.005, Low: inp * 0.995, Close: inp}
		x, _ := NewCoronaTrendVigorDefault()
		prime(x)
		check(x.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: tm, Bid: inp, Ask: inp}
		x, _ := NewCoronaTrendVigorDefault()
		prime(x)
		check(x.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: tm, Price: inp}
		x, _ := NewCoronaTrendVigorDefault()
		prime(x)
		check(x.UpdateTrade(&r))
	})
}

//nolint:funlen
func TestNewCoronaTrendVigor(t *testing.T) {
	t.Parallel()

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("default", func(t *testing.T) {
		t.Parallel()

		x, err := NewCoronaTrendVigorDefault()
		check("err == nil", true, err == nil)
		check("mnemonic", "ctv(50, 20, -10, 10, 30, hl/2)", x.mnemonic)
		check("MinParameterValue", -10.0, x.minParameterValue)
		check("MaxParameterValue", 10.0, x.maxParameterValue)
		check("RasterLength", 50, x.rasterLength)
	})

	t.Run("RasterLength < 2", func(t *testing.T) {
		t.Parallel()

		_, err := NewCoronaTrendVigorParams(&Params{RasterLength: 1})
		if err == nil || err.Error() !=
			"invalid corona trend vigor parameters: RasterLength should be >= 2" {
			t.Errorf("unexpected: %v", err)
		}
	})

	t.Run("MaxParameterValue <= MinParameterValue", func(t *testing.T) {
		t.Parallel()

		_, err := NewCoronaTrendVigorParams(&Params{
			MinParameterValue: 5,
			MaxParameterValue: 5,
		})
		if err == nil || err.Error() !=
			"invalid corona trend vigor parameters: MaxParameterValue should be > MinParameterValue" {
			t.Errorf("unexpected: %v", err)
		}
	})

	t.Run("HighPassFilterCutoff < 2", func(t *testing.T) {
		t.Parallel()

		_, err := NewCoronaTrendVigorParams(&Params{HighPassFilterCutoff: 1})
		if err == nil || err.Error() !=
			"invalid corona trend vigor parameters: HighPassFilterCutoff should be >= 2" {
			t.Errorf("unexpected: %v", err)
		}
	})

	t.Run("MinimalPeriod < 2", func(t *testing.T) {
		t.Parallel()

		_, err := NewCoronaTrendVigorParams(&Params{MinimalPeriod: 1})
		if err == nil || err.Error() !=
			"invalid corona trend vigor parameters: MinimalPeriod should be >= 2" {
			t.Errorf("unexpected: %v", err)
		}
	})

	t.Run("MaximalPeriod <= MinimalPeriod", func(t *testing.T) {
		t.Parallel()

		_, err := NewCoronaTrendVigorParams(&Params{
			MinimalPeriod: 10,
			MaximalPeriod: 10,
		})
		if err == nil || err.Error() !=
			"invalid corona trend vigor parameters: MaximalPeriod should be > MinimalPeriod" {
			t.Errorf("unexpected: %v", err)
		}
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()

		_, err := NewCoronaTrendVigorParams(&Params{
			BarComponent: entities.BarComponent(9999),
		})
		if err == nil {
			t.Error("expected error")
		}
	})
}
