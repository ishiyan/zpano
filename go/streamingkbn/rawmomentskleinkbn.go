package streamingkbn

import "math"

// RawMomentsKleinKBN provides streaming mean, variance, skewness, and kurtosis
// via raw power sums (x¹..x⁴) with KBN (Kahan-Babuška-Neumaier) double-compensated
// accumulation.
//
// It accumulates Σx, Σx², Σx³, Σx⁴ using KleinKBNAccumulator for each,
// plus a separate Welford-style variance tracker (also KBN-compensated).
// Raw sums are converted to central moments at query time.
//
// Supports both LIFO revert (undo the most recent update) and FIFO
// rolling window (via the revert/update cycle) because subtracting
// from a linear sum preserves the KBN compensation state.
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
type RawMomentsKleinKBN struct {
	n      int
	x1     *KleinKBNAccumulator
	x2     *KleinKBNAccumulator
	x3     *KleinKBNAccumulator
	x4     *KleinKBNAccumulator
	mean   *KleinKBNAccumulator
	s      *KleinKBNAccumulator
	ddof   int
	bias   bool
	fisher bool
}

// NewRawMomentsKleinKBN creates a new RawMomentsKleinKBN with the given parameters.
func NewRawMomentsKleinKBN(ddof int, bias, fisher bool) *RawMomentsKleinKBN {
	return &RawMomentsKleinKBN{
		x1:     &KleinKBNAccumulator{},
		x2:     &KleinKBNAccumulator{},
		x3:     &KleinKBNAccumulator{},
		x4:     &KleinKBNAccumulator{},
		mean:   &KleinKBNAccumulator{},
		s:      &KleinKBNAccumulator{},
		ddof:   ddof,
		bias:   bias,
		fisher: fisher,
	}
}

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

// Update adds a new sample x to the accumulator.
func (m *RawMomentsKleinKBN) Update(x float64) {
	m.n++
	m.x1.Update(x)
	x2 := x * x
	m.x2.Update(x2)
	x3 := x2 * x
	m.x3.Update(x3)
	x4 := x3 * x
	m.x4.Update(x4)

	n := float64(m.n)
	delta := x - m.mean.Value()
	m.mean.Update(delta / n)
	m.s.Update(delta * (x - m.mean.Value()))
}

// Revert removes a previously added sample x from the accumulator.
func (m *RawMomentsKleinKBN) Revert(x float64) {
	m.n--
	m.x1.Revert(x)
	x2 := x * x
	m.x2.Revert(x2)
	x3 := x2 * x
	m.x3.Revert(x3)
	x4 := x3 * x
	m.x4.Revert(x4)

	delta := x - m.mean.Value()
	n := float64(m.n)
	m.mean.Revert(delta / n)
	m.s.Revert(delta * (x - m.mean.Value()))
}

// Mean returns the current arithmetic mean.
func (m *RawMomentsKleinKBN) Mean() float64 {
	return m.mean.Value()
}

// Variance returns the current variance.
// Returns NaN if n <= ddof.
func (m *RawMomentsKleinKBN) Variance() float64 {
	n := float64(m.n - m.ddof)
	if n <= 0 {
		return math.NaN()
	}
	s := m.s.Value()
	if s < 0 {
		m.s.Reset()
		return math.NaN()
	}
	return s / n
}

// StandardDeviation returns the current standard deviation.
// Returns NaN if n <= ddof.
func (m *RawMomentsKleinKBN) StandardDeviation() float64 {
	n := float64(m.n - m.ddof)
	if n <= 0 {
		return math.NaN()
	}
	return math.Sqrt(m.s.Value() / n)
}

// Skewness returns the current skewness.
// Returns NaN if n < 3.
func (m *RawMomentsKleinKBN) Skewness() float64 {
	N := float64(m.n)
	if m.n < 3 {
		return math.NaN()
	}
	A := m.x1.Value() / N
	B := m.x2.Value()/N - A*A
	if B <= 1e-14 {
		return math.NaN()
	}
	R := math.Sqrt(B)
	C := m.x3.Value()/N - A*A*A - 3*A*B
	g1 := C / (R * R * R)
	if m.bias {
		return g1
	}
	return g1 * math.Sqrt(N*(N-1)) / (N - 2)
}

// Kurtosis returns the current kurtosis.
// Returns NaN if n < 4.
func (m *RawMomentsKleinKBN) Kurtosis() float64 {
	N := float64(m.n)
	if m.n < 4 {
		return math.NaN()
	}
	A := m.x1.Value() / N
	R := A * A
	B := m.x2.Value()/N - R
	if B <= 1e-14 {
		return math.NaN()
	}
	R *= A
	C := m.x3.Value()/N - R - 3*A*B
	R *= A
	D := m.x4.Value()/N - R - 6*B*A*A - 4*C*A
	raw := D / (B * B)
	if !m.bias {
		adj := ((N*N-1)*raw - 3*(N-1)*(N-1)) / ((N - 2) * (N - 3))
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
