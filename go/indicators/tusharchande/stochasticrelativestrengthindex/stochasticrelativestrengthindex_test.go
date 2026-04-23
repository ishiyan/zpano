//nolint:testpackage
package stochasticrelativestrengthindex

//nolint: gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

// Test data from TA-Lib (252 entries).
func testInput() []float64 {
	return []float64{
		91.500000, 94.815000, 94.375000, 95.095000, 93.780000, 94.625000, 92.530000, 92.750000, 90.315000, 92.470000, 96.125000,
		97.250000, 98.500000, 89.875000, 91.000000, 92.815000, 89.155000, 89.345000, 91.625000, 89.875000, 88.375000, 87.625000,
		84.780000, 83.000000, 83.500000, 81.375000, 84.440000, 89.250000, 86.375000, 86.250000, 85.250000, 87.125000, 85.815000,
		88.970000, 88.470000, 86.875000, 86.815000, 84.875000, 84.190000, 83.875000, 83.375000, 85.500000, 89.190000, 89.440000,
		91.095000, 90.750000, 91.440000, 89.000000, 91.000000, 90.500000, 89.030000, 88.815000, 84.280000, 83.500000, 82.690000,
		84.750000, 85.655000, 86.190000, 88.940000, 89.280000, 88.625000, 88.500000, 91.970000, 91.500000, 93.250000, 93.500000,
		93.155000, 91.720000, 90.000000, 89.690000, 88.875000, 85.190000, 83.375000, 84.875000, 85.940000, 97.250000, 99.875000,
		104.940000, 106.000000, 102.500000, 102.405000, 104.595000, 106.125000, 106.000000, 106.065000, 104.625000, 108.625000,
		109.315000, 110.500000, 112.750000, 123.000000, 119.625000, 118.750000, 119.250000, 117.940000, 116.440000, 115.190000,
		111.875000, 110.595000, 118.125000, 116.000000, 116.000000, 112.000000, 113.750000, 112.940000, 116.000000, 120.500000,
		116.620000, 117.000000, 115.250000, 114.310000, 115.500000, 115.870000, 120.690000, 120.190000, 120.750000, 124.750000,
		123.370000, 122.940000, 122.560000, 123.120000, 122.560000, 124.620000, 129.250000, 131.000000, 132.250000, 131.000000,
		132.810000, 134.000000, 137.380000, 137.810000, 137.880000, 137.250000, 136.310000, 136.250000, 134.630000, 128.250000,
		129.000000, 123.870000, 124.810000, 123.000000, 126.250000, 128.380000, 125.370000, 125.690000, 122.250000, 119.370000,
		118.500000, 123.190000, 123.500000, 122.190000, 119.310000, 123.310000, 121.120000, 123.370000, 127.370000, 128.500000,
		123.870000, 122.940000, 121.750000, 124.440000, 122.000000, 122.370000, 122.940000, 124.000000, 123.190000, 124.560000,
		127.250000, 125.870000, 128.860000, 132.000000, 130.750000, 134.750000, 135.000000, 132.380000, 133.310000, 131.940000,
		130.000000, 125.370000, 130.130000, 127.120000, 125.190000, 122.000000, 125.000000, 123.000000, 123.500000, 120.060000,
		121.000000, 117.750000, 119.870000, 122.000000, 119.190000, 116.370000, 113.500000, 114.250000, 110.000000, 105.060000,
		107.000000, 107.870000, 107.000000, 107.120000, 107.000000, 91.000000, 93.940000, 93.870000, 95.500000, 93.000000,
		94.940000, 98.250000, 96.750000, 94.810000, 94.370000, 91.560000, 90.250000, 93.940000, 93.620000, 97.000000, 95.000000,
		95.870000, 94.060000, 94.620000, 93.750000, 98.000000, 103.940000, 107.870000, 106.060000, 104.500000, 105.000000,
		104.190000, 103.060000, 103.420000, 105.270000, 111.870000, 116.000000, 116.620000, 118.280000, 113.370000, 109.000000,
		109.700000, 109.250000, 107.000000, 109.190000, 110.000000, 109.200000, 110.120000, 108.000000, 108.620000, 109.750000,
		109.810000, 109.000000, 108.750000, 107.870000,
	}
}

func testTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

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
