//nolint:testpackage
package stochasticrelativestrengthindex

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

// Test case 1: period=14, fastK=14, fastD=1, SMA.
// begIndex=27, first value: FastK=94.156709, FastD=94.156709.
// last value (index 251): FastK=0.0, FastD=0.0.
func TestStochasticRSI_14_14_1_SMA(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-4

	input := testInput()

	ind, err := NewStochasticRelativeStrengthIndex(&StochasticRelativeStrengthIndexParams{
		Length:      14,
		FastKLength: 14,
		FastDLength: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	// First 27 values should produce NaN for FastK.
	for i := 0; i < 27; i++ {
		fastK, _ := ind.Update(input[i])
		if !math.IsNaN(fastK) {
			t.Errorf("[%d] expected NaN FastK, got %v", i, fastK)
		}
	}

	// Index 27: first value.
	fastK, fastD := ind.Update(input[27])
	if math.IsNaN(fastK) {
		t.Errorf("[27] expected non-NaN FastK, got NaN")
	}

	if math.Abs(fastK-94.156709) > tolerance {
		t.Errorf("[27] FastK: expected ~94.156709, got %v", fastK)
	}

	if math.Abs(fastD-94.156709) > tolerance {
		t.Errorf("[27] FastD: expected ~94.156709, got %v", fastD)
	}

	// Feed remaining and check last value.
	for i := 28; i < 251; i++ {
		ind.Update(input[i])
	}

	fastK, fastD = ind.Update(input[251])

	if math.Abs(fastK-0.0) > tolerance {
		t.Errorf("[251] FastK: expected ~0.0, got %v", fastK)
	}

	if math.Abs(fastD-0.0) > tolerance {
		t.Errorf("[251] FastD: expected ~0.0, got %v", fastD)
	}
}

// Test case 2: period=14, fastK=45, fastD=1, SMA.
// begIndex=58, first value: FastK=79.729186, FastD=79.729186.
// last value (index 251): FastK=48.1550743, FastD=48.1550743.
func TestStochasticRSI_14_45_1_SMA(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-4

	input := testInput()

	ind, err := NewStochasticRelativeStrengthIndex(&StochasticRelativeStrengthIndexParams{
		Length:      14,
		FastKLength: 45,
		FastDLength: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	// First 58 values should produce NaN for FastK.
	for i := 0; i < 58; i++ {
		fastK, _ := ind.Update(input[i])
		if !math.IsNaN(fastK) {
			t.Errorf("[%d] expected NaN FastK, got %v", i, fastK)
		}
	}

	// Index 58: first value.
	fastK, fastD := ind.Update(input[58])
	if math.IsNaN(fastK) {
		t.Errorf("[58] expected non-NaN FastK, got NaN")
	}

	if math.Abs(fastK-79.729186) > tolerance {
		t.Errorf("[58] FastK: expected ~79.729186, got %v", fastK)
	}

	if math.Abs(fastD-79.729186) > tolerance {
		t.Errorf("[58] FastD: expected ~79.729186, got %v", fastD)
	}

	// Feed remaining and check last value.
	for i := 59; i < 251; i++ {
		ind.Update(input[i])
	}

	fastK, fastD = ind.Update(input[251])

	if math.Abs(fastK-48.1550743) > tolerance {
		t.Errorf("[251] FastK: expected ~48.1550743, got %v", fastK)
	}

	if math.Abs(fastD-48.1550743) > tolerance {
		t.Errorf("[251] FastD: expected ~48.1550743, got %v", fastD)
	}
}

// Test case 3: period=11, fastK=13, fastD=16, SMA.
// begIndex=38, first value: FastK=5.25947, FastD=57.1711.
// last value (index 251): FastK=0.0, FastD=15.7303.
func TestStochasticRSI_11_13_16_SMA(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-3

	input := testInput()

	ind, err := NewStochasticRelativeStrengthIndex(&StochasticRelativeStrengthIndexParams{
		Length:      11,
		FastKLength: 13,
		FastDLength: 16,
	})
	if err != nil {
		t.Fatal(err)
	}

	// begIndex=38: RSI lookback(11)=11, STOCHF lookback(13, 16, SMA) = 12 + 15 = 27, total = 38.
	// But FastD (SMA(16)) won't be primed until 16 FastK values are collected.
	// First primed at: RSI primes after 11+1=12 inputs (indices 0..11, first RSI at index 11).
	// Wait, RSI lookback = 11 means first RSI at index 11.
	// FastK needs 13 RSI values: first FastK at index 11+12 = 23.
	// FastD SMA(16) needs 16 FastK values: first FastD at index 23+15 = 38.
	// So begIndex=38: FastD is first valid here.

	// Feed first 38 values.
	for i := 0; i < 38; i++ {
		ind.Update(input[i])
	}

	// Index 38: first primed value.
	fastK, fastD := ind.Update(input[38])

	if math.Abs(fastK-5.25947) > tolerance {
		t.Errorf("[38] FastK: expected ~5.25947, got %v", fastK)
	}

	if math.Abs(fastD-57.1711) > tolerance {
		t.Errorf("[38] FastD: expected ~57.1711, got %v", fastD)
	}

	if !ind.IsPrimed() {
		t.Error("expected primed at index 38")
	}

	// Feed remaining and check last value.
	for i := 39; i < 251; i++ {
		ind.Update(input[i])
	}

	fastK, fastD = ind.Update(input[251])

	if math.Abs(fastK-0.0) > tolerance {
		t.Errorf("[251] FastK: expected ~0.0, got %v", fastK)
	}

	if math.Abs(fastD-15.7303) > tolerance {
		t.Errorf("[251] FastD: expected ~15.7303, got %v", fastD)
	}
}

func TestStochasticRSIIsPrimed(t *testing.T) {
	t.Parallel()

	ind, err := NewStochasticRelativeStrengthIndex(&StochasticRelativeStrengthIndexParams{
		Length:      14,
		FastKLength: 14,
		FastDLength: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	input := testInput()

	if ind.IsPrimed() {
		t.Error("expected not primed initially")
	}

	for i := 0; i < 27; i++ {
		ind.Update(input[i])
		if ind.IsPrimed() {
			t.Errorf("[%d] expected not primed", i)
		}
	}

	ind.Update(input[27])

	if !ind.IsPrimed() {
		t.Error("expected primed after index 27")
	}
}

func TestStochasticRSINaN(t *testing.T) {
	t.Parallel()

	ind, err := NewStochasticRelativeStrengthIndex(&StochasticRelativeStrengthIndexParams{
		Length:      14,
		FastKLength: 14,
		FastDLength: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	fastK, fastD := ind.Update(math.NaN())
	if !math.IsNaN(fastK) {
		t.Errorf("expected NaN FastK, got %v", fastK)
	}

	if !math.IsNaN(fastD) {
		t.Errorf("expected NaN FastD, got %v", fastD)
	}
}

func TestStochasticRSIMetadata(t *testing.T) {
	t.Parallel()

	ind, err := NewStochasticRelativeStrengthIndex(&StochasticRelativeStrengthIndexParams{
		Length:      14,
		FastKLength: 14,
		FastDLength: 3,
	})
	if err != nil {
		t.Fatal(err)
	}

	meta := ind.Metadata()

	if meta.Identifier != core.StochasticRelativeStrengthIndex {
		t.Errorf("expected identifier StochasticRelativeStrengthIndex, got %v", meta.Identifier)
	}

	exp := "stochrsi(14/14/SMA3)"
	if meta.Mnemonic != exp {
		t.Errorf("expected mnemonic '%s', got '%s'", exp, meta.Mnemonic)
	}

	if len(meta.Outputs) != 2 {
		t.Fatalf("expected 2 outputs, got %d", len(meta.Outputs))
	}

	if meta.Outputs[0].Kind != int(FastK) {
		t.Errorf("expected output 0 kind %d, got %d", FastK, meta.Outputs[0].Kind)
	}

	if meta.Outputs[0].Shape != shape.Scalar {
		t.Errorf("expected scalar output type, got %v", meta.Outputs[0].Shape)
	}

	if meta.Outputs[1].Kind != int(FastD) {
		t.Errorf("expected output 1 kind %d, got %d", FastD, meta.Outputs[1].Kind)
	}
}

func TestStochasticRSIUpdateEntity(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-4

	input := testInput()

	ind, err := NewStochasticRelativeStrengthIndex(&StochasticRelativeStrengthIndexParams{
		Length:      14,
		FastKLength: 14,
		FastDLength: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	tm := testTime()

	for i := 0; i < 27; i++ {
		scalar := &entities.Scalar{Time: tm, Value: input[i]}
		out := ind.UpdateScalar(scalar)

		v := out[0].(entities.Scalar).Value //nolint:forcetypeassert
		if !math.IsNaN(v) {
			t.Errorf("[%d] expected NaN, got %v", i, v)
		}
	}

	scalar := &entities.Scalar{Time: tm, Value: input[27]}
	out := ind.UpdateScalar(scalar)

	fastK := out[0].(entities.Scalar).Value //nolint:forcetypeassert
	fastD := out[1].(entities.Scalar).Value //nolint:forcetypeassert

	if math.Abs(fastK-94.156709) > tolerance {
		t.Errorf("[27] FastK: expected ~94.156709, got %v", fastK)
	}

	if math.Abs(fastD-94.156709) > tolerance {
		t.Errorf("[27] FastD: expected ~94.156709, got %v", fastD)
	}
}

func TestStochasticRSIInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                             string
		length, fastKLength, fastDLength int
	}{
		{"length too small", 1, 14, 3},
		{"fastK too small", 14, 0, 3},
		{"fastD too small", 14, 14, 0},
		{"length negative", -1, 14, 3},
	}

	for _, tt := range tests {
		_, err := NewStochasticRelativeStrengthIndex(&StochasticRelativeStrengthIndexParams{
			Length:      tt.length,
			FastKLength: tt.fastKLength,
			FastDLength: tt.fastDLength,
		})
		if err == nil {
			t.Errorf("%s: expected error, got nil", tt.name)
		}
	}
}
