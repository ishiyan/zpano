//nolint:testpackage
package combbandpassspectrum

import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
	"zpano/indicators/core/outputs/shape"
)

func TestCombBandPassSpectrumUpdate(t *testing.T) {
	t.Parallel()

	input := testCbpsInput()
	t0 := testCbpsTime()

	x, err := NewCombBandPassSpectrumDefault()
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

		if si < len(cbpsSnapshots) && cbpsSnapshots[si].i == i {
			snap := cbpsSnapshots[si]
			if math.Abs(h.ValueMin-snap.valueMin) > testCbpsMinMaxTol {
				t.Errorf("[%d] ValueMin: expected %v, got %v", i, snap.valueMin, h.ValueMin)
			}

			if math.Abs(h.ValueMax-snap.valueMax) > testCbpsMinMaxTol {
				t.Errorf("[%d] ValueMax: expected %v, got %v", i, snap.valueMax, h.ValueMax)
			}

			for _, sp := range snap.spots {
				if math.Abs(h.Values[sp.i]-sp.v) > testCbpsTolerance {
					t.Errorf("[%d] Values[%d]: expected %v, got %v", i, sp.i, sp.v, h.Values[sp.i])
				}
			}

			si++
		}
	}

	if si != len(cbpsSnapshots) {
		t.Errorf("did not hit all %d snapshots, reached %d", len(cbpsSnapshots), si)
	}
}

func TestCombBandPassSpectrumPrimesAtBar47(t *testing.T) {
	t.Parallel()

	x, _ := NewCombBandPassSpectrumDefault()
	if x.IsPrimed() {
		t.Error("expected not primed at start")
	}

	input := testCbpsInput()
	t0 := testCbpsTime()
	primedAt := -1

	for i := range input {
		x.Update(input[i], t0.Add(time.Duration(i)*time.Minute))

		if x.IsPrimed() && primedAt < 0 {
			primedAt = i
		}
	}

	if primedAt != 47 {
		t.Errorf("expected priming at index 47, got %d", primedAt)
	}
}

func TestCombBandPassSpectrumNaNInput(t *testing.T) {
	t.Parallel()

	x, _ := NewCombBandPassSpectrumDefault()

	h := x.Update(math.NaN(), testCbpsTime())

	if h == nil || !h.IsEmpty() {
		t.Errorf("expected empty heatmap for NaN input, got %v", h)
	}

	if x.IsPrimed() {
		t.Error("NaN input must not prime the indicator")
	}
}

// TestCombBandPassSpectrumSyntheticSine injects a pure sinusoid at a known
// period and verifies the spectrum peak lands at that period bin. The
// band-pass filter tuned to period P resonates strongly when driven by a
// sinusoid of period P, so this provides a clean independent sanity check
// of the Ehlers comb filter bank.
func TestCombBandPassSpectrumSyntheticSine(t *testing.T) {
	t.Parallel()

	const (
		period = 20.0 // Mid-range period well within [10, 48].
		bars   = 400  // Enough to let AGC settle and BP filters stabilize.
	)

	x, err := NewCombBandPassSpectrumParams(&Params{
		DisableSpectralDilationCompensation: true,
		DisableAutomaticGainControl:         true,
		FixedNormalization:                  true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t0 := testCbpsTime()

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

func TestCombBandPassSpectrumMetadata(t *testing.T) {
	t.Parallel()

	x, _ := NewCombBandPassSpectrumDefault()
	md := x.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s: expected %v, actual %v", what, exp, act)
		}
	}

	mn := "cbps(10, 48, hl/2)"

	check("Identifier", core.CombBandPassSpectrum, md.Identifier)
	check("Mnemonic", mn, md.Mnemonic)
	check("Description", "Comb band-pass spectrum "+mn, md.Description)
	check("len(Outputs)", 1, len(md.Outputs))
	check("Outputs[0].Kind", int(Value), md.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Heatmap, md.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", mn, md.Outputs[0].Mnemonic)
}

//nolint:funlen
func TestCombBandPassSpectrumMnemonicFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    Params
		mn   string
	}{
		{"default", Params{}, "cbps(10, 48, hl/2)"},
		{
			"bandwidth override",
			Params{Bandwidth: 0.5},
			"cbps(10, 48, bw=0.5, hl/2)",
		},
		{
			"no-sdc",
			Params{DisableSpectralDilationCompensation: true},
			"cbps(10, 48, no-sdc, hl/2)",
		},
		{
			"no-agc",
			Params{DisableAutomaticGainControl: true},
			"cbps(10, 48, no-agc, hl/2)",
		},
		{
			"agc override",
			Params{AutomaticGainControlDecayFactor: 0.8},
			"cbps(10, 48, agc=0.8, hl/2)",
		},
		{"no-fn", Params{FixedNormalization: true}, "cbps(10, 48, no-fn, hl/2)"},
		{
			"all flags",
			Params{
				Bandwidth:                           0.5,
				DisableSpectralDilationCompensation: true,
				DisableAutomaticGainControl:         true,
				FixedNormalization:                  true,
			},
			"cbps(10, 48, bw=0.5, no-sdc, no-agc, no-fn, hl/2)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			x, err := NewCombBandPassSpectrumParams(&tt.p)
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
func TestCombBandPassSpectrumValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    Params
		msg  string
	}{
		{
			"MinPeriod < 2",
			Params{MinPeriod: 1, MaxPeriod: 48, Bandwidth: 0.3},
			"invalid comb band-pass spectrum parameters: MinPeriod should be >= 2",
		},
		{
			"MaxPeriod <= MinPeriod",
			Params{MinPeriod: 10, MaxPeriod: 10, Bandwidth: 0.3},
			"invalid comb band-pass spectrum parameters: MaxPeriod should be > MinPeriod",
		},
		{
			"Bandwidth <= 0",
			Params{Bandwidth: -0.1},
			"invalid comb band-pass spectrum parameters: Bandwidth should be in (0, 1)",
		},
		{
			"Bandwidth >= 1",
			Params{Bandwidth: 1.0},
			"invalid comb band-pass spectrum parameters: Bandwidth should be in (0, 1)",
		},
		{
			"AGC decay <= 0",
			Params{AutomaticGainControlDecayFactor: -0.1},
			"invalid comb band-pass spectrum parameters: AutomaticGainControlDecayFactor should be in (0, 1)",
		},
		{
			"AGC decay >= 1",
			Params{AutomaticGainControlDecayFactor: 1.0},
			"invalid comb band-pass spectrum parameters: AutomaticGainControlDecayFactor should be in (0, 1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewCombBandPassSpectrumParams(&tt.p)
			if err == nil || err.Error() != tt.msg {
				t.Errorf("expected %q, got %v", tt.msg, err)
			}
		})
	}
}

func TestCombBandPassSpectrumInvalidBarComponent(t *testing.T) {
	t.Parallel()

	_, err := NewCombBandPassSpectrumParams(&Params{BarComponent: entities.BarComponent(9999)})
	if err == nil {
		t.Error("expected error")
	}
}

//nolint:funlen
func TestCombBandPassSpectrumUpdateEntity(t *testing.T) {
	t.Parallel()

	const (
		primeCount = 60
		inp        = 100.
		outputLen  = 1
	)

	tm := testCbpsTime()
	input := testCbpsInput()

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
		x, _ := NewCombBandPassSpectrumDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}

		check(x.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: tm, High: inp, Low: inp, Close: inp}
		x, _ := NewCombBandPassSpectrumDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}

		check(x.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: tm, Bid: inp, Ask: inp}
		x, _ := NewCombBandPassSpectrumDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}

		check(x.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: tm, Price: inp}
		x, _ := NewCombBandPassSpectrumDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}

		check(x.UpdateTrade(&r))
	})
}
