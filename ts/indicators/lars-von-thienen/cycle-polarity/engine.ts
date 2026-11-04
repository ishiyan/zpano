import * as hodrickprescottsmoothinganddetrending from '../hodrick-prescott-smoothing-and-detrending/solver.js';
import * as harmoniccycle from '../harmonic-cycle/model.js';
import { CyclePolarityParams } from "./params.js";
import { validateInteger, validateReal, finite, validateSample, throwError, isInfinite, mod, push } from '../harmonic-cycle/validation.js';
// Native numerical port of the Go implementation; Lars von Thienen, CC BY 4.0.
// Research and startup conventions: description.md.
export class Snapshot {
    polarity: number = 0;
    shortValue: number = 0;
    longValue: number = 0;
    shortSlope: number = 0;
    longSlope: number = 0;
    constructor(values: Partial<Snapshot> = {}) { Object.assign(this, values); }
}
export function emptySnapshot(): Snapshot {
    let nan = NaN;
    return new Snapshot({ polarity: nan, shortValue: nan, longValue: nan, shortSlope: nan, longSlope: nan });
}
export class Engine {
    p: CyclePolarityParams = undefined as unknown as CyclePolarityParams;
    prices: number[] = [];
    constructor(values: Partial<Engine> = {}) { Object.assign(this, values); }
    reset() {
        this.prices = [];
    }
    primed(): boolean {
        return (this.prices.length === this.p.window);
    }
    warmup(): number {
        return (this.p.window - 1);
    }
    update(price: number): Snapshot {
        harmoniccycle.validateSample(price);
        if (Number.isNaN(price)) {
            return emptySnapshot();
        }
        let prices = harmoniccycle.push([...[], ...this.prices], price, this.p.window);
        let output = emptySnapshot();
        if ((prices.length === this.p.window)) {
            let values = hodrickprescottsmoothinganddetrending.cycleData(prices, this.p.smoothing);
            let short = harmoniccycle.goertzel(values, this.p.shortPeriod);
            let long = harmoniccycle.goertzel(values, this.p.longPeriod);
            let shortSlope = harmoniccycle.cycleSlope(short, (this.p.window - 1));
            let longSlope = harmoniccycle.cycleSlope(long, (this.p.window - 1));
            output = new Snapshot({ polarity: harmoniccycle.cyclePolarity(shortSlope, longSlope, this.p.tolerance), shortValue: harmoniccycle.cycleValue(short, (this.p.window - 1)), longValue: harmoniccycle.cycleValue(long, (this.p.window - 1)), shortSlope: shortSlope, longSlope: longSlope });
        }
        this.prices = prices;
        return output;
    }
}
