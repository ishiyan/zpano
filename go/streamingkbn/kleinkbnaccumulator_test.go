package streamingkbn

import (
	"math"
	"math/rand/v2"
	"testing"
)

// https://en.wikipedia.org/wiki/Kahan_summation_algorithm
// A simple example due to Peters: summing [1.0, +1e100, 1.0, -1e100]
// in double precision, Kahan's algorithm yields 0.0, whereas
// Neumaier's algorithm yields the correct value 2.0.
func petersData() []float64 {
	return []float64{1.0, +1e100, 1.0, -1e100}
}

// https://github.com/numpy/numpy/issues/8786
// A badly conditioned sum, condition number ~2.188e+14.
func numpyData() []float64 {
	return []float64{
		-0.41253261766461263,
		41287272281118.43,
		-1.4727977348624173e-14,
		5670.3302557520055,
		2.119245229045646e-11,
		-0.003679264134906428,
		-6.892634568678797e-14,
		-0.0006984744181630712,
		-4054136.048352595,
		-1003.101760720037,
		-1.4436349910427172e-17,
		-41287268231649.57,
	}
}

const numpyExpected = -0.377392919181026

// A sequence where the second-level correction is non-zero at more
// than one step, so it must be accumulated (ccs += cc), not
// overwritten (ccs = cc). Overwriting yields -1.0000000000000001e-16.
func kleinData() []float64 {
	return []float64{1e-16, -1e16, 1.0, 1e-16, -1.0, -1e-16, -1e-32, 1e16}
}

const kleinExpected = 9.999999999999999e-17 // math.fsum(klein_data)

func kbnSum(data []float64) float64 {
	var kbn KleinKBNAccumulator
	for _, x := range data {
		kbn.Update(x)
	}
	return kbn.Value()
}

func TestKleinKBNAccumulator_InitialValueIsZero(t *testing.T) {
	t.Parallel()

	var kbn KleinKBNAccumulator
	assertEqual(t, "value", kbn.Value(), 0.0)
}

func TestKleinKBNAccumulator_Peters(t *testing.T) {
	t.Parallel()

	assertEqual(t, "naive", naiveSum(petersData()), 0.0)
	assertEqual(t, "kbn", kbnSum(petersData()), 2.0)
}

func TestKleinKBNAccumulator_NumpyIssue(t *testing.T) {
	t.Parallel()

	assertAlmostEqual(t, "kbn", kbnSum(numpyData()), numpyExpected, 16)
	if v := naiveSum(numpyData()); almostEqual(v, numpyExpected, places(3)) {
		t.Errorf("naive = %v, should not be almost equal to %v (places=3)", v, numpyExpected)
	}
}

func TestKleinKBNAccumulator_SecondLevelCorrectionIsAccumulated(t *testing.T) {
	t.Parallel()

	assertEqual(t, "kbn", kbnSum(kleinData()), kleinExpected)
}

func TestKleinKBNAccumulator_MatchesFsumOnMixedMagnitudes(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewPCG(42, 0))
	exponents := []int{-8, 0, 8}
	for range 200 {
		data := make([]float64, 100)
		for i := range data {
			u := -1.0 + 2.0*rng.Float64() // uniform(-1, 1)
			data[i] = u * math.Pow10(exponents[rng.IntN(len(exponents))])
		}
		assertEqual(t, "kbn vs fsum", kbnSum(data), fsum(data))
	}
}

func TestKleinKBNAccumulator_BetterAccuracyThanNaive(t *testing.T) {
	t.Parallel()

	// Add and then subtract the same values, so the exact sum is 0.
	rng := rand.New(rand.NewPCG(42, 0))
	const count = 100000
	data := make([]float64, 2*count)
	for i := range count {
		data[i] = 1e7 * rng.Float64() // uniform(0, 1e7)
	}
	for i := range count {
		data[count+i] = -data[i]
	}
	k := kbnSum(data)
	v := naiveSum(data)
	assertEqual(t, "kbn", k, 0.0)
	if v == 0.0 {
		t.Errorf("naive = %v, want non-zero", v)
	}
}

func TestKleinKBNAccumulator_UpdateZeroKeepsCompensation(t *testing.T) {
	t.Parallel()

	var kbn KleinKBNAccumulator
	for _, x := range kleinData() {
		kbn.Update(x)
	}
	kbn.Update(0.0)
	assertEqual(t, "value", kbn.Value(), kleinExpected)
}

func TestKleinKBNAccumulator_Revert(t *testing.T) {
	t.Parallel()

	var kbn KleinKBNAccumulator
	kbn.Update(1.5)
	kbn.Update(2.5)
	kbn.Revert(2.5)
	assertEqual(t, "value", kbn.Value(), 1.5)
	kbn.Revert(1.5)
	assertEqual(t, "value", kbn.Value(), 0.0)
}

func TestKleinKBNAccumulator_RevertNotMostRecent(t *testing.T) {
	t.Parallel()

	var kbn KleinKBNAccumulator
	for _, x := range petersData() {
		kbn.Update(x)
	}
	kbn.Revert(1e100) // not the most recent value
	assertEqual(t, "value", kbn.Value(), 2.0-1e100)
	kbn.Revert(-1e100)
	assertEqual(t, "value", kbn.Value(), 2.0)
}

func TestKleinKBNAccumulator_RevertRestoresCompensatedSum(t *testing.T) {
	t.Parallel()

	var kbn KleinKBNAccumulator
	for _, x := range numpyData() {
		kbn.Update(x)
	}
	kbn.Update(1e20)
	kbn.Revert(1e20)
	assertAlmostEqual(t, "value", kbn.Value(), numpyExpected, 16)
}

func TestKleinKBNAccumulator_Set(t *testing.T) {
	t.Parallel()

	var kbn KleinKBNAccumulator
	for _, x := range petersData() {
		kbn.Update(x)
	}
	kbn.Set(5.0)
	assertEqual(t, "value", kbn.Value(), 5.0)
	// Compensation terms are cleared, so only the new values count.
	kbn.Update(1e100)
	kbn.Update(-1e100)
	assertEqual(t, "value", kbn.Value(), 5.0)
}

func TestKleinKBNAccumulator_Reset(t *testing.T) {
	t.Parallel()

	var kbn KleinKBNAccumulator
	for _, x := range petersData() {
		kbn.Update(x)
	}
	kbn.Reset()
	assertEqual(t, "value", kbn.Value(), 0.0)
	kbn.Update(1.5)
	assertEqual(t, "value", kbn.Value(), 1.5)
}
