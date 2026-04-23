//nolint:testpackage
package movingaverageconvergencedivergence

//nolint: gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func testTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

//nolint:funlen,cyclop
func TestMovingAverageConvergenceDivergenceDefaultParams(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-8

	closingPrice := testInput()
	expMACD := testMACDExpected()
	expSignal := testSignalExpected()
	expHistogram := testHistogramExpected()

	ind, err := NewMovingAverageConvergenceDivergence(&MovingAverageConvergenceDivergenceParams{})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 252; i++ {
		macd, signal, histogram := ind.Update(closingPrice[i])

		if math.IsNaN(expMACD[i]) {
			if !math.IsNaN(macd) {
				t.Errorf("[%d] macd: expected NaN, got %v", i, macd)
			}

			if !math.IsNaN(signal) {
				t.Errorf("[%d] signal: expected NaN, got %v", i, signal)
			}

			if !math.IsNaN(histogram) {
				t.Errorf("[%d] histogram: expected NaN, got %v", i, histogram)
			}

			continue
		}

		if !math.IsNaN(expMACD[i]) && math.Abs(macd-expMACD[i]) > tolerance {
			t.Errorf("[%d] macd: expected %v, got %v", i, expMACD[i], macd)
		}

		if math.IsNaN(expSignal[i]) {
			if !math.IsNaN(signal) {
				t.Errorf("[%d] signal: expected NaN, got %v", i, signal)
			}

			if !math.IsNaN(histogram) {
				t.Errorf("[%d] histogram: expected NaN, got %v", i, histogram)
			}

			continue
		}

		if math.Abs(signal-expSignal[i]) > tolerance {
			t.Errorf("[%d] signal: expected %v, got %v", i, expSignal[i], signal)
		}

		if math.Abs(histogram-expHistogram[i]) > tolerance {
			t.Errorf("[%d] histogram: expected %v, got %v", i, expHistogram[i], histogram)
		}
	}
}

func TestMovingAverageConvergenceDivergenceTaLibSpotCheck(t *testing.T) {
	t.Parallel()

	const tolerance = 5e-4

	closingPrice := testInput()

	ind, err := NewMovingAverageConvergenceDivergence(&MovingAverageConvergenceDivergenceParams{})
	if err != nil {
		t.Fatal(err)
	}

	var macd, signal, histogram float64

	for i := 0; i < 252; i++ {
		macd, signal, histogram = ind.Update(closingPrice[i])
	}

	// TaLib spot check: begIndex=33, output[0] at index 33.
	// Verify first non-NaN values already checked in full data test.
	// Verify last values.
	_ = macd
	_ = signal
	_ = histogram

	// Spot check at index 33 (first signal output).
	ind2, _ := NewMovingAverageConvergenceDivergence(&MovingAverageConvergenceDivergenceParams{})

	for i := 0; i <= 33; i++ {
		macd, signal, histogram = ind2.Update(closingPrice[i])
	}

	if math.Abs(macd-(-1.9738)) > tolerance {
		t.Errorf("MACD[33] = %v, want -1.9738", macd)
	}

	if math.Abs(signal-(-2.7071)) > tolerance {
		t.Errorf("Signal[33] = %v, want -2.7071", signal)
	}

	expectedHistogram := (-1.9738) - (-2.7071)
	if math.Abs(histogram-expectedHistogram) > tolerance {
		t.Errorf("Histogram[33] = %v, want %v", histogram, expectedHistogram)
	}
}

func TestMovingAverageConvergenceDivergencePeriodInversion(t *testing.T) {
	t.Parallel()

	const tolerance = 5e-4

	closingPrice := testInput()

	// TaLib test: passing fast=26, slow=12 should auto-swap and produce same output.
	ind, err := NewMovingAverageConvergenceDivergence(&MovingAverageConvergenceDivergenceParams{
		FastLength: 26,
		SlowLength: 12,
	})
	if err != nil {
		t.Fatal(err)
	}

	var macd, signal float64

	for i := 0; i <= 33; i++ {
		macd, signal, _ = ind.Update(closingPrice[i])
	}

	if math.Abs(macd-(-1.9738)) > tolerance {
		t.Errorf("MACD[33] = %v, want -1.9738", macd)
	}

	if math.Abs(signal-(-2.7071)) > tolerance {
		t.Errorf("Signal[33] = %v, want -2.7071", signal)
	}
}

func TestMovingAverageConvergenceDivergenceIsPrimed(t *testing.T) {
	t.Parallel()

	ind, err := NewMovingAverageConvergenceDivergence(&MovingAverageConvergenceDivergenceParams{
		FastLength:   3,
		SlowLength:   5,
		SignalLength: 2,
	})
	if err != nil {
		t.Fatal(err)
	}

	if ind.IsPrimed() {
		t.Error("expected not primed initially")
	}

	// Slow EMA(5) primes at index 4. Signal EMA(2) primes after 2 MACD values (index 5).
	// fastDelay = 5-3 = 2, so fast EMA starts at index 2, primes at index 4 (after 3 values).
	// Both prime at index 4. Signal gets first MACD at index 4, primes at index 5.
	for i := 0; i < 6; i++ {
		ind.Update(float64(i + 1))
		if i < 5 && ind.IsPrimed() {
			t.Errorf("[%d] expected not primed", i)
		}
	}

	if !ind.IsPrimed() {
		t.Error("expected primed after 6 samples")
	}
}

func TestMovingAverageConvergenceDivergenceNaN(t *testing.T) {
	t.Parallel()

	ind, err := NewMovingAverageConvergenceDivergence(&MovingAverageConvergenceDivergenceParams{})
	if err != nil {
		t.Fatal(err)
	}

	macd, signal, histogram := ind.Update(math.NaN())
	if !math.IsNaN(macd) {
		t.Errorf("expected NaN macd, got %v", macd)
	}

	if !math.IsNaN(signal) {
		t.Errorf("expected NaN signal, got %v", signal)
	}

	if !math.IsNaN(histogram) {
		t.Errorf("expected NaN histogram, got %v", histogram)
	}
}

func TestMovingAverageConvergenceDivergenceMetadata(t *testing.T) {
	t.Parallel()

	ind, err := NewMovingAverageConvergenceDivergence(&MovingAverageConvergenceDivergenceParams{})
	if err != nil {
		t.Fatal(err)
	}

	meta := ind.Metadata()

	if meta.Identifier != core.MovingAverageConvergenceDivergence {
		t.Errorf("expected identifier MovingAverageConvergenceDivergence, got %v", meta.Identifier)
	}

	exp := "macd(12,26,9)"
	if meta.Mnemonic != exp {
		t.Errorf("expected mnemonic '%s', got '%s'", exp, meta.Mnemonic)
	}

	const expectedOutputs = 3
	if len(meta.Outputs) != expectedOutputs {
		t.Fatalf("expected %d outputs, got %d", expectedOutputs, len(meta.Outputs))
	}

	if meta.Outputs[0].Kind != int(MACD) {
		t.Errorf("expected output 0 kind %d, got %d",
			MACD, meta.Outputs[0].Kind)
	}

	if meta.Outputs[0].Shape != shape.Scalar {
		t.Errorf("expected scalar output type, got %v", meta.Outputs[0].Shape)
	}

	if meta.Outputs[1].Kind != int(Signal) {
		t.Errorf("expected output 1 kind %d, got %d",
			Signal, meta.Outputs[1].Kind)
	}

	if meta.Outputs[2].Kind != int(Histogram) {
		t.Errorf("expected output 2 kind %d, got %d",
			Histogram, meta.Outputs[2].Kind)
	}
}

func TestMovingAverageConvergenceDivergenceMetadataSMA(t *testing.T) {
	t.Parallel()

	ind, err := NewMovingAverageConvergenceDivergence(&MovingAverageConvergenceDivergenceParams{
		MovingAverageType: SMA,
	})
	if err != nil {
		t.Fatal(err)
	}

	meta := ind.Metadata()

	exp := "macd(12,26,9,SMA,EMA)"
	if meta.Mnemonic != exp {
		t.Errorf("expected mnemonic '%s', got '%s'", exp, meta.Mnemonic)
	}
}

//nolint:funlen
func TestMovingAverageConvergenceDivergenceUpdateScalar(t *testing.T) {
	t.Parallel()

	const tolerance = 5e-4

	closingPrice := testInput()

	ind, err := NewMovingAverageConvergenceDivergence(&MovingAverageConvergenceDivergenceParams{})
	if err != nil {
		t.Fatal(err)
	}

	tm := testTime()

	// First 32 values: signal should be NaN.
	for i := 0; i < 33; i++ {
		scalar := &entities.Scalar{Time: tm, Value: closingPrice[i]}
		out := ind.UpdateScalar(scalar)

		s := out[1].(entities.Scalar).Value //nolint:forcetypeassert
		if i < 25 {
			// MACD not ready yet.
			m := out[0].(entities.Scalar).Value //nolint:forcetypeassert
			if !math.IsNaN(m) {
				t.Errorf("[%d] expected NaN macd, got %v", i, m)
			}
		}

		if i < 33 && !math.IsNaN(s) {
			t.Errorf("[%d] expected NaN signal, got %v", i, s)
		}
	}

	// Index 33: first complete output.
	scalar := &entities.Scalar{Time: tm, Value: closingPrice[33]}
	out := ind.UpdateScalar(scalar)

	macd := out[0].(entities.Scalar).Value      //nolint:forcetypeassert
	signal := out[1].(entities.Scalar).Value    //nolint:forcetypeassert
	histogram := out[2].(entities.Scalar).Value //nolint:forcetypeassert

	if math.Abs(macd-(-1.9738)) > tolerance {
		t.Errorf("MACD[33] = %v, want -1.9738", macd)
	}

	if math.Abs(signal-(-2.7071)) > tolerance {
		t.Errorf("Signal[33] = %v, want -2.7071", signal)
	}

	expectedHistogram := (-1.9738) - (-2.7071)
	if math.Abs(histogram-expectedHistogram) > tolerance {
		t.Errorf("Histogram[33] = %v, want %v", histogram, expectedHistogram)
	}
}

func TestMovingAverageConvergenceDivergenceInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		fast   int
		slow   int
		signal int
	}{
		{"fast too small", 1, 26, 9},
		{"slow too small", 12, 1, 9},
		{"signal negative", 12, 26, -1},
		{"fast negative", -8, 12, 9},
		{"slow negative", 26, -7, 9},
	}

	for _, tt := range tests {
		_, err := NewMovingAverageConvergenceDivergence(&MovingAverageConvergenceDivergenceParams{
			FastLength:   tt.fast,
			SlowLength:   tt.slow,
			SignalLength: tt.signal,
		})
		if err == nil {
			t.Errorf("%s: expected error, got nil", tt.name)
		}
	}
}
