//nolint:testpackage
package commoditychannelindex

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

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

	if meta.Identifier != core.CommodityChannelIndex {
		t.Errorf("expected identifier CommodityChannelIndex, got %v", meta.Identifier)
	}

	exp := "cci(20)"
	if meta.Mnemonic != exp {
		t.Errorf("expected mnemonic '%s', got '%s'", exp, meta.Mnemonic)
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
