import * as harmoniccycle from '../harmonic-cycle/model.js';
import * as cyclescanner from '../cycle-scanner/scanner.js';
import { PhasingOscillatorParams } from "./params.js";
import { validateInteger, validateReal, finite, validateSample, throwError, isInfinite, mod, push } from '../harmonic-cycle/validation.js';
// Native numerical port of the Go implementation; Lars von Thienen, CC BY 4.0.
// Research and startup conventions: description.md.
export class Snapshot {
    value: number = 0;
    shortScore: number = 0;
    longScore: number = 0;
    shortPeriod: number = 0;
    longPeriod: number = 0;
    constructor(values: Partial<Snapshot> = {}) { Object.assign(this, values); }
}
export function emptySnapshot(): Snapshot {
    let nan = NaN;
    return new Snapshot({ value: nan, shortScore: nan, longScore: nan, shortPeriod: nan, longPeriod: nan });
}
export class Engine {
    p: PhasingOscillatorParams = undefined as unknown as PhasingOscillatorParams;
    short: cyclescanner.RollingScanner = undefined as unknown as cyclescanner.RollingScanner;
    long: cyclescanner.RollingScanner = undefined as unknown as cyclescanner.RollingScanner;
    constructor(values: Partial<Engine> = {}) { Object.assign(this, values); }
    reset() {
        [this.short,] = cyclescanner.newRollingScanner(new cyclescanner.ScanConfig({ window: this.p.window, minPeriod: this.p.shortMin, maxPeriod: this.p.shortMax, smoothing: this.p.smoothing, genuineThreshold: this.p.genuineThreshold, phaseCycles: this.p.phaseCycles, topK: 1, peakDistance: 2, maxSegments: 5 }));
        [this.long,] = cyclescanner.newRollingScanner(new cyclescanner.ScanConfig({ window: this.p.window, minPeriod: this.p.longMin, maxPeriod: this.p.longMax, smoothing: this.p.smoothing, genuineThreshold: this.p.genuineThreshold, phaseCycles: this.p.phaseCycles, topK: 1, peakDistance: 2, maxSegments: 5 }));
    }
    primed(): boolean {
        return (this.short.isPrimed() && this.long.isPrimed());
    }
    warmup(): number {
        return (this.p.window - 1);
    }
    update(price: number): Snapshot {
        harmoniccycle.validateSample(price);
        if (Number.isNaN(price)) {
            return emptySnapshot();
        }
        let short = this.short.update(price).cycles[0];
        let long = this.long.update(price).cycles[0];
        let [shortScore, longScore] = [harmoniccycle.phaseState(short.phase), harmoniccycle.phaseState(long.phase)];
        return new Snapshot({ value: (shortScore + longScore), shortScore: shortScore, longScore: longScore, shortPeriod: short.period, longPeriod: long.period });
    }
}
