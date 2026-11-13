package corona_test

import (
	"math"
	"testing"
	"time"

	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
	"zpano/indicators/factory"
	"zpano/indicators/johnehlers/corona"
	"zpano/indicators/johnehlers/coronasignaltonoiseratio"
	"zpano/indicators/johnehlers/coronaspectrum"
	"zpano/indicators/johnehlers/coronaswingposition"
	"zpano/indicators/johnehlers/coronatrendvigor"
)

const engineTolerance = 5e-12

func sameFloat(actual, expected, tolerance float64) bool {
	if math.IsNaN(expected) {
		return math.IsNaN(actual)
	}
	if math.IsInf(expected, 0) {
		return actual == expected
	}
	if math.IsNaN(actual) || math.IsInf(actual, 0) {
		return false
	}
	bound := max(tolerance, math.Abs(expected)*tolerance)
	return math.Abs(actual-expected) <= bound
}

func sameSlice(actual, expected []float64, tolerance float64) bool {
	if len(actual) != len(expected) {
		return false
	}
	for i := range expected {
		if !sameFloat(actual[i], expected[i], tolerance) {
			return false
		}
	}
	return true
}

func sameHeatmap(actual, expected *outputs.Heatmap, tolerance float64) bool {
	if (actual == nil) != (expected == nil) {
		return false
	}
	if actual == nil {
		return true
	}
	return actual.Time.Equal(expected.Time) &&
		sameFloat(actual.ParameterFirst, expected.ParameterFirst, tolerance) &&
		sameFloat(actual.ParameterLast, expected.ParameterLast, tolerance) &&
		sameFloat(actual.ParameterResolution, expected.ParameterResolution, tolerance) &&
		sameFloat(actual.ValueMin, expected.ValueMin, tolerance) &&
		sameFloat(actual.ValueMax, expected.ValueMax, tolerance) &&
		sameSlice(actual.Values, expected.Values, tolerance)
}

func TestCoronaEngineStartupAndEmptySelectionHolds(t *testing.T) {
	t.Parallel()

	engine, err := corona.NewCorona(corona.DefaultParams())
	if err != nil {
		t.Fatalf("NewCorona: %v", err)
	}
	if engine.FilterBankLength() != 49 {
		t.Fatalf("FilterBankLength = %d, want 49", engine.FilterBankLength())
	}
	if engine.MinimalPeriodTimesTwo() != 12 {
		t.Fatalf("MinimalPeriodTimesTwo = %d, want 12", engine.MinimalPeriodTimesTwo())
	}
	if engine.MaximalPeriodTimesTwo() != 60 {
		t.Fatalf("MaximalPeriodTimesTwo = %d, want 60", engine.MaximalPeriodTimesTwo())
	}

	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	snapshots := make([]corona.CoronaSnapshot, len(testInput))
	for i, price := range testInput {
		snapshots[i] = engine.Update(price, baseTime.AddDate(0, 0, i))
	}
	for i := 0; i < 12; i++ {
		if snapshots[i].Primed || snapshots[i].MaximumPower != 0.0 {
			t.Fatalf("snapshot[%d]: expected unprimed with MaximumPower=0, got Primed=%v MaximumPower=%v",
				i, snapshots[i].Primed, snapshots[i].MaximumPower)
		}
	}
	if !snapshots[12].Primed {
		t.Fatal("snapshot[12]: expected Primed=true")
	}

	holds := 0
	for i := 1; i < len(snapshots); i++ {
		if snapshots[i].SelectedBins == 0 {
			holds++
			if snapshots[i].DominantCycle != snapshots[i-1].DominantCycle {
				t.Fatalf("snapshot[%d]: expected DominantCycle hold %v, got %v",
					i, snapshots[i-1].DominantCycle, snapshots[i].DominantCycle)
			}
			if snapshots[i].HasCycleEstimate {
				t.Fatalf("snapshot[%d]: expected HasCycleEstimate=false when SelectedBins==0", i)
			}
		}
	}
	if holds == 0 {
		t.Fatal("expected at least one zero-bin hold across 252 samples")
	}
	if !sameFloat(snapshots[len(snapshots)-1].DominantCycle, 11.5, engineTolerance) {
		t.Fatalf("final DominantCycle = %v, want 11.5", snapshots[len(snapshots)-1].DominantCycle)
	}
}

func TestSharedAndStandaloneConsumersMatch(t *testing.T) {
	t.Parallel()

	engine, err := corona.NewCorona(nil)
	if err != nil {
		t.Fatalf("NewCorona: %v", err)
	}
	specStandalone, _ := coronaspectrum.NewCoronaSpectrumDefault()
	specShared, _ := coronaspectrum.NewCoronaSpectrumDefault()
	swingStandalone, _ := coronaswingposition.NewCoronaSwingPositionDefault()
	swingShared, _ := coronaswingposition.NewCoronaSwingPositionDefault()
	trendStandalone, _ := coronatrendvigor.NewCoronaTrendVigorDefault()
	trendShared, _ := coronatrendvigor.NewCoronaTrendVigorDefault()
	snrStandalone, _ := coronasignaltonoiseratio.NewCoronaSignalToNoiseRatioDefault()
	snrShared, _ := coronasignaltonoiseratio.NewCoronaSignalToNoiseRatioDefault()

	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	var lastSnapshot corona.CoronaSnapshot
	for i, price := range testInput {
		barTime := baseTime.AddDate(0, 0, i)
		snapshot := engine.Update(price, barTime)
		lastSnapshot = snapshot

		specA := specStandalone.Update(price, barTime)
		specB := specShared.UpdateFromSnapshot(snapshot)
		if !sameHeatmap(specA.Heatmap, specB.Heatmap, engineTolerance) ||
			!sameFloat(specA.DominantCycle, specB.DominantCycle, engineTolerance) ||
			!sameFloat(specA.DominantCycleMedian, specB.DominantCycleMedian, engineTolerance) ||
			!sameSlice(specA.Power, specB.Power, engineTolerance) ||
			!sameSlice(specA.RelativePower, specB.RelativePower, engineTolerance) {
			t.Fatalf("spectrum mismatch at bar %d", i)
		}

		swingA := swingStandalone.Update(price, barTime)
		swingB := swingShared.UpdateFromSnapshot(snapshot)
		if !sameHeatmap(swingA.Heatmap, swingB.Heatmap, engineTolerance) ||
			!sameFloat(swingA.SwingPosition, swingB.SwingPosition, engineTolerance) {
			t.Fatalf("swing mismatch at bar %d", i)
		}

		trendA := trendStandalone.Update(price, barTime)
		trendB := trendShared.UpdateFromSnapshot(snapshot)
		if !sameHeatmap(trendA.Heatmap, trendB.Heatmap, engineTolerance) ||
			!sameFloat(trendA.TrendVigor, trendB.TrendVigor, engineTolerance) {
			t.Fatalf("trend mismatch at bar %d", i)
		}

		snrA := snrStandalone.Update(price, testLow[i], testHigh[i], barTime)
		snrB := snrShared.UpdateFromSnapshot(snapshot, testLow[i], testHigh[i])
		if !sameHeatmap(snrA.Heatmap, snrB.Heatmap, engineTolerance) ||
			!sameFloat(snrA.SignalToNoiseRatio, snrB.SignalToNoiseRatio, engineTolerance) {
			t.Fatalf("snr mismatch at bar %d", i)
		}
	}

	if engine.SampleCount() != 252 {
		t.Fatalf("SampleCount = %d, want 252", engine.SampleCount())
	}
	if _, err := specShared.TryUpdateFromSnapshot(lastSnapshot); err == nil {
		t.Fatal("expected error re-feeding duplicate snapshot")
	}

	incompatibleParams := coronaspectrum.DefaultParams()
	incompatibleParams.HighPassFilterCutoff = 48
	incompatible, err := coronaspectrum.NewCoronaSpectrumParams(incompatibleParams)
	if err != nil {
		t.Fatalf("NewCoronaSpectrumParams: %v", err)
	}
	freshEngine, _ := corona.NewCorona(nil)
	firstSnapshot := freshEngine.Update(testInput[0], baseTime)
	if _, err := incompatible.TryUpdateFromSnapshot(firstSnapshot); err == nil {
		t.Fatal("expected error feeding snapshot from incompatible engine")
	}
}

func TestNaNSkipsAndSnapshotOwnsSlices(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	specA, _ := coronaspectrum.NewCoronaSpectrumDefault()
	specB, _ := coronaspectrum.NewCoronaSpectrumDefault()
	swingA, _ := coronaswingposition.NewCoronaSwingPositionDefault()
	swingB, _ := coronaswingposition.NewCoronaSwingPositionDefault()
	trendA, _ := coronatrendvigor.NewCoronaTrendVigorDefault()
	trendB, _ := coronatrendvigor.NewCoronaTrendVigorDefault()
	snrA, _ := coronasignaltonoiseratio.NewCoronaSignalToNoiseRatioDefault()
	snrB, _ := coronasignaltonoiseratio.NewCoronaSignalToNoiseRatioDefault()

	var specLastA coronaspectrum.Snapshot
	var swingLastA coronaswingposition.Snapshot
	var trendLastA coronatrendvigor.Snapshot
	var snrLastA coronasignaltonoiseratio.Snapshot

	for i, price := range testInput[:50] {
		barTime := baseTime.AddDate(0, 0, i)
		if i == 25 {
			if !specA.Update(math.NaN(), barTime).Heatmap.IsEmpty() {
				t.Fatal("expected empty spectrum heatmap on NaN")
			}
			if !swingA.Update(math.NaN(), barTime).Heatmap.IsEmpty() {
				t.Fatal("expected empty swing heatmap on NaN")
			}
			if !trendA.Update(math.NaN(), barTime).Heatmap.IsEmpty() {
				t.Fatal("expected empty trend heatmap on NaN")
			}
			if !snrA.UpdateRange(math.NaN(), 2.0, barTime).Heatmap.IsEmpty() {
				t.Fatal("expected empty snr heatmap on NaN")
			}
		}
		specLastA = specA.Update(price, barTime)
		specOutB := specB.Update(price, barTime)
		if !sameHeatmap(specLastA.Heatmap, specOutB.Heatmap, engineTolerance) {
			t.Fatalf("spectrum NaN skip mismatch at %d", i)
		}

		swingLastA = swingA.Update(price, barTime)
		swingOutB := swingB.Update(price, barTime)
		if !sameHeatmap(swingLastA.Heatmap, swingOutB.Heatmap, engineTolerance) {
			t.Fatalf("swing NaN skip mismatch at %d", i)
		}

		trendLastA = trendA.Update(price, barTime)
		trendOutB := trendB.Update(price, barTime)
		if !sameHeatmap(trendLastA.Heatmap, trendOutB.Heatmap, engineTolerance) {
			t.Fatalf("trend NaN skip mismatch at %d", i)
		}

		snrLastA = snrA.UpdateRange(price, 2.0, barTime)
		snrOutB := snrB.UpdateRange(price, 2.0, barTime)
		if !sameHeatmap(snrLastA.Heatmap, snrOutB.Heatmap, engineTolerance) {
			t.Fatalf("snr NaN skip mismatch at %d", i)
		}
	}

	specLastA.Heatmap.Values[0] = -999.0
	swingLastA.Heatmap.Values[0] = -999.0
	trendLastA.Heatmap.Values[0] = -999.0
	snrLastA.Heatmap.Values[0] = -999.0

	extraTime := baseTime.AddDate(0, 0, 60)
	if !sameHeatmap(specA.Update(102.0, extraTime).Heatmap, specB.Update(102.0, extraTime).Heatmap, engineTolerance) {
		t.Fatal("mutating returned spectrum heatmap affected internal state")
	}
	if !sameHeatmap(swingA.Update(102.0, extraTime).Heatmap, swingB.Update(102.0, extraTime).Heatmap, engineTolerance) {
		t.Fatal("mutating returned swing heatmap affected internal state")
	}
	if !sameHeatmap(trendA.Update(102.0, extraTime).Heatmap, trendB.Update(102.0, extraTime).Heatmap, engineTolerance) {
		t.Fatal("mutating returned trend heatmap affected internal state")
	}
	if !sameHeatmap(snrA.UpdateRange(102.0, 2.0, extraTime).Heatmap, snrB.UpdateRange(102.0, 2.0, extraTime).Heatmap, engineTolerance) {
		t.Fatal("mutating returned snr heatmap affected internal state")
	}
}

func TestRejectBadParametersAndNonFiniteBeforeMutation(t *testing.T) {
	t.Parallel()

	badConfigs := []corona.Params{
		{HighPassFilterCutoff: 30, MinimalPeriod: 2, MaximalPeriod: 30, DecibelsLowerThreshold: 6, DecibelsUpperThreshold: 20},
		{HighPassFilterCutoff: 30, MinimalPeriod: 0, MaximalPeriod: 30, DecibelsLowerThreshold: 6, DecibelsUpperThreshold: 20},
		{HighPassFilterCutoff: 30, MinimalPeriod: 6, MaximalPeriod: 6, DecibelsLowerThreshold: 6, DecibelsUpperThreshold: 20},
		{HighPassFilterCutoff: 4, MinimalPeriod: 6, MaximalPeriod: 30, DecibelsLowerThreshold: 6, DecibelsUpperThreshold: 20},
		{HighPassFilterCutoff: 0, MinimalPeriod: 6, MaximalPeriod: 30, DecibelsLowerThreshold: 6, DecibelsUpperThreshold: 20},
		{HighPassFilterCutoff: 30, MinimalPeriod: 6, MaximalPeriod: 30, DecibelsLowerThreshold: math.NaN(), DecibelsUpperThreshold: 20},
		{HighPassFilterCutoff: 30, MinimalPeriod: 6, MaximalPeriod: 30, DecibelsLowerThreshold: -1.0, DecibelsUpperThreshold: 20},
		{HighPassFilterCutoff: 30, MinimalPeriod: 6, MaximalPeriod: 30, DecibelsLowerThreshold: 20.0, DecibelsUpperThreshold: 20.0},
		{HighPassFilterCutoff: 30, MinimalPeriod: 6, MaximalPeriod: 30, DecibelsLowerThreshold: 6.0, DecibelsUpperThreshold: math.Inf(1)},
	}
	for i, bad := range badConfigs {
		candidate := bad
		if _, err := corona.NewCorona(&candidate); err == nil {
			t.Fatalf("badConfigs[%d]: expected error, got nil", i)
		}
	}

	zeroLower := corona.DefaultParams()
	zeroLower.DecibelsLowerThreshold = 0.0
	engineZero, err := corona.NewCorona(zeroLower)
	if err != nil {
		t.Fatalf("explicit 0.0 DecibelsLowerThreshold rejected: %v", err)
	}
	if engineZero.Parameters().DecibelsLowerThreshold != 0.0 {
		t.Fatalf("DecibelsLowerThreshold = %v, want 0.0", engineZero.Parameters().DecibelsLowerThreshold)
	}

	engine, _ := corona.NewCorona(nil)
	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for i, price := range testInput[:30] {
		engine.Update(price, baseTime.AddDate(0, 0, i))
	}
	before, hasBefore := engine.LastSnapshot()
	if !hasBefore {
		t.Fatal("expected LastSnapshot after 30 samples")
	}
	for _, badPrice := range []float64{math.Inf(1), math.Inf(-1), 1e308} {
		if _, err := engine.TryUpdate(badPrice, baseTime); err == nil {
			t.Fatalf("expected error for price=%v", badPrice)
		}
		if engine.SampleCount() != 30 {
			t.Fatalf("SampleCount mutated to %d", engine.SampleCount())
		}
		after, hasAfter := engine.LastSnapshot()
		if !hasAfter || after.DominantCycle != before.DominantCycle {
			t.Fatal("LastSnapshot mutated after failed TryUpdate")
		}
	}

	engine.Reset()
	_, hasReset := engine.LastSnapshot()
	if engine.SampleCount() != 0 || hasReset {
		t.Fatal("Reset did not clear engine state")
	}
}

func TestFullCycleTrendAndPhaseLead(t *testing.T) {
	t.Parallel()

	consumer, _ := coronatrendvigor.NewCoronaTrendVigorDefault()
	engine, _ := corona.NewCorona(nil)
	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 120; i++ {
		price := 100.0 + 0.5*float64(i) + 3.0*math.Sin(corona.TwoPi*float64(i)/20.0)
		snapshot := engine.Update(price, baseTime.AddDate(0, 0, i))
		snapshot.DominantCycle = 20.0
		snapshot.DominantCycleMedian = 20.0
		output := consumer.UpdateFromSnapshot(snapshot)
		if i < 20 {
			if !output.Heatmap.IsEmpty() {
				t.Fatalf("bar %d: expected empty heatmap", i)
			}
		} else {
			if output.Lookback != 20.0 {
				t.Fatalf("bar %d: Lookback = %v, want 20.0", i, output.Lookback)
			}
			if math.Abs(output.TrendDisplacement-10.0) > 1e-12 {
				t.Fatalf("bar %d: TrendDisplacement = %v, want 10.0", i, output.TrendDisplacement)
			}
		}
	}
	if !consumer.IsPrimed() {
		t.Fatal("expected consumer IsPrimed()=true")
	}

	bandpass := corona.NewAdaptiveBandPass()
	previousBP := 0.0
	for i := 0; i < 100; i++ {
		cycle := bandpass.Update(100.0+math.Sin(corona.TwoPi*float64(i)/20.0), 20.0)
		if i >= 2 {
			expectedQuadrature := (cycle.BandPass - previousBP) / (corona.TwoPi / 20.0)
			if math.Abs(cycle.Quadrature-expectedQuadrature) > 1e-14 {
				t.Fatalf("bar %d: Quadrature = %v, want %v", i, cycle.Quadrature, expectedQuadrature)
			}
			expectedLead := 0.5*cycle.BandPass + corona.SqrtThreeOverTwo*cycle.Quadrature
			if math.Abs(cycle.Lead60-expectedLead) > 1e-14 {
				t.Fatalf("bar %d: Lead60 = %v, want %v", i, cycle.Lead60, expectedLead)
			}
		}
		previousBP = cycle.BandPass
	}
}

func TestConstantZeroPricesSNRAndRaster(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, price := range []float64{0.0, 100.0} {
		spec, _ := coronaspectrum.NewCoronaSpectrumDefault()
		swing, _ := coronaswingposition.NewCoronaSwingPositionDefault()
		trend, _ := coronatrendvigor.NewCoronaTrendVigorDefault()
		snr, _ := coronasignaltonoiseratio.NewCoronaSignalToNoiseRatioDefault()

		var specOut coronaspectrum.Snapshot
		var swingOut coronaswingposition.Snapshot
		var trendOut coronatrendvigor.Snapshot
		var snrOut coronasignaltonoiseratio.Snapshot
		for i := 0; i < 100; i++ {
			barTime := baseTime.AddDate(0, 0, i)
			specOut = spec.Update(price, barTime)
			swingOut = swing.Update(price, barTime)
			trendOut = trend.Update(price, barTime)
			snrOut = snr.UpdateRange(price, 2.0, barTime)
		}
		if specOut.Heatmap.IsEmpty() || specOut.HasCycleEstimate != 0.0 {
			t.Fatalf("price %v: unexpected spectrum output", price)
		}
		if swingOut.NormalizedPosition != 0.5 || swingOut.SwingPosition != 0.0 {
			t.Fatalf("price %v: unexpected swing position %v/%v", price, swingOut.NormalizedPosition, swingOut.SwingPosition)
		}
		for _, cell := range swingOut.Heatmap.Values {
			if cell != 20.0 {
				t.Fatalf("price %v: expected flat swing raster cell 20.0, got %v", price, cell)
			}
		}
		if trendOut.Heatmap.IsEmpty() || trendOut.HasCycleEstimate != 0.0 {
			t.Fatalf("price %v: unexpected trend output", price)
		}
		if price > 0.0 {
			if !math.IsInf(snrOut.PhysicalDecibels, -1) {
				t.Fatalf("price %v: expected -Inf PhysicalDecibels, got %v", price, snrOut.PhysicalDecibels)
			}
		} else if !math.IsNaN(snrOut.PhysicalDecibels) {
			t.Fatalf("price %v: expected NaN PhysicalDecibels, got %v", price, snrOut.PhysicalDecibels)
		}
		if snrOut.SignalToNoiseRatio != 1.0 {
			t.Fatalf("price %v: SignalToNoiseRatio = %v, want 1.0", price, snrOut.SignalToNoiseRatio)
		}
	}

	consumer, _ := coronasignaltonoiseratio.NewCoronaSignalToNoiseRatioDefault()
	engine, _ := corona.NewCorona(nil)
	var snrSnapshot coronasignaltonoiseratio.Snapshot
	for i := 0; i < 350; i++ {
		snapshot := engine.Update(100.0, baseTime.AddDate(0, 0, i))
		snapshot.MaximumPower = 4.0
		snrSnapshot = consumer.UpdateFromSnapshotRange(snapshot, 2.0)
	}
	if math.Abs(snrSnapshot.Signal-4.0) > 1e-12 || math.Abs(snrSnapshot.Noise-2.0) > 1e-12 {
		t.Fatalf("unexpected converged Signal/Noise: %v / %v", snrSnapshot.Signal, snrSnapshot.Noise)
	}
	expectedDB := 20.0 * math.Log10(2.0)
	if math.Abs(snrSnapshot.PhysicalDecibels-expectedDB) > 1e-12 {
		t.Fatalf("PhysicalDecibels = %v, want %v", snrSnapshot.PhysicalDecibels, expectedDB)
	}
	if math.Abs(snrSnapshot.SignalToNoiseRatio-(1.0+snrSnapshot.PhysicalDecibels+3.5)) > 1e-12 {
		t.Fatalf("SignalToNoiseRatio = %v", snrSnapshot.SignalToNoiseRatio)
	}

	snrInvalid, _ := coronasignaltonoiseratio.NewCoronaSignalToNoiseRatioDefault()
	if _, err := snrInvalid.TryUpdate(100.0, 101.0, 99.0, baseTime); err == nil {
		t.Fatal("expected error for low > high")
	}
	if _, err := snrInvalid.TryUpdateRange(100.0, -1.0, baseTime); err == nil {
		t.Fatal("expected error for negative barRange")
	}
	if _, err := snrInvalid.TryUpdate(100.0, 100.0, math.Inf(1), baseTime); err == nil {
		t.Fatal("expected error for infinite high")
	}
	if snrInvalid.IsPrimed() {
		t.Fatal("failed TryUpdate should not prime SNR")
	}
	if !snrInvalid.Update(100.0, 99.0, math.NaN(), baseTime).Heatmap.IsEmpty() {
		t.Fatal("expected NaN high to skip bar")
	}

	raster, err := corona.NewRaster(50, -5.0, 5.0, 20.0)
	if err != nil {
		t.Fatalf("NewRaster: %v", err)
	}
	raster.Update(1.0, 0.1, corona.RasterSwing, false)
	column := raster.Column(baseTime)
	if column.ParameterFirst != -4.8 || column.ParameterLast != 5.0 || column.ParameterResolution != 5.0 {
		t.Fatalf("unexpected raster geometry: first=%v last=%v resolution=%v",
			column.ParameterFirst, column.ParameterLast, column.ParameterResolution)
	}
	if column.Values[49] != 0.0 || column.ValueMin != 0.0 || column.ValueMax != 20.0 {
		t.Fatalf("unexpected raster endpoint values: last=%v min=%v max=%v",
			column.Values[49], column.ValueMin, column.ValueMax)
	}

	seed := make([]float64, 50)
	for i := range seed {
		seed[i] = 10.0
	}
	raster.SetValues(seed)
	raster.Update(0.51, 2.0, corona.RasterSwing, false)
	if raster.ValuesCopy()[25] != 5.0 {
		t.Fatalf("half-up marker index 25 = %v, want 5.0", raster.ValuesCopy()[25])
	}

	snrRaster, _ := corona.NewRaster(50, 1.0, 11.0, 20.0)
	snrRaster.SetValues(seed)
	snrRaster.Update(0.5, 4.0, corona.RasterSNR, false)
	snrValues := snrRaster.ValuesCopy()
	if snrValues[24] != 5.0 || snrValues[25] != 20.0 || snrValues[23] >= 20.0 {
		t.Fatalf("unexpected SNR corona plume: [23]=%v [24]=%v [25]=%v",
			snrValues[23], snrValues[24], snrValues[25])
	}
}

func TestFactoryMergesOntoDefaults(t *testing.T) {
	t.Parallel()

	cases := []struct {
		id       core.Identifier
		json     string
		mnemonic string
	}{
		{core.CoronaSpectrum, `{"minRasterValue": 0}`, "cspect(0, 20, 6, 30, 30)"},
		{core.CoronaSwingPosition, `{"leadWindow": 41}`, "cswing(50, 20, -5, 5, 30, lead=41)"},
		{core.CoronaTrendVigor, `{"lookbackOffset": -1}`, "ctv(50, 20, -10, 10, 30, offset=-1)"},
		{core.CoronaSignalToNoiseRatio, `{"minParameterValue": -2, "maxParameterValue": 8}`, "csnr(50, 20, -2, 8, 30)"},
	}
	for _, tc := range cases {
		indicator, err := factory.New(tc.id, tc.json)
		if err != nil {
			t.Fatalf("factory.New(%s, %s): %v", tc.id, tc.json, err)
		}
		if got := indicator.Metadata().Mnemonic; got != tc.mnemonic {
			t.Fatalf("factory.New(%s, %s) mnemonic = %q, want %q", tc.id, tc.json, got, tc.mnemonic)
		}
	}
}
