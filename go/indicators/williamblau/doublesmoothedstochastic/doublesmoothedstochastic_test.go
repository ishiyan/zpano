//nolint:testpackage
package doublesmoothedstochastic

import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
)

const tolerance = 1e-10

type dssCombo struct {
	name      string
	q         int
	r         int
	s         int
	g         int
	expDss    []float64
	expSignal []float64
}

func dssCombos() []dssCombo {
	return []dssCombo{
		{"Q5_R7_S3_G3", 5, 7, 3, 3, expectedDsQ5_R7_S3_G3, expectedSigQ5_R7_S3_G3},
		{"Q2_R3_S15_G3", 2, 3, 15, 3, expectedDsQ2_R3_S15_G3, expectedSigQ2_R3_S15_G3},
		{"Q5_R20_S5_G3", 5, 20, 5, 3, expectedDsQ5_R20_S5_G3, expectedSigQ5_R20_S5_G3},
		{"Q5_R7_S3_G1", 5, 7, 3, 1, expectedDsQ5_R7_S3_G1, expectedSigQ5_R7_S3_G1},
		{"Q2_R3_S15_G1", 2, 3, 15, 1, expectedDsQ2_R3_S15_G1, expectedSigQ2_R3_S15_G1},
		{"Q1_R1_S1_G1", 1, 1, 1, 1, expectedDsQ1_R1_S1_G1, expectedSigQ1_R1_S1_G1},
		{"Q1_R5_S5_G3", 1, 5, 5, 3, expectedDsQ1_R5_S5_G3, expectedSigQ1_R5_S5_G3},
		{"Q8_R5_S3_G3", 8, 5, 3, 3, expectedDsQ8_R5_S3_G3, expectedSigQ8_R5_S3_G3},
		{"Q21_R13_S4_G3", 21, 13, 4, 3, expectedDsQ21_R13_S4_G3, expectedSigQ21_R13_S4_G3},
		{"Q5_R1_S1_G3", 5, 1, 1, 3, expectedDsQ5_R1_S1_G3, expectedSigQ5_R1_S1_G3},
		{"Q3_R10_S10_G5", 3, 10, 10, 5, expectedDsQ3_R10_S10_G5, expectedSigQ3_R10_S10_G5},
		{"Q34_R5_S5_G3", 34, 5, 5, 3, expectedDsQ34_R5_S5_G3, expectedSigQ34_R5_S5_G3},
		{"Q2_R3_S15_G5", 2, 3, 15, 5, expectedDsQ2_R3_S15_G5, expectedSigQ2_R3_S15_G5},
		{"Q10_R7_S3_G3", 10, 7, 3, 3, expectedDsQ10_R7_S3_G3, expectedSigQ10_R7_S3_G3},
	}
}

func checkVal(t *testing.T, name string, i int, exp, act float64) {
	t.Helper()

	if math.IsNaN(exp) {
		if !math.IsNaN(act) {
			t.Errorf("%s[%d]: expected NaN, got %v", name, i, act)
		}

		return
	}

	if math.Abs(act-exp) > tolerance {
		t.Errorf("%s[%d]: expected %v, got %v", name, i, exp, act)
	}
}

func TestDoubleSmoothedStochasticData(t *testing.T) {
	t.Parallel()

	for _, combo := range dssCombos() {
		t.Run(combo.name, func(t *testing.T) {
			t.Parallel()

			dss, err := NewDoubleSmoothedStochastic(&Params{
				Q: combo.q, R: combo.r, S: combo.s, G: combo.g,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for i := range testInput {
				dssValue, signalValue := dss.Update(testHigh[i], testLow[i], testInput[i])

				checkVal(t, "dss", i, combo.expDss[i], dssValue)
				checkVal(t, "signal", i, combo.expSignal[i], signalValue)
			}
		})
	}
}

func TestDoubleSmoothedStochasticPassthrough(t *testing.T) {
	t.Parallel()

	// One-bar HLC index with both EMA stages and the signal SMA as passthroughs,
	// so the oscillator is 100*(close-low)/(high-low) and the signal equals it.
	dss, err := NewDoubleSmoothedStochastic(&Params{Q: 1, R: 1, S: 1, G: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name  string
		high  float64
		low   float64
		close float64
		exp   float64
	}{
		{"close at high", 12.0, 10.0, 12.0, 100.0},
		{"close at low", 12.0, 10.0, 10.0, 0.0},
		{"close at midpoint", 12.0, 10.0, 11.0, 50.0},
		{"flat window", 11.0, 11.0, 11.0, 0.0},
	}

	for i, tt := range tests {
		dssValue, signalValue := dss.Update(tt.high, tt.low, tt.close)

		checkVal(t, tt.name+" dss", i, tt.exp, dssValue)
		checkVal(t, tt.name+" signal", i, tt.exp, signalValue)
	}
}

func TestDoubleSmoothedStochasticSignal(t *testing.T) {
	t.Parallel()

	t.Run("expanding then rolling window", func(t *testing.T) {
		t.Parallel()

		// The signal SMA averages the oscillator values seen so far until g of
		// them exist, then rolls over the last g values.
		dss, err := NewDoubleSmoothedStochastic(&Params{Q: 1, R: 1, S: 1, G: 3})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		tests := []struct {
			close float64
			exp   float64
		}{
			{12.0, 100.0},       // dss 100
			{10.0, 50.0},        // dss 0
			{11.0, 50.0},        // dss 50
			{11.0, 100.0 / 3.0}, // dss 50, 100 dropped
		}

		for i, tt := range tests {
			_, signalValue := dss.Update(12.0, 10.0, tt.close)

			checkVal(t, "signal", i, tt.exp, signalValue)
		}
	})

	t.Run("g = 1 makes signal equal dss", func(t *testing.T) {
		t.Parallel()

		dss, err := NewDoubleSmoothedStochastic(&Params{Q: 5, R: 7, S: 3, G: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for i := range testInput {
			dssValue, signalValue := dss.Update(testHigh[i], testLow[i], testInput[i])

			if math.IsNaN(dssValue) {
				checkVal(t, "signal", i, math.NaN(), signalValue)
			} else if dssValue != signalValue {
				t.Errorf("[%d]: expected signal %v to equal dss %v", i, signalValue, dssValue)
			}
		}
	})
}

func TestDoubleSmoothedStochasticIsPrimed(t *testing.T) {
	t.Parallel()

	t.Run("default q", func(t *testing.T) {
		t.Parallel()

		dss, err := NewDoubleSmoothedStochastic(DefaultParams())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if dss.IsPrimed() {
			t.Errorf("IsPrimed: expected false before the first update")
		}

		// Bars 0..q-2 are the NaN warm-up region.
		const q = 5

		for i := range q - 1 {
			dssValue, signalValue := dss.Update(testHigh[i], testLow[i], testInput[i])

			if dss.IsPrimed() {
				t.Errorf("[%d] IsPrimed: expected false", i)
			}

			checkVal(t, "dss", i, math.NaN(), dssValue)
			checkVal(t, "signal", i, math.NaN(), signalValue)
		}

		for i := q - 1; i < len(testInput); i++ {
			dss.Update(testHigh[i], testLow[i], testInput[i])

			if !dss.IsPrimed() {
				t.Errorf("[%d] IsPrimed: expected true", i)
			}
		}
	})

	t.Run("q = 1 has no warm-up", func(t *testing.T) {
		t.Parallel()

		dss, err := NewDoubleSmoothedStochastic(&Params{Q: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if dss.IsPrimed() {
			t.Errorf("IsPrimed: expected false before the first update")
		}

		dss.Update(testHigh[0], testLow[0], testInput[0])

		if !dss.IsPrimed() {
			t.Errorf("IsPrimed: expected true after the first update")
		}
	})
}

func TestDoubleSmoothedStochasticUpdateBar(t *testing.T) {
	t.Parallel()

	dss, err := NewDoubleSmoothedStochastic(&Params{Q: 5, R: 7, S: 3, G: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tm := time.Now()

	var output core.Output

	for i := range testInput {
		output = dss.UpdateBar(&entities.Bar{
			Time: tm, High: testHigh[i], Low: testLow[i], Close: testInput[i],
		})
	}

	if len(output) != 2 {
		t.Fatalf("outputs: expected 2, got %d", len(output))
	}

	last := len(testInput) - 1

	if act := output[0].(entities.Scalar).Time; !act.Equal(tm) {
		t.Errorf("time: expected %v, got %v", tm, act)
	}

	checkVal(t, "dss", last, expectedDsQ5_R7_S3_G3[last], output[0].(entities.Scalar).Value)
	checkVal(t, "signal", last, expectedSigQ5_R7_S3_G3[last], output[1].(entities.Scalar).Value)
}

func TestDoubleSmoothedStochasticUpdateEntities(t *testing.T) {
	t.Parallel()

	tm := time.Now()

	t.Run("scalar is used as high, low and close", func(t *testing.T) {
		t.Parallel()

		dss, err := NewDoubleSmoothedStochastic(&Params{Q: 2, R: 1, S: 1, G: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		output := dss.UpdateScalar(&entities.Scalar{Time: tm, Value: 10.0})

		checkVal(t, "dss", 0, math.NaN(), output[0].(entities.Scalar).Value)
		checkVal(t, "signal", 0, math.NaN(), output[1].(entities.Scalar).Value)

		// Close at the 2-bar high.
		output = dss.UpdateScalar(&entities.Scalar{Time: tm, Value: 12.0})

		checkVal(t, "dss", 1, 100.0, output[0].(entities.Scalar).Value)
		checkVal(t, "signal", 1, 100.0, output[1].(entities.Scalar).Value)
	})

	t.Run("quote mid price is used as high, low and close", func(t *testing.T) {
		t.Parallel()

		dss, err := NewDoubleSmoothedStochastic(&Params{Q: 2, R: 1, S: 1, G: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		dss.UpdateQuote(&entities.Quote{Time: tm, Bid: 12.0, Ask: 14.0})

		// Mid 11 is at the 2-bar low.
		output := dss.UpdateQuote(&entities.Quote{Time: tm, Bid: 10.0, Ask: 12.0})

		checkVal(t, "dss", 1, 0.0, output[0].(entities.Scalar).Value)
		checkVal(t, "signal", 1, 0.0, output[1].(entities.Scalar).Value)
	})

	t.Run("trade price is used as high, low and close", func(t *testing.T) {
		t.Parallel()

		dss, err := NewDoubleSmoothedStochastic(&Params{Q: 2, R: 1, S: 1, G: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		dss.UpdateTrade(&entities.Trade{Time: tm, Price: 10.0, Volume: 1.0})
		output := dss.UpdateTrade(&entities.Trade{Time: tm, Price: 12.0, Volume: 1.0})

		checkVal(t, "dss", 1, 100.0, output[0].(entities.Scalar).Value)
		checkVal(t, "signal", 1, 100.0, output[1].(entities.Scalar).Value)
	})
}

func TestDoubleSmoothedStochasticMnemonic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params *Params
		exp    string
	}{
		{"all parameters zero", &Params{}, "dss(5,7,3,3)"},
		{"default parameters", DefaultParams(), "dss(5,7,3,3)"},
		{"book alternative", &Params{Q: 2, R: 3, S: 15, G: 3}, "dss(2,3,15,3)"},
		{"passthrough signal", &Params{Q: 5, R: 7, S: 3, G: 1}, "dss(5,7,3,1)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dss, err := NewDoubleSmoothedStochastic(tt.params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if dss.mnemonic != tt.exp {
				t.Errorf("mnemonic: expected '%s', got '%s'", tt.exp, dss.mnemonic)
			}

			expDesc := "Double Smoothed Stochastic " + tt.exp
			if meta := dss.Metadata(); meta.Description != expDesc {
				t.Errorf("description: expected '%s', got '%s'", expDesc, meta.Description)
			}
		})
	}
}

func TestDoubleSmoothedStochasticMetadata(t *testing.T) {
	t.Parallel()

	dss, err := NewDoubleSmoothedStochastic(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	meta := dss.Metadata()

	if meta.Identifier != core.DoubleSmoothedStochastic {
		t.Errorf("identifier: expected DoubleSmoothedStochastic, got %v", meta.Identifier)
	}

	if meta.Mnemonic != "dss(5,7,3,3)" {
		t.Errorf("mnemonic: expected 'dss(5,7,3,3)', got '%s'", meta.Mnemonic)
	}

	if len(meta.Outputs) != 2 {
		t.Fatalf("outputs: expected 2, got %d", len(meta.Outputs))
	}

	if act := meta.Outputs[0].Mnemonic; act != "dss(5,7,3,3) dss" {
		t.Errorf("outputs[0].Mnemonic: expected 'dss(5,7,3,3) dss', got '%s'", act)
	}

	if act := meta.Outputs[1].Mnemonic; act != "dss(5,7,3,3) signal" {
		t.Errorf("outputs[1].Mnemonic: expected 'dss(5,7,3,3) signal', got '%s'", act)
	}
}

func TestDoubleSmoothedStochasticInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params *Params
	}{
		{"q too small", &Params{Q: -1, R: 7, S: 3, G: 3}},
		{"r too small", &Params{Q: 5, R: -1, S: 3, G: 3}},
		{"s too small", &Params{Q: 5, R: 7, S: -1, G: 3}},
		{"g too small", &Params{Q: 5, R: 7, S: 3, G: -1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := NewDoubleSmoothedStochastic(tt.params); err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}
