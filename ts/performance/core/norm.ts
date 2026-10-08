// Python's `math.erf` (CPython 3.13+) delegates to the C library `erf`,
// which in glibc is the fdlibm (Sun Microsystems) implementation.
// JavaScript has no `Math.erf`, so the fdlibm algorithm is ported below.

const ERX = 8.45062911510467529297e-01;
// Coefficients for approximation to erf in [0, 0.84375].
const EFX = 1.28379167095512586316e-01;
const EFX8 = 1.02703333676410069053e+00;
const PP0 = 1.28379167095512558561e-01;
const PP1 = -3.25042107247001499370e-01;
const PP2 = -2.84817495755985104766e-02;
const PP3 = -5.77027029648944159157e-03;
const PP4 = -2.37630166566501626084e-05;
const QQ1 = 3.97917223959155352819e-01;
const QQ2 = 6.50222499887672944485e-02;
const QQ3 = 5.08130628187576562776e-03;
const QQ4 = 1.32494738004321644526e-04;
const QQ5 = -3.96022827877536812320e-06;
// Coefficients for approximation to erf in [0.84375, 1.25].
const PA0 = -2.36211856075265944077e-03;
const PA1 = 4.14856118683748331666e-01;
const PA2 = -3.72207876035701323847e-01;
const PA3 = 3.18346619901161753674e-01;
const PA4 = -1.10894694282396677476e-01;
const PA5 = 3.54783043256182359371e-02;
const PA6 = -2.16637559486879084300e-03;
const QA1 = 1.06420880400844228286e-01;
const QA2 = 5.40397917702171048937e-01;
const QA3 = 7.18286544141962662868e-02;
const QA4 = 1.26171219808761642112e-01;
const QA5 = 1.36370839120290507362e-02;
const QA6 = 1.19844998467991074170e-02;
// Coefficients for approximation to erfc in [1.25, 1/0.35].
const RA0 = -9.86494403484714822705e-03;
const RA1 = -6.93858572707181764372e-01;
const RA2 = -1.05586262253232909814e+01;
const RA3 = -6.23753324503260060396e+01;
const RA4 = -1.62396669462573470355e+02;
const RA5 = -1.84605092906711035994e+02;
const RA6 = -8.12874355063065934246e+01;
const RA7 = -9.81432934416914548592e+00;
const SA1 = 1.96512716674392571292e+01;
const SA2 = 1.37657754143519042600e+02;
const SA3 = 4.34565877475229228821e+02;
const SA4 = 6.45387271733267880336e+02;
const SA5 = 4.29008140027567833386e+02;
const SA6 = 1.08635005541779435134e+02;
const SA7 = 6.57024977031928170135e+00;
const SA8 = -6.04244152148580987438e-02;
// Coefficients for approximation to erfc in [1/0.35, 28].
const RB0 = -9.86494292470009928597e-03;
const RB1 = -7.99283237680523006574e-01;
const RB2 = -1.77579549177547519889e+01;
const RB3 = -1.60636384855821916062e+02;
const RB4 = -6.37566443368389627722e+02;
const RB5 = -1.02509513161107724954e+03;
const RB6 = -4.83519191608651397019e+02;
const SB1 = 3.03380607434824582924e+01;
const SB2 = 3.25792512996573918826e+02;
const SB3 = 1.53672958608443695994e+03;
const SB4 = 3.19985821950859553908e+03;
const SB5 = 2.55305040643316442583e+03;
const SB6 = 4.74528541206955367215e+02;
const SB7 = -2.24409524465858183362e+01;

/** 2**-28. */
const ERF_SMALL = 1.0 / (1 << 28);
/** 0x0080000000000000 (fdlibm's 0x00800000 high-word threshold). */
const ERF_VERY_TINY = 2.848094538889218e-306;
/** fdlibm's |x| < 1/0.35 high-word threshold 0x4006DB6E as a double. */
const ERF_ONE_OVER_035 = 2.8571434020996094;

const erfBuffer = new DataView(new ArrayBuffer(8));

/** Truncates the low 32 bits of the mantissa of x (fdlibm's `SET_LOW_WORD(z, 0)`). */
function clearLowWord(x: number): number {
    erfBuffer.setFloat64(0, x);
    erfBuffer.setUint32(4, 0);
    return erfBuffer.getFloat64(0);
}

/**
 * The error function, a port of the fdlibm `erf` used by glibc and therefore
 * by Python's `math.erf`.
 */
function erf(x: number): number {
    if (Number.isNaN(x)) {
        return NaN;
    }
    if (x === Infinity) {
        return 1;
    }
    if (x === -Infinity) {
        return -1;
    }
    let sign = false;
    if (x < 0) {
        x = -x;
        sign = true;
    }
    if (x < 0.84375) {
        let temp: number;
        if (x < ERF_SMALL) {
            if (x < ERF_VERY_TINY) {
                temp = 0.125 * (8.0 * x + EFX8 * x); // avoid underflow
            } else {
                temp = x + EFX * x;
            }
        } else {
            const z = x * x;
            const r = PP0 + z * (PP1 + z * (PP2 + z * (PP3 + z * PP4)));
            const s = 1 + z * (QQ1 + z * (QQ2 + z * (QQ3 + z * (QQ4 + z * QQ5))));
            const y = r / s;
            temp = x + x * y;
        }
        return sign ? -temp : temp;
    }
    if (x < 1.25) {
        const s = x - 1;
        const p = PA0 + s * (PA1 + s * (PA2 + s * (PA3 + s * (PA4 + s * (PA5 + s * PA6)))));
        const q = 1 + s * (QA1 + s * (QA2 + s * (QA3 + s * (QA4 + s * (QA5 + s * QA6)))));
        return sign ? -ERX - p / q : ERX + p / q;
    }
    if (x >= 6) {
        return sign ? -1 : 1;
    }
    const s = 1 / (x * x);
    let rr: number;
    let ss: number;
    if (x < ERF_ONE_OVER_035) {
        rr = RA0 + s * (RA1 + s * (RA2 + s * (RA3 + s * (RA4 + s * (RA5 + s * (RA6 + s * RA7))))));
        ss = 1 + s * (SA1 + s * (SA2 + s * (SA3 + s * (SA4 + s * (SA5 + s * (SA6 + s * (SA7 + s * SA8)))))));
    } else {
        rr = RB0 + s * (RB1 + s * (RB2 + s * (RB3 + s * (RB4 + s * (RB5 + s * RB6)))));
        ss = 1 + s * (SB1 + s * (SB2 + s * (SB3 + s * (SB4 + s * (SB5 + s * (SB6 + s * SB7))))));
    }
    const z = clearLowWord(x);
    const r = Math.exp(-z * z - 0.5625) * Math.exp((z - x) * (z + x) + rr / ss);
    return sign ? r / x - 1 : 1 - r / x;
}

// _SQRT2 = math.sqrt(2.0)
const SQRT2 = 1.4142135623730950488016887242097;

/**
 * Standard normal cumulative distribution function (CDF).
 *
 * Computes P(Z ≤ z) for Z ~ N(0, 1):
 *
 *     Φ(z) = ½ (1 + erf(z / √2))
 *
 * `erf` is a port of the fdlibm implementation used by Python's `math.erf`
 * and is accurate to near machine precision.
 *
 * For a normal distribution N(μ, σ²), standardize first:
 *
 *     Φ((x - μ) / σ)
 */
export function normCdf(z: number): number {
    return 0.5 * (1.0 + erf(z / SQRT2));
}

// _INV_SQRT_2PI = 1.0 / math.sqrt(2.0 * math.pi)
const INV_SQRT_2PI = 0.39894228040143267793994605993438;

/**
 * Standard normal probability density function (PDF).
 *
 * Computes
 *
 *     φ(z) = exp(-z² / 2) / √(2π)
 *
 * for Z ~ N(0, 1).
 *
 * For a normal distribution N(μ, σ²),
 *
 *     φ((x - μ) / σ) / σ
 */
export function normPdf(z: number): number {
    return INV_SQRT_2PI * Math.exp(-0.5 * z * z);
}

// The inverse standard normal CDF has no closed-form expression.
//
// This implementation uses Peter J. Acklam's piecewise rational
// approximation, which divides the probability domain into lower
// tail, central region, and upper tail.
//
// Separate rational functions are used in each region because no
// single polynomial provides uniform accuracy over (0, 1).
//
// The upper-tail approximation exploits the symmetry
//
//     Φ⁻¹(1 - p) = -Φ⁻¹(p)
//
// to reuse the lower-tail coefficients.

// Acklam's published polynomial coefficients.
//
// A, B : central region
// C, D : lower and upper tails

const A = [
    -3.969683028665376e+01, 2.209460984245205e+02,
    -2.759285104469687e+02, 1.383577518672690e+02,
    -3.066479806614716e+01, 2.506628277459239e+00,
] as const;

const B = [
    -5.447609879822406e+01, 1.615858368580409e+02,
    -1.556989798598866e+02, 6.680131188771972e+01,
    -1.328068155288572e+01,
] as const;

const C = [
    -7.784894002430293e-03, -3.223964580411365e-01,
    -2.400758277161838e+00, -2.549732539343734e+00,
    4.374664141464968e+00, 2.938163982698783e+00,
] as const;

const D = [
    7.784695709041462e-03, 3.224671290700398e-01,
    2.445134137142996e+00, 3.754408661907416e+00,
] as const;

/**
 * Standard normal percent-point function (inverse CDF).
 *
 * Computes z such that
 *
 *     P(Z ≤ z) = p
 *
 * for Z ~ N(0, 1).
 *
 * Uses Peter J. Acklam's piecewise rational approximation.
 *
 * The approximation is accurate to roughly 1×10⁻⁹ relative error
 * over most of the domain, consistent with Acklam's published results.
 *
 * Throws an Error if p is not in (0, 1).
 */
export function normPpf(p: number): number {
    if (p <= 0 || p >= 1) {
        throw new Error('p must be between 0 and 1 (exclusive)');
    }

    const pLow = 0.02425;
    const pHigh = 1.0 - pLow;
    if (p < pLow) {
        // Lower tail
        const q = Math.sqrt(-2.0 * Math.log(p));
        return (((((C[0] * q + C[1]) * q + C[2]) * q + C[3]) * q + C[4]) * q + C[5]) /
            ((((D[0] * q + D[1]) * q + D[2]) * q + D[3]) * q + 1.0);
    } else if (p <= pHigh) {
        // Central region
        const q = p - 0.5;
        const r = q * q;
        return (((((A[0] * r + A[1]) * r + A[2]) * r + A[3]) * r + A[4]) * r + A[5]) * q /
            (((((B[0] * r + B[1]) * r + B[2]) * r + B[3]) * r + B[4]) * r + 1.0);
    } else {
        // Upper tail
        const q = Math.sqrt(-2.0 * Math.log(1.0 - p));
        return -(((((C[0] * q + C[1]) * q + C[2]) * q + C[3]) * q + C[4]) * q + C[5]) /
            ((((D[0] * q + D[1]) * q + D[2]) * q + D[3]) * q + 1.0);
    }
}
