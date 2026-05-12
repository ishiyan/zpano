import { MembershipShape, muLess, muGreater, muGreaterEqual, muNear, muDirection } from '../../fuzzy/index.ts';
import { Criterion } from './criterion.ts';
import {
    DEFAULT_LONG_BODY, DEFAULT_VERY_LONG_BODY, DEFAULT_SHORT_BODY, DEFAULT_DOJI_BODY,
    DEFAULT_LONG_SHADOW, DEFAULT_VERY_LONG_SHADOW, DEFAULT_SHORT_SHADOW, DEFAULT_VERY_SHORT_SHADOW,
    DEFAULT_NEAR, DEFAULT_FAR, DEFAULT_EQUAL,
} from './defaults.ts';

// Minimum history size: 5-candle patterns + 10 default criterion period + 5 margin.
const MIN_HISTORY = 20;

/** OHLC bar. */
export interface OHLC {
    o: number;
    h: number;
    l: number;
    c: number;
}

/** Options for CandlestickPatterns engine. */
export interface CandlestickPatternsOptions {
    longBody?: Criterion;
    veryLongBody?: Criterion;
    shortBody?: Criterion;
    dojiBody?: Criterion;
    longShadow?: Criterion;
    veryLongShadow?: Criterion;
    shortShadow?: Criterion;
    veryShortShadow?: Criterion;
    near?: Criterion;
    far?: Criterion;
    equal?: Criterion;
    fuzzRatio?: number;
    shape?: MembershipShape;
}

/**
 * CriterionState maintains a running total for a single Criterion over a sliding window.
 */
export class CriterionState {
    readonly criterion: Criterion;
    private readonly ring: number[];
    private readonly ringSize: number;
    private ringStart: number = 0;
    private ringLen: number = 0;
    private total: number = 0;

    constructor(criterion: Criterion, maxShift: number) {
        this.criterion = criterion;
        this.ringSize = criterion.averagePeriod > 0 ? criterion.averagePeriod + maxShift : 0;
        this.ring = new Array(this.ringSize).fill(0);
    }

    push(o: number, h: number, l: number, c: number): void {
        if (this.ringSize === 0) return;
        const val = this.criterion.candleContribution(o, h, l, c);
        if (this.ringLen === this.ringSize) {
            this.total -= this.ring[this.ringStart];
            this.ring[this.ringStart] = val;
            this.ringStart = (this.ringStart + 1) % this.ringSize;
        } else {
            const idx = (this.ringStart + this.ringLen) % this.ringSize;
            this.ring[idx] = val;
            this.ringLen++;
        }
        this.total += val;
    }

    totalAt(shift: number): number {
        if (this.ringSize === 0 || this.criterion.averagePeriod <= 0) return 0.0;
        const period = this.criterion.averagePeriod;
        const n = this.ringLen;
        const end = n - shift;
        const start = end - period;
        if (start < 0 || end <= 0) return 0.0;
        let total = 0.0;
        for (let i = start; i < end; i++) {
            total += this.ring[(this.ringStart + i) % this.ringSize];
        }
        return total;
    }

    avg(shift: number, o: number, h: number, l: number, c: number): number {
        return this.criterion.averageValueFromTotal(this.totalAt(shift), o, h, l, c);
    }
}

/**
 * CandlestickPatternsEngine is the core engine with all internal state
 * and helper methods exposed for use by standalone pattern functions.
 *
 * This class is not intended for direct use by consumers — use the
 * CandlestickPatterns wrapper in the parent module instead.
 */
export class CandlestickPatternsEngine {
    readonly fuzzRatio: number;
    readonly shape: MembershipShape;

    readonly longBody: CriterionState;
    readonly veryLongBody: CriterionState;
    readonly shortBody: CriterionState;
    readonly dojiBody: CriterionState;
    readonly longShadow: CriterionState;
    readonly veryLongShadow: CriterionState;
    readonly shortShadow: CriterionState;
    readonly veryShortShadow: CriterionState;
    readonly near: CriterionState;
    readonly far: CriterionState;
    readonly equal: CriterionState;
    private readonly allStates: CriterionState[];

    private readonly history: OHLC[];
    private readonly histSize: number;
    private histStart: number = 0;
    private histLen: number = 0;
    count: number = 0;

    // Hikkake modified stateful fields.
    hikmodPatternResult: number = 0;
    hikmodPatternIdx: number = 0;
    hikmodConfirmed: boolean = false;
    hikmodLastSignal: number = 0;

    constructor(opts?: CandlestickPatternsOptions) {
        this.fuzzRatio = opts?.fuzzRatio ?? 0.2;
        this.shape = opts?.shape ?? MembershipShape.SIGMOID;

        this.longBody = new CriterionState(opts?.longBody?.copy() ?? DEFAULT_LONG_BODY.copy(), 5);
        this.veryLongBody = new CriterionState(opts?.veryLongBody?.copy() ?? DEFAULT_VERY_LONG_BODY.copy(), 5);
        this.shortBody = new CriterionState(opts?.shortBody?.copy() ?? DEFAULT_SHORT_BODY.copy(), 5);
        this.dojiBody = new CriterionState(opts?.dojiBody?.copy() ?? DEFAULT_DOJI_BODY.copy(), 5);
        this.longShadow = new CriterionState(opts?.longShadow?.copy() ?? DEFAULT_LONG_SHADOW.copy(), 5);
        this.veryLongShadow = new CriterionState(opts?.veryLongShadow?.copy() ?? DEFAULT_VERY_LONG_SHADOW.copy(), 5);
        this.shortShadow = new CriterionState(opts?.shortShadow?.copy() ?? DEFAULT_SHORT_SHADOW.copy(), 5);
        this.veryShortShadow = new CriterionState(opts?.veryShortShadow?.copy() ?? DEFAULT_VERY_SHORT_SHADOW.copy(), 5);
        this.near = new CriterionState(opts?.near?.copy() ?? DEFAULT_NEAR.copy(), 5);
        this.far = new CriterionState(opts?.far?.copy() ?? DEFAULT_FAR.copy(), 5);
        this.equal = new CriterionState(opts?.equal?.copy() ?? DEFAULT_EQUAL.copy(), 5);

        this.allStates = [
            this.longBody, this.veryLongBody, this.shortBody, this.dojiBody,
            this.longShadow, this.veryLongShadow, this.shortShadow, this.veryShortShadow,
            this.near, this.far, this.equal,
        ];

        let maxPeriod = 0;
        for (const s of this.allStates) {
            if (s.criterion.averagePeriod > maxPeriod) {
                maxPeriod = s.criterion.averagePeriod;
            }
        }
        let historySize = maxPeriod + 10;
        if (historySize < MIN_HISTORY) historySize = MIN_HISTORY;
        this.history = new Array(historySize);
        for (let i = 0; i < historySize; i++) {
            this.history[i] = { o: 0, h: 0, l: 0, c: 0 };
        }
        this.histSize = historySize;
    }

    /** Feeds a new OHLC bar into the engine (ring buffer + criterion states). */
    updateBar(o: number, h: number, l: number, c: number): void {
        const bar: OHLC = { o, h, l, c };
        if (this.histLen === this.histSize) {
            this.history[this.histStart] = bar;
            this.histStart = (this.histStart + 1) % this.histSize;
        } else {
            const idx = (this.histStart + this.histLen) % this.histSize;
            this.history[idx] = bar;
            this.histLen++;
        }
        for (const s of this.allStates) {
            s.push(o, h, l, c);
        }
        this.count++;
    }

    /** Gets OHLC of a bar. shift=1 is most recent, shift=2 is one before, etc. */
    bar(shift: number): OHLC {
        const idx = (this.histStart + this.histLen - shift) % this.histSize;
        return this.history[idx];
    }

    /** Checks if we have at least n bars in history. */
    has(n: number): boolean {
        return this.histLen >= n;
    }

    /** Checks if we have sufficient bars for a pattern requiring nCandles. */
    enough(nCandles: number, ...criteria: CriterionState[]): boolean {
        const avail = this.histLen - nCandles;
        for (const cs of criteria) {
            if (avail < cs.criterion.averagePeriod) return false;
        }
        return true;
    }

    /** Gets the criterion average value at a given shift. */
    avgCS(cs: CriterionState, shift: number): number {
        const b = this.bar(shift);
        return cs.avg(shift, b.o, b.h, b.l, b.c);
    }

    // Fuzzy membership helpers
    muLessCS(value: number, cs: CriterionState, shift: number): number {
        const avg = this.avgCS(cs, shift);
        let w = this.fuzzRatio * avg;
        if (avg <= 0.0) w = 0.0;
        return muLess(value, avg, w, this.shape);
    }

    muGreaterCS(value: number, cs: CriterionState, shift: number): number {
        const avg = this.avgCS(cs, shift);
        let w = this.fuzzRatio * avg;
        if (avg <= 0.0) w = 0.0;
        return muGreater(value, avg, w, this.shape);
    }

    muNearValue(value: number, target: number, cs: CriterionState, shift: number): number {
        const avg = this.avgCS(cs, shift);
        let w = this.fuzzRatio * avg;
        if (avg <= 0.0) w = 0.0;
        return muNear(value, target, w, this.shape);
    }

    muGeRaw(value: number, threshold: number, width: number): number {
        return muGreaterEqual(value, threshold, width, this.shape);
    }

    muGtRaw(value: number, threshold: number, width: number): number {
        return muGreater(value, threshold, width, this.shape);
    }

    muLtRaw(value: number, threshold: number, width: number): number {
        return muLess(value, threshold, width, this.shape);
    }

    muBullish(o: number, c: number, shift: number): number {
        const d = this.muDirectionRaw(o, c, shift);
        return d > 0.0 ? d : 0.0;
    }

    muBearish(o: number, c: number, shift: number): number {
        const d = this.muDirectionRaw(o, c, shift);
        return -d > 0.0 ? -d : 0.0;
    }

    muDirectionRaw(o: number, c: number, shift: number): number {
        const avg = this.avgCS(this.shortBody, shift);
        return muDirection(o, c, avg, 2.0);
    }

    /** Stateful update for hikkake_modified pattern. Called from Update(). */
    hikkakeModifiedUpdate(): void {
        if (this.count < 4) return;

        const b1 = this.bar(4);
        const b2 = this.bar(3);
        const b3 = this.bar(2);
        const b4 = this.bar(1);

        if (b2.h < b1.h && b2.l > b1.l &&
            b3.h < b2.h && b3.l > b2.l) {
            const nearAvg = this.avgCS(this.near, 3);
            if (b4.h < b3.h && b4.l < b3.l && b2.c <= b2.l + nearAvg) {
                this.hikmodPatternResult = 100.0;
                this.hikmodPatternIdx = this.count;
                return;
            }
            if (b4.h > b3.h && b4.l > b3.l && b2.c >= b2.h - nearAvg) {
                this.hikmodPatternResult = -100.0;
                this.hikmodPatternIdx = this.count;
                return;
            }
        }

        if (this.hikmodPatternResult !== 0 && this.count <= this.hikmodPatternIdx + 3) {
            const shift3rd = this.count - this.hikmodPatternIdx + 2;
            const b3rd = this.bar(shift3rd);

            if (this.hikmodPatternResult > 0 && b4.c > b3rd.h) {
                this.hikmodLastSignal = 200.0;
                this.hikmodPatternResult = 0.0;
                this.hikmodPatternIdx = 0;
                this.hikmodConfirmed = true;
                return;
            }
            if (this.hikmodPatternResult < 0 && b4.c < b3rd.l) {
                this.hikmodLastSignal = -200.0;
                this.hikmodPatternResult = 0.0;
                this.hikmodPatternIdx = 0;
                this.hikmodConfirmed = true;
                return;
            }
        }

        if (this.hikmodPatternResult !== 0 && this.count > this.hikmodPatternIdx + 3) {
            this.hikmodPatternResult = 0.0;
            this.hikmodPatternIdx = 0;
        }
    }
}
