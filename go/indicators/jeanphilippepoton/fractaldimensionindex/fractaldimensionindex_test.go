//nolint:testpackage
package fractaldimensionindex

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/indicators/core"
)

func testFractalDimensionIndexCreate(period int) *FractalDimensionIndex {
	fdi, _ := NewFractalDimensionIndex(&Params{Period: period})
	return fdi
}

func TestFractalDimensionIndexUpdate(t *testing.T) {
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

	t.Run("period = 5", func(t *testing.T) {
		t.Parallel()

		fdi := testFractalDimensionIndexCreate(5)
		expected := expected_P5

		for i, v := range input {
			act := fdi.Update(v)
			check(t, i, expected[i], act)
		}
	})

	t.Run("period = 10", func(t *testing.T) {
		t.Parallel()

		fdi := testFractalDimensionIndexCreate(10)
		expected := expected_P10

		for i, v := range input {
			act := fdi.Update(v)
			check(t, i, expected[i], act)
		}
	})

	t.Run("period = 15", func(t *testing.T) {
		t.Parallel()

		fdi := testFractalDimensionIndexCreate(15)
		expected := expected_P15

		for i, v := range input {
			act := fdi.Update(v)
			check(t, i, expected[i], act)
		}
	})

	t.Run("period = 20", func(t *testing.T) {
		t.Parallel()

		fdi := testFractalDimensionIndexCreate(20)
		expected := expected_P20

		for i, v := range input {
			act := fdi.Update(v)
			check(t, i, expected[i], act)
		}
	})

	t.Run("period = 30", func(t *testing.T) {
		t.Parallel()

		fdi := testFractalDimensionIndexCreate(30)
		expected := expected_P30

		for i, v := range input {
			act := fdi.Update(v)
			check(t, i, expected[i], act)
		}
	})

	t.Run("period = 50", func(t *testing.T) {
		t.Parallel()

		fdi := testFractalDimensionIndexCreate(50)
		expected := expected_P50

		for i, v := range input {
			act := fdi.Update(v)
			check(t, i, expected[i], act)
		}
	})

	t.Run("period = 80", func(t *testing.T) {
		t.Parallel()

		fdi := testFractalDimensionIndexCreate(80)
		expected := expected_P80

		for i, v := range input {
			act := fdi.Update(v)
			check(t, i, expected[i], act)
		}
	})

	t.Run("period = 120", func(t *testing.T) {
		t.Parallel()

		fdi := testFractalDimensionIndexCreate(120)
		expected := expected_P120

		for i, v := range input {
			act := fdi.Update(v)
			check(t, i, expected[i], act)
		}
	})
}

func TestFractalDimensionIndexIsPrimed(t *testing.T) {
	t.Parallel()

	fdi := testFractalDimensionIndexCreate(30)
	input := testInput

	for i := 0; i < 30; i++ {
		fdi.Update(input[i])

		if fdi.IsPrimed() {
			t.Errorf("should not be primed at index %d", i)
		}
	}

	fdi.Update(input[30])

	if !fdi.IsPrimed() {
		t.Error("should be primed after period+1 updates")
	}
}

func TestFractalDimensionIndexNaN(t *testing.T) {
	t.Parallel()

	fdi := testFractalDimensionIndexCreate(5)
	result := fdi.Update(math.NaN())

	if !math.IsNaN(result) {
		t.Errorf("expected NaN, got %v", result)
	}
}

func TestFractalDimensionIndexInvalidParams(t *testing.T) {
	t.Parallel()

	_, err := NewFractalDimensionIndex(&Params{Period: 1})
	if err == nil {
		t.Error("expected error for period=1")
	}
}

func TestFractalDimensionIndexMetadata(t *testing.T) {
	t.Parallel()

	fdi := testFractalDimensionIndexCreate(30)
	meta := fdi.Metadata()

	if meta.Identifier != core.FractalDimensionIndex {
		t.Errorf("expected identifier %v, got %v", core.FractalDimensionIndex, meta.Identifier)
	}
}
