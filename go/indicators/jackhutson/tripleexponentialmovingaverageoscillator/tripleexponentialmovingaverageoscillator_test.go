//nolint:testpackage
package tripleexponentialmovingaverageoscillator

import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func TestTripleExponentialMovingAverageOscillatorValues(t *testing.T) {
	t.Parallel()

	closes := testCloses()
	expected := testExpected()

	ind, err := NewTripleExponentialMovingAverageOscillator(
		&TripleExponentialMovingAverageOscillatorParams{Length: 5},
	)
	if err != nil {
		t.Fatal(err)
	}

	for i, c := range closes {
		result := ind.Update(c)

		if math.IsNaN(expected[i]) {
			if !math.IsNaN(result) {
				t.Errorf("[%d] expected NaN, got %v", i, result)
			}
		} else {
			if math.IsNaN(result) {
				t.Errorf("[%d] expected %v, got NaN", i, expected[i])
			} else if !inDelta(expected[i], result, tolerance) {
				t.Errorf("[%d] expected %v, got %v", i, expected[i], result)
			}
		}
	}
}

func TestTripleExponentialMovingAverageOscillatorSpotChecks(t *testing.T) {
	t.Parallel()

	const spotTolerance = 1e-4

	closes := testCloses()

	ind, err := NewTripleExponentialMovingAverageOscillator(
		&TripleExponentialMovingAverageOscillatorParams{Length: 5},
	)
	if err != nil {
		t.Fatal(err)
	}

	results := make([]float64, len(closes))
	for i, c := range closes {
		results[i] = ind.Update(c)
	}

	// TaLib spot checks: begIndex=13, nbElement=239.
	if !inDelta(0.2589, results[13], spotTolerance) {
		t.Errorf("spot check output[0]: expected ~0.2589, got %v", results[13])
	}

	if !inDelta(0.010495, results[14], spotTolerance) {
		t.Errorf("spot check output[1]: expected ~0.010495, got %v", results[14])
	}

	if !inDelta(-0.058, results[250], 1e-3) {
		t.Errorf("spot check output[237]: expected ~-0.058, got %v", results[250])
	}

	if !inDelta(-0.095, results[251], 1e-3) {
		t.Errorf("spot check output[238]: expected ~-0.095, got %v", results[251])
	}
}

func TestTripleExponentialMovingAverageOscillatorIsPrimed(t *testing.T) {
	t.Parallel()

	closes := testCloses()

	ind, err := NewTripleExponentialMovingAverageOscillator(
		&TripleExponentialMovingAverageOscillatorParams{Length: 5},
	)
	if err != nil {
		t.Fatal(err)
	}

	// Lookback = 3*(5-1) + 1 = 13. First primed at index 13.
	for i := 0; i < 13; i++ {
		ind.Update(closes[i])
		if ind.IsPrimed() {
			t.Errorf("should not be primed at index %d", i)
		}
	}

	ind.Update(closes[13])
	if !ind.IsPrimed() {
		t.Error("should be primed at index 13")
	}
}

func TestTripleExponentialMovingAverageOscillatorMetadata(t *testing.T) {
	t.Parallel()

	ind, err := NewTripleExponentialMovingAverageOscillator(
		&TripleExponentialMovingAverageOscillatorParams{Length: 30},
	)
	if err != nil {
		t.Fatal(err)
	}

	meta := ind.Metadata()
	if meta.Identifier != core.TripleExponentialMovingAverageOscillator {
		t.Errorf("expected identifier TripleExponentialMovingAverageOscillator, got %v", meta.Identifier)
	}

	if meta.Mnemonic != "trix(30)" {
		t.Errorf("expected mnemonic trix(30), got %s", meta.Mnemonic)
	}

	if meta.Description != "Triple exponential moving average oscillator trix(30)" {
		t.Errorf("expected description mismatch, got %s", meta.Description)
	}

	if len(meta.Outputs) != 1 {
		t.Fatalf("expected 1 output, got %d", len(meta.Outputs))
	}

	if meta.Outputs[0].Kind != int(Value) {
		t.Errorf("expected output kind %d, got %d", Value, meta.Outputs[0].Kind)
	}

	if meta.Outputs[0].Shape != shape.Scalar {
		t.Errorf("expected scalar output type, got %v", meta.Outputs[0].Shape)
	}
}

func TestTripleExponentialMovingAverageOscillatorInvalidParams(t *testing.T) {
	t.Parallel()

	_, err := NewTripleExponentialMovingAverageOscillator(
		&TripleExponentialMovingAverageOscillatorParams{Length: 0},
	)
	if err == nil {
		t.Error("expected error for zero length")
	}
}

func TestTripleExponentialMovingAverageOscillatorNaN(t *testing.T) {
	t.Parallel()

	ind, err := NewTripleExponentialMovingAverageOscillator(
		&TripleExponentialMovingAverageOscillatorParams{Length: 5},
	)
	if err != nil {
		t.Fatal(err)
	}

	result := ind.Update(math.NaN())
	if !math.IsNaN(result) {
		t.Errorf("expected NaN for NaN input, got %v", result)
	}
}

func TestTripleExponentialMovingAverageOscillatorMnemonic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		params   TripleExponentialMovingAverageOscillatorParams
		mnemonic string
	}{
		{
			name:     "default components",
			params:   TripleExponentialMovingAverageOscillatorParams{Length: 5},
			mnemonic: "trix(5)",
		},
		{
			name: "bar component",
			params: TripleExponentialMovingAverageOscillatorParams{
				Length:       5,
				BarComponent: entities.BarMedianPrice,
			},
			mnemonic: "trix(5, hl/2)",
		},
		{
			name: "quote component",
			params: TripleExponentialMovingAverageOscillatorParams{
				Length:         5,
				QuoteComponent: entities.QuoteBidPrice,
			},
			mnemonic: "trix(5, b)",
		},
		{
			name: "trade component",
			params: TripleExponentialMovingAverageOscillatorParams{
				Length:         5,
				TradeComponent: entities.TradeVolume,
			},
			mnemonic: "trix(5, v)",
		},
		{
			name: "bar and quote components",
			params: TripleExponentialMovingAverageOscillatorParams{
				Length:         5,
				BarComponent:   entities.BarOpenPrice,
				QuoteComponent: entities.QuoteBidPrice,
			},
			mnemonic: "trix(5, o, b)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ind, err := NewTripleExponentialMovingAverageOscillator(&tt.params)
			if err != nil {
				t.Fatal(err)
			}

			if ind.Metadata().Mnemonic != tt.mnemonic {
				t.Errorf("expected mnemonic %s, got %s", tt.mnemonic, ind.Metadata().Mnemonic)
			}
		})
	}
}

func TestTripleExponentialMovingAverageOscillatorUpdateBar(t *testing.T) {
	t.Parallel()

	ind, err := NewTripleExponentialMovingAverageOscillator(
		&TripleExponentialMovingAverageOscillatorParams{Length: 5},
	)
	if err != nil {
		t.Fatal(err)
	}

	bar := &entities.Bar{Time: time.Now(), Close: 100.0}
	result := ind.UpdateBar(bar)
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestTripleExponentialMovingAverageOscillatorUpdateQuote(t *testing.T) {
	t.Parallel()

	ind, err := NewTripleExponentialMovingAverageOscillator(
		&TripleExponentialMovingAverageOscillatorParams{Length: 5},
	)
	if err != nil {
		t.Fatal(err)
	}

	quote := &entities.Quote{Time: time.Now(), Bid: 100.0, Ask: 102.0}
	result := ind.UpdateQuote(quote)
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestTripleExponentialMovingAverageOscillatorUpdateTrade(t *testing.T) {
	t.Parallel()

	ind, err := NewTripleExponentialMovingAverageOscillator(
		&TripleExponentialMovingAverageOscillatorParams{Length: 5},
	)
	if err != nil {
		t.Fatal(err)
	}

	trade := &entities.Trade{Time: time.Now(), Price: 100.0}
	result := ind.UpdateTrade(trade)
	if result == nil {
		t.Error("expected non-nil result")
	}
}
