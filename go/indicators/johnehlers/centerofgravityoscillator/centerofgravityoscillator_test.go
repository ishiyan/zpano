//nolint:testpackage
package centerofgravityoscillator

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
	length          int
	centered        bool
	expectedValue   []float64
	expectedTrigger []float64
}

func referenceCases() []referenceCase {
	return []referenceCase{
		{1, false, expectedL1_C0_CG, expectedL1_C0_TR},
		{1, true, expectedL1_C1_CG, expectedL1_C1_TR},
		{2, false, expectedL2_C0_CG, expectedL2_C0_TR},
		{2, true, expectedL2_C1_CG, expectedL2_C1_TR},
		{3, false, expectedL3_C0_CG, expectedL3_C0_TR},
		{3, true, expectedL3_C1_CG, expectedL3_C1_TR},
		{5, false, expectedL5_C0_CG, expectedL5_C0_TR},
		{5, true, expectedL5_C1_CG, expectedL5_C1_TR},
		{8, false, expectedL8_C0_CG, expectedL8_C0_TR},
		{8, true, expectedL8_C1_CG, expectedL8_C1_TR},
		{10, false, expectedL10_C0_CG, expectedL10_C0_TR},
		{10, true, expectedL10_C1_CG, expectedL10_C1_TR},
		{14, false, expectedL14_C0_CG, expectedL14_C0_TR},
		{14, true, expectedL14_C1_CG, expectedL14_C1_TR},
		{20, false, expectedL20_C0_CG, expectedL20_C0_TR},
		{20, true, expectedL20_C1_CG, expectedL20_C1_TR},
		{32, false, expectedL32_C0_CG, expectedL32_C0_TR},
		{32, true, expectedL32_C1_CG, expectedL32_C1_TR},
		{64, false, expectedL64_C0_CG, expectedL64_C0_TR},
		{64, true, expectedL64_C1_CG, expectedL64_C1_TR},
		{126, false, expectedL126_C0_CG, expectedL126_C0_TR},
		{126, true, expectedL126_C1_CG, expectedL126_C1_TR},
		{252, false, expectedL252_C0_CG, expectedL252_C0_TR},
		{252, true, expectedL252_C1_CG, expectedL252_C1_TR},
	}
}

func publishedLoopOutputs(prices []float64, length int, centered bool) [][2]float64 {
	previous := 0.0
	outputs := make([][2]float64, len(prices))
	for step := range prices {
		if step < length-1 {
			outputs[step] = [2]float64{math.NaN(), math.NaN()}
			continue
		}
		trigger := math.NaN()
		if step >= length {
			trigger = previous
		}
		numerator := 0.0
		denominator := 0.0
		for lag := 0; lag < length; lag++ {
			numerator += float64(1+lag) * prices[step-lag]
			denominator += prices[step-lag]
		}
		if denominator != 0.0 {
			previous = -numerator / denominator
			if centered {
				previous += float64(length+1) / 2.0
			}
		}
		outputs[step] = [2]float64{previous, trigger}
	}
	return outputs
}

func assertCloseOrNaN(t *testing.T, label string, expected, actual, tolerance float64) {
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
		t.Fatalf("%s: expected %.17g, got %.17g (diff=%.3e)", label, expected, actual, math.Abs(actual-expected))
	}
}

func TestCenterOfGravityOscillatorReferenceCases(t *testing.T) {
	t.Parallel()

	if len(testInput) != 252 {
		t.Fatalf("expected 252 input prices, got %d", len(testInput))
	}

	for _, tc := range referenceCases() {
		tc := tc
		params := Params{Length: tc.length, Centered: tc.centered}
		t.Run(params.Mnemonic(), func(t *testing.T) {
			t.Parallel()

			indicator, err := NewCenterOfGravityOscillator(&params)
			if err != nil {
				t.Fatalf("unexpected constructor error: %v", err)
			}
			if indicator.Length() != tc.length {
				t.Fatalf("expected Length()=%d, got %d", tc.length, indicator.Length())
			}
			if indicator.Centered() != tc.centered {
				t.Fatalf("expected Centered()=%v, got %v", tc.centered, indicator.Centered())
			}
			if indicator.IsPrimed() || indicator.IsTriggerPrimed() {
				t.Fatal("expected unprimed indicator at startup")
			}
			if !math.IsNaN(indicator.Value()) || !math.IsNaN(indicator.Trigger()) {
				t.Fatal("expected NaN accessors at startup")
			}

			oracle := publishedLoopOutputs(testInput, tc.length, tc.centered)
			barIndicator, err := NewCenterOfGravityOscillator(&params)
			if err != nil {
				t.Fatalf("unexpected bar constructor error: %v", err)
			}
			baseTime := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)

			for index, price := range testInput {
				actualValue := indicator.Update(price)
				actualTrigger := indicator.Trigger()

				expectedPrimed := index >= tc.length-1
				expectedTriggerPrimed := index >= tc.length
				if indicator.IsPrimed() != expectedPrimed {
					t.Fatalf("bar %d: IsPrimed()=%v, want %v", index, indicator.IsPrimed(), expectedPrimed)
				}
				if indicator.IsTriggerPrimed() != expectedTriggerPrimed {
					t.Fatalf("bar %d: IsTriggerPrimed()=%v, want %v", index, indicator.IsTriggerPrimed(), expectedTriggerPrimed)
				}

				assertCloseOrNaN(t, fmt.Sprintf("bar %d value", index), tc.expectedValue[index], actualValue, 1e-13)
				assertCloseOrNaN(t, fmt.Sprintf("bar %d trigger", index), tc.expectedTrigger[index], actualTrigger, 1e-13)
				assertCloseOrNaN(t, fmt.Sprintf("bar %d Value()", index), actualValue, indicator.Value(), 0.0)

				// Exact bit-for-bit agreement with the published EasyLanguage loop.
				assertCloseOrNaN(t, fmt.Sprintf("bar %d oracle value", index), oracle[index][0], actualValue, 0.0)
				assertCloseOrNaN(t, fmt.Sprintf("bar %d oracle trigger", index), oracle[index][1], actualTrigger, 0.0)

				stamp := baseTime.AddDate(0, 0, index)
				barOutput := barIndicator.UpdateBar(&entities.Bar{
					Time:   stamp,
					Open:   testOpen[index],
					High:   testHigh[index],
					Low:    testLow[index],
					Close:  price,
					Volume: 1000.0,
				})
				if len(barOutput) != 2 {
					t.Fatalf("bar %d: expected 2 outputs, got %d", index, len(barOutput))
				}
				outValue := barOutput[0].(entities.Scalar)
				outTrigger := barOutput[1].(entities.Scalar)
				if !outValue.Time.Equal(stamp) || !outTrigger.Time.Equal(stamp) {
					t.Fatalf("bar %d: timestamp mismatch", index)
				}
				assertCloseOrNaN(t, fmt.Sprintf("bar %d entity value", index), actualValue, outValue.Value, 0.0)
				assertCloseOrNaN(t, fmt.Sprintf("bar %d entity trigger", index), actualTrigger, outTrigger.Value, 0.0)
			}
		})
	}
}

func TestCenterOfGravityOscillatorMathematicalProperties(t *testing.T) {
	t.Parallel()

	t.Run("newest price has weight one and sign is negative", func(t *testing.T) {
		t.Parallel()

		article, _ := NewCenterOfGravityOscillator(&Params{Length: 3, Centered: false})
		book, _ := NewCenterOfGravityOscillator(&Params{Length: 3, Centered: true})

		var articleValue, bookValue float64
		for _, price := range []float64{10.0, 20.0, 30.0} {
			articleValue = article.Update(price)
			bookValue = book.Update(price)
		}
		// Numerator = 1*30 + 2*20 + 3*10 = 100, Denominator = 60 => -5/3 and +1/3.
		assertCloseOrNaN(t, "article [10,20,30]", -5.0/3.0, articleValue, 1e-15)
		assertCloseOrNaN(t, "book [10,20,30]", 1.0/3.0, bookValue, 1e-15)
		if !math.IsNaN(article.Trigger()) || !math.IsNaN(book.Trigger()) {
			t.Fatal("trigger must still be NaN on first full window")
		}

		// Next sample 60 => window [20, 30, 60]: Num = 1*60 + 2*30 + 3*20 = 180, Denom = 110 => -18/11? Wait:
		// 1*60 + 2*30 + 3*20 = 60 + 60 + 60 = 180, Denom = 110 => -18/11.
		nextArticle := article.Update(60.0)
		assertCloseOrNaN(t, "article [20,30,60] value", -18.0/11.0, nextArticle, 1e-15)
		assertCloseOrNaN(t, "article [20,30,60] trigger", -5.0/3.0, article.Trigger(), 1e-15)
	})

	t.Run("constant positive price, length one boundary, and positive scale invariance", func(t *testing.T) {
		t.Parallel()

		for _, length := range []int{1, 2, 3, 10} {
			for _, centered := range []bool{false, true} {
				indicator, _ := NewCenterOfGravityOscillator(&Params{Length: length, Centered: centered})
				baseline := -(float64(length+1) / 2.0)
				if centered {
					baseline = 0.0
				}
				for step := 0; step < length+2; step++ {
					actualValue := indicator.Update(8.0)
					expectedValue := baseline
					if step < length-1 {
						expectedValue = math.NaN()
					}
					expectedTrigger := baseline
					if step < length {
						expectedTrigger = math.NaN()
					}
					assertCloseOrNaN(t, "constant value", expectedValue, actualValue, 1e-14)
					assertCloseOrNaN(t, "constant trigger", expectedTrigger, indicator.Trigger(), 1e-14)
				}
			}
		}

		base, _ := NewCenterOfGravityOscillator(&Params{Length: 5, Centered: true})
		scaled, _ := NewCenterOfGravityOscillator(&Params{Length: 5, Centered: true})
		for index := 0; index < 25; index++ {
			price := testInput[index]
			baseOut := base.Update(price)
			scaledOut := scaled.Update(3.75 * price)
			assertCloseOrNaN(t, "scale invariance value", baseOut, scaledOut, 1e-13)
			assertCloseOrNaN(t, "scale invariance trigger", base.Trigger(), scaled.Trigger(), 1e-13)
		}
	})

	t.Run("monotone rising vs falling windows and trigger crossover", func(t *testing.T) {
		t.Parallel()

		rising, _ := NewCenterOfGravityOscillator(&Params{Length: 4, Centered: true})
		falling, _ := NewCenterOfGravityOscillator(&Params{Length: 4, Centered: true})
		for step := 0; step < 4; step++ {
			rising.Update(float64(10 + step))
			falling.Update(float64(13 - step))
		}
		if !(rising.Value() > 0.0 && falling.Value() < 0.0) {
			t.Fatalf("expected rising > 0 and falling < 0 in centered mode, got %v and %v", rising.Value(), falling.Value())
		}
		// Reversing direction causes value to cross below trigger.
		rising.Update(14.0)
		if !(rising.Value() < rising.Trigger() || math.Abs(rising.Value()-rising.Trigger()) < 1.0) {
			// Drop sharply so newest bar has low mass and balance point shifts toward older bars (more negative).
		}
		afterDrop := rising.Update(1.0)
		if !(afterDrop < rising.Trigger()) {
			t.Fatalf("expected sharp drop to move CG below trigger: value=%v, trigger=%v", afterDrop, rising.Trigger())
		}
	})
}

func TestCenterOfGravityOscillatorEdgeCasesAndState(t *testing.T) {
	t.Parallel()

	t.Run("zero denominator holds previous value and initial zero has no centered offset", func(t *testing.T) {
		t.Parallel()

		for _, centered := range []bool{false, true} {
			initialZero, _ := NewCenterOfGravityOscillator(&Params{Length: 2, Centered: centered})
			initialZero.Update(1.0)
			first := initialZero.Update(-1.0)
			assertCloseOrNaN(t, "initial zero value", 0.0, first, 0.0)
			if !math.IsNaN(initialZero.Trigger()) {
				t.Fatal("expected NaN trigger on first full window")
			}

			subsequentZero, _ := NewCenterOfGravityOscillator(&Params{Length: 2, Centered: centered})
			subsequentZero.Update(1.0)
			previous := subsequentZero.Update(2.0)
			held := subsequentZero.Update(-2.0)
			assertCloseOrNaN(t, "held value", previous, held, 0.0)
			assertCloseOrNaN(t, "held trigger", previous, subsequentZero.Trigger(), 0.0)

			resumed := subsequentZero.Update(4.0)
			expectedResumed := 0.0
			if centered {
				expectedResumed = 1.5
			}
			assertCloseOrNaN(t, "resumed value", expectedResumed, resumed, 0.0)
			assertCloseOrNaN(t, "resumed trigger", previous, subsequentZero.Trigger(), 0.0)
		}
	})

	t.Run("subnormal denominator uses exact nonzero guard", func(t *testing.T) {
		t.Parallel()

		equalSubnormal, _ := NewCenterOfGravityOscillator(&Params{Length: 2})
		equalSubnormal.Update(1e-309)
		assertCloseOrNaN(t, "equal subnormal", -1.5, equalSubnormal.Update(1e-309), 1e-14)

		unequalSubnormal, _ := NewCenterOfGravityOscillator(&Params{Length: 2})
		unequalSubnormal.Update(1e-309)
		assertCloseOrNaN(t, "unequal subnormal", -1.25, unequalSubnormal.Update(3e-309), 1e-14)
	})

	t.Run("NaN skips without advancing window or clearing accessors", func(t *testing.T) {
		t.Parallel()

		skipped, _ := NewCenterOfGravityOscillator(&Params{Length: 3})
		uninterrupted, _ := NewCenterOfGravityOscillator(&Params{Length: 3})
		baseTime := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

		for step, price := range []float64{1.0, 2.0, 3.0, 4.0, 5.0} {
			if !math.IsNaN(skipped.Update(math.NaN())) {
				t.Fatal("expected NaN scalar return on NaN input")
			}
			gapEntity := skipped.UpdateScalar(&entities.Scalar{Time: baseTime.AddDate(0, 0, step), Value: math.NaN()})
			if !math.IsNaN(gapEntity[0].(entities.Scalar).Value) || !math.IsNaN(gapEntity[1].(entities.Scalar).Value) {
				t.Fatal("expected [NaN, NaN] entity output on NaN input")
			}
			actual := skipped.Update(price)
			reference := uninterrupted.Update(price)
			assertCloseOrNaN(t, "skipped vs uninterrupted value", reference, actual, 0.0)
			assertCloseOrNaN(t, "skipped vs uninterrupted trigger", uninterrupted.Trigger(), skipped.Trigger(), 0.0)
		}
		if !skipped.IsTriggerPrimed() {
			t.Fatal("expected IsTriggerPrimed() after 5 finite observations")
		}
	})

	t.Run("infinity is rejected before mutating state and Update panics", func(t *testing.T) {
		t.Parallel()

		indicator, _ := NewCenterOfGravityOscillator(&Params{Length: 1})
		indicator.Update(4.0)

		for _, invalid := range []float64{math.Inf(1), math.Inf(-1)} {
			result, err := indicator.TryUpdate(invalid)
			if err == nil || !math.IsNaN(result) {
				t.Fatalf("expected error and NaN for %v, got (%v, %v)", invalid, result, err)
			}

			didPanic := false
			func() {
				defer func() {
					if recover() != nil {
						didPanic = true
					}
				}()
				indicator.Update(invalid)
			}()
			if !didPanic {
				t.Fatalf("expected Update(%v) to panic", invalid)
			}
		}

		next := indicator.Update(5.0)
		assertCloseOrNaN(t, "post-inf value", -1.0, next, 0.0)
		assertCloseOrNaN(t, "post-inf trigger", -1.0, indicator.Trigger(), 0.0)
	})

	t.Run("Reset restores warmup and persistent zero", func(t *testing.T) {
		t.Parallel()

		indicator, _ := NewCenterOfGravityOscillator(&Params{Length: 2, Centered: true})
		for _, price := range []float64{1.0, 2.0, 3.0} {
			indicator.Update(price)
		}
		indicator.Reset()
		if indicator.IsPrimed() || indicator.IsTriggerPrimed() {
			t.Fatal("expected Reset() to clear priming flags")
		}
		if !math.IsNaN(indicator.Value()) || !math.IsNaN(indicator.Trigger()) {
			t.Fatal("expected Reset() to restore NaN accessors")
		}
		if !math.IsNaN(indicator.Update(1.0)) {
			t.Fatal("expected NaN on first post-reset sample")
		}
		assertCloseOrNaN(t, "post-reset zero sum", 0.0, indicator.Update(-1.0), 0.0)
	})
}

func TestCenterOfGravityOscillatorValidationMetadataEntitiesAndConcurrency(t *testing.T) {
	t.Parallel()

	t.Run("parameter validation", func(t *testing.T) {
		t.Parallel()

		if _, err := NewCenterOfGravityOscillator(nil); err == nil {
			t.Fatal("expected error for nil params")
		}
		for _, badLength := range []int{0, -1, -10} {
			if _, err := NewCenterOfGravityOscillator(&Params{Length: badLength}); err == nil {
				t.Fatalf("expected error for length=%d", badLength)
			}
		}
		if _, err := NewCenterOfGravityOscillator(&Params{Length: 10, BarComponent: entities.BarComponent(9999)}); err == nil {
			t.Fatal("expected error for invalid bar component")
		}
		if _, err := NewCenterOfGravityOscillator(&Params{Length: 10, QuoteComponent: entities.QuoteComponent(9999)}); err == nil {
			t.Fatal("expected error for invalid quote component")
		}
		if _, err := NewCenterOfGravityOscillator(&Params{Length: 10, TradeComponent: entities.TradeComponent(9999)}); err == nil {
			t.Fatal("expected error for invalid trade component")
		}
	})

	t.Run("mnemonics and metadata", func(t *testing.T) {
		t.Parallel()

		defaultIndicator, err := NewCenterOfGravityOscillatorDefault()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		meta := defaultIndicator.Metadata()
		if meta.Identifier != core.CenterOfGravityOscillator {
			t.Fatalf("unexpected identifier: %v", meta.Identifier)
		}
		if meta.Mnemonic != "cog(10)" || meta.Description != "Center of Gravity oscillator cog(10)" {
			t.Fatalf("unexpected default metadata: %+v", meta)
		}
		if len(meta.Outputs) != 2 {
			t.Fatalf("expected 2 outputs, got %d", len(meta.Outputs))
		}
		if meta.Outputs[0].Kind != int(Value) || meta.Outputs[0].Shape != shape.Scalar || meta.Outputs[0].Mnemonic != "cog(10)" {
			t.Fatalf("unexpected output[0]: %+v", meta.Outputs[0])
		}
		if meta.Outputs[1].Kind != int(Trigger) || meta.Outputs[1].Shape != shape.Scalar || meta.Outputs[1].Mnemonic != "cogTrig(10)" || meta.Outputs[1].Description != "Center of Gravity trigger cogTrig(10)" {
			t.Fatalf("unexpected output[1]: %+v", meta.Outputs[1])
		}

		centeredOnly := Params{Length: 8, Centered: true}
		if centeredOnly.Mnemonic() != "cog(8, centered)" || centeredOnly.TriggerMnemonic() != "cogTrig(8, centered)" {
			t.Fatalf("unexpected centered mnemonics: %s / %s", centeredOnly.Mnemonic(), centeredOnly.TriggerMnemonic())
		}

		custom := Params{
			Length:         14,
			Centered:       true,
			BarComponent:   entities.BarMedianPrice,
			QuoteComponent: entities.QuoteBidPrice,
			TradeComponent: entities.TradeVolume,
		}
		if custom.Mnemonic() != "cog(14, centered, hl/2, b, v)" || custom.TriggerMnemonic() != "cogTrig(14, centered, hl/2, b, v)" {
			t.Fatalf("unexpected custom mnemonics: %s / %s", custom.Mnemonic(), custom.TriggerMnemonic())
		}
	})

	t.Run("entity adapters for Bar Median, Quote, Trade, and Scalar", func(t *testing.T) {
		t.Parallel()

		baseTime := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
		barMedian, _ := NewCenterOfGravityOscillator(&Params{Length: 4, Centered: true, BarComponent: entities.BarMedianPrice})
		quoteMid, _ := NewCenterOfGravityOscillator(&Params{Length: 4, Centered: true})
		tradePrice, _ := NewCenterOfGravityOscillator(&Params{Length: 4, Centered: true})
		scalarDirect, _ := NewCenterOfGravityOscillator(&Params{Length: 4, Centered: true})

		for step := 0; step < 12; step++ {
			stamp := baseTime.AddDate(0, 0, step)
			median := (testHigh[step] + testLow[step]) / 2.0
			barOut := barMedian.UpdateBar(&entities.Bar{Time: stamp, Open: testOpen[step], High: testHigh[step], Low: testLow[step], Close: testInput[step]})
			quoteOut := quoteMid.UpdateQuote(&entities.Quote{Time: stamp, Bid: testLow[step], Ask: testHigh[step]})
			tradeOut := tradePrice.UpdateTrade(&entities.Trade{Time: stamp, Price: median})
			scalarOut := scalarDirect.UpdateScalar(&entities.Scalar{Time: stamp, Value: median})

			for slot := 0; slot < 2; slot++ {
				expected := scalarOut[slot].(entities.Scalar).Value
				assertCloseOrNaN(t, "bar vs scalar", expected, barOut[slot].(entities.Scalar).Value, 1e-15)
				assertCloseOrNaN(t, "quote vs scalar", expected, quoteOut[slot].(entities.Scalar).Value, 1e-15)
				assertCloseOrNaN(t, "trade vs scalar", expected, tradeOut[slot].(entities.Scalar).Value, 1e-15)
			}
		}
	})

	t.Run("concurrent goroutine access", func(t *testing.T) {
		t.Parallel()

		indicator, _ := NewCenterOfGravityOscillatorDefault()
		var group sync.WaitGroup
		for worker := 0; worker < 4; worker++ {
			group.Add(1)
			go func() {
				defer group.Done()
				for index := 0; index < 50; index++ {
					indicator.Update(testInput[index])
					_ = indicator.Value()
					_ = indicator.Trigger()
					_ = indicator.IsPrimed()
					_ = indicator.IsTriggerPrimed()
				}
			}()
		}
		group.Wait()
	})
}
