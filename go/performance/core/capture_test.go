package core

import (
	"fmt"
	"testing"
)

type pair struct{ a, b float64 }

var captureGetters = []struct {
	name string
	get  func(*Capture) float64
}{
	{"upside_capture_ratio_geometric", (*Capture).UpsideCaptureRatioGeometric},
	{"upside_capture_ratio_arithmetic", (*Capture).UpsideCaptureRatioArithmetic},
	{"downside_capture_ratio_geometric", (*Capture).DownsideCaptureRatioGeometric},
	{"downside_capture_ratio_arithmetic", (*Capture).DownsideCaptureRatioArithmetic},
	{"up_number_ratio", (*Capture).UpNumberRatio},
	{"down_number_ratio", (*Capture).DownNumberRatio},
	{"up_percentage_ratio", (*Capture).UpPercentageRatio},
	{"down_percentage_ratio", (*Capture).DownPercentageRatio},
}

// captureReference computes naive capture ratios following the
// PerformanceAnalytics conventions: upside periods have benchmark > 0;
// downside capture and down-number use benchmark <= 0; down-percentage
// uses benchmark < 0.
func captureReference(pairs []pair) map[string]float64 {
	var up, dn, dnStrict []pair
	for _, p := range pairs {
		if p.b > 0 {
			up = append(up, p)
		}
		if p.b <= 0 {
			dn = append(dn, p)
		}
		if p.b < 0 {
			dnStrict = append(dnStrict, p)
		}
	}
	as := func(ps []pair) []float64 {
		out := make([]float64, len(ps))
		for i, p := range ps {
			out[i] = p.a
		}
		return out
	}
	bs := func(ps []pair) []float64 {
		out := make([]float64, len(ps))
		for i, p := range ps {
			out[i] = p.b
		}
		return out
	}
	onePlus := func(x float64) float64 { return 1 + x }
	count := func(ps []pair, keep func(pair) bool) float64 {
		n := 0
		for _, p := range ps {
			if keep(p) {
				n++
			}
		}
		return float64(n)
	}
	return map[string]float64{
		"upside_capture_ratio_geometric": nanDiv(
			prod(mapf(as(up), onePlus))-1, prod(mapf(bs(up), onePlus))-1),
		"upside_capture_ratio_arithmetic": nanDiv(pySum(as(up)), pySum(bs(up))),
		"downside_capture_ratio_geometric": nanDiv(
			prod(mapf(as(dn), onePlus))-1, prod(mapf(bs(dn), onePlus))-1),
		"downside_capture_ratio_arithmetic": nanDiv(pySum(as(dn)), pySum(bs(dn))),
		"up_number_ratio":                   nanDiv(count(up, func(p pair) bool { return p.a > 0 }), float64(len(up))),
		"down_number_ratio":                 nanDiv(count(dn, func(p pair) bool { return p.a < 0 }), float64(len(dn))),
		"up_percentage_ratio":               nanDiv(count(up, func(p pair) bool { return p.a > p.b }), float64(len(up))),
		"down_percentage_ratio":             nanDiv(count(dnStrict, func(p pair) bool { return p.a > p.b }), float64(len(dnStrict))),
	}
}

func assertCaptureMatches(t *testing.T, c *Capture, pairs []pair, tol float64, msg string) {
	t.Helper()
	ref := captureReference(pairs)
	for _, g := range captureGetters {
		assertNaNOrAlmostEqual(t, g.get(c), ref[g.name], tol, msg+" "+g.name)
	}
}

func assertCaptureEmpty(t *testing.T, c *Capture) {
	t.Helper()
	for _, g := range captureGetters {
		assertNaN(t, g.get(c), g.name)
	}
}

func randomPairs(seed uint64, n int) []pair {
	rng := newRNG(seed)
	pairs := make([]pair, n)
	for i := range pairs {
		a := zeroOrGauss(rng, 0.03)
		b := zeroOrGauss(rng, 0.03)
		pairs[i] = pair{a, b}
	}
	return pairs
}

func TestCaptureEmpty(t *testing.T) {
	t.Parallel()
	assertCaptureEmpty(t, NewCapture())
}

func TestCaptureHandComputed(t *testing.T) {
	t.Parallel()
	pairs := []pair{{0.02, 0.01}, {-0.01, -0.02}, {0.03, 0.04}, {-0.03, 0.0}, {0.01, -0.01}}
	c := NewCapture()
	for _, p := range pairs {
		c.Update(p.a, p.b)
	}
	// Runtime variables so the expressions are evaluated in float64.
	v005, v102, v103, v101, v104 := 0.05, 1.02, 1.03, 1.01, 1.04
	vm001, vm003, v001, vm002, v0 := -0.01, -0.03, 0.01, -0.02, 0.0
	two, three := 2.0, 3.0
	// Up periods (b > 0): (0.02, 0.01), (0.03, 0.04)
	assertAlmostEqual(t, c.UpsideCaptureRatioArithmetic(), v005/v005, places(15), "upside arithmetic")
	assertAlmostEqual(t, c.UpsideCaptureRatioGeometric(),
		(v102*v103-1)/(v101*v104-1), places(14), "upside geometric")
	if c.UpNumberRatio() != 1.0 {
		t.Errorf("up number ratio: expected 1, got %v", c.UpNumberRatio())
	}
	if c.UpPercentageRatio() != 0.5 {
		t.Errorf("up percentage ratio: expected 0.5, got %v", c.UpPercentageRatio())
	}
	// Down periods (b <= 0) include the zero-benchmark period.
	assertAlmostEqual(t, c.DownsideCaptureRatioArithmetic(),
		(vm001+vm003+v001)/(vm002+v0+vm001), places(15), "downside arithmetic")
	assertAlmostEqual(t, c.DownNumberRatio(), two/three, places(15), "down number ratio")
	// Down-percentage only counts strictly negative benchmark periods.
	if c.DownPercentageRatio() != 1.0 {
		t.Errorf("down percentage ratio: expected 1, got %v", c.DownPercentageRatio())
	}
}

func TestCaptureMatchesReference(t *testing.T) {
	t.Parallel()
	pairs := randomPairs(42, 200)
	c := NewCapture()
	for _, p := range pairs {
		c.Update(p.a, p.b)
	}
	assertCaptureMatches(t, c, pairs, places(12), "")
}

func TestCaptureRollingWindowMatchesReference(t *testing.T) {
	t.Parallel()
	pairs := randomPairs(7, 120)
	const w = 9
	c := NewCapture()
	for i, p := range pairs {
		if i >= w {
			old := pairs[i-w]
			c.Revert(old.a, old.b)
		}
		c.Update(p.a, p.b)
		assertCaptureMatches(t, c, pairs[max(0, i-w+1):i+1], places(12), fmt.Sprintf("step %d", i))
	}
}

func TestCaptureReset(t *testing.T) {
	t.Parallel()
	c := NewCapture()
	c.Update(0.01, 0.02)
	c.Update(-0.01, -0.02)
	c.Reset()
	assertCaptureEmpty(t, c)
}
