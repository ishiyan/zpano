import * as harmoniccycle from '../harmonic-cycle/model.js';
import * as cyclescanner from '../cycle-scanner/scanner.js';
import { CyclePhaseScoreParams } from "./params.js";
import { validateInteger, validateReal, finite, validateSample, throwError, isInfinite, mod, push } from '../harmonic-cycle/validation.js';
// Native numerical port of the Go implementation; Lars von Thienen, CC BY 4.0.
// Research and startup conventions: description.md.
export class Snapshot {
    value: number = 0;
    period: number = 0;
    phase: number = 0;
    constructor(values: Partial<Snapshot> = {}) { Object.assign(this, values); }
}
export function emptySnapshot(): Snapshot {
    let nan = NaN;
    return new Snapshot({ value: nan, period: nan, phase: nan });
}
export class Engine {
    p: CyclePhaseScoreParams = undefined as unknown as CyclePhaseScoreParams;
    scanner: cyclescanner.RollingScanner = undefined as unknown as cyclescanner.RollingScanner;
    constructor(values: Partial<Engine> = {}) { Object.assign(this, values); }
    reset() {
        [this.scanner,] = cyclescanner.newRollingScanner(new cyclescanner.ScanConfig({ window: this.p.window, minPeriod: this.p.minPeriod, maxPeriod: this.p.maxPeriod, smoothing: this.p.smoothing, genuineThreshold: this.p.genuineThreshold, phaseCycles: this.p.phaseCycles, topK: this.p.topK, peakDistance: this.p.peakDistance, maxSegments: this.p.maxSegments }));
    }
    primed(): boolean {
        return this.scanner.isPrimed();
    }
    warmup(): number {
        return (this.p.window - 1);
    }
    update(price: number): Snapshot {
        harmoniccycle.validateSample(price);
        if (Number.isNaN(price)) {
            return emptySnapshot();
        }
        let scan = this.scanner.update(price);
        let cycle = scan.cycles[0];
        return new Snapshot({ value: harmoniccycle.phaseState(cycle.phase), period: cycle.period, phase: cycle.phase });
    }
}
