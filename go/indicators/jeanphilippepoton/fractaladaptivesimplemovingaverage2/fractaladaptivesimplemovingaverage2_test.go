//nolint:testpackage
package fractaladaptivesimplemovingaverage2

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/indicators/core"
)

func testCreate(period, normalSpeed int) *FractalAdaptiveSimpleMovingAverage2 {
	f, _ := NewFractalAdaptiveSimpleMovingAverage2(&Params{Period: period, NormalSpeed: normalSpeed})
	return f
}

func TestFractalAdaptiveSimpleMovingAverage2Update(t *testing.T) {
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

	t.Run("period=5, normal_speed=20", func(t *testing.T) {
		t.Parallel()

		f := testCreate(5, 20)
		expected := expectedP5

		for i, v := range input {
			act := f.Update(v)
			check(t, i, expected[i], act)
		}
	})

	t.Run("period=10, normal_speed=20", func(t *testing.T) {
		t.Parallel()

		f := testCreate(10, 20)
		expected := expectedP10

		for i, v := range input {
			act := f.Update(v)
			check(t, i, expected[i], act)
		}
	})

	t.Run("period=15, normal_speed=20", func(t *testing.T) {
		t.Parallel()

		f := testCreate(15, 20)
		expected := expectedP15

		for i, v := range input {
			act := f.Update(v)
			check(t, i, expected[i], act)
		}
	})

	t.Run("period=20, normal_speed=20", func(t *testing.T) {
		t.Parallel()

		f := testCreate(20, 20)
		expected := expectedP20

		for i, v := range input {
			act := f.Update(v)
			check(t, i, expected[i], act)
		}
	})

	t.Run("period=30, normal_speed=20", func(t *testing.T) {
		t.Parallel()

		f := testCreate(30, 20)
		expected := expectedP30

		for i, v := range input {
			act := f.Update(v)
			check(t, i, expected[i], act)
		}
	})

	t.Run("period=50, normal_speed=20", func(t *testing.T) {
		t.Parallel()

		f := testCreate(50, 20)
		expected := expectedP50

		for i, v := range input {
			act := f.Update(v)
			check(t, i, expected[i], act)
		}
	})

	t.Run("period=80, normal_speed=20", func(t *testing.T) {
		t.Parallel()

		f := testCreate(80, 20)
		expected := expectedP80

		for i, v := range input {
			act := f.Update(v)
			check(t, i, expected[i], act)
		}
	})

	t.Run("period=120, normal_speed=20", func(t *testing.T) {
		t.Parallel()

		f := testCreate(120, 20)
		expected := expectedP120

		for i, v := range input {
			act := f.Update(v)
			check(t, i, expected[i], act)
		}
	})
}

func TestFractalAdaptiveSimpleMovingAverage2IsPrimed(t *testing.T) {
	t.Parallel()

	f := testCreate(30, 20)
	input := testInput

	for i := 0; i < 29; i++ {
		f.Update(input[i])

		if f.IsPrimed() {
			t.Errorf("should not be primed at index %d", i)
		}
	}

	f.Update(input[29])

	if !f.IsPrimed() {
		t.Error("should be primed after period updates")
	}
}

func TestFractalAdaptiveSimpleMovingAverage2NaN(t *testing.T) {
	t.Parallel()

	f := testCreate(5, 20)
	result := f.Update(math.NaN())

	if !math.IsNaN(result) {
		t.Errorf("expected NaN, got %v", result)
	}
}

func TestFractalAdaptiveSimpleMovingAverage2InvalidParams(t *testing.T) {
	t.Parallel()

	_, err := NewFractalAdaptiveSimpleMovingAverage2(&Params{Period: 1, NormalSpeed: 20})
	if err == nil {
		t.Error("expected error for period=1")
	}

	_, err = NewFractalAdaptiveSimpleMovingAverage2(&Params{Period: 5, NormalSpeed: 0})
	if err == nil {
		t.Error("expected error for normal_speed=0")
	}
}

func TestFractalAdaptiveSimpleMovingAverage2Metadata(t *testing.T) {
	t.Parallel()

	f := testCreate(30, 20)
	meta := f.Metadata()

	if meta.Identifier != core.FractalAdaptiveSimpleMovingAverage2 {
		t.Errorf("expected identifier %v, got %v", core.FractalAdaptiveSimpleMovingAverage2, meta.Identifier)
	}
}
