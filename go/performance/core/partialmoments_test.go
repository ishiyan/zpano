package core

import (
	"fmt"
	"math"
	"testing"
)

type namedGetter[T any] struct {
	name string
	get  func(T) float64
}

func intGetter[T any](f func(T) int) func(T) float64 {
	return func(x T) float64 { return float64(f(x)) }
}

var partialMomentsGetters = []namedGetter[*PartialMoments]{
	{"total_count", intGetter((*PartialMoments).TotalCount)},
	{"lower_excess_count", intGetter((*PartialMoments).LowerExcessCount)},
	{"upper_excess_count", intGetter((*PartialMoments).UpperExcessCount)},
	{"downside_frequency", (*PartialMoments).DownsideFrequency},
	{"upside_frequency", (*PartialMoments).UpsideFrequency},
	{"downside_potential", (*PartialMoments).DownsidePotential},
	{"upper_excess_moment_1_sum", (*PartialMoments).UpperExcessMoment1Sum},
	{"upper_excess_moment_2_sum", (*PartialMoments).UpperExcessMoment2Sum},
	{"lower_excess_moment_2_sum", (*PartialMoments).LowerExcessMoment2Sum},
	{"lower_partial_moment_1", (*PartialMoments).LowerPartialMoment1},
	{"lower_partial_moment_2", (*PartialMoments).LowerPartialMoment2},
	{"lower_partial_moment_3", (*PartialMoments).LowerPartialMoment3},
	{"lower_partial_moment_4", (*PartialMoments).LowerPartialMoment4},
	{"higher_partial_moment_1", (*PartialMoments).HigherPartialMoment1},
	{"higher_partial_moment_2", (*PartialMoments).HigherPartialMoment2},
	{"higher_partial_moment_3", (*PartialMoments).HigherPartialMoment3},
	{"higher_partial_moment_4", (*PartialMoments).HigherPartialMoment4},
	{"upper_excess_moment_1", (*PartialMoments).UpperExcessMoment1},
	{"upper_excess_moment_2", (*PartialMoments).UpperExcessMoment2},
	{"upper_excess_moment_3", (*PartialMoments).UpperExcessMoment3},
	{"upper_excess_moment_4", (*PartialMoments).UpperExcessMoment4},
	{"lower_excess_moment_1", (*PartialMoments).LowerExcessMoment1},
	{"lower_excess_moment_2", (*PartialMoments).LowerExcessMoment2},
	{"lower_excess_moment_3", (*PartialMoments).LowerExcessMoment3},
	{"lower_excess_moment_4", (*PartialMoments).LowerExcessMoment4},
}

var rawPartialMomentsGetters = []namedGetter[*RawPartialMoments]{
	{"count", intGetter((*RawPartialMoments).Count)},
	{"lower_partial_moment_1", (*RawPartialMoments).LowerPartialMoment1},
	{"higher_partial_moment_1", (*RawPartialMoments).HigherPartialMoment1},
	{"count_negative", intGetter((*RawPartialMoments).CountNegative)},
	{"sum_negative", (*RawPartialMoments).SumNegative},
	{"count_positive", intGetter((*RawPartialMoments).CountPositive)},
	{"sum_positive", (*RawPartialMoments).SumPositive},
}

// partialMomentsReference computes naive partial moments about the
// threshold.
func partialMomentsReference(returns []float64, threshold float64) map[string]float64 {
	var lower, upper []float64
	for _, r := range returns {
		if r < threshold {
			lower = append(lower, threshold-r)
		}
	}
	for _, r := range returns {
		if r > threshold {
			upper = append(upper, r-threshold)
		}
	}
	n := float64(len(returns))
	freq := func(k int) float64 {
		if n != 0 {
			return float64(k) / n
		}
		return math.NaN()
	}
	sq := func(x float64) float64 { return x * x }
	shortfall := func(r float64) float64 { return math.Max(threshold-r, 0.0) }
	excess := func(r float64) float64 { return math.Max(r-threshold, 0.0) }
	expected := map[string]float64{
		"total_count":               n,
		"lower_excess_count":        float64(len(lower)),
		"upper_excess_count":        float64(len(upper)),
		"downside_frequency":        freq(len(lower)),
		"upside_frequency":          freq(len(upper)),
		"downside_potential":        pyMean(mapf(returns, shortfall)),
		"upper_excess_moment_1_sum": pySum(upper),
		"upper_excess_moment_2_sum": pySum(mapf(upper, sq)),
		"lower_excess_moment_2_sum": pySum(mapf(lower, sq)),
	}
	for k := 1; k <= 4; k++ {
		pow := func(f func(float64) float64) func(float64) float64 {
			return func(x float64) float64 { return math.Pow(f(x), float64(k)) }
		}
		id := func(x float64) float64 { return x }
		expected[fmt.Sprintf("lower_partial_moment_%d", k)] = pyMean(mapf(returns, pow(shortfall)))
		expected[fmt.Sprintf("higher_partial_moment_%d", k)] = pyMean(mapf(returns, pow(excess)))
		expected[fmt.Sprintf("upper_excess_moment_%d", k)] = pyMean(mapf(upper, pow(id)))
		expected[fmt.Sprintf("lower_excess_moment_%d", k)] = pyMean(mapf(lower, pow(id)))
	}
	return expected
}

// rawPartialMomentsReference computes naive raw partial moments (sums,
// threshold 0).
func rawPartialMomentsReference(returns []float64) map[string]float64 {
	neg := filter(returns, func(r float64) bool { return r < 0 })
	pos := filter(returns, func(r float64) bool { return r > 0 })
	return map[string]float64{
		"count":                   float64(len(returns)),
		"lower_partial_moment_1":  pySum(mapf(returns, func(r float64) float64 { return math.Max(-r, 0.0) })),
		"higher_partial_moment_1": pySum(mapf(returns, func(r float64) float64 { return math.Max(r, 0.0) })),
		"count_negative":          float64(len(neg)),
		"sum_negative":            pySum(neg),
		"count_positive":          float64(len(pos)),
		"sum_positive":            pySum(pos),
	}
}

func assertGettersMatch[T any](t *testing.T, obj T, getters []namedGetter[T], expected map[string]float64,
	tol float64, msg string,
) {
	t.Helper()
	if len(getters) != len(expected) {
		t.Fatalf("%d getters, %d expected values", len(getters), len(expected))
	}
	for _, g := range getters {
		e, ok := expected[g.name]
		if !ok {
			t.Fatalf("no expected value for %s", g.name)
		}
		assertNaNOrAlmostEqual(t, g.get(obj), e, tol, msg+" "+g.name)
	}
}

// randomPartialMomentsReturns mixes in values exactly equal to the
// threshold and to zero.
func randomPartialMomentsReturns(seed uint64, n int, threshold float64) []float64 {
	rng := newRNG(seed)
	returns := make([]float64, n)
	for i := range returns {
		g1 := gauss(rng, 0.0, 0.03)
		g2 := gauss(rng, 0.0, 0.03)
		returns[i] = choice(rng, threshold, 0.0, g1, g2)
	}
	return returns
}

func TestPartialMomentsEmpty(t *testing.T) {
	t.Parallel()
	pm := NewPartialMoments(0.01)
	if pm.TotalCount() != 0 {
		t.Errorf("total count %d, expected 0", pm.TotalCount())
	}
	assertNaN(t, pm.DownsideFrequency(), "downside frequency")
	assertNaN(t, pm.UpsideFrequency(), "upside frequency")
	assertNaN(t, pm.LowerPartialMoment2(), "lower partial moment 2")
}

func TestPartialMomentsHandComputed(t *testing.T) {
	t.Parallel()
	pm := NewPartialMoments(0.01)
	for _, r := range []float64{0.03, -0.01, 0.01, 0.00} {
		pm.Update(r)
	}
	// Runtime variables so the expressions are evaluated in float64.
	v002, v001, v00004, v00001 := 0.02, 0.01, 0.0004, 0.0001
	// Shortfalls below 1%: 0.02, 0.01; excesses above: 0.02.
	assertAlmostEqual(t, pm.LowerPartialMoment1(), (v002+v001)/4, places(16), "lpm1")
	assertAlmostEqual(t, pm.LowerPartialMoment2(), (v00004+v00001)/4, places(16), "lpm2")
	assertAlmostEqual(t, pm.HigherPartialMoment1(), v002/4, places(16), "hpm1")
	if pm.LowerExcessCount() != 2 {
		t.Errorf("lower excess count %d, expected 2", pm.LowerExcessCount())
	}
	if pm.UpperExcessCount() != 1 {
		t.Errorf("upper excess count %d, expected 1", pm.UpperExcessCount())
	}
	assertExact(t, pm.DownsideFrequency(), 0.5, "downside frequency")
	assertExact(t, pm.UpsideFrequency(), 0.25, "upside frequency")
}

func TestPartialMomentsMatchesReference(t *testing.T) {
	t.Parallel()
	for _, threshold := range []float64{0.0, 0.005} {
		returns := randomPartialMomentsReturns(42, 200, threshold)
		pm := NewPartialMoments(threshold)
		for _, r := range returns {
			pm.Update(r)
		}
		assertGettersMatch(t, pm, partialMomentsGetters, partialMomentsReference(returns, threshold),
			places(14), fmt.Sprintf("threshold %v", threshold))
	}
}

func TestPartialMomentsRollingWindowMatchesReference(t *testing.T) {
	t.Parallel()
	threshold := 0.005
	returns := randomPartialMomentsReturns(7, 120, threshold)
	const w = 8
	pm := NewPartialMoments(threshold)
	for i, r := range returns {
		if i >= w {
			pm.Revert(returns[i-w])
		}
		pm.Update(r)
		assertGettersMatch(t, pm, partialMomentsGetters, partialMomentsReference(window(returns, w, i), threshold),
			places(13), fmt.Sprintf("step %d", i))
	}
}

func TestPartialMomentsReset(t *testing.T) {
	t.Parallel()
	pm := NewPartialMoments(0.0)
	for _, r := range []float64{0.01, -0.02} {
		pm.Update(r)
	}
	pm.Reset()
	if pm.TotalCount() != 0 || pm.LowerExcessCount() != 0 || pm.UpperExcessCount() != 0 {
		t.Errorf("counts %d %d %d, expected 0", pm.TotalCount(), pm.LowerExcessCount(), pm.UpperExcessCount())
	}
}

func TestRawPartialMomentsEmpty(t *testing.T) {
	t.Parallel()
	assertGettersMatch(t, NewRawPartialMoments(), rawPartialMomentsGetters,
		rawPartialMomentsReference(nil), places(14), "")
}

func TestRawPartialMomentsMatchesReference(t *testing.T) {
	t.Parallel()
	returns := randomPartialMomentsReturns(42, 200, 0.0)
	pm := NewRawPartialMoments()
	for _, r := range returns {
		pm.Update(r)
	}
	assertGettersMatch(t, pm, rawPartialMomentsGetters, rawPartialMomentsReference(returns), places(14), "")
}

func TestRawPartialMomentsRollingWindowMatchesReference(t *testing.T) {
	t.Parallel()
	returns := randomPartialMomentsReturns(7, 120, 0.0)
	const w = 8
	pm := NewRawPartialMoments()
	for i, r := range returns {
		if i >= w {
			pm.Revert(returns[i-w])
		}
		pm.Update(r)
		assertGettersMatch(t, pm, rawPartialMomentsGetters, rawPartialMomentsReference(window(returns, w, i)),
			places(14), fmt.Sprintf("step %d", i))
	}
}

func TestRawPartialMomentsReset(t *testing.T) {
	t.Parallel()
	pm := NewRawPartialMoments()
	for _, r := range []float64{0.01, -0.02} {
		pm.Update(r)
	}
	pm.Reset()
	assertGettersMatch(t, pm, rawPartialMomentsGetters, rawPartialMomentsReference(nil), places(14), "")
}
