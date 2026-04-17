//nolint:testpackage
package commoditychannelindex

//nolint: gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
)

// Test data from TA-Lib (252 entries), used by MBST C# tests.
// Typical price input: test_CCI.xsl, TYPPRICE, F4…F255.
func testInput() []float64 {
	return []float64{
		91.83333333333330, 93.72000000000000, 95.00000000000000, 94.92833333333330, 94.19833333333330, 94.28166666666670, 93.17666666666670, 92.07333333333330, 90.74166666666670, 91.94833333333330,
		95.04166666666670, 97.73000000000000, 97.88500000000000, 90.48000000000000, 89.68833333333330, 92.33500000000000, 90.48833333333330, 89.59333333333330, 90.94833333333330, 90.62500000000000,
		88.74000000000000, 87.55166666666670, 85.88500000000000, 83.59333333333330, 83.16833333333330, 82.33333333333330, 83.37666666666670, 87.57333333333330, 86.69833333333330, 87.11500000000000,
		85.60333333333330, 86.49000000000000, 86.23000000000000, 87.82333333333330, 88.78166666666670, 87.76166666666670, 86.14666666666670, 84.68833333333330, 83.83500000000000, 84.26166666666670,
		83.45833333333330, 86.35500000000000, 88.51166666666670, 89.32333333333330, 90.93833333333330, 90.77166666666670, 91.72000000000000, 89.90666666666670, 90.24000000000000, 90.78166666666670,
		89.34333333333330, 88.05333333333330, 85.76000000000000, 84.02166666666670, 82.83500000000000, 84.41666666666670, 85.67666666666670, 86.47000000000000, 88.50166666666670, 89.44833333333330,
		89.20833333333330, 88.23000000000000, 91.07333333333330, 91.99000000000000, 92.19833333333330, 92.89500000000000, 93.06166666666670, 91.35500000000000, 90.65666666666670, 90.14833333333330,
		88.45833333333330, 86.33500000000000, 83.85333333333330, 83.75000000000000, 84.81500000000000, 97.65666666666670, 99.87500000000000, 103.82333333333300, 105.95833333333300, 103.16666666666700,
		102.87500000000000, 103.93833333333300, 105.13500000000000, 106.54333333333300, 105.32333333333300, 105.20833333333300, 107.63500000000000, 109.59500000000000, 110.06333333333300, 111.57333333333300,
		121.00000000000000, 119.79166666666700, 118.18833333333300, 119.35500000000000, 117.94833333333300, 116.96000000000000, 115.49166666666700, 112.69833333333300, 111.36500000000000, 115.72000000000000,
		115.16333333333300, 115.64666666666700, 112.35333333333300, 112.60333333333300, 113.27000000000000, 114.81333333333300, 119.89666666666700, 117.51666666666700, 118.14333333333300, 115.10333333333300,
		114.45666666666700, 115.16666666666700, 116.31000000000000, 120.35333333333300, 120.39666666666700, 120.64666666666700, 124.37333333333300, 124.70666666666700, 122.96000000000000, 122.85333333333300,
		123.99666666666700, 123.14666666666700, 124.22666666666700, 128.54000000000000, 130.10333333333300, 131.33333333333300, 131.89666666666700, 132.31333333333300, 133.87666666666700, 136.23333333333300,
		137.29333333333300, 137.60666666666700, 137.31333333333300, 136.31333333333300, 136.37666666666700, 135.73333333333300, 128.81333333333300, 128.54000000000000, 125.29000000000000, 124.29000000000000,
		123.62333333333300, 125.43666666666700, 127.62666666666700, 125.53666666666700, 125.58333333333300, 123.35333333333300, 120.22666666666700, 119.47666666666700, 121.58333333333300, 123.83333333333300,
		122.58333333333300, 120.25000000000000, 122.29000000000000, 121.97666666666700, 123.29000000000000, 126.58000000000000, 127.87333333333300, 125.66666666666700, 123.02000000000000, 122.39333333333300,
		123.87333333333300, 122.20666666666700, 122.43333333333300, 123.62333333333300, 123.98000000000000, 123.83333333333300, 124.47666666666700, 127.08333333333300, 125.62333333333300, 128.87000000000000,
		131.02333333333300, 131.79333333333300, 134.29333333333300, 135.69000000000000, 133.31333333333300, 132.93666666666700, 132.96000000000000, 130.64666666666700, 126.85333333333300, 129.00333333333300,
		127.66666666666700, 125.60333333333300, 123.75000000000000, 123.48000000000000, 123.72666666666700, 123.31333333333300, 120.95666666666700, 120.95666666666700, 118.10333333333300, 118.87333333333300,
		121.43666666666700, 120.33333333333300, 116.87333333333300, 113.20666666666700, 114.68666666666700, 111.62333333333300, 106.97666666666700, 106.31333333333300, 106.87000000000000, 106.89666666666700,
		107.02000000000000, 109.02000000000000, 91.00000000000000, 93.68666666666670, 93.70333333333330, 95.37333333333330, 93.79000000000000, 94.83333333333330, 97.83333333333330, 97.31000000000000,
		95.10333333333330, 94.60333333333330, 92.00000000000000, 91.12666666666670, 92.79333333333330, 93.74666666666670, 96.06000000000000, 95.79000000000000, 95.04000000000000, 94.76666666666670,
		94.20666666666670, 93.74666666666670, 96.60333333333330, 102.47666666666700, 106.91666666666700, 107.31000000000000, 103.77000000000000, 105.04000000000000, 104.16666666666700, 103.22666666666700,
		103.37000000000000, 104.98333333333300, 110.89333333333300, 115.00000000000000, 117.08333333333300, 118.26000000000000, 115.91333333333300, 109.50000000000000, 109.67000000000000, 108.77000000000000,
		106.48000000000000, 108.21000000000000, 109.89333333333300, 109.13000000000000, 109.43333333333300, 108.77000000000000, 109.08333333333300, 109.29000000000000, 109.87333333333300, 109.41666666666700,
		109.27000000000000, 107.99666666666700,
	}
}

func testTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

func TestCommodityChannelIndexLength11(t *testing.T) {
	t.Parallel()

	const tolerance = 5e-8

	input := testInput()

	cci, err := NewCommodityChannelIndex(&CommodityChannelIndexParams{
		Length: 11,
	})
	if err != nil {
		t.Fatal(err)
	}

	// First 10 values should be NaN.
	for i := 0; i < 10; i++ {
		v := cci.Update(input[i])
		if !math.IsNaN(v) {
			t.Errorf("[%d] expected NaN, got %v", i, v)
		}
	}

	// Index 10: first value.
	v := cci.Update(input[10])
	if math.IsNaN(v) {
		t.Errorf("[10] expected non-NaN, got NaN")
	}

	if math.Abs(v-87.92686612269590) > tolerance {
		t.Errorf("[10] expected ~87.92686612269590, got %v", v)
	}

	// Index 11.
	v = cci.Update(input[11])
	if math.Abs(v-180.00543014506300) > tolerance {
		t.Errorf("[11] expected ~180.00543014506300, got %v", v)
	}

	// Feed remaining and check last.
	for i := 12; i < 251; i++ {
		cci.Update(input[i])
	}

	v = cci.Update(input[251])
	if math.Abs(v-(-169.65514382823800)) > tolerance {
		t.Errorf("[251] expected ~-169.65514382823800, got %v", v)
	}

	if !cci.IsPrimed() {
		t.Error("expected primed")
	}
}

func TestCommodityChannelIndexLength2(t *testing.T) {
	t.Parallel()

	const tolerance = 5e-7

	input := testInput()

	cci, err := NewCommodityChannelIndex(&CommodityChannelIndexParams{
		Length: 2,
	})
	if err != nil {
		t.Fatal(err)
	}

	// First value should be NaN.
	v := cci.Update(input[0])
	if !math.IsNaN(v) {
		t.Errorf("[0] expected NaN, got %v", v)
	}

	// Index 1: first value.
	v = cci.Update(input[1])
	if math.IsNaN(v) {
		t.Errorf("[1] expected non-NaN, got NaN")
	}

	if math.Abs(v-66.66666666666670) > tolerance {
		t.Errorf("[1] expected ~66.66666666666670, got %v", v)
	}

	// Feed remaining and check last.
	for i := 2; i < 251; i++ {
		cci.Update(input[i])
	}

	v = cci.Update(input[251])
	if math.Abs(v-(-66.66666666666590)) > tolerance {
		t.Errorf("[251] expected ~-66.66666666666590, got %v", v)
	}
}

func TestCommodityChannelIndexIsPrimed(t *testing.T) {
	t.Parallel()

	cci, err := NewCommodityChannelIndex(&CommodityChannelIndexParams{
		Length: 5,
	})
	if err != nil {
		t.Fatal(err)
	}

	if cci.IsPrimed() {
		t.Error("expected not primed initially")
	}

	for i := 1; i <= 4; i++ {
		cci.Update(float64(i))
		if cci.IsPrimed() {
			t.Errorf("[%d] expected not primed", i)
		}
	}

	cci.Update(5)

	if !cci.IsPrimed() {
		t.Error("expected primed after 5 samples")
	}

	cci.Update(6)

	if !cci.IsPrimed() {
		t.Error("expected still primed after 6 samples")
	}
}

func TestCommodityChannelIndexNaN(t *testing.T) {
	t.Parallel()

	cci, err := NewCommodityChannelIndex(&CommodityChannelIndexParams{
		Length: 5,
	})
	if err != nil {
		t.Fatal(err)
	}

	v := cci.Update(math.NaN())
	if !math.IsNaN(v) {
		t.Errorf("expected NaN, got %v", v)
	}
}

func TestCommodityChannelIndexMetadata(t *testing.T) {
	t.Parallel()

	cci, err := NewCommodityChannelIndex(&CommodityChannelIndexParams{
		Length: 20,
	})
	if err != nil {
		t.Fatal(err)
	}

	meta := cci.Metadata()

	if meta.Type != core.CommodityChannelIndex {
		t.Errorf("expected type CommodityChannelIndex, got %v", meta.Type)
	}

	exp := "cci(20)"
	if meta.Mnemonic != exp {
		t.Errorf("expected mnemonic '%s', got '%s'", exp, meta.Mnemonic)
	}

	if len(meta.Outputs) != 1 {
		t.Fatalf("expected 1 output, got %d", len(meta.Outputs))
	}

	if meta.Outputs[0].Kind != int(CommodityChannelIndexValue) {
		t.Errorf("expected output kind %d, got %d", CommodityChannelIndexValue, meta.Outputs[0].Kind)
	}

	if meta.Outputs[0].Type != outputs.ScalarType {
		t.Errorf("expected scalar output type, got %v", meta.Outputs[0].Type)
	}
}

func TestCommodityChannelIndexUpdateEntity(t *testing.T) {
	t.Parallel()

	input := testInput()

	cci, err := NewCommodityChannelIndex(&CommodityChannelIndexParams{
		Length: 11,
	})
	if err != nil {
		t.Fatal(err)
	}

	tm := testTime()

	for i := 0; i < 10; i++ {
		scalar := &entities.Scalar{Time: tm, Value: input[i]}
		out := cci.UpdateScalar(scalar)

		v := out[0].(entities.Scalar).Value //nolint:forcetypeassert
		if !math.IsNaN(v) {
			t.Errorf("[%d] expected NaN, got %v", i, v)
		}
	}

	scalar := &entities.Scalar{Time: tm, Value: input[10]}
	out := cci.UpdateScalar(scalar)

	v := out[0].(entities.Scalar).Value //nolint:forcetypeassert
	if math.IsNaN(v) {
		t.Errorf("[10] expected non-NaN, got NaN")
	}
}

func TestCommodityChannelIndexInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		length int
	}{
		{"length too small", 1},
		{"length zero", 0},
		{"length negative", -8},
	}

	for _, tt := range tests {
		_, err := NewCommodityChannelIndex(&CommodityChannelIndexParams{
			Length: tt.length,
		})
		if err == nil {
			t.Errorf("%s: expected error, got nil", tt.name)
		}
	}
}

func TestCommodityChannelIndexCustomScalingFactor(t *testing.T) {
	t.Parallel()

	// With custom inverse scaling factor, values should scale differently.
	cci, err := NewCommodityChannelIndex(&CommodityChannelIndexParams{
		Length:               5,
		InverseScalingFactor: 0.03,
	})
	if err != nil {
		t.Fatal(err)
	}

	for i := 1; i <= 5; i++ {
		cci.Update(float64(i))
	}

	if !cci.IsPrimed() {
		t.Error("expected primed")
	}
}
