//nolint:testpackage
package maximumentropyspectrum

import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
	"zpano/indicators/core/outputs/shape"
)

func TestMaximumEntropySpectrumUpdate(t *testing.T) {
	t.Parallel()

	input := testMesInput()
	t0 := testMesTime()

	x, err := NewMaximumEntropySpectrumDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	si := 0

	for i := range input {
		h := x.Update(input[i], t0.Add(time.Duration(i)*time.Minute))
		if h == nil {
			t.Fatalf("[%d] heatmap must not be nil", i)
		}

		if h.ParameterFirst != 2 || h.ParameterLast != 59 || h.ParameterResolution != 1 {
			t.Errorf("[%d] axis incorrect: first=%v last=%v result=%v",
				i, h.ParameterFirst, h.ParameterLast, h.ParameterResolution)
		}

		if !x.IsPrimed() {
			if !h.IsEmpty() {
				t.Errorf("[%d] expected empty heatmap before priming, got len=%d", i, len(h.Values))
			}

			continue
		}

		if len(h.Values) != 58 {
			t.Errorf("[%d] expected values len=58, got %d", i, len(h.Values))
		}

		if si < len(mesSnapshots) && mesSnapshots[si].i == i {
			snap := mesSnapshots[si]
			if math.Abs(h.ValueMin-snap.valueMin) > testMesMinMaxTol {
				t.Errorf("[%d] ValueMin: expected %v, got %v", i, snap.valueMin, h.ValueMin)
			}

			if math.Abs(h.ValueMax-snap.valueMax) > testMesMinMaxTol {
				t.Errorf("[%d] ValueMax: expected %v, got %v", i, snap.valueMax, h.ValueMax)
			}

			for _, sp := range snap.spots {
				if math.Abs(h.Values[sp.i]-sp.v) > testMesTolerance {
					t.Errorf("[%d] Values[%d]: expected %v, got %v", i, sp.i, sp.v, h.Values[sp.i])
				}
			}

			si++
		}
	}

	if si != len(mesSnapshots) {
		t.Errorf("did not hit all %d snapshots, reached %d", len(mesSnapshots), si)
	}
}

func TestMaximumEntropySpectrumPrimesAtBar59(t *testing.T) {
	t.Parallel()

	x, _ := NewMaximumEntropySpectrumDefault()
	if x.IsPrimed() {
		t.Error("expected not primed at start")
	}

	input := testMesInput()
	t0 := testMesTime()
	primedAt := -1

	for i := range input {
		x.Update(input[i], t0.Add(time.Duration(i)*time.Minute))

		if x.IsPrimed() && primedAt < 0 {
			primedAt = i
		}
	}

	if primedAt != 59 {
		t.Errorf("expected priming at index 59, got %d", primedAt)
	}
}

func TestMaximumEntropySpectrumNaNInput(t *testing.T) {
	t.Parallel()

	x, _ := NewMaximumEntropySpectrumDefault()

	h := x.Update(math.NaN(), testMesTime())

	if h == nil || !h.IsEmpty() {
		t.Errorf("expected empty heatmap for NaN input, got %v", h)
	}

	if x.IsPrimed() {
		t.Error("NaN input must not prime the indicator")
	}
}

func TestMaximumEntropySpectrumMetadata(t *testing.T) {
	t.Parallel()

	x, _ := NewMaximumEntropySpectrumDefault()
	md := x.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s: expected %v, actual %v", what, exp, act)
		}
	}

	mn := "mespect(60, 30, 2, 59, 1, hl/2)"

	check("Identifier", core.MaximumEntropySpectrum, md.Identifier)
	check("Mnemonic", mn, md.Mnemonic)
	check("Description", "Maximum entropy spectrum "+mn, md.Description)
	check("len(Outputs)", 1, len(md.Outputs))
	check("Outputs[0].Kind", int(Value), md.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Heatmap, md.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", mn, md.Outputs[0].Mnemonic)
}

//nolint:funlen
func TestMaximumEntropySpectrumMnemonicFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    Params
		mn   string
	}{
		{"default", Params{}, "mespect(60, 30, 2, 59, 1, hl/2)"},
		{"no-agc", Params{DisableAutomaticGainControl: true}, "mespect(60, 30, 2, 59, 1, no-agc, hl/2)"},
		{
			"agc override",
			Params{AutomaticGainControlDecayFactor: 0.8},
			"mespect(60, 30, 2, 59, 1, agc=0.8, hl/2)",
		},
		{"no-fn", Params{FixedNormalization: true}, "mespect(60, 30, 2, 59, 1, no-fn, hl/2)"},
		{
			"all flags",
			Params{
				DisableAutomaticGainControl: true,
				FixedNormalization:          true,
			},
			"mespect(60, 30, 2, 59, 1, no-agc, no-fn, hl/2)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			x, err := NewMaximumEntropySpectrumParams(&tt.p)
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
func TestMaximumEntropySpectrumValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    Params
		msg  string
	}{
		{
			"Length < 2",
			Params{Length: 1, Degree: 1, MinPeriod: 2, MaxPeriod: 4, SpectrumResolution: 1},
			"invalid maximum entropy spectrum parameters: Length should be >= 2",
		},
		{
			"Degree >= Length",
			Params{Length: 4, Degree: 4, MinPeriod: 2, MaxPeriod: 4, SpectrumResolution: 1},
			"invalid maximum entropy spectrum parameters: Degree should be > 0 and < Length",
		},
		{
			"MinPeriod < 2",
			Params{Length: 60, Degree: 30, MinPeriod: 1, MaxPeriod: 59, SpectrumResolution: 1},
			"invalid maximum entropy spectrum parameters: MinPeriod should be >= 2",
		},
		{
			"MaxPeriod <= MinPeriod",
			Params{Length: 60, Degree: 30, MinPeriod: 10, MaxPeriod: 10, SpectrumResolution: 1},
			"invalid maximum entropy spectrum parameters: MaxPeriod should be > MinPeriod",
		},
		{
			"MaxPeriod > 2*Length",
			Params{Length: 10, Degree: 5, MinPeriod: 2, MaxPeriod: 59, SpectrumResolution: 1},
			"invalid maximum entropy spectrum parameters: MaxPeriod should be <= 2 * Length",
		},
		{
			"AGC decay <= 0",
			Params{AutomaticGainControlDecayFactor: -0.1},
			"invalid maximum entropy spectrum parameters: AutomaticGainControlDecayFactor should be in (0, 1)",
		},
		{
			"AGC decay >= 1",
			Params{AutomaticGainControlDecayFactor: 1.0},
			"invalid maximum entropy spectrum parameters: AutomaticGainControlDecayFactor should be in (0, 1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewMaximumEntropySpectrumParams(&tt.p)
			if err == nil || err.Error() != tt.msg {
				t.Errorf("expected %q, got %v", tt.msg, err)
			}
		})
	}
}

func TestMaximumEntropySpectrumInvalidBarComponent(t *testing.T) {
	t.Parallel()

	_, err := NewMaximumEntropySpectrumParams(&Params{BarComponent: entities.BarComponent(9999)})
	if err == nil {
		t.Error("expected error")
	}
}

//nolint:funlen
func TestMaximumEntropySpectrumUpdateEntity(t *testing.T) {
	t.Parallel()

	const (
		primeCount = 70
		inp        = 100.
		outputLen  = 1
	)

	tm := testMesTime()
	input := testMesInput()

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
		x, _ := NewMaximumEntropySpectrumDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}

		check(x.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: tm, High: inp, Low: inp, Close: inp}
		x, _ := NewMaximumEntropySpectrumDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}

		check(x.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: tm, Bid: inp, Ask: inp}
		x, _ := NewMaximumEntropySpectrumDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}

		check(x.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: tm, Price: inp}
		x, _ := NewMaximumEntropySpectrumDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)], tm)
		}

		check(x.UpdateTrade(&r))
	})
}
