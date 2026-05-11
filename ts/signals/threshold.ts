/**
 * Threshold crossing signals.
 *
 * Fuzzy membership for indicator values relative to fixed thresholds
 * (e.g., RSI > 70, Stochastic < 20).
 */
import { MembershipShape, muGreater, muLess } from '../fuzzy/index.ts';

/** Degree to which value is above threshold. */
export function muAbove(
    value: number, threshold: number,
    width: number = 5.0, shape: MembershipShape = MembershipShape.SIGMOID
): number {
    return muGreater(value, threshold, width, shape);
}

/** Degree to which value is below threshold. Complement of muAbove. */
export function muBelow(
    value: number, threshold: number,
    width: number = 5.0, shape: MembershipShape = MembershipShape.SIGMOID
): number {
    return muLess(value, threshold, width, shape);
}

/** Degree of overbought condition. Default level 70 matches common RSI interpretation. */
export function muOverbought(
    value: number, level: number = 70.0,
    width: number = 5.0, shape: MembershipShape = MembershipShape.SIGMOID
): number {
    return muGreater(value, level, width, shape);
}

/** Degree of oversold condition. Default level 30 matches common RSI interpretation. */
export function muOversold(
    value: number, level: number = 30.0,
    width: number = 5.0, shape: MembershipShape = MembershipShape.SIGMOID
): number {
    return muLess(value, level, width, shape);
}
