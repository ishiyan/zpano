//nolint:testpackage
package goertzelspectrum

import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
	"zpano/indicators/core/outputs/shape"
)

func TestGoertzelSpectrumUpdate(t *testing.T) {
	t.Parallel()

	input := testGSInput()
	t0 := testGSTime()

	x, err := NewGoertzelSpectrumDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	si := 0

	for i := range input {
		h := x.Update(input[i], t0.Add(time.Duration(i)*time.Minute))
		if h == nil {
			t.Fatalf("[%d] heatmap must not be nil", i)
		}

		if h.ParameterFirst != 2 || h.ParameterLast != 64 || h.ParameterResolution != 1 {
			t.Errorf("[%d] axis incorrect: first=%v last=%v result=%v",
				i, h.ParameterFirst, h.ParameterLast, h.ParameterResolution)
		}

		if !x.IsPrimed() {
			if !h.IsEmpty() {
				t.Errorf("[%d] expected empty heatmap before priming, got len=%d", i, len(h.Values))
			}

			continue
		}

		if len(h.Values) != 63 {
			t.Errorf("[%d] expected values len=63, got %d", i, len(h.Values))
		}

		if si < len(goertzelSnapshots) && goertzelSnapshots[si].i == i {
			snap := goertzelSnapshots[si]
			if math.Abs(h.ValueMin-snap.valueMin) > testGSMinMaxTol {
				t.Errorf("[%d] ValueMin: expected %v, got %v", i, snap.valueMin, h.ValueMin)
			}

			if math.Abs(h.ValueMax-snap.valueMax) > testGSMinMaxTol {
				t.Errorf("[%d] ValueMax: expected %v, got %v", i, snap.valueMax, h.ValueMax)
			}

			for _, sp := range snap.spots {
				if math.Abs(h.Values[sp.i]-sp.v) > testGSTolerance {
					t.Errorf("[%d] Values[%d]: expected %v, got %v", i, sp.i, sp.v, h.Values[sp.i])
				}
			}

			si++
		}
	}

	if si != len(goertzelSnapshots) {
		t.Errorf("did not hit all %d snapshots, reached %d", len(goertzelSnapshots), si)
	}
}

func TestGoertzelSpectrumPrimesAtBar63(t *testing.T) {
	t.Parallel()

	x, _ := NewGoertzelSpectrumDefault()
	if x.IsPrimed() {
		t.Error("expected not primed at start")
	}

	input := testGSInput()
	t0 := testGSTime()
	primedAt := -1

	for i := range input {
		x.Update(input[i], t0.Add(time.Duration(i)*time.Minute))

		if x.IsPrimed() && primedAt < 0 {
			primedAt = i
		}
	}

	if primedAt != 63 {
		t.Errorf("expected priming at index 63, got %d", primedAt)
	}
}

func TestGoertzelSpectrumNaNInput(t *testing.T) {
	t.Parallel()

	x, _ := NewGoertzelSpectrumDefault()

	h := x.Update(math.NaN(), testGSTime())

	if h == nil || !h.IsEmpty() {
		t.Errorf("expected empty heatmap for NaN input, got %v", h)
	}

	if x.IsPrimed() {
		t.Error("NaN input must not prime the indicator")
	}
}

func TestGoertzelSpectrumMetadata(t *testing.T) {
	t.Parallel()

	x, _ := NewGoertzelSpectrumDefault()
	md := x.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s: expected %v, actual %v", what, exp, act)
		}
	}

	mn := "gspect(64, 2, 64, 1, hl/2)"

	check("Identifier", core.GoertzelSpectrum, md.Identifier)
	check("Mnemonic", mn, md.Mnemonic)
	check("Description", "Goertzel spectrum "+mn, md.Description)
	check("len(Outputs)", 1, len(md.Outputs))
	check("Outputs[0].Kind", int(Value), md.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Heatmap, md.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", mn, md.Outputs[0].Mnemonic)
}

//nolint:funlen
func TestGoertzelSpectrumMnemonicFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    Params
		mn   string
	}{
		{"default", Params{}, "gspect(64, 2, 64, 1, hl/2)"},
		{"first-order", Params{IsFirstOrder: true}, "gspect(64, 2, 64, 1, fo, hl/2)"},
		{"no-sdc", Params{DisableSpectralDilationCompensation: true}, "gspect(64, 2, 64, 1, no-sdc, hl/2)"},
		{"no-agc", Params{DisableAutomaticGainControl: true}, "gspect(64, 2, 64, 1, no-agc, hl/2)"},
		{
			"agc override",
			Params{AutomaticGainControlDecayFactor: 0.8},
			"gspect(64, 2, 64, 1, agc=0.8, hl/2)",
		},
		{"no-fn", Params{FixedNormalization: true}, "gspect(64, 2, 64, 1, no-fn, hl/2)"},
		{
			"all flags",
			Params{
				IsFirstOrder:                        true,
				DisableSpectralDilationCompensation: true,
				DisableAutomaticGainControl:         true,
				FixedNormalization:                  true,
			},
			"gspect(64, 2, 64, 1, fo, no-sdc, no-agc, no-fn, hl/2)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			x, err := NewGoertzelSpectrumParams(&tt.p)
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
func TestGoertzelSpectrumValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    Params
		msg  string
	}{
		{
			"Length < 2",
			Params{Length: 1, MinPeriod: 2, MaxPeriod: 64, SpectrumResolution: 1},
			"invalid goertzel spectrum parameters: Length should be >= 2",
		},
		{
			"MinPeriod < 2",
			Params{Length: 64, MinPeriod: 1, MaxPeriod: 64, SpectrumResolution: 1},
			"invalid goertzel spectrum parameters: MinPeriod should be >= 2",
		},
		{
			"MaxPeriod <= MinPeriod",
			Params{Length: 64, MinPeriod: 10, MaxPeriod: 10, SpectrumResolution: 1},
			"invalid goertzel spectrum parameters: MaxPeriod should be > MinPeriod",
		},
		{
			"MaxPeriod > 2*Length",
			Params{Length: 16, MinPeriod: 2, MaxPeriod: 64, SpectrumResolution: 1},
			"invalid goertzel spectrum parameters: MaxPeriod should be <= 2 * Length",
		},
		{
			"AGC decay <= 0",
			Params{AutomaticGainControlDecayFactor: -0.1},
			"invalid goertzel spectrum parameters: AutomaticGainControlDecayFactor should be in (0, 1)",
		},
		{
			"AGC decay >= 1",
			Params{AutomaticGainControlDecayFactor: 1.0},
			"invalid goertzel spectrum parameters: AutomaticGainControlDecayFactor should be in (0, 1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewGoertzelSpectrumParams(&tt.p)
			if err == nil || err.Error() != tt.msg {
				t.Errorf("expected %q, got %v", tt.msg, err)
			}
		})
	}
}

func TestGoertzelSpectrumInvalidBarComponent(t *testing.T) {
	t.Parallel()

	_, err := NewGoertzelSpectrumParams(&Params{BarComponent: entities.BarComponent(9999)})
	if err == nil {
		t.Error("expected error")
	}
}

//nolint:funlen
func TestGoertzelSpectrumUpdateEntity(t *testing.T) {
	t.Parallel()

	const (
		primeCount = 70
		inp        = 100.
		outputLen  = 1
	)

	tm := testGSTime()
	input := testGSInput()

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
		x, _ := NewGoertzelSpectrumDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}

		check(x.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: tm, High: inp, Low: inp, Close: inp}
		x, _ := NewGoertzelSpectrumDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}

		check(x.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: tm, Bid: inp, Ask: inp}
		x, _ := NewGoertzelSpectrumDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}

		check(x.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: tm, Price: inp}
		x, _ := NewGoertzelSpectrumDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}

		check(x.UpdateTrade(&r))
	})
}
