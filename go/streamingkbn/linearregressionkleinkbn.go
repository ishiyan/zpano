package streamingkbn

import "math"

// LinearRegressionKleinKBN provides streaming simple linear regression
// (y = slope * x + intercept) with KBN-compensated accumulation.
//
// It internally uses two RawMomentsKleinKBN (ddof=0) for x and y moments,
// and a KleinKBNAccumulator for the cross-product S_xy.
//
// Supports both LIFO revert and FIFO rolling window via the revert/update cycle.
type LinearRegressionKleinKBN struct {
	n        int
	xMoments *RawMomentsKleinKBN
	yMoments *RawMomentsKleinKBN
	sXY      *KleinKBNAccumulator
}

// NewLinearRegressionKleinKBN creates a new LinearRegressionKleinKBN.
func NewLinearRegressionKleinKBN() *LinearRegressionKleinKBN {
	return &LinearRegressionKleinKBN{
		xMoments: NewRawMomentsKleinKBN(0, true, true),
		yMoments: NewRawMomentsKleinKBN(0, true, true),
		sXY:      &KleinKBNAccumulator{},
	}
}

// Reset clears all accumulated state.
func (r *LinearRegressionKleinKBN) Reset() {
	r.n = 0
	r.xMoments.Reset()
	r.yMoments.Reset()
	r.sXY.Reset()
}

// Update adds a new (x, y) observation.
func (r *LinearRegressionKleinKBN) Update(x, y float64) {
	nOld := r.n
	r.n++
	term := (r.xMoments.Mean() - x) * (r.yMoments.Mean() - y) * float64(nOld) / float64(nOld+1)
	r.sXY.Update(term)
	r.xMoments.Update(x)
	r.yMoments.Update(y)
}

// Revert removes a previously added (x, y) observation.
func (r *LinearRegressionKleinKBN) Revert(x, y float64) {
	if r.n == 0 {
		return
	}
	if r.n == 1 {
		r.Reset()
		return
	}
	r.xMoments.Revert(x)
	r.yMoments.Revert(y)
	n := r.n - 1
	term := (r.xMoments.Mean() - x) * (r.yMoments.Mean() - y) * float64(n) / float64(n+1)
	r.sXY.Revert(term)
	r.n = n
}

// Slope returns the current slope coefficient. Returns NaN if n < 2 or S_xx == 0.
func (r *LinearRegressionKleinKBN) Slope() float64 {
	if r.n < 2 {
		return math.NaN()
	}
	Sxx := r.xMoments.Variance() * float64(r.n)
	if Sxx == 0 {
		return math.NaN()
	}
	return r.sXY.Value() / Sxx
}

// Intercept returns the current intercept coefficient.
func (r *LinearRegressionKleinKBN) Intercept() float64 {
	return r.yMoments.Mean() - r.Slope()*r.xMoments.Mean()
}

// Correlation returns the current Pearson correlation coefficient. Returns NaN if n < 2 or
// either standard deviation is zero.
func (r *LinearRegressionKleinKBN) Correlation() float64 {
	if r.n < 2 {
		return math.NaN()
	}
	t := r.xMoments.StandardDeviation() * r.yMoments.StandardDeviation()
	if t == 0 {
		return math.NaN()
	}
	return r.sXY.Value() / (t * float64(r.n))
}
