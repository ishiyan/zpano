//nolint:testpackage
package combbandpassspectrum

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
	fixtureTolerance = 1e-11
	oracleTolerance  = 2e-11
	exactTolerance   = 1e-13
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
			name:            "P010_P148_HP48_LP10_SDC1_E0_B0_3",
			params:          Params{MinPeriod: 10, MaxPeriod: 48, Bandwidth: 0.3, HighpassPeriod: 48, SmoothingPeriod: 10, SpectralDilationCompensation: true, EnhanceResolution: false},
			expectedHeatmap: expectedP010_P148_HP48_LP10_SDC1_E0_B0_3_HEATMAP,
			expectedMin:     expectedP010_P148_HP48_LP10_SDC1_E0_B0_3_MIN,
			expectedMax:     expectedP010_P148_HP48_LP10_SDC1_E0_B0_3_MAX,
			expectedDC:      expectedP010_P148_HP48_LP10_SDC1_E0_B0_3_DOMINANT_CYCLE,
		},
		{
			name:            "P010_P148_HP48_LP10_SDC0_E0_B0_3",
			params:          Params{MinPeriod: 10, MaxPeriod: 48, Bandwidth: 0.3, HighpassPeriod: 48, SmoothingPeriod: 10, SpectralDilationCompensation: false, EnhanceResolution: false},
			expectedHeatmap: expectedP010_P148_HP48_LP10_SDC0_E0_B0_3_HEATMAP,
			expectedMin:     expectedP010_P148_HP48_LP10_SDC0_E0_B0_3_MIN,
			expectedMax:     expectedP010_P148_HP48_LP10_SDC0_E0_B0_3_MAX,
			expectedDC:      expectedP010_P148_HP48_LP10_SDC0_E0_B0_3_DOMINANT_CYCLE,
		},
		{
			name:            "P010_P148_HP48_LP10_SDC1_E1_B0_3",
			params:          Params{MinPeriod: 10, MaxPeriod: 48, Bandwidth: 0.3, HighpassPeriod: 48, SmoothingPeriod: 10, SpectralDilationCompensation: true, EnhanceResolution: true},
			expectedHeatmap: expectedP010_P148_HP48_LP10_SDC1_E1_B0_3_HEATMAP,
			expectedMin:     expectedP010_P148_HP48_LP10_SDC1_E1_B0_3_MIN,
			expectedMax:     expectedP010_P148_HP48_LP10_SDC1_E1_B0_3_MAX,
			expectedDC:      expectedP010_P148_HP48_LP10_SDC1_E1_B0_3_DOMINANT_CYCLE,
		},
		{
			name:            "P010_P148_HP48_LP10_SDC0_E1_B0_3",
			params:          Params{MinPeriod: 10, MaxPeriod: 48, Bandwidth: 0.3, HighpassPeriod: 48, SmoothingPeriod: 10, SpectralDilationCompensation: false, EnhanceResolution: true},
			expectedHeatmap: expectedP010_P148_HP48_LP10_SDC0_E1_B0_3_HEATMAP,
			expectedMin:     expectedP010_P148_HP48_LP10_SDC0_E1_B0_3_MIN,
			expectedMax:     expectedP010_P148_HP48_LP10_SDC0_E1_B0_3_MAX,
			expectedDC:      expectedP010_P148_HP48_LP10_SDC0_E1_B0_3_DOMINANT_CYCLE,
		},
		{
			name:            "P010_P148_HP48_LP10_SDC1_E0_B0_1",
			params:          Params{MinPeriod: 10, MaxPeriod: 48, Bandwidth: 0.1, HighpassPeriod: 48, SmoothingPeriod: 10, SpectralDilationCompensation: true, EnhanceResolution: false},
			expectedHeatmap: expectedP010_P148_HP48_LP10_SDC1_E0_B0_1_HEATMAP,
			expectedMin:     expectedP010_P148_HP48_LP10_SDC1_E0_B0_1_MIN,
			expectedMax:     expectedP010_P148_HP48_LP10_SDC1_E0_B0_1_MAX,
			expectedDC:      expectedP010_P148_HP48_LP10_SDC1_E0_B0_1_DOMINANT_CYCLE,
		},
		{
			name:            "P010_P148_HP48_LP10_SDC1_E0_B0_7",
			params:          Params{MinPeriod: 10, MaxPeriod: 48, Bandwidth: 0.7, HighpassPeriod: 48, SmoothingPeriod: 10, SpectralDilationCompensation: true, EnhanceResolution: false},
			expectedHeatmap: expectedP010_P148_HP48_LP10_SDC1_E0_B0_7_HEATMAP,
			expectedMin:     expectedP010_P148_HP48_LP10_SDC1_E0_B0_7_MIN,
			expectedMax:     expectedP010_P148_HP48_LP10_SDC1_E0_B0_7_MAX,
			expectedDC:      expectedP010_P148_HP48_LP10_SDC1_E0_B0_7_DOMINANT_CYCLE,
		},
		{
			name:            "P010_P148_HP48_LP10_5_SDC1_E0_B0_3",
			params:          Params{MinPeriod: 10, MaxPeriod: 48, Bandwidth: 0.3, HighpassPeriod: 48, SmoothingPeriod: 10.5, SpectralDilationCompensation: true, EnhanceResolution: false},
			expectedHeatmap: expectedP010_P148_HP48_LP10_5_SDC1_E0_B0_3_HEATMAP,
			expectedMin:     expectedP010_P148_HP48_LP10_5_SDC1_E0_B0_3_MIN,
			expectedMax:     expectedP010_P148_HP48_LP10_5_SDC1_E0_B0_3_MAX,
			expectedDC:      expectedP010_P148_HP48_LP10_5_SDC1_E0_B0_3_DOMINANT_CYCLE,
		},
		{
			name:            "P010_P148_HP80_LP20_SDC1_E0_B0_3",
			params:          Params{MinPeriod: 10, MaxPeriod: 48, Bandwidth: 0.3, HighpassPeriod: 80, SmoothingPeriod: 20, SpectralDilationCompensation: true, EnhanceResolution: false},
			expectedHeatmap: expectedP010_P148_HP80_LP20_SDC1_E0_B0_3_HEATMAP,
			expectedMin:     expectedP010_P148_HP80_LP20_SDC1_E0_B0_3_MIN,
			expectedMax:     expectedP010_P148_HP80_LP20_SDC1_E0_B0_3_MAX,
			expectedDC:      expectedP010_P148_HP80_LP20_SDC1_E0_B0_3_DOMINANT_CYCLE,
		},
		{
			name:            "P06_P124_HP24_LP6_SDC1_E0_B0_3",
			params:          Params{MinPeriod: 6, MaxPeriod: 24, Bandwidth: 0.3, HighpassPeriod: 24, SmoothingPeriod: 6, SpectralDilationCompensation: true, EnhanceResolution: false},
			expectedHeatmap: expectedP06_P124_HP24_LP6_SDC1_E0_B0_3_HEATMAP,
			expectedMin:     expectedP06_P124_HP24_LP6_SDC1_E0_B0_3_MIN,
			expectedMax:     expectedP06_P124_HP24_LP6_SDC1_E0_B0_3_MAX,
			expectedDC:      expectedP06_P124_HP24_LP6_SDC1_E0_B0_3_DOMINANT_CYCLE,
		},
		{
			name:            "P020_P120_HP48_LP10_SDC1_E0_B0_3",
			params:          Params{MinPeriod: 20, MaxPeriod: 20, Bandwidth: 0.3, HighpassPeriod: 48, SmoothingPeriod: 10, SpectralDilationCompensation: true, EnhanceResolution: false},
			expectedHeatmap: expectedP020_P120_HP48_LP10_SDC1_E0_B0_3_HEATMAP,
			expectedMin:     expectedP020_P120_HP48_LP10_SDC1_E0_B0_3_MIN,
			expectedMax:     expectedP020_P120_HP48_LP10_SDC1_E0_B0_3_MAX,
			expectedDC:      expectedP020_P120_HP48_LP10_SDC1_E0_B0_3_DOMINANT_CYCLE,
		},
		{
			name:            "P024_P160_HP60_LP12_SDC1_E0_B0_3",
			params:          Params{MinPeriod: 24, MaxPeriod: 60, Bandwidth: 0.3, HighpassPeriod: 60, SmoothingPeriod: 12, SpectralDilationCompensation: true, EnhanceResolution: false},
			expectedHeatmap: expectedP024_P160_HP60_LP12_SDC1_E0_B0_3_HEATMAP,
			expectedMin:     expectedP024_P160_HP60_LP12_SDC1_E0_B0_3_MIN,
			expectedMax:     expectedP024_P160_HP60_LP12_SDC1_E0_B0_3_MAX,
			expectedDC:      expectedP024_P160_HP60_LP12_SDC1_E0_B0_3_DOMINANT_CYCLE,
		},
		{
			name:            "P010_P1240_HP240_LP10_SDC1_E0_B0_3",
			params:          Params{MinPeriod: 10, MaxPeriod: 240, Bandwidth: 0.3, HighpassPeriod: 240, SmoothingPeriod: 10, SpectralDilationCompensation: true, EnhanceResolution: false},
			expectedHeatmap: expectedP010_P1240_HP240_LP10_SDC1_E0_B0_3_HEATMAP,
			expectedMin:     expectedP010_P1240_HP240_LP10_SDC1_E0_B0_3_MIN,
			expectedMax:     expectedP010_P1240_HP240_LP10_SDC1_E0_B0_3_MAX,
			expectedDC:      expectedP010_P1240_HP240_LP10_SDC1_E0_B0_3_DOMINANT_CYCLE,
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
	assertClose(t, actual.HighPass, expected.HighPass, exactTolerance, label+".hp")
	assertClose(t, actual.Filtered, expected.Filtered, exactTolerance, label+".filt")
	assertClose(t, actual.NormalizationMaximum, expected.NormalizationMaximum, exactTolerance, label+".max")
	assertClose(t, actual.InternalDominantCycle, expected.InternalDominantCycle, exactTolerance, label+".idc")
	if len(actual.RawPower) != len(expected.RawPower) {
		t.Fatalf("%s.raw len: got %d, want %d", label, len(actual.RawPower), len(expected.RawPower))
	}
	for index := range actual.RawPower {
		assertClose(t, actual.RawPower[index], expected.RawPower[index], exactTolerance, label+".raw")
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
	assertClose(t, actual.Heatmap.ValueMin, expected.Heatmap.ValueMin, exactTolerance, label+".min")
	assertClose(t, actual.Heatmap.ValueMax, expected.Heatmap.ValueMax, exactTolerance, label+".max")
	if len(actual.Heatmap.Values) != len(expected.Heatmap.Values) {
		t.Fatalf("%s.values len: got %d, want %d", label, len(actual.Heatmap.Values), len(expected.Heatmap.Values))
	}
	for index := range actual.Heatmap.Values {
		assertClose(t, actual.Heatmap.Values[index], expected.Heatmap.Values[index], exactTolerance, label+".values")
	}
	assertClose(t, actual.DominantCycle, expected.DominantCycle, exactTolerance, label+".dc")
}

type listingOracle struct {
	params     Params
	prices     []float64
	hp         []float64
	filt       []float64
	bp         map[int][]float64
	maxPower   float64
	dominant   float64
	identified bool
}

func newListingOracle(params Params) *listingOracle {
	bp := make(map[int][]float64, params.MaxPeriod-params.MinPeriod+1)
	for period := params.MinPeriod; period <= params.MaxPeriod; period++ {
		bp[period] = make([]float64, params.MaxPeriod+1)
	}
	return &listingOracle{
		params: params,
		bp:     bp,
	}
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
		row := o.bp[period]
		for m := params.MaxPeriod; m > 1; m-- {
			row[m] = row[m-1]
		}
		row[1] = 0
		if len(o.prices) > 12 {
			beta := math.Cos(2.0 * math.Pi / float64(period))
			gamma := 1.0 / math.Cos(2.0*math.Pi*params.Bandwidth/float64(period))
			pole := gamma - math.Sqrt(gamma*gamma-1.0)
			row[1] = 0.5*(1.0-pole)*(filtered-o.previous(o.filt, 2)) +
				beta*(1.0+pole)*row[2] - pole*row[3]
		}
		power := 0.0
		for m := 1; m <= period; m++ {
			power += (row[m] / compensation) * (row[m] / compensation)
		}
		powers[index] = power
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

func TestCombBandPassSpectrumReferenceCases(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, rc := range allReferenceCases() {
		t.Run(rc.name, func(t *testing.T) {
			t.Parallel()

			paramsCopy := rc.params
			indicator, err := NewCombBandPassSpectrumParams(&paramsCopy)
			if err != nil {
				t.Fatalf("unexpected constructor error: %v", err)
			}
			oracle := newListingOracle(indicator.Parameters())
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
						assertClose(t, result.Heatmap.Values[binIndex], rc.expectedHeatmap[barIndex][binIndex], fixtureTolerance, rc.name+".heatmap")
						assertClose(t, result.Heatmap.Values[binIndex], oracleValues[binIndex], oracleTolerance, rc.name+".oracle")
					}
					assertClose(t, result.DominantCycle, oracleDC, oracleTolerance, rc.name+".oracle_dc")
				}

				assertClose(t, result.Heatmap.ValueMin, rc.expectedMin[barIndex], fixtureTolerance, rc.name+".min")
				assertClose(t, result.Heatmap.ValueMax, rc.expectedMax[barIndex], fixtureTolerance, rc.name+".max")
				assertClose(t, result.DominantCycle, rc.expectedDC[barIndex], fixtureTolerance, rc.name+".dc")
			}

			indicator.Reset()
			if indicator.IsPrimed() || indicator.SamplesSeen() != 0 {
				t.Fatalf("reset failed to clear priming state")
			}
		})
	}
}

func TestCombBandPassSpectrumSourceTwelveBarGate(t *testing.T) {
	t.Parallel()

	indicator, err := NewCombBandPassSpectrumDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for index, price := range testInput[:12] {
		indicator.Update(price, baseTime.AddDate(0, 0, index))
		for _, row := range indicator.BandPassHistory() {
			for _, value := range row {
				if value != 0.0 {
					t.Fatalf("expected zero band-pass state during startup gate at bar %d, got %v", index, value)
				}
			}
		}
	}
	indicator.Update(testInput[12], baseTime.AddDate(0, 0, 12))
	hasNonZero := false
	for _, row := range indicator.BandPassHistory() {
		if row[0] != 0.0 {
			hasNonZero = true
			break
		}
	}
	if !hasNonZero {
		t.Fatalf("expected nonzero band-pass output at bar 13")
	}
}

func TestCombBandPassSpectrumMissingInfinityAndOverflow(t *testing.T) {
	t.Parallel()

	indicator, err := NewCombBandPassSpectrumDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	control, err := NewCombBandPassSpectrumDefault()
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

	missingTime := baseTime.AddDate(0, 0, 999)
	nanOut := indicator.Update(math.NaN(), missingTime)
	if !nanOut.Heatmap.IsEmpty() || !math.IsNaN(nanOut.DominantCycle) {
		t.Fatalf("expected empty heatmap and NaN dominant cycle on NaN sample")
	}
	if !nanOut.Heatmap.Time.Equal(missingTime) {
		t.Fatalf("expected NaN heatmap time %v, got %v", missingTime, nanOut.Heatmap.Time)
	}
	assertSameState(t, indicator.State(), savedState, "after_nan_state")
	assertSameOutput(t, indicator.Output(), savedOutput, "after_nan_output")
	if indicator.SamplesSeen() != savedCount {
		t.Fatalf("NaN sample mutated count: got %d, want %d", indicator.SamplesSeen(), savedCount)
	}

	for _, badSample := range []float64{math.Inf(1), math.Inf(-1), 1e308} {
		if _, tryErr := indicator.TryUpdate(badSample, missingTime); tryErr == nil {
			t.Fatalf("expected error for invalid/overflowing sample %v", badSample)
		}
		assertSameState(t, indicator.State(), savedState, "after_bad_state")
		assertSameOutput(t, indicator.Output(), savedOutput, "after_bad_output")
		if indicator.SamplesSeen() != savedCount {
			t.Fatalf("rejected sample mutated count")
		}
	}

	for offset, price := range testInput[120:150] {
		sampleTime := baseTime.AddDate(0, 0, 120+offset)
		assertSameOutput(
			t,
			indicator.Update(price, sampleTime),
			control.Update(price, sampleTime),
			"resume_after_rejected",
		)
	}
}

func TestCombBandPassSpectrumOwnedSnapshotAndReset(t *testing.T) {
	t.Parallel()

	indicator, _ := NewCombBandPassSpectrumDefault()
	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for index, price := range testInput[:70] {
		indicator.Update(price, baseTime.AddDate(0, 0, index))
	}
	snapshot := indicator.Output()
	originalValues := slices.Clone(snapshot.Heatmap.Values)
	snapshot.Heatmap.Values[0] = -999.0

	bpSnapshot := indicator.BandPassHistory()
	bpSnapshot[0][0] = -777.0
	if indicator.BandPassHistory()[0][0] == -777.0 {
		t.Fatalf("BandPassHistory returned aliased slice")
	}

	indicator.Update(testInput[70], baseTime.AddDate(0, 0, 70))
	if snapshot.Heatmap.Values[1] != originalValues[1] {
		t.Fatalf("subsequent update mutated previously returned heatmap slice")
	}

	indicator.Reset()
	if indicator.IsPrimed() || indicator.SamplesSeen() != 0 {
		t.Fatalf("reset did not clear priming")
	}
	if !indicator.Output().Heatmap.IsEmpty() || !math.IsNaN(indicator.Output().DominantCycle) {
		t.Fatalf("reset did not clear cached output")
	}

	fresh, _ := NewCombBandPassSpectrumDefault()
	for index, price := range testInput[:80] {
		sampleTime := baseTime.AddDate(0, 0, index)
		assertSameOutput(t, indicator.Update(price, sampleTime), fresh.Update(price, sampleTime), "replay")
	}
}

func TestCombBandPassSpectrumZeroSingleBinAndBoundary(t *testing.T) {
	t.Parallel()

	zeroInd, _ := NewCombBandPassSpectrumDefault()
	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	var zeroOut SpectrumOutput
	for step := range 70 {
		zeroOut = zeroInd.Update(0.0, baseTime.AddDate(0, 0, step))
	}
	if !zeroInd.IsPrimed() || zeroOut.Heatmap.ValueMin != 0.0 || zeroOut.Heatmap.ValueMax != 0.0 || !math.IsNaN(zeroOut.DominantCycle) {
		t.Fatalf("unexpected zero-input result: %+v", zeroOut)
	}

	single, err := NewCombBandPassSpectrumParams(&Params{
		MinPeriod:                    20,
		MaxPeriod:                    20,
		Bandwidth:                    0.3,
		HighpassPeriod:               48,
		SmoothingPeriod:              10,
		SpectralDilationCompensation: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if single.WarmupBars() != 31 {
		t.Fatalf("single warmup = %d, want 31", single.WarmupBars())
	}
	var singleOut SpectrumOutput
	for index, price := range testInput {
		singleOut = single.Update(price, baseTime.AddDate(0, 0, index))
		if !math.IsNaN(singleOut.DominantCycle) {
			assertClose(t, singleOut.DominantCycle, 20.0, exactTolerance, "single_dc")
		}
	}
	if len(singleOut.Heatmap.Values) != 1 {
		t.Fatalf("len(single values) = %d, want 1", len(singleOut.Heatmap.Values))
	}

	boundary, err := NewCombBandPassSpectrumParams(&Params{
		MinPeriod:                    10,
		MaxPeriod:                    240,
		Bandwidth:                    0.3,
		HighpassPeriod:               240,
		SmoothingPeriod:              10,
		SpectralDilationCompensation: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if boundary.WarmupBars() != 251 {
		t.Fatalf("boundary warmup = %d, want 251", boundary.WarmupBars())
	}
	for index, price := range testInput {
		out := boundary.Update(price, baseTime.AddDate(0, 0, index))
		if index < 251 && !out.Heatmap.IsEmpty() {
			t.Fatalf("expected empty heatmap before bar 251 at %d", index)
		}
		if index == 251 && out.Heatmap.IsEmpty() {
			t.Fatalf("expected primed heatmap on bar 251")
		}
	}
}

func TestCombBandPassSpectrumCompensationEnhancementAndSine(t *testing.T) {
	t.Parallel()

	compensated, _ := NewCombBandPassSpectrumDefault()
	uncompensatedParams := DefaultParams()
	uncompensatedParams.SpectralDilationCompensation = false
	uncompensated, _ := NewCombBandPassSpectrumParams(uncompensatedParams)
	enhancedParams := DefaultParams()
	enhancedParams.EnhanceResolution = true
	enhanced, _ := NewCombBandPassSpectrumParams(enhancedParams)

	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for barIndex, price := range testInput {
		sampleTime := baseTime.AddDate(0, 0, barIndex)
		outComp := compensated.Update(price, sampleTime)
		uncompensated.Update(price, sampleTime)
		outEnh := enhanced.Update(price, sampleTime)

		assertSameState(t, compensated.State(), enhanced.State(), "comp_vs_enh")
		assertClose(t, outComp.DominantCycle, outEnh.DominantCycle, exactTolerance, "enh_dc")

		compRaw := compensated.State().RawPower
		uncompRaw := uncompensated.State().RawPower
		for binIndex := range compRaw {
			period := float64(10 + binIndex)
			assertClose(t, compRaw[binIndex], uncompRaw[binIndex]/(period*period), exactTolerance, "sdc_ratio")
		}
		if !outComp.Heatmap.IsEmpty() {
			for binIndex, value := range outComp.Heatmap.Values {
				assertClose(t, outEnh.Heatmap.Values[binIndex], value*value*value, exactTolerance, "cube")
			}
		}
	}

	sineInd, err := NewCombBandPassSpectrumDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var sineOut SpectrumOutput
	for step := range 1000 {
		sineOut = sineInd.Update(math.Sin(2.0*math.Pi*float64(step)/20.0), baseTime.AddDate(0, 0, step))
	}
	peakIndex := 0
	for index, value := range sineOut.Heatmap.Values {
		if value > sineOut.Heatmap.Values[peakIndex] {
			peakIndex = index
		}
	}
	peakPeriod := 10 + peakIndex
	if peakPeriod < 18 || peakPeriod > 22 {
		t.Fatalf("peakPeriod = %d, want within [18, 22]", peakPeriod)
	}
	if !(sineOut.DominantCycle > 17.0 && sineOut.DominantCycle < 23.0) {
		t.Fatalf("dominantCycle = %v, want in (17, 23)", sineOut.DominantCycle)
	}
}

func TestCombBandPassSpectrumMetadataMnemonicsEntitiesAndValidation(t *testing.T) {
	t.Parallel()

	indicator, err := NewCombBandPassSpectrumDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	meta := indicator.Metadata()
	if meta.Identifier != core.CombBandPassSpectrum {
		t.Fatalf("identifier = %v, want %v", meta.Identifier, core.CombBandPassSpectrum)
	}
	if meta.Mnemonic != "cbps(10, 48, 0.3, 48, 10)" {
		t.Fatalf("mnemonic = %q, want %q", meta.Mnemonic, "cbps(10, 48, 0.3, 48, 10)")
	}
	if len(meta.Outputs) != 2 {
		t.Fatalf("len(outputs) = %d, want 2", len(meta.Outputs))
	}
	if meta.Outputs[0].Shape != shape.Heatmap || meta.Outputs[0].Mnemonic != "cbps(10, 48, 0.3, 48, 10)" {
		t.Fatalf("unexpected primary output metadata: %+v", meta.Outputs[0])
	}
	if meta.Outputs[1].Shape != shape.Scalar || meta.Outputs[1].Mnemonic != "cbps-dc(10, 48, 0.3, 48, 10)" {
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
		inst, constructorErr := NewCombBandPassSpectrumParams(params)
		if constructorErr != nil {
			t.Fatalf("unexpected error for combo %+v: %v", combo, constructorErr)
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
		wantPrimary := "cbps(10, 48, 0.3, 48, 10" + suffix + ")"
		wantDC := "cbps-dc(10, 48, 0.3, 48, 10" + suffix + ")"
		if inst.Metadata().Mnemonic != wantPrimary || params.Mnemonic() != wantPrimary {
			t.Fatalf("got primary mnemonic %q, want %q", inst.Metadata().Mnemonic, wantPrimary)
		}
		if inst.Metadata().Outputs[1].Mnemonic != wantDC || params.DominantCycleMnemonic() != wantDC {
			t.Fatalf("got dc mnemonic %q, want %q", inst.Metadata().Outputs[1].Mnemonic, wantDC)
		}
	}

	flagParams := DefaultParams()
	flagParams.Bandwidth = 0.4
	flagParams.SmoothingPeriod = 10.5
	flagParams.SpectralDilationCompensation = false
	flagParams.EnhanceResolution = true
	flagParams.BarComponent = entities.BarMedianPrice
	if got := flagParams.Mnemonic(); got != "cbps(10, 48, 0.4, 48, 10.5, no-sdc, enh, hl/2)" {
		t.Fatalf("got flag mnemonic %q", got)
	}
	if got := flagParams.DominantCycleMnemonic(); got != "cbps-dc(10, 48, 0.4, 48, 10.5, no-sdc, hl/2)" {
		t.Fatalf("got flag dc mnemonic %q", got)
	}

	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	barInd, _ := NewCombBandPassSpectrumDefault()
	scalarInd, _ := NewCombBandPassSpectrumDefault()
	quoteInd, _ := NewCombBandPassSpectrumDefault()
	tradeInd, _ := NewCombBandPassSpectrumDefault()
	for index := range 70 {
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
		assertClose(t, dcBar, dcScalar, exactTolerance, "bar_vs_scalar")
		assertClose(t, dcBar, dcQuote, exactTolerance, "bar_vs_quote")
		assertClose(t, dcBar, dcTrade, exactTolerance, "bar_vs_trade")
	}

	invalidParams := []Params{
		{MinPeriod: 2, MaxPeriod: 48, Bandwidth: 0.3, HighpassPeriod: 48, SmoothingPeriod: 10, SpectralDilationCompensation: true},
		{MinPeriod: 0, MaxPeriod: 48, Bandwidth: 0.3, HighpassPeriod: 48, SmoothingPeriod: 10, SpectralDilationCompensation: true},
		{MinPeriod: 10, MaxPeriod: 9, Bandwidth: 0.3, HighpassPeriod: 48, SmoothingPeriod: 10, SpectralDilationCompensation: true},
		{MinPeriod: 10, MaxPeriod: 48, Bandwidth: 0, HighpassPeriod: 48, SmoothingPeriod: 10, SpectralDilationCompensation: true},
		{MinPeriod: 10, MaxPeriod: 48, Bandwidth: -0.1, HighpassPeriod: 48, SmoothingPeriod: 10, SpectralDilationCompensation: true},
		{MinPeriod: 10, MaxPeriod: 48, Bandwidth: 2.5, HighpassPeriod: 48, SmoothingPeriod: 10, SpectralDilationCompensation: true},
		{MinPeriod: 3, MaxPeriod: 10, Bandwidth: 0.75, HighpassPeriod: 48, SmoothingPeriod: 10, SpectralDilationCompensation: true},
		{MinPeriod: 10, MaxPeriod: 48, Bandwidth: math.NaN(), HighpassPeriod: 48, SmoothingPeriod: 10, SpectralDilationCompensation: true},
		{MinPeriod: 10, MaxPeriod: 48, Bandwidth: math.Inf(1), HighpassPeriod: 48, SmoothingPeriod: 10, SpectralDilationCompensation: true},
		{MinPeriod: 10, MaxPeriod: 48, Bandwidth: 0.3, HighpassPeriod: 4, SmoothingPeriod: 10, SpectralDilationCompensation: true},
		{MinPeriod: 10, MaxPeriod: 48, Bandwidth: 0.3, HighpassPeriod: math.NaN(), SmoothingPeriod: 10, SpectralDilationCompensation: true},
		{MinPeriod: 10, MaxPeriod: 48, Bandwidth: 0.3, HighpassPeriod: math.Inf(1), SmoothingPeriod: 10, SpectralDilationCompensation: true},
		{MinPeriod: 10, MaxPeriod: 48, Bandwidth: 0.3, HighpassPeriod: 48, SmoothingPeriod: 1, SpectralDilationCompensation: true},
		{MinPeriod: 10, MaxPeriod: 48, Bandwidth: 0.3, HighpassPeriod: 48, SmoothingPeriod: math.NaN(), SpectralDilationCompensation: true},
		{MinPeriod: 10, MaxPeriod: 48, Bandwidth: 0.3, HighpassPeriod: 48, SmoothingPeriod: 10, BarComponent: entities.BarComponent(999)},
	}
	for index, bad := range invalidParams {
		badCopy := bad
		if _, constructorErr := NewCombBandPassSpectrumParams(&badCopy); constructorErr == nil {
			t.Fatalf("expected error for invalidParams[%d]: %+v", index, bad)
		}
	}
}
