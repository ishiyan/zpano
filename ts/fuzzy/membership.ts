/**
 * Membership functions for fuzzy logic.
 *
 * Each function maps a crisp value to a membership degree μ ∈ [0, 1].
 * Two shapes are supported: SIGMOID (default, smooth) and LINEAR (trapezoidal ramp).
 * All functions degrade to crisp step functions when width = 0.
 */

/**
 * Shape of the fuzzy membership transition curve.
 */
export enum MembershipShape {
    /** Smooth logistic curve. Default for most applications. */
    SIGMOID = 0,
    /** Piecewise-linear ramp (trapezoidal/triangular). */
    LINEAR = 1,
}

/**
 * Steepness constant for sigmoid shape.
 * k = SIGMOID_K / width gives ≈0.997 at threshold ± width/2.
 */
const SIGMOID_K = 12.0;

/**
 * Logistic sigmoid: 1 / (1 + exp(k * (x - threshold))).
 *
 * Returns the "less-than" membership: high when x << threshold,
 * low when x >> threshold, exactly 0.5 at x == threshold.
 */
function sigmoid(x: number, threshold: number, k: number): number {
    const exponent = k * (x - threshold);
    // Clamp to avoid overflow in exp().
    if (exponent > 500.0) {
        return 0.0;
    }
    if (exponent < -500.0) {
        return 1.0;
    }
    return 1.0 / (1.0 + Math.exp(exponent));
}

/**
 * Degree to which x is less than threshold.
 *
 * At threshold: μ = 0.5.
 * At threshold - width/2: μ ≈ 0.997 (sigmoid) or 1.0 (linear).
 * At threshold + width/2: μ ≈ 0.003 (sigmoid) or 0.0 (linear).
 *
 * When width = 0 (crisp): 1.0 if x < threshold, 0.5 if x == threshold,
 * 0.0 if x > threshold.
 */
export function muLess(
    x: number, threshold: number, width: number,
    shape: MembershipShape = MembershipShape.SIGMOID
): number {
    if (width <= 0.0) {
        if (x < threshold) return 1.0;
        if (x > threshold) return 0.0;
        return 0.5;
    }

    if (shape === MembershipShape.LINEAR) {
        const half = width * 0.5;
        if (x <= threshold - half) return 1.0;
        if (x >= threshold + half) return 0.0;
        return (threshold + half - x) / width;
    }
    // sigmoid
    return sigmoid(x, threshold, SIGMOID_K / width);
}

/**
 * Degree to which x ≤ threshold.
 * Identical to muLess for continuous values — the distinction is conceptual.
 */
export function muLessEqual(
    x: number, threshold: number, width: number,
    shape: MembershipShape = MembershipShape.SIGMOID
): number {
    return muLess(x, threshold, width, shape);
}

/**
 * Degree to which x > threshold. Complement of muLess.
 */
export function muGreater(
    x: number, threshold: number, width: number,
    shape: MembershipShape = MembershipShape.SIGMOID
): number {
    return 1.0 - muLess(x, threshold, width, shape);
}

/**
 * Degree to which x ≥ threshold. Complement of muLessEqual.
 */
export function muGreaterEqual(
    x: number, threshold: number, width: number,
    shape: MembershipShape = MembershipShape.SIGMOID
): number {
    return 1.0 - muLessEqual(x, threshold, width, shape);
}

/**
 * Bell-shaped membership: degree to which x ≈ target.
 *
 * μ = 1.0 at x == target.
 * μ ≈ 0 at |x - target| ≥ width.
 *
 * For sigmoid shape: Gaussian bell exp(-k * (x - target)²).
 * For linear shape: triangular peak at target with base 2 * width.
 */
export function muNear(
    x: number, target: number, width: number,
    shape: MembershipShape = MembershipShape.SIGMOID
): number {
    if (width <= 0.0) {
        return x === target ? 1.0 : 0.0;
    }

    if (shape === MembershipShape.LINEAR) {
        const dist = Math.abs(x - target);
        if (dist >= width) return 0.0;
        return 1.0 - dist / width;
    }
    // sigmoid → Gaussian bell
    // σ chosen so that μ ≈ 0.003 at |x - target| = width.
    const sigma = width / 2.41;
    const d = (x - target) / sigma;
    return Math.exp(-d * d);
}

/**
 * Fuzzy candle direction ∈ [-1, +1].
 *
 * +1 = fully bullish (large white body).
 *  0 = neutral (doji-like).
 * -1 = fully bearish (large black body).
 *
 * Uses tanh(steepness * (c - o) / bodyAvg).
 * When bodyAvg ≤ 0: returns +1.0 if c ≥ o, else -1.0 (crisp).
 */
export function muDirection(
    o: number, c: number, bodyAvg: number,
    steepness: number = 2.0
): number {
    if (bodyAvg <= 0.0) {
        return c >= o ? 1.0 : -1.0;
    }
    return Math.tanh(steepness * (c - o) / bodyAvg);
}
