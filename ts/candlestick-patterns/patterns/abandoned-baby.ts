/** Abandoned Baby pattern (3-candle reversal). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, isWhite, isBlack, isHighLowGapUp, isHighLowGapDown } from '../core/primitives.ts';

const ABANDONED_BABY_PENETRATION = 0.3;

/**
 * Abandoned Baby: a three-candle reversal pattern.
 *
 * Must have:
 * - first candle: long real body,
 * - second candle: doji,
 * - third candle: real body longer than short, opposite color to 1st,
 *   closes well within 1st body,
 * - upside/downside gap between 1st and doji (shadows don't touch),
 * - downside/upside gap between doji and 3rd (shadows don't touch).
 *
 * Category C: both branches evaluated, return stronger signal.
 *
 * Returns:
 *     Continuous float in [-100, +100].
 */
export function abandonedBaby(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(3, cp.longBody, cp.dojiBody, cp.shortBody)) return 0.0;

    const b1 = cp.bar(3);
    const b2 = cp.bar(2);
    const b3 = cp.bar(1);

    // Shared fuzzy conditions: 1st long, 2nd doji, 3rd > short.
    const muLong1 = cp.muGreaterCS(realBody(b1.o, b1.c), cp.longBody, 3);
    const muDoji2 = cp.muLessCS(realBody(b2.o, b2.c), cp.dojiBody, 2);
    const muShort3 = cp.muGreaterCS(realBody(b3.o, b3.c), cp.shortBody, 1);
    const penetration = ABANDONED_BABY_PENETRATION;

    // Bearish: white-doji-black, gap up then gap down.
    let bearSignal = 0.0;
    if (isWhite(b1.o, b1.c) && isBlack(b3.o, b3.c)) {
        if (isHighLowGapUp(b1.h, b2.l) && isHighLowGapDown(b2.l, b3.h)) {
            const rb1 = realBody(b1.o, b1.c);
            const penThreshold = b1.c - rb1 * penetration;
            let penWidth = cp.fuzzRatio * rb1;
            if (rb1 <= 0.0) penWidth = 0.0;
            const muPen = cp.muLtRaw(b3.c, penThreshold, penWidth);
            const confBear = tProductAll(muLong1, muDoji2, muShort3, muPen);
            bearSignal = -confBear * 100.0;
        }
    }

    // Bullish: black-doji-white, gap down then gap up.
    let bullSignal = 0.0;
    if (isBlack(b1.o, b1.c) && isWhite(b3.o, b3.c)) {
        if (isHighLowGapDown(b1.l, b2.h) && isHighLowGapUp(b2.h, b3.l)) {
            const rb1 = realBody(b1.o, b1.c);
            const penThreshold = b1.c + rb1 * penetration;
            let penWidth = cp.fuzzRatio * rb1;
            if (rb1 <= 0.0) penWidth = 0.0;
            const muPen = cp.muGtRaw(b3.c, penThreshold, penWidth);
            const confBull = tProductAll(muLong1, muDoji2, muShort3, muPen);
            bullSignal = confBull * 100.0;
        }
    }

    return Math.abs(bullSignal) >= Math.abs(bearSignal) ? bullSignal : bearSignal;
}
