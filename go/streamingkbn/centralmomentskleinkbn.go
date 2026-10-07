package streamingkbn

import "math"

// Central moments with Klein KBN (Kahan-Babuška-Neumaier)
// compensated summation for improved numerical stability.
//
// References:
//   - P. Pébay, "Formulas for Robust, One-Pass Parallel Computation
//     of Covariances and Arbitrary-Order Statistical Moments",
//     Sandia Report SAND2008-6212 (2008).
//   - https://www.johndcook.com/skewness_kurtosis.html
//   - https://github.com/kuiperzone/Compensated-Accumulators

// CentralMomentsKleinKBN computes streaming mean, variance, skewness, kurtosis
// via Pébay's central moment update with Klein KBN (Kahan-Babuška-Neumaier)
// compensated accumulation.
//
// It maintains the mean M₁ and the sums of central powers
//
//	M₂ = Σ(x - x̄)²,   M₃ = Σ(x - x̄)³,   M₄ = Σ(x - x̄)⁴
//
// (each as a [KleinKBNAccumulator]), updated in O(1) per sample.
// The population central moments are μₖ = Mₖ / n.
//
// Avoids the catastrophic cancellation inherent in converting raw power
// sums Σxᵏ to central moments. This matters for data with a large mean
// relative to its spread. Inverse updates can remove any previously
// added sample, including the oldest sample in a FIFO rolling window.
// Reversion clears the compensation terms, so repeated removals can
// accumulate rounding error.
//
// # Parameters
//
//   - ddof (nonnegative int, conventionally 1): delta degrees of freedom for
//     variance. variance = M₂ / (n - ddof). ddof=0 gives population, ddof=1
//     gives sample.
//   - bias (conventionally true): if true, return the biased (population)
//     skewness and kurtosis. If false, apply the bias corrections (see Notes).
//   - fisher (conventionally true): if true, return excess kurtosis
//     (subtract 3 so Gaussian→0). If false, return raw (Pearson) kurtosis
//     (Gaussian→3). Applied after the bias correction when bias=false.
//
// # Notes
//
// The results match scipy.stats.skew(bias=...) and
// scipy.stats.kurtosis(bias=..., fisher=...).
//
// Skewness (bias=true), requires n ≥ 2:
//
//	g₁ = μ₃ / μ₂^1.5 = √n · M₃ / M₂^1.5
//
// Skewness (bias=false), requires n ≥ 3:
//
//	G₁ = g₁ · √(n·(n-1)) / (n-2)
//
// Kurtosis (bias=true), requires n ≥ 2:
//
//	β₂ = μ₄ / μ₂² = n · M₄ / M₂²
//	fisher=true:  g₂ = β₂ - 3
//	fisher=false: β₂
//
// Kurtosis (bias=false), requires n ≥ 4:
//
//	G₂ = ((n²-1) · β₂  -  3·(n-1)²) / ((n-2)·(n-3))
//	fisher=true:  G₂
//	fisher=false: G₂ + 3
//
// Skewness and kurtosis are NaN when M₂ = 0 (constant data).
type CentralMomentsKleinKBN struct {
	ddof   int
	bias   bool
	fisher bool
	n      int
	m1     KleinKBNAccumulator
	m2     KleinKBNAccumulator
	m3     KleinKBNAccumulator
	m4     KleinKBNAccumulator
}

// NewCentralMomentsKleinKBN returns a new empty CentralMomentsKleinKBN.
// The conventional defaults are ddof=1, bias=true, fisher=true.
//
// Panics if ddof is negative.
func NewCentralMomentsKleinKBN(ddof int, bias, fisher bool) *CentralMomentsKleinKBN {
	if ddof < 0 {
		panic("ddof must be a nonnegative integer")
	}
	return &CentralMomentsKleinKBN{ddof: ddof, bias: bias, fisher: fisher}
}

// Ddof returns the delta degrees of freedom used by Variance and
// StandardDeviation.
func (m *CentralMomentsKleinKBN) Ddof() int { return m.ddof }

// SetDdof sets the delta degrees of freedom used by Variance and
// StandardDeviation.
//
// Panics if ddof is negative.
func (m *CentralMomentsKleinKBN) SetDdof(ddof int) {
	if ddof < 0 {
		panic("ddof must be a nonnegative integer")
	}
	m.ddof = ddof
}

// Bias returns the bias parameter used by Skewness and Kurtosis.
func (m *CentralMomentsKleinKBN) Bias() bool { return m.bias }

// SetBias sets the bias parameter used by Skewness and Kurtosis.
func (m *CentralMomentsKleinKBN) SetBias(bias bool) { m.bias = bias }

// Fisher returns the fisher parameter used by Kurtosis.
func (m *CentralMomentsKleinKBN) Fisher() bool { return m.fisher }

// SetFisher sets the fisher parameter used by Kurtosis.
func (m *CentralMomentsKleinKBN) SetFisher(fisher bool) { m.fisher = fisher }

// Reset clears all accumulated state.
func (m *CentralMomentsKleinKBN) Reset() {
	m.n = 0
	m.m1.Reset()
	m.m2.Reset()
	m.m3.Reset()
	m.m4.Reset()
}

// Update adds a sample x using Pébay's update (n = count after adding x):
//
//	δ    = x − M₁
//	δₙ   = δ / n
//	term = δ · δₙ · (n − 1)
//
//	M₁ += δₙ
//	M₄ += term·δₙ²·(n²−3n+3) + 6·δₙ²·M₂ − 4·δₙ·M₃
//	M₃ += term·δₙ·(n−2) − 3·δₙ·M₂
//	M₂ += term
//
// M₄ and M₃ are updated before M₂ and M₃ respectively, because
// they use the values from before x was added.
func (m *CentralMomentsKleinKBN) Update(x float64) {
	nOld := m.n
	nNew := nOld + 1
	m.n = nNew
	fn := float64(nNew)
	delta := x - m.m1.Value()
	deltaN := delta / float64(nNew)
	deltaN2 := deltaN * deltaN
	term := delta * deltaN * float64(nOld)
	m2 := m.m2.Value()
	m3 := m.m3.Value()
	m.m1.Update(deltaN)
	m.m4.Update(term*deltaN2*(fn*fn-3*fn+3) + 6*deltaN2*m2 - 4*deltaN*m3)
	m.m3.Update(term*deltaN*float64(nNew-2) - 3*deltaN*m2)
	m.m2.Update(term)
}

// Revert removes a previously added sample x, regardless of insertion order.
// Reverting a value that was never added corrupts the state.
//
// The restored M₁–M₄ are written with [KleinKBNAccumulator.Set], which
// clears their compensation terms. Subsequent updates rebuild the
// compensation from the restored values. Repeated reverts can
// accumulate rounding error, especially for large-offset data.
//
// Inverse formulas (where nₙ = count before revert, nₒ = nₙ − 1):
//
//	M₁_old = (nₙ · M₁_new − x) / nₒ            [mean undo]
//	δ      = x − M₁_old
//	δₙ     = δ / nₙ
//	term   = δ · δₙ · nₒ
//
//	M₂_old = M₂_new − term
//	M₃_old = M₃_new − (term·δₙ·(nₙ−2) − 3·δₙ·M₂_old)
//	M₄_old = M₄_new − (term·δₙ²·(nₙ²−3nₙ+3)
//	                    + 6·δₙ²·M₂_old − 4·δₙ·M₃_old)
//
// Panics if there are no samples.
func (m *CentralMomentsKleinKBN) Revert(x float64) {
	nNew := m.n
	if nNew == 0 {
		panic("Cannot revert from an empty accumulator")
	}
	nOld := nNew - 1
	if nOld == 0 {
		m.Reset()
		return
	}
	fn := float64(nNew)

	m1New := m.m1.Value()
	m2New := m.m2.Value()
	m3New := m.m3.Value()
	m4New := m.m4.Value()

	m1Old := (float64(nNew)*m1New - x) / float64(nOld)
	delta := x - m1Old
	deltaN := delta / float64(nNew)
	deltaN2 := deltaN * deltaN
	term := delta * deltaN * float64(nOld)

	m2Old := m2New - term
	m3Old := m3New - (term*deltaN*float64(nNew-2) - 3*deltaN*m2Old)
	m4Old := m4New - (term*deltaN2*(fn*fn-3*fn+3) + 6*deltaN2*m2Old - 4*deltaN*m3Old)

	m.n = nOld
	m.m1.Set(m1Old)
	m.m2.Set(m2Old)
	m.m3.Set(m3Old)
	m.m4.Set(m4Old)
}

// N returns the number of samples.
func (m *CentralMomentsKleinKBN) N() int { return m.n }

// Mean returns the arithmetic mean (0.0 when empty).
func (m *CentralMomentsKleinKBN) Mean() float64 {
	return m.m1.Value()
}

// Variance returns the variance M₂ / (n - ddof), NaN when n ≤ ddof.
//
// A slightly negative M₂ caused by rounding after Revert
// is clamped to zero.
func (m *CentralMomentsKleinKBN) Variance() float64 {
	d := m.n - m.ddof
	if d <= 0 {
		return math.NaN()
	}
	m2 := m.m2.Value()
	if m2 < 0 {
		m2 = 0
	}
	return m2 / float64(d)
}

// StandardDeviation returns the square root of the variance, NaN when n ≤ ddof.
func (m *CentralMomentsKleinKBN) StandardDeviation() float64 {
	v := m.Variance()
	if math.IsNaN(v) {
		return v
	}
	return math.Sqrt(v)
}

// Skewness returns the skewness g₁ (bias=true) or G₁ (bias=false); see the
// type Notes.
func (m *CentralMomentsKleinKBN) Skewness() float64 {
	n := m.n
	m2 := m.m2.Value()
	if n < 2 || m2 <= 0 {
		return math.NaN()
	}
	g1 := math.Sqrt(float64(n)) * m.m3.Value() / (m2 * math.Sqrt(m2))
	if m.bias {
		return g1
	}
	if n < 3 {
		return math.NaN()
	}
	fn := float64(n)
	return g1 * math.Sqrt(fn*(fn-1)) / (fn - 2)
}

// Kurtosis returns the kurtosis selected by bias and fisher; see the type
// Notes.
func (m *CentralMomentsKleinKBN) Kurtosis() float64 {
	n := m.n
	m2 := m.m2.Value()
	if n < 2 || m2 <= 0 {
		return math.NaN()
	}
	b2 := float64(n) * m.m4.Value() / (m2 * m2)
	if m.bias {
		if m.fisher {
			return b2 - 3.0
		}
		return b2
	}
	if n < 4 {
		return math.NaN()
	}
	fn := float64(n)
	g2 := ((fn*fn-1)*b2 - 3*(fn-1)*(fn-1)) / ((fn - 2) * (fn - 3))
	if m.fisher {
		return g2
	}
	return g2 + 3.0
}
