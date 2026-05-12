/** Tasuki Gap pattern (3-candle continuation). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, isWhite, isBlack, isRealBodyGapUp, isRealBodyGapDown } from '../core/primitives.ts';

/**
 * Tasuki Gap: a three-candle continuation pattern.
 *
 * Upside Tasuki Gap (bullish):
 * - real-body gap up between 1st and 2nd candles,
 * - 2nd candle: white,
 * - 3rd candle: black, opens within 2nd white body, closes below 2nd
 *   open but above 1st candle's real body top (inside the gap),
 * - 2nd and 3rd have near-equal body sizes.
 *
 * Downside Tasuki Gap (bearish):
 * - real-body gap down between 1st and 2nd candles,
 * - 2nd candle: black,
 * - 3rd candle: white, opens within 2nd black body, closes above 2nd
 *   open but below 1st candle's real body bottom (inside the gap),
 * - 2nd and 3rd have near-equal body sizes.
 *
 * Category C: both branches evaluated, return stronger signal.
 *
 * Returns:
 *     Continuous float in [-100, +100].
 */
export function tasukiGap(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(3, cp.near)) return 0.0;

    const b1 = cp.bar(3);
    const b2 = cp.bar(2);
    const b3 = cp.bar(1);

    // Upside Tasuki Gap (bullish).
    let bullSignal = 0.0;
    if (isRealBodyGapUp(b1.o, b1.c, b2.o, b2.c) &&
        isWhite(b2.o, b2.c) && isBlack(b3.o, b3.c)) {
        const rb2 = realBody(b2.o, b2.c);
        const rb3 = realBody(b3.o, b3.c);
        let width = 0.0;
        if (rb2 > 0.0) width = cp.fuzzRatio * rb2;
        // o3 within 2nd body: o3 < c2 and o3 > o2
        const muO3LtC2 = cp.muLtRaw(b3.o, b2.c, width);
        const muO3GtO2 = cp.muGtRaw(b3.o, b2.o, width);
        // c3 below o2
        const muC3LtO2 = cp.muLtRaw(b3.c, b2.o, width);
        // c3 above 1st body top (inside gap)
        const body1Top = Math.max(b1.c, b1.o);
        const muC3GtTop1 = cp.muGtRaw(b3.c, body1Top, width);
        // near-equal bodies
        const muNear = cp.muLessCS(Math.abs(rb2 - rb3), cp.near, 2);
        bullSignal = tProductAll(muO3LtC2, muO3GtO2, muC3LtO2, muC3GtTop1, muNear) * 100.0;
    }

    // Downside Tasuki Gap (bearish).
    let bearSignal = 0.0;
    if (isRealBodyGapDown(b1.o, b1.c, b2.o, b2.c) &&
        isBlack(b2.o, b2.c) && isWhite(b3.o, b3.c)) {
        const rb2 = realBody(b2.o, b2.c);
        const rb3 = realBody(b3.o, b3.c);
        let width = 0.0;
        if (rb2 > 0.0) width = cp.fuzzRatio * rb2;
        // o3 within 2nd body: o3 < o2 and o3 > c2
        const muO3LtO2 = cp.muLtRaw(b3.o, b2.o, width);
        const muO3GtC2 = cp.muGtRaw(b3.o, b2.c, width);
        // c3 above o2
        const muC3GtO2 = cp.muGtRaw(b3.c, b2.o, width);
        // c3 below 1st body bottom (inside gap)
        const body1Bot = Math.min(b1.c, b1.o);
        const muC3LtBot1 = cp.muLtRaw(b3.c, body1Bot, width);
        // near-equal bodies
        const muNear = cp.muLessCS(Math.abs(rb2 - rb3), cp.near, 2);
        bearSignal = -tProductAll(muO3LtO2, muO3GtC2, muC3GtO2, muC3LtBot1, muNear) * 100.0;
    }

    return Math.abs(bullSignal) >= Math.abs(bearSignal) ? bullSignal : bearSignal;
}
