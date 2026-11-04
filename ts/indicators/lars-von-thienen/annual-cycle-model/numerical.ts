import * as cyclescanner from '../cycle-scanner/scanner.js';
import * as harmoniccycle from '../harmonic-cycle/model.js';
import * as hodrickprescottsmoothinganddetrending from '../hodrick-prescott-smoothing-and-detrending/solver.js';
import { throwError } from '../harmonic-cycle/validation.js';
// Native numerical port of the Go implementation; Lars von Thienen, CC BY 4.0.
// Research and startup conventions: description.md.
export class Params {
    minPeriod: number = 0;
    maxPeriod: number = 0;
    smoothing: number = 0;
    genuineThreshold: number = 0;
    candidatePool: number = 0;
    maxComponents: number = 0;
    peakDistance: number = 0;
    constructor(values: Partial<Params> = {}) { Object.assign(this, values); }
}
export function defaultParams(): Params {
    return new Params({ minPeriod: 5, maxPeriod: 125, smoothing: 1600, genuineThreshold: 49, candidatePool: 10, maxComponents: 7, peakDistance: 2 });
}
export class Component {
    period: number = 0;
    amplitude: number = 0;
    phase: number = 0;
    strength: number = 0;
    genuinePercent: number = 0;
    constructor(values: Partial<Component> = {}) { Object.assign(this, values); }
}
export class Model {
    year: number = 0;
    sampleCount: number = 0;
    components: Component[] = [];
    correlation: number = 0;
    subsetsChecked: number = 0;
    candidateCount: number = 0;
    constructor(values: Partial<Model> = {}) { Object.assign(this, values); }
}
export function pearson(left: number[], right: number[]): number {
    if (((left.length < 1) || (left.length !== right.length))) {
        throwError("correlation requires nonempty equal-length vectors");
    }
    let [a, b] = [(harmoniccycle.sum(left) / left.length), (harmoniccycle.sum(right) / right.length)];
    let [xx, yy, xy] = [Array.from({ length: left.length }, () => 0), Array.from({ length: left.length }, () => 0), Array.from({ length: left.length }, () => 0)];
    for (let i = 0; i < left.length; i++) {
        let [x, y] = [(left[i] - a), (right[i] - b)];
        xx[i] = (x * x);
        yy[i] = (y * y);
        xy[i] = (x * y);
    }
    let [x, y] = [harmoniccycle.sum(xx), harmoniccycle.sum(yy)];
    if (((x === 0) || (y === 0))) {
        return NaN;
    }
    return Math.max(-1, Math.min(1, (harmoniccycle.sum(xy) / Math.sqrt((x * y)))));
}
export function chooseSubset(target: number[], waves: number[][], periods: number[], maxComponents: number): readonly [
    number[],
    number,
    number,
    Error | null
] {
    if (((((((waves.length < 1) || (waves.length > 16)) || (waves.length !== periods.length)) || (target.length < 6)) || (maxComponents < 1)) || (maxComponents > 7))) {
        return throwError(0);
    }
    let n = target.length;
    let mean = (harmoniccycle.sum(target) / n);
    let tc = Array.from({ length: n }, () => 0);
    let energyValues = Array.from({ length: n }, () => 0);
    for (const [i, value] of target.entries()) {
        tc[i] = (value - mean);
        energyValues[i] = (tc[i] * tc[i]);
    }
    let energy = harmoniccycle.sum(energyValues);
    if ((energy === 0)) {
        return throwError(0);
    }
    let centered = Array.from({ length: waves.length }, () => [] as number[]);
    for (const [i, wave] of waves.entries()) {
        if ((wave.length !== n)) {
            return throwError(0);
        }
        let mean = (harmoniccycle.sum(wave) / n);
        centered[i] = Array.from({ length: n }, () => 0);
        for (const [j, value] of wave.entries()) {
            centered[i][j] = (value - mean);
        }
    }
    let cov = Array.from({ length: waves.length }, () => 0);
    let gram = Array.from({ length: waves.length }, () => [] as number[]);
    for (const [i, left] of centered.entries()) {
        let products = Array.from({ length: n }, () => 0);
        for (let k = 0; k < n; k++) {
            products[k] = (tc[k] * left[k]);
        }
        cov[i] = harmoniccycle.sum(products);
        gram[i] = Array.from({ length: waves.length }, () => 0);
        for (const [j, right] of centered.entries()) {
            for (let k = 0; k < n; k++) {
                products[k] = (left[k] * right[k]);
            }
            gram[i][j] = harmoniccycle.sum(products);
        }
    }
    let best: number[] = [];
    let bestScore = -Infinity;
    let bestPeriods: number[] = [];
    let checked = 0;
    for (let count = 1; (count <= Math.min(maxComponents, waves.length)); count++) {
        for (let mask = 1; (mask < ((1 << waves.length))); mask++) {
            let subset: number[] = [];
            for (let i = 0; (i < waves.length); i++) {
                if (((mask & ((1 << i))) !== 0)) {
                    subset = [...subset, i];
                }
            }
            if ((subset.length !== count)) {
                continue;
            }
            checked++;
            let numerators = Array.from({ length: count }, () => 0);
            let denominators: number[] = [];
            let key = Array.from({ length: count }, () => 0);
            for (const [k, i] of subset.entries()) {
                numerators[k] = cov[i];
                key[k] = periods[i];
                for (const j of subset) {
                    denominators = [...denominators, gram[i][j]];
                }
            }
            let denominator = harmoniccycle.sum(denominators);
            if ((denominator <= 0)) {
                continue;
            }
            let score = Math.max(-1, Math.min(1, (harmoniccycle.sum(numerators) / Math.sqrt((energy * denominator)))));
            key.sort((a, b) => a - b);
            let smaller = (best.length > count);
            if ((best.length === count)) {
                for (let i = 0; i < key.length; i++) {
                    if ((key[i] !== bestPeriods[i])) {
                        smaller = (key[i] < bestPeriods[i]);
                        break;
                    }
                }
            }
            if (((score > (bestScore + 1e-12)) || ((Math.abs((score - bestScore)) <= 1e-12) && smaller))) {
                best = [...[], ...subset];
                bestScore = score;
                bestPeriods = key;
            }
        }
    }
    if (((best.length < 1) || (bestScore <= 0))) {
        return throwError(0);
    }
    return [best, bestScore, checked, null] as const;
}
export function modulo(value: number): number {
    value %= 2 * Math.PI;
    if ((value < 0)) {
        value += (2 * Math.PI);
    }
    return value;
}
export function fitAnnualValues(prices: number[], year: number, p: Params): readonly [
    Model,
    Error | null
] {
    if ((((((year < 1) || (p.candidatePool < 1)) || (p.candidatePool > 16)) || (p.maxComponents < 1)) || (p.maxComponents > 7))) {
        return [new Model({}), new Error("invalid annual bounds")] as const;
    }
    let config = new cyclescanner.ScanConfig({ window: prices.length, minPeriod: p.minPeriod, maxPeriod: p.maxPeriod, smoothing: p.smoothing, genuineThreshold: p.genuineThreshold, phaseCycles: 0, topK: p.candidatePool, peakDistance: p.peakDistance, maxSegments: 5 });
    {
        let err = config.validate();
        if ((err !== null)) {
            return [new Model({}), err] as const;
        }
    }
    let scan = cyclescanner.scanCycles(prices, config);
    let cycles = [] as cyclescanner.RankedCycle[];
    for (const cycle of scan.cycles) {
        if (!Number.isNaN(cycle.period)) {
            cycles = [...cycles, cycle];
        }
    }
    if ((cycles.length < 1)) {
        return [new Model({}), new Error("no accepted annual cycles")] as const;
    }
    let n = prices.length;
    let waves = Array.from({ length: cycles.length }, () => [] as number[]);
    let periods = Array.from({ length: cycles.length }, () => 0);
    for (const [i, cycle] of cycles.entries()) {
        periods[i] = cycle.period;
        waves[i] = Array.from({ length: n }, () => 0);
        for (let j = 0; j < n; j++) {
            waves[i][j] = cyclescanner.modeledValue(cycle, (j - ((n - 1))));
        }
    }
    let [subset, correlation, checked, err] = chooseSubset(hodrickprescottsmoothinganddetrending.cycleData(prices, p.smoothing), waves, periods, p.maxComponents);
    if ((err !== null)) {
        return [new Model({}), err] as const;
    }
    let components = Array.from({ length: subset.length }, () => new Component());
    for (const [i, index] of subset.entries()) {
        let cycle = cycles[index];
        components[i] = new Component({ period: cycle.period, amplitude: cycle.fitAmplitude, phase: modulo((cycle.phase - (((2 * Math.PI) * (n - 1)) / cycle.period))), strength: cycle.strength, genuinePercent: cycle.genuinePercent });
    }
    for (let i = 1; (i < components.length); i++) {
        for (let j = i; ((j > 0) && (components[j].period < components[(j - 1)].period)); j--) {
            [components[j], components[(j - 1)]] = [components[(j - 1)], components[j]];
        }
    }
    return [new Model({ year: year, sampleCount: n, components: components, correlation: correlation, subsetsChecked: checked, candidateCount: cycles.length }), null] as const;
}
export function annualWave(model: Model, index: number): number {
    let values = Array.from({ length: model.components.length }, () => 0);
    for (const [i, c] of model.components.entries()) {
        values[i] = (c.amplitude * Math.sin(((((2 * Math.PI) * index) / c.period) + c.phase)));
    }
    return harmoniccycle.sum(values);
}
