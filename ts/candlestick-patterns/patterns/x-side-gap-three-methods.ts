/** Up/Down-side Gap Three Methods pattern (3-candle continuation). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, isWhite, isBlack, isRealBodyGapUp, isRealBodyGapDown } from '../core/primitives.ts';

/**
 * Up/Down-side Gap Three Methods: a three-candle continuation pattern.
 *
 * Must have:
 * - first and second candles are the same color with a gap between them,
 * - third candle is opposite color, opens within the second candle's
 *   real body and closes within the first candle's real body (fills the
 *   gap).
 *
 * Upside gap: two white candles with gap up, third is black = bullish.
 * Downside gap: two black candles with gap down, third is white = bearish.
 *
 * Category C: both branches evaluated, return stronger signal.
 *
 * Returns:
 *     Continuous float in [-100, +100].
 */
export function xSideGapThreeMethods(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(3)) return 0.0;

    const b1 = cp.bar(3);
    const b2 = cp.bar(2);
    const b3 = cp.bar(1);

    // Upside gap: two whites gap up, third black fills.
    let bullSignal = 0.0;
    if (isWhite(b1.o, b1.c) && isWhite(b2.o, b2.c) && isBlack(b3.o, b3.c) &&
        isRealBodyGapUp(b1.o, b1.c, b2.o, b2.c)) {
        const rb2 = realBody(b2.o, b2.c);
        let width = 0.0;
        if (rb2 > 0.0) width = cp.fuzzRatio * rb2;
        // o3 within 2nd body: o3 < c2 and o3 > o2
        const muO3LtC2 = cp.muLtRaw(b3.o, b2.c, width);
        const muO3GtO2 = cp.muGtRaw(b3.o, b2.o, width);
        // c3 within 1st body: c3 > o1 and c3 < c1
        const rb1 = realBody(b1.o, b1.c);
        let width1 = 0.0;
        if (rb1 > 0.0) width1 = cp.fuzzRatio * rb1;
        const muC3GtO1 = cp.muGtRaw(b3.c, b1.o, width1);
        const muC3LtC1 = cp.muLtRaw(b3.c, b1.c, width1);
        bullSignal = tProductAll(muO3LtC2, muO3GtO2, muC3GtO1, muC3LtC1) * 100.0;
    }

    // Downside gap: two blacks gap down, third white fills.
    let bearSignal = 0.0;
    if (isBlack(b1.o, b1.c) && isBlack(b2.o, b2.c) && isWhite(b3.o, b3.c) &&
        isRealBodyGapDown(b1.o, b1.c, b2.o, b2.c)) {
        const rb2 = realBody(b2.o, b2.c);
        let width = 0.0;
        if (rb2 > 0.0) width = cp.fuzzRatio * rb2;
        // o3 within 2nd body: o3 > c2 and o3 < o2
        const muO3GtC2 = cp.muGtRaw(b3.o, b2.c, width);
        const muO3LtO2 = cp.muLtRaw(b3.o, b2.o, width);
        // c3 within 1st body: c3 < o1 and c3 > c1
        const rb1 = realBody(b1.o, b1.c);
        let width1 = 0.0;
        if (rb1 > 0.0) width1 = cp.fuzzRatio * rb1;
        const muC3LtO1 = cp.muLtRaw(b3.c, b1.o, width1);
        const muC3GtC1 = cp.muGtRaw(b3.c, b1.c, width1);
        bearSignal = -tProductAll(muO3GtC2, muO3LtO2, muC3LtO1, muC3GtC1) * 100.0;
    }

    return Math.abs(bullSignal) >= Math.abs(bearSignal) ? bullSignal : bearSignal;
}
