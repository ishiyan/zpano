import * as hodrickprescottsmoothinganddetrending from '../hodrick-prescott-smoothing-and-detrending/solver.js';
import * as harmoniccycle from '../harmonic-cycle/model.js';
import { CompositeCycleParams } from "./params.js";
import { validateInteger, validateReal, finite, validateSample, throwError, isInfinite, mod, push } from '../harmonic-cycle/validation.js';
// Native numerical port of the Go implementation; Lars von Thienen, CC BY 4.0.
// Research and startup conventions: description.md.
export class Snapshot {
    value: number = 0;
    slope: number = 0;
    nextValue: number = 0;
    constructor(values: Partial<Snapshot> = {}) { Object.assign(this, values); }
}
export function emptySnapshot(): Snapshot {
    let nan = NaN;
    return new Snapshot({ value: nan, slope: nan, nextValue: nan });
}
export class Engine {
    p: CompositeCycleParams = undefined as unknown as CompositeCycleParams;
    prices: number[] = [];
    models: harmoniccycle.CycleModel[] = [];
    constructor(values: Partial<Engine> = {}) { Object.assign(this, values); }
    reset() {
        this.prices = [];
        this.models = [];
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
            let models = Array.from({ length: this.p.periods.length }, () => new harmoniccycle.CycleModel());
            let sum: number = 0;
            for (const [i, period] of this.p.periods.entries()) {
                models[i] = harmoniccycle.goertzel(values, period);
                sum += harmoniccycle.cycleSlope(models[i], (this.p.window - 1));
            }
            output = new Snapshot({ value: harmoniccycle.compositeValue(models, (this.p.window - 1)), slope: harmoniccycle.finite(sum), nextValue: harmoniccycle.compositeValue(models, this.p.window) });
            this.models = models;
        }
        this.prices = prices;
        return output;
    }
}
