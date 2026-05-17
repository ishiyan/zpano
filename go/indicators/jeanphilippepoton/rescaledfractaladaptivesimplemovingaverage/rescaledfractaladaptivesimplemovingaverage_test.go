//nolint:testpackage
package rescaledfractaladaptivesimplemovingaverage

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/indicators/core"
)

func testCreate(period, normalSpeed int, priceScale float64) *RescaledFractalAdaptiveSimpleMovingAverage {
	f, _ := NewRescaledFractalAdaptiveSimpleMovingAverage(&Params{Period: period, NormalSpeed: normalSpeed, PriceScale: priceScale})
	return f
}

func TestRescaledFractalAdaptiveSimpleMovingAverageUpdate(t *testing.T) {
	t.Parallel()

	const epsilon = 1e-13

	check := func(t *testing.T, index int, exp, act float64) {
		t.Helper()

		if math.IsNaN(exp) {
			if !math.IsNaN(act) {
				t.Errorf("[%v] expected NaN, got %v", index, act)
			}

			return
		}

		if math.Abs(exp-act) > epsilon {
			t.Errorf("[%v] expected %v, got %v", index, exp, act)
		}
	}

	input := testInput

	t.Run("period=4, normal_speed=30, price_scale=1.0", func(t *testing.T) {
		t.Parallel()

		f := testCreate(4, 30, 1.0)
		expected := expectedP4_S1

		for i, v := range input {
			act := f.Update(v)
			check(t, i, expected[i], act)
		}
	})

	t.Run("period=8, normal_speed=30, price_scale=1.0", func(t *testing.T) {
		t.Parallel()

		f := testCreate(8, 30, 1.0)
		expected := expectedP8_S1

		for i, v := range input {
			act := f.Update(v)
			check(t, i, expected[i], act)
		}
	})

	t.Run("period=16, normal_speed=30, price_scale=1.0", func(t *testing.T) {
		t.Parallel()

		f := testCreate(16, 30, 1.0)
		expected := expectedP16_S1

		for i, v := range input {
			act := f.Update(v)
			check(t, i, expected[i], act)
		}
	})

	t.Run("period=32, normal_speed=30, price_scale=1.0", func(t *testing.T) {
		t.Parallel()

		f := testCreate(32, 30, 1.0)
		expected := expectedP32_S1

		for i, v := range input {
			act := f.Update(v)
			check(t, i, expected[i], act)
		}
	})

	t.Run("period=64, normal_speed=30, price_scale=1.0", func(t *testing.T) {
		t.Parallel()

		f := testCreate(64, 30, 1.0)
		expected := expectedP64_S1

		for i, v := range input {
			act := f.Update(v)
			check(t, i, expected[i], act)
		}
	})

	t.Run("period=128, normal_speed=30, price_scale=1.0", func(t *testing.T) {
		t.Parallel()

		f := testCreate(128, 30, 1.0)
		expected := expectedP128_S1

		for i, v := range input {
			act := f.Update(v)
			check(t, i, expected[i], act)
		}
	})

	t.Run("period=32, normal_speed=30, price_scale=100.0", func(t *testing.T) {
		t.Parallel()

		f := testCreate(32, 30, 100.0)
		expected := expectedP32_S100

		for i, v := range input {
			act := f.Update(v)
			check(t, i, expected[i], act)
		}
	})

	t.Run("period=32, normal_speed=30, price_scale=10000.0", func(t *testing.T) {
		t.Parallel()

		f := testCreate(32, 30, 10000.0)
		expected := expectedP32_S10000

		for i, v := range input {
			act := f.Update(v)
			check(t, i, expected[i], act)
		}
	})
}

func TestRescaledFractalAdaptiveSimpleMovingAverageIsPrimed(t *testing.T) {
	t.Parallel()

	f := testCreate(64, 30, 1.0)
	input := testInput

	for i := 0; i < 64; i++ {
		f.Update(input[i])

		if f.IsPrimed() {
			t.Errorf("should not be primed at index %d", i)
		}
	}

	f.Update(input[64])

	if !f.IsPrimed() {
		t.Error("should be primed after period+1 updates")
	}
}

func TestRescaledFractalAdaptiveSimpleMovingAverageNaN(t *testing.T) {
	t.Parallel()

	f := testCreate(4, 30, 1.0)
	result := f.Update(math.NaN())

	if !math.IsNaN(result) {
		t.Errorf("expected NaN, got %v", result)
	}
}

func TestRescaledFractalAdaptiveSimpleMovingAverageInvalidParams(t *testing.T) {
	t.Parallel()

	_, err := NewRescaledFractalAdaptiveSimpleMovingAverage(&Params{Period: 2, NormalSpeed: 30, PriceScale: 1.0})
	if err == nil {
		t.Error("expected error for period=2")
	}

	_, err = NewRescaledFractalAdaptiveSimpleMovingAverage(&Params{Period: 6, NormalSpeed: 30, PriceScale: 1.0})
	if err == nil {
		t.Error("expected error for period=6 (not power of 2)")
	}

	_, err = NewRescaledFractalAdaptiveSimpleMovingAverage(&Params{Period: 4, NormalSpeed: 0, PriceScale: 1.0})
	if err == nil {
		t.Error("expected error for normal_speed=0")
	}
}

func TestRescaledFractalAdaptiveSimpleMovingAverageMetadata(t *testing.T) {
	t.Parallel()

	f := testCreate(64, 30, 1.0)
	meta := f.Metadata()

	if meta.Identifier != core.RescaledFractalAdaptiveSimpleMovingAverage {
		t.Errorf("expected identifier %v, got %v", core.RescaledFractalAdaptiveSimpleMovingAverage, meta.Identifier)
	}
}
