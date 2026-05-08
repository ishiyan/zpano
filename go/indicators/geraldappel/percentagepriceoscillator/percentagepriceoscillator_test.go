//nolint:testpackage
package percentagepriceoscillator

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func TestPercentagePriceOscillatorSMA32(t *testing.T) {
	t.Parallel()

	const tolerance = 5e-4 // C# test rounds to 4 decimal places.

	input := testInput()

	ppo, err := NewPercentagePriceOscillator(&PercentagePriceOscillatorParams{
		FastLength: 2,
		SlowLength: 3,
	})
	if err != nil {
		t.Fatal(err)
	}

	// First 2 values should be NaN (slow SMA(3) not primed until index 2).
	for i := 0; i < 2; i++ {
		v := ppo.Update(input[i])
		if !math.IsNaN(v) {
			t.Errorf("[%d] expected NaN, got %v", i, v)
		}
	}

	// Index 2: first value.
	v := ppo.Update(input[2])
	if math.IsNaN(v) {
		t.Errorf("[2] expected non-NaN, got NaN")
	}

	if math.Abs(v-1.10264) > tolerance {
		t.Errorf("[2] expected ~1.10264, got %v", v)
	}

	// Index 3: second value.
	v = ppo.Update(input[3])
	if math.Abs(v-(-0.02813)) > tolerance {
		t.Errorf("[3] expected ~-0.02813, got %v", v)
	}

	// Feed remaining and check last.
	for i := 4; i < 251; i++ {
		ppo.Update(input[i])
	}

	v = ppo.Update(input[251])
	if math.Abs(v-(-0.21191)) > tolerance {
		t.Errorf("[251] expected ~-0.21191, got %v", v)
	}

	if !ppo.IsPrimed() {
		t.Error("expected primed")
	}
}

func TestPercentagePriceOscillatorSMA2612(t *testing.T) {
	t.Parallel()

	const tolerance = 5e-4

	input := testInput()

	ppo, err := NewPercentagePriceOscillator(&PercentagePriceOscillatorParams{
		FastLength: 12,
		SlowLength: 26,
	})
	if err != nil {
		t.Fatal(err)
	}

	// First 25 values should be NaN.
	for i := 0; i < 25; i++ {
		v := ppo.Update(input[i])
		if !math.IsNaN(v) {
			t.Errorf("[%d] expected NaN, got %v", i, v)
		}
	}

	// Index 25: first value.
	v := ppo.Update(input[25])
	if math.IsNaN(v) {
		t.Errorf("[25] expected non-NaN, got NaN")
	}

	if math.Abs(v-(-3.6393)) > tolerance {
		t.Errorf("[25] expected ~-3.6393, got %v", v)
	}

	// Index 26: second value.
	v = ppo.Update(input[26])
	if math.Abs(v-(-3.9534)) > tolerance {
		t.Errorf("[26] expected ~-3.9534, got %v", v)
	}

	// Feed remaining and check last.
	for i := 27; i < 251; i++ {
		ppo.Update(input[i])
	}

	v = ppo.Update(input[251])
	if math.Abs(v-(-0.15281)) > tolerance {
		t.Errorf("[251] expected ~-0.15281, got %v", v)
	}
}

func TestPercentagePriceOscillatorEMA2612(t *testing.T) {
	t.Parallel()

	const tolerance = 5e-3 // C# test rounds to 3 decimal places for EMA.

	input := testInput()

	ppo, err := NewPercentagePriceOscillator(&PercentagePriceOscillatorParams{
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
		v := ppo.Update(input[i])
		if !math.IsNaN(v) {
			t.Errorf("[%d] expected NaN, got %v", i, v)
		}
	}

	// Index 25: first value.
	v := ppo.Update(input[25])
	if math.IsNaN(v) {
		t.Errorf("[25] expected non-NaN, got NaN")
	}

	if math.Abs(v-(-2.7083)) > tolerance {
		t.Errorf("[25] expected ~-2.7083, got %v", v)
	}

	// Index 26: second value.
	v = ppo.Update(input[26])
	if math.Abs(v-(-2.7390)) > tolerance {
		t.Errorf("[26] expected ~-2.7390, got %v", v)
	}

	// Feed remaining and check last.
	for i := 27; i < 251; i++ {
		ppo.Update(input[i])
	}

	v = ppo.Update(input[251])
	if math.Abs(v-0.83644) > tolerance {
		t.Errorf("[251] expected ~0.83644, got %v", v)
	}
}

func TestPercentagePriceOscillatorIsPrimed(t *testing.T) {
	t.Parallel()

	ppo, err := NewPercentagePriceOscillator(&PercentagePriceOscillatorParams{
		FastLength: 3,
		SlowLength: 5,
	})
	if err != nil {
		t.Fatal(err)
	}

	if ppo.IsPrimed() {
		t.Error("expected not primed initially")
	}

	for i := 1; i < 5; i++ {
		ppo.Update(float64(i))
		if ppo.IsPrimed() {
			t.Errorf("[%d] expected not primed", i)
		}
	}

	ppo.Update(5)

	if !ppo.IsPrimed() {
		t.Error("expected primed after 5 samples")
	}

	for i := 6; i < 10; i++ {
		ppo.Update(float64(i))
		if !ppo.IsPrimed() {
			t.Errorf("[%d] expected primed", i)
		}
	}
}

func TestPercentagePriceOscillatorNaN(t *testing.T) {
	t.Parallel()

	ppo, err := NewPercentagePriceOscillator(&PercentagePriceOscillatorParams{
		FastLength: 2,
		SlowLength: 3,
	})
	if err != nil {
		t.Fatal(err)
	}

	v := ppo.Update(math.NaN())
	if !math.IsNaN(v) {
		t.Errorf("expected NaN, got %v", v)
	}
}

func TestPercentagePriceOscillatorMetadata(t *testing.T) {
	t.Parallel()

	ppo, err := NewPercentagePriceOscillator(&PercentagePriceOscillatorParams{
		FastLength: 12,
		SlowLength: 26,
	})
	if err != nil {
		t.Fatal(err)
	}

	meta := ppo.Metadata()

	if meta.Identifier != core.PercentagePriceOscillator {
		t.Errorf("expected identifier PercentagePriceOscillator, got %v", meta.Identifier)
	}

	exp := "ppo(SMA12/SMA26)"
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

func TestPercentagePriceOscillatorMetadataEMA(t *testing.T) {
	t.Parallel()

	ppo, err := NewPercentagePriceOscillator(&PercentagePriceOscillatorParams{
		FastLength:        12,
		SlowLength:        26,
		MovingAverageType: EMA,
	})
	if err != nil {
		t.Fatal(err)
	}

	meta := ppo.Metadata()

	exp := "ppo(EMA12/EMA26)"
	if meta.Mnemonic != exp {
		t.Errorf("expected mnemonic '%s', got '%s'", exp, meta.Mnemonic)
	}
}

func TestPercentagePriceOscillatorUpdateEntity(t *testing.T) {
	t.Parallel()

	const tolerance = 5e-4

	input := testInput()

	ppo, err := NewPercentagePriceOscillator(&PercentagePriceOscillatorParams{
		FastLength: 2,
		SlowLength: 3,
	})
	if err != nil {
		t.Fatal(err)
	}

	tm := testTime()

	for i := 0; i < 2; i++ {
		scalar := &entities.Scalar{Time: tm, Value: input[i]}
		out := ppo.UpdateScalar(scalar)

		v := out[0].(entities.Scalar).Value //nolint:forcetypeassert
		if !math.IsNaN(v) {
			t.Errorf("[%d] expected NaN, got %v", i, v)
		}
	}

	scalar := &entities.Scalar{Time: tm, Value: input[2]}
	out := ppo.UpdateScalar(scalar)

	v := out[0].(entities.Scalar).Value //nolint:forcetypeassert
	if math.Abs(v-1.10264) > tolerance {
		t.Errorf("[2] expected ~1.10264, got %v", v)
	}
}

func TestPercentagePriceOscillatorInvalidParams(t *testing.T) {
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
		_, err := NewPercentagePriceOscillator(&PercentagePriceOscillatorParams{
			FastLength: tt.fast,
			SlowLength: tt.slow,
		})
		if err == nil {
			t.Errorf("%s: expected error, got nil", tt.name)
		}
	}
}
