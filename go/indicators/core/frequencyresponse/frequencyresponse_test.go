//nolint:testpackage
package frequencyresponse

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/indicators/core"
)

type testFrequencyResponseIdentytyFilter int

func (s testFrequencyResponseIdentytyFilter) Metadata() core.Metadata {
	return core.Metadata{Mnemonic: "identity"}
}

func (s testFrequencyResponseIdentytyFilter) Update(sample float64) float64 {
	return sample
}

func TestFrequencyResponse(t *testing.T) {
	t.Parallel()

	check := func(index int, exp, act float64) {
		t.Helper()

		if math.Abs(exp-act) > math.SmallestNonzeroFloat64 {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	t.Run("validate signal length", func(t *testing.T) {
		t.Parallel()

		const maxLength = 8199

		isValid := func(length int) bool {
			switch length {
			case 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192:
				return true
			default:
				return false
			}
		}

		for i := -1; i < maxLength; i++ {
			exp := isValid(i)
			act := isValidSignalLength(i)

			if exp != act {
				t.Errorf("isValidSignalLength(%v): expected '%v', actual '%v'", i, exp, act)
			}
		}
	})

	t.Run("prepare frequency domain", func(t *testing.T) {
		t.Parallel()

		const l = float64(7)

		expected := []float64{1 / l, 2 / l, 3 / l, 4 / l, 5 / l, 6 / l, 7 / l}
		actual := []float64{0, 0, 0, 0, 0, 0, 0}

		prepareFrequencyDomain(int(l), actual)

		for i := 0; i < len(expected); i++ {
			exp := expected[i]
			act := actual[i]
			check(i, exp, act)
		}
	})

	t.Run("prepare filtered signal", func(t *testing.T) {
		t.Parallel()

		const length = 7
		const warmup = 5

		expected := []float64{1000, 0, 0, 0, 0, 0, 0}
		filter := testFrequencyResponseIdentytyFilter(0)

		actual := prepareFilteredSignal(length, filter, warmup)

		for i := 0; i < length; i++ {
			exp := expected[i]
			act := actual[i]
			check(i, exp, act)
		}
	})

	t.Run("calculate FFT", func(t *testing.T) {
		t.Parallel()

		expected := []float64{16, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
		actual := []float64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}

		directRealFastFourierTransform(actual)

		for i := 0; i < len(expected); i++ {
			exp := expected[i]
			act := actual[i]
			check(i, exp, act)
		}
	})

	t.Run("normalize array", func(t *testing.T) {
		t.Parallel()

		t.Run("zero max", func(t *testing.T) {
			t.Parallel()

			expected := []float64{1, 2, 3, 4, 5}
			actual := []float64{1, 2, 3, 4, 5}

			normalize(len(actual), actual, 0)

			for i := 0; i < len(expected); i++ {
				exp := expected[i]
				act := actual[i]
				check(i, exp, act)
			}
		})

		t.Run("positive max", func(t *testing.T) {
			t.Parallel()

			const max = float64(6)

			expected := []float64{2 / max, 3 / max, 4 / max, 5 / max, 6 / max}
			actual := []float64{2, 3, 4, 5, 6}

			normalize(len(actual), actual, max)

			for i := 0; i < len(expected); i++ {
				exp := expected[i]
				act := actual[i]
				check(i, exp, act)
			}
		})
	})

	t.Run("calculate", func(t *testing.T) {
		t.Parallel()

		const length = 512
		const warmup = 128

		filter := testFrequencyResponseIdentytyFilter(0)

		actual, _ := Calculate(length, filter, warmup)

		t.Log(actual != nil)
	})

	t.Run("calculates FFT", func(t *testing.T) {
		t.Parallel()
	})

	t.Run("calculates FFT", func(t *testing.T) {
		t.Parallel()
	})
}
