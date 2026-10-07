package streamingkbn

import "math"

// LinearRegressionKleinKBN computes streaming ordinary least squares (OLS)
// regression y = a + b·x with Klein KBN (Kahan-Babuška-Neumaier) compensated
// accumulation.
//
// It tracks the means and variances of x and y (as [RawMomentsKleinKBN] with
// ddof=0) and the co-moment S_xy = Σ(x − x̄)(y − ȳ), updated in O(1) per
// sample (n = count before adding the sample):
//
//	S_xy += (x − x̄)·(y − ȳ)·n / (n + 1)
//
// where x̄ and ȳ are the means before adding the sample.
//
// Like [RawMomentsKleinKBN], Revert may remove any previously added
// (x, y) pair, not only the most recent one, so the type works for FIFO
// rolling windows.
//
// Derived quantities, with S_xx = Σ(x − x̄)² and S_yy = Σ(y − ȳ)²:
//
//	slope       b = S_xy / S_xx
//	intercept   a = ȳ − b·x̄
//	correlation r = S_xy / √(S_xx·S_yy)
//	covariance      S_xy / n   (population)
type LinearRegressionKleinKBN struct {
	n        int
	xMoments *RawMomentsKleinKBN
	yMoments *RawMomentsKleinKBN
	sxy      KleinKBNAccumulator
}

// NewLinearRegressionKleinKBN returns a new empty LinearRegressionKleinKBN.
func NewLinearRegressionKleinKBN() *LinearRegressionKleinKBN {
	return &LinearRegressionKleinKBN{
		xMoments: NewRawMomentsKleinKBN(0, true, true),
		yMoments: NewRawMomentsKleinKBN(0, true, true),
	}
}

// Reset clears all accumulated state.
func (r *LinearRegressionKleinKBN) Reset() {
	r.n = 0
	r.xMoments.Reset()
	r.yMoments.Reset()
	r.sxy.Reset()
}

// Update adds a sample (x, y).
func (r *LinearRegressionKleinKBN) Update(x, y float64) {
	nOld := r.n
	r.n++
	term := (r.xMoments.Mean() - x) * (r.yMoments.Mean() - y) * float64(nOld) / float64(nOld+1)
	r.sxy.Update(term)
	r.xMoments.Update(x)
	r.yMoments.Update(y)
}

// Revert removes a previously added sample (x, y), not necessarily the most
// recent one.
//
// Panics if there are no samples.
func (r *LinearRegressionKleinKBN) Revert(x, y float64) {
	if r.n == 0 {
		panic("Cannot revert from an empty regression")
	}
	if r.n == 1 {
		r.Reset()
		return
	}
	r.xMoments.Revert(x)
	r.yMoments.Revert(y)
	// The means are now those without (x, y), as in Update.
	n := r.n - 1
	term := (r.xMoments.Mean() - x) * (r.yMoments.Mean() - y) * float64(n) / float64(n+1)
	r.sxy.Revert(term)
	r.n = n
}

// N returns the number of samples.
func (r *LinearRegressionKleinKBN) N() int { return r.n }

// MeanX returns the mean of x (0.0 when empty).
func (r *LinearRegressionKleinKBN) MeanX() float64 { return r.xMoments.Mean() }

// MeanY returns the mean of y (0.0 when empty).
func (r *LinearRegressionKleinKBN) MeanY() float64 { return r.yMoments.Mean() }

// VarianceX returns the population variance of x, S_xx / n (NaN when empty).
func (r *LinearRegressionKleinKBN) VarianceX() float64 { return r.xMoments.Variance() }

// VarianceY returns the population variance of y, S_yy / n (NaN when empty).
func (r *LinearRegressionKleinKBN) VarianceY() float64 { return r.yMoments.Variance() }

// CoMoment returns the co-moment S_xy = Σ(x − x̄)(y − ȳ) (0.0 when empty).
func (r *LinearRegressionKleinKBN) CoMoment() float64 { return r.sxy.Value() }

// Covariance returns the population covariance S_xy / n (NaN when empty).
func (r *LinearRegressionKleinKBN) Covariance() float64 {
	n := r.n
	if n < 1 {
		return math.NaN()
	}
	return r.sxy.Value() / float64(n)
}

// Slope returns the OLS slope b = S_xy / S_xx.
//
// NaN when n < 2 or all x are equal (S_xx = 0).
func (r *LinearRegressionKleinKBN) Slope() float64 {
	n := r.n
	if n < 2 {
		return math.NaN()
	}
	sxx := r.xMoments.Variance() * float64(n)
	if sxx != 0 {
		return r.sxy.Value() / sxx
	}
	return math.NaN()
}

// Intercept returns the OLS intercept a = ȳ − b·x̄ (NaN when the slope is NaN).
func (r *LinearRegressionKleinKBN) Intercept() float64 {
	return r.yMoments.Mean() - r.Slope()*r.xMoments.Mean()
}

// Correlation returns the Pearson correlation coefficient
// r = S_xy / √(S_xx·S_yy), clamped to [−1, 1] to absorb rounding.
//
// NaN when n < 2 or either x or y is constant.
func (r *LinearRegressionKleinKBN) Correlation() float64 {
	n := r.n
	if n < 2 {
		return math.NaN()
	}
	t := r.xMoments.StandardDeviation() * r.yMoments.StandardDeviation()
	if t == 0 {
		return math.NaN()
	}
	c := r.sxy.Value() / (t * float64(n))
	if c > 1.0 {
		return 1.0
	}
	if c < -1.0 {
		return -1.0
	}
	return c
}
