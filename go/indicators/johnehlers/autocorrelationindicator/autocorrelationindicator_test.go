//nolint:testpackage
package autocorrelationindicator

import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
	"zpano/indicators/core/outputs/shape"
)

func TestAutoCorrelationIndicatorUpdate(t *testing.T) {
	t.Parallel()

	input := testAciInput()
	t0 := testAciTime()

	x, err := NewAutoCorrelationIndicatorDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	si := 0

	for i := range input {
		h := x.Update(input[i], t0.Add(time.Duration(i)*time.Minute))
		if h == nil {
			t.Fatalf("[%d] heatmap must not be nil", i)
		}

		if h.ParameterFirst != 3 || h.ParameterLast != 48 || h.ParameterResolution != 1 {
			t.Errorf("[%d] axis incorrect: first=%v last=%v result=%v",
				i, h.ParameterFirst, h.ParameterLast, h.ParameterResolution)
		}

		if !x.IsPrimed() {
			if !h.IsEmpty() {
				t.Errorf("[%d] expected empty heatmap before priming, got len=%d", i, len(h.Values))
			}

			continue
		}

		if len(h.Values) != 46 {
			t.Errorf("[%d] expected values len=46, got %d", i, len(h.Values))
		}

		if si < len(aciSnapshots) && aciSnapshots[si].i == i {
			snap := aciSnapshots[si]
			if len(snap.spots) == 0 {
				// Placeholder: emit a capture line to paste back.
				t.Logf("CAPTURE [%d] valueMin=%.15f valueMax=%.15f", i, h.ValueMin, h.ValueMax)
				for _, bin := range []int{0, 9, 19, 28, 44} {
					t.Logf("CAPTURE [%d] Values[%d]=%.15f", i, bin, h.Values[bin])
				}

				si++

				continue
			}

			if math.Abs(h.ValueMin-snap.valueMin) > testAciMinMaxTol {
				t.Errorf("[%d] ValueMin: expected %v, got %v", i, snap.valueMin, h.ValueMin)
			}

			if math.Abs(h.ValueMax-snap.valueMax) > testAciMinMaxTol {
				t.Errorf("[%d] ValueMax: expected %v, got %v", i, snap.valueMax, h.ValueMax)
			}

			for _, sp := range snap.spots {
				if math.Abs(h.Values[sp.i]-sp.v) > testAciTolerance {
					t.Errorf("[%d] Values[%d]: expected %v, got %v", i, sp.i, sp.v, h.Values[sp.i])
				}
			}

			si++
		}
	}

	if si != len(aciSnapshots) {
		t.Errorf("did not hit all %d snapshots, reached %d", len(aciSnapshots), si)
	}
}

// TestAutoCorrelationIndicatorSyntheticSine injects a pure sinusoid at a known
// period and verifies the autocorrelation peak lands at that lag. A pure sine
// of period P autocorrelates to 1.0 at every lag that is a multiple of P, so
// we choose a period for which only one multiple fits in the lag range.
func TestAutoCorrelationIndicatorSyntheticSine(t *testing.T) {
	t.Parallel()

	const (
		// period=35: only multiple in [MinLag=3, MaxLag=48] is 35 itself.
		period = 35.0
		bars   = 600
	)

	x, err := NewAutoCorrelationIndicatorDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t0 := testAciTime()

	var last *outputs.Heatmap

	for i := 0; i < bars; i++ {
		sample := 100 + math.Sin(2*math.Pi*float64(i)/period)
		last = x.Update(sample, t0.Add(time.Duration(i)*time.Minute))
	}

	if last == nil || last.IsEmpty() {
		t.Fatal("expected primed non-empty heatmap")
	}

	peakBin := 0
	for i := range last.Values {
		if last.Values[i] > last.Values[peakBin] {
			peakBin = i
		}
	}

	// Bin k corresponds to lag MinLag+k. MinLag=3, period=35 -> bin 32.
	expectedBin := int(period - last.ParameterFirst)
	if peakBin != expectedBin {
		t.Errorf("peak bin: expected %d (lag %.0f), got %d (lag %.0f)",
			expectedBin, period, peakBin, last.ParameterFirst+float64(peakBin))
	}
}

func TestAutoCorrelationIndicatorNaNInput(t *testing.T) {
	t.Parallel()

	x, _ := NewAutoCorrelationIndicatorDefault()

	h := x.Update(math.NaN(), testAciTime())

	if h == nil || !h.IsEmpty() {
		t.Errorf("expected empty heatmap for NaN input, got %v", h)
	}

	if x.IsPrimed() {
		t.Error("NaN input must not prime the indicator")
	}
}

func TestAutoCorrelationIndicatorMetadata(t *testing.T) {
	t.Parallel()

	x, _ := NewAutoCorrelationIndicatorDefault()
	md := x.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s: expected %v, actual %v", what, exp, act)
		}
	}

	mn := "aci(3, 48, 10, hl/2)"

	check("Identifier", core.AutoCorrelationIndicator, md.Identifier)
	check("Mnemonic", mn, md.Mnemonic)
	check("Description", "Autocorrelation indicator "+mn, md.Description)
	check("len(Outputs)", 1, len(md.Outputs))
	check("Outputs[0].Kind", int(Value), md.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Heatmap, md.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", mn, md.Outputs[0].Mnemonic)
}

func TestAutoCorrelationIndicatorMnemonicFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    Params
		mn   string
	}{
		{"default", Params{}, "aci(3, 48, 10, hl/2)"},
		{"average override", Params{AveragingLength: 5}, "aci(3, 48, 10, average=5, hl/2)"},
		{
			"custom range",
			Params{MinLag: 5, MaxLag: 30, SmoothingPeriod: 8},
			"aci(5, 30, 8, hl/2)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			x, err := NewAutoCorrelationIndicatorParams(&tt.p)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if x.mnemonic != tt.mn {
				t.Errorf("expected %q, got %q", tt.mn, x.mnemonic)
			}
		})
	}
}

func TestAutoCorrelationIndicatorValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    Params
		msg  string
	}{
		{
			"MinLag < 1",
			Params{MinLag: -1, MaxLag: 48, SmoothingPeriod: 10},
			"invalid autocorrelation indicator parameters: MinLag should be >= 1",
		},
		{
			"MaxLag <= MinLag",
			Params{MinLag: 10, MaxLag: 10, SmoothingPeriod: 10},
			"invalid autocorrelation indicator parameters: MaxLag should be > MinLag",
		},
		{
			"SmoothingPeriod < 2",
			Params{MinLag: 3, MaxLag: 48, SmoothingPeriod: 1},
			"invalid autocorrelation indicator parameters: SmoothingPeriod should be >= 2",
		},
		{
			"AveragingLength < 0",
			Params{AveragingLength: -1},
			"invalid autocorrelation indicator parameters: AveragingLength should be >= 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewAutoCorrelationIndicatorParams(&tt.p)
			if err == nil || err.Error() != tt.msg {
				t.Errorf("expected %q, got %v", tt.msg, err)
			}
		})
	}
}

func TestAutoCorrelationIndicatorInvalidBarComponent(t *testing.T) {
	t.Parallel()

	_, err := NewAutoCorrelationIndicatorParams(&Params{BarComponent: entities.BarComponent(9999)})
	if err == nil {
		t.Error("expected error")
	}
}

//nolint:funlen
func TestAutoCorrelationIndicatorUpdateEntity(t *testing.T) {
	t.Parallel()

	const (
		primeCount = 200
		inp        = 100.
		outputLen  = 1
	)

	tm := testAciTime()
	input := testAciInput()

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
	}

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		s := entities.Scalar{Time: tm, Value: inp}
		x, _ := NewAutoCorrelationIndicatorDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}

		check(x.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: tm, High: inp, Low: inp, Close: inp}
		x, _ := NewAutoCorrelationIndicatorDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}

		check(x.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: tm, Bid: inp, Ask: inp}
		x, _ := NewAutoCorrelationIndicatorDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}

		check(x.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: tm, Price: inp}
		x, _ := NewAutoCorrelationIndicatorDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}

		check(x.UpdateTrade(&r))
	})
}
