/**
 * Fuzzy logic operators: t-norms, s-norms, and negation.
 *
 * T-norms implement fuzzy AND. S-norms implement fuzzy OR.
 * All operators take membership degrees in [0, 1] and return [0, 1].
 */

// -----------------------------------------------------------------------
// T-norms (fuzzy AND)
// -----------------------------------------------------------------------

/** Product t-norm: a * b. All conditions contribute proportionally. */
export function tProduct(a: number, b: number): number {
    return a * b;
}

/** Minimum t-norm (Zadeh): min(a, b). Dominated by the weakest condition. */
export function tMin(a: number, b: number): number {
    return Math.min(a, b);
}

/** Łukasiewicz t-norm: max(0, a + b - 1). Very strict. */
export function tLukasiewicz(a: number, b: number): number {
    return Math.max(0.0, a + b - 1.0);
}

// -----------------------------------------------------------------------
// S-norms (fuzzy OR)
// -----------------------------------------------------------------------

/** Probabilistic sum: a + b - a*b. Dual of the product t-norm. */
export function sProbabilistic(a: number, b: number): number {
    return a + b - a * b;
}

/** Maximum s-norm (Zadeh): max(a, b). Dual of the minimum t-norm. */
export function sMax(a: number, b: number): number {
    return Math.max(a, b);
}

// -----------------------------------------------------------------------
// Negation
// -----------------------------------------------------------------------

/** Standard fuzzy negation: 1 - a. */
export function fNot(a: number): number {
    return 1.0 - a;
}

// -----------------------------------------------------------------------
// Variadic helpers
// -----------------------------------------------------------------------

/**
 * Product t-norm over multiple arguments.
 * Returns 1.0 for zero arguments (identity element of product).
 */
export function tProductAll(...args: number[]): number {
    let result = 1.0;
    for (const a of args) {
        result *= a;
    }
    return result;
}

/**
 * Minimum t-norm over multiple arguments.
 * Returns 1.0 for zero arguments (identity element of min).
 */
export function tMinAll(...args: number[]): number {
    if (args.length === 0) return 1.0;
    let result = args[0];
    for (let i = 1; i < args.length; i++) {
        if (args[i] < result) result = args[i];
    }
    return result;
}
