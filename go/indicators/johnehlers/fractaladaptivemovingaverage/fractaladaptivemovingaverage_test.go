//nolint:testpackage
package fractaladaptivemovingaverage

import (
	"fmt"
	"math"
	"sync"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

type referenceCase struct {
	params             Params
	expectedValues     []float64
	expectedDimensions []float64
}

func referenceCases() []referenceCase {
	return []referenceCase{
		{Params{Length: 16, SlowestSmoothingFactor: 0.01, CoefficientMode: 1}, expectedL16_A0_01_C1_VALUE, expectedL16_A0_01_C1_FDIM},
		{Params{Length: 2, SlowestSmoothingFactor: 0.01, CoefficientMode: 1}, expectedL2_A0_01_C1_VALUE, expectedL2_A0_01_C1_FDIM},
		{Params{Length: 4, SlowestSmoothingFactor: 0.01, CoefficientMode: 1}, expectedL4_A0_01_C1_VALUE, expectedL4_A0_01_C1_FDIM},
		{Params{Length: 8, SlowestSmoothingFactor: 0.01, CoefficientMode: 1}, expectedL8_A0_01_C1_VALUE, expectedL8_A0_01_C1_FDIM},
		{Params{Length: 20, SlowestSmoothingFactor: 0.01, CoefficientMode: 1}, expectedL20_A0_01_C1_VALUE, expectedL20_A0_01_C1_FDIM},
		{Params{Length: 30, SlowestSmoothingFactor: 0.01, CoefficientMode: 1}, expectedL30_A0_01_C1_VALUE, expectedL30_A0_01_C1_FDIM},
		{Params{Length: 64, SlowestSmoothingFactor: 0.01, CoefficientMode: 1}, expectedL64_A0_01_C1_VALUE, expectedL64_A0_01_C1_FDIM},
		{Params{Length: 128, SlowestSmoothingFactor: 0.01, CoefficientMode: 1}, expectedL128_A0_01_C1_VALUE, expectedL128_A0_01_C1_FDIM},
		{Params{Length: 250, SlowestSmoothingFactor: 0.01, CoefficientMode: 1}, expectedL250_A0_01_C1_VALUE, expectedL250_A0_01_C1_FDIM},
		{Params{Length: 252, SlowestSmoothingFactor: 0.01, CoefficientMode: 1}, expectedL252_A0_01_C1_VALUE, expectedL252_A0_01_C1_FDIM},
		{Params{Length: 254, SlowestSmoothingFactor: 0.01, CoefficientMode: 1}, expectedL254_A0_01_C1_VALUE, expectedL254_A0_01_C1_FDIM},
		{Params{Length: 16, SlowestSmoothingFactor: 0.001, CoefficientMode: 1}, expectedL16_A0_001_C1_VALUE, expectedL16_A0_001_C1_FDIM},
		{Params{Length: 16, SlowestSmoothingFactor: 0.1, CoefficientMode: 1}, expectedL16_A0_1_C1_VALUE, expectedL16_A0_1_C1_FDIM},
		{Params{Length: 16, SlowestSmoothingFactor: 1.0, CoefficientMode: 1}, expectedL16_A1_C1_VALUE, expectedL16_A1_C1_FDIM},
		{Params{Length: 16, SlowestSmoothingFactor: 0.01, CoefficientMode: 0}, expectedL16_A0_01_C0_VALUE, expectedL16_A0_01_C0_FDIM},
		{Params{Length: 30, SlowestSmoothingFactor: 0.01, CoefficientMode: 0}, expectedL30_A0_01_C0_VALUE, expectedL30_A0_01_C0_FDIM},
	}
}

func articleBatch(
	prices []float64,
	length int,
	slowest float64,
	mode int,
	highs, lows []float64,
) ([]float64, []float64) {
	if highs == nil {
		highs = prices
	}
	if lows == nil {
		lows = prices
	}
	half := length / 2
	scale := -4.6
	if mode == 1 {
		scale = math.Log(slowest)
	}
	dimension := 0.0
	filtered := make([]float64, len(prices))
	outValues := make([]float64, len(prices))
	outDimensions := make([]float64, len(prices))

	for index, price := range prices {
		if index+1 >= length {
			max1, min1 := highs[index], lows[index]
			for offset := 1; offset < half; offset++ {
				if highs[index-offset] > max1 {
					max1 = highs[index-offset]
				}
				if lows[index-offset] < min1 {
					min1 = lows[index-offset]
				}
			}
			max2, min2 := highs[index-half], lows[index-half]
			for offset := half + 1; offset < length; offset++ {
				if highs[index-offset] > max2 {
					max2 = highs[index-offset]
				}
				if lows[index-offset] < min2 {
					min2 = lows[index-offset]
				}
			}
			max3, min3 := math.Max(max1, max2), math.Min(min1, min2)
			n1 := (max1 - min1) / float64(half)
			n2 := (max2 - min2) / float64(half)
			n3 := (max3 - min3) / float64(length)
			if n1 > 0.0 && n2 > 0.0 && n3 > 0.0 {
				dimension = (math.Log(n1+n2) - math.Log(n3)) / math.Log(2.0)
			}
			rawAlpha := 1.0
			if dimension > 1.0 {
				rawAlpha = math.Exp(scale * (dimension - 1.0))
			}
			alpha := math.Max(slowest, math.Min(1.0, rawAlpha))
			if index+1 < length+1 {
				filtered[index] = price
			} else {
				filtered[index] = alpha*price + (1.0-alpha)*filtered[index-1]
			}
			outValues[index] = filtered[index]
			outDimensions[index] = dimension
		} else {
			filtered[index] = price
			outValues[index] = math.NaN()
			outDimensions[index] = math.NaN()
		}
	}
	return outValues, outDimensions
}

func TestFractalAdaptiveMovingAverageReferenceFixtures(t *testing.T) {
	t.Parallel()

	if len(testInput) != 252 {
		t.Fatalf("expected 252 input bars, got %d", len(testInput))
	}

	for _, tc := range referenceCases() {
		tc := tc
		t.Run(tc.params.Mnemonic(), func(t *testing.T) {
			t.Parallel()

			params := tc.params
			indicator, err := NewFractalAdaptiveMovingAverage(&params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			warmup := params.Length - 1
			if indicator.WarmupBars() != warmup {
				t.Fatalf("expected warmup %d, got %d", warmup, indicator.WarmupBars())
			}
			if indicator.IsPrimed() {
				t.Fatal("expected unprimed initially")
			}
			if !math.IsNaN(indicator.Value()) || !math.IsNaN(indicator.FractalDimension()) || !math.IsNaN(indicator.Alpha()) {
				t.Fatal("expected initial accessors to return NaN")
			}

			for index, sample := range testInput {
				actualValue := indicator.UpdatePrice(sample)
				actualDimension := indicator.FractalDimension()
				targetValue := tc.expectedValues[index]
				targetDimension := tc.expectedDimensions[index]

				if index < warmup {
					if indicator.IsPrimed() {
						t.Fatalf("bar %d unexpectedly primed", index)
					}
					if !math.IsNaN(actualValue) || !math.IsNaN(actualDimension) {
						t.Fatalf("bar %d expected NaNs, got (%v, %v)", index, actualValue, actualDimension)
					}
					if !math.IsNaN(targetValue) || !math.IsNaN(targetDimension) {
						t.Fatalf("bar %d fixture expected NaNs", index)
					}
					continue
				}

				if !indicator.IsPrimed() {
					t.Fatalf("bar %d expected primed", index)
				}
				valueBound := math.Max(1e-12, math.Abs(targetValue)*1e-12)
				dimBound := math.Max(1e-12, math.Abs(targetDimension)*1e-12)
				if math.Abs(actualValue-targetValue) > valueBound {
					t.Fatalf("bar %d value mismatch: got %.17g, want %.17g", index, actualValue, targetValue)
				}
				if math.Abs(actualDimension-targetDimension) > dimBound {
					t.Fatalf("bar %d dimension mismatch: got %.17g, want %.17g", index, actualDimension, targetDimension)
				}
				if math.Abs(indicator.Value()-actualValue) > 1e-15 {
					t.Fatalf("bar %d Value() accessor mismatch", index)
				}
			}
		})
	}
}

func TestFractalAdaptiveMovingAverageBatchOracleAndMidpointBars(t *testing.T) {
	t.Parallel()

	midpoints := make([]float64, len(testHigh))
	for index := range testHigh {
		midpoints[index] = (testHigh[index] + testLow[index]) / 2.0
	}

	params := Params{Length: 16, SlowestSmoothingFactor: 0.01, CoefficientMode: 0}
	stream, err := NewFractalAdaptiveMovingAverage(&params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedValues, expectedDimensions := articleBatch(midpoints, 16, 0.01, 0, testHigh, testLow)
	for index := range midpoints {
		actualValue := stream.Update(midpoints[index], testHigh[index], testLow[index])
		actualDimension := stream.FractalDimension()
		if math.IsNaN(expectedValues[index]) {
			if !math.IsNaN(actualValue) || !math.IsNaN(actualDimension) {
				t.Fatalf("bar %d expected NaN", index)
			}
		} else {
			if math.Abs(actualValue-expectedValues[index]) > 1e-12 {
				t.Fatalf("bar %d value mismatch: %v vs %v", index, actualValue, expectedValues[index])
			}
			if math.Abs(actualDimension-expectedDimensions[index]) > 1e-12 {
				t.Fatalf("bar %d dimension mismatch: %v vs %v", index, actualDimension, expectedDimensions[index])
			}
		}
	}
}

func TestFractalAdaptiveMovingAverageInvariants(t *testing.T) {
	t.Parallel()

	t.Run("seed_and_first_adaptive_step", func(t *testing.T) {
		t.Parallel()
		stream, err := NewFractalAdaptiveMovingAverageDefault()
		if err != nil {
			t.Fatal(err)
		}
		var seedValue float64
		for index, price := range testInput {
			actualValue := stream.UpdatePrice(price)
			if stream.SamplesSeen() != index+1 {
				t.Fatalf("samples seen mismatch at %d", index)
			}
			if index == 15 {
				if actualValue != price {
					t.Fatalf("expected bar 15 final seed %v, got %v", price, actualValue)
				}
				seedValue = actualValue
			}
			if index == 16 {
				currentAlpha := stream.Alpha()
				expected := currentAlpha*price + (1.0-currentAlpha)*seedValue
				if math.Abs(actualValue-expected) > 1e-15 {
					t.Fatalf("bar 16 adaptive step mismatch: %v vs %v", actualValue, expected)
				}
			}
		}
	})

	t.Run("ramp_gap_constant_flat_half_and_alpha_modes", func(t *testing.T) {
		t.Parallel()
		expectedRampDim := math.Log2(28.0 / 15.0)
		for _, offset := range []float64{100.0, 0.0, -100.0} {
			stream, _ := NewFractalAdaptiveMovingAverageDefault()
			for step := 0; step < 40; step++ {
				sample := offset + float64(step)
				actualValue := stream.UpdatePrice(sample)
				if step >= 15 {
					if math.Abs(stream.FractalDimension()-expectedRampDim) > 1e-13 {
						t.Fatalf("ramp dimension mismatch: %v", stream.FractalDimension())
					}
					if stream.FractalDimension() >= 1.0 || stream.Alpha() != 1.0 || actualValue != sample {
						t.Fatalf("ramp adaptation mismatch at step %d", step)
					}
				}
			}
		}

		gapParams := Params{Length: 4, SlowestSmoothingFactor: 0.01, CoefficientMode: 1}
		gapStream, _ := NewFractalAdaptiveMovingAverage(&gapParams)
		var gapValue float64
		for _, price := range []float64{0.0, 1.0, 100.0, 101.0} {
			gapValue = gapStream.UpdatePrice(price)
		}
		if gapStream.FractalDimension() >= 0.0 || math.Abs(gapStream.FractalDimension()-math.Log2(4.0/101.0)) > 1e-13 {
			t.Fatalf("unexpected gap dimension: %v", gapStream.FractalDimension())
		}
		if gapStream.Alpha() != 1.0 || gapValue != 101.0 {
			t.Fatalf("unexpected gap alpha/value: %v, %v", gapStream.Alpha(), gapValue)
		}

		for _, level := range []float64{-123.0, 0.0, 123.0} {
			flatStream, _ := NewFractalAdaptiveMovingAverage(&gapParams)
			for step := 0; step < 20; step++ {
				actualValue := flatStream.UpdatePrice(level)
				if flatStream.IsPrimed() && (actualValue != level || flatStream.FractalDimension() != 0.0) {
					t.Fatalf("constant guard failed for %v", level)
				}
			}
		}

		holdStream, _ := NewFractalAdaptiveMovingAverage(&gapParams)
		for _, price := range []float64{0.0, 1.0, 0.0, 1.0} {
			holdStream.UpdatePrice(price)
		}
		if holdStream.FractalDimension() != 2.0 {
			t.Fatalf("expected dimension 2.0, got %v", holdStream.FractalDimension())
		}
		for step := 0; step < 10; step++ {
			holdStream.UpdatePrice(1.0)
			if holdStream.FractalDimension() != 2.0 {
				t.Fatalf("expected held dimension 2.0, got %v", holdStream.FractalDimension())
			}
		}

		preciseParams := Params{Length: 4, SlowestSmoothingFactor: 0.01, CoefficientMode: 1}
		articleParams := Params{Length: 4, SlowestSmoothingFactor: 0.01, CoefficientMode: 0}
		precise, _ := NewFractalAdaptiveMovingAverage(&preciseParams)
		article, _ := NewFractalAdaptiveMovingAverage(&articleParams)
		for step := 0; step < 5; step++ {
			precise.Update(1.0, 2.0, 0.0)
			article.Update(1.0, 2.0, 0.0)
		}
		if math.Abs(precise.Alpha()-0.01) > 1e-15 || article.Alpha() != math.Exp(-4.6) || article.Alpha() <= precise.Alpha() {
			t.Fatalf("alpha endpoint mismatch: precise=%v article=%v", precise.Alpha(), article.Alpha())
		}
	})

	t.Run("affine_invariance_nan_gap_and_reset", func(t *testing.T) {
		t.Parallel()
		original, _ := NewFractalAdaptiveMovingAverageDefault()
		transformed, _ := NewFractalAdaptiveMovingAverageDefault()
		withGaps, _ := NewFractalAdaptiveMovingAverageDefault()

		for _, price := range testInput {
			leftVal := original.UpdatePrice(price)
			rightVal := transformed.UpdatePrice(3.0*price - 1000.0)
			gapVal := withGaps.UpdatePrice(price)
			seenBefore := withGaps.SamplesSeen()
			savedVal := withGaps.Value()
			if !math.IsNaN(withGaps.UpdatePrice(math.NaN())) || !math.IsNaN(withGaps.Update(price, math.NaN(), price)) {
				t.Fatal("expected NaN on missing observation")
			}
			if withGaps.SamplesSeen() != seenBefore {
				t.Fatal("NaN advanced SamplesSeen")
			}
			if math.IsNaN(savedVal) != math.IsNaN(withGaps.Value()) || (!math.IsNaN(savedVal) && savedVal != withGaps.Value()) {
				t.Fatal("NaN mutated saved Value")
			}
			if !math.IsNaN(leftVal) {
				if leftVal != gapVal {
					t.Fatalf("NaN gap altered trajectory: %v vs %v", leftVal, gapVal)
				}
				if math.Abs(original.FractalDimension()-transformed.FractalDimension()) > 1e-11 {
					t.Fatal("affine dimension mismatch")
				}
				if math.Abs((3.0*leftVal-1000.0)-rightVal) > 1e-10 {
					t.Fatal("affine value mismatch")
				}
			}
		}

		withGaps.Reset()
		if withGaps.IsPrimed() || withGaps.SamplesSeen() != 0 || !math.IsNaN(withGaps.Value()) {
			t.Fatal("reset did not clear state")
		}
		fresh, _ := NewFractalAdaptiveMovingAverageDefault()
		for _, price := range testInput {
			a := withGaps.UpdatePrice(price)
			b := fresh.UpdatePrice(price)
			if math.IsNaN(b) != math.IsNaN(a) || (!math.IsNaN(b) && a != b) {
				t.Fatal("post-reset output mismatch")
			}
		}
	})
}

func TestFractalAdaptiveMovingAverageValidationMetadataEntitiesAndConcurrency(t *testing.T) {
	t.Parallel()

	invalidParams := []Params{
		{Length: 1, SlowestSmoothingFactor: 0.01, CoefficientMode: 1},
		{Length: 17, SlowestSmoothingFactor: 0.01, CoefficientMode: 1},
		{Length: 16, SlowestSmoothingFactor: 0.0, CoefficientMode: 1},
		{Length: 16, SlowestSmoothingFactor: -0.1, CoefficientMode: 1},
		{Length: 16, SlowestSmoothingFactor: 1.1, CoefficientMode: 1},
		{Length: 16, SlowestSmoothingFactor: math.NaN(), CoefficientMode: 1},
		{Length: 16, SlowestSmoothingFactor: math.Inf(1), CoefficientMode: 1},
		{Length: 16, SlowestSmoothingFactor: 0.01, CoefficientMode: 2},
		{Length: 16, SlowestSmoothingFactor: 0.01, CoefficientMode: -1},
		{Length: 16, SlowestSmoothingFactor: 0.1, CoefficientMode: 0},
		{Length: 16, SlowestSmoothingFactor: 0.01, CoefficientMode: 1, BarComponent: entities.BarComponent(99)},
		{Length: 16, SlowestSmoothingFactor: 0.01, CoefficientMode: 1, QuoteComponent: entities.QuoteComponent(99)},
		{Length: 16, SlowestSmoothingFactor: 0.01, CoefficientMode: 1, TradeComponent: entities.TradeComponent(99)},
	}
	for index, bad := range invalidParams {
		badCopy := bad
		if _, err := NewFractalAdaptiveMovingAverage(&badCopy); err == nil {
			t.Fatalf("case %d: expected error for invalid params %+v", index, badCopy)
		}
	}

	twoBarParams := Params{Length: 2, SlowestSmoothingFactor: 0.01, CoefficientMode: 1}
	atomicStream, err := NewFractalAdaptiveMovingAverage(&twoBarParams)
	if err != nil {
		t.Fatal(err)
	}
	atomicStream.UpdatePrice(0.0)
	badObservations := [][3]float64{
		{math.Inf(1), math.Inf(1), math.Inf(1)},
		{1.0, 0.0, 2.0},
		{1.0, math.Inf(1), 0.0},
		{0.0, 1e308, -1e308},
	}
	for _, obs := range badObservations {
		if _, err := atomicStream.TryUpdate(obs[0], obs[1], obs[2]); err == nil {
			t.Fatalf("expected TryUpdate error for %v", obs)
		}
		if atomicStream.SamplesSeen() != 1 || atomicStream.IsPrimed() {
			t.Fatal("failed TryUpdate mutated state")
		}
	}

	defaultInd, _ := NewFractalAdaptiveMovingAverageDefault()
	meta := defaultInd.Metadata()
	if meta.Identifier != core.FractalAdaptiveMovingAverage {
		t.Fatalf("unexpected identifier: %v", meta.Identifier)
	}
	if meta.Mnemonic != "frama(16, 0.01)" || meta.Description != "Fractal adaptive moving average frama(16, 0.01)" {
		t.Fatalf("unexpected metadata: %+v", meta)
	}
	if len(meta.Outputs) != 2 || meta.Outputs[0].Shape != shape.Scalar || meta.Outputs[1].Mnemonic != "framaDim(16, 0.01)" {
		t.Fatalf("unexpected outputs metadata: %+v", meta.Outputs)
	}

	customParams := Params{
		Length:                 30,
		SlowestSmoothingFactor: 0.01,
		CoefficientMode:        0,
		BarComponent:           entities.BarMedianPrice,
		QuoteComponent:         entities.QuoteBidPrice,
		TradeComponent:         entities.TradeVolume,
	}
	if got := customParams.Mnemonic(); got != "frama(30, 0.01, mode=0, hl/2, b, v)" {
		t.Fatalf("unexpected custom mnemonic: %s", got)
	}
	if got := customParams.DimensionMnemonic(); got != "framaDim(30, 0.01, mode=0, hl/2, b, v)" {
		t.Fatalf("unexpected custom dimension mnemonic: %s", got)
	}

	stamp := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	entityParams := Params{Length: 2, SlowestSmoothingFactor: 0.01, CoefficientMode: 1}
	entityInd, _ := NewFractalAdaptiveMovingAverage(&entityParams)
	entityInd.UpdateBar(&entities.Bar{Time: stamp, Open: 10, High: 12, Low: 9, Close: 11, Volume: 100})
	barOut := entityInd.UpdateBar(&entities.Bar{Time: stamp.Add(time.Hour), Open: 11, High: 14, Low: 10, Close: 13, Volume: 100})
	if len(barOut) != 2 {
		t.Fatalf("expected 2 bar outputs, got %d", len(barOut))
	}
	firstScalar, ok := barOut[0].(entities.Scalar)
	if !ok || firstScalar.Value != 13.0 {
		t.Fatalf("unexpected bar output[0]: %+v", barOut[0])
	}

	nanOut := entityInd.UpdateScalar(&entities.Scalar{Time: stamp.Add(2 * time.Hour), Value: math.NaN()})
	if !math.IsNaN(nanOut[0].(entities.Scalar).Value) || !math.IsNaN(nanOut[1].(entities.Scalar).Value) {
		t.Fatalf("expected NaN entity outputs on skipped sample, got %+v", nanOut)
	}

	quoteInd, _ := NewFractalAdaptiveMovingAverage(&entityParams)
	quoteInd.UpdateQuote(&entities.Quote{Time: stamp, Bid: 9, Ask: 11, BidSize: 1, AskSize: 1})
	quoteOut := quoteInd.UpdateQuote(&entities.Quote{Time: stamp.Add(time.Hour), Bid: 10, Ask: 12, BidSize: 1, AskSize: 1})
	if quoteOut[0].(entities.Scalar).Value != 11.0 {
		t.Fatalf("unexpected quote output: %+v", quoteOut)
	}

	tradeInd, _ := NewFractalAdaptiveMovingAverage(&entityParams)
	tradeInd.UpdateTrade(&entities.Trade{Time: stamp, Price: 5, Volume: 1})
	tradeOut := tradeInd.UpdateTrade(&entities.Trade{Time: stamp.Add(time.Hour), Price: 7, Volume: 1})
	if tradeOut[0].(entities.Scalar).Value != 7.0 {
		t.Fatalf("unexpected trade output: %+v", tradeOut)
	}

	concurrentInd, _ := NewFractalAdaptiveMovingAverageDefault()
	var wg sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		wg.Add(1)
		go func(offset int) {
			defer wg.Done()
			for step := 0; step < 50; step++ {
				price := float64(100 + offset + step)
				_ = concurrentInd.UpdatePrice(price)
				_ = concurrentInd.IsPrimed()
				_ = concurrentInd.Value()
				_ = concurrentInd.FractalDimension()
				_ = concurrentInd.Alpha()
				_ = fmt.Sprint(concurrentInd.Metadata().Mnemonic)
			}
		}(worker)
	}
	wg.Wait()
}
