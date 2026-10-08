package core

import (
	"errors"
	"math"
)

// sqrt2 is √2.
const sqrt2 = 1.4142135623730950488016887242097

// invSqrt2Pi is 1/√(2π).
const invSqrt2Pi = 0.39894228040143267793994605993438

// NormCDF returns the standard normal cumulative distribution function
// (CDF), P(Z ≤ z) for Z ~ N(0, 1):
//
//	Φ(z) = ½ (1 + erf(z / √2))
//
// For a normal distribution N(μ, σ²), standardize first: Φ((x - μ) / σ).
func NormCDF(z float64) float64 {
	return 0.5 * (1.0 + math.Erf(z/sqrt2))
}

// NormPDF returns the standard normal probability density function (PDF),
//
//	φ(z) = exp(-z² / 2) / √(2π)
//
// for Z ~ N(0, 1). For a normal distribution N(μ, σ²), use
// φ((x - μ) / σ) / σ.
func NormPDF(z float64) float64 {
	return invSqrt2Pi * math.Exp(-0.5*z*z)
}

// Acklam's published polynomial coefficients for the inverse standard
// normal CDF.
//
// acklamA, acklamB: central region; acklamC, acklamD: lower and upper tails.
var (
	acklamA = [6]float64{
		-3.969683028665376e+01, 2.209460984245205e+02,
		-2.759285104469687e+02, 1.383577518672690e+02,
		-3.066479806614716e+01, 2.506628277459239e+00,
	}
	acklamB = [5]float64{
		-5.447609879822406e+01, 1.615858368580409e+02,
		-1.556989798598866e+02, 6.680131188771972e+01,
		-1.328068155288572e+01,
	}
	acklamC = [6]float64{
		-7.784894002430293e-03, -3.223964580411365e-01,
		-2.400758277161838e+00, -2.549732539343734e+00,
		4.374664141464968e+00, 2.938163982698783e+00,
	}
	acklamD = [4]float64{
		7.784695709041462e-03, 3.224671290700398e-01,
		2.445134137142996e+00, 3.754408661907416e+00,
	}
)

// errNormPPFDomain is returned by NormPPF for p outside (0, 1).
var errNormPPFDomain = errors.New("p must be between 0 and 1 (exclusive)")

// NormPPF returns the standard normal percent-point function (inverse
// CDF): z such that P(Z ≤ z) = p for Z ~ N(0, 1).
//
// The inverse standard normal CDF has no closed-form expression. This
// implementation uses Peter J. Acklam's piecewise rational approximation,
// which divides the probability domain into lower tail, central region,
// and upper tail. The upper-tail approximation exploits the symmetry
// Φ⁻¹(1 - p) = -Φ⁻¹(p) to reuse the lower-tail coefficients.
//
// The approximation is accurate to roughly 1×10⁻⁹ relative error over
// most of the domain, consistent with Acklam's published results.
//
// Returns an error if p is not strictly between 0 and 1.
func NormPPF(p float64) (float64, error) {
	if p <= 0 || p >= 1 {
		return math.NaN(), errNormPPFDomain
	}

	c, d := &acklamC, &acklamD
	// Variables (not constants) so that 1 - pLow is rounded like Python.
	pLow := 0.02425
	pHigh := 1.0 - pLow
	if p < pLow {
		// Lower tail
		q := math.Sqrt(-2.0 * math.Log(p))
		return (((((c[0]*q+c[1])*q+c[2])*q+c[3])*q+c[4])*q + c[5]) /
			((((d[0]*q+d[1])*q+d[2])*q+d[3])*q + 1.0), nil
	} else if p <= pHigh {
		// Central region
		a, b := &acklamA, &acklamB
		q := p - 0.5
		r := q * q
		return (((((a[0]*r+a[1])*r+a[2])*r+a[3])*r+a[4])*r + a[5]) * q /
			(((((b[0]*r+b[1])*r+b[2])*r+b[3])*r+b[4])*r + 1.0), nil
	}
	// Upper tail
	q := math.Sqrt(-2.0 * math.Log(1.0-p))
	return -(((((c[0]*q+c[1])*q+c[2])*q+c[3])*q+c[4])*q + c[5]) /
		((((d[0]*q+d[1])*q+d[2])*q+d[3])*q + 1.0), nil
}
