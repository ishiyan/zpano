/**
 * Crossover signals.
 *
 * Fuzzy membership for line crossings (e.g., fast MA crossing slow MA)
 * and threshold crossings (e.g., RSI crossing above 30).
 */
import { MembershipShape, muGreater, muLess, tProduct } from '../fuzzy/index.ts';

/** Degree to which a value crossed above threshold from below. */
export function muCrossesAbove(
    prevValue: number, currValue: number, threshold: number,
    width: number = 0.0, shape: MembershipShape = MembershipShape.SIGMOID
): number {
    const wasBelow = muLess(prevValue, threshold, width, shape);
    const isAbove = muGreater(currValue, threshold, width, shape);
    return tProduct(wasBelow, isAbove);
}

/** Degree to which a value crossed below threshold from above. */
export function muCrossesBelow(
    prevValue: number, currValue: number, threshold: number,
    width: number = 0.0, shape: MembershipShape = MembershipShape.SIGMOID
): number {
    const wasAbove = muGreater(prevValue, threshold, width, shape);
    const isBelow = muLess(currValue, threshold, width, shape);
    return tProduct(wasAbove, isBelow);
}

/** Degree to which a fast line crossed above a slow line. */
export function muLineCrossesAbove(
    prevFast: number, currFast: number,
    prevSlow: number, currSlow: number,
    width: number = 0.0, shape: MembershipShape = MembershipShape.SIGMOID
): number {
    const prevDiff = prevFast - prevSlow;
    const currDiff = currFast - currSlow;
    return muCrossesAbove(prevDiff, currDiff, 0.0, width, shape);
}

/** Degree to which a fast line crossed below a slow line. */
export function muLineCrossesBelow(
    prevFast: number, currFast: number,
    prevSlow: number, currSlow: number,
    width: number = 0.0, shape: MembershipShape = MembershipShape.SIGMOID
): number {
    const prevDiff = prevFast - prevSlow;
    const currDiff = currFast - currSlow;
    return muCrossesBelow(prevDiff, currDiff, 0.0, width, shape);
}
