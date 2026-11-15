//nolint:testpackage
package zerolagerrorcorrectingexponentialmovingaverage

//nolint: gofumpt
import (
	"fmt"
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

type articleRow struct {
	ec         float64
	ema        float64
	bestGain   float64
	leastError float64
}

func articleBatch(prices []float64, length int, gainLimit int) []articleRow {
	alpha := 2.0 / (float64(length) + 1.0)
	ema := make([]float64, len(prices)+1)
	ec := make([]float64, len(prices)+1)
	bestGain := 0.0
	rows := make([]articleRow, 0, len(prices))
	for step, price := range prices {
		index := step + 1
		ema[index] = alpha*price + (1.0-alpha)*ema[index-1]
		leastError := 1_000_000.0
		for integerGain := -gainLimit; integerGain <= gainLimit; integerGain++ {
			gain := float64(integerGain) / 10.0
			trial := alpha*(ema[index]+gain*(price-ec[index-1])) + (1.0-alpha)*ec[index-1]
			errorValue := price - trial
			if math.Abs(errorValue) < leastError {
				leastError = math.Abs(errorValue)
				bestGain = gain
			}
		}
		ec[index] = alpha*(ema[index]+bestGain*(price-ec[index-1])) + (1.0-alpha)*ec[index-1]
		rows = append(rows, articleRow{
			ec:         ec[index],
			ema:        ema[index],
			bestGain:   bestGain,
			leastError: leastError,
		})
	}
	return rows
}

func assertCloseOrNaN(t *testing.T, label string, actual, expected, tolerance float64) {
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
	scale := math.Max(1.0, math.Max(math.Abs(actual), math.Abs(expected)))
	if math.Abs(actual-expected) > tolerance*scale {
		t.Fatalf("%s: expected %.17g, got %.17g (diff=%.3e)", label, expected, actual, math.Abs(actual-expected))
	}
}

func testZecemaTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

func TestZecemaReferenceFixturesAndArticleBatch(t *testing.T) {
	t.Parallel()

	cases := []struct {
		length      int
		gainLimit   int
		expectedEc  []float64
		expectedEma []float64
	}{
		{20, 50, expectedL20_G50_EC, expectedL20_G50_EMA},
		{12, 50, expectedL12_G50_EC, expectedL12_G50_EMA},
		{32, 22, expectedL32_G22_EC, expectedL32_G22_EMA},
		{20, 0, expectedL20_G0_EC, expectedL20_G0_EMA},
		{20, 1, expectedL20_G1_EC, expectedL20_G1_EMA},
		{20, 10, expectedL20_G10_EC, expectedL20_G10_EMA},
		{20, 22, expectedL20_G22_EC, expectedL20_G22_EMA},
		{20, 100, expectedL20_G100_EC, expectedL20_G100_EMA},
		{1, 50, expectedL1_G50_EC, expectedL1_G50_EMA},
		{2, 50, expectedL2_G50_EC, expectedL2_G50_EMA},
		{5, 50, expectedL5_G50_EC, expectedL5_G50_EMA},
		{10, 50, expectedL10_G50_EC, expectedL10_G50_EMA},
		{40, 50, expectedL40_G50_EC, expectedL40_G50_EMA},
		{100, 50, expectedL100_G50_EC, expectedL100_G50_EMA},
		{252, 50, expectedL252_G50_EC, expectedL252_G50_EMA},
		{20, 49, expectedL20_G49_EC, expectedL20_G49_EMA},
	}

	baseTime := testZecemaTime()
	for _, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("L%d_G%d", tc.length, tc.gainLimit), func(t *testing.T) {
			t.Parallel()

			stream, err := NewZeroLagErrorCorrectingExponentialMovingAverageLength(&LengthParams{
				Length:    tc.length,
				GainLimit: tc.gainLimit,
			})
			if err != nil {
				t.Fatalf("unexpected constructor error: %v", err)
			}
			entityStream, err := NewZeroLagErrorCorrectingExponentialMovingAverageLength(&LengthParams{
				Length:    tc.length,
				GainLimit: tc.gainLimit,
			})
			if err != nil {
				t.Fatalf("unexpected constructor error: %v", err)
			}

			reference := articleBatch(testInput, tc.length, tc.gainLimit)
			for barIndex, price := range testInput {
				ec := stream.Update(price)
				ema := stream.Ema()
				assertCloseOrNaN(t, fmt.Sprintf("bar %d fixture EC", barIndex), ec, tc.expectedEc[barIndex], 2e-12)
				assertCloseOrNaN(t, fmt.Sprintf("bar %d fixture EMA", barIndex), ema, tc.expectedEma[barIndex], 2e-12)
				assertCloseOrNaN(t, fmt.Sprintf("bar %d batch EC", barIndex), ec, reference[barIndex].ec, 2e-12)
				assertCloseOrNaN(t, fmt.Sprintf("bar %d batch EMA", barIndex), ema, reference[barIndex].ema, 2e-12)
				if stream.BestGain() != reference[barIndex].bestGain {
					t.Fatalf("bar %d bestGain: expected %v, got %v", barIndex, reference[barIndex].bestGain, stream.BestGain())
				}
				if stream.LeastError() != reference[barIndex].leastError {
					t.Fatalf("bar %d leastError: expected %v, got %v", barIndex, reference[barIndex].leastError, stream.LeastError())
				}

				stamp := baseTime.AddDate(0, 0, barIndex)
				entityOut := entityStream.UpdateScalar(&entities.Scalar{Time: stamp, Value: price})
				if len(entityOut) != 2 {
					t.Fatalf("bar %d len(entityOut) = %d, want 2", barIndex, len(entityOut))
				}
				outEc := entityOut[0].(entities.Scalar)
				outEma := entityOut[1].(entities.Scalar)
				if outEc.Time != stamp || outEma.Time != stamp {
					t.Fatalf("bar %d timestamp mismatch", barIndex)
				}
				if outEc.Value != ec || outEma.Value != ema {
					t.Fatalf("bar %d entity output mismatch: got (%v, %v), want (%v, %v)", barIndex, outEc.Value, outEma.Value, ec, ema)
				}
			}
		})
	}
}

func TestZecemaZeroStateStartupAndFirstObservationPriming(t *testing.T) {
	t.Parallel()

	stream, err := NewZeroLagErrorCorrectingExponentialMovingAverageDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stream.IsPrimed() {
		t.Fatal("should not be primed before any updates")
	}
	if !math.IsNaN(stream.Value()) || !math.IsNaN(stream.Ema()) || !math.IsNaN(stream.BestGain()) || !math.IsNaN(stream.LeastError()) {
		t.Fatal("expected NaN diagnostic getters before priming")
	}
	if stream.Alpha() != 2.0/21.0 || stream.Alpha() == 0.095 {
		t.Fatalf("expected exact alpha 2/21, got %.17g", stream.Alpha())
	}
	if stream.WarmupBars() != 0 {
		t.Fatalf("expected WarmupBars() == 0, got %d", stream.WarmupBars())
	}
	if stream.Length() != 20 || stream.GainLimit() != 50 {
		t.Fatalf("unexpected default parameters: (%d, %d)", stream.Length(), stream.GainLimit())
	}

	price := testInput[0]
	ec := stream.Update(price)
	expectedEma := (2.0 / 21.0) * price
	expectedEc := (2.0 / 21.0) * (expectedEma + 5.0*price)
	if stream.Ema() != expectedEma {
		t.Fatalf("expected ema %.17g, got %.17g", expectedEma, stream.Ema())
	}
	if ec != expectedEc || stream.Value() != expectedEc {
		t.Fatalf("expected ec %.17g, got %.17g", expectedEc, ec)
	}
	if stream.BestGain() != 5.0 {
		t.Fatalf("expected bestGain 5.0, got %v", stream.BestGain())
	}
	if !stream.IsPrimed() || stream.SamplesSeen() != 1 {
		t.Fatalf("expected primed with SamplesSeen()==1, got primed=%v count=%d", stream.IsPrimed(), stream.SamplesSeen())
	}
}

func TestZecemaInclusiveEndpointsFrozenPreviousEcAndTieBreaking(t *testing.T) {
	t.Parallel()

	for _, limit := range []int{1, 22, 49, 50} {
		stream, err := NewZeroLagErrorCorrectingExponentialMovingAverageLength(&LengthParams{
			Length:    20,
			GainLimit: limit,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		stream.Update(testInput[0])
		expectedGain := float64(limit) / 10.0
		if stream.BestGain() != expectedGain {
			t.Fatalf("limit=%d: expected bestGain %v, got %v", limit, expectedGain, stream.BestGain())
		}
	}

	frozenStream, err := NewZeroLagErrorCorrectingExponentialMovingAverageLength(&LengthParams{
		Length:    20,
		GainLimit: 22,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, price := range testInput[:60] {
		previousEc := frozenStream.ecState
		previousEma := frozenStream.emaState
		alpha := frozenStream.Alpha()
		ema := alpha*price + (1.0-alpha)*previousEma
		bestError := math.Inf(1)
		winner := 0.0
		winnerGain := 0.0
		for integerGain := -22; integerGain <= 22; integerGain++ {
			gain := float64(integerGain) / 10.0
			candidate := alpha*(ema+gain*(price-previousEc)) + (1.0-alpha)*previousEc
			errorValue := math.Abs(price - candidate)
			if errorValue < bestError {
				bestError = errorValue
				winner = candidate
				winnerGain = gain
			}
		}
		ec := frozenStream.Update(price)
		if ec != winner || frozenStream.BestGain() != winnerGain || frozenStream.Ema() != ema {
			t.Fatalf("frozen previous EC check failed: got (%v, %v, %v), want (%v, %v, %v)", ec, frozenStream.BestGain(), frozenStream.Ema(), winner, winnerGain, ema)
		}
	}

	tieStream, err := NewZeroLagErrorCorrectingExponentialMovingAverageLength(&LengthParams{
		Length:    20,
		GainLimit: 49,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for step := 0; step < 5; step++ {
		ec := tieStream.Update(0.0)
		if ec != 0.0 || tieStream.Ema() != 0.0 {
			t.Fatalf("step %d: expected (0, 0), got (%v, %v)", step, ec, tieStream.Ema())
		}
		if tieStream.LeastError() != 0.0 || tieStream.BestGain() != -4.9 {
			t.Fatalf("step %d: expected leastError=0 bestGain=-4.9, got %v %v", step, tieStream.LeastError(), tieStream.BestGain())
		}
	}
}

func TestZecemaGainZeroAndBaselineIndependence(t *testing.T) {
	t.Parallel()

	zeroGainStream, err := NewZeroLagErrorCorrectingExponentialMovingAverageLength(&LengthParams{
		Length:    20,
		GainLimit: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	alpha := 2.0 / 21.0
	complement := 1.0 - alpha
	var lastEc, lastEma float64
	for step := 0; step < 50; step++ {
		sample := 0.0
		if step == 0 {
			sample = 1.0
		}
		lastEc = zeroGainStream.Update(sample)
		lastEma = zeroGainStream.Ema()
		expectedEma := alpha * math.Pow(complement, float64(step))
		expectedEc := float64(step+1) * alpha * alpha * math.Pow(complement, float64(step))
		assertCloseOrNaN(t, fmt.Sprintf("step %d ema", step), lastEma, expectedEma, 1e-14)
		assertCloseOrNaN(t, fmt.Sprintf("step %d ec", step), lastEc, expectedEc, 1e-14)
		if zeroGainStream.BestGain() != 0.0 {
			t.Fatalf("step %d: expected bestGain 0.0, got %v", step, zeroGainStream.BestGain())
		}
	}
	if lastEc == lastEma {
		t.Fatal("expected cascaded EMA output to differ from single baseline EMA")
	}

	limits := []int{0, 1, 22, 50, 100}
	streams := make([]*ZeroLagErrorCorrectingExponentialMovingAverage, len(limits))
	for index, limit := range limits {
		streams[index], err = NewZeroLagErrorCorrectingExponentialMovingAverageLength(&LengthParams{
			Length:    20,
			GainLimit: limit,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	ema := 0.0
	for _, price := range testInput {
		ema = alpha*price + complement*ema
		for _, stream := range streams {
			stream.Update(price)
			if stream.Ema() != ema {
				t.Fatalf("expected baseline ema %.17g, got %.17g", ema, stream.Ema())
			}
		}
	}
}

func TestZecemaFiniteSentinelAlphaOneLongLengthNaNResetAndOverflow(t *testing.T) {
	t.Parallel()

	sentinelStream, _ := NewZeroLagErrorCorrectingExponentialMovingAverageDefault()
	sentinelStream.Update(0.0)
	if sentinelStream.BestGain() != -5.0 {
		t.Fatalf("expected -5.0, got %v", sentinelStream.BestGain())
	}
	largeEc := sentinelStream.Update(100_000_000.0)
	if sentinelStream.BestGain() != -5.0 || sentinelStream.LeastError() != 1_000_000.0 {
		t.Fatalf("expected retained bestGain=-5.0 and sentinel leastError=1e6, got %v and %v", sentinelStream.BestGain(), sentinelStream.LeastError())
	}
	if sentinelStream.LeastError() == math.Abs(100_000_000.0-largeEc) || !(largeEc < 0.0) {
		t.Fatalf("unexpected sentinel EC behavior: ec=%v leastError=%v", largeEc, sentinelStream.LeastError())
	}
	expected := articleBatch([]float64{0.0, 100_000_000.0}, 20, 50)[1]
	assertCloseOrNaN(t, "sentinel EC", largeEc, expected.ec, 2e-12)
	assertCloseOrNaN(t, "sentinel EMA", sentinelStream.Ema(), expected.ema, 2e-12)

	alphaOne, _ := NewZeroLagErrorCorrectingExponentialMovingAverageLength(&LengthParams{Length: 1, GainLimit: 50})
	for _, price := range testInput {
		if ec := alphaOne.Update(price); ec != price || alphaOne.Ema() != price {
			t.Fatalf("length=1 expected (%v, %v), got (%v, %v)", price, price, ec, alphaOne.Ema())
		}
	}

	longStream, _ := NewZeroLagErrorCorrectingExponentialMovingAverageLength(&LengthParams{Length: 252, GainLimit: 50})
	for _, price := range testInput {
		ec := longStream.Update(price)
		if !isFinite(ec) || !isFinite(longStream.Ema()) || !longStream.IsPrimed() {
			t.Fatal("length=252 should produce finite outputs from bar 0")
		}
	}

	nanStream, _ := NewZeroLagErrorCorrectingExponentialMovingAverageDefault()
	controlStream, _ := NewZeroLagErrorCorrectingExponentialMovingAverageDefault()
	for _, price := range testInput[:80] {
		ec := nanStream.Update(price)
		controlEc := controlStream.Update(price)
		if ec != controlEc || nanStream.Ema() != controlStream.Ema() {
			t.Fatal("stream diverged before NaN gap")
		}
		savedValue, savedEma := nanStream.Value(), nanStream.Ema()
		savedCount, savedGain := nanStream.SamplesSeen(), nanStream.BestGain()
		if !math.IsNaN(nanStream.Update(math.NaN())) {
			t.Fatal("expected NaN return on NaN sample")
		}
		nanEntity := nanStream.UpdateScalar(&entities.Scalar{Time: testZecemaTime(), Value: math.NaN()})
		if !math.IsNaN(nanEntity[0].(entities.Scalar).Value) || !math.IsNaN(nanEntity[1].(entities.Scalar).Value) {
			t.Fatal("expected [NaN, NaN] entity output on NaN sample")
		}
		if nanStream.Value() != savedValue || nanStream.Ema() != savedEma || nanStream.SamplesSeen() != savedCount || nanStream.BestGain() != savedGain {
			t.Fatal("NaN sample mutated internal state")
		}
	}
	nanStream.Reset()
	if nanStream.IsPrimed() || nanStream.SamplesSeen() != 0 || !math.IsNaN(nanStream.Value()) || !math.IsNaN(nanStream.Ema()) {
		t.Fatal("Reset did not restore cold-start state")
	}
	freshStream, _ := NewZeroLagErrorCorrectingExponentialMovingAverageDefault()
	for _, price := range testInput {
		if nanStream.Update(price) != freshStream.Update(price) || nanStream.Ema() != freshStream.Ema() {
			t.Fatal("Reset stream diverged from fresh instance")
		}
	}

	overflowStream, _ := NewZeroLagErrorCorrectingExponentialMovingAverageDefault()
	overflowStream.Update(testInput[0])
	savedValue := overflowStream.Value()
	savedEma := overflowStream.Ema()
	savedCount := overflowStream.SamplesSeen()
	savedBestGain := overflowStream.BestGain()
	savedLeastError := overflowStream.LeastError()
	for _, badSample := range []float64{math.Inf(1), math.Inf(-1), 1e308} {
		if _, err := overflowStream.TryUpdate(badSample); err == nil {
			t.Fatalf("expected TryUpdate error for %v", badSample)
		}
		if overflowStream.Value() != savedValue || overflowStream.Ema() != savedEma ||
			overflowStream.SamplesSeen() != savedCount || overflowStream.BestGain() != savedBestGain ||
			overflowStream.LeastError() != savedLeastError {
			t.Fatalf("failed sample %v mutated indicator state", badSample)
		}
	}
}

func TestZecemaSmoothingFactorVariantAndEntities(t *testing.T) {
	t.Parallel()

	fromLength, err := NewZeroLagErrorCorrectingExponentialMovingAverageDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fromAlpha, err := NewZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactor(DefaultSmoothingFactorParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fromAlpha.Length() != 20 || fromAlpha.GainLimit() != 50 {
		t.Fatalf("unexpected smoothing-factor derived properties: length=%d gainLimit=%d", fromAlpha.Length(), fromAlpha.GainLimit())
	}
	alphaMeta := fromAlpha.Metadata()
	if alphaMeta.Outputs[0].Mnemonic != "zecema(20, 0.09523810, 50)" {
		t.Fatalf("unexpected smoothing-factor EC mnemonic: %s", alphaMeta.Outputs[0].Mnemonic)
	}
	if alphaMeta.Outputs[1].Mnemonic != "zecemaEma(20, 0.09523810, 50)" {
		t.Fatalf("unexpected smoothing-factor EMA mnemonic: %s", alphaMeta.Outputs[1].Mnemonic)
	}
	for barIndex, price := range testInput {
		if fromAlpha.Update(price) != fromLength.Update(price) || fromAlpha.Ema() != fromLength.Ema() {
			t.Fatalf("bar %d: smoothing-factor variant diverged from length variant", barIndex)
		}
	}

	stamp := testZecemaTime()
	const sample = 100.0
	checkEntity := func(output core.Output) {
		t.Helper()
		if len(output) != 2 {
			t.Fatalf("expected 2 outputs, got %d", len(output))
		}
		for index := 0; index < 2; index++ {
			scalarOut, ok := output[index].(entities.Scalar)
			if !ok {
				t.Fatalf("output[%d] is not entities.Scalar", index)
			}
			if scalarOut.Time != stamp || math.IsNaN(scalarOut.Value) {
				t.Fatalf("output[%d] unexpected scalar %+v", index, scalarOut)
			}
		}
	}

	s1, _ := NewZeroLagErrorCorrectingExponentialMovingAverageDefault()
	checkEntity(s1.UpdateScalar(&entities.Scalar{Time: stamp, Value: sample}))
	s2, _ := NewZeroLagErrorCorrectingExponentialMovingAverageDefault()
	checkEntity(s2.UpdateBar(&entities.Bar{Time: stamp, Close: sample}))
	s3, _ := NewZeroLagErrorCorrectingExponentialMovingAverageDefault()
	checkEntity(s3.UpdateQuote(&entities.Quote{Time: stamp, Bid: sample, Ask: sample}))
	s4, _ := NewZeroLagErrorCorrectingExponentialMovingAverageDefault()
	checkEntity(s4.UpdateTrade(&entities.Trade{Time: stamp, Price: sample}))
}

func TestZecemaMetadata(t *testing.T) {
	t.Parallel()

	z, _ := NewZeroLagErrorCorrectingExponentialMovingAverageDefault()
	act := z.Metadata()
	desc, ok := core.DescriptorOf(core.ZeroLagErrorCorrectingExponentialMovingAverage)
	if !ok {
		t.Fatal("missing descriptor for ZeroLagErrorCorrectingExponentialMovingAverage")
	}

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	check("Identifier", core.ZeroLagErrorCorrectingExponentialMovingAverage, act.Identifier)
	check("Adaptivity", core.Adaptive, desc.Adaptivity)
	check("len(Outputs)", 2, len(act.Outputs))
	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", "zecema(20, 50)", act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", "Zero-lag Error-Correcting Exponential Moving Average zecema(20, 50)", act.Outputs[0].Description)
	check("Outputs[1].Kind", int(Ema), act.Outputs[1].Kind)
	check("Outputs[1].Shape", shape.Scalar, act.Outputs[1].Shape)
	check("Outputs[1].Mnemonic", "zecemaEma(20, 50)", act.Outputs[1].Mnemonic)
	check("Outputs[1].Description", "Zero-lag Error-Correcting Exponential Moving Average baseline EMA zecemaEma(20, 50)", act.Outputs[1].Description)
}

func TestZecemaValidationAndMnemonics(t *testing.T) {
	t.Parallel()

	if _, err := NewZeroLagErrorCorrectingExponentialMovingAverageLength(nil); err == nil {
		t.Fatal("expected error for nil LengthParams")
	}
	if _, err := NewZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactor(nil); err == nil {
		t.Fatal("expected error for nil SmoothingFactorParams")
	}
	for _, badLength := range []int{0, -1, math.MaxInt} {
		if _, err := NewZeroLagErrorCorrectingExponentialMovingAverageLength(&LengthParams{Length: badLength, GainLimit: 50}); err == nil {
			t.Fatalf("expected error for length=%d", badLength)
		}
	}
	if _, err := NewZeroLagErrorCorrectingExponentialMovingAverageLength(&LengthParams{Length: 20, GainLimit: -1}); err == nil {
		t.Fatal("expected error for negative GainLimit")
	}
	for _, badAlpha := range []float64{0.0, -0.1, 1.1, 1e-30, math.NaN(), math.Inf(1)} {
		if _, err := NewZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactor(&SmoothingFactorParams{SmoothingFactor: badAlpha, GainLimit: 50}); err == nil {
			t.Fatalf("expected error for smoothingFactor=%v", badAlpha)
		}
	}
	if _, err := NewZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactor(&SmoothingFactorParams{SmoothingFactor: 1.0, GainLimit: 0}); err != nil {
		t.Fatalf("expected smoothingFactor=1.0 gainLimit=0 to succeed, got %v", err)
	}
	if _, err := NewZeroLagErrorCorrectingExponentialMovingAverageLength(&LengthParams{Length: 20, GainLimit: 50, BarComponent: entities.BarComponent(9999)}); err == nil {
		t.Fatal("expected error for invalid BarComponent")
	}
	if _, err := NewZeroLagErrorCorrectingExponentialMovingAverageLength(&LengthParams{Length: 20, GainLimit: 50, QuoteComponent: entities.QuoteComponent(9999)}); err == nil {
		t.Fatal("expected error for invalid QuoteComponent")
	}
	if _, err := NewZeroLagErrorCorrectingExponentialMovingAverageLength(&LengthParams{Length: 20, GainLimit: 50, TradeComponent: entities.TradeComponent(9999)}); err == nil {
		t.Fatal("expected error for invalid TradeComponent")
	}

	mnemonicCases := []struct {
		params  LengthParams
		wantEc  string
		wantEma string
	}{
		{LengthParams{Length: 20, GainLimit: 50}, "zecema(20, 50)", "zecemaEma(20, 50)"},
		{LengthParams{Length: 12, GainLimit: 0, BarComponent: entities.BarMedianPrice}, "zecema(12, 0, hl/2)", "zecemaEma(12, 0, hl/2)"},
		{LengthParams{Length: 20, GainLimit: 50, QuoteComponent: entities.QuoteBidPrice}, "zecema(20, 50, b)", "zecemaEma(20, 50, b)"},
		{LengthParams{Length: 20, GainLimit: 50, TradeComponent: entities.TradeVolume}, "zecema(20, 50, v)", "zecemaEma(20, 50, v)"},
		{LengthParams{Length: 32, GainLimit: 22, BarComponent: entities.BarOpenPrice, QuoteComponent: entities.QuoteAskPrice}, "zecema(32, 22, o, a)", "zecemaEma(32, 22, o, a)"},
	}
	for _, mc := range mnemonicCases {
		if got := mc.params.Mnemonic(); got != mc.wantEc {
			t.Fatalf("Mnemonic() = %q, want %q", got, mc.wantEc)
		}
		if got := mc.params.EmaMnemonic(); got != mc.wantEma {
			t.Fatalf("EmaMnemonic() = %q, want %q", got, mc.wantEma)
		}
	}

	sfParams := SmoothingFactorParams{
		SmoothingFactor: 0.2,
		GainLimit:       22,
		BarComponent:    entities.BarTypicalPrice,
		TradeComponent:  entities.TradeVolume,
	}
	if got := sfParams.Mnemonic(); got != "zecema(9, 0.20000000, 22, hlc/3, v)" {
		t.Fatalf("SmoothingFactorParams.Mnemonic() = %q", got)
	}
	if got := sfParams.EmaMnemonic(); got != "zecemaEma(9, 0.20000000, 22, hlc/3, v)" {
		t.Fatalf("SmoothingFactorParams.EmaMnemonic() = %q", got)
	}
}
