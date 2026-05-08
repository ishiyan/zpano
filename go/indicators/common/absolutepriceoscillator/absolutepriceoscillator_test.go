//nolint:testpackage
package absolutepriceoscillator

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

func TestAbsolutePriceOscillatorSMA1226(t *testing.T) {
	t.Parallel()

	const tolerance = 5e-4

	input := testInput()

	apo, err := NewAbsolutePriceOscillator(&AbsolutePriceOscillatorParams{
		FastLength: 12,
		SlowLength: 26,
	})
	if err != nil {
		t.Fatal(err)
	}

	// First 25 values should be NaN.
	for i := 0; i < 25; i++ {
		v := apo.Update(input[i])
		if !math.IsNaN(v) {
			t.Errorf("[%d] expected NaN, got %v", i, v)
		}
	}

	// Index 25: first value.
	v := apo.Update(input[25])
	if math.IsNaN(v) {
		t.Errorf("[25] expected non-NaN, got NaN")
	}

	if math.Abs(v-(-3.3124)) > tolerance {
		t.Errorf("[25] expected ~-3.3124, got %v", v)
	}

	// Index 26: second value.
	v = apo.Update(input[26])
	if math.Abs(v-(-3.5876)) > tolerance {
		t.Errorf("[26] expected ~-3.5876, got %v", v)
	}

	// Feed remaining and check last.
	for i := 27; i < 251; i++ {
		apo.Update(input[i])
	}

	v = apo.Update(input[251])
	if math.Abs(v-(-0.1667)) > tolerance {
		t.Errorf("[251] expected ~-0.1667, got %v", v)
	}

	if !apo.IsPrimed() {
		t.Error("expected primed")
	}
}

func TestAbsolutePriceOscillatorEMA1226(t *testing.T) {
	t.Parallel()

	const tolerance = 5e-4

	input := testInput()

	apo, err := NewAbsolutePriceOscillator(&AbsolutePriceOscillatorParams{
		FastLength:        12,
		SlowLength:        26,
		MovingAverageType: EMA,
		FirstIsAverage:    false,
	})
	if err != nil {
		t.Fatal(err)
	}

	// First 25 values should be NaN.
	for i := 0; i < 25; i++ {
		v := apo.Update(input[i])
		if !math.IsNaN(v) {
			t.Errorf("[%d] expected NaN, got %v", i, v)
		}
	}

	// Index 25: first value.
	v := apo.Update(input[25])
	if math.IsNaN(v) {
		t.Errorf("[25] expected non-NaN, got NaN")
	}

	if math.Abs(v-(-2.4193)) > tolerance {
		t.Errorf("[25] expected ~-2.4193, got %v", v)
	}

	// Index 26: second value.
	v = apo.Update(input[26])
	if math.Abs(v-(-2.4367)) > tolerance {
		t.Errorf("[26] expected ~-2.4367, got %v", v)
	}

	// Feed remaining and check last.
	for i := 27; i < 251; i++ {
		apo.Update(input[i])
	}

	v = apo.Update(input[251])
	if math.Abs(v-0.90401) > tolerance {
		t.Errorf("[251] expected ~0.90401, got %v", v)
	}
}

func TestAbsolutePriceOscillatorIsPrimed(t *testing.T) {
	t.Parallel()

	apo, err := NewAbsolutePriceOscillator(&AbsolutePriceOscillatorParams{
		FastLength: 3,
		SlowLength: 5,
	})
	if err != nil {
		t.Fatal(err)
	}

	if apo.IsPrimed() {
		t.Error("expected not primed initially")
	}

	for i := 1; i < 5; i++ {
		apo.Update(float64(i))
		if apo.IsPrimed() {
			t.Errorf("[%d] expected not primed", i)
		}
	}

	apo.Update(5)

	if !apo.IsPrimed() {
		t.Error("expected primed after 5 samples")
	}

	for i := 6; i < 10; i++ {
		apo.Update(float64(i))
		if !apo.IsPrimed() {
			t.Errorf("[%d] expected primed", i)
		}
	}
}

func TestAbsolutePriceOscillatorNaN(t *testing.T) {
	t.Parallel()

	apo, err := NewAbsolutePriceOscillator(&AbsolutePriceOscillatorParams{
		FastLength: 2,
		SlowLength: 3,
	})
	if err != nil {
		t.Fatal(err)
	}

	v := apo.Update(math.NaN())
	if !math.IsNaN(v) {
		t.Errorf("expected NaN, got %v", v)
	}
}

func TestAbsolutePriceOscillatorMetadata(t *testing.T) {
	t.Parallel()

	apo, err := NewAbsolutePriceOscillator(&AbsolutePriceOscillatorParams{
		FastLength: 12,
		SlowLength: 26,
	})
	if err != nil {
		t.Fatal(err)
	}

	meta := apo.Metadata()

	if meta.Identifier != core.AbsolutePriceOscillator {
		t.Errorf("expected identifier AbsolutePriceOscillator, got %v", meta.Identifier)
	}

	exp := "apo(SMA12/SMA26)"
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

func TestAbsolutePriceOscillatorMetadataEMA(t *testing.T) {
	t.Parallel()

	apo, err := NewAbsolutePriceOscillator(&AbsolutePriceOscillatorParams{
		FastLength:        12,
		SlowLength:        26,
		MovingAverageType: EMA,
	})
	if err != nil {
		t.Fatal(err)
	}

	meta := apo.Metadata()

	exp := "apo(EMA12/EMA26)"
	if meta.Mnemonic != exp {
		t.Errorf("expected mnemonic '%s', got '%s'", exp, meta.Mnemonic)
	}
}

func TestAbsolutePriceOscillatorUpdateEntity(t *testing.T) {
	t.Parallel()

	const tolerance = 5e-4

	input := testInput()

	apo, err := NewAbsolutePriceOscillator(&AbsolutePriceOscillatorParams{
		FastLength: 2,
		SlowLength: 3,
	})
	if err != nil {
		t.Fatal(err)
	}

	tm := testTime()

	for i := 0; i < 2; i++ {
		scalar := &entities.Scalar{Time: tm, Value: input[i]}
		out := apo.UpdateScalar(scalar)

		v := out[0].(entities.Scalar).Value //nolint:forcetypeassert
		if !math.IsNaN(v) {
			t.Errorf("[%d] expected NaN, got %v", i, v)
		}
	}

	scalar := &entities.Scalar{Time: tm, Value: input[2]}
	out := apo.UpdateScalar(scalar)

	v := out[0].(entities.Scalar).Value //nolint:forcetypeassert
	// SMA(2) of [91.5, 94.815] = 93.1575, SMA(3) of [91.5, 94.815, 94.375] = 93.5633
	// APO = 94.595 - 93.5633 = 1.0317 (approximate)
	if math.IsNaN(v) {
		t.Errorf("[2] expected non-NaN, got NaN")
	}
}

func TestAbsolutePriceOscillatorInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		fast int
		slow int
	}{
		{"fast too small", 1, 26},
		{"slow too small", 12, 1},
		{"fast negative", -8, 12},
		{"slow negative", 26, -7},
	}

	for _, tt := range tests {
		_, err := NewAbsolutePriceOscillator(&AbsolutePriceOscillatorParams{
			FastLength: tt.fast,
			SlowLength: tt.slow,
		})
		if err == nil {
			t.Errorf("%s: expected error, got nil", tt.name)
		}
	}
}
