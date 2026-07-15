package streamingkbn

import "math"

// CentralMomentsKleinKBN provides streaming mean, variance, skewness, and kurtosis
// via Pébay's central moment update with KBN (Kahan-Babuška-Neumaier) double-compensated
// accumulation.
//
// It maintains running sums of central moments m2, m3, m4 (as KleinKBNAccumulators)
// updated in O(1) per sample. Preferred over RawMomentsKleinKBN for forward-only
// computation (no revert) because it avoids the numerical cancellation
// inherent in converting raw power sums to central moments.
//
// Parameters:
//   - ddof: Delta degrees of freedom for variance.
//     variance = m2 / (n - ddof). ddof=0 gives population, ddof=1 gives sample.
//   - bias: If true, compute population standardized moments (m3/m2^1.5).
//     If false, apply the Fisher-Pearson adjusted (bias-corrected) factor:
//     skewness_bcf = skewness_pop * sqrt(n*(n-1)) / (n-2)
//   - fisher: If true, return excess kurtosis (subtract 3 so Gaussian→0).
//     If false, return raw kurtosis (Gaussian→3).
//     Applied after the bias correction when bias=false.
//
// Formulas:
//
// Skewness (bias=true):
//
//	g1 = sqrt(n) * m3 / m2^1.5
//
// Skewness (bias=false):
//
//	G1 = g1 * sqrt(n*(n-1)) / (n-2)
//
// Kurtosis (bias=true, fisher=true):
//
//	g2 = n * m4 / m2^2 - 3
//
// Kurtosis (bias=true, fisher=false):
//
//	g2 = n * m4 / m2^2
//
// Kurtosis (bias=false, fisher=true):
//
//	G2 = ((n^2-1) * n*m4/m2^2 - 3*(n-1)^2) / ((n-2)*(n-3))
//
// Kurtosis (bias=false, fisher=false):
//
//	G2 = ((n^2-1) * n*m4/m2^2 - 3*(n-1)^2) / ((n-2)*(n-3)) + 3
type CentralMomentsKleinKBN struct {
	n      int
	m1     *KleinKBNAccumulator
	m2     *KleinKBNAccumulator
	m3     *KleinKBNAccumulator
	m4     *KleinKBNAccumulator
	ddof   int
	bias   bool
	fisher bool
}

// NewCentralMomentsKleinKBN creates a new CentralMomentsKleinKBN with the given parameters.
func NewCentralMomentsKleinKBN(ddof int, bias, fisher bool) *CentralMomentsKleinKBN {
	return &CentralMomentsKleinKBN{
		m1:     &KleinKBNAccumulator{},
		m2:     &KleinKBNAccumulator{},
		m3:     &KleinKBNAccumulator{},
		m4:     &KleinKBNAccumulator{},
		ddof:   ddof,
		bias:   bias,
		fisher: fisher,
	}
}

// Reset clears all accumulated state.
func (m *CentralMomentsKleinKBN) Reset() {
	m.n = 0
	m.m1.Reset()
	m.m2.Reset()
	m.m3.Reset()
	m.m4.Reset()
}

// Update adds a new sample x using Pébay's central moment update formulas.
func (m *CentralMomentsKleinKBN) Update(x float64) {
	nOld := m.n
	nNew := nOld + 1
	m.n = nNew
	delta := x - m.m1.Value()
	deltaN := delta / float64(nNew)
	deltaN2 := deltaN * deltaN
	term := delta * deltaN * float64(nOld)

	m.m1.Update(deltaN)
	m.m4.Update(term*deltaN2*(float64(nNew*nNew)-3*float64(nNew)+3) + 6*deltaN2*m.m2.Value() - 4*deltaN*m.m3.Value())
	m.m3.Update(term*deltaN*(float64(nNew)-2) - 3*deltaN*m.m2.Value())
	m.m2.Update(term)
}

// Revert removes the most recently added sample x (LIFO).
// Uses inverse Pébay formulas to restore prior state. The KleinKBNAccumulator.Set()
// method is used for m1–m4, which resets the compensation terms to zero.
// Only the most recent sample can be reverted (LIFO stack, not FIFO queue).
//
// Inverse formulas (where nN = count before revert, nO = nN - 1):
//
//	m1_old = (nN * m1_new - x) / nO
//	δ      = x - m1_old
//	δN     = δ / nN
//	δN2    = δN * δN
//	term   = δ * δN * nO
//	m2_old = m2_new - term
//	m3_old = m3_new - (term*δN*(nN-2) - 3*δN*m2_old)
//	m4_old = m4_new - (term*δN2*(nN^2-3*nN+3) + 6*δN2*m2_old - 4*δN*m3_old)
func (m *CentralMomentsKleinKBN) Revert(x float64) {
	nNew := m.n
	if nNew == 0 {
		panic("cannot revert below 0")
	}
	nOld := nNew - 1
	if nOld == 0 {
		m.n = 0
		m.m1.Reset()
		m.m2.Reset()
		m.m3.Reset()
		m.m4.Reset()
		return
	}

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
	m3Old := m3New - (term*deltaN*(float64(nNew)-2) - 3*deltaN*m2Old)
	m4Old := m4New - (term*deltaN2*(float64(nNew*nNew)-3*float64(nNew)+3) + 6*deltaN2*m2Old - 4*deltaN*m3Old)

	m.n = nOld
	m.m1.Set(m1Old)
	m.m2.Set(m2Old)
	m.m3.Set(m3Old)
	m.m4.Set(m4Old)
}

// Mean returns the current arithmetic mean.
func (m *CentralMomentsKleinKBN) Mean() float64 {
	return m.m1.Value()
}

// Variance returns the current variance. Returns 0 if n <= ddof.
func (m *CentralMomentsKleinKBN) Variance() float64 {
	n := float64(m.n - m.ddof)
	if n <= 0 {
		return math.NaN()
	}
	return m.m2.Value() / n
}

// StandardDeviation returns the current standard deviation. Returns 0 if n <= ddof.
func (m *CentralMomentsKleinKBN) StandardDeviation() float64 {
	n := float64(m.n - m.ddof)
	if n <= 0 {
		return math.NaN()
	}
	return math.Sqrt(m.m2.Value() / n)
}

// Skewness returns the current skewness. Returns 0 if n < 3 or m2 <= 0.
func (m *CentralMomentsKleinKBN) Skewness() float64 {
	n := float64(m.n)
	if m.n < 3 || m.m2.Value() <= 0 {
		return math.NaN()
	}
	g1 := math.Sqrt(n) * m.m3.Value() / math.Pow(m.m2.Value(), 1.5)
	if m.bias {
		return g1
	}
	return g1 * math.Sqrt(n*(n-1)) / (n - 2)
}

// Kurtosis returns the current kurtosis. Returns 0 if n < 4 or m2 <= 0.
func (m *CentralMomentsKleinKBN) Kurtosis() float64 {
	n := float64(m.n)
	if m.n < 4 || m.m2.Value() <= 0 {
		return math.NaN()
	}
	raw := n * m.m4.Value() / (m.m2.Value() * m.m2.Value())
	if !m.bias {
		adj := ((n*n-1)*raw - 3*(n-1)*(n-1)) / ((n - 2) * (n - 3))
		if m.fisher {
			return adj
		}
		return adj + 3.0
	}
	if m.fisher {
		return raw - 3.0
	}
	return raw
}
