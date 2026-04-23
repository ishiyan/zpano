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

func testDftsTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, time.UTC)
}

// testDftsInput is the 252-entry TA-Lib MAMA reference series.
//
//nolint:dupl
func testDftsInput() []float64 {
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

const (
	testDftsTolerance = 1e-12
	testDftsMinMaxTol = 1e-10
)

type dftsSpot struct {
	i int
	v float64
}

type dftsSnap struct {
	i        int
	valueMin float64
	valueMax float64
	spots    []dftsSpot
}

// Snapshots captured from the Go implementation. The DFT math is additionally
// sanity-checked in TestDiscreteFourierTransformSpectrumSyntheticSine below
// (which verifies the peak bin lands at the injected period).
//
//nolint:gochecknoglobals
var dftsSnapshots = []dftsSnap{
	{
		i: 47, valueMin: 0, valueMax: 1,
		spots: []dftsSpot{
			{0, 0.092590275198666},
			{9, 0.027548278511410},
			{19, 0.157582436454944},
			{28, 0.818873586056752},
			{38, 0.994469163657834},
		},
	},
	{
		i: 60, valueMin: 0, valueMax: 1,
		spots: []dftsSpot{
			{0, 0.051096698360059},
			{9, 0.058377239289306},
			{19, 0.248187339190831},
			{28, 0.136096222727665},
			{38, 0.059069344951534},
		},
	},
	{
		i: 100, valueMin: 0, valueMax: 1,
		spots: []dftsSpot{
			{0, 0.150638336509665},
			{9, 0.046435987728045},
			{19, 0.103832850895319},
			{28, 0.402801132284104},
			{38, 1.000000000000000},
		},
	},
	{
		i: 150, valueMin: 0, valueMax: 0.5272269971142493,
		spots: []dftsSpot{
			{0, 0.000000000000000},
			{9, 0.091857989427651},
			{19, 0.219395988856534},
			{28, 0.516960894560452},
			{38, 0.468682020733700},
		},
	},
	{
		i: 200, valueMin: 0, valueMax: 0.6015223942655807,
		spots: []dftsSpot{
			{0, 0.107853213261092},
			{9, 0.164118955219278},
			{19, 0.306440501928972},
			{28, 0.569768020155262},
			{38, 0.585690371992475},
		},
	},
}

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
