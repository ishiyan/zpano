import * as adaptivebands from '../adaptive-bands/bands.js';
import * as hodrickprescottsmoothinganddetrending from '../hodrick-prescott-smoothing-and-detrending/solver.js';
import * as harmoniccycle from '../harmonic-cycle/model.js';
import { CycleSwingMomentumParams } from "./params.js";
import { validateInteger, validateReal, finite, validateSample, throwError, isInfinite, mod, push } from '../harmonic-cycle/validation.js';
// Native numerical port of the Go implementation; Lars von Thienen, CC BY 4.0.
// Research and startup conventions: description.md.
export class Snapshot {
    value: number = 0;
    lower: number = 0;
    upper: number = 0;
    score: number = 0;
    constructor(values: Partial<Snapshot> = {}) { Object.assign(this, values); }
}
export function emptySnapshot(): Snapshot {
    let nan = NaN;
    return new Snapshot({ value: nan, lower: nan, upper: nan, score: nan });
}
export class Engine {
    p: CycleSwingMomentumParams = undefined as unknown as CycleSwingMomentumParams;
    prices: number[] = [];
    values: number[] = [];
    constructor(values: Partial<Engine> = {}) { Object.assign(this, values); }
    reset() {
        this.prices = [];
        this.values = [];
    }
    primed(): boolean {
        return (this.values.length === this.p.bandMemory);
    }
    warmup(): number {
        return ((this.p.window + this.p.bandMemory) - 2);
    }
    update(price: number): Snapshot {
        harmoniccycle.validateSample(price);
        if (Number.isNaN(price)) {
            return emptySnapshot();
        }
        let prices = harmoniccycle.push([...[], ...this.prices], price, this.p.window);
        let output = emptySnapshot();
        if ((prices.length === this.p.window)) {
            let fast = hodrickprescottsmoothinganddetrending.hpTrend(prices, this.p.fastSmoothing);
            let slow = hodrickprescottsmoothinganddetrending.hpTrend(prices, this.p.slowSmoothing);
            let value = harmoniccycle.finite((fast[(fast.length - 1)] - slow[(slow.length - 1)]));
            this.values = harmoniccycle.push(this.values, value, this.p.bandMemory);
            output.value = value;
            if (this.primed()) {
                [output.lower, output.upper] = adaptivebands.gridBands(this.values, this.p.leveling);
                output.score = 0;
                if ((value >= output.upper)) {
                    output.score = 1;
                }
                else if ((value <= output.lower)) {
                    output.score = -1;
                }
            }
        }
        this.prices = prices;
        return output;
    }
}
