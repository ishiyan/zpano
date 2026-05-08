//nolint:testpackage
package discretefouriertransformspectrum

import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
	"zpano/indicators/core/outputs/shape"
)

func TestDiscreteFourierTransformSpectrumUpdate(t *testing.T) {
	t.Parallel()

	input := testDftsInput()
	t0 := testDftsTime()

	x, err := NewDiscreteFourierTransformSpectrumDefault()
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

		if si < len(dftsSnapshots) && dftsSnapshots[si].i == i {
			snap := dftsSnapshots[si]
			if math.Abs(h.ValueMin-snap.valueMin) > testDftsMinMaxTol {
				t.Errorf("[%d] ValueMin: expected %v, got %v", i, snap.valueMin, h.ValueMin)
			}

			if math.Abs(h.ValueMax-snap.valueMax) > testDftsMinMaxTol {
				t.Errorf("[%d] ValueMax: expected %v, got %v", i, snap.valueMax, h.ValueMax)
			}

			for _, sp := range snap.spots {
				if math.Abs(h.Values[sp.i]-sp.v) > testDftsTolerance {
					t.Errorf("[%d] Values[%d]: expected %v, got %v", i, sp.i, sp.v, h.Values[sp.i])
				}
			}

			si++
		}
	}

	if si != len(dftsSnapshots) {
		t.Errorf("did not hit all %d snapshots, reached %d", len(dftsSnapshots), si)
	}
}

func TestDiscreteFourierTransformSpectrumPrimesAtBar47(t *testing.T) {
	t.Parallel()

	x, _ := NewDiscreteFourierTransformSpectrumDefault()
	if x.IsPrimed() {
		t.Error("expected not primed at start")
	}

	input := testDftsInput()
	t0 := testDftsTime()
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

func TestDiscreteFourierTransformSpectrumNaNInput(t *testing.T) {
	t.Parallel()

	x, _ := NewDiscreteFourierTransformSpectrumDefault()

	h := x.Update(math.NaN(), testDftsTime())

	if h == nil || !h.IsEmpty() {
		t.Errorf("expected empty heatmap for NaN input, got %v", h)
	}

	if x.IsPrimed() {
		t.Error("NaN input must not prime the indicator")
	}
}

// TestDiscreteFourierTransformSpectrumSyntheticSine injects a pure sinusoid at
// a known period and verifies the spectrum peak lands at that period bin. This
// provides an independent sanity check on the DFT math, since the DFTS algorithm
// as implemented in MBST deviates from Ehlers' EasyLanguage listing 9-1 (which
// additionally highpass + Super Smoother filters its input).
func TestDiscreteFourierTransformSpectrumSyntheticSine(t *testing.T) {
	t.Parallel()

	const (
		period = 16.0 // 3 integer cycles in the default length=48 window (no DFT leakage).
		bars   = 200
	)

	// Disable AGC/SDC/FloatingNormalization so the peak reflects the raw DFT magnitude.
	x, err := NewDiscreteFourierTransformSpectrumParams(&Params{
		DisableSpectralDilationCompensation: true,
		DisableAutomaticGainControl:         true,
		FixedNormalization:                  true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t0 := testDftsTime()

	var last *outputs.Heatmap

	for i := 0; i < bars; i++ {
		sample := 100 + math.Sin(2*math.Pi*float64(i)/period)
		last = x.Update(sample, t0.Add(time.Duration(i)*time.Minute))
	}

	if last == nil || last.IsEmpty() {
		t.Fatal("expected primed non-empty heatmap")
	}

	// Peak bin should correspond to period=16. Axis is MinPeriod..MaxPeriod step 1,
	// so bin k corresponds to period MinPeriod+k. With defaults MinPeriod=10,
	// period=16 -> bin index 6.
	peakBin := 0
	for i := range last.Values {
		if last.Values[i] > last.Values[peakBin] {
			peakBin = i
		}
	}

	expectedBin := int(period - last.ParameterFirst)
	if peakBin != expectedBin {
		t.Errorf("peak bin: expected %d (period %.0f), got %d (period %.0f)",
			expectedBin, period, peakBin, last.ParameterFirst+float64(peakBin))
	}
}

func TestDiscreteFourierTransformSpectrumMetadata(t *testing.T) {
	t.Parallel()

	x, _ := NewDiscreteFourierTransformSpectrumDefault()
	md := x.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s: expected %v, actual %v", what, exp, act)
		}
	}

	mn := "dftps(48, 10, 48, 1, hl/2)"

	check("Identifier", core.DiscreteFourierTransformSpectrum, md.Identifier)
	check("Mnemonic", mn, md.Mnemonic)
	check("Description", "Discrete Fourier transform spectrum "+mn, md.Description)
	check("len(Outputs)", 1, len(md.Outputs))
	check("Outputs[0].Kind", int(Value), md.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Heatmap, md.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", mn, md.Outputs[0].Mnemonic)
}

//nolint:funlen
func TestDiscreteFourierTransformSpectrumMnemonicFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    Params
		mn   string
	}{
		{"default", Params{}, "dftps(48, 10, 48, 1, hl/2)"},
		{
			"no-sdc",
			Params{DisableSpectralDilationCompensation: true},
			"dftps(48, 10, 48, 1, no-sdc, hl/2)",
		},
		{
			"no-agc",
			Params{DisableAutomaticGainControl: true},
			"dftps(48, 10, 48, 1, no-agc, hl/2)",
		},
		{
			"agc override",
			Params{AutomaticGainControlDecayFactor: 0.8},
			"dftps(48, 10, 48, 1, agc=0.8, hl/2)",
		},
		{"no-fn", Params{FixedNormalization: true}, "dftps(48, 10, 48, 1, no-fn, hl/2)"},
		{
			"all flags",
			Params{
				DisableSpectralDilationCompensation: true,
				DisableAutomaticGainControl:         true,
				FixedNormalization:                  true,
			},
			"dftps(48, 10, 48, 1, no-sdc, no-agc, no-fn, hl/2)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			x, err := NewDiscreteFourierTransformSpectrumParams(&tt.p)
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
func TestDiscreteFourierTransformSpectrumValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    Params
		msg  string
	}{
		{
			"Length < 2",
			Params{Length: 1, MinPeriod: 10, MaxPeriod: 48, SpectrumResolution: 1},
			"invalid discrete Fourier transform spectrum parameters: Length should be >= 2",
		},
		{
			"MinPeriod < 2",
			Params{Length: 48, MinPeriod: 1, MaxPeriod: 48, SpectrumResolution: 1},
			"invalid discrete Fourier transform spectrum parameters: MinPeriod should be >= 2",
		},
		{
			"MaxPeriod <= MinPeriod",
			Params{Length: 48, MinPeriod: 10, MaxPeriod: 10, SpectrumResolution: 1},
			"invalid discrete Fourier transform spectrum parameters: MaxPeriod should be > MinPeriod",
		},
		{
			"MaxPeriod > 2*Length",
			Params{Length: 10, MinPeriod: 2, MaxPeriod: 48, SpectrumResolution: 1},
			"invalid discrete Fourier transform spectrum parameters: MaxPeriod should be <= 2 * Length",
		},
		{
			"AGC decay <= 0",
			Params{AutomaticGainControlDecayFactor: -0.1},
			"invalid discrete Fourier transform spectrum parameters: AutomaticGainControlDecayFactor should be in (0, 1)",
		},
		{
			"AGC decay >= 1",
			Params{AutomaticGainControlDecayFactor: 1.0},
			"invalid discrete Fourier transform spectrum parameters: AutomaticGainControlDecayFactor should be in (0, 1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewDiscreteFourierTransformSpectrumParams(&tt.p)
			if err == nil || err.Error() != tt.msg {
				t.Errorf("expected %q, got %v", tt.msg, err)
			}
		})
	}
}

func TestDiscreteFourierTransformSpectrumInvalidBarComponent(t *testing.T) {
	t.Parallel()

	_, err := NewDiscreteFourierTransformSpectrumParams(&Params{BarComponent: entities.BarComponent(9999)})
	if err == nil {
		t.Error("expected error")
	}
}

//nolint:funlen
func TestDiscreteFourierTransformSpectrumUpdateEntity(t *testing.T) {
	t.Parallel()

	const (
		primeCount = 60
		inp        = 100.
		outputLen  = 1
	)

	tm := testDftsTime()
	input := testDftsInput()

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
		x, _ := NewDiscreteFourierTransformSpectrumDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}

		check(x.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: tm, High: inp, Low: inp, Close: inp}
		x, _ := NewDiscreteFourierTransformSpectrumDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}

		check(x.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: tm, Bid: inp, Ask: inp}
		x, _ := NewDiscreteFourierTransformSpectrumDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}

		check(x.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: tm, Price: inp}
		x, _ := NewDiscreteFourierTransformSpectrumDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}

		check(x.UpdateTrade(&r))
	})
}
