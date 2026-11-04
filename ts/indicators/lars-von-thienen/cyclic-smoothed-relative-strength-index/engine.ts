import * as adaptivebands from '../adaptive-bands/bands.js';
import * as harmoniccycle from '../harmonic-cycle/model.js';
import { WilderRSI, newWilderRSI } from './rsi.js';
import { CyclicSmoothedRelativeStrengthIndexParams } from "./params.js";
import { validateInteger, validateReal, finite, validateSample, throwError, isInfinite, mod, push } from '../harmonic-cycle/validation.js';
// Native numerical port of the Go implementation; Lars von Thienen, CC BY 4.0.
// Research and startup conventions: description.md.
export class Snapshot {
    value: number = 0;
    lower: number = 0;
    upper: number = 0;
    constructor(values: Partial<Snapshot> = {}) { Object.assign(this, values); }
}
export function emptySnapshot(): Snapshot {
    let nan = NaN;
    return new Snapshot({ value: nan, lower: nan, upper: nan });
}
export class Engine {
    p: CyclicSmoothedRelativeStrengthIndexParams = undefined as unknown as CyclicSmoothedRelativeStrengthIndexParams;
    rsi: WilderRSI = undefined as unknown as WilderRSI;
    rsiHistory: number[] = [];
    values: number[] = [];
    previousValue: number = 0;
    constructor(values: Partial<Engine> = {}) { Object.assign(this, values); }
    reset() {
        this.rsi = newWilderRSI(Math.trunc(this.p.dominantCycle / 2));
        this.rsiHistory = [];
        this.values = [];
        this.previousValue = 0;
    }
    primed(): boolean {
        return (this.values.length === (2 * this.p.dominantCycle));
    }
    warmup(): number {
        return (((Math.trunc(this.p.dominantCycle / 2) + Math.trunc(((this.p.vibration - 1)) / 2)) + (2 * this.p.dominantCycle)) - 1);
    }
    update(price: number): Snapshot {
        harmoniccycle.validateSample(price);
        if (Number.isNaN(price)) {
            return emptySnapshot();
        }
        let rsi = this.rsi.update(price);
        if (Number.isNaN(rsi)) {
            return emptySnapshot();
        }
        let lag = Math.trunc(((this.p.vibration - 1)) / 2);
        this.rsiHistory = harmoniccycle.push(this.rsiHistory, rsi, (lag + 1));
        if ((this.rsiHistory.length < (lag + 1))) {
            return emptySnapshot();
        }
        let torque = (2 / (this.p.vibration + 1));
        let value = harmoniccycle.finite(((torque * (((2 * rsi) - this.rsiHistory[0]))) + (((1 - torque)) * this.previousValue)));
        this.previousValue = value;
        this.values = harmoniccycle.push(this.values, value, (2 * this.p.dominantCycle));
        let output = emptySnapshot();
        output.value = value;
        if (this.primed()) {
            [output.lower, output.upper] = adaptivebands.gridBands(this.values, this.p.leveling);
        }
        return output;
    }
}
