/** Counterattack pattern (2-candle reversal). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody } from '../core/primitives.ts';

/**
 * Counterattack: a two-candle reversal pattern.
 *
 * Two long candles of opposite color with closes that are equal
 * (or very near equal).
 *
 * - bullish: first candle is long black, second is long white,
 *   closes are equal,
 * - bearish: first candle is long white, second is long black,
 *   closes are equal.
 *
 * The meaning of "long" is specified with longBody.
 * The meaning of "equal" is specified with equal.
 *
 * Category B: direction from 2nd candle color (continuous).
 *
 * Returns:
 *     Continuous float in [-100, +100].
 */
export function counterattack(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(2, cp.longBody, cp.equal)) return 0.0;

    const b1 = cp.bar(2);
    const b2 = cp.bar(1);

    // Opposite colors — crisp gate.
    const color1 = b1.c < b1.o ? -1 : 1;
    const color2 = b2.c < b2.o ? -1 : 1;
    if (color1 === color2) return 0.0;

    // Fuzzy conditions.
    const muLong1 = cp.muGreaterCS(realBody(b1.o, b1.c), cp.longBody, 2);
    const muLong2 = cp.muGreaterCS(realBody(b2.o, b2.c), cp.longBody, 1);
    // Closes near equal: model as muLess(absDiff, eqAvg) — crossover at eq boundary.
    const muEq = cp.muLessCS(Math.abs(b2.c - b1.c), cp.equal, 2);

    const confidence = tProductAll(muLong1, muLong2, muEq);
    // Direction from 2nd candle color.
    const direction = b2.c < b2.o ? -1.0 : 1.0;
    return direction * confidence * 100.0;
}
