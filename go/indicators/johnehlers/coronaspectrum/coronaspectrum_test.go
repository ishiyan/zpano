//nolint:testpackage
package coronaspectrum

import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
	"zpano/indicators/core/outputs/shape"
)

func TestCoronaSpectrumUpdate(t *testing.T) {
	t.Parallel()

	input := testCSInput()
	t0 := testCSTime()

	// Snapshot values captured from a first run and locked in here.
	type snap struct {
		i   int
		dc  float64
		dcm float64
	}
	snapshots := []snap{
		{11, 17.7604672565, 17.7604672565},
		{12, 6.0000000000, 6.0000000000},
		{50, 15.9989078712, 15.9989078712},
		{100, 14.7455497547, 14.7455497547},
		{150, 17.5000000000, 17.2826036069},
		{200, 19.7557338512, 20.0000000000},
		{251, 6.0000000000, 6.0000000000},
	}

	x, err := NewCoronaSpectrumDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	si := 0
	for i := range input {
		h, dc, dcm := x.Update(input[i], t0.Add(time.Duration(i)*time.Minute))

		if h == nil {
			t.Fatalf("[%d] heatmap must not be nil", i)
		}

		// Heatmap axis invariants are always present.
		if h.ParameterFirst != 6 || h.ParameterLast != 30 || h.ParameterResolution != 2 {
			t.Errorf("[%d] heatmap axis incorrect: first=%v last=%v result=%v",
				i, h.ParameterFirst, h.ParameterLast, h.ParameterResolution)
		}

		if !x.IsPrimed() {
			if !h.IsEmpty() {
				t.Errorf("[%d] expected empty heatmap before priming, got len=%d", i, len(h.Values))
			}

			if !math.IsNaN(dc) || !math.IsNaN(dcm) {
				t.Errorf("[%d] expected NaN scalars before priming, got dc=%v dcm=%v", i, dc, dcm)
			}

			continue
		}

		if len(h.Values) != 49 {
			t.Errorf("[%d] heatmap values length: expected 49, got %d", i, len(h.Values))
		}

		if si < len(snapshots) && snapshots[si].i == i {
			if math.Abs(snapshots[si].dc-dc) > testCSTolerance {
				t.Errorf("[%d] dc: expected %v, got %v", i, snapshots[si].dc, dc)
			}

			if math.Abs(snapshots[si].dcm-dcm) > testCSTolerance {
				t.Errorf("[%d] dcm: expected %v, got %v", i, snapshots[si].dcm, dcm)
			}

			si++
		}
	}

	if si != len(snapshots) {
		t.Errorf("did not hit all %d snapshots, reached %d", len(snapshots), si)
	}
}

func TestCoronaSpectrumPrimesAtBar11(t *testing.T) {
	t.Parallel()

	x, _ := NewCoronaSpectrumDefault()

	if x.IsPrimed() {
		t.Error("expected not primed at start")
	}

	input := testCSInput()
	t0 := testCSTime()
	primedAt := -1

	for i := range input {
		x.Update(input[i], t0.Add(time.Duration(i)*time.Minute))

		if x.IsPrimed() && primedAt < 0 {
			primedAt = i
		}
	}

	// MinimalPeriodTimesTwo = 12, so priming first happens at sample index 11 (0-indexed).
	if primedAt != 11 {
		t.Errorf("expected priming at index 11, got %d", primedAt)
	}
}

func TestCoronaSpectrumNaNInput(t *testing.T) {
	t.Parallel()

	x, _ := NewCoronaSpectrumDefault()

	h, dc, dcm := x.Update(math.NaN(), testCSTime())

	if h == nil || !h.IsEmpty() {
		t.Errorf("expected empty heatmap for NaN input, got %v", h)
	}

	if !math.IsNaN(dc) || !math.IsNaN(dcm) {
		t.Errorf("expected NaN scalars for NaN input, got dc=%v dcm=%v", dc, dcm)
	}

	// NaN must not prime the indicator.
	if x.IsPrimed() {
		t.Error("NaN input must not prime the indicator")
	}
}

func TestCoronaSpectrumMetadata(t *testing.T) {
	t.Parallel()

	x, _ := NewCoronaSpectrumDefault()
	md := x.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s: expected %v, actual %v", what, exp, act)
		}
	}

	mnValue := "cspect(6, 20, 6, 30, 30, hl/2)"
	mnDC := "cspect-dc(30, hl/2)"
	mnDCM := "cspect-dcm(30, hl/2)"

	check("Identifier", core.CoronaSpectrum, md.Identifier)
	check("Mnemonic", mnValue, md.Mnemonic)
	check("Description", "Corona spectrum "+mnValue, md.Description)
	check("len(Outputs)", 3, len(md.Outputs))

	check("Outputs[0].Kind", int(Value), md.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Heatmap, md.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", mnValue, md.Outputs[0].Mnemonic)

	check("Outputs[1].Kind", int(DominantCycle), md.Outputs[1].Kind)
	check("Outputs[1].Shape", shape.Scalar, md.Outputs[1].Shape)
	check("Outputs[1].Mnemonic", mnDC, md.Outputs[1].Mnemonic)

	check("Outputs[2].Kind", int(DominantCycleMedian), md.Outputs[2].Kind)
	check("Outputs[2].Shape", shape.Scalar, md.Outputs[2].Shape)
	check("Outputs[2].Mnemonic", mnDCM, md.Outputs[2].Mnemonic)
}

//nolint:funlen
func TestCoronaSpectrumUpdateEntity(t *testing.T) {
	t.Parallel()

	const (
		primeCount = 50
		inp        = 100.
		outputLen  = 3
	)

	tm := testCSTime()
	input := testCSInput()

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

		for i := 1; i < outputLen; i++ {
			s, ok := act[i].(entities.Scalar)
			if !ok {
				t.Errorf("output[%d] is not a scalar", i)

				continue
			}

			if s.Time != tm {
				t.Errorf("output[%d].Time: expected %v, actual %v", i, tm, s.Time)
			}
		}
	}

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		s := entities.Scalar{Time: tm, Value: inp}
		x, _ := NewCoronaSpectrumDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}

		check(x.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: tm, High: inp, Low: inp, Close: inp}
		x, _ := NewCoronaSpectrumDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}

		check(x.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: tm, Bid: inp, Ask: inp}
		x, _ := NewCoronaSpectrumDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}

		check(x.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: tm, Price: inp}
		x, _ := NewCoronaSpectrumDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}

		check(x.UpdateTrade(&r))
	})
}

//nolint:funlen
func TestNewCoronaSpectrum(t *testing.T) {
	t.Parallel()

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("default", func(t *testing.T) {
		t.Parallel()

		x, err := NewCoronaSpectrumDefault()
		check("err == nil", true, err == nil)
		check("mnemonic", "cspect(6, 20, 6, 30, 30, hl/2)", x.mnemonic)
		check("MinParameterValue", 6.0, x.minParameterValue)
		check("MaxParameterValue", 30.0, x.maxParameterValue)
		check("ParameterResolution", 2.0, x.parameterResolution)
	})

	t.Run("custom ranges round to integers", func(t *testing.T) {
		t.Parallel()

		x, err := NewCoronaSpectrumParams(&Params{
			MinRasterValue:       4,
			MaxRasterValue:       25,
			MinParameterValue:    8.7,  // ceils to 9
			MaxParameterValue:    40.4, // floors to 40
			HighPassFilterCutoff: 20,
		})
		check("err == nil", true, err == nil)
		check("MinParameterValue", 9.0, x.minParameterValue)
		check("MaxParameterValue", 40.0, x.maxParameterValue)
		check("mnemonic", "cspect(4, 25, 9, 40, 20, hl/2)", x.mnemonic)
	})

	t.Run("MaxRasterValue <= MinRasterValue", func(t *testing.T) {
		t.Parallel()

		_, err := NewCoronaSpectrumParams(&Params{
			MinRasterValue:    10,
			MaxRasterValue:    10,
			MinParameterValue: 6,
			MaxParameterValue: 30,
		})
		if err == nil || err.Error() !=
			"invalid corona spectrum parameters: MaxRasterValue should be > MinRasterValue" {
			t.Errorf("unexpected: %v", err)
		}
	})

	t.Run("MinParameterValue < 2", func(t *testing.T) {
		t.Parallel()

		_, err := NewCoronaSpectrumParams(&Params{
			MinParameterValue: 1,
			MaxParameterValue: 30,
		})
		if err == nil || err.Error() !=
			"invalid corona spectrum parameters: MinParameterValue should be >= 2" {
			t.Errorf("unexpected: %v", err)
		}
	})

	t.Run("MaxParameterValue <= MinParameterValue", func(t *testing.T) {
		t.Parallel()

		_, err := NewCoronaSpectrumParams(&Params{
			MinParameterValue: 20,
			MaxParameterValue: 20,
		})
		if err == nil || err.Error() !=
			"invalid corona spectrum parameters: MaxParameterValue should be > MinParameterValue" {
			t.Errorf("unexpected: %v", err)
		}
	})

	t.Run("HighPassFilterCutoff < 2", func(t *testing.T) {
		t.Parallel()

		_, err := NewCoronaSpectrumParams(&Params{
			HighPassFilterCutoff: 1,
		})
		if err == nil || err.Error() !=
			"invalid corona spectrum parameters: HighPassFilterCutoff should be >= 2" {
			t.Errorf("unexpected: %v", err)
		}
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()

		_, err := NewCoronaSpectrumParams(&Params{
			BarComponent: entities.BarComponent(9999),
		})
		if err == nil {
			t.Error("expected error")
		}
	})
}
