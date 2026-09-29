//nolint:testpackage
package tickvolumeindicator

import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

const tolerance = 1e-10

type tviCombo struct {
	name     string
	r        int
	s        int
	u        int
	expected []float64
}

func tviCombos() []tviCombo {
	return []tviCombo{
		{"R12_S12_U1", 12, 12, 1, expectedR12_S12_U1},
		{"R25_S13_U1", 25, 13, 1, expectedR25_S13_U1},
		{"R32_S32_U5", 32, 32, 5, expectedR32_S32_U5},
		{"R1_S1_U1", 1, 1, 1, expectedR1_S1_U1},
		{"R32_S5_U1", 32, 5, 1, expectedR32_S5_U1},
		{"R12_S12_U5", 12, 12, 5, expectedR12_S12_U5},
		{"R20_S5_U3", 20, 5, 3, expectedR20_S5_U3},
		{"R5_S5_U5", 5, 5, 5, expectedR5_S5_U5},
		{"R32_S32_U1", 32, 32, 1, expectedR32_S32_U1},
		{"R10_S10_U1", 10, 10, 1, expectedR10_S10_U1},
		{"R50_S25_U1", 50, 25, 1, expectedR50_S25_U1},
		{"R12_S26_U9", 12, 26, 9, expectedR12_S26_U9},
		{"R3_S3_U3", 3, 3, 3, expectedR3_S3_U3},
		{"R7_S4_U2", 7, 4, 2, expectedR7_S4_U2},
		{"R64_S1_U1", 64, 1, 1, expectedR64_S1_U1},
		{"R12_S12_U3", 12, 12, 3, expectedR12_S12_U3},
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

func checkOutput(t *testing.T, name string, i int, tm time.Time, exp float64, output core.Output) {
	t.Helper()

	if len(output) != 1 {
		t.Fatalf("%s[%d]: outputs: expected 1, got %d", name, i, len(output))
	}

	scalar := output[0].(entities.Scalar)
	if scalar.Time != tm {
		t.Errorf("%s[%d]: time: expected %v, got %v", name, i, tm, scalar.Time)
	}

	checkVal(t, name, i, exp, scalar.Value)
}

func TestTickVolumeIndicatorData(t *testing.T) {
	t.Parallel()

	for _, combo := range tviCombos() {
		t.Run(combo.name, func(t *testing.T) {
			t.Parallel()

			tvi, err := NewTickVolumeIndicator(&Params{R: combo.r, S: combo.s, U: combo.u})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for i := range testUpticks {
				checkVal(t, "tvi", i, combo.expected[i], tvi.Update(testUpticks[i], testDownticks[i]))
			}
		})
	}
}

func TestTickVolumeIndicatorPassthrough(t *testing.T) {
	t.Parallel()

	// All EMA stages are passthroughs, so TVI = 100*(up-down)/(up+down).
	tvi, err := NewTickVolumeIndicator(&Params{R: 1, S: 1, U: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name      string
		upticks   float64
		downticks float64
		exp       float64
	}{
		{"mostly up", 8.0, 2.0, 60.0},
		{"all down", 0.0, 5.0, -100.0},
		{"flat market", 0.0, 0.0, 0.0},
	}

	for i, tt := range tests {
		checkVal(t, tt.name, i, tt.exp, tvi.Update(tt.upticks, tt.downticks))
	}
}

func TestTickVolumeIndicatorIsPrimed(t *testing.T) {
	t.Parallel()

	t.Run("update", func(t *testing.T) {
		t.Parallel()

		tvi, err := NewTickVolumeIndicator(DefaultParams())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if tvi.IsPrimed() {
			t.Errorf("IsPrimed: expected false before the first update")
		}

		tvi.Update(testUpticks[0], testDownticks[0])

		if !tvi.IsPrimed() {
			t.Errorf("IsPrimed: expected true after the first update")
		}
	})

	t.Run("tick rule", func(t *testing.T) {
		t.Parallel()

		tvi, err := NewTickVolumeIndicator(DefaultParams())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if tvi.IsPrimed() {
			t.Errorf("IsPrimed: expected false before the first update")
		}

		tvi.UpdateScalar(&entities.Scalar{Time: time.Now(), Value: 10.0})

		if !tvi.IsPrimed() {
			t.Errorf("IsPrimed: expected true after the first update")
		}
	})
}

func TestTickVolumeIndicatorUpdateBar(t *testing.T) {
	t.Parallel()

	// A bar maps up = close - low, down = high - close: the fixture proxy.
	tvi, err := NewTickVolumeIndicator(&Params{R: 12, S: 12, U: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tm := time.Now()

	for i := range testInput {
		output := tvi.UpdateBar(&entities.Bar{
			Time: tm, Open: testOpen[i], High: testHigh[i],
			Low: testLow[i], Close: testInput[i],
		})
		checkOutput(t, "bar", i, tm, expectedR12_S12_U1[i], output)
	}
}

func TestTickVolumeIndicatorUpdateEntities(t *testing.T) {
	t.Parallel()

	tm := time.Now()

	// Tick rule with passthrough stages: first -> 0, up -> +100, down -> -100, flat -> 0.
	values := []float64{10.0, 12.0, 11.0, 11.0}
	expected := []float64{0.0, 100.0, -100.0, 0.0}

	t.Run("scalar applies the tick rule to the value", func(t *testing.T) {
		t.Parallel()

		tvi, err := NewTickVolumeIndicator(&Params{R: 1, S: 1, U: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for i, value := range values {
			checkOutput(t, "scalar", i, tm, expected[i], tvi.UpdateScalar(&entities.Scalar{Time: tm, Value: value}))
		}
	})

	t.Run("trade applies the tick rule to the price", func(t *testing.T) {
		t.Parallel()

		tvi, err := NewTickVolumeIndicator(&Params{R: 1, S: 1, U: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for i, value := range values {
			checkOutput(t, "trade", i, tm, expected[i],
				tvi.UpdateTrade(&entities.Trade{Time: tm, Price: value, Volume: 1.0}))
		}
	})

	t.Run("quote applies the tick rule to the mid price", func(t *testing.T) {
		t.Parallel()

		tvi, err := NewTickVolumeIndicator(&Params{R: 1, S: 1, U: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		quotes := []entities.Quote{
			{Time: tm, Bid: 9.0, Ask: 11.0, BidSize: 1.0, AskSize: 1.0},
			{Time: tm, Bid: 11.0, Ask: 13.0, BidSize: 1.0, AskSize: 1.0},
			{Time: tm, Bid: 10.0, Ask: 12.0, BidSize: 1.0, AskSize: 1.0},
			{Time: tm, Bid: 10.5, Ask: 11.5, BidSize: 1.0, AskSize: 1.0},
		}

		for i := range quotes {
			checkOutput(t, "quote", i, tm, expected[i], tvi.UpdateQuote(&quotes[i]))
		}
	})

	t.Run("scalar matches update of the value changes", func(t *testing.T) {
		t.Parallel()

		tvi, err := NewTickVolumeIndicator(&Params{R: 12, S: 12, U: 3})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ref, err := NewTickVolumeIndicator(&Params{R: 12, S: 12, U: 3})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for i := range testInput {
			diff := 0.0
			if i > 0 {
				diff = testInput[i] - testInput[i-1]
			}

			exp := ref.Update(max(diff, 0.0), max(-diff, 0.0))
			checkOutput(t, "scalar", i, tm, exp, tvi.UpdateScalar(&entities.Scalar{Time: tm, Value: testInput[i]}))
		}
	})
}

func TestTickVolumeIndicatorMnemonic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params *Params
		exp    string
	}{
		{"all parameters zero", &Params{}, "tvi(12,12,1)"},
		{"default parameters", DefaultParams(), "tvi(12,12,1)"},
		{"triple smoothing", &Params{R: 32, S: 32, U: 5}, "tvi(32,32,5)"},
		{"only r set", &Params{R: 25}, "tvi(25,12,1)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tvi, err := NewTickVolumeIndicator(tt.params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tvi.mnemonic != tt.exp {
				t.Errorf("mnemonic: expected '%s', got '%s'", tt.exp, tvi.mnemonic)
			}

			expDesc := "Tick Volume Indicator " + tt.exp
			if meta := tvi.Metadata(); meta.Description != expDesc {
				t.Errorf("description: expected '%s', got '%s'", expDesc, meta.Description)
			}
		})
	}
}

func TestTickVolumeIndicatorMetadata(t *testing.T) {
	t.Parallel()

	tvi, err := NewTickVolumeIndicator(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	const (
		mnemonic    = "tvi(12,12,1)"
		description = "Tick Volume Indicator tvi(12,12,1)"
	)

	meta := tvi.Metadata()

	if meta.Identifier != core.TickVolumeIndicator {
		t.Errorf("identifier: expected TickVolumeIndicator, got %v", meta.Identifier)
	}

	if meta.Mnemonic != mnemonic {
		t.Errorf("mnemonic: expected '%s', got '%s'", mnemonic, meta.Mnemonic)
	}

	if meta.Description != description {
		t.Errorf("description: expected '%s', got '%s'", description, meta.Description)
	}

	if len(meta.Outputs) != 1 {
		t.Fatalf("outputs: expected 1, got %d", len(meta.Outputs))
	}

	if meta.Outputs[0].Kind != int(Value) {
		t.Errorf("Outputs[0].Kind: expected %d, got %d", int(Value), meta.Outputs[0].Kind)
	}

	if meta.Outputs[0].Shape != shape.Scalar {
		t.Errorf("Outputs[0].Shape: expected %v, got %v", shape.Scalar, meta.Outputs[0].Shape)
	}

	if meta.Outputs[0].Mnemonic != mnemonic {
		t.Errorf("Outputs[0].Mnemonic: expected '%s', got '%s'", mnemonic, meta.Outputs[0].Mnemonic)
	}

	if meta.Outputs[0].Description != description {
		t.Errorf("Outputs[0].Description: expected '%s', got '%s'", description, meta.Outputs[0].Description)
	}
}

func TestTickVolumeIndicatorInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params *Params
		exp    string
	}{
		{"r too small", &Params{R: -1, S: 12, U: 1}, "invalid tick volume indicator parameters: r should be greater than 0"},
		{"s too small", &Params{R: 12, S: -1, U: 1}, "invalid tick volume indicator parameters: s should be greater than 0"},
		{"u too small", &Params{R: 12, S: 12, U: -1}, "invalid tick volume indicator parameters: u should be greater than 0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewTickVolumeIndicator(tt.params)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}

			if err.Error() != tt.exp {
				t.Errorf("error: expected '%s', got '%s'", tt.exp, err.Error())
			}
		})
	}
}
