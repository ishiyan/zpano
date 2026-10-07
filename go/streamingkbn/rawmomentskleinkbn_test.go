package streamingkbn

import (
	"fmt"
	"math"
	"testing"
)

// https://github.com/medo64/Medo/blob/main/tests/Tests.Medo/Math/WelfordVariance.cs
// https://github.com/andrewuhl/RollingWindow/blob/master/src/RollingWindow.cpp
// https://github.com/ajcr/rolling/blob/master/rolling/similarity.py

// rawProps maps the Python property names to the Go getter methods.
var rawProps = map[string]func(*RawMomentsKleinKBN) float64{
	"mean":                      (*RawMomentsKleinKBN).Mean,
	"variance":                  (*RawMomentsKleinKBN).Variance,
	"standard_deviation":        (*RawMomentsKleinKBN).StandardDeviation,
	"variance_ddof_0":           (*RawMomentsKleinKBN).VarianceDdof0,
	"variance_ddof_1":           (*RawMomentsKleinKBN).VarianceDdof1,
	"standard_deviation_ddof_0": (*RawMomentsKleinKBN).StandardDeviationDdof0,
	"standard_deviation_ddof_1": (*RawMomentsKleinKBN).StandardDeviationDdof1,
	"skewness":                  (*RawMomentsKleinKBN).Skewness,
	"skewness_moment":           (*RawMomentsKleinKBN).SkewnessMoment,
	"skewness_fisher":           (*RawMomentsKleinKBN).SkewnessFisher,
	"skewness_sample":           (*RawMomentsKleinKBN).SkewnessSample,
	"kurtosis":                  (*RawMomentsKleinKBN).Kurtosis,
	"kurtosis_moment":           (*RawMomentsKleinKBN).KurtosisMoment,
	"kurtosis_excess":           (*RawMomentsKleinKBN).KurtosisExcess,
	"kurtosis_sample_excess":    (*RawMomentsKleinKBN).KurtosisSampleExcess,
	"kurtosis_sample":           (*RawMomentsKleinKBN).KurtosisSample,
	"kurtosis_sample_corrected": (*RawMomentsKleinKBN).KurtosisSampleCorrected,
	"x1_sum":                    (*RawMomentsKleinKBN).X1Sum,
	"x2_sum":                    (*RawMomentsKleinKBN).X2Sum,
	"x3_sum":                    (*RawMomentsKleinKBN).X3Sum,
	"x4_sum":                    (*RawMomentsKleinKBN).X4Sum,
	"x1":                        (*RawMomentsKleinKBN).X1,
	"x2":                        (*RawMomentsKleinKBN).X2,
	"x3":                        (*RawMomentsKleinKBN).X3,
	"x4":                        (*RawMomentsKleinKBN).X4,
}

func rawGet(m *RawMomentsKleinKBN, name string) float64 {
	f, ok := rawProps[name]
	if !ok {
		panic("unknown property " + name)
	}
	return f(m)
}

type namedValue struct {
	name string
	want float64
}

// Reference values for bacon, computed with exact rational arithmetic
// (fractions.Fraction on the binary float inputs, square roots with
// 50-digit decimal.Decimal), rounded to the nearest float.
var rawExpected = []namedValue{
	{"mean", 0.009000000000000001},
	{"variance_ddof_0", 0.0014989166666666668},
	{"variance_ddof_1", 0.0015640869565217393},
	{"standard_deviation_ddof_0", 0.03871584516275819},
	{"standard_deviation_ddof_1", 0.039548539246370897},
	{"skewness_moment", -0.08256245520856804},         // scipy skew(bias=True)
	{"skewness_fisher", -0.08817174934967535},         // scipy skew(bias=False)
	{"skewness_sample", -0.09398413873544505},         // R PerformanceAnalytics "sample"
	{"kurtosis_moment", 2.4324537941078743},           // scipy kurtosis(bias=True, fisher=False)
	{"kurtosis_excess", -0.5675462058921257},          // scipy kurtosis(bias=True, fisher=True)
	{"kurtosis_sample_excess", -0.40766032118608714},  // scipy kurtosis(bias=False, fisher=True)
	{"kurtosis_sample", 2.592339678813913},            // scipy kurtosis(bias=False, fisher=False)
	{"kurtosis_sample_corrected", 3.027404613878848},  // R PerformanceAnalytics "sample"
	{"x1_sum", 0.21600000000000003},
	{"x2_sum", 0.037918},
	{"x3_sum", 0.0008738040000000003},
	{"x4_sum", 0.00014466403},
	{"x1", 0.009000000000000001},
	{"x2", 0.0015799166666666668},
	{"x3", 3.640850000000001e-05},
	{"x4", 6.0276679166666674e-06},
}

func rawExpectedValue(name string) float64 {
	for _, nv := range rawExpected {
		if nv.name == name {
			return nv.want
		}
	}
	panic("unknown expected value " + name)
}

// Defaults are ddof=1, bias=true, fisher=true.
var rawExpectedByDispatch = []namedValue{
	{"mean", rawExpectedValue("mean")},
	{"variance", rawExpectedValue("variance_ddof_1")},
	{"skewness", rawExpectedValue("skewness_moment")},
	{"kurtosis", rawExpectedValue("kurtosis_excess")},
}

func feedRaw(m *RawMomentsKleinKBN, data []float64) *RawMomentsKleinKBN {
	for _, x := range data {
		m.Update(x)
	}
	return m
}

// newRawDefault mirrors Python's RawMomentsKleinKBN() defaults.
func newRawDefault() *RawMomentsKleinKBN {
	return NewRawMomentsKleinKBN(1, true, true)
}

func TestRawMomentsKleinKBN_SimpleUpdate(t *testing.T) {
	t.Parallel()

	m := feedRaw(NewRawMomentsKleinKBN(0, true, true), []float64{1.0, 2.0, 3.0, 4.0})
	assertEqualInt(t, "n", m.N(), 4)
	assertAlmostEqual(t, "mean", m.Mean(), 2.5, 15)
	assertAlmostEqual(t, "variance", m.Variance(), 1.25, 15)
	assertAlmostEqual(t, "skewness", m.Skewness(), 0.0, 14)
	assertAlmostEqual(t, "kurtosis", m.Kurtosis(), -1.36, 13)
}

func TestRawMomentsKleinKBN_BaconAllProperties(t *testing.T) {
	t.Parallel()

	m := feedRaw(newRawDefault(), bacon)
	for _, nv := range rawExpected {
		assertAlmostEqual(t, nv.name, rawGet(m, nv.name), nv.want, 14)
	}
}

func TestRawMomentsKleinKBN_Dispatch(t *testing.T) {
	t.Parallel()

	cases := []struct {
		bias, fisher bool
		skew, kurt   string
	}{
		{true, true, "skewness_moment", "kurtosis_excess"},
		{true, false, "skewness_moment", "kurtosis_moment"},
		{false, true, "skewness_fisher", "kurtosis_sample_excess"},
		{false, false, "skewness_fisher", "kurtosis_sample"},
	}
	for _, c := range cases {
		m := feedRaw(NewRawMomentsKleinKBN(1, c.bias, c.fisher), bacon)
		label := fmt.Sprintf("bias=%v fisher=%v", c.bias, c.fisher)
		assertEqual(t, label+" skewness", m.Skewness(), rawGet(m, c.skew))
		assertEqual(t, label+" kurtosis", m.Kurtosis(), rawGet(m, c.kurt))
		assertAlmostEqual(t, label+" skewness", m.Skewness(), rawExpectedValue(c.skew), 14)
		assertAlmostEqual(t, label+" kurtosis", m.Kurtosis(), rawExpectedValue(c.kurt), 13)
	}
}

func TestRawMomentsKleinKBN_Ddof(t *testing.T) {
	t.Parallel()

	for _, ddof := range []int{0, 1} {
		m := feedRaw(NewRawMomentsKleinKBN(ddof, true, true), bacon)
		assertEqual(t, fmt.Sprintf("ddof=%d variance", ddof),
			m.Variance(), rawGet(m, fmt.Sprintf("variance_ddof_%d", ddof)))
		assertEqual(t, fmt.Sprintf("ddof=%d standard_deviation", ddof),
			m.StandardDeviation(), rawGet(m, fmt.Sprintf("standard_deviation_ddof_%d", ddof)))
	}
	m := feedRaw(NewRawMomentsKleinKBN(1, true, true), []float64{1.0, 2.0, 3.0})
	assertAlmostEqual(t, "variance", m.Variance(), 1.0, 15)
	assertAlmostEqual(t, "standard_deviation", m.StandardDeviation(), 1.0, 15)
}

func TestRawMomentsKleinKBN_InvalidDdof(t *testing.T) {
	t.Parallel()

	// Non-integer and bool ddof values are rejected by the Go type system.
	assertPanics(t, "ddof must be a nonnegative integer", func() {
		NewRawMomentsKleinKBN(-1, true, true)
	})
	assertPanics(t, "ddof must be a nonnegative integer", func() {
		newRawDefault().SetDdof(-1)
	})
}

func TestRawMomentsKleinKBN_KurtosisSampleCorrectedDifference(t *testing.T) {
	t.Parallel()

	// kurtosis_sample_corrected - kurtosis_sample = (9n-15) / ((n-2)(n-3))
	m := feedRaw(newRawDefault(), bacon)
	n := len(bacon)
	assertAlmostEqual(t, "difference", m.KurtosisSampleCorrected()-m.KurtosisSample(),
		float64(9*n-15)/float64((n-2)*(n-3)), 14)
}

func TestRawMomentsKleinKBN_Empty(t *testing.T) {
	t.Parallel()

	m := newRawDefault()
	assertEqualInt(t, "n", m.N(), 0)
	assertEqual(t, "mean", m.Mean(), 0.0)
	for _, name := range []string{"variance", "standard_deviation", "skewness", "kurtosis",
		"x1", "x2", "x3", "x4"} {
		assertNaN(t, name, rawGet(m, name))
	}
	assertEqual(t, "x1_sum", m.X1Sum(), 0.0)
}

func TestRawMomentsKleinKBN_MinimumSampleSizes(t *testing.T) {
	t.Parallel()

	data := []float64{1.0, 2.0, 4.0, 8.0}
	minimumN := []struct {
		name string
		minN int
	}{
		{"skewness_moment", 2},
		{"skewness_fisher", 3},
		{"skewness_sample", 3},
		{"kurtosis_moment", 2},
		{"kurtosis_excess", 2},
		{"kurtosis_sample_excess", 4},
		{"kurtosis_sample", 4},
		{"kurtosis_sample_corrected", 4},
	}
	m := newRawDefault()
	for i, x := range data {
		m.Update(x)
		n := i + 1
		for _, c := range minimumN {
			got := rawGet(m, c.name)
			if math.IsNaN(got) != (n < c.minN) {
				t.Errorf("n=%d %s = %v, want NaN=%v", n, c.name, got, n < c.minN)
			}
		}
	}
}

func TestRawMomentsKleinKBN_ConstantData(t *testing.T) {
	t.Parallel()

	m := feedRaw(NewRawMomentsKleinKBN(0, true, true), []float64{0.1, 0.1, 0.1, 0.1, 0.1})
	assertAlmostEqual(t, "mean", m.Mean(), 0.1, 16)
	assertAlmostEqual(t, "variance", m.Variance(), 0.0, 16)
	assertNaN(t, "skewness", m.Skewness())
	assertNaN(t, "kurtosis", m.Kurtosis())
}

func TestRawMomentsKleinKBN_ScaleInvariance(t *testing.T) {
	t.Parallel()

	// The cancellation threshold is relative, so tiny values work.
	scaled := make([]float64, len(bacon))
	for i, x := range bacon {
		scaled[i] = x * 1e-6
	}
	m := feedRaw(newRawDefault(), scaled)
	assertAlmostEqual(t, "skewness_moment", m.SkewnessMoment(), rawExpectedValue("skewness_moment"), 13)
	assertAlmostEqual(t, "kurtosis_excess", m.KurtosisExcess(), rawExpectedValue("kurtosis_excess"), 13)
}

func TestRawMomentsKleinKBN_LargeOffsetPreservesVarianceButNotHigherMoments(t *testing.T) {
	t.Parallel()

	// Welford's variance remains usable when raw-power cancellation
	// makes skewness and kurtosis unreliable.
	m := feedRaw(NewRawMomentsKleinKBN(0, true, true), []float64{1e8, 1e8 + 1, 1e8 + 2})
	assertAlmostEqual(t, "mean", m.Mean(), 1e8+1, 10)
	assertAlmostEqual(t, "variance", m.Variance(), 2.0/3.0, 14)
	assertNaN(t, "skewness", m.Skewness())
	assertNaN(t, "kurtosis", m.Kurtosis())
}

func TestRawMomentsKleinKBN_RevertPartial(t *testing.T) {
	t.Parallel()

	data := []float64{10.0, 18.0, 5.0, 12.0, 7.0}
	mFull := feedRaw(NewRawMomentsKleinKBN(0, true, true), data)
	mPart := feedRaw(NewRawMomentsKleinKBN(0, true, true), data[:4])
	mFull.Revert(data[4])
	assertEqualInt(t, "n", mFull.N(), 4)
	assertAlmostEqual(t, "mean", mFull.Mean(), mPart.Mean(), 15)
	assertAlmostEqual(t, "variance", mFull.Variance(), mPart.Variance(), 15)
	assertAlmostEqual(t, "skewness", mFull.Skewness(), mPart.Skewness(), 14)
	assertAlmostEqual(t, "kurtosis", mFull.Kurtosis(), mPart.Kurtosis(), 13)
}

func TestRawMomentsKleinKBN_RevertNotMostRecent(t *testing.T) {
	t.Parallel()

	m := feedRaw(newRawDefault(), append(append([]float64{}, bacon...), 0.5))
	m.Revert(0.5)
	m2 := feedRaw(newRawDefault(), append([]float64{0.5}, bacon...))
	m2.Revert(0.5) // the oldest sample
	for _, nv := range rawExpectedByDispatch {
		assertAlmostEqual(t, "m "+nv.name, rawGet(m, nv.name), nv.want, 13)
		assertAlmostEqual(t, "m2 "+nv.name, rawGet(m2, nv.name), nv.want, 13)
	}
}

func TestRawMomentsKleinKBN_RevertToEmpty(t *testing.T) {
	t.Parallel()

	m := feedRaw(NewRawMomentsKleinKBN(0, true, true), bacon)
	for _, x := range bacon {
		m.Revert(x)
	}
	assertEqualInt(t, "n", m.N(), 0)
	assertEqual(t, "mean", m.Mean(), 0.0)
	assertEqual(t, "x1_sum", m.X1Sum(), 0.0)
	assertNaN(t, "variance", m.Variance())
	feedRaw(m, []float64{1.0, 2.0, 3.0, 4.0})
	assertAlmostEqual(t, "variance", m.Variance(), 1.25, 15)
}

func TestRawMomentsKleinKBN_RevertEmptyRaises(t *testing.T) {
	t.Parallel()

	m := newRawDefault()
	assertPanics(t, "Cannot revert from an empty accumulator", func() { m.Revert(1.0) })
}

func TestRawMomentsKleinKBN_RollingWindow(t *testing.T) {
	t.Parallel()

	names := []string{"mean", "variance", "standard_deviation", "skewness", "kurtosis",
		"skewness_sample", "kurtosis_sample_corrected", "x1", "x2", "x3", "x4"}
	const w = 5
	m := NewRawMomentsKleinKBN(1, false, true)
	for i, x := range bacon {
		m.Update(x)
		if i >= w {
			m.Revert(bacon[i-w])
		}
		ref := feedRaw(NewRawMomentsKleinKBN(1, false, true), window(bacon, i, w))
		assertEqualInt(t, fmt.Sprintf("step=%d n", i), m.N(), ref.N())
		for _, name := range names {
			assertNaNOrAlmostEqual(t, fmt.Sprintf("step=%d %s", i, name),
				rawGet(m, name), rawGet(ref, name), 13)
		}
	}
}

func TestRawMomentsKleinKBN_StandardDeviationIsRealAfterRevert(t *testing.T) {
	t.Parallel()

	m := NewRawMomentsKleinKBN(0, true, true)
	for _, x := range []float64{0.1, 0.1, 0.7} {
		m.Update(x)
	}
	m.Revert(0.7)
	if math.IsNaN(m.StandardDeviation()) {
		t.Errorf("standard_deviation = NaN, want a real number")
	}
	if !(m.Variance() >= 0.0) {
		t.Errorf("variance = %v, want >= 0", m.Variance())
	}
	assertAlmostEqual(t, "standard_deviation", m.StandardDeviation(), 0.0, 15)
}

func TestRawMomentsKleinKBN_VarianceGetterHasNoSideEffects(t *testing.T) {
	t.Parallel()

	m := feedRaw(NewRawMomentsKleinKBN(0, true, true), bacon)
	v := m.Variance()
	_ = m.StandardDeviation()
	assertEqual(t, "variance", m.Variance(), v)
	assertEqualInt(t, "n", m.N(), len(bacon))
}

func TestRawMomentsKleinKBN_Reset(t *testing.T) {
	t.Parallel()

	m := feedRaw(newRawDefault(), bacon)
	m.Reset()
	assertEqualInt(t, "n", m.N(), 0)
	assertEqual(t, "mean", m.Mean(), 0.0)
	assertEqual(t, "x4_sum", m.X4Sum(), 0.0)
	assertNaN(t, "variance", m.Variance())
	feedRaw(m, []float64{1.0, 2.0, 3.0})
	assertAlmostEqual(t, "variance", m.Variance(), 1.0, 15)
}
