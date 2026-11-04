import { validateInteger, validateReal, finite, validateSample, throwError, isInfinite, mod, push } from '../harmonic-cycle/validation.js';
import * as bartelsperiodicityscore from '../bartels-periodicity-score/scorer.js';
import * as hodrickprescottsmoothinganddetrending from '../hodrick-prescott-smoothing-and-detrending/solver.js';
import * as harmoniccycle from '../harmonic-cycle/model.js';
// Native numerical port of the Go implementation; Lars von Thienen, CC BY 4.0.
// Research and startup conventions: description.md.
export class ScanConfig {
    window: number = 0;
    minPeriod: number = 0;
    maxPeriod: number = 0;
    smoothing: number = 0;
    genuineThreshold: number = 0;
    phaseCycles: number = 0;
    topK: number = 0;
    peakDistance: number = 0;
    maxSegments: number = 0;
    constructor(values: Partial<ScanConfig> = {}) { Object.assign(this, values); }
    validate(): Error | null {
        if (((((((this.window < 6) || (this.minPeriod < 3)) || (this.maxPeriod < this.minPeriod)) || (this.topK < 1)) || (this.peakDistance < 1)) || (this.maxSegments < 2))) {
            return new Error("invalid scanner integer parameters");
        }
        for (const [name, value] of new Map<string, number>([["smoothing", this.smoothing], ["genuine_threshold", this.genuineThreshold], ["phase_cycles", this.phaseCycles]]).entries()) {
            {
                let err = harmoniccycle.validateReal(value, name, 0);
                if ((err !== null)) {
                    return err;
                }
            }
        }
        if (((this.genuineThreshold > 100) || ((this.phaseCycles !== 0) && (this.phaseCycles < 1)))) {
            return new Error("invalid scanner threshold or phase cycles");
        }
        return null;
    }
}
export function defaultConfig(): ScanConfig {
    return new ScanConfig({ window: 200, minPeriod: 5, maxPeriod: 80, smoothing: 1600, genuineThreshold: 49, phaseCycles: 2, topK: 3, peakDistance: 2, maxSegments: 5 });
}
export class RankedCycle {
    period: number = 0;
    amplitude: number = 0;
    phase: number = 0;
    strength: number = 0;
    genuinePercent: number = 0;
    segments: number = 0;
    fitAmplitude: number = 0;
    fitSamples: number = 0;
    value: number = 0;
    slope: number = 0;
    etaTop: number = 0;
    etaBottom: number = 0;
    constructor(values: Partial<RankedCycle> = {}) { Object.assign(this, values); }
}
export class ScanOutput {
    amplitude: number[] = [];
    genuinePercent: number[] = [];
    cycles: RankedCycle[] = [];
    acceptedCount: number = 0;
    constructor(values: Partial<ScanOutput> = {}) { Object.assign(this, values); }
}
export function emptyCycle(): RankedCycle {
    let nan = NaN;
    return new RankedCycle({ period: nan, amplitude: nan, phase: nan, strength: nan, genuinePercent: nan, segments: nan, fitAmplitude: nan, fitSamples: nan, value: nan, slope: nan, etaTop: nan, etaBottom: nan });
}
export function emptyScan(config: ScanConfig): ScanOutput {
    let count = ((config.maxPeriod - config.minPeriod) + 1);
    let output = new ScanOutput({ amplitude: Array.from({ length: count }, () => 0), genuinePercent: Array.from({ length: count }, () => 0), cycles: Array.from({ length: config.topK }, () => new RankedCycle()), acceptedCount: NaN });
    for (let i = 0; i < count; i++) {
        output.amplitude[i] = NaN;
        output.genuinePercent[i] = NaN;
    }
    for (let i = 0; i < output.cycles.length; i++) {
        output.cycles[i] = emptyCycle();
    }
    return output;
}
export function modeledValue(cycle: RankedCycle, offset: number): number {
    return (cycle.fitAmplitude * Math.sin((cycle.phase + (((2 * Math.PI) * offset) / cycle.period))));
}
export function modulo(value: number): number {
    value = mod(value, (2 * Math.PI));
    if ((value < 0)) {
        value += (2 * Math.PI);
    }
    return value;
}
export function scanCycles(prices: number[], config: ScanConfig): ScanOutput {
    {
        let err = config.validate();
        if ((err !== null)) {
            throwError(err);
        }
    }
    if ((prices.length < 6)) {
        throwError("scan requires >= 6 prices");
    }
    for (const value of prices) {
        harmoniccycle.validateSample(value);
        if (Number.isNaN(value)) {
            throwError("scan batch data must be finite");
        }
    }
    let residual = hodrickprescottsmoothinganddetrending.cycleData(prices, config.smoothing);
    let output = emptyScan(config);
    let count = output.amplitude.length;
    let models = Array.from({ length: count }, () => new harmoniccycle.CycleModel());
    let evidence = Array.from({ length: count }, () => new bartelsperiodicityscore.BartelsOutput());
    for (let i = 0; i < count; i++) {
        let period = (config.minPeriod + i);
        models[i] = harmoniccycle.goertzel(residual, period);
        let segments = Math.min(Math.trunc(prices.length / period), config.maxSegments);
        evidence[i] = bartelsperiodicityscore.bartelsScore(prices.slice((prices.length - (segments * period))), period, config.maxSegments, 2);
        output.amplitude[i] = models[i].amplitude;
        output.genuinePercent[i] = evidence[i].genuinePercent;
    }
    let candidates = [] as number[];
    for (const [i, model] of models.entries()) {
        let [left, right] = [-Infinity, -Infinity];
        if ((i > 0)) {
            left = output.amplitude[(i - 1)];
        }
        if (((i + 1) < count)) {
            right = output.amplitude[(i + 1)];
        }
        let score = evidence[i];
        if ((((((model.amplitude > 0) && (model.amplitude > left)) && (model.amplitude >= right)) && !Number.isNaN(score.genuinePercent)) && (score.genuinePercent > config.genuineThreshold))) {
            candidates = [...candidates, i];
        }
    }
    candidates.sort((i, j) => {
        const less = (i: number, j: number) => {
            let [left, right] = [models[i], models[j]];
            let [a, b] = [(left.amplitude / left.period), (right.amplitude / right.period)];
            if ((a === b)) {
                return (left.period < right.period);
            }
            return (a > b);
        };
        return less(i, j) ? -1 : less(j, i) ? 1 : 0;
    });
    let accepted = [] as RankedCycle[];
    for (const i of candidates) {
        let [model, score] = [models[i], evidence[i]];
        let suppressed = false;
        for (const prior of accepted) {
            if ((Math.abs((model.period - prior.period)) < config.peakDistance)) {
                suppressed = true;
                break;
            }
        }
        if (suppressed) {
            continue;
        }
        let fitN = prices.length;
        if ((config.phaseCycles !== 0)) {
            fitN = Math.min(fitN, Math.max(6, Math.trunc(Math.ceil((config.phaseCycles * model.period)))));
        }
        let fit = model;
        if ((fitN !== prices.length)) {
            fit = harmoniccycle.goertzel(hodrickprescottsmoothinganddetrending.cycleData(prices.slice((prices.length - fitN)), config.smoothing), model.period);
        }
        let phase = modulo((fit.phase + (((2 * Math.PI) * (fitN - 1)) / model.period)));
        let omega = ((2 * Math.PI) / model.period);
        let value = (fit.amplitude * Math.sin(phase));
        let slope = ((fit.amplitude * omega) * Math.cos(phase));
        let etaTop = (modulo(((Math.PI / 2) - phase)) / omega);
        let etaBottom = (modulo((((3 * Math.PI) / 2) - phase)) / omega);
        accepted = [...accepted, new RankedCycle({ period: model.period, amplitude: model.amplitude, phase: phase, strength: (model.amplitude / model.period), genuinePercent: score.genuinePercent, segments: score.segments, fitAmplitude: fit.amplitude, fitSamples: fitN, value: value, slope: slope, etaTop: etaTop, etaBottom: etaBottom })];
    }
    output.cycles.splice(0, Math.min(output.cycles.length, accepted.length), ...accepted.slice(0, output.cycles.length));
    output.acceptedCount = accepted.length;
    return output;
}
export class RollingScanner {
    config: ScanConfig = undefined as unknown as ScanConfig;
    prices: number[] = [];
    samplesSeen: number = 0;
    constructor(values: Partial<RollingScanner> = {}) { Object.assign(this, values); }
    reset() {
        this.prices = this.prices.slice(0, 0);
        this.samplesSeen = 0;
    }
    isPrimed(): boolean {
        return (this.prices.length === this.config.window);
    }
    update(price: number): ScanOutput {
        harmoniccycle.validateSample(price);
        if (Number.isNaN(price)) {
            return emptyScan(this.config);
        }
        let prices = harmoniccycle.push([...[] as number[], ...this.prices], price, this.config.window);
        let output = emptyScan(this.config);
        if ((prices.length === this.config.window)) {
            output = scanCycles(prices, this.config);
        }
        this.prices = prices;
        this.samplesSeen++;
        return output;
    }
}
export function newRollingScanner(config: ScanConfig): readonly [
    RollingScanner,
    Error | null
] {
    {
        let err = config.validate();
        if ((err !== null)) {
            return throwError(err);
        }
    }
    return [new RollingScanner({ config: config, prices: [] as number[] }), null] as const;
}
