package streamingkbn

import "math"

// Raw moments with Klein KBN (Kahan-Babuška-Neumaier)
// compensated summation for improved numerical stability.
//
// References:
//   - https://github.com/kuiperzone/Compensated-Accumulators
//   - https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance

// cancellationEpsilon is the relative threshold below which the population
// variance μ₂, computed from raw power sums as Σx²/n − (Σx/n)², is
// considered to be lost in rounding error (i.e. indistinguishable from zero).
const cancellationEpsilon = 1e-14

// RawMomentsKleinKBN computes streaming mean, variance, skewness, kurtosis via
// raw power sums (x¹..x⁴) with Klein KBN (Kahan-Babuška-Neumaier) compensated
// accumulation.
//
// It accumulates Σx, Σx², Σx³, Σx⁴ using a [KleinKBNAccumulator] for each,
// plus a separate Welford mean/variance tracker (also KBN-compensated).
// Mean and variance come from the Welford tracker; skewness and kurtosis
// are converted from the raw power sums at query time.
//
// Supports removal of any previously added sample with Revert, not
// only the most recent one, because both the power sums and Welford's
// mean/M₂ are symmetric functions of the samples. This makes the type
// suitable for FIFO rolling windows (update the new sample, revert the
// oldest one).
//
// Accuracy caveat: converting raw power sums to central moments suffers
// catastrophic cancellation when the mean is large compared to the spread
// (e.g. prices rather than returns). For such data skewness and kurtosis
// lose precision. Use [CentralMomentsKleinKBN] for such data; it supports
// FIFO removal, but repeated removals clear compensation and can accumulate
// rounding error.
//
// # Notation
//
// μₖ' = Σxᵏ / n are the raw moments (methods X1..X4) and μₖ are the
// population central moments derived from them:
//
//	μ₂ = μ₂' − μ₁'²
//	μ₃ = μ₃' − μ₁'³ − 3·μ₁'·μ₂
//	μ₄ = μ₄' − μ₁'⁴ − 6·μ₁'²·μ₂ − 4·μ₁'·μ₃
//
// g₁ = μ₃ / μ₂^1.5 is the population skewness and β₂ = μ₄ / μ₂² is the
// population (Pearson) kurtosis. All skewness and kurtosis methods
// return NaN when n < 2 or μ₂ is zero relative to μ₂' (constant data).
//
// # Parameters
//
//   - ddof (nonnegative int, conventionally 1): delta degrees of freedom for
//     Variance and StandardDeviation. variance = Σ(x - x̄)² / (n - ddof).
//     ddof=0 gives population, ddof=1 gives sample.
//   - bias (conventionally true): selects the Skewness and Kurtosis
//     methods, see below.
//   - fisher (conventionally true): selects the Kurtosis method, see below.
//
// The Skewness and Kurtosis methods dispatch as follows
// (matching scipy.stats.skew and scipy.stats.kurtosis):
//
//	bias   fisher  Skewness                Kurtosis
//	true   true    SkewnessMoment, g₁      KurtosisExcess, β₂ − 3
//	true   false   SkewnessMoment, g₁      KurtosisMoment, β₂
//	false  true    SkewnessFisher, G₁      KurtosisSampleExcess, G₂
//	false  false   SkewnessFisher, G₁      KurtosisSample, G₂ + 3
//
// The remaining variants (SkewnessSample, KurtosisSampleCorrected),
// which match the R PerformanceAnalytics package, are available as
// separate methods.
type RawMomentsKleinKBN struct {
	n      int
	x1     KleinKBNAccumulator
	x2     KleinKBNAccumulator
	x3     KleinKBNAccumulator
	x4     KleinKBNAccumulator
	ddof   int
	bias   bool
	fisher bool
	// Welford's mean and sum of squared deviations Σ(x - x̄)².
	mean KleinKBNAccumulator
	s    KleinKBNAccumulator
}

// NewRawMomentsKleinKBN returns a new empty RawMomentsKleinKBN.
// The conventional defaults are ddof=1, bias=true, fisher=true.
//
// Panics if ddof is negative.
func NewRawMomentsKleinKBN(ddof int, bias, fisher bool) *RawMomentsKleinKBN {
	if ddof < 0 {
		panic("ddof must be a nonnegative integer")
	}
	return &RawMomentsKleinKBN{ddof: ddof, bias: bias, fisher: fisher}
}

// Ddof returns the delta degrees of freedom used by Variance and
// StandardDeviation.
func (m *RawMomentsKleinKBN) Ddof() int { return m.ddof }

// SetDdof sets the delta degrees of freedom used by Variance and
// StandardDeviation.
//
// Panics if ddof is negative.
func (m *RawMomentsKleinKBN) SetDdof(ddof int) {
	if ddof < 0 {
		panic("ddof must be a nonnegative integer")
	}
	m.ddof = ddof
}

// Bias returns the bias parameter selecting Skewness and Kurtosis.
func (m *RawMomentsKleinKBN) Bias() bool { return m.bias }

// SetBias sets the bias parameter selecting Skewness and Kurtosis.
func (m *RawMomentsKleinKBN) SetBias(bias bool) { m.bias = bias }

// Fisher returns the fisher parameter selecting Kurtosis.
func (m *RawMomentsKleinKBN) Fisher() bool { return m.fisher }

// SetFisher sets the fisher parameter selecting Kurtosis.
func (m *RawMomentsKleinKBN) SetFisher(fisher bool) { m.fisher = fisher }

// Reset clears all accumulated state.
func (m *RawMomentsKleinKBN) Reset() {
	m.n = 0
	m.x1.Reset()
	m.x2.Reset()
	m.x3.Reset()
	m.x4.Reset()
	m.mean.Reset()
	m.s.Reset()
}

// Update adds a sample x.
func (m *RawMomentsKleinKBN) Update(x float64) {
	m.n++
	m.x1.Update(x)
	x2 := x * x
	m.x2.Update(x2)
	x3 := x2 * x
	m.x3.Update(x3)
	x4 := x3 * x
	m.x4.Update(x4)
	// Welford: mean += (x - mean_old) / n;  S += (x - mean_old)·(x - mean_new)
	delta := x - m.mean.Value()
	m.mean.Update(delta / float64(m.n))
	m.s.Update(delta * (x - m.mean.Value()))
}

// Revert removes a previously added sample x (any sample, not only the most
// recent one). Reverting a value that was never added corrupts the state.
//
// Panics if there are no samples.
func (m *RawMomentsKleinKBN) Revert(x float64) {
	if m.n <= 0 {
		panic("Cannot revert from an empty accumulator")
	}
	if m.n == 1 {
		m.Reset()
		return
	}
	m.n--
	m.x1.Revert(x)
	x2 := x * x
	m.x2.Revert(x2)
	x3 := x2 * x
	m.x3.Revert(x3)
	x4 := x3 * x
	m.x4.Revert(x4)
	// Inverse Welford: mean_old = mean_new - (x - mean_new) / (n - 1);
	// S -= (x - mean_new)·(x - mean_old)
	delta := x - m.mean.Value()
	m.mean.Revert(delta / float64(m.n))
	m.s.Revert(delta * (x - m.mean.Value()))
}

func (m *RawMomentsKleinKBN) variance(ddof int) float64 {
	// A slightly negative S caused by rounding after Revert is clamped to zero.
	d := m.n - ddof
	if d <= 0 {
		return math.NaN()
	}
	s := m.s.Value()
	if s < 0 {
		s = 0
	}
	return s / float64(d)
}

func (m *RawMomentsKleinKBN) standardDeviation(ddof int) float64 {
	v := m.variance(ddof)
	if math.IsNaN(v) {
		return v
	}
	return math.Sqrt(v)
}

// Mean returns the arithmetic mean (0.0 when empty).
func (m *RawMomentsKleinKBN) Mean() float64 {
	return m.mean.Value()
}

// Variance returns the variance Σ(x - x̄)² / (n - ddof), NaN when n ≤ ddof.
func (m *RawMomentsKleinKBN) Variance() float64 {
	return m.variance(m.ddof)
}

// VarianceDdof0 returns the population variance Σ(x - x̄)² / n, regardless of ddof.
func (m *RawMomentsKleinKBN) VarianceDdof0() float64 {
	return m.variance(0)
}

// VarianceDdof1 returns the sample variance Σ(x - x̄)² / (n - 1), regardless of ddof.
func (m *RawMomentsKleinKBN) VarianceDdof1() float64 {
	return m.variance(1)
}

// StandardDeviation returns the square root of Variance.
func (m *RawMomentsKleinKBN) StandardDeviation() float64 {
	return m.standardDeviation(m.ddof)
}

// StandardDeviationDdof0 returns the population standard deviation, regardless of ddof.
func (m *RawMomentsKleinKBN) StandardDeviationDdof0() float64 {
	return m.standardDeviation(0)
}

// StandardDeviationDdof1 returns the sample standard deviation, regardless of ddof.
func (m *RawMomentsKleinKBN) StandardDeviationDdof1() float64 {
	return m.standardDeviation(1)
}

// centralMoments converts the raw power sums to population central moments
// (μ₂, μ₃, μ₄), see the type Notation.
//
// Returns ok=false when n < 2 or μ₂ is lost in rounding error.
func (m *RawMomentsKleinKBN) centralMoments() (mu2, mu3, mu4 float64, ok bool) {
	n := m.n
	if n < 2 {
		return 0, 0, 0, false
	}
	fn := float64(n)
	mu1 := m.x1.Value() / fn
	r := mu1 * mu1
	meanX2 := m.x2.Value() / fn
	mu2 = meanX2 - r
	if mu2 <= cancellationEpsilon*meanX2 {
		return 0, 0, 0, false
	}
	r *= mu1
	mu3 = m.x3.Value()/fn - r - 3*mu1*mu2
	r *= mu1
	mu4 = m.x4.Value()/fn - r - 6*mu2*mu1*mu1 - 4*mu3*mu1
	return mu2, mu3, mu4, true
}

// g1 returns the population skewness g₁ = μ₃ / μ₂^1.5.
func (m *RawMomentsKleinKBN) g1() float64 {
	mu2, mu3, _, ok := m.centralMoments()
	if !ok {
		return math.NaN()
	}
	return mu3 / (mu2 * math.Sqrt(mu2))
}

// b2 returns the population (Pearson) kurtosis β₂ = μ₄ / μ₂².
func (m *RawMomentsKleinKBN) b2() float64 {
	mu2, _, mu4, ok := m.centralMoments()
	if !ok {
		return math.NaN()
	}
	return mu4 / (mu2 * mu2)
}

// SkewnessMoment returns the 'moment' (biased, population) skewness,
// requires n ≥ 2:
//
//	g₁ = μ₃ / μ₂^1.5
//
// Matches scipy.stats.skew(bias=True) and PerformanceAnalytics
// skewness(method="moment").
func (m *RawMomentsKleinKBN) SkewnessMoment() float64 {
	return m.g1()
}

// SkewnessFisher returns the 'fisher' (bias-adjusted Fisher-Pearson)
// skewness, requires n ≥ 3:
//
//	G₁ = g₁ · √(n(n−1)) / (n−2)
//
// Matches scipy.stats.skew(bias=False) and PerformanceAnalytics
// skewness(method="fisher").
func (m *RawMomentsKleinKBN) SkewnessFisher() float64 {
	g1 := m.g1()
	if math.IsNaN(g1) {
		return math.NaN()
	}
	n := m.n
	if n < 3 {
		return math.NaN()
	}
	fn := float64(n)
	return g1 * math.Sqrt(fn*(fn-1)) / (fn - 2)
}

// SkewnessSample returns the 'sample' skewness, requires n ≥ 3:
//
//	g₁ · n² / ((n−1)(n−2))
//
// Matches PerformanceAnalytics skewness(method="sample").
// Doesn't depend on the bias parameter.
func (m *RawMomentsKleinKBN) SkewnessSample() float64 {
	g1 := m.g1()
	if math.IsNaN(g1) {
		return math.NaN()
	}
	n := m.n
	if n < 3 {
		return math.NaN()
	}
	fn := float64(n)
	return g1 * (fn * fn) / ((fn - 1) * (fn - 2))
}

// Skewness returns the skewness selected by the bias parameter:
//
//   - bias=true:  SkewnessMoment, g₁
//   - bias=false: SkewnessFisher, G₁ = g₁ · √(n(n−1)) / (n−2)
//
// The third variant, SkewnessSample, doesn't depend on bias.
func (m *RawMomentsKleinKBN) Skewness() float64 {
	if m.bias {
		return m.SkewnessMoment()
	}
	return m.SkewnessFisher()
}

// KurtosisMoment returns the 'moment' (biased, population) Pearson kurtosis,
// requires n ≥ 2:
//
//	β₂ = μ₄ / μ₂²
//
// Matches scipy.stats.kurtosis(bias=True, fisher=False) and
// PerformanceAnalytics kurtosis(method="moment").
func (m *RawMomentsKleinKBN) KurtosisMoment() float64 {
	return m.b2()
}

// KurtosisExcess returns the 'excess' (biased, population) excess kurtosis,
// requires n ≥ 2:
//
//	β₂ − 3
//
// Matches scipy.stats.kurtosis(bias=True, fisher=True) and
// PerformanceAnalytics kurtosis(method="excess").
func (m *RawMomentsKleinKBN) KurtosisExcess() float64 {
	b2 := m.b2()
	if math.IsNaN(b2) {
		return math.NaN()
	}
	return b2 - 3
}

// KurtosisSampleExcess returns the 'sample excess' (unbiased) excess
// kurtosis, requires n ≥ 4:
//
//	G₂ = ((n²−1)·β₂ − 3(n−1)²) / ((n−2)(n−3))
//
// Matches scipy.stats.kurtosis(bias=False, fisher=True) and
// PerformanceAnalytics kurtosis(method="sample_excess").
func (m *RawMomentsKleinKBN) KurtosisSampleExcess() float64 {
	b2 := m.b2()
	if math.IsNaN(b2) {
		return math.NaN()
	}
	n := m.n
	if n <= 3 {
		return math.NaN()
	}
	fn := float64(n)
	return ((fn*fn-1)*b2 - 3*(fn-1)*(fn-1)) / ((fn - 2) * (fn - 3))
}

// KurtosisSample returns the 'sample' (unbiased) Pearson kurtosis, requires
// n ≥ 4, calculated as the 'sample excess' kurtosis plus 3:
//
//	G₂ + 3 = ((n²−1)·β₂ − 3(n−1)²) / ((n−2)(n−3)) + 3
//
// Matches scipy.stats.kurtosis(bias=False, fisher=False).
//
// The PerformanceAnalytics kurtosis(method="sample") variant is
// available as KurtosisSampleCorrected; it is larger than this one
// by (9n−15) / ((n−2)(n−3)), approximately 0.44 for n = 24.
func (m *RawMomentsKleinKBN) KurtosisSample() float64 {
	b2 := m.b2()
	if math.IsNaN(b2) {
		return math.NaN()
	}
	n := m.n
	if n <= 3 {
		return math.NaN()
	}
	fn := float64(n)
	return ((fn*fn-1)*b2-3*(fn-1)*(fn-1))/((fn-2)*(fn-3)) + 3
}

// KurtosisSampleCorrected returns the PerformanceAnalytics 'sample'
// (unbiased) Pearson kurtosis, requires n ≥ 4:
//
//	(n²−1)·β₂ / ((n−2)(n−3))
//
// Matches PerformanceAnalytics kurtosis(method="sample").
// Doesn't depend on the bias and fisher parameters.
//
// It differs from KurtosisSample (G₂ + 3) by (9n−15) / ((n−2)(n−3)),
// approximately 0.44 for n = 24.
func (m *RawMomentsKleinKBN) KurtosisSampleCorrected() float64 {
	b2 := m.b2()
	if math.IsNaN(b2) {
		return math.NaN()
	}
	n := m.n
	if n <= 3 {
		return math.NaN()
	}
	fn := float64(n)
	return b2 * (fn*fn - 1) / ((fn - 2) * (fn - 3))
}

// Kurtosis returns the kurtosis selected by the bias and fisher parameters:
//
//   - bias=true,  fisher=true:  KurtosisExcess, β₂ − 3
//   - bias=true,  fisher=false: KurtosisMoment, β₂
//   - bias=false, fisher=true:  KurtosisSampleExcess, G₂
//   - bias=false, fisher=false: KurtosisSample, G₂ + 3
func (m *RawMomentsKleinKBN) Kurtosis() float64 {
	if m.bias {
		if m.fisher {
			return m.KurtosisExcess()
		}
		return m.KurtosisMoment()
	}
	if m.fisher {
		return m.KurtosisSampleExcess()
	}
	return m.KurtosisSample()
}

// X1Sum returns the sum Σx.
func (m *RawMomentsKleinKBN) X1Sum() float64 { return m.x1.Value() }

// X2Sum returns the sum Σx².
func (m *RawMomentsKleinKBN) X2Sum() float64 { return m.x2.Value() }

// X3Sum returns the sum Σx³.
func (m *RawMomentsKleinKBN) X3Sum() float64 { return m.x3.Value() }

// X4Sum returns the sum Σx⁴.
func (m *RawMomentsKleinKBN) X4Sum() float64 { return m.x4.Value() }

func (m *RawMomentsKleinKBN) rawMoment(a *KleinKBNAccumulator) float64 {
	if m.n > 0 {
		return a.Value() / float64(m.n)
	}
	return math.NaN()
}

// X1 returns the first raw moment Σx / n (NaN when empty).
func (m *RawMomentsKleinKBN) X1() float64 { return m.rawMoment(&m.x1) }

// X2 returns the second raw moment Σx² / n (NaN when empty).
func (m *RawMomentsKleinKBN) X2() float64 { return m.rawMoment(&m.x2) }

// X3 returns the third raw moment Σx³ / n (NaN when empty).
func (m *RawMomentsKleinKBN) X3() float64 { return m.rawMoment(&m.x3) }

// X4 returns the fourth raw moment Σx⁴ / n (NaN when empty).
func (m *RawMomentsKleinKBN) X4() float64 { return m.rawMoment(&m.x4) }

// N returns the number of samples.
func (m *RawMomentsKleinKBN) N() int { return m.n }
