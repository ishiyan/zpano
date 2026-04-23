//nolint:testpackage
package bollingerbandstrend

//nolint:gofumpt
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

func TestBollingerBandsTrend_SampleStdDev_FullData(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-8

	closingPrice := testClosingPrice()
	expected := testSampleExpected()

	boolTrue := true
	ind, err := NewBollingerBandsTrend(&BollingerBandsTrendParams{
		FastLength: 20,
		SlowLength: 50,
		IsUnbiased: &boolTrue,
	})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 252; i++ {
		v := ind.Update(closingPrice[i])

		if math.IsNaN(expected[i]) {
			if !math.IsNaN(v) {
				t.Errorf("[%d] expected NaN, got %v", i, v)
			}

			continue
		}

		if math.Abs(v-expected[i]) > tolerance {
			t.Errorf("[%d] expected %v, got %v", i, expected[i], v)
		}
	}
}

func TestBollingerBandsTrend_PopulationStdDev_FullData(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-8

	closingPrice := testClosingPrice()
	expected := testPopulationExpected()

	boolFalse := false
	ind, err := NewBollingerBandsTrend(&BollingerBandsTrendParams{
		FastLength: 20,
		SlowLength: 50,
		IsUnbiased: &boolFalse,
	})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 252; i++ {
		v := ind.Update(closingPrice[i])

		if math.IsNaN(expected[i]) {
			if !math.IsNaN(v) {
				t.Errorf("[%d] expected NaN, got %v", i, v)
			}

			continue
		}

		if math.Abs(v-expected[i]) > tolerance {
			t.Errorf("[%d] expected %v, got %v", i, expected[i], v)
		}
	}
}

func TestBollingerBandsTrendIsPrimed(t *testing.T) {
	t.Parallel()

	ind, err := NewBollingerBandsTrend(&BollingerBandsTrendParams{
		FastLength: 20,
		SlowLength: 50,
	})
	if err != nil {
		t.Fatal(err)
	}

	closingPrice := testClosingPrice()

	if ind.IsPrimed() {
		t.Error("expected not primed initially")
	}

	for i := 0; i < 49; i++ {
		ind.Update(closingPrice[i])
		if ind.IsPrimed() {
			t.Errorf("[%d] expected not primed", i)
		}
	}

	ind.Update(closingPrice[49])

	if !ind.IsPrimed() {
		t.Error("expected primed after index 49")
	}
}

func TestBollingerBandsTrendNaN(t *testing.T) {
	t.Parallel()

	ind, err := NewBollingerBandsTrend(&BollingerBandsTrendParams{
		FastLength: 20,
		SlowLength: 50,
	})
	if err != nil {
		t.Fatal(err)
	}

	v := ind.Update(math.NaN())
	if !math.IsNaN(v) {
		t.Errorf("expected NaN, got %v", v)
	}
}

func TestBollingerBandsTrendMetadata(t *testing.T) {
	t.Parallel()

	ind, err := NewBollingerBandsTrend(&BollingerBandsTrendParams{
		FastLength: 20,
		SlowLength: 50,
	})
	if err != nil {
		t.Fatal(err)
	}

	meta := ind.Metadata()

	if meta.Identifier != core.BollingerBandsTrend {
		t.Errorf("expected identifier BollingerBandsTrend, got %v", meta.Identifier)
	}

	const expectedOutputs = 1
	if len(meta.Outputs) != expectedOutputs {
		t.Fatalf("expected %d outputs, got %d", expectedOutputs, len(meta.Outputs))
	}

	if meta.Outputs[0].Kind != int(Value) {
		t.Errorf("expected output 0 kind %d, got %d", Value, meta.Outputs[0].Kind)
	}

	if meta.Outputs[0].Shape != shape.Scalar {
		t.Errorf("expected scalar output type, got %v", meta.Outputs[0].Shape)
	}
}

func TestBollingerBandsTrendUpdateScalar(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-8

	closingPrice := testClosingPrice()
	expected := testSampleExpected()

	boolTrue := true
	ind, err := NewBollingerBandsTrend(&BollingerBandsTrendParams{
		FastLength: 20,
		SlowLength: 50,
		IsUnbiased: &boolTrue,
	})
	if err != nil {
		t.Fatal(err)
	}

	tm := testTime()

	// Feed first 49 samples — all NaN.
	for i := 0; i < 49; i++ {
		out := ind.UpdateScalar(&entities.Scalar{Time: tm, Value: closingPrice[i]})

		v := out[0].(entities.Scalar).Value //nolint:forcetypeassert
		if !math.IsNaN(v) {
			t.Errorf("[%d] expected NaN scalar, got %v", i, v)
		}
	}

	// Feed index 49 — first primed value.
	out := ind.UpdateScalar(&entities.Scalar{Time: tm, Value: closingPrice[49]})

	v := out[0].(entities.Scalar).Value //nolint:forcetypeassert

	if math.Abs(v-expected[49]) > tolerance {
		t.Errorf("[49] expected %v, got %v", expected[49], v)
	}
}

func TestBollingerBandsTrendDefaultParams(t *testing.T) {
	t.Parallel()

	// Default fast=20, slow=50.
	ind, err := NewBollingerBandsTrend(&BollingerBandsTrendParams{})
	if err != nil {
		t.Fatal(err)
	}

	closingPrice := testClosingPrice()

	// Feed 49 samples, should not be primed.
	for i := 0; i < 49; i++ {
		ind.Update(closingPrice[i])
	}

	if ind.IsPrimed() {
		t.Error("expected not primed after 49 samples")
	}

	// 50th sample should prime it.
	ind.Update(closingPrice[49])

	if !ind.IsPrimed() {
		t.Error("expected primed after 50 samples")
	}
}

func TestBollingerBandsTrendInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		fastLength int
		slowLength int
	}{
		{"fast length too small", 1, 50},
		{"fast length negative", -1, 50},
		{"slow length too small", 20, 1},
		{"slow length negative", 20, -1},
		{"slow not greater than fast", 20, 20},
		{"slow less than fast", 50, 20},
	}

	for _, tt := range tests {
		_, err := NewBollingerBandsTrend(&BollingerBandsTrendParams{
			FastLength: tt.fastLength,
			SlowLength: tt.slowLength,
		})
		if err == nil {
			t.Errorf("%s: expected error, got nil", tt.name)
		}
	}
}
