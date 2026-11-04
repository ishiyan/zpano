import * as harmoniccycle from '../harmonic-cycle/model.js';
import { ZLEMA, newZLEMA } from './smoother.js';
import { LarsVonThienenZeroLagExponentialMovingAverageParams } from "./params.js";
import { validateInteger, validateReal, finite, validateSample, throwError, isInfinite, mod, push } from '../harmonic-cycle/validation.js';
// Native numerical port of the Go implementation; Lars von Thienen, CC BY 4.0.
// Research and startup conventions: description.md.
export class Snapshot {
    value: number = 0;
    constructor(values: Partial<Snapshot> = {}) { Object.assign(this, values); }
}
export function emptySnapshot(): Snapshot {
    let nan = NaN;
    return new Snapshot({ value: nan });
}
export class Engine {
    p: LarsVonThienenZeroLagExponentialMovingAverageParams = undefined as unknown as LarsVonThienenZeroLagExponentialMovingAverageParams;
    smoother: ZLEMA = undefined as unknown as ZLEMA;
    samplesSeen: number = 0;
    constructor(values: Partial<Engine> = {}) { Object.assign(this, values); }
    reset() {
        this.smoother = newZLEMA(this.p.length);
        this.samplesSeen = 0;
    }
    primed(): boolean {
        return (this.samplesSeen >= (Math.trunc(((this.p.length - 1)) / 2) + 1));
    }
    warmup(): number {
        return Math.trunc(((this.p.length - 1)) / 2);
    }
    update(price: number): Snapshot {
        harmoniccycle.validateSample(price);
        if (Number.isNaN(price)) {
            return emptySnapshot();
        }
        this.samplesSeen++;
        return new Snapshot({ value: this.smoother.update(price) });
    }
}
