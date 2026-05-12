/** Kicking By Length pattern (2-candle). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, upperShadow, lowerShadow, isHighLowGapUp, isHighLowGapDown } from '../core/primitives.ts';

/**
 * Kicking By Length: like Kicking but direction determined by the longer marubozu.
 *
 * Must have:
 * - first candle: marubozu (long body, very short shadows),
 * - second candle: opposite-color marubozu with a high-low gap,
 * - bull/bear determined by which marubozu has the longer real body.
 *
 * Category B: direction from longer marubozu's color.
 *
 * Returns:
 *     Continuous float in [-100, +100].
 */
export function kickingByLength(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(2, cp.veryShortShadow, cp.longBody)) return 0.0;

    const b1 = cp.bar(2);
    const b2 = cp.bar(1);

    const color1 = b1.c < b1.o ? -1 : 1;
    const color2 = b2.c < b2.o ? -1 : 1;
    // Crisp: opposite colors.
    if (color1 === color2) return 0.0;

    // Crisp: gap check.
    let hasGap = false;
    if (color1 === -1 && isHighLowGapUp(b1.h, b2.l)) hasGap = true;
    else if (color1 === 1 && isHighLowGapDown(b1.l, b2.h)) hasGap = true;
    if (!hasGap) return 0.0;

    const rb1 = realBody(b1.o, b1.c);
    const rb2 = realBody(b2.o, b2.c);

    // Fuzzy: both are marubozu (long body, very short shadows).
    const muLong1 = cp.muGreaterCS(rb1, cp.longBody, 2);
    const muVSUS1 = cp.muLessCS(upperShadow(b1.o, b1.h, b1.c), cp.veryShortShadow, 2);
    const muVSLS1 = cp.muLessCS(lowerShadow(b1.o, b1.l, b1.c), cp.veryShortShadow, 2);

    const muLong2 = cp.muGreaterCS(rb2, cp.longBody, 1);
    const muVSUS2 = cp.muLessCS(upperShadow(b2.o, b2.h, b2.c), cp.veryShortShadow, 1);
    const muVSLS2 = cp.muLessCS(lowerShadow(b2.o, b2.l, b2.c), cp.veryShortShadow, 1);

    const confidence = tProductAll(muLong1, muVSUS1, muVSLS1, muLong2, muVSUS2, muVSLS2);

    // Direction determined by the longer marubozu's color.
    let direction = color1;
    if (rb2 > rb1) direction = color2;
    return direction * confidence * 100.0;
}
