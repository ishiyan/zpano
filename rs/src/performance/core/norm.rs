//! Standard normal distribution: CDF, PDF and inverse CDF (PPF).

// Python: _SQRT2 = math.sqrt(2.0)
#[allow(clippy::excessive_precision, clippy::approx_constant)]
const SQRT2: f64 = 1.4142135623730950488016887242097;

/// Standard normal cumulative distribution function (CDF).
///
/// Computes P(Z ≤ z) for Z ~ N(0, 1):
///
/// ```text
/// Φ(z) = ½ (1 + erf(z / √2))
/// ```
///
/// `erf` is a port of the fdlibm / glibc double-precision `erf` used by
/// Python's `math.erf`, accurate to near machine precision.
///
/// For a normal distribution N(μ, σ²), standardize first: Φ((x - μ) / σ).
pub fn norm_cdf(z: f64) -> f64 {
    0.5 * (1.0 + erf(z / SQRT2))
}

// Python: _INV_SQRT_2PI = 1.0 / math.sqrt(2.0 * math.pi)
#[allow(clippy::excessive_precision)]
const INV_SQRT_2PI: f64 = 0.39894228040143267793994605993438;

/// Standard normal probability density function (PDF).
///
/// ```text
/// φ(z) = exp(-z² / 2) / √(2π)
/// ```
///
/// For a normal distribution N(μ, σ²): φ((x - μ) / σ) / σ.
pub fn norm_pdf(z: f64) -> f64 {
    INV_SQRT_2PI * (-0.5 * z * z).exp()
}

// The inverse standard normal CDF has no closed-form expression.
//
// This implementation uses Peter J. Acklam's piecewise rational
// approximation, which divides the probability domain into lower
// tail, central region, and upper tail. The upper-tail approximation
// exploits the symmetry Φ⁻¹(1 - p) = -Φ⁻¹(p) to reuse the lower-tail
// coefficients.
//
// A, B: central region; C, D: lower and upper tails.

#[allow(clippy::excessive_precision)]
const A: [f64; 6] = [
    -3.969683028665376e+01,
    2.209460984245205e+02,
    -2.759285104469687e+02,
    1.383577518672690e+02,
    -3.066479806614716e+01,
    2.506628277459239e+00,
];

#[allow(clippy::excessive_precision)]
const B: [f64; 5] = [
    -5.447609879822406e+01,
    1.615858368580409e+02,
    -1.556989798598866e+02,
    6.680131188771972e+01,
    -1.328068155288572e+01,
];

#[allow(clippy::excessive_precision)]
const C: [f64; 6] = [
    -7.784894002430293e-03,
    -3.223964580411365e-01,
    -2.400758277161838e+00,
    -2.549732539343734e+00,
    4.374664141464968e+00,
    2.938163982698783e+00,
];

#[allow(clippy::excessive_precision)]
const D: [f64; 4] = [
    7.784695709041462e-03,
    3.224671290700398e-01,
    2.445134137142996e+00,
    3.754408661907416e+00,
];

/// Standard normal percent-point function (inverse CDF).
///
/// Computes z such that P(Z ≤ z) = p for Z ~ N(0, 1), using Peter J.
/// Acklam's piecewise rational approximation (relative error roughly
/// 1×10⁻⁹ over most of the domain).
///
/// # Errors
///
/// Returns `"p must be between 0 and 1 (exclusive)"` when p ≤ 0 or p ≥ 1.
/// A NaN p is not rejected and yields NaN.
pub fn norm_ppf(p: f64) -> Result<f64, String> {
    if p <= 0.0 || p >= 1.0 {
        return Err("p must be between 0 and 1 (exclusive)".to_string());
    }

    let p_low = 0.02425;
    let p_high = 1.0 - p_low;
    if p < p_low {
        // Lower tail
        let q = (-2.0 * p.ln()).sqrt();
        Ok((((((C[0] * q + C[1]) * q + C[2]) * q + C[3]) * q + C[4]) * q + C[5])
            / ((((D[0] * q + D[1]) * q + D[2]) * q + D[3]) * q + 1.0))
    } else if p <= p_high {
        // Central region
        let q = p - 0.5;
        let r = q * q;
        Ok((((((A[0] * r + A[1]) * r + A[2]) * r + A[3]) * r + A[4]) * r + A[5]) * q
            / (((((B[0] * r + B[1]) * r + B[2]) * r + B[3]) * r + B[4]) * r + 1.0))
    } else {
        // Upper tail
        let q = (-2.0 * (1.0 - p).ln()).sqrt();
        Ok(-(((((C[0] * q + C[1]) * q + C[2]) * q + C[3]) * q + C[4]) * q + C[5])
            / ((((D[0] * q + D[1]) * q + D[2]) * q + D[3]) * q + 1.0))
    }
}

// ---------------------------------------------------------------------------
// erf: port of fdlibm s_erf.c (as used by glibc and hence CPython math.erf).
// ---------------------------------------------------------------------------

#[allow(clippy::excessive_precision)]
mod erf_coefficients {
    pub const ERX: f64 = 8.45062911510467529297e-01;
    pub const EFX: f64 = 1.28379167095512586316e-01;
    pub const EFX8: f64 = 1.02703333676410069053e+00;
    pub const PP0: f64 = 1.28379167095512558561e-01;
    pub const PP1: f64 = -3.25042107247001499370e-01;
    pub const PP2: f64 = -2.84817495755985104766e-02;
    pub const PP3: f64 = -5.77027029648944159157e-03;
    pub const PP4: f64 = -2.37630166566501626084e-05;
    pub const QQ1: f64 = 3.97917223959155352819e-01;
    pub const QQ2: f64 = 6.50222499887672944485e-02;
    pub const QQ3: f64 = 5.08130628187576562776e-03;
    pub const QQ4: f64 = 1.32494738004321644526e-04;
    pub const QQ5: f64 = -3.96022827877536812320e-06;
    pub const PA0: f64 = -2.36211856075265944077e-03;
    pub const PA1: f64 = 4.14856118683748331666e-01;
    pub const PA2: f64 = -3.72207876035701323847e-01;
    pub const PA3: f64 = 3.18346619901161753674e-01;
    pub const PA4: f64 = -1.10894694282396677476e-01;
    pub const PA5: f64 = 3.54783043256182359371e-02;
    pub const PA6: f64 = -2.16637559486879084300e-03;
    pub const QA1: f64 = 1.06420880400844228286e-01;
    pub const QA2: f64 = 5.40397917702171048937e-01;
    pub const QA3: f64 = 7.18286544141962662868e-02;
    pub const QA4: f64 = 1.26171219808761642112e-01;
    pub const QA5: f64 = 1.36370839120290507362e-02;
    pub const QA6: f64 = 1.19844998467991074170e-02;
    pub const RA0: f64 = -9.86494403484714822705e-03;
    pub const RA1: f64 = -6.93858572707181764372e-01;
    pub const RA2: f64 = -1.05586262253232909814e+01;
    pub const RA3: f64 = -6.23753324503260060396e+01;
    pub const RA4: f64 = -1.62396669462573470355e+02;
    pub const RA5: f64 = -1.84605092906711035994e+02;
    pub const RA6: f64 = -8.12874355063065934246e+01;
    pub const RA7: f64 = -9.81432934416914548592e+00;
    pub const SA1: f64 = 1.96512716674392571292e+01;
    pub const SA2: f64 = 1.37657754143519042600e+02;
    pub const SA3: f64 = 4.34565877475229228821e+02;
    pub const SA4: f64 = 6.45387271733267880336e+02;
    pub const SA5: f64 = 4.29008140027567833386e+02;
    pub const SA6: f64 = 1.08635005541779435134e+02;
    pub const SA7: f64 = 6.57024977031928170135e+00;
    pub const SA8: f64 = -6.04244152148580987438e-02;
    pub const RB0: f64 = -9.86494292470009928597e-03;
    pub const RB1: f64 = -7.99283237680523006574e-01;
    pub const RB2: f64 = -1.77579549177547519889e+01;
    pub const RB3: f64 = -1.60636384855821916062e+02;
    pub const RB4: f64 = -6.37566443368389627722e+02;
    pub const RB5: f64 = -1.02509513161107724954e+03;
    pub const RB6: f64 = -4.83519191608651397019e+02;
    pub const SB1: f64 = 3.03380607434824582924e+01;
    pub const SB2: f64 = 3.25792512996573918826e+02;
    pub const SB3: f64 = 1.53672958608443695994e+03;
    pub const SB4: f64 = 3.19985821950859553908e+03;
    pub const SB5: f64 = 2.55305040643316442583e+03;
    pub const SB6: f64 = 4.74528541206955367215e+02;
    pub const SB7: f64 = -2.24409524465858183362e+01;
}

/// The error function erf(x), a port of fdlibm `s_erf.c` (the
/// implementation behind glibc's and hence CPython's `math.erf`).
///
/// The Rust standard library has no stable `erf`.
pub(crate) fn erf(x: f64) -> f64 {
    use erf_coefficients::*;
    const TINY: f64 = 1e-300;

    let hx = (x.to_bits() >> 32) as u32 as i32;
    let ix = hx & 0x7fff_ffff;
    if ix >= 0x7ff0_0000 {
        // erf(nan) = nan, erf(+-inf) = +-1
        let i = (((hx as u32) >> 31) << 1) as i32;
        return (1 - i) as f64 + 1.0 / x;
    }

    if ix < 0x3feb_0000 {
        // |x| < 0.84375
        if ix < 0x3e30_0000 {
            // |x| < 2**-28
            if ix < 0x0080_0000 {
                // avoid underflow
                return 0.125 * (8.0 * x + EFX8 * x);
            }
            return x + EFX * x;
        }
        let z = x * x;
        let r = PP0 + z * (PP1 + z * (PP2 + z * (PP3 + z * PP4)));
        let s = 1.0 + z * (QQ1 + z * (QQ2 + z * (QQ3 + z * (QQ4 + z * QQ5))));
        let y = r / s;
        return x + x * y;
    }

    if ix < 0x3ff4_0000 {
        // 0.84375 <= |x| < 1.25
        let s = x.abs() - 1.0;
        let p = PA0 + s * (PA1 + s * (PA2 + s * (PA3 + s * (PA4 + s * (PA5 + s * PA6)))));
        let q = 1.0 + s * (QA1 + s * (QA2 + s * (QA3 + s * (QA4 + s * (QA5 + s * QA6)))));
        return if hx >= 0 { ERX + p / q } else { -ERX - p / q };
    }

    if ix >= 0x4018_0000 {
        // |x| >= 6
        return if hx >= 0 { 1.0 - TINY } else { TINY - 1.0 };
    }

    let xa = x.abs();
    let s = 1.0 / (xa * xa);
    let (r, ss) = if ix < 0x4006_DB6E {
        // |x| < 1/0.35
        (
            RA0 + s * (RA1 + s * (RA2 + s * (RA3 + s * (RA4 + s * (RA5 + s * (RA6 + s * RA7)))))),
            1.0 + s
                * (SA1
                    + s * (SA2 + s * (SA3 + s * (SA4 + s * (SA5 + s * (SA6 + s * (SA7 + s * SA8))))))),
        )
    } else {
        // |x| >= 1/0.35
        (
            RB0 + s * (RB1 + s * (RB2 + s * (RB3 + s * (RB4 + s * (RB5 + s * RB6))))),
            1.0 + s * (SB1 + s * (SB2 + s * (SB3 + s * (SB4 + s * (SB5 + s * (SB6 + s * SB7)))))),
        )
    };
    // z = xa with the low 32 bits cleared.
    let z = f64::from_bits(xa.to_bits() & 0xffff_ffff_0000_0000);
    let r = (-z * z - 0.5625).exp() * ((z - xa) * (z + xa) + r / ss).exp();
    if hx >= 0 { 1.0 - r / xa } else { r / xa - 1.0 }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::performance::core::test_support::almost_equal;

    fn assert_delta(actual: f64, expected: f64, delta: f64, msg: &str) {
        assert!(
            actual == expected || (actual - expected).abs() <= delta,
            "{msg}: expected {expected}, got {actual} (delta {delta})"
        );
    }

    #[test]
    fn test_erf_matches_python_math_erf() {
        // Values from CPython math.erf.
        let cases: [(f64, f64); 19] = [
            (0.0, 0.0),
            (1e-30, 1.1283791670955127e-30),
            (0.1, 0.1124629160182849),
            (0.3, 0.3286267594591274),
            (0.5, 0.5204998778130465),
            (-0.5, -0.5204998778130465),
            (0.9, 0.7969082124228322),
            (1.0, 0.8427007929497149),
            (-1.0, -0.8427007929497149),
            (1.2, 0.9103139782296353),
            (1.3, 0.9340079449406524),
            (2.0, 0.9953222650189527),
            (-2.5, -0.999593047982555),
            (2.9, 0.9999589021219005),
            (3.0, 0.9999779095030014),
            (4.0, 0.9999999845827421),
            (5.0, 0.9999999999984626),
            (5.9, 0.9999999999999999),
            (7.0, 1.0),
        ];
        for (x, e) in cases {
            assert_eq!(erf(x), e, "erf({x})");
        }
        assert!(erf(f64::NAN).is_nan());
        assert_eq!(erf(f64::INFINITY), 1.0);
        assert_eq!(erf(f64::NEG_INFINITY), -1.0);
    }

    #[test]
    fn test_pdf_scipy_compliance() {
        // These values cover symmetry, peak, and tails
        let inputs = [-8.0, -5.0, -3.0, -2.0, -1.0, 0.0, 1.0, 2.0, 3.0, 5.0, 8.0];
        let outputs = [
            5.052271083536893e-15,
            1.4867195147342979e-06,
            0.0044318484119380075,
            0.053990966513188056,
            0.24197072451914337,
            0.3989422804014327,
            0.24197072451914337,
            0.053990966513188056,
            0.0044318484119380075,
            1.4867195147342979e-06,
            5.052271083536893e-15,
        ];
        for (i, (&x, &expected)) in inputs.iter().zip(outputs.iter()).enumerate() {
            let actual = norm_pdf(x);
            assert!(almost_equal(actual, expected, 15), "step {i}: expected {expected}, got {actual}");

            // Symmetry: norm_pdf(x) == norm_pdf(-x)
            let expected = norm_pdf(x);
            let actual = norm_pdf(-x);
            assert!(
                almost_equal(actual, expected, 15),
                "step {i} symmetry: expected {expected}, got {actual}"
            );
        }
    }

    #[test]
    fn test_cdf_scipy_compliance() {
        let inputs = [-8.0, -5.0, -3.0, -2.0, -1.0, 0.0, 1.0, 2.0, 3.0, 5.0, 8.0];
        let outputs = [
            6.22096057427174e-16,
            2.866515718791933e-07,
            0.0013498980316300933,
            0.022750131948179195,
            0.15865525393145707,
            0.5,
            0.8413447460685429,
            0.9772498680518208,
            0.9986501019683699,
            0.9999997133484281,
            0.9999999999999993,
        ];
        for (i, (&x, &expected)) in inputs.iter().zip(outputs.iter()).enumerate() {
            let actual = norm_cdf(x);
            assert!(almost_equal(actual, expected, 15), "step {i}: expected {expected}, got {actual}");

            // Symmetry: norm_cdf(x) + norm_cdf(-x) = 1.0
            let actual = norm_cdf(x) + norm_cdf(-x);
            assert!(almost_equal(actual, 1.0, 15), "step {i} symmetry: expected 1.0, got {actual}");
        }
    }

    #[test]
    fn test_ppf_scipy_compliance() {
        let p_values = [
            // Center
            0.5, 0.75, 0.9, //
            // Branch boundaries (values around the transition)
            0.02425, 0.025, 0.975, 0.97575, //
            // Common quantiles
            0.001, 0.005, 0.01, 0.025, 0.05, 0.10, 0.90, 0.95, 0.975, 0.99, 0.995, 0.999,
            // Extreme tails (exercise numerical stability)
            1e-12, 1e-10, 1e-8, 0.99999999, 0.9999999999, 0.999999999999,
        ];
        // Hardcoded 1-p_values
        let p_symmetric_values = [
            0.5, 0.25, 0.1, //
            0.97575, 0.975, 0.025, 0.02425, //
            0.999, 0.995, 0.99, 0.975, 0.95, 0.90, 0.10, 0.05, 0.025, 0.01, 0.005, 0.001,
            0.999999999999, 0.9999999999, 0.99999999, 1e-8, 1e-10, 1e-12,
        ];
        let z_values = [
            0.0,
            0.6744897501960817,
            1.2815515655446004,
            -1.972961051311885,
            -1.9599639845400545,
            1.959963984540054,
            1.972961051311885,
            -3.090232306167813,
            -2.575829303548901,
            -2.3263478740408408,
            -1.9599639845400545,
            -1.6448536269514729,
            -1.2815515655446004,
            1.2815515655446004,
            1.6448536269514722,
            1.959963984540054,
            2.3263478740408408,
            2.5758293035489004,
            3.090232306167,
            -7.034483825301131,
            -6.361340902404056,
            -5.612001244174789,
            5.612001243305505,
            6.361340889697422,
            7.0344869100478356,
        ];

        for i in 0..p_values.len() {
            let (p, p_symmetric, expected) = (p_values[i], p_symmetric_values[i], z_values[i]);
            let actual = norm_ppf(p).unwrap();
            // Extreme tails have much lower accuracy.
            let delta = if i < 19 { 3e-9 } else { 8e-9 };
            assert_delta(actual, expected, delta, &format!("step {i}"));

            // p ≈ Φ(Φ⁻¹(p))
            let actual = norm_cdf(norm_ppf(p).unwrap());
            assert_delta(actual, p, delta, &format!("step {i} roundtrip"));

            // norm_ppf(p) == - norm_ppf(1.0 - p)
            let expected = norm_ppf(p).unwrap();
            let actual = -norm_ppf(p_symmetric).unwrap();
            let delta = if i < 19 { 1e-13 } else { 4e-6 };
            assert_delta(actual, expected, delta, &format!("step {i} symmetry"));
        }
    }

    #[test]
    fn test_ppf_out_of_range() {
        for p in [0.0, 1.0, -0.5, 1.5] {
            assert_eq!(norm_ppf(p), Err("p must be between 0 and 1 (exclusive)".to_string()));
        }
        assert!(norm_ppf(f64::NAN).unwrap().is_nan());
    }
}
