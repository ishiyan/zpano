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

func testCSwingTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

// testCSwingInput is the 252-entry TA-Lib MAMA reference series.
//
//nolint:dupl
func testCSwingInput() []float64 {
	return []float64{
		92.0000, 93.1725, 95.3125, 94.8450, 94.4075, 94.1100, 93.5000, 91.7350, 90.9550, 91.6875,
		94.5000, 97.9700, 97.5775, 90.7825, 89.0325, 92.0950, 91.1550, 89.7175, 90.6100, 91.0000,
		88.9225, 87.5150, 86.4375, 83.8900, 83.0025, 82.8125, 82.8450, 86.7350, 86.8600, 87.5475,
		85.7800, 86.1725, 86.4375, 87.2500, 88.9375, 88.2050, 85.8125, 84.5950, 83.6575, 84.4550,
		83.5000, 86.7825, 88.1725, 89.2650, 90.8600, 90.7825, 91.8600, 90.3600, 89.8600, 90.9225,
		89.5000, 87.6725, 86.5000, 84.2825, 82.9075, 84.2500, 85.6875, 86.6100, 88.2825, 89.5325,
		89.5000, 88.0950, 90.6250, 92.2350, 91.6725, 92.5925, 93.0150, 91.1725, 90.9850, 90.3775,
		88.2500, 86.9075, 84.0925, 83.1875, 84.2525, 97.8600, 99.8750, 103.2650, 105.9375, 103.5000,
		103.1100, 103.6100, 104.6400, 106.8150, 104.9525, 105.5000, 107.1400, 109.7350, 109.8450, 110.9850,
		120.0000, 119.8750, 117.9075, 119.4075, 117.9525, 117.2200, 115.6425, 113.1100, 111.7500, 114.5175,
		114.7450, 115.4700, 112.5300, 112.0300, 113.4350, 114.2200, 119.5950, 117.9650, 118.7150, 115.0300,
		114.5300, 115.0000, 116.5300, 120.1850, 120.5000, 120.5950, 124.1850, 125.3750, 122.9700, 123.0000,
		124.4350, 123.4400, 124.0300, 128.1850, 129.6550, 130.8750, 132.3450, 132.0650, 133.8150, 135.6600,
		137.0350, 137.4700, 137.3450, 136.3150, 136.4400, 136.2850, 129.0950, 128.3100, 126.0000, 124.0300,
		123.9350, 125.0300, 127.2500, 125.6200, 125.5300, 123.9050, 120.6550, 119.9650, 120.7800, 124.0000,
		122.7800, 120.7200, 121.7800, 122.4050, 123.2500, 126.1850, 127.5600, 126.5650, 123.0600, 122.7150,
		123.5900, 122.3100, 122.4650, 123.9650, 123.9700, 124.1550, 124.4350, 127.0000, 125.5000, 128.8750,
		130.5350, 132.3150, 134.0650, 136.0350, 133.7800, 132.7500, 133.4700, 130.9700, 127.5950, 128.4400,
		127.9400, 125.8100, 124.6250, 122.7200, 124.0900, 123.2200, 121.4050, 120.9350, 118.2800, 118.3750,
		121.1550, 120.9050, 117.1250, 113.0600, 114.9050, 112.4350, 107.9350, 105.9700, 106.3700, 106.8450,
		106.9700, 110.0300, 91.0000, 93.5600, 93.6200, 95.3100, 94.1850, 94.7800, 97.6250, 97.5900,
		95.2500, 94.7200, 92.2200, 91.5650, 92.2200, 93.8100, 95.5900, 96.1850, 94.6250, 95.1200,
		94.0000, 93.7450, 95.9050, 101.7450, 106.4400, 107.9350, 103.4050, 105.0600, 104.1550, 103.3100,
		103.3450, 104.8400, 110.4050, 114.5000, 117.3150, 118.2500, 117.1850, 109.7500, 109.6550, 108.5300,
		106.2200, 107.7200, 109.8400, 109.0950, 109.0900, 109.1550, 109.3150, 109.0600, 109.9050, 109.6250,
		109.5300, 108.0600,
	}
}

const testCSwingTolerance = 1e-4

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
