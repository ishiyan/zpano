import * as harmoniccycle from '../harmonic-cycle/model.js';
import { hpTrend } from './solver.js';
import { HodrickPrescottSmoothingAndDetrendingParams } from "./params.js";
import { validateInteger, validateReal, finite, validateSample, throwError, isInfinite, mod, push } from '../harmonic-cycle/validation.js';
// Native numerical port of the Go implementation; Lars von Thienen, CC BY 4.0.
// Research and startup conventions: description.md.
export class Snapshot {
    trend: number = 0;
    cycle: number = 0;
    constructor(values: Partial<Snapshot> = {}) { Object.assign(this, values); }
}
export function emptySnapshot(): Snapshot {
    let nan = NaN;
    return new Snapshot({ trend: nan, cycle: nan });
}
export class Engine {
    p: HodrickPrescottSmoothingAndDetrendingParams = undefined as unknown as HodrickPrescottSmoothingAndDetrendingParams;
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
            let trend = hpTrend(prices, this.p.smoothing);
            output = new Snapshot({ trend: trend[(trend.length - 1)], cycle: harmoniccycle.finite((price - trend[(trend.length - 1)])) });
        }
        this.prices = prices;
        return output;
    }
}
