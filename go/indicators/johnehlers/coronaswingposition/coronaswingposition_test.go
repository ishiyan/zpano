//nolint:testpackage
package coronaswingposition

import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
	"zpano/indicators/core/outputs/shape"
)

func TestCoronaSwingPositionUpdate(t *testing.T) {
	t.Parallel()

	input := testCSwingInput()
	t0 := testCSwingTime()

	type snap struct {
		i   int
		sp  float64
		vmn float64
		vmx float64
	}
	snapshots := []snap{
		{11, 5.0000000000, 20.0000000000, 20.0000000000},
		{12, 5.0000000000, 20.0000000000, 20.0000000000},
		{50, 4.5384908349, 20.0000000000, 20.0000000000},
		{100, -3.8183742675, 3.4957777081, 20.0000000000},
		{150, -1.8516194371, 5.3792287864, 20.0000000000},
		{200, -3.6944428668, 4.2580825738, 20.0000000000},
		{251, -0.8524812061, 4.4822539784, 20.0000000000},
	}

	x, err := NewCoronaSwingPositionDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	si := 0
	for i := range input {
		h, sp := x.Update(input[i], t0.Add(time.Duration(i)*time.Minute))

		if h == nil {
			t.Fatalf("[%d] heatmap must not be nil", i)
		}

		if h.ParameterFirst != -5 || h.ParameterLast != 5 || math.Abs(h.ParameterResolution-4.9) > 1e-9 {
			t.Errorf("[%d] heatmap axis incorrect: first=%v last=%v result=%v",
				i, h.ParameterFirst, h.ParameterLast, h.ParameterResolution)
		}

		if !x.IsPrimed() {
			if !h.IsEmpty() {
				t.Errorf("[%d] expected empty heatmap before priming, got len=%d", i, len(h.Values))
			}

			if !math.IsNaN(sp) {
				t.Errorf("[%d] expected NaN sp before priming, got %v", i, sp)
			}

			continue
		}

		if len(h.Values) != 50 {
			t.Errorf("[%d] heatmap values length: expected 50, got %d", i, len(h.Values))
		}

		if si < len(snapshots) && snapshots[si].i == i {
			if math.Abs(snapshots[si].sp-sp) > testCSwingTolerance {
				t.Errorf("[%d] sp: expected %v, got %v", i, snapshots[si].sp, sp)
			}

			if math.Abs(snapshots[si].vmn-h.ValueMin) > testCSwingTolerance {
				t.Errorf("[%d] vmin: expected %v, got %v", i, snapshots[si].vmn, h.ValueMin)
			}

			if math.Abs(snapshots[si].vmx-h.ValueMax) > testCSwingTolerance {
				t.Errorf("[%d] vmax: expected %v, got %v", i, snapshots[si].vmx, h.ValueMax)
			}

			si++
		}
	}

	if si != len(snapshots) {
		t.Errorf("did not hit all %d snapshots, reached %d", len(snapshots), si)
	}
}

func TestCoronaSwingPositionPrimesAtBar11(t *testing.T) {
	t.Parallel()

	x, _ := NewCoronaSwingPositionDefault()

	if x.IsPrimed() {
		t.Error("expected not primed at start")
	}

	input := testCSwingInput()
	t0 := testCSwingTime()
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

func TestCoronaSwingPositionNaNInput(t *testing.T) {
	t.Parallel()

	x, _ := NewCoronaSwingPositionDefault()

	h, sp := x.Update(math.NaN(), testCSwingTime())

	if h == nil || !h.IsEmpty() {
		t.Errorf("expected empty heatmap for NaN input, got %v", h)
	}

	if !math.IsNaN(sp) {
		t.Errorf("expected NaN sp for NaN input, got %v", sp)
	}

	if x.IsPrimed() {
		t.Error("NaN input must not prime the indicator")
	}
}

func TestCoronaSwingPositionMetadata(t *testing.T) {
	t.Parallel()

	x, _ := NewCoronaSwingPositionDefault()
	md := x.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s: expected %v, actual %v", what, exp, act)
		}
	}

	mnValue := "cswing(50, 20, -5, 5, 30, hl/2)"
	mnSP := "cswing-sp(30, hl/2)"

	check("Identifier", core.CoronaSwingPosition, md.Identifier)
	check("Mnemonic", mnValue, md.Mnemonic)
	check("Description", "Corona swing position "+mnValue, md.Description)
	check("len(Outputs)", 2, len(md.Outputs))

	check("Outputs[0].Kind", int(Value), md.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Heatmap, md.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", mnValue, md.Outputs[0].Mnemonic)

	check("Outputs[1].Kind", int(SwingPosition), md.Outputs[1].Kind)
	check("Outputs[1].Shape", shape.Scalar, md.Outputs[1].Shape)
	check("Outputs[1].Mnemonic", mnSP, md.Outputs[1].Mnemonic)
}

//nolint:funlen
func TestCoronaSwingPositionUpdateEntity(t *testing.T) {
	t.Parallel()

	const (
		primeCount = 50
		inp        = 100.
		outputLen  = 2
	)

	tm := testCSwingTime()
	input := testCSwingInput()

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

	prime := func(x *CoronaSwingPosition) {
		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}
	}

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		s := entities.Scalar{Time: tm, Value: inp}
		x, _ := NewCoronaSwingPositionDefault()
		prime(x)
		check(x.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: tm, High: inp * 1.005, Low: inp * 0.995, Close: inp}
		x, _ := NewCoronaSwingPositionDefault()
		prime(x)
		check(x.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: tm, Bid: inp, Ask: inp}
		x, _ := NewCoronaSwingPositionDefault()
		prime(x)
		check(x.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: tm, Price: inp}
		x, _ := NewCoronaSwingPositionDefault()
		prime(x)
		check(x.UpdateTrade(&r))
	})
}

//nolint:funlen
func TestNewCoronaSwingPosition(t *testing.T) {
	t.Parallel()

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("default", func(t *testing.T) {
		t.Parallel()

		x, err := NewCoronaSwingPositionDefault()
		check("err == nil", true, err == nil)
		check("mnemonic", "cswing(50, 20, -5, 5, 30, hl/2)", x.mnemonic)
		check("MinParameterValue", -5.0, x.minParameterValue)
		check("MaxParameterValue", 5.0, x.maxParameterValue)
		check("RasterLength", 50, x.rasterLength)
	})

	t.Run("RasterLength < 2", func(t *testing.T) {
		t.Parallel()

		_, err := NewCoronaSwingPositionParams(&Params{RasterLength: 1})
		if err == nil || err.Error() !=
			"invalid corona swing position parameters: RasterLength should be >= 2" {
			t.Errorf("unexpected: %v", err)
		}
	})

	t.Run("MaxParameterValue <= MinParameterValue", func(t *testing.T) {
		t.Parallel()

		_, err := NewCoronaSwingPositionParams(&Params{
			MinParameterValue: 5,
			MaxParameterValue: 5,
		})
		if err == nil || err.Error() !=
			"invalid corona swing position parameters: MaxParameterValue should be > MinParameterValue" {
			t.Errorf("unexpected: %v", err)
		}
	})

	t.Run("HighPassFilterCutoff < 2", func(t *testing.T) {
		t.Parallel()

		_, err := NewCoronaSwingPositionParams(&Params{HighPassFilterCutoff: 1})
		if err == nil || err.Error() !=
			"invalid corona swing position parameters: HighPassFilterCutoff should be >= 2" {
			t.Errorf("unexpected: %v", err)
		}
	})

	t.Run("MinimalPeriod < 2", func(t *testing.T) {
		t.Parallel()

		_, err := NewCoronaSwingPositionParams(&Params{MinimalPeriod: 1})
		if err == nil || err.Error() !=
			"invalid corona swing position parameters: MinimalPeriod should be >= 2" {
			t.Errorf("unexpected: %v", err)
		}
	})

	t.Run("MaximalPeriod <= MinimalPeriod", func(t *testing.T) {
		t.Parallel()

		_, err := NewCoronaSwingPositionParams(&Params{
			MinimalPeriod: 10,
			MaximalPeriod: 10,
		})
		if err == nil || err.Error() !=
			"invalid corona swing position parameters: MaximalPeriod should be > MinimalPeriod" {
			t.Errorf("unexpected: %v", err)
		}
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()

		_, err := NewCoronaSwingPositionParams(&Params{
			BarComponent: entities.BarComponent(9999),
		})
		if err == nil {
			t.Error("expected error")
		}
	})
}
