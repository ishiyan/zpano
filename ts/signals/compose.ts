/**
 * Signal composition utilities.
 *
 * Combine multiple fuzzy signals using t-norms, s-norms, and negation.
 */
import { tProductAll, sProbabilistic, fNot } from '../fuzzy/index.ts';

/** Combine signals with product t-norm (fuzzy AND). */
export function signalAnd(...signals: number[]): number {
    return tProductAll(...signals);
}

/** Combine two signals with probabilistic s-norm (fuzzy OR). */
export function signalOr(a: number, b: number): number {
    return sProbabilistic(a, b);
}

/** Negate a signal (fuzzy complement). Returns 1 - signal. */
export function signalNot(signal: number): number {
    return fNot(signal);
}

/**
 * Filter weak signals below minStrength to zero.
 * Signals at or above the threshold pass through unchanged.
 */
export function signalStrength(signal: number, minStrength: number = 0.5): number {
    return signal >= minStrength ? signal : 0.0;
}
