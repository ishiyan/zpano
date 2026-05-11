/**
 * Histogram sign-change signals.
 *
 * Fuzzy membership for oscillator histograms turning positive or negative
 * (e.g., MACD histogram crossing the zero line).
 */
import { MembershipShape, muGreater, muLess, tProduct } from '../fuzzy/index.ts';

/** Degree to which a histogram turned from non-positive to positive. */
export function muTurnsPositive(
    prevValue: number, currValue: number,
    width: number = 0.0, shape: MembershipShape = MembershipShape.SIGMOID
): number {
    const wasNonpositive = muLess(prevValue, 0.0, width, shape);
    const isPositive = muGreater(currValue, 0.0, width, shape);
    return tProduct(wasNonpositive, isPositive);
}

/** Degree to which a histogram turned from non-negative to negative. */
export function muTurnsNegative(
    prevValue: number, currValue: number,
    width: number = 0.0, shape: MembershipShape = MembershipShape.SIGMOID
): number {
    const wasNonnegative = muGreater(prevValue, 0.0, width, shape);
    const isNegative = muLess(currValue, 0.0, width, shape);
    return tProduct(wasNonnegative, isNegative);
}
