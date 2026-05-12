/** Harami Cross pattern (2-candle reversal). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody } from '../core/primitives.ts';

/**
 * Harami Cross: a two-candle reversal pattern.
 *
 * Like Harami, but the second candle is a doji instead of just short.
 *
 * Must have:
 * - first candle: long real body,
 * - second candle: doji body contained within the first candle's real body.
 *
 * Category B: direction from 1st candle color (continuous).
 *
 * Returns:
 *     Continuous number in [-100, +100].
 */
export function haramiCross(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(2, cp.longBody, cp.dojiBody)) return 0.0;

    const b1 = cp.bar(2);
    const b2 = cp.bar(1);

    // Fuzzy size conditions.
    const muLong1 = cp.muGreaterCS(realBody(b1.o, b1.c), cp.longBody, 2);
    const muDoji2 = cp.muLessCS(realBody(b2.o, b2.c), cp.dojiBody, 1);

    // Fuzzy containment: 1st body encloses 2nd body.
    const eqAvg = cp.avgCS(cp.equal, 1);
    let eqWidth = cp.fuzzRatio * eqAvg;
    if (eqAvg <= 0.0) eqWidth = 0.0;

    const muEncUpper = cp.muGeRaw(Math.max(b1.o, b1.c), Math.max(b2.o, b2.c), eqWidth);
    const muEncLower = cp.muLtRaw(Math.min(b1.o, b1.c), Math.min(b2.o, b2.c), eqWidth);

    const confidence = tProductAll(muLong1, muDoji2, muEncUpper, muEncLower);
    // Direction: opposite of 1st candle color.
    const direction = b1.c < b1.o ? 1.0 : -1.0;
    return direction * confidence * 100.0;
}
