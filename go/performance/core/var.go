package core

import (
	"math"

	"zpano/streamingkbn"
)

// VarHistorical returns the historical value at risk: the negated
// (1 - confidence) percentile of the excess returns r - riskFreeRate
// (the conventional defaults are riskFreeRate = 0, confidence = 0.95).
//
// The returns slice is not modified. Returns NaN when returns is nil or
// empty, or when confidence is outside [0, 1].
func VarHistorical(returns []float64, riskFreeRate, confidence float64) float64 {
	w := returns
	if len(w) < 1 {
		return math.NaN()
	}
	q := 1 - confidence
	if riskFreeRate != 0 {
		excess := make([]float64, len(w))
		for i, r := range w {
			excess[i] = r - riskFreeRate
		}
		w = excess
	}
	p, err := Percentile(w, q)
	if err != nil {
		return math.NaN()
	}
	return -p
}

// VarGaussian returns the Gaussian (parametric) value at risk,
// -(μ + z·σ) with z = Φ⁻¹(1 - confidence) and σ the population (ddof=0)
// standard deviation (the conventional default is confidence = 0.95).
//
// Returns NaN when the standard deviation is unavailable or confidence is
// not strictly between 0 and 1.
func VarGaussian(returnsKBN *streamingkbn.RawMomentsKleinKBN, confidence float64) float64 {
	mean := returnsKBN.Mean()
	std := returnsKBN.StandardDeviationDdof0()
	if math.IsNaN(std) {
		return math.NaN()
	}
	z, err := NormPPF(1 - confidence)
	if err != nil {
		return math.NaN()
	}
	return -(mean + z*std)
}

// VarCornishFisher returns the modified (Cornish-Fisher) value at risk,
// -(μ + z_cf·σ), where z_cf is the Cornish-Fisher expansion of
// z = Φ⁻¹(1 - confidence) using the biased skewness and excess kurtosis
// (the conventional default is confidence = 0.95).
//
// Skewness and kurtosis are unavailable for very small samples; the
// Gaussian VaR is returned then. Returns NaN when the standard deviation
// is unavailable or confidence is not strictly between 0 and 1.
func VarCornishFisher(returnsKBN *streamingkbn.RawMomentsKleinKBN, confidence float64) float64 {
	mean := returnsKBN.Mean()
	std := returnsKBN.StandardDeviationDdof0()
	if math.IsNaN(std) {
		return math.NaN()
	}
	// Cornish-Fisher expansion for z-score adjustment.
	z, err := NormPPF(1 - confidence)
	if err != nil {
		return math.NaN()
	}
	skew := returnsKBN.SkewnessMoment()     // bias=true
	kurtosis := returnsKBN.KurtosisExcess() // bias=true, fisher=true
	// Skewness and kurtosis are unavailable for very small samples.
	// Fall back to Gaussian VaR.
	if math.IsNaN(skew) || math.IsNaN(kurtosis) {
		return -(mean + z*std)
	}
	// Cornish-Fisher expansion.
	z2 := z * z
	z3 := z2 * z
	z = z +
		(z2-1)*skew/6 +
		(z3-3*z)*kurtosis/24 -
		(2*z3-5*z)*skew*skew/36
	return -(mean + z*std)
}
