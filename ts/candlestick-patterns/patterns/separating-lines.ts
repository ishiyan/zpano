/** Separating Lines pattern (2-candle continuation). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, upperShadow, lowerShadow } from '../core/primitives.ts';

/**
 * Separating Lines: a two-candle continuation pattern.
 *
 * Opposite colors with the same open. The second candle is a belt hold
 * (long body with no shadow on the opening side).
 *
 * - bullish: first candle is black, second is white with same open,
 *   long body, very short lower shadow,
 * - bearish: first candle is white, second is black with same open,
 *   long body, very short upper shadow.
 *
 * The meaning of "long" is specified with `longBody`.
 * The meaning of "very short" for shadows is specified with
 * `veryShortShadow`.
 * The meaning of "equal" is specified with `equal`.
 *
 * Category C: both branches evaluated, return stronger signal.
 *
 * Returns:
 *     Continuous float in [-100, +100].
 */
export function separatingLines(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(2, cp.longBody, cp.veryShortShadow, cp.equal)) return 0.0;

    const b1 = cp.bar(2);
    const b2 = cp.bar(1);

    // Opposite colors — crisp gate.
    const color1 = b1.c < b1.o ? -1 : 1;
    const color2 = b2.c < b2.o ? -1 : 1;
    if (color1 === color2) return 0.0;

    // Opens near equal — fuzzy (crisp was abs(o2-o1) <= eq).
    const muEq = cp.muLessCS(Math.abs(b2.o - b1.o), cp.equal, 2);
    // Long body on 2nd candle — fuzzy.
    const muLong = cp.muGreaterCS(realBody(b2.o, b2.c), cp.longBody, 1);

    // Bullish: white belt hold (very short lower shadow).
    let bullSignal = 0.0;
    if (color2 === 1) {
        const muVS = cp.muLessCS(lowerShadow(b2.o, b2.l, b2.c), cp.veryShortShadow, 1);
        bullSignal = tProductAll(muEq, muLong, muVS) * 100.0;
    }

    // Bearish: black belt hold (very short upper shadow).
    let bearSignal = 0.0;
    if (color2 === -1) {
        const muVS = cp.muLessCS(upperShadow(b2.o, b2.h, b2.c), cp.veryShortShadow, 1);
        bearSignal = -tProductAll(muEq, muLong, muVS) * 100.0;
    }

    return Math.abs(bullSignal) >= Math.abs(bearSignal) ? bullSignal : bearSignal;
}
