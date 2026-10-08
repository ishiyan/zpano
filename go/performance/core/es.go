package core

import (
	"math"

	"zpano/streamingkbn"
)

// EsHistorical returns the historical expected shortfall: the negated mean
// of the excess returns r - riskFreeRate at or below -VaR, where VaR is
// [VarHistorical] (the conventional defaults are riskFreeRate = 0,
// confidence = 0.95).
//
// The returns slice is not modified. Returns NaN when returns is nil or
// empty, when VaR is NaN, or when the tail is empty.
func EsHistorical(returns []float64, riskFreeRate, confidence float64) float64 {
	if len(returns) == 0 {
		return math.NaN()
	}
	v := VarHistorical(returns, riskFreeRate, confidence)
	if math.IsNaN(v) {
		return math.NaN()
	}

	sumTail := 0.0
	count := 0
	for _, r := range returns {
		excess := r - riskFreeRate
		if excess <= -v {
			sumTail += excess
			count++
		}
	}

	if count != 0 {
		return -sumTail / float64(count)
	}
	return math.NaN()
}

// EsGaussian returns the Gaussian (parametric) expected shortfall,
// -μ + φ(z)·σ / (1 - confidence) with z = Φ⁻¹(confidence) and σ the
// population (ddof=0) standard deviation (the conventional default is
// confidence = 0.95).
//
// Returns NaN when the standard deviation is unavailable or confidence is
// not strictly between 0 and 1.
func EsGaussian(returnsKBN *streamingkbn.RawMomentsKleinKBN, confidence float64) float64 {
	mean := returnsKBN.Mean()
	std := returnsKBN.StandardDeviationDdof0()
	if math.IsNaN(std) {
		return math.NaN()
	}
	z, err := NormPPF(confidence)
	if err != nil {
		return math.NaN()
	}
	phiZ := NormPDF(z)
	return -mean + phiZ*std/(1-confidence)
}

// EsCornishFisher returns the modified (Cornish-Fisher) expected shortfall
// using the biased skewness and excess kurtosis (the conventional default
// is confidence = 0.95).
//
// Skewness and kurtosis are unavailable for very small samples; the
// Gaussian ES ([EsGaussian]) is returned then. Returns NaN when confidence
// is not strictly between 0 and 1.
func EsCornishFisher(returnsKBN *streamingkbn.RawMomentsKleinKBN, confidence float64) float64 {
	alpha := 1.0 - confidence
	z, err := NormPPF(alpha)
	if err != nil {
		return math.NaN()
	}
	mean := returnsKBN.Mean()
	sigma := returnsKBN.StandardDeviationDdof0()
	skew := returnsKBN.SkewnessMoment()     // bias=true
	kurtosis := returnsKBN.KurtosisExcess() // bias=true, fisher=true
	// Skewness and kurtosis are unavailable for very small samples.
	// Fall back to Gaussian ES.
	if math.IsNaN(skew) || math.IsNaN(kurtosis) {
		return EsGaussian(returnsKBN, confidence)
	}
	z2 := z * z
	z3 := z2 * z
	h := z +
		(z2-1)*skew/6 +
		(z3-3*z)*kurtosis/24 -
		(2*z3-5*z)*skew*skew/36
	h2 := h * h
	h4 := h2 * h2
	mes := NormPDF(h) * (1 +
		h2*h*skew/6 +
		(h4*h2-9*h4+9*h2+3)*skew*skew/72 +
		(h4-2*h2-1)*kurtosis/24)
	// Python min(a, b): b if b < a else a.
	m := -mes / alpha
	if h < m {
		m = h
	}
	return -mean - sigma*m
}
