//nolint:testpackage
package ultimateoscillator

//nolint: gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func TestNewUltimateOscillator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		params  UltimateOscillatorParams
		wantErr bool
	}{
		{"default params", UltimateOscillatorParams{}, false},
		{"custom params", UltimateOscillatorParams{Length1: 5, Length2: 10, Length3: 20}, false},
		{"length1 too small", UltimateOscillatorParams{Length1: 1}, true},
		{"length2 too small", UltimateOscillatorParams{Length2: 1}, true},
		{"length3 too small", UltimateOscillatorParams{Length3: 1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewUltimateOscillator(&tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewUltimateOscillator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUltimateOscillatorUpdate(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-4

	highs := testInputHigh()
	lows := testInputLow()
	closes := testInputClose()
	expected := testExpectedUltOsc()

	t.Run("period 7-14-28 all 252 rows", func(t *testing.T) {
		t.Parallel()

		ind, err := NewUltimateOscillator(&UltimateOscillatorParams{})
		if err != nil {
			t.Fatalf("NewUltimateOscillator() error = %v", err)
		}

		for i := range highs {
			result := ind.Update(closes[i], highs[i], lows[i])

			if math.IsNaN(expected[i]) {
				if !math.IsNaN(result) {
					t.Errorf("index %d: expected NaN, got %v", i, result)
				}

				continue
			}

			if math.IsNaN(result) {
				t.Errorf("index %d: expected %v, got NaN", i, expected[i])

				continue
			}

			if diff := math.Abs(result - expected[i]); diff > tolerance {
				t.Errorf("index %d: expected %v, got %v (diff %v)", i, expected[i], result, diff)
			}
		}
	})
}

func TestUltimateOscillatorIsPrimed(t *testing.T) {
	t.Parallel()

	highs := testInputHigh()
	lows := testInputLow()
	closes := testInputClose()

	tests := []struct {
		name     string
		params   UltimateOscillatorParams
		primedAt int // index at which IsPrimed should first be true
	}{
		{"default 7-14-28", UltimateOscillatorParams{}, 28},
		{"custom 2-3-4", UltimateOscillatorParams{Length1: 2, Length2: 3, Length3: 4}, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ind, err := NewUltimateOscillator(&tt.params)
			if err != nil {
				t.Fatalf("error: %v", err)
			}

			for i := 0; i < tt.primedAt && i < len(highs); i++ {
				ind.Update(closes[i], highs[i], lows[i])

				if ind.IsPrimed() {
					t.Errorf("expected not primed at index %d", i)
				}
			}

			if tt.primedAt < len(highs) {
				ind.Update(closes[tt.primedAt], highs[tt.primedAt], lows[tt.primedAt])

				if !ind.IsPrimed() {
					t.Errorf("expected primed at index %d", tt.primedAt)
				}
			}
		})
	}
}

func TestUltimateOscillatorMetadata(t *testing.T) {
	t.Parallel()

	ind, err := NewUltimateOscillator(&UltimateOscillatorParams{})
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	meta := ind.Metadata()

	if meta.Identifier != core.UltimateOscillator {
		t.Errorf("expected type %v, got %v", core.UltimateOscillator, meta.Identifier)
	}

	if meta.Mnemonic != "ultosc(7, 14, 28)" {
		t.Errorf("expected mnemonic 'ultosc(7, 14, 28)', got '%v'", meta.Mnemonic)
	}

	if len(meta.Outputs) != 1 {
		t.Fatalf("expected 1 output, got %d", len(meta.Outputs))
	}

	if meta.Outputs[0].Kind != int(Value) {
		t.Errorf("expected output kind %d, got %d", Value, meta.Outputs[0].Kind)
	}

	if meta.Outputs[0].Shape != shape.Scalar {
		t.Errorf("expected output type %v, got %v", shape.Scalar, meta.Outputs[0].Shape)
	}
}

func TestUltimateOscillatorUpdateEntity(t *testing.T) {
	t.Parallel()

	now := time.Now()

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		ind, _ := NewUltimateOscillator(&UltimateOscillatorParams{})

		bar := &entities.Bar{Time: now, High: 100, Low: 90, Close: 95, Open: 92, Volume: 1000}
		output := ind.UpdateBar(bar)

		if len(output) != 1 {
			t.Fatalf("expected 1 output, got %d", len(output))
		}

		scalar, ok := output[0].(entities.Scalar)
		if !ok {
			t.Fatal("expected entities.Scalar")
		}

		if scalar.Time != now {
			t.Errorf("expected time %v, got %v", now, scalar.Time)
		}
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		ind, _ := NewUltimateOscillator(&UltimateOscillatorParams{})

		quote := &entities.Quote{Time: now, Bid: 100, Ask: 102}
		output := ind.UpdateQuote(quote)

		if len(output) != 1 {
			t.Fatalf("expected 1 output, got %d", len(output))
		}
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		ind, _ := NewUltimateOscillator(&UltimateOscillatorParams{})

		trade := &entities.Trade{Time: now, Price: 100, Volume: 500}
		output := ind.UpdateTrade(trade)

		if len(output) != 1 {
			t.Fatalf("expected 1 output, got %d", len(output))
		}
	})

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		ind, _ := NewUltimateOscillator(&UltimateOscillatorParams{})

		scalar := &entities.Scalar{Time: now, Value: 100}
		output := ind.UpdateScalar(scalar)

		if len(output) != 1 {
			t.Fatalf("expected 1 output, got %d", len(output))
		}
	})
}

func TestUltimateOscillatorNaN(t *testing.T) {
	t.Parallel()

	ind, _ := NewUltimateOscillator(&UltimateOscillatorParams{})

	// NaN input should return NaN without corrupting state.
	result := ind.Update(math.NaN(), 100, 90)
	if !math.IsNaN(result) {
		t.Errorf("expected NaN for NaN close, got %v", result)
	}

	result = ind.Update(95, math.NaN(), 90)
	if !math.IsNaN(result) {
		t.Errorf("expected NaN for NaN high, got %v", result)
	}

	result = ind.Update(95, 100, math.NaN())
	if !math.IsNaN(result) {
		t.Errorf("expected NaN for NaN low, got %v", result)
	}
}

func TestUltimateOscillatorTaLibSpotChecks(t *testing.T) {
	t.Parallel()

	// Spot checks from test_per_hlc.c:
	// { 0, TA_ULTOSC_TEST, 0, 251, 7, 14, 28, TA_SUCCESS, 0, 47.1713, 28, 252-28 }
	// { 0, TA_ULTOSC_TEST, 0, 251, 7, 14, 28, TA_SUCCESS, 1, 46.2802, 28, 252-28 }
	// { 1, TA_ULTOSC_TEST, 0, 251, 7, 14, 28, TA_SUCCESS, 252-29, 40.0854, 28, 252-28 }
	// Output index 0 at begIndex 28 -> input index 28
	// Output index 1 at begIndex 28 -> input index 29
	// Output index 252-29=223 at begIndex 28 -> input index 251

	const tolerance = 1e-4

	highs := testInputHigh()
	lows := testInputLow()
	closes := testInputClose()

	ind, err := NewUltimateOscillator(&UltimateOscillatorParams{})
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	results := make([]float64, len(highs))
	for i := range highs {
		results[i] = ind.Update(closes[i], highs[i], lows[i])
	}

	spots := []struct {
		index int
		value float64
	}{
		{28, 47.1713},
		{29, 46.2802},
		{251, 40.0854},
	}

	for _, s := range spots {
		if diff := math.Abs(results[s.index] - s.value); diff > tolerance {
			t.Errorf("spot check index %d: expected %v, got %v (diff %v)", s.index, s.value, results[s.index], diff)
		}
	}
}
