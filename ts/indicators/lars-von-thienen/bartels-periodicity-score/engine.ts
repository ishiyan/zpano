import * as hodrickprescottsmoothinganddetrending from '../hodrick-prescott-smoothing-and-detrending/solver.js';
import * as harmoniccycle from '../harmonic-cycle/model.js';
import { bartelsScore } from './scorer.js';
import { BartelsPeriodicityScoreParams } from "./params.js";
import { validateInteger, validateReal, finite, validateSample, throwError, isInfinite, mod, push } from '../harmonic-cycle/validation.js';
// Native numerical port of the Go implementation; Lars von Thienen, CC BY 4.0.
// Research and startup conventions: description.md.
export class Snapshot {
    score: number = 0;
    genuinePercent: number = 0;
    segments: number = 0;
    constructor(values: Partial<Snapshot> = {}) { Object.assign(this, values); }
}
export function emptySnapshot(): Snapshot {
    let nan = NaN;
    return new Snapshot({ score: nan, genuinePercent: nan, segments: nan });
}
export class Engine {
    p: BartelsPeriodicityScoreParams = undefined as unknown as BartelsPeriodicityScoreParams;
    prices: number[] = [];
    constructor(values: Partial<Engine> = {}) { Object.assign(this, values); }
    reset() {
        this.prices = [];
    }
    primed(): boolean {
        return (this.prices.length === (this.p.period * this.p.segments));
    }
    warmup(): number {
        return ((this.p.period * this.p.segments) - 1);
    }
    update(price: number): Snapshot {
        harmoniccycle.validateSample(price);
        if (Number.isNaN(price)) {
            return emptySnapshot();
        }
        let prices = harmoniccycle.push([...[], ...this.prices], price, (this.p.period * this.p.segments));
        let output = emptySnapshot();
        if ((prices.length === (this.p.period * this.p.segments))) {
            let score = bartelsScore(hodrickprescottsmoothinganddetrending.cycleData(prices, this.p.smoothing), this.p.period, this.p.segments, this.p.segments);
            output = new Snapshot({ score: score.score, genuinePercent: score.genuinePercent, segments: score.segments });
        }
        this.prices = prices;
        return output;
    }
}
