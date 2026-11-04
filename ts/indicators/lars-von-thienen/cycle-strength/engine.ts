import * as hodrickprescottsmoothinganddetrending from '../hodrick-prescott-smoothing-and-detrending/solver.js';
import * as harmoniccycle from '../harmonic-cycle/model.js';
import { CycleStrengthParams } from "./params.js";
import { validateInteger, validateReal, finite, validateSample, throwError, isInfinite, mod, push } from '../harmonic-cycle/validation.js';
// Native numerical port of the Go implementation; Lars von Thienen, CC BY 4.0.
// Research and startup conventions: description.md.
export class Snapshot {
    strength: number = 0;
    amplitude: number = 0;
    phase: number = 0;
    constructor(values: Partial<Snapshot> = {}) { Object.assign(this, values); }
}
export function emptySnapshot(): Snapshot {
    let nan = NaN;
    return new Snapshot({ strength: nan, amplitude: nan, phase: nan });
}
export class Engine {
    p: CycleStrengthParams = undefined as unknown as CycleStrengthParams;
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
            let model = harmoniccycle.goertzel(values, this.p.period);
            output = new Snapshot({ strength: harmoniccycle.cycleStrength(model.amplitude, this.p.period), amplitude: model.amplitude, phase: model.phase });
        }
        this.prices = prices;
        return output;
    }
}
