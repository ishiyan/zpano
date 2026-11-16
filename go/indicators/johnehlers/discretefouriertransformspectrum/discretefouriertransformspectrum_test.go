//nolint:testpackage
package discretefouriertransformspectrum

import (
	"math"
	"slices"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

const (
	referenceTolerance = 1e-13
	oracleTolerance    = 2e-11
)

type referenceCase struct {
	name            string
	params          Params
	expectedHeatmap [][]float64
	expectedMin     []float64
	expectedMax     []float64
	expectedDC      []float64
}

func allReferenceCases() []referenceCase {
	return []referenceCase{
		{
			name:            "P010_P148_HP48_LP10_SDC1_E0_W48",
			params:          Params{MinPeriod: 10, MaxPeriod: 48, WindowLength: 48, HighpassPeriod: 48, SmoothingPeriod: 10, SpectralDilationCompensation: true, EnhanceResolution: false},
			expectedHeatmap: expectedP010_P148_HP48_LP10_SDC1_E0_W48_HEATMAP,
			expectedMin:     expectedP010_P148_HP48_LP10_SDC1_E0_W48_MIN,
			expectedMax:     expectedP010_P148_HP48_LP10_SDC1_E0_W48_MAX,
			expectedDC:      expectedP010_P148_HP48_LP10_SDC1_E0_W48_DOMINANT_CYCLE,
		},
		{
			name:            "P010_P148_HP48_LP10_SDC0_E0_W48",
			params:          Params{MinPeriod: 10, MaxPeriod: 48, WindowLength: 48, HighpassPeriod: 48, SmoothingPeriod: 10, SpectralDilationCompensation: false, EnhanceResolution: false},
			expectedHeatmap: expectedP010_P148_HP48_LP10_SDC0_E0_W48_HEATMAP,
			expectedMin:     expectedP010_P148_HP48_LP10_SDC0_E0_W48_MIN,
			expectedMax:     expectedP010_P148_HP48_LP10_SDC0_E0_W48_MAX,
			expectedDC:      expectedP010_P148_HP48_LP10_SDC0_E0_W48_DOMINANT_CYCLE,
		},
		{
			name:            "P010_P148_HP48_LP10_SDC1_E1_W48",
			params:          Params{MinPeriod: 10, MaxPeriod: 48, WindowLength: 48, HighpassPeriod: 48, SmoothingPeriod: 10, SpectralDilationCompensation: true, EnhanceResolution: true},
			expectedHeatmap: expectedP010_P148_HP48_LP10_SDC1_E1_W48_HEATMAP,
			expectedMin:     expectedP010_P148_HP48_LP10_SDC1_E1_W48_MIN,
			expectedMax:     expectedP010_P148_HP48_LP10_SDC1_E1_W48_MAX,
			expectedDC:      expectedP010_P148_HP48_LP10_SDC1_E1_W48_DOMINANT_CYCLE,
		},
		{
			name:            "P010_P148_HP48_LP10_SDC0_E1_W48",
			params:          Params{MinPeriod: 10, MaxPeriod: 48, WindowLength: 48, HighpassPeriod: 48, SmoothingPeriod: 10, SpectralDilationCompensation: false, EnhanceResolution: true},
			expectedHeatmap: expectedP010_P148_HP48_LP10_SDC0_E1_W48_HEATMAP,
			expectedMin:     expectedP010_P148_HP48_LP10_SDC0_E1_W48_MIN,
			expectedMax:     expectedP010_P148_HP48_LP10_SDC0_E1_W48_MAX,
			expectedDC:      expectedP010_P148_HP48_LP10_SDC0_E1_W48_DOMINANT_CYCLE,
		},
		{
			name:            "P010_P148_HP48_LP10_SDC1_E0_W24",
			params:          Params{MinPeriod: 10, MaxPeriod: 48, WindowLength: 24, HighpassPeriod: 48, SmoothingPeriod: 10, SpectralDilationCompensation: true, EnhanceResolution: false},
			expectedHeatmap: expectedP010_P148_HP48_LP10_SDC1_E0_W24_HEATMAP,
			expectedMin:     expectedP010_P148_HP48_LP10_SDC1_E0_W24_MIN,
			expectedMax:     expectedP010_P148_HP48_LP10_SDC1_E0_W24_MAX,
			expectedDC:      expectedP010_P148_HP48_LP10_SDC1_E0_W24_DOMINANT_CYCLE,
		},
		{
			name:            "P010_P148_HP48_LP10_SDC1_E0_W96",
			params:          Params{MinPeriod: 10, MaxPeriod: 48, WindowLength: 96, HighpassPeriod: 48, SmoothingPeriod: 10, SpectralDilationCompensation: true, EnhanceResolution: false},
			expectedHeatmap: expectedP010_P148_HP48_LP10_SDC1_E0_W96_HEATMAP,
			expectedMin:     expectedP010_P148_HP48_LP10_SDC1_E0_W96_MIN,
			expectedMax:     expectedP010_P148_HP48_LP10_SDC1_E0_W96_MAX,
			expectedDC:      expectedP010_P148_HP48_LP10_SDC1_E0_W96_DOMINANT_CYCLE,
		},
		{
			name:            "P010_P148_HP48_LP10_SDC1_E0_W252",
			params:          Params{MinPeriod: 10, MaxPeriod: 48, WindowLength: 252, HighpassPeriod: 48, SmoothingPeriod: 10, SpectralDilationCompensation: true, EnhanceResolution: false},
			expectedHeatmap: expectedP010_P148_HP48_LP10_SDC1_E0_W252_HEATMAP,
			expectedMin:     expectedP010_P148_HP48_LP10_SDC1_E0_W252_MIN,
			expectedMax:     expectedP010_P148_HP48_LP10_SDC1_E0_W252_MAX,
			expectedDC:      expectedP010_P148_HP48_LP10_SDC1_E0_W252_DOMINANT_CYCLE,
		},
		{
			name:            "P010_P148_HP48_LP10_5_SDC1_E0_W48",
			params:          Params{MinPeriod: 10, MaxPeriod: 48, WindowLength: 48, HighpassPeriod: 48, SmoothingPeriod: 10.5, SpectralDilationCompensation: true, EnhanceResolution: false},
			expectedHeatmap: expectedP010_P148_HP48_LP10_5_SDC1_E0_W48_HEATMAP,
			expectedMin:     expectedP010_P148_HP48_LP10_5_SDC1_E0_W48_MIN,
			expectedMax:     expectedP010_P148_HP48_LP10_5_SDC1_E0_W48_MAX,
			expectedDC:      expectedP010_P148_HP48_LP10_5_SDC1_E0_W48_DOMINANT_CYCLE,
		},
		{
			name:            "P010_P148_HP80_LP20_SDC1_E0_W48",
			params:          Params{MinPeriod: 10, MaxPeriod: 48, WindowLength: 48, HighpassPeriod: 80, SmoothingPeriod: 20, SpectralDilationCompensation: true, EnhanceResolution: false},
			expectedHeatmap: expectedP010_P148_HP80_LP20_SDC1_E0_W48_HEATMAP,
			expectedMin:     expectedP010_P148_HP80_LP20_SDC1_E0_W48_MIN,
			expectedMax:     expectedP010_P148_HP80_LP20_SDC1_E0_W48_MAX,
			expectedDC:      expectedP010_P148_HP80_LP20_SDC1_E0_W48_DOMINANT_CYCLE,
		},
		{
			name:            "P06_P124_HP24_LP6_SDC1_E0_W24",
			params:          Params{MinPeriod: 6, MaxPeriod: 24, WindowLength: 24, HighpassPeriod: 24, SmoothingPeriod: 6, SpectralDilationCompensation: true, EnhanceResolution: false},
			expectedHeatmap: expectedP06_P124_HP24_LP6_SDC1_E0_W24_HEATMAP,
			expectedMin:     expectedP06_P124_HP24_LP6_SDC1_E0_W24_MIN,
			expectedMax:     expectedP06_P124_HP24_LP6_SDC1_E0_W24_MAX,
			expectedDC:      expectedP06_P124_HP24_LP6_SDC1_E0_W24_DOMINANT_CYCLE,
		},
		{
			name:            "P020_P120_HP48_LP10_SDC1_E0_W48",
			params:          Params{MinPeriod: 20, MaxPeriod: 20, WindowLength: 48, HighpassPeriod: 48, SmoothingPeriod: 10, SpectralDilationCompensation: true, EnhanceResolution: false},
			expectedHeatmap: expectedP020_P120_HP48_LP10_SDC1_E0_W48_HEATMAP,
			expectedMin:     expectedP020_P120_HP48_LP10_SDC1_E0_W48_MIN,
			expectedMax:     expectedP020_P120_HP48_LP10_SDC1_E0_W48_MAX,
			expectedDC:      expectedP020_P120_HP48_LP10_SDC1_E0_W48_DOMINANT_CYCLE,
		},
		{
			name:            "P024_P160_HP60_LP12_SDC1_E0_W60",
			params:          Params{MinPeriod: 24, MaxPeriod: 60, WindowLength: 60, HighpassPeriod: 60, SmoothingPeriod: 12, SpectralDilationCompensation: true, EnhanceResolution: false},
			expectedHeatmap: expectedP024_P160_HP60_LP12_SDC1_E0_W60_HEATMAP,
			expectedMin:     expectedP024_P160_HP60_LP12_SDC1_E0_W60_MIN,
			expectedMax:     expectedP024_P160_HP60_LP12_SDC1_E0_W60_MAX,
			expectedDC:      expectedP024_P160_HP60_LP12_SDC1_E0_W60_DOMINANT_CYCLE,
		},
	}
}

func assertClose(t *testing.T, actual, expected, tolerance float64, label string) {
	t.Helper()
	if math.IsNaN(expected) {
		if !math.IsNaN(actual) {
			t.Fatalf("%s: expected NaN, got %.17g", label, actual)
		}
		return
	}
	if math.IsNaN(actual) {
		t.Fatalf("%s: expected %.17g, got NaN", label, expected)
	}
	bound := math.Max(tolerance, math.Abs(expected)*tolerance)
	if math.Abs(actual-expected) > bound {
		t.Fatalf("%s: expected %.17g, got %.17g (diff=%.17g, bound=%.17g)", label, expected, actual, math.Abs(actual-expected), bound)
	}
}

func assertSameState(t *testing.T, actual, expected SpectrumState, label string) {
	t.Helper()
	assertClose(t, actual.HighPass, expected.HighPass, referenceTolerance, label+".hp")
	assertClose(t, actual.Filtered, expected.Filtered, referenceTolerance, label+".filt")
	assertClose(t, actual.NormalizationMaximum, expected.NormalizationMaximum, referenceTolerance, label+".max")
	assertClose(t, actual.InternalDominantCycle, expected.InternalDominantCycle, referenceTolerance, label+".idc")
	if len(actual.RawPower) != len(expected.RawPower) {
		t.Fatalf("%s.raw len: got %d, want %d", label, len(actual.RawPower), len(expected.RawPower))
	}
	for index := range actual.RawPower {
		assertClose(t, actual.RawPower[index], expected.RawPower[index], referenceTolerance, label+".raw")
	}
}

func assertSameOutput(t *testing.T, actual, expected SpectrumOutput, label string) {
	t.Helper()
	if !actual.Heatmap.Time.Equal(expected.Heatmap.Time) {
		t.Fatalf("%s.time: got %v, want %v", label, actual.Heatmap.Time, expected.Heatmap.Time)
	}
	if actual.Heatmap.ParameterFirst != expected.Heatmap.ParameterFirst ||
		actual.Heatmap.ParameterLast != expected.Heatmap.ParameterLast ||
		actual.Heatmap.ParameterResolution != expected.Heatmap.ParameterResolution {
		t.Fatalf("%s.axis mismatch", label)
	}
	assertClose(t, actual.Heatmap.ValueMin, expected.Heatmap.ValueMin, referenceTolerance, label+".min")
	assertClose(t, actual.Heatmap.ValueMax, expected.Heatmap.ValueMax, referenceTolerance, label+".max")
	if len(actual.Heatmap.Values) != len(expected.Heatmap.Values) {
		t.Fatalf("%s.values len: got %d, want %d", label, len(actual.Heatmap.Values), len(expected.Heatmap.Values))
	}
	for index := range actual.Heatmap.Values {
		assertClose(t, actual.Heatmap.Values[index], expected.Heatmap.Values[index], referenceTolerance, label+".val")
	}
	assertClose(t, actual.DominantCycle, expected.DominantCycle, referenceTolerance, label+".dc")
}

type listingOracle struct {
	params     Params
	prices     []float64
	hp         []float64
	filt       []float64
	maxPower   float64
	dominant   float64
	identified bool
}

func (o *listingOracle) previous(series []float64, lag int) float64 {
	if len(series) > lag {
		return series[len(series)-1-lag]
	}
	return 0
}

func (o *listingOracle) update(price float64) ([]float64, float64, float64, float64) {
	params := o.params
	theta := math.Sqrt2 * math.Pi / params.HighpassPeriod
	alpha := (math.Cos(theta) + math.Sin(theta) - 1.0) / math.Cos(theta)
	high := (1.0-alpha/2.0)*(1.0-alpha/2.0)*(price-2.0*o.previous(o.prices, 0)+o.previous(o.prices, 1)) +
		2.0*(1.0-alpha)*o.previous(o.hp, 0) -
		(1.0-alpha)*(1.0-alpha)*o.previous(o.hp, 1)
	a1 := math.Exp(-math.Sqrt2 * math.Pi / params.SmoothingPeriod)
	c2 := 2.0 * a1 * math.Cos(math.Sqrt2*math.Pi/params.SmoothingPeriod)
	c3 := -a1 * a1
	c1 := 1.0 - c2 - c3
	filtered := c1*(high+o.previous(o.hp, 0))/2.0 + c2*o.previous(o.filt, 0) + c3*o.previous(o.filt, 1)
	o.prices = append(o.prices, price)
	o.hp = append(o.hp, high)
	o.filt = append(o.filt, filtered)

	cells := params.MaxPeriod - params.MinPeriod + 1
	powers := make([]float64, cells)
	for index := range cells {
		period := params.MinPeriod + index
		compensation := 1.0
		if params.SpectralDilationCompensation {
			compensation = float64(period)
		}
		cosine := 0.0
		sine := 0.0
		for n := range params.WindowLength {
			value := o.previous(o.filt, n)
			angle := 2.0 * math.Pi * float64(n) / float64(period)
			cosine += value * math.Cos(angle) / compensation
			sine += value * math.Sin(angle) / compensation
		}
		powers[index] = cosine*cosine + sine*sine
	}

	o.maxPower *= 0.995
	for _, power := range powers {
		if power > o.maxPower {
			o.maxPower = power
		}
	}
	for index := range powers {
		if o.maxPower != 0.0 {
			powers[index] /= o.maxPower
		} else {
			powers[index] = 0.0
		}
	}

	sp := 0.0
	spx := 0.0
	for index, power := range powers {
		period := params.MinPeriod + index
		if power >= 0.5 {
			spx += float64(period) * power
			sp += power
		}
	}
	if sp != 0.0 {
		o.dominant = spx / sp
		o.identified = true
	}
	values := make([]float64, cells)
	for index, power := range powers {
		if params.EnhanceResolution {
			values[index] = power * power * power
		} else {
			values[index] = power
		}
	}
	dc := math.NaN()
	if o.identified {
		dc = o.dominant
	}
	return values, dc, high, filtered
}

func TestDiscreteFourierTransformSpectrumReferenceCases(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, rc := range allReferenceCases() {
		t.Run(rc.name, func(t *testing.T) {
			t.Parallel()

			paramsCopy := rc.params
			indicator, err := NewDiscreteFourierTransformSpectrumParams(&paramsCopy)
			if err != nil {
				t.Fatalf("unexpected constructor error: %v", err)
			}
			oracle := &listingOracle{params: indicator.Parameters()}
			cells := rc.params.MaxPeriod - rc.params.MinPeriod + 1

			for barIndex, price := range testInput {
				barTime := baseTime.AddDate(0, 0, barIndex)
				result := indicator.Update(price, barTime)
				oracleValues, oracleDC, oracleHP, oracleFilt := oracle.update(price)

				state := indicator.State()
				assertClose(t, state.HighPass, oracleHP, oracleTolerance, rc.name+".hp")
				assertClose(t, state.Filtered, oracleFilt, oracleTolerance, rc.name+".filt")

				if !result.Heatmap.Time.Equal(barTime) {
					t.Fatalf("[%d] heatmap time = %v, want %v", barIndex, result.Heatmap.Time, barTime)
				}
				if result.Heatmap.ParameterFirst != float64(rc.params.MinPeriod) ||
					result.Heatmap.ParameterLast != float64(rc.params.MaxPeriod) ||
					result.Heatmap.ParameterResolution != 1.0 {
					t.Fatalf("[%d] unexpected heatmap axis", barIndex)
				}

				if barIndex < indicator.WarmupBars() {
					if indicator.IsPrimed() {
						t.Fatalf("[%d] expected not primed", barIndex)
					}
					if !result.Heatmap.IsEmpty() {
						t.Fatalf("[%d] expected empty heatmap before priming", barIndex)
					}
				} else {
					if !indicator.IsPrimed() {
						t.Fatalf("[%d] expected primed", barIndex)
					}
					if len(result.Heatmap.Values) != cells {
						t.Fatalf("[%d] len(values) = %d, want %d", barIndex, len(result.Heatmap.Values), cells)
					}
					for binIndex := range cells {
						assertClose(t, result.Heatmap.Values[binIndex], rc.expectedHeatmap[barIndex][binIndex], referenceTolerance, rc.name+".heatmap")
						assertClose(t, result.Heatmap.Values[binIndex], oracleValues[binIndex], oracleTolerance, rc.name+".oracle")
					}
					assertClose(t, result.DominantCycle, oracleDC, oracleTolerance, rc.name+".oracle_dc")
				}

				assertClose(t, result.Heatmap.ValueMin, rc.expectedMin[barIndex], referenceTolerance, rc.name+".min")
				assertClose(t, result.Heatmap.ValueMax, rc.expectedMax[barIndex], referenceTolerance, rc.name+".max")
				assertClose(t, result.DominantCycle, rc.expectedDC[barIndex], referenceTolerance, rc.name+".dc")
			}

			indicator.Reset()
			if indicator.IsPrimed() || indicator.SamplesSeen() != 0 {
				t.Fatalf("reset failed to clear priming state")
			}
		})
	}
}

func TestDiscreteFourierTransformSpectrumMissingInfinityAndOverflow(t *testing.T) {
	t.Parallel()

	indicator, err := NewDiscreteFourierTransformSpectrumDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	control, err := NewDiscreteFourierTransformSpectrumDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for index, price := range testInput[:120] {
		sampleTime := baseTime.AddDate(0, 0, index)
		indicator.Update(price, sampleTime)
		control.Update(price, sampleTime)
	}

	savedState := indicator.State()
	savedOutput := indicator.Output()
	savedCount := indicator.SamplesSeen()
	missingTime := time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC)

	missingResult := indicator.Update(math.NaN(), missingTime)
	if !missingResult.Heatmap.IsEmpty() || !math.IsNaN(missingResult.DominantCycle) {
		t.Fatalf("expected empty heatmap and NaN dominant cycle on NaN input")
	}
	if !missingResult.Heatmap.Time.Equal(missingTime) {
		t.Fatalf("expected missing time %v, got %v", missingTime, missingResult.Heatmap.Time)
	}
	assertSameState(t, indicator.State(), savedState, "after_nan")
	assertSameOutput(t, indicator.Output(), savedOutput, "after_nan_output")
	if indicator.SamplesSeen() != savedCount {
		t.Fatalf("expected samplesSeen %d, got %d", savedCount, indicator.SamplesSeen())
	}

	for _, badValue := range []float64{math.Inf(1), math.Inf(-1), 1e308} {
		if _, err := indicator.TryUpdate(badValue, missingTime); err == nil {
			t.Fatalf("expected error for sample %v", badValue)
		}
		assertSameState(t, indicator.State(), savedState, "after_bad")
		assertSameOutput(t, indicator.Output(), savedOutput, "after_bad_output")
		if indicator.SamplesSeen() != savedCount {
			t.Fatalf("samplesSeen mutated on error")
		}
	}

	nextTime := baseTime.AddDate(0, 0, 120)
	assertSameOutput(t, indicator.Update(testInput[120], nextTime), control.Update(testInput[120], nextTime), "resumed")
}

func TestDiscreteFourierTransformSpectrumOwnedSnapshotResetZeroAndProperties(t *testing.T) {
	t.Parallel()

	indicator, err := NewDiscreteFourierTransformSpectrumDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	var lastResult SpectrumOutput
	for index, price := range testInput {
		lastResult = indicator.Update(price, baseTime.AddDate(0, 0, index))
	}
	snapshot := slices.Clone(lastResult.Heatmap.Values)
	indicator.Update(101.0, baseTime.AddDate(0, 0, 252))
	if !slices.Equal(lastResult.Heatmap.Values, snapshot) {
		t.Fatalf("returned heatmap slice mutated on subsequent update")
	}

	lastResult.Heatmap.Values[0] = -999.0
	control, _ := NewDiscreteFourierTransformSpectrumDefault()
	for index, price := range testInput {
		control.Update(price, baseTime.AddDate(0, 0, index))
	}
	control.Update(101.0, baseTime.AddDate(0, 0, 252))
	assertSameOutput(
		t,
		indicator.Update(102.0, baseTime.AddDate(0, 0, 253)),
		control.Update(102.0, baseTime.AddDate(0, 0, 253)),
		"mutation_isolation",
	)

	indicator.Reset()
	reference, _ := NewDiscreteFourierTransformSpectrumDefault()
	for index, price := range testInput {
		sampleTime := baseTime.AddDate(0, 0, index)
		assertSameOutput(t, indicator.Update(price, sampleTime), reference.Update(price, sampleTime), "replay")
	}

	zeroIndicator, _ := NewDiscreteFourierTransformSpectrumDefault()
	var zeroOut SpectrumOutput
	for step := range zeroIndicator.WarmupBars() + 4 {
		zeroOut = zeroIndicator.Update(0.0, baseTime.AddDate(0, 0, step))
	}
	if !zeroIndicator.IsPrimed() {
		t.Fatalf("expected primed after warmup+4 zero samples")
	}
	for _, value := range zeroOut.Heatmap.Values {
		if value != 0.0 {
			t.Fatalf("expected 0.0 in zero-input heatmap, got %v", value)
		}
	}
	if !math.IsNaN(zeroOut.DominantCycle) || zeroIndicator.State().NormalizationMaximum != 0.0 {
		t.Fatalf("expected NaN dominant cycle and zero normalization maximum")
	}

	plainParams := DefaultParams()
	plainParams.SpectralDilationCompensation = false
	plain, _ := NewDiscreteFourierTransformSpectrumParams(plainParams)
	enhParams := DefaultParams()
	enhParams.EnhanceResolution = true
	enhanced, _ := NewDiscreteFourierTransformSpectrumParams(enhParams)
	compensated, _ := NewDiscreteFourierTransformSpectrumDefault()
	for index, price := range testInput {
		sampleTime := baseTime.AddDate(0, 0, index)
		compOut := compensated.Update(price, sampleTime)
		plain.Update(price, sampleTime)
		enhOut := enhanced.Update(price, sampleTime)
		assertSameState(t, compensated.State(), enhanced.State(), "enh_state")
		assertClose(t, compOut.DominantCycle, enhOut.DominantCycle, referenceTolerance, "enh_dc")
		if !compOut.Heatmap.IsEmpty() {
			for binIndex, baseVal := range compOut.Heatmap.Values {
				assertClose(t, enhOut.Heatmap.Values[binIndex], baseVal*baseVal*baseVal, referenceTolerance, "enh_cube")
			}
		}
	}
	compState := compensated.State()
	plainState := plain.State()
	for binIndex := range 39 {
		period := float64(10 + binIndex)
		assertClose(t, compState.RawPower[binIndex], plainState.RawPower[binIndex]/(period*period), oracleTolerance, "sdc_ratio")
	}

	sineParams := DefaultParams()
	sineParams.WindowLength = 120
	sineIndicator, _ := NewDiscreteFourierTransformSpectrumParams(sineParams)
	var sineOut SpectrumOutput
	for step := range 1000 {
		sineOut = sineIndicator.Update(math.Sin(2.0*math.Pi*float64(step)/20.0), baseTime.AddDate(0, 0, step))
	}
	peakIndex := 0
	for index, value := range sineOut.Heatmap.Values {
		if value > sineOut.Heatmap.Values[peakIndex] {
			peakIndex = index
		}
	}
	peakPeriod := 10 + peakIndex
	if peakPeriod < 19 || peakPeriod > 21 {
		t.Fatalf("unexpected sine peak period %d", peakPeriod)
	}
	if !(sineOut.DominantCycle > 17.0 && sineOut.DominantCycle < 23.0) {
		t.Fatalf("unexpected sine dominant cycle %v", sineOut.DominantCycle)
	}
}

func TestDiscreteFourierTransformSpectrumMetadataMnemonicsEntitiesAndValidation(t *testing.T) {
	t.Parallel()

	indicator, err := NewDiscreteFourierTransformSpectrumDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	meta := indicator.Metadata()
	if meta.Identifier != core.DiscreteFourierTransformSpectrum {
		t.Fatalf("Identifier = %v, want %v", meta.Identifier, core.DiscreteFourierTransformSpectrum)
	}
	if meta.Mnemonic != "dftps(10, 48, 48, 48, 10)" {
		t.Fatalf("Mnemonic = %q", meta.Mnemonic)
	}
	if len(meta.Outputs) != 2 {
		t.Fatalf("len(Outputs) = %d, want 2", len(meta.Outputs))
	}
	if meta.Outputs[0].Shape != shape.Heatmap || meta.Outputs[0].Mnemonic != "dftps(10, 48, 48, 48, 10)" {
		t.Fatalf("unexpected primary output metadata: %+v", meta.Outputs[0])
	}
	if meta.Outputs[1].Shape != shape.Scalar || meta.Outputs[1].Mnemonic != "dftps-dc(10, 48, 48, 48, 10)" {
		t.Fatalf("unexpected secondary output metadata: %+v", meta.Outputs[1])
	}

	combinations := []struct {
		barComp   entities.BarComponent
		quoteComp entities.QuoteComponent
		tradeComp entities.TradeComponent
	}{
		{0, 0, 0},
		{entities.BarOpenPrice, 0, 0},
		{0, entities.QuoteBidPrice, 0},
		{0, 0, entities.TradeVolume},
		{entities.BarMedianPrice, entities.QuoteAskPrice, 0},
		{entities.BarMedianPrice, 0, entities.TradeVolume},
		{0, entities.QuoteAskPrice, entities.TradeVolume},
		{entities.BarOpenPrice, entities.QuoteBidPrice, entities.TradeVolume},
	}
	for _, combo := range combinations {
		params := DefaultParams()
		params.BarComponent = combo.barComp
		params.QuoteComponent = combo.quoteComp
		params.TradeComponent = combo.tradeComp
		inst, err := NewDiscreteFourierTransformSpectrumParams(params)
		if err != nil {
			t.Fatalf("unexpected error for combo %+v: %v", combo, err)
		}
		resolvedBar := combo.barComp
		if resolvedBar == 0 {
			resolvedBar = entities.DefaultBarComponent
		}
		resolvedQuote := combo.quoteComp
		if resolvedQuote == 0 {
			resolvedQuote = entities.DefaultQuoteComponent
		}
		resolvedTrade := combo.tradeComp
		if resolvedTrade == 0 {
			resolvedTrade = entities.DefaultTradeComponent
		}
		suffix := core.ComponentTripleMnemonic(resolvedBar, resolvedQuote, resolvedTrade)
		wantPrimary := "dftps(10, 48, 48, 48, 10" + suffix + ")"
		wantDC := "dftps-dc(10, 48, 48, 48, 10" + suffix + ")"
		if inst.Metadata().Mnemonic != wantPrimary || params.Mnemonic() != wantPrimary {
			t.Fatalf("got primary mnemonic %q, want %q", inst.Metadata().Mnemonic, wantPrimary)
		}
		if inst.Metadata().Outputs[1].Mnemonic != wantDC || params.DominantCycleMnemonic() != wantDC {
			t.Fatalf("got dc mnemonic %q, want %q", inst.Metadata().Outputs[1].Mnemonic, wantDC)
		}
	}

	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	barInd, _ := NewDiscreteFourierTransformSpectrumDefault()
	scalarInd, _ := NewDiscreteFourierTransformSpectrumDefault()
	quoteInd, _ := NewDiscreteFourierTransformSpectrumDefault()
	tradeInd, _ := NewDiscreteFourierTransformSpectrumDefault()
	for index := range 60 {
		sampleTime := baseTime.AddDate(0, 0, index)
		price := testInput[index]
		outBar := barInd.UpdateBar(&entities.Bar{
			Time:   sampleTime,
			Open:   testOpen[index],
			High:   testHigh[index],
			Low:    testLow[index],
			Close:  price,
			Volume: 1000,
		})
		outScalar := scalarInd.UpdateScalar(&entities.Scalar{Time: sampleTime, Value: price})
		outQuote := quoteInd.UpdateQuote(&entities.Quote{Time: sampleTime, Bid: price - 0.5, Ask: price + 0.5, BidSize: 10, AskSize: 10})
		outTrade := tradeInd.UpdateTrade(&entities.Trade{Time: sampleTime, Price: price, Volume: 10})
		if len(outBar) != 2 {
			t.Fatalf("len(outBar) = %d, want 2", len(outBar))
		}
		dcBar := outBar[1].(entities.Scalar).Value
		dcScalar := outScalar[1].(entities.Scalar).Value
		dcQuote := outQuote[1].(entities.Scalar).Value
		dcTrade := outTrade[1].(entities.Scalar).Value
		assertClose(t, dcBar, dcScalar, referenceTolerance, "bar_vs_scalar")
		assertClose(t, dcBar, dcQuote, referenceTolerance, "bar_vs_quote")
		assertClose(t, dcBar, dcTrade, referenceTolerance, "bar_vs_trade")
	}

	invalidParams := []Params{
		{MinPeriod: 2, MaxPeriod: 48, WindowLength: 48, HighpassPeriod: 48, SmoothingPeriod: 10, SpectralDilationCompensation: true},
		{MinPeriod: 0, MaxPeriod: 48, WindowLength: 48, HighpassPeriod: 48, SmoothingPeriod: 10, SpectralDilationCompensation: true},
		{MinPeriod: 10, MaxPeriod: 9, WindowLength: 48, HighpassPeriod: 48, SmoothingPeriod: 10, SpectralDilationCompensation: true},
		{MinPeriod: 10, MaxPeriod: 48, WindowLength: 1, HighpassPeriod: 48, SmoothingPeriod: 10, SpectralDilationCompensation: true},
		{MinPeriod: 10, MaxPeriod: 48, WindowLength: 48, HighpassPeriod: 4, SmoothingPeriod: 10, SpectralDilationCompensation: true},
		{MinPeriod: 10, MaxPeriod: 48, WindowLength: 48, HighpassPeriod: math.NaN(), SmoothingPeriod: 10, SpectralDilationCompensation: true},
		{MinPeriod: 10, MaxPeriod: 48, WindowLength: 48, HighpassPeriod: math.Inf(1), SmoothingPeriod: 10, SpectralDilationCompensation: true},
		{MinPeriod: 10, MaxPeriod: 48, WindowLength: 48, HighpassPeriod: 48, SmoothingPeriod: 1, SpectralDilationCompensation: true},
		{MinPeriod: 10, MaxPeriod: 48, WindowLength: 48, HighpassPeriod: 48, SmoothingPeriod: math.NaN(), SpectralDilationCompensation: true},
		{MinPeriod: 10, MaxPeriod: 48, WindowLength: 48, HighpassPeriod: 48, SmoothingPeriod: 10, BarComponent: entities.BarComponent(999)},
	}
	for index, bad := range invalidParams {
		badCopy := bad
		if _, err := NewDiscreteFourierTransformSpectrumParams(&badCopy); err == nil {
			t.Fatalf("expected error for invalidParams[%d]: %+v", index, bad)
		}
	}
}
