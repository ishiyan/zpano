import math

#_SQRT2 = math.sqrt(2.0)
_SQRT2 = 1.4142135623730950488016887242097

def norm_cdf(z: float) -> float:
    """
    Standard normal cumulative distribution function (CDF).

    Computes P(Z ≤ z) for Z ~ N(0, 1) using ``math.erf`` from the
    Python standard library:

        Φ(z) = ½ (1 + erf(z / √2))

    The underlying ``erf`` implementation is written in C and is
    typically accurate to near machine precision.

    For a normal distribution N(μ, σ²), standardize first:

        Φ((x - μ) / σ)
    """
    return 0.5 * (1.0 + math.erf(z / _SQRT2))

#_INV_SQRT_2PI = 1.0 / math.sqrt(2.0 * math.pi)
_INV_SQRT_2PI = 0.39894228040143267793994605993438

def norm_pdf(z: float) -> float:
    """
    Standard normal probability density function (PDF).

    Computes

        φ(z) = exp(-z² / 2) / √(2π)

    for Z ~ N(0, 1).

    For a normal distribution N(μ, σ²),

        φ((x - μ) / σ) / σ
    """
    return _INV_SQRT_2PI * math.exp(-0.5 * z * z)

# The inverse standard normal CDF has no closed-form expression.
#
# This implementation uses Peter J. Acklam's piecewise rational
# approximation, which divides the probability domain into lower
# tail, central region, and upper tail.
#
# Separate rational functions are used in each region because no
# single polynomial provides uniform accuracy over (0, 1).
#
# The upper-tail approximation exploits the symmetry
#
#     Φ⁻¹(1 - p) = -Φ⁻¹(p)
#
# to reuse the lower-tail coefficients.

# Acklam's published polynomial coefficients.
#
# _a, _b : central region
# _c, _d : lower and upper tails

_a = (
    -3.969683028665376e+01,  2.209460984245205e+02,
    -2.759285104469687e+02,  1.383577518672690e+02,
    -3.066479806614716e+01,  2.506628277459239e+00
)

_b = (
    -5.447609879822406e+01,  1.615858368580409e+02,
    -1.556989798598866e+02,  6.680131188771972e+01,
    -1.328068155288572e+01
)

_c = (
    -7.784894002430293e-03, -3.223964580411365e-01,
    -2.400758277161838e+00, -2.549732539343734e+00,
    4.374664141464968e+00,  2.938163982698783e+00
)

_d = (
    7.784695709041462e-03,  3.224671290700398e-01,
    2.445134137142996e+00,  3.754408661907416e+00
)

def norm_ppf(p: float) -> float:
    """
    Standard normal percent-point function (inverse CDF).

    Computes z such that

        P(Z ≤ z) = p

    for Z ~ N(0, 1).

    Uses Peter J. Acklam's piecewise rational approximation.

    The approximation is accurate to roughly 1×10⁻⁹ relative error
    over most of the domain, consistent with Acklam's published results.
    """
    if p <= 0 or p >= 1:
        raise ValueError("p must be between 0 and 1 (exclusive)")

    p_low = 0.02425
    p_high = 1.0 - p_low
    if p < p_low:
        # Lower tail
        q = math.sqrt(-2.0 * math.log(p))
        return (((((_c[0]*q + _c[1])*q + _c[2])*q + _c[3])*q + _c[4])*q + _c[5]) / \
               ((((_d[0]*q + _d[1])*q + _d[2])*q + _d[3])*q + 1.0)
    elif p <= p_high:
        # Central region
        q = p - 0.5
        r = q * q
        return (((((_a[0]*r + _a[1])*r + _a[2])*r + _a[3])*r + _a[4])*r + _a[5]) * q / \
               (((((_b[0]*r + _b[1])*r + _b[2])*r + _b[3])*r + _b[4])*r + 1.0)
    else:
        # Upper tail
        q = math.sqrt(-2.0 * math.log(1.0 - p))
        return -(((((_c[0]*q + _c[1])*q + _c[2])*q + _c[3])*q + _c[4])*q + _c[5]) / \
                ((((_d[0]*q + _d[1])*q + _d[2])*q + _d[3])*q + 1.0)
