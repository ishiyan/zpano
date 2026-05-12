/** Three-Line Strike pattern (4-candle). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { isWhite } from '../core/primitives.ts';

/**
 * Three-Line Strike: a four-candle pattern.
 *
 * Bullish: three white candles with rising closes, each opening within/near
 * the prior body, 4th black opens above 3rd close and closes below 1st open.
 *
 * Bearish: three black candles with falling closes, each opening within/near
 * the prior body, 4th white opens below 3rd close and closes above 1st open.
 *
 * Category C: both branches evaluated, return stronger signal.
 *
 * Returns:
 *     Continuous float in [-100, +100].
 */
export function threeLineStrike(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(4, cp.near)) return 0.0;

    const b1 = cp.bar(4);
    const b2 = cp.bar(3);
    const b3 = cp.bar(2);
    const b4 = cp.bar(1);

    // Three same color — crisp gate.
    const color1 = isWhite(b1.o, b1.c) ? 1 : -1;
    const color2 = isWhite(b2.o, b2.c) ? 1 : -1;
    const color3 = isWhite(b3.o, b3.c) ? 1 : -1;
    const color4 = isWhite(b4.o, b4.c) ? 1 : -1;

    if (!(color1 === color2 && color2 === color3 && color4 === -color3)) return 0.0;

    // 2nd opens within/near 1st real body — fuzzy.
    const near4 = cp.avgCS(cp.near, 4);
    const near3 = cp.avgCS(cp.near, 3);
    let nearWidth4 = 0.0;
    if (near4 > 0.0) nearWidth4 = cp.fuzzRatio * near4;
    let nearWidth3 = 0.0;
    if (near3 > 0.0) nearWidth3 = cp.fuzzRatio * near3;

    const muO2Ge = cp.muGeRaw(b2.o, Math.min(b1.o, b1.c) - near4, nearWidth4);
    const muO2Le = cp.muLtRaw(b2.o, Math.max(b1.o, b1.c) + near4, nearWidth4);
    // 3rd opens within/near 2nd real body — fuzzy.
    const muO3Ge = cp.muGeRaw(b3.o, Math.min(b2.o, b2.c) - near3, nearWidth3);
    const muO3Le = cp.muLtRaw(b3.o, Math.max(b2.o, b2.c) + near3, nearWidth3);

    // Bullish: three white, rising closes, 4th opens above 3rd close, closes below 1st open.
    let bullSignal = 0.0;
    if (color3 === 1 && b3.c > b2.c && b2.c > b1.c) {
        const rb1 = Math.abs(b1.c - b1.o);
        let width = 0.0;
        if (rb1 > 0.0) width = cp.fuzzRatio * rb1;
        const muO4Above = cp.muGtRaw(b4.o, b3.c, width);
        const muC4Below = cp.muLtRaw(b4.c, b1.o, width);
        bullSignal = tProductAll(muO2Ge, muO2Le, muO3Ge, muO3Le, muO4Above, muC4Below) * 100.0;
    }

    // Bearish: three black, falling closes, 4th opens below 3rd close, closes above 1st open.
    let bearSignal = 0.0;
    if (color3 === -1 && b3.c < b2.c && b2.c < b1.c) {
        const rb1 = Math.abs(b1.c - b1.o);
        let width = 0.0;
        if (rb1 > 0.0) width = cp.fuzzRatio * rb1;
        const muO4Below = cp.muLtRaw(b4.o, b3.c, width);
        const muC4Above = cp.muGtRaw(b4.c, b1.o, width);
        bearSignal = -tProductAll(muO2Ge, muO2Le, muO3Ge, muO3Le, muO4Below, muC4Above) * 100.0;
    }

    return Math.abs(bullSignal) >= Math.abs(bearSignal) ? bullSignal : bearSignal;
}
