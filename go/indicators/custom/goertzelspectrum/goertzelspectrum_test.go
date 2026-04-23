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

func testGSTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

// testGSInput is the 252-entry TA-Lib MAMA reference series (Price D5…D256).
//
//nolint:dupl
func testGSInput() []float64 {
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

const testGSTolerance = 1e-10

// spotValue represents a single (index, value) pair inside a heatmap column.
type spotValue struct {
	i int
	v float64
}

// gsSnap is a locked snapshot for a given input index.
type gsSnap struct {
	i        int
	valueMin float64
	valueMax float64
	spots    []spotValue
}

// snapshots were captured from the Go implementation and hand-verified at i=63
// against an independent Python implementation of the Goertzel spectrum (match
// better than 1e-14).
//
//nolint:gochecknoglobals
var goertzelSnapshots = []gsSnap{
	{
		i: 63, valueMin: 0, valueMax: 1,
		spots: []spotValue{
			{0, 0.002212390126817},
			{15, 0.393689637083521},
			{31, 0.561558825583766},
			{47, 0.486814514368002},
			{62, 0.487856217300954},
		},
	},
	{
		i: 64, valueMin: 0, valueMax: 0.9945044963,
		spots: []spotValue{
			{0, 0.006731833921830},
			{15, 0.435945652220356},
			{31, 0.554419782890674},
			{47, 0.489761317874540},
			{62, 0.490802995079533},
		},
	},
	{
		i: 100, valueMin: 0, valueMax: 1,
		spots: []spotValue{
			{0, 0.008211812272033},
			{15, 0.454499290767355},
			{31, 0.450815700228196},
			{47, 0.432349912501093},
			{62, 1.0},
		},
	},
	{
		i: 150, valueMin: 0, valueMax: 0.4526639264,
		spots: []spotValue{
			{0, 0.003721075091811},
			{15, 0.050467362919035},
			{31, 0.053328277804150},
			{47, 0.351864884608844},
			{62, 0.451342692411903},
		},
	},
	{
		i: 200, valueMin: 0, valueMax: 0.5590969243,
		spots: []spotValue{
			{0, 0.041810380001389},
			{15, 0.388762084039364},
			{31, 0.412461432112096},
			{47, 0.446271463994143},
			{62, 0.280061782526868},
		},
	},
}

// Relaxed tolerance for the valueMax-only checks (values captured to 10 sf).
const testGSMinMaxTol = 1e-9

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
