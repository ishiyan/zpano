//! Standard normal distribution functions: CDF, PDF and inverse CDF (PPF).

const std = @import("std");
const math = std.math;
const testing = std.testing;

/// Errors returned by `normPpf`.
pub const Error = error{InvalidArgument};

//const sqrt2 = math.sqrt(2.0);
const sqrt2: f64 = 1.4142135623730950488016887242097;

// ── erf ────────────────────────────────────────────────────────────────────
//
// The Zig standard library has no erf.  Python's math.erf delegates to the
// C library erf, which (glibc, musl, FreeBSD) derives from Sun's fdlibm
// s_erf.c.  This is a port of that algorithm (via Go's math/erf.go).
//
// Copyright (C) 1993 by Sun Microsystems, Inc. All rights reserved.
// Developed at SunPro, a Sun Microsystems, Inc. business.
// Permission to use, copy, modify, and distribute this software is freely
// granted, provided that this notice is preserved.

const erx: f64 = 8.45062911510467529297e-01;
// Coefficients for approximation to erf in [0, 0.84375]
const efx: f64 = 1.28379167095512586316e-01;
const efx8: f64 = 1.02703333676410069053e+00;
const pp0: f64 = 1.28379167095512558561e-01;
const pp1: f64 = -3.25042107247001499370e-01;
const pp2: f64 = -2.84817495755985104766e-02;
const pp3: f64 = -5.77027029648944159157e-03;
const pp4: f64 = -2.37630166566501626084e-05;
const qq1: f64 = 3.97917223959155352819e-01;
const qq2: f64 = 6.50222499887672944485e-02;
const qq3: f64 = 5.08130628187576562776e-03;
const qq4: f64 = 1.32494738004321644526e-04;
const qq5: f64 = -3.96022827877536812320e-06;
// Coefficients for approximation to erf in [0.84375, 1.25]
const pa0: f64 = -2.36211856075265944077e-03;
const pa1: f64 = 4.14856118683748331666e-01;
const pa2: f64 = -3.72207876035701323847e-01;
const pa3: f64 = 3.18346619901161753674e-01;
const pa4: f64 = -1.10894694282396677476e-01;
const pa5: f64 = 3.54783043256182359371e-02;
const pa6: f64 = -2.16637559486879084300e-03;
const qa1: f64 = 1.06420880400844228286e-01;
const qa2: f64 = 5.40397917702171048937e-01;
const qa3: f64 = 7.18286544141962662868e-02;
const qa4: f64 = 1.26171219808761642112e-01;
const qa5: f64 = 1.36370839120290507362e-02;
const qa6: f64 = 1.19844998467991074170e-02;
// Coefficients for approximation to erfc in [1.25, 1/0.35]
const ra0: f64 = -9.86494403484714822705e-03;
const ra1: f64 = -6.93858572707181764372e-01;
const ra2: f64 = -1.05586262253232909814e+01;
const ra3: f64 = -6.23753324503260060396e+01;
const ra4: f64 = -1.62396669462573470355e+02;
const ra5: f64 = -1.84605092906711035994e+02;
const ra6: f64 = -8.12874355063065934246e+01;
const ra7: f64 = -9.81432934416914548592e+00;
const sa1: f64 = 1.96512716674392571292e+01;
const sa2: f64 = 1.37657754143519042600e+02;
const sa3: f64 = 4.34565877475229228821e+02;
const sa4: f64 = 6.45387271733267880336e+02;
const sa5: f64 = 4.29008140027567833386e+02;
const sa6: f64 = 1.08635005541779435134e+02;
const sa7: f64 = 6.57024977031928170135e+00;
const sa8: f64 = -6.04244152148580987438e-02;
// Coefficients for approximation to erfc in [1/.35, 28]
const rb0: f64 = -9.86494292470009928597e-03;
const rb1: f64 = -7.99283237680523006574e-01;
const rb2: f64 = -1.77579549177547519889e+01;
const rb3: f64 = -1.60636384855821916062e+02;
const rb4: f64 = -6.37566443368389627722e+02;
const rb5: f64 = -1.02509513161107724954e+03;
const rb6: f64 = -4.83519191608651397019e+02;
const sb1: f64 = 3.03380607434824582924e+01;
const sb2: f64 = 3.25792512996573918826e+02;
const sb3: f64 = 1.53672958608443695994e+03;
const sb4: f64 = 3.19985821950859553908e+03;
const sb5: f64 = 2.55305040643316442583e+03;
const sb6: f64 = 4.74528541206955367215e+02;
const sb7: f64 = -2.24409524465858183362e+01;

/// The error function erf(x) = 2/√π ∫₀ˣ exp(−t²) dt (fdlibm algorithm,
/// matching the C library erf used by Python's `math.erf`).
pub fn erf(x_in: f64) f64 {
    const very_tiny: f64 = 2.848094538889218e-306; // 0x0080000000000000
    const small: f64 = 1.0 / @as(f64, 1 << 28); // 2**-28
    if (math.isNan(x_in)) return math.nan(f64);
    if (math.isPositiveInf(x_in)) return 1;
    if (math.isNegativeInf(x_in)) return -1;
    var x = x_in;
    var sign = false;
    if (x < 0) {
        x = -x;
        sign = true;
    }
    if (x < 0.84375) { // |x| < 0.84375
        var temp: f64 = undefined;
        if (x < small) { // |x| < 2**-28
            if (x < very_tiny) {
                temp = 0.125 * (8.0 * x + efx8 * x); // avoid underflow
            } else {
                temp = x + efx * x;
            }
        } else {
            const z = x * x;
            const r = pp0 + z * (pp1 + z * (pp2 + z * (pp3 + z * pp4)));
            const s = 1 + z * (qq1 + z * (qq2 + z * (qq3 + z * (qq4 + z * qq5))));
            const y = r / s;
            temp = x + x * y;
        }
        return if (sign) -temp else temp;
    }
    if (x < 1.25) { // 0.84375 <= |x| < 1.25
        const s = x - 1;
        const p = pa0 + s * (pa1 + s * (pa2 + s * (pa3 + s * (pa4 + s * (pa5 + s * pa6)))));
        const q = 1 + s * (qa1 + s * (qa2 + s * (qa3 + s * (qa4 + s * (qa5 + s * qa6)))));
        return if (sign) -erx - p / q else erx + p / q;
    }
    if (x >= 6) { // inf > |x| >= 6
        return if (sign) -1 else 1;
    }
    const s = 1 / (x * x);
    var r_: f64 = undefined;
    var s_: f64 = undefined;
    if (x < 1.0 / 0.35) { // |x| < 1 / 0.35  ~ 2.857143
        r_ = ra0 + s * (ra1 + s * (ra2 + s * (ra3 + s * (ra4 + s * (ra5 + s * (ra6 + s * ra7))))));
        s_ = 1 + s * (sa1 + s * (sa2 + s * (sa3 + s * (sa4 + s * (sa5 + s * (sa6 + s * (sa7 + s * sa8)))))));
    } else { // |x| >= 1 / 0.35  ~ 2.857143
        r_ = rb0 + s * (rb1 + s * (rb2 + s * (rb3 + s * (rb4 + s * (rb5 + s * rb6)))));
        s_ = 1 + s * (sb1 + s * (sb2 + s * (sb3 + s * (sb4 + s * (sb5 + s * (sb6 + s * sb7))))));
    }
    // pseudo-single (20-bit) precision x
    const z: f64 = @bitCast(@as(u64, @bitCast(x)) & 0xffffffff00000000);
    const r = @exp(-z * z - 0.5625) * @exp((z - x) * (z + x) + r_ / s_);
    return if (sign) r / x - 1 else 1 - r / x;
}

/// Standard normal cumulative distribution function (CDF).
///
/// Computes P(Z ≤ z) for Z ~ N(0, 1):
///
///     Φ(z) = ½ (1 + erf(z / √2))
///
/// For a normal distribution N(μ, σ²), standardize first:
///
///     Φ((x - μ) / σ)
pub fn normCdf(z: f64) f64 {
    return 0.5 * (1.0 + erf(z / sqrt2));
}

//const inv_sqrt_2pi = 1.0 / math.sqrt(2.0 * math.pi);
const inv_sqrt_2pi: f64 = 0.39894228040143267793994605993438;

/// Standard normal probability density function (PDF).
///
/// Computes
///
///     φ(z) = exp(-z² / 2) / √(2π)
///
/// for Z ~ N(0, 1).
///
/// For a normal distribution N(μ, σ²),
///
///     φ((x - μ) / σ) / σ
pub fn normPdf(z: f64) f64 {
    return inv_sqrt_2pi * @exp(-0.5 * z * z);
}

// The inverse standard normal CDF has no closed-form expression.
//
// This implementation uses Peter J. Acklam's piecewise rational
// approximation, which divides the probability domain into lower
// tail, central region, and upper tail.
//
// The upper-tail approximation exploits the symmetry
//
//     Φ⁻¹(1 - p) = -Φ⁻¹(p)
//
// to reuse the lower-tail coefficients.
//
// Acklam's published polynomial coefficients.
//
// a, b : central region
// c, d : lower and upper tails

const a = [_]f64{
    -3.969683028665376e+01, 2.209460984245205e+02,
    -2.759285104469687e+02, 1.383577518672690e+02,
    -3.066479806614716e+01, 2.506628277459239e+00,
};

const b = [_]f64{
    -5.447609879822406e+01, 1.615858368580409e+02,
    -1.556989798598866e+02, 6.680131188771972e+01,
    -1.328068155288572e+01,
};

const c = [_]f64{
    -7.784894002430293e-03, -3.223964580411365e-01,
    -2.400758277161838e+00, -2.549732539343734e+00,
    4.374664141464968e+00,  2.938163982698783e+00,
};

const d = [_]f64{
    7.784695709041462e-03, 3.224671290700398e-01,
    2.445134137142996e+00, 3.754408661907416e+00,
};

/// Standard normal percent-point function (inverse CDF).
///
/// Computes z such that
///
///     P(Z ≤ z) = p
///
/// for Z ~ N(0, 1).
///
/// Uses Peter J. Acklam's piecewise rational approximation, accurate to
/// roughly 1×10⁻⁹ relative error over most of the domain.
///
/// Returns `error.InvalidArgument` unless 0 < p < 1 (Python ValueError).
/// A NaN p is not rejected (as in Python) and yields NaN.
pub fn normPpf(p: f64) Error!f64 {
    if (p <= 0 or p >= 1) return error.InvalidArgument;

    const p_low = 0.02425;
    const p_high = 1.0 - p_low;
    if (p < p_low) {
        // Lower tail
        const q = @sqrt(-2.0 * @log(p));
        return (((((c[0] * q + c[1]) * q + c[2]) * q + c[3]) * q + c[4]) * q + c[5]) /
            ((((d[0] * q + d[1]) * q + d[2]) * q + d[3]) * q + 1.0);
    } else if (p <= p_high) {
        // Central region
        const q = p - 0.5;
        const r = q * q;
        return (((((a[0] * r + a[1]) * r + a[2]) * r + a[3]) * r + a[4]) * r + a[5]) * q /
            (((((b[0] * r + b[1]) * r + b[2]) * r + b[3]) * r + b[4]) * r + 1.0);
    } else {
        // Upper tail
        const q = @sqrt(-2.0 * @log(1.0 - p));
        return -(((((c[0] * q + c[1]) * q + c[2]) * q + c[3]) * q + c[4]) * q + c[5]) /
            ((((d[0] * q + d[1]) * q + d[2]) * q + d[3]) * q + 1.0);
    }
}

// ── Tests ──────────────────────────────────────────────────────────────────

fn expectAlmostEqual(expected: f64, actual: f64, places: u5) !void {
    if (expected == actual) return;
    const tol = 0.5 * math.pow(f64, 10.0, -@as(f64, @floatFromInt(places)));
    if (!(@abs(expected - actual) <= tol)) {
        std.debug.print("expected {d}, got {d} (places {d})\n", .{ expected, actual, places });
        return error.TestExpectedApproxEq;
    }
}

fn expectWithinDelta(expected: f64, actual: f64, delta: f64) !void {
    if (expected == actual) return;
    if (!(@abs(expected - actual) <= delta)) {
        std.debug.print("expected {d}, got {d} (delta {e})\n", .{ expected, actual, delta });
        return error.TestExpectedApproxEq;
    }
}

const sample_inputs = [_]f64{ -8.0, -5.0, -3.0, -2.0, -1.0, 0.0, 1.0, 2.0, 3.0, 5.0, 8.0 };

test "pdf scipy compliance" {
    const outputs = [_]f64{
        5.052271083536893e-15,  1.4867195147342979e-06, 0.0044318484119380075,
        0.053990966513188056,   0.24197072451914337,    0.3989422804014327,
        0.24197072451914337,    0.053990966513188056,   0.0044318484119380075,
        1.4867195147342979e-06, 5.052271083536893e-15,
    };
    for (sample_inputs, outputs) |x, expected| {
        try expectAlmostEqual(expected, normPdf(x), 15);
        // Symmetry: norm_pdf(x) == norm_pdf(-x)
        try expectAlmostEqual(normPdf(x), normPdf(-x), 15);
    }
}

test "cdf scipy compliance" {
    // Notice that 2 is just inside Acklam's upper-tail threshold (CDF(2) ≈ 0.97725 > 0.97575),
    // so tests naturally cover the branch boundary.
    const outputs = [_]f64{
        6.22096057427174e-16, 2.866515718791933e-07, 0.0013498980316300933,
        0.022750131948179195, 0.15865525393145707,   0.5,
        0.8413447460685429,   0.9772498680518208,    0.9986501019683699,
        0.9999997133484281,   0.9999999999999993,
    };
    for (sample_inputs, outputs) |x, expected| {
        try expectAlmostEqual(expected, normCdf(x), 15);
        // Symmetry: norm_cdf(x) + norm_cdf(-x) = 1.0
        try expectAlmostEqual(1.0, normCdf(x) + normCdf(-x), 15);
    }
}

test "ppf scipy compliance" {
    const p_values = [_]f64{
        // Center
        0.5,            0.75,       0.9,
        // Branch boundaries (values around the transition)
        0.02425,        0.025,      0.975,
        0.97575,
        // Common quantiles
               0.001,      0.005,
        0.01,           0.025,      0.05,
        0.10,           0.90,       0.95,
        0.975,          0.99,       0.995,
        0.999,
        // Extreme tails (exercise numerical stability)
                 1e-12,      1e-10,
        1e-8,           0.99999999, 0.9999999999,
        0.999999999999,
    };
    // Hardcoded 1-p_values
    const p_symmetric_values = [_]f64{
        // Center
        0.5,        0.25,           0.1,
        // Branch boundaries (values around the transition)
        0.97575,    0.975,          0.025,
        0.02425,
        // Common quantiles
           0.999,          0.995,
        0.99,       0.975,          0.95,
        0.90,       0.10,           0.05,
        0.025,      0.01,           0.005,
        0.001,
        // Extreme tails (exercise numerical stability)
             0.999999999999, 0.9999999999,
        0.99999999, 1e-8,           1e-10,
        1e-12,
    };
    const z_values = [_]f64{
        // Center
        0.0,                 0.6744897501960817,  1.2815515655446004,
        // Branch boundaries (values around the transition)
        -1.972961051311885,  -1.9599639845400545, 1.959963984540054,
        1.972961051311885,
        // Common quantiles
          -3.090232306167813,  -2.575829303548901,
        -2.3263478740408408, -1.9599639845400545, -1.6448536269514729,
        -1.2815515655446004, 1.2815515655446004,  1.6448536269514722,
        1.959963984540054,   2.3263478740408408,  2.5758293035489004,
        3.090232306167,
        // Extreme tails (exercise numerical stability)
             -7.034483825301131,  -6.361340902404056,
        -5.612001244174789,  5.612001243305505,   6.361340889697422,
        7.0344869100478356,
    };
    for (p_values, p_symmetric_values, z_values, 0..) |p, p_symmetric, expected, i| {
        // Extreme tails have much lower accuracy.
        const delta: f64 = if (i < 19) 3e-9 else 8e-9;
        try expectWithinDelta(expected, try normPpf(p), delta);

        // Verify that norm_ppf is the inverse of norm_cdf: p ≈ Φ(Φ⁻¹(p))
        try expectWithinDelta(p, normCdf(try normPpf(p)), delta);

        // Symmetry: norm_ppf(p) == - norm_ppf(1.0 - p)
        const sym_delta: f64 = if (i < 19) 1e-13 else 4e-6;
        try expectWithinDelta(try normPpf(p), -(try normPpf(p_symmetric)), sym_delta);
    }
}

test "ppf rejects p outside (0, 1)" {
    try testing.expectError(error.InvalidArgument, normPpf(0.0));
    try testing.expectError(error.InvalidArgument, normPpf(1.0));
    try testing.expectError(error.InvalidArgument, normPpf(-0.5));
    try testing.expectError(error.InvalidArgument, normPpf(1.5));
}
