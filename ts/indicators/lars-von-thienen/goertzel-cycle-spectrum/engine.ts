import * as hodrickprescottsmoothinganddetrending from '../hodrick-prescott-smoothing-and-detrending/solver.js';
import * as harmoniccycle from '../harmonic-cycle/model.js';
import { GoertzelCycleSpectrumParams } from "./params.js";
import { validateInteger, validateReal, finite, validateSample, throwError, isInfinite, mod, push } from '../harmonic-cycle/validation.js';
// Native numerical port of the Go implementation; Lars von Thienen, CC BY 4.0.
// Research and startup conventions: description.md.
export class Snapshot {
    amplitude: number[] = [];
    phases: number[] = [];
    troughOffsets: number[] = [];
    dominantPeriod: number = 0;
    dominantAmplitude: number = 0;
    dominantPhase: number = 0;
    constructor(values: Partial<Snapshot> = {}) { Object.assign(this, values); }
}
export function emptySnapshot(): Snapshot {
    let nan = NaN;
    return new Snapshot({ amplitude: [], phases: [], troughOffsets: [], dominantPeriod: nan, dominantAmplitude: nan, dominantPhase: nan });
}
export class Engine {
    p: GoertzelCycleSpectrumParams = undefined as unknown as GoertzelCycleSpectrumParams;
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
            let count = (Math.trunc(((this.p.maxPeriod - this.p.minPeriod)) / this.p.step) + 1);
            output.amplitude = Array.from({ length: count }, () => 0);
            output.phases = Array.from({ length: count }, () => 0);
            output.troughOffsets = Array.from({ length: count }, () => 0);
            let best = 0;
            for (let i = 0; i < count; i++) {
                let model = harmoniccycle.goertzel(values, (this.p.minPeriod + (i * this.p.step)));
                output.amplitude[i] = model.amplitude;
                output.phases[i] = model.phase;
                output.troughOffsets[i] = model.troughOffset;
                if ((model.amplitude > output.amplitude[best])) {
                    best = i;
                }
            }
            output.dominantAmplitude = output.amplitude[best];
            output.dominantPhase = output.phases[best];
            if ((output.dominantAmplitude > 0)) {
                output.dominantPeriod = (this.p.minPeriod + (best * this.p.step));
            }
        }
        this.prices = prices;
        return output;
    }
}
