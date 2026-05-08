//nolint:testpackage
package coronasignaltonoiseratio

import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
	"zpano/indicators/core/outputs/shape"
)

func TestCoronaSignalToNoiseRatioUpdate(t *testing.T) {
	t.Parallel()

	input := testCSNRInput()
	t0 := testCSNRTime()

	// Snapshot values captured from a first run and locked in here.
	type snap struct {
		i   int
		snr float64
		vmn float64
		vmx float64
	}
	snapshots := []snap{
		{11, 1.0000000000, 0.0000000000, 20.0000000000},
		{12, 1.0000000000, 0.0000000000, 20.0000000000},
		{50, 1.0000000000, 0.0000000000, 20.0000000000},
		{100, 2.9986583538, 4.2011609652, 20.0000000000},
		{150, 1.0000000000, 0.0000000035, 20.0000000000},
		{200, 1.0000000000, 0.0000000000, 20.0000000000},
		{251, 1.0000000000, 0.0000000026, 20.0000000000},
	}

	x, err := NewCoronaSignalToNoiseRatioDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	si := 0
	for i := range input {
		low, high := makeHL(i, input[i])
		h, snr := x.Update(input[i], low, high, t0.Add(time.Duration(i)*time.Minute))

		if h == nil {
			t.Fatalf("[%d] heatmap must not be nil", i)
		}

		if h.ParameterFirst != 1 || h.ParameterLast != 11 || math.Abs(h.ParameterResolution-4.9) > 1e-9 {
			t.Errorf("[%d] heatmap axis incorrect: first=%v last=%v result=%v",
				i, h.ParameterFirst, h.ParameterLast, h.ParameterResolution)
		}

		if !x.IsPrimed() {
			if !h.IsEmpty() {
				t.Errorf("[%d] expected empty heatmap before priming, got len=%d", i, len(h.Values))
			}

			if !math.IsNaN(snr) {
				t.Errorf("[%d] expected NaN snr before priming, got %v", i, snr)
			}

			continue
		}

		if len(h.Values) != 50 {
			t.Errorf("[%d] heatmap values length: expected 50, got %d", i, len(h.Values))
		}

		if si < len(snapshots) && snapshots[si].i == i {
			if math.Abs(snapshots[si].snr-snr) > testCSNRTolerance {
				t.Errorf("[%d] snr: expected %v, got %v", i, snapshots[si].snr, snr)
			}

			if math.Abs(snapshots[si].vmn-h.ValueMin) > testCSNRTolerance {
				t.Errorf("[%d] vmin: expected %v, got %v", i, snapshots[si].vmn, h.ValueMin)
			}

			if math.Abs(snapshots[si].vmx-h.ValueMax) > testCSNRTolerance {
				t.Errorf("[%d] vmax: expected %v, got %v", i, snapshots[si].vmx, h.ValueMax)
			}

			si++
		}
	}

	if si != len(snapshots) {
		t.Errorf("did not hit all %d snapshots, reached %d", len(snapshots), si)
	}
}

func TestCoronaSignalToNoiseRatioPrimesAtBar11(t *testing.T) {
	t.Parallel()

	x, _ := NewCoronaSignalToNoiseRatioDefault()

	if x.IsPrimed() {
		t.Error("expected not primed at start")
	}

	input := testCSNRInput()
	t0 := testCSNRTime()
	primedAt := -1

	for i := range input {
		low, high := makeHL(i, input[i])
		x.Update(input[i], low, high, t0.Add(time.Duration(i)*time.Minute))

		if x.IsPrimed() && primedAt < 0 {
			primedAt = i
		}
	}

	if primedAt != 11 {
		t.Errorf("expected priming at index 11, got %d", primedAt)
	}
}

func TestCoronaSignalToNoiseRatioNaNInput(t *testing.T) {
	t.Parallel()

	x, _ := NewCoronaSignalToNoiseRatioDefault()

	h, snr := x.Update(math.NaN(), math.NaN(), math.NaN(), testCSNRTime())

	if h == nil || !h.IsEmpty() {
		t.Errorf("expected empty heatmap for NaN input, got %v", h)
	}

	if !math.IsNaN(snr) {
		t.Errorf("expected NaN snr for NaN input, got %v", snr)
	}

	if x.IsPrimed() {
		t.Error("NaN input must not prime the indicator")
	}
}

func TestCoronaSignalToNoiseRatioMetadata(t *testing.T) {
	t.Parallel()

	x, _ := NewCoronaSignalToNoiseRatioDefault()
	md := x.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s: expected %v, actual %v", what, exp, act)
		}
	}

	mnValue := "csnr(50, 20, 1, 11, 30, hl/2)"
	mnSNR := "csnr-snr(30, hl/2)"

	check("Identifier", core.CoronaSignalToNoiseRatio, md.Identifier)
	check("Mnemonic", mnValue, md.Mnemonic)
	check("Description", "Corona signal to noise ratio "+mnValue, md.Description)
	check("len(Outputs)", 2, len(md.Outputs))

	check("Outputs[0].Kind", int(Value), md.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Heatmap, md.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", mnValue, md.Outputs[0].Mnemonic)

	check("Outputs[1].Kind", int(SignalToNoiseRatio), md.Outputs[1].Kind)
	check("Outputs[1].Shape", shape.Scalar, md.Outputs[1].Shape)
	check("Outputs[1].Mnemonic", mnSNR, md.Outputs[1].Mnemonic)
}

//nolint:funlen
func TestCoronaSignalToNoiseRatioUpdateEntity(t *testing.T) {
	t.Parallel()

	const (
		primeCount = 50
		inp        = 100.
		outputLen  = 2
	)

	tm := testCSNRTime()
	input := testCSNRInput()

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

	prime := func(x *CoronaSignalToNoiseRatio) {
		for i := 0; i < primeCount; i++ {
			low, high := makeHL(i, input[i%len(input)])
			x.Update(input[i%len(input)], low, high, tm)
		}
	}

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		s := entities.Scalar{Time: tm, Value: inp}
		x, _ := NewCoronaSignalToNoiseRatioDefault()
		prime(x)
		check(x.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: tm, High: inp * 1.005, Low: inp * 0.995, Close: inp}
		x, _ := NewCoronaSignalToNoiseRatioDefault()
		prime(x)
		check(x.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: tm, Bid: inp, Ask: inp}
		x, _ := NewCoronaSignalToNoiseRatioDefault()
		prime(x)
		check(x.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: tm, Price: inp}
		x, _ := NewCoronaSignalToNoiseRatioDefault()
		prime(x)
		check(x.UpdateTrade(&r))
	})
}

//nolint:funlen
func TestNewCoronaSignalToNoiseRatio(t *testing.T) {
	t.Parallel()

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("default", func(t *testing.T) {
		t.Parallel()

		x, err := NewCoronaSignalToNoiseRatioDefault()
		check("err == nil", true, err == nil)
		check("mnemonic", "csnr(50, 20, 1, 11, 30, hl/2)", x.mnemonic)
		check("MinParameterValue", 1.0, x.minParameterValue)
		check("MaxParameterValue", 11.0, x.maxParameterValue)
		check("RasterLength", 50, x.rasterLength)
	})

	t.Run("RasterLength < 2", func(t *testing.T) {
		t.Parallel()

		_, err := NewCoronaSignalToNoiseRatioParams(&Params{RasterLength: 1})
		if err == nil || err.Error() !=
			"invalid corona signal to noise ratio parameters: RasterLength should be >= 2" {
			t.Errorf("unexpected: %v", err)
		}
	})

	t.Run("MaxParameterValue <= MinParameterValue", func(t *testing.T) {
		t.Parallel()

		_, err := NewCoronaSignalToNoiseRatioParams(&Params{
			MinParameterValue: 5,
			MaxParameterValue: 5,
		})
		if err == nil || err.Error() !=
			"invalid corona signal to noise ratio parameters: MaxParameterValue should be > MinParameterValue" {
			t.Errorf("unexpected: %v", err)
		}
	})

	t.Run("HighPassFilterCutoff < 2", func(t *testing.T) {
		t.Parallel()

		_, err := NewCoronaSignalToNoiseRatioParams(&Params{HighPassFilterCutoff: 1})
		if err == nil || err.Error() !=
			"invalid corona signal to noise ratio parameters: HighPassFilterCutoff should be >= 2" {
			t.Errorf("unexpected: %v", err)
		}
	})

	t.Run("MinimalPeriod < 2", func(t *testing.T) {
		t.Parallel()

		_, err := NewCoronaSignalToNoiseRatioParams(&Params{MinimalPeriod: 1})
		if err == nil || err.Error() !=
			"invalid corona signal to noise ratio parameters: MinimalPeriod should be >= 2" {
			t.Errorf("unexpected: %v", err)
		}
	})

	t.Run("MaximalPeriod <= MinimalPeriod", func(t *testing.T) {
		t.Parallel()

		_, err := NewCoronaSignalToNoiseRatioParams(&Params{
			MinimalPeriod: 10,
			MaximalPeriod: 10,
		})
		if err == nil || err.Error() !=
			"invalid corona signal to noise ratio parameters: MaximalPeriod should be > MinimalPeriod" {
			t.Errorf("unexpected: %v", err)
		}
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()

		_, err := NewCoronaSignalToNoiseRatioParams(&Params{
			BarComponent: entities.BarComponent(9999),
		})
		if err == nil {
			t.Error("expected error")
		}
	})
}
