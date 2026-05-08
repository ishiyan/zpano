//nolint:testpackage
package autocorrelationperiodogram

import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
	"zpano/indicators/core/outputs/shape"
)

//nolint:funlen
func TestAutoCorrelationPeriodogramUpdate(t *testing.T) {
	t.Parallel()

	input := testAcpInput()
	t0 := testAcpTime()

	x, err := NewAutoCorrelationPeriodogramDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	si := 0

	for i := range input {
		h := x.Update(input[i], t0.Add(time.Duration(i)*time.Minute))
		if h == nil {
			t.Fatalf("[%d] heatmap must not be nil", i)
		}

		if h.ParameterFirst != 10 || h.ParameterLast != 48 || h.ParameterResolution != 1 {
			t.Errorf("[%d] axis incorrect: first=%v last=%v result=%v",
				i, h.ParameterFirst, h.ParameterLast, h.ParameterResolution)
		}

		if !x.IsPrimed() {
			if !h.IsEmpty() {
				t.Errorf("[%d] expected empty heatmap before priming, got len=%d", i, len(h.Values))
			}

			continue
		}

		if len(h.Values) != 39 {
			t.Errorf("[%d] expected values len=39, got %d", i, len(h.Values))
		}

		if si < len(acpSnapshots) && acpSnapshots[si].i == i {
			snap := acpSnapshots[si]
			if len(snap.spots) == 0 {
				t.Logf("CAPTURE [%d] valueMin=%.15f valueMax=%.15f", i, h.ValueMin, h.ValueMax)
				for _, bin := range []int{0, 9, 19, 28, 38} {
					t.Logf("CAPTURE [%d] Values[%d]=%.15f", i, bin, h.Values[bin])
				}

				si++

				continue
			}

			if math.Abs(h.ValueMin-snap.valueMin) > testAcpMinMaxTol {
				t.Errorf("[%d] ValueMin: expected %v, got %v", i, snap.valueMin, h.ValueMin)
			}

			if math.Abs(h.ValueMax-snap.valueMax) > testAcpMinMaxTol {
				t.Errorf("[%d] ValueMax: expected %v, got %v", i, snap.valueMax, h.ValueMax)
			}

			for _, sp := range snap.spots {
				if math.Abs(h.Values[sp.i]-sp.v) > testAcpTolerance {
					t.Errorf("[%d] Values[%d]: expected %v, got %v", i, sp.i, sp.v, h.Values[sp.i])
				}
			}

			si++
		}
	}

	if si != len(acpSnapshots) {
		t.Errorf("did not hit all %d snapshots, reached %d", len(acpSnapshots), si)
	}
}

// TestAutoCorrelationPeriodogramSyntheticSine injects a pure sinusoid at a
// known period and verifies the periodogram peak lands at that period bin.
func TestAutoCorrelationPeriodogramSyntheticSine(t *testing.T) {
	t.Parallel()

	const (
		period = 20.0
		bars   = 600
	)

	x, err := NewAutoCorrelationPeriodogramParams(&Params{
		DisableAutomaticGainControl: true,
		FixedNormalization:          true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t0 := testAcpTime()

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

	// Bin k corresponds to period MinPeriod+k. MinPeriod=10, period=20 -> bin 10.
	expectedBin := int(period - last.ParameterFirst)
	if peakBin != expectedBin {
		t.Errorf("peak bin: expected %d (period %.0f), got %d (period %.0f)",
			expectedBin, period, peakBin, last.ParameterFirst+float64(peakBin))
	}
}

func TestAutoCorrelationPeriodogramNaNInput(t *testing.T) {
	t.Parallel()

	x, _ := NewAutoCorrelationPeriodogramDefault()

	h := x.Update(math.NaN(), testAcpTime())

	if h == nil || !h.IsEmpty() {
		t.Errorf("expected empty heatmap for NaN input, got %v", h)
	}

	if x.IsPrimed() {
		t.Error("NaN input must not prime the indicator")
	}
}

func TestAutoCorrelationPeriodogramMetadata(t *testing.T) {
	t.Parallel()

	x, _ := NewAutoCorrelationPeriodogramDefault()
	md := x.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s: expected %v, actual %v", what, exp, act)
		}
	}

	mn := "acp(10, 48, hl/2)"

	check("Identifier", core.AutoCorrelationPeriodogram, md.Identifier)
	check("Mnemonic", mn, md.Mnemonic)
	check("Description", "Autocorrelation periodogram "+mn, md.Description)
	check("len(Outputs)", 1, len(md.Outputs))
	check("Outputs[0].Kind", int(Value), md.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Heatmap, md.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", mn, md.Outputs[0].Mnemonic)
}

//nolint:funlen
func TestAutoCorrelationPeriodogramMnemonicFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    Params
		mn   string
	}{
		{"default", Params{}, "acp(10, 48, hl/2)"},
		{"average override", Params{AveragingLength: 5}, "acp(10, 48, average=5, hl/2)"},
		{
			"no-sqr",
			Params{DisableSpectralSquaring: true},
			"acp(10, 48, no-sqr, hl/2)",
		},
		{
			"no-smooth",
			Params{DisableSmoothing: true},
			"acp(10, 48, no-smooth, hl/2)",
		},
		{
			"no-agc",
			Params{DisableAutomaticGainControl: true},
			"acp(10, 48, no-agc, hl/2)",
		},
		{
			"agc override",
			Params{AutomaticGainControlDecayFactor: 0.8},
			"acp(10, 48, agc=0.8, hl/2)",
		},
		{"no-fn", Params{FixedNormalization: true}, "acp(10, 48, no-fn, hl/2)"},
		{
			"all flags",
			Params{
				AveragingLength:             5,
				DisableSpectralSquaring:     true,
				DisableSmoothing:            true,
				DisableAutomaticGainControl: true,
				FixedNormalization:          true,
			},
			"acp(10, 48, average=5, no-sqr, no-smooth, no-agc, no-fn, hl/2)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			x, err := NewAutoCorrelationPeriodogramParams(&tt.p)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if x.mnemonic != tt.mn {
				t.Errorf("expected %q, got %q", tt.mn, x.mnemonic)
			}
		})
	}
}

//nolint:funlen
func TestAutoCorrelationPeriodogramValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    Params
		msg  string
	}{
		{
			"MinPeriod < 2",
			Params{MinPeriod: 1, MaxPeriod: 48, AveragingLength: 3},
			"invalid autocorrelation periodogram parameters: MinPeriod should be >= 2",
		},
		{
			"MaxPeriod <= MinPeriod",
			Params{MinPeriod: 10, MaxPeriod: 10, AveragingLength: 3},
			"invalid autocorrelation periodogram parameters: MaxPeriod should be > MinPeriod",
		},
		{
			"AveragingLength < 1",
			Params{AveragingLength: -1},
			"invalid autocorrelation periodogram parameters: AveragingLength should be >= 1",
		},
		{
			"AGC decay <= 0",
			Params{AutomaticGainControlDecayFactor: -0.1},
			"invalid autocorrelation periodogram parameters: AutomaticGainControlDecayFactor should be in (0, 1)",
		},
		{
			"AGC decay >= 1",
			Params{AutomaticGainControlDecayFactor: 1.0},
			"invalid autocorrelation periodogram parameters: AutomaticGainControlDecayFactor should be in (0, 1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewAutoCorrelationPeriodogramParams(&tt.p)
			if err == nil || err.Error() != tt.msg {
				t.Errorf("expected %q, got %v", tt.msg, err)
			}
		})
	}
}

func TestAutoCorrelationPeriodogramInvalidBarComponent(t *testing.T) {
	t.Parallel()

	_, err := NewAutoCorrelationPeriodogramParams(&Params{BarComponent: entities.BarComponent(9999)})
	if err == nil {
		t.Error("expected error")
	}
}

//nolint:funlen
func TestAutoCorrelationPeriodogramUpdateEntity(t *testing.T) {
	t.Parallel()

	const (
		primeCount = 100
		inp        = 100.
		outputLen  = 1
	)

	tm := testAcpTime()
	input := testAcpInput()

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
		x, _ := NewAutoCorrelationPeriodogramDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}

		check(x.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: tm, High: inp, Low: inp, Close: inp}
		x, _ := NewAutoCorrelationPeriodogramDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}

		check(x.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: tm, Bid: inp, Ask: inp}
		x, _ := NewAutoCorrelationPeriodogramDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}

		check(x.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: tm, Price: inp}
		x, _ := NewAutoCorrelationPeriodogramDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}

		check(x.UpdateTrade(&r))
	})
}
