package performance

import (
	"fmt"
	"math"
	"math/rand/v2"
	"testing"
)

// From the 'Portfolio bacon' dataset of the PerformanceAnalytics R package:
// page 65 (portfolio) and page 66 (benchmark) of Carl R. Bacon, Practical
// portfolio performance: measurement and attribution, 2nd ed, Wiley, 2008.
var baconPortfolioReturns = []float64{
	0.003, 0.026, 0.011, -0.010,
	0.015, 0.025, 0.016, 0.067,
	-0.014, 0.040, -0.005, 0.081,
	0.040, -0.037, -0.061, 0.017,
	-0.049, -0.022, 0.070, 0.058,
	-0.065, 0.024, -0.005, -0.009,
}

var baconBenchmarkReturns = []float64{
	0.002, 0.025, 0.018, -0.011,
	0.014, 0.018, 0.014, 0.065,
	-0.015, 0.042, -0.006, 0.083,
	0.039, -0.038, -0.062, 0.015,
	-0.048, 0.021, 0.060, 0.056,
	-0.067, 0.019, -0.003, 0.000,
}

var baconPortfolioLen = len(baconPortfolioReturns)

// Extended Bacon 2023 (3rd edition) portfolio data: Carl R. Bacon,
// Practical portfolio performance measurement and attribution, 3rd ed,
// Wiley, 2023.
var bacon2023PortfolioReturns = []float64{
	0.003, 0.026, 0.011, -0.009, 0.014, 0.024, 0.015, 0.066, -0.014, 0.039,
	-0.005, 0.081, 0.040, -0.037, -0.061, 0.014, -0.049, -0.021, 0.062, 0.058,
	-0.064, 0.017, -0.004, -0.002, -0.021, 0.011, 0.047, 0.024, 0.033, -0.007,
	0.047, 0.006, 0.010, -0.002, 0.034, 0.010,
}

var bacon2023DrawdownContinuousWithoutZeroes = []float64{
	-0.0090, -0.0140, -0.0050, -0.0960, -0.0690,
	-0.0640, -0.0270, -0.0070, -0.0020,
}

var bacon2023DrawdownFromPeak = []float64{
	0, 0, 0, -0.0090, 0, 0, 0, 0, -0.0140, 0,
	-0.0050, 0, 0, -0.0370, -0.0957, -0.0831, -0.1280, -0.1463, -0.0934, -0.0408,
	-0.1022, -0.0869, -0.0906, -0.0924, -0.1115, -0.1017, -0.0595, -0.0369, -0.0051, -0.0121,
	0, 0, 0, -0.0020, 0, 0,
}

var bacon2023PortfolioLen = len(bacon2023PortfolioReturns)

// tolerance mirrors the places / delta / rel_tol arguments of the Python
// assertion helpers.
type tolerance struct {
	places   int
	delta    float64
	hasDelta bool
	relTol   float64
	hasRel   bool
}

// places mirrors places=n (round(a-b, n) == 0, approximately
// |a-b| <= 0.5·10⁻ⁿ).
func places(n int) tolerance { return tolerance{places: n} }

// delta mirrors delta=d.
func delta(d float64) tolerance { return tolerance{delta: d, hasDelta: true} }

// relTol mirrors rel_tol=r with delta=d as the absolute tolerance
// (math.isclose(actual, expected, rel_tol=r, abs_tol=d)).
func relTol(r, d float64) tolerance {
	return tolerance{relTol: r, hasRel: true, delta: d, hasDelta: true}
}

// checkFloatEqual mirrors assertFloatEqual: NaN and ±Inf must match
// exactly, finite values within the tolerance.
func checkFloatEqual(actual, expected float64, tol tolerance, prefix string) error {
	switch {
	case math.IsNaN(expected):
		if !math.IsNaN(actual) {
			return fmt.Errorf("%s: expected NaN, actual %v", prefix, actual)
		}
	case math.IsInf(expected, 1):
		if !math.IsInf(actual, 1) {
			return fmt.Errorf("%s: expected +Inf, actual %v", prefix, actual)
		}
	case math.IsInf(expected, -1):
		if !math.IsInf(actual, -1) {
			return fmt.Errorf("%s: expected -Inf, actual %v", prefix, actual)
		}
	default:
		if math.IsNaN(actual) || math.IsInf(actual, 0) {
			return fmt.Errorf("%s: expected %v, got %v", prefix, expected, actual)
		}
		diff := math.Abs(actual - expected)
		var ok bool
		switch {
		case tol.hasRel:
			ok = diff <= math.Max(tol.relTol*math.Max(math.Abs(actual), math.Abs(expected)), tol.delta)
		case tol.hasDelta:
			ok = actual == expected || diff <= tol.delta
		default:
			ok = actual == expected || diff <= 0.5*math.Pow10(-tol.places)
		}
		if !ok {
			return fmt.Errorf("%s: expected %v, got %v (diff %g)", prefix, expected, actual, diff)
		}
	}
	return nil
}

// checkSeriesEqual mirrors assertSeriesEqual: equal lengths, and every
// element from index skip on equal within the tolerance.
func checkSeriesEqual(actual, expected []float64, tol tolerance, prefix string, skip int) error {
	if len(actual) != len(expected) {
		return fmt.Errorf("%s: series length %d, expected %d", prefix, len(actual), len(expected))
	}
	for i := range actual {
		if i >= skip {
			if err := checkFloatEqual(actual[i], expected[i], tol, fmt.Sprintf("%s step %d", prefix, i)); err != nil {
				return err
			}
		}
	}
	return nil
}

func assertFloatEqual(t *testing.T, actual, expected float64, tol tolerance, prefix string) {
	t.Helper()
	if err := checkFloatEqual(actual, expected, tol, prefix); err != nil {
		t.Error(err)
	}
}

func assertSeriesEqual(t *testing.T, actual, expected []float64, tol tolerance, prefix string) {
	t.Helper()
	if err := checkSeriesEqual(actual, expected, tol, prefix, 0); err != nil {
		t.Error(err)
	}
}

func assertSeriesEqualSkip(t *testing.T, actual, expected []float64, tol tolerance, skip int, prefix string) {
	t.Helper()
	if err := checkSeriesEqual(actual, expected, tol, prefix, skip); err != nil {
		t.Error(err)
	}
}

// assertAlmostEqual mirrors unittest.assertAlmostEqual(places=n).
func assertAlmostEqual(t *testing.T, actual, expected float64, n int, prefix string) {
	t.Helper()
	if !(actual == expected || math.Abs(actual-expected) <= 0.5*math.Pow10(-n)) {
		t.Errorf("%s: expected %v, got %v (diff %g)", prefix, expected, actual, math.Abs(actual-expected))
	}
}

func periodsPerAnnum(daily, monthly bool) float64 {
	if daily && monthly {
		panic("Only one of daily or monthly can be True")
	}
	if daily {
		return 252
	}
	if monthly {
		return 12
	}
	return 1
}

// stream mirrors the keyword arguments of the Python run_stream_* helpers.
// Nil returns / benchmark default to the Bacon data.
type stream struct {
	daily, monthly     bool
	annualRiskFreeRate float64
	annualTargetReturn float64
	returns            []float64
	benchmark          []float64
	window             int
	start              int
}

func newTestMeasures(t *testing.T, ppa, annualRf, annualMar float64, window int) *Measures {
	t.Helper()
	m, err := NewMeasures(ppa, annualRf, annualMar, window)
	if err != nil {
		t.Fatalf("NewMeasures: %v", err)
	}
	return m
}

// runStreamAny mirrors run_stream_callback for any value type.
func runStreamAny[T any](t *testing.T, s stream, f func(*Measures) T) []T {
	t.Helper()
	returns, benchmark := s.returns, s.benchmark
	if returns == nil {
		returns = baconPortfolioReturns
	}
	if benchmark == nil {
		benchmark = baconBenchmarkReturns
	}
	m := newTestMeasures(t, periodsPerAnnum(s.daily, s.monthly),
		s.annualRiskFreeRate, s.annualTargetReturn, s.window)
	m.Reset()
	results := make([]T, 0, len(returns))
	for i := s.start; i < len(returns); i++ {
		m.AddReturn(returns[i], benchmark[i])
		results = append(results, f(m))
	}
	return results
}

// runStream mirrors run_stream_property / run_stream_method /
// run_stream_callback for float-valued measures.
func runStream(t *testing.T, s stream, f func(*Measures) float64) []float64 {
	t.Helper()
	return runStreamAny(t, s, f)
}

// must returns v, panicking (and so failing the test) on a non-nil error.
func must[T any](v T, err error) T {
	if err != nil {
		panic(fmt.Sprintf("unexpected error: %v", err))
	}
	return v
}

func makeMeasures(t *testing.T, window int, annualRf, annualMar float64, daily, monthly bool) *Measures {
	t.Helper()
	m := newTestMeasures(t, periodsPerAnnum(daily, monthly), annualRf, annualMar, window)
	m.Reset()
	return m
}

// addBacon mirrors add_bacon; nil returns / benchmark default to the Bacon
// data.
func addBacon(m *Measures, returns, benchmark []float64) {
	if returns == nil {
		returns = baconPortfolioReturns
	}
	if benchmark == nil {
		benchmark = baconBenchmarkReturns
	}
	for i := range baconPortfolioLen {
		m.AddReturn(returns[i], benchmark[i])
	}
}

// repeat returns n copies of v (Python [v] * n).
func repeat(v float64, n int) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = v
	}
	return out
}

// annualize returns (1 + r)^p - 1.
func annualize(r, p float64) float64 {
	return math.Pow(1+r, p) - 1
}

// fsum mirrors Python's math.fsum (Shewchuk's exact summation with
// correct rounding).
func fsum(values []float64) float64 {
	partials := []float64{}
	for _, x := range values {
		i := 0
		for _, y := range partials {
			if math.Abs(x) < math.Abs(y) {
				x, y = y, x
			}
			hi := x + y
			lo := y - (hi - x)
			if lo != 0 {
				partials[i] = lo
				i++
			}
			x = hi
		}
		partials = append(partials[:i], x)
	}
	n := len(partials)
	hi := 0.0
	if n > 0 {
		n--
		hi = partials[n]
		lo := 0.0
		for n > 0 {
			x := hi
			n--
			y := partials[n]
			hi = x + y
			yr := hi - x
			lo = y - yr
			if lo != 0 {
				break
			}
		}
		if n > 0 && ((lo < 0 && partials[n-1] < 0) || (lo > 0 && partials[n-1] > 0)) {
			y := lo * 2
			x := hi + y
			yr := x - hi
			if y == yr {
				hi = x
			}
		}
	}
	return hi
}

// prod mirrors math.prod: the product of values, left to right.
func prod(values []float64) float64 {
	p := 1.0
	for _, v := range values {
		p *= v
	}
	return p
}

// onePlus returns 1 + v for every value.
func onePlus(values []float64) []float64 {
	out := make([]float64, len(values))
	for i, v := range values {
		out[i] = 1 + v
	}
	return out
}

// fmean mirrors statistics.fmean.
func fmean(values []float64) float64 {
	return fsum(values) / float64(len(values))
}

// sumSquaredDeviations returns Σ(x - mean)² using exact summation.
func sumSquaredDeviations(values []float64) float64 {
	mean := fmean(values)
	sq := make([]float64, len(values))
	for i, v := range values {
		sq[i] = (v - mean) * (v - mean)
	}
	return fsum(sq)
}

// pstdev approximates statistics.pstdev.
func pstdev(values []float64) float64 {
	return math.Sqrt(sumSquaredDeviations(values) / float64(len(values)))
}

// stdev approximates statistics.stdev.
func stdev(values []float64) float64 {
	return math.Sqrt(sumSquaredDeviations(values) / float64(len(values)-1))
}

// newRNG is the counterpart of Python's random.Random(seed).
func newRNG(seed uint64) *rand.Rand {
	return rand.New(rand.NewPCG(seed, 0))
}

// gauss mirrors random.gauss(mu, sigma).
func gauss(rng *rand.Rand, mu, sigma float64) float64 {
	return mu + sigma*rng.NormFloat64()
}
