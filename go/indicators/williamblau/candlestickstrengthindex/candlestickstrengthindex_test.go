//nolint:testpackage
package candlestickstrengthindex

import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
)

const tolerance = 1e-9

// ul is the signal-line EMA period used for every expected signal array.
const ul = 3

type csiCombo struct {
	name      string
	r         int
	s         int
	u         int
	expCsi    []float64
	expSignal []float64
}

func csiCombos() []csiCombo {
	return []csiCombo{
		{"R20_S5_U3", 20, 5, 3, expectedR20_S5_U3, expectedR20_S5_U3_SIG_UL3},
		{"R32_S32_U1", 32, 32, 1, expectedR32_S32_U1, expectedR32_S32_U1_SIG_UL3},
		{"R1_S1_U1", 1, 1, 1, expectedR1_S1_U1, expectedR1_S1_U1_SIG_UL3},
		{"R25_S13_U1", 25, 13, 1, expectedR25_S13_U1, expectedR25_S13_U1_SIG_UL3},
		{"R13_S13_U1", 13, 13, 1, expectedR13_S13_U1, expectedR13_S13_U1_SIG_UL3},
		{"R5_S5_U5", 5, 5, 5, expectedR5_S5_U5, expectedR5_S5_U5_SIG_UL3},
		{"R9_S3_U1", 9, 3, 1, expectedR9_S3_U1, expectedR9_S3_U1_SIG_UL3},
		{"R64_S64_U1", 64, 64, 1, expectedR64_S64_U1, expectedR64_S64_U1_SIG_UL3},
		{"R32_S32_U3", 32, 32, 3, expectedR32_S32_U3, expectedR32_S32_U3_SIG_UL3},
		{"R40_S20_U1", 40, 20, 1, expectedR40_S20_U1, expectedR40_S20_U1_SIG_UL3},
		{"R2_S2_U2", 2, 2, 2, expectedR2_S2_U2, expectedR2_S2_U2_SIG_UL3},
		{"R7_S4_U2", 7, 4, 2, expectedR7_S4_U2, expectedR7_S4_U2_SIG_UL3},
		{"R12_S12_U12", 12, 12, 12, expectedR12_S12_U12, expectedR12_S12_U12_SIG_UL3},
		{"R3_S10_U10", 3, 10, 10, expectedR3_S10_U10, expectedR3_S10_U10_SIG_UL3},
		{"R50_S1_U1", 50, 1, 1, expectedR50_S1_U1, expectedR50_S1_U1_SIG_UL3},
		{"R32_S5_U3", 32, 5, 3, expectedR32_S5_U3, expectedR32_S5_U3_SIG_UL3},
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

func TestCandlestickStrengthIndexData(t *testing.T) {
	t.Parallel()

	for _, combo := range csiCombos() {
		combo := combo
		t.Run(combo.name, func(t *testing.T) {
			t.Parallel()

			csi, err := NewCandlestickStrengthIndex(&Params{
				R: combo.r, S: combo.s, U: combo.u, Ul: ul,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for i := 0; i < len(testInput); i++ {
				csiVal, signalVal := csi.Update(testOpen[i], testHigh[i], testLow[i], testInput[i])

				checkVal(t, "csi", i, combo.expCsi[i], csiVal)
				checkVal(t, "signal", i, combo.expSignal[i], signalVal)
			}
		})
	}
}

func TestCandlestickStrengthIndexPassthrough(t *testing.T) {
	t.Parallel()

	// All EMA stages and the signal EMA are passthroughs, so the oscillator is
	// 100*(close-open)/(high-low) and the signal equals it.
	csi, err := NewCandlestickStrengthIndex(&Params{R: 1, S: 1, U: 1, Ul: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name  string
		open  float64
		high  float64
		low   float64
		close float64
		exp   float64
	}{
		{"body spans the whole range up", 10.0, 12.0, 10.0, 12.0, 100.0},
		{"body spans the whole range down", 12.0, 12.0, 10.0, 10.0, -100.0},
		{"zero range", 11.0, 11.0, 11.0, 11.0, 0.0},
	}

	for i, tt := range tests {
		csiVal, signalVal := csi.Update(tt.open, tt.high, tt.low, tt.close)

		checkVal(t, tt.name+" csi", i, tt.exp, csiVal)
		checkVal(t, tt.name+" signal", i, tt.exp, signalVal)
	}
}

func TestCandlestickStrengthIndexIsPrimed(t *testing.T) {
	t.Parallel()

	// Both intra-bar series are defined from bar 0, so there is no NaN warm-up.
	csi, err := NewCandlestickStrengthIndex(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if csi.IsPrimed() {
		t.Errorf("IsPrimed: expected false before the first update")
	}

	csi.Update(testOpen[0], testHigh[0], testLow[0], testInput[0])

	if !csi.IsPrimed() {
		t.Errorf("IsPrimed: expected true after the first update")
	}
}

func TestCandlestickStrengthIndexUpdateBar(t *testing.T) {
	t.Parallel()

	csi, err := NewCandlestickStrengthIndex(&Params{R: 20, S: 5, U: 3, Ul: ul})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tm := time.Now()

	var output core.Output

	for i := 0; i < len(testInput); i++ {
		output = csi.UpdateBar(&entities.Bar{
			Time: tm, Open: testOpen[i], High: testHigh[i],
			Low: testLow[i], Close: testInput[i],
		})
	}

	if len(output) != 2 {
		t.Fatalf("outputs: expected 2, got %d", len(output))
	}

	last := len(testInput) - 1

	checkVal(t, "csi", last, expectedR20_S5_U3[last], output[0].(entities.Scalar).Value)
	checkVal(t, "signal", last, expectedR20_S5_U3_SIG_UL3[last], output[1].(entities.Scalar).Value)
}

func TestCandlestickStrengthIndexUpdateEntities(t *testing.T) {
	t.Parallel()

	tm := time.Now()

	t.Run("quote maps bid to open and low, ask to close and high", func(t *testing.T) {
		t.Parallel()

		csi, err := NewCandlestickStrengthIndex(&Params{R: 1, S: 1, U: 1, Ul: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		output := csi.UpdateQuote(&entities.Quote{Time: tm, Bid: 10.0, Ask: 12.0})

		checkVal(t, "csi", 0, 100.0, output[0].(entities.Scalar).Value)
		checkVal(t, "signal", 0, 100.0, output[1].(entities.Scalar).Value)
	})

	t.Run("scalar has a zero candle body and a zero range", func(t *testing.T) {
		t.Parallel()

		csi, err := NewCandlestickStrengthIndex(&Params{R: 1, S: 1, U: 1, Ul: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		output := csi.UpdateScalar(&entities.Scalar{Time: tm, Value: 10.0})

		checkVal(t, "csi", 0, 0.0, output[0].(entities.Scalar).Value)
		checkVal(t, "signal", 0, 0.0, output[1].(entities.Scalar).Value)
	})

	t.Run("trade has a zero candle body and a zero range", func(t *testing.T) {
		t.Parallel()

		csi, err := NewCandlestickStrengthIndex(&Params{R: 1, S: 1, U: 1, Ul: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		output := csi.UpdateTrade(&entities.Trade{Time: tm, Price: 10.0, Volume: 1.0})

		checkVal(t, "csi", 0, 0.0, output[0].(entities.Scalar).Value)
		checkVal(t, "signal", 0, 0.0, output[1].(entities.Scalar).Value)
	})
}

func TestCandlestickStrengthIndexMnemonic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params *Params
		exp    string
	}{
		{"all parameters zero", &Params{}, "csi(20,5,3,3)"},
		{"default parameters", DefaultParams(), "csi(20,5,3,3)"},
		{"double smoothing", &Params{R: 25, S: 13, U: 1, Ul: 3}, "csi(25,13,1,3)"},
		{"passthrough signal", &Params{R: 20, S: 5, U: 3, Ul: 1}, "csi(20,5,3,1)"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			csi, err := NewCandlestickStrengthIndex(tt.params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if csi.mnemonic != tt.exp {
				t.Errorf("mnemonic: expected '%s', got '%s'", tt.exp, csi.mnemonic)
			}

			expDesc := "Candlestick Strength Index " + tt.exp
			if meta := csi.Metadata(); meta.Description != expDesc {
				t.Errorf("description: expected '%s', got '%s'", expDesc, meta.Description)
			}
		})
	}
}

func TestCandlestickStrengthIndexMetadata(t *testing.T) {
	t.Parallel()

	csi, err := NewCandlestickStrengthIndex(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	meta := csi.Metadata()

	if meta.Identifier != core.CandlestickStrengthIndex {
		t.Errorf("identifier: expected CandlestickStrengthIndex, got %v", meta.Identifier)
	}

	if meta.Mnemonic != "csi(20,5,3,3)" {
		t.Errorf("mnemonic: expected 'csi(20,5,3,3)', got '%s'", meta.Mnemonic)
	}

	if len(meta.Outputs) != 2 {
		t.Errorf("outputs: expected 2, got %d", len(meta.Outputs))
	}
}

func TestCandlestickStrengthIndexInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params *Params
	}{
		{"r too small", &Params{R: -1, S: 5, U: 3, Ul: 3}},
		{"s too small", &Params{R: 20, S: -1, U: 3, Ul: 3}},
		{"u too small", &Params{R: 20, S: 5, U: -1, Ul: 3}},
		{"ul too small", &Params{R: 20, S: 5, U: 3, Ul: -1}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := NewCandlestickStrengthIndex(tt.params); err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}
