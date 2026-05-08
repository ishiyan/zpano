//nolint:testpackage
package parabolicstopandreverse

//nolint: gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func TestParabolicStopAndReverse252Bar(t *testing.T) {
	t.Parallel()

	const tol = 1e-6

	sar, err := NewParabolicStopAndReverse(&ParabolicStopAndReverseParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	highs := testHighs()
	lows := testLows()
	expected := testExpected()

	for i := range highs {
		result := sar.UpdateHL(highs[i], lows[i])

		if math.IsNaN(expected[i]) {
			if !math.IsNaN(result) {
				t.Errorf("[%d] expected NaN, got %v", i, result)
			}

			continue
		}

		diff := math.Abs(result - expected[i])
		if diff > tol {
			t.Errorf("[%d] expected %.10f, got %.10f, diff %.10f", i, expected[i], result, diff)
		}
	}
}

func TestParabolicStopAndReverseWilder(t *testing.T) {
	t.Parallel()

	const tol = 1e-3

	sar, err := NewParabolicStopAndReverse(&ParabolicStopAndReverseParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	highs := wilderHighs()
	lows := wilderLows()
	results := make([]float64, len(highs))

	for i := range highs {
		results[i] = sar.UpdateHL(highs[i], lows[i])
	}

	// Wilder spot checks from test_sar.c (TA_SAR, absolute values).
	// expectedBegIndex = 1, so output[0] corresponds to results[1].
	// TA_SAR always returns positive values, SAREXT returns signed.
	spotChecks := []struct {
		outIndex int     // index into output array (begIndex-relative)
		expected float64 // TA_SAR absolute expected value
	}{
		{0, 50.00},
		{1, 50.047},
		{4, 50.182},
		{35, 52.93},
		{36, 50.00},
	}

	for _, sc := range spotChecks {
		actual := math.Abs(results[sc.outIndex+1]) // +1 because results[0] = NaN
		diff := math.Abs(actual - sc.expected)

		if diff > tol {
			t.Errorf("Wilder spot check output[%d]: expected %.4f, got %.4f, diff %.6f",
				sc.outIndex, sc.expected, actual, diff)
		}
	}
}

func TestParabolicStopAndReverseIsPrimed(t *testing.T) {
	t.Parallel()

	sar, err := NewParabolicStopAndReverse(&ParabolicStopAndReverseParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sar.IsPrimed() {
		t.Error("expected not primed before any data")
	}

	// First bar — still not primed.
	sar.UpdateHL(93.25, 90.75)

	if sar.IsPrimed() {
		t.Error("expected not primed after 1 bar")
	}

	// Second bar — should be primed.
	sar.UpdateHL(94.94, 91.405)

	if !sar.IsPrimed() {
		t.Error("expected primed after 2 bars")
	}
}

func TestParabolicStopAndReverseMetadata(t *testing.T) {
	t.Parallel()

	sar, err := NewParabolicStopAndReverse(&ParabolicStopAndReverseParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	meta := sar.Metadata()

	if meta.Identifier != core.ParabolicStopAndReverse {
		t.Errorf("expected identifier ParabolicStopAndReverse, got %v", meta.Identifier)
	}

	if meta.Mnemonic != "sar()" {
		t.Errorf("expected mnemonic 'sar()', got '%s'", meta.Mnemonic)
	}

	if len(meta.Outputs) != 1 {
		t.Fatalf("expected 1 output, got %d", len(meta.Outputs))
	}

	if meta.Outputs[0].Kind != int(Value) {
		t.Errorf("expected output kind %d, got %d", Value, meta.Outputs[0].Kind)
	}

	if meta.Outputs[0].Shape != shape.Scalar {
		t.Errorf("expected output type ScalarType, got %v", meta.Outputs[0].Shape)
	}
}

func TestParabolicStopAndReverseConstructorValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params ParabolicStopAndReverseParams
		valid  bool
	}{
		{"defaults", ParabolicStopAndReverseParams{}, true},
		{"negative long init", ParabolicStopAndReverseParams{AccelerationInitLong: -0.01}, false},
		{"negative short step", ParabolicStopAndReverseParams{AccelerationShort: -0.01}, false},
		{"negative offset", ParabolicStopAndReverseParams{OffsetOnReverse: -0.01}, false},
		{"custom valid", ParabolicStopAndReverseParams{
			AccelerationInitLong:  0.01,
			AccelerationLong:      0.01,
			AccelerationMaxLong:   0.10,
			AccelerationInitShort: 0.03,
			AccelerationShort:     0.03,
			AccelerationMaxShort:  0.30,
		}, true},
		{"start value positive", ParabolicStopAndReverseParams{StartValue: 100.0}, true},
		{"start value negative", ParabolicStopAndReverseParams{StartValue: -100.0}, true},
	}

	for _, tt := range tests {
		_, err := NewParabolicStopAndReverse(&tt.params)
		if tt.valid && err != nil {
			t.Errorf("%s: unexpected error: %v", tt.name, err)
		}

		if !tt.valid && err == nil {
			t.Errorf("%s: expected error, got nil", tt.name)
		}
	}
}

func TestParabolicStopAndReverseUpdateBar(t *testing.T) {
	t.Parallel()

	sar, err := NewParabolicStopAndReverse(&ParabolicStopAndReverseParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	now := time.Now()

	bar1 := &entities.Bar{Time: now, Open: 91, High: 93.25, Low: 90.75, Close: 91.5, Volume: 1000}
	out1 := sar.UpdateBar(bar1)
	scalar1 := out1[0].(entities.Scalar)

	if !math.IsNaN(scalar1.Value) {
		t.Errorf("expected NaN for first bar, got %v", scalar1.Value)
	}

	bar2 := &entities.Bar{Time: now.Add(time.Minute), Open: 92, High: 94.94, Low: 91.405, Close: 94.815, Volume: 1000}
	out2 := sar.UpdateBar(bar2)
	scalar2 := out2[0].(entities.Scalar)

	if math.IsNaN(scalar2.Value) {
		t.Error("expected valid value for second bar, got NaN")
	}
}

func TestParabolicStopAndReverseNaN(t *testing.T) {
	t.Parallel()

	sar, err := NewParabolicStopAndReverse(&ParabolicStopAndReverseParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Feed two valid bars to prime.
	sar.UpdateHL(93.25, 90.75)
	sar.UpdateHL(94.94, 91.405)

	// Feed NaN — should not corrupt state.
	result := sar.UpdateHL(math.NaN(), 92.0)
	if !math.IsNaN(result) {
		t.Errorf("expected NaN for NaN input, got %v", result)
	}

	// Feed valid data — should still work.
	result = sar.UpdateHL(96.375, 94.25)
	if math.IsNaN(result) {
		t.Error("expected valid output after NaN, got NaN")
	}
}

func TestParabolicStopAndReverseForcedStartLong(t *testing.T) {
	t.Parallel()

	sar, err := NewParabolicStopAndReverse(&ParabolicStopAndReverseParams{
		StartValue: 85.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	highs := testHighs()
	lows := testLows()

	// First bar: NaN.
	result := sar.UpdateHL(highs[0], lows[0])
	if !math.IsNaN(result) {
		t.Errorf("expected NaN for first bar, got %v", result)
	}

	// Second bar: should be positive (long).
	result = sar.UpdateHL(highs[1], lows[1])
	if result <= 0 {
		t.Errorf("expected positive (long) SAR with forced long start, got %v", result)
	}
}

func TestParabolicStopAndReverseForcedStartShort(t *testing.T) {
	t.Parallel()

	sar, err := NewParabolicStopAndReverse(&ParabolicStopAndReverseParams{
		StartValue: -100.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	highs := testHighs()
	lows := testLows()

	// First bar: NaN.
	result := sar.UpdateHL(highs[0], lows[0])
	if !math.IsNaN(result) {
		t.Errorf("expected NaN for first bar, got %v", result)
	}

	// Second bar: should be negative (short).
	result = sar.UpdateHL(highs[1], lows[1])
	if result >= 0 {
		t.Errorf("expected negative (short) SAR with forced short start, got %v", result)
	}
}
