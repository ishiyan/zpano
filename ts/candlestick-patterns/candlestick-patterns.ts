import { CandlestickPatternsEngine, CandlestickPatternsOptions } from './core/engine.ts';
import { PatternIdentifier, PATTERN_COUNT } from './core/pattern-identifier.ts';
import {
    abandonedBaby, advanceBlock, beltHold, breakaway, closingMarubozu,
    concealingBabySwallow, counterattack, darkCloudCover, doji, dojiStar,
    dragonflyDoji, engulfing, eveningDojiStar, eveningStar, gravestoneDoji,
    hammer, hangingMan, harami, haramiCross, highWave,
    hikkake, hikkakeModified, homingPigeon, identicalThreeCrows, inNeck,
    invertedHammer, kicking, kickingByLength, ladderBottom, longLeggedDoji,
    longLine, marubozu, matchingLow, matHold, morningDojiStar,
    morningStar, onNeck, piercing, rickshawMan, risingFallingThreeMethods,
    separatingLines, shootingStar, shortLine, spinningTop, stalled,
    stickSandwich, takuri, tasukiGap, threeBlackCrows, threeInside,
    threeLineStrike, threeOutside, threeStarsInTheSouth, threeWhiteSoldiers,
    thrusting, tristar, twoCrows, uniqueThreeRiver,
    upDownGapSideBySideWhiteLines, upsideGapTwoCrows, xSideGapThreeMethods,
} from './patterns/index.ts';

// Re-export types from core for public API consumers.
export { OHLC, CandlestickPatternsOptions } from './core/engine.ts';

/**
 * CandlestickPatterns is the streaming candlestick pattern recognition engine.
 *
 * Call update(o, h, l, c) for each new bar, then call any pattern method
 * to get the result: a continuous confidence value in [-100, +100], where
 * positive values indicate bullish signals and negative values indicate
 * bearish signals. The magnitude reflects the fuzzy confidence of the match.
 * Use {@link alphaCut} to convert to crisp {-100, 0, +100} if needed.
 */
export class CandlestickPatterns {
    private readonly engine: CandlestickPatternsEngine;

    constructor(opts?: CandlestickPatternsOptions) {
        this.engine = new CandlestickPatternsEngine(opts);
    }

    /** Returns the number of bars fed so far. */
    get count(): number {
        return this.engine.count;
    }

    /** Feeds a new OHLC bar into the engine. */
    update(o: number, h: number, l: number, c: number): void {
        this.engine.updateBar(o, h, l, c);
        // Reset and update stateful patterns.
        this.engine.hikmodConfirmed = false;
        this.engine.hikmodLastSignal = 0;
        this.engine.hikkakeModifiedUpdate();
    }

    /** Evaluates a single pattern by its identifier. */
    evaluate(id: PatternIdentifier): number {
        const fn = patternDispatch[id];
        if (fn === undefined) return 0;
        return fn(this.engine);
    }

    // -----------------------------------------------------------------------
    // Pattern methods — each delegates to the standalone function in patterns/.
    // -----------------------------------------------------------------------

    abandonedBaby(): number { return abandonedBaby(this.engine); }
    advanceBlock(): number { return advanceBlock(this.engine); }
    beltHold(): number { return beltHold(this.engine); }
    breakaway(): number { return breakaway(this.engine); }
    closingMarubozu(): number { return closingMarubozu(this.engine); }
    concealingBabySwallow(): number { return concealingBabySwallow(this.engine); }
    counterattack(): number { return counterattack(this.engine); }
    darkCloudCover(): number { return darkCloudCover(this.engine); }
    doji(): number { return doji(this.engine); }
    dojiStar(): number { return dojiStar(this.engine); }
    dragonflyDoji(): number { return dragonflyDoji(this.engine); }
    engulfing(): number { return engulfing(this.engine); }
    eveningDojiStar(): number { return eveningDojiStar(this.engine); }
    eveningStar(): number { return eveningStar(this.engine); }
    gravestoneDoji(): number { return gravestoneDoji(this.engine); }
    hammer(): number { return hammer(this.engine); }
    hangingMan(): number { return hangingMan(this.engine); }
    harami(): number { return harami(this.engine); }
    haramiCross(): number { return haramiCross(this.engine); }
    highWave(): number { return highWave(this.engine); }
    hikkake(): number { return hikkake(this.engine); }
    hikkakeModified(): number { return hikkakeModified(this.engine); }
    homingPigeon(): number { return homingPigeon(this.engine); }
    identicalThreeCrows(): number { return identicalThreeCrows(this.engine); }
    inNeck(): number { return inNeck(this.engine); }
    invertedHammer(): number { return invertedHammer(this.engine); }
    kicking(): number { return kicking(this.engine); }
    kickingByLength(): number { return kickingByLength(this.engine); }
    ladderBottom(): number { return ladderBottom(this.engine); }
    longLeggedDoji(): number { return longLeggedDoji(this.engine); }
    longLine(): number { return longLine(this.engine); }
    marubozu(): number { return marubozu(this.engine); }
    matchingLow(): number { return matchingLow(this.engine); }
    matHold(): number { return matHold(this.engine); }
    morningDojiStar(): number { return morningDojiStar(this.engine); }
    morningStar(): number { return morningStar(this.engine); }
    onNeck(): number { return onNeck(this.engine); }
    piercing(): number { return piercing(this.engine); }
    rickshawMan(): number { return rickshawMan(this.engine); }
    risingFallingThreeMethods(): number { return risingFallingThreeMethods(this.engine); }
    separatingLines(): number { return separatingLines(this.engine); }
    shootingStar(): number { return shootingStar(this.engine); }
    shortLine(): number { return shortLine(this.engine); }
    spinningTop(): number { return spinningTop(this.engine); }
    stalled(): number { return stalled(this.engine); }
    stickSandwich(): number { return stickSandwich(this.engine); }
    takuri(): number { return takuri(this.engine); }
    tasukiGap(): number { return tasukiGap(this.engine); }
    threeBlackCrows(): number { return threeBlackCrows(this.engine); }
    threeInside(): number { return threeInside(this.engine); }
    threeLineStrike(): number { return threeLineStrike(this.engine); }
    threeOutside(): number { return threeOutside(this.engine); }
    threeStarsInTheSouth(): number { return threeStarsInTheSouth(this.engine); }
    threeWhiteSoldiers(): number { return threeWhiteSoldiers(this.engine); }
    thrusting(): number { return thrusting(this.engine); }
    tristar(): number { return tristar(this.engine); }
    twoCrows(): number { return twoCrows(this.engine); }
    uniqueThreeRiver(): number { return uniqueThreeRiver(this.engine); }
    upDownGapSideBySideWhiteLines(): number { return upDownGapSideBySideWhiteLines(this.engine); }
    upsideGapTwoCrows(): number { return upsideGapTwoCrows(this.engine); }
    xSideGapThreeMethods(): number { return xSideGapThreeMethods(this.engine); }
}

/** Dispatch table: PatternIdentifier → pattern function. */
const patternDispatch: Array<(engine: CandlestickPatternsEngine) => number> = [
    abandonedBaby,
    advanceBlock,
    beltHold,
    breakaway,
    closingMarubozu,
    concealingBabySwallow,
    counterattack,
    darkCloudCover,
    doji,
    dojiStar,
    dragonflyDoji,
    engulfing,
    eveningDojiStar,
    eveningStar,
    gravestoneDoji,
    hammer,
    hangingMan,
    harami,
    haramiCross,
    highWave,
    hikkake,
    hikkakeModified,
    homingPigeon,
    identicalThreeCrows,
    inNeck,
    invertedHammer,
    kicking,
    kickingByLength,
    ladderBottom,
    longLeggedDoji,
    longLine,
    marubozu,
    matchingLow,
    matHold,
    morningDojiStar,
    morningStar,
    onNeck,
    piercing,
    rickshawMan,
    risingFallingThreeMethods,
    separatingLines,
    shootingStar,
    shortLine,
    spinningTop,
    stalled,
    stickSandwich,
    takuri,
    tasukiGap,
    threeBlackCrows,
    threeInside,
    threeLineStrike,
    threeOutside,
    threeStarsInTheSouth,
    threeWhiteSoldiers,
    thrusting,
    tristar,
    twoCrows,
    uniqueThreeRiver,
    upDownGapSideBySideWhiteLines,
    upsideGapTwoCrows,
    xSideGapThreeMethods,
];
