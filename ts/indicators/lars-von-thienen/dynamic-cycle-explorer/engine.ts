import * as harmoniccycle from '../harmonic-cycle/model.js';
import * as cyclescanner from '../cycle-scanner/scanner.js';
import { DynamicCycleExplorerParams } from "./params.js";
import { validateInteger, validateReal, finite, validateSample, throwError, isInfinite, mod, push } from '../harmonic-cycle/validation.js';
// Native numerical port of the Go implementation; Lars von Thienen, CC BY 4.0.
// Research and startup conventions: description.md.
export class Snapshot {
    value: number = 0;
    period: number = 0;
    phase: number = 0;
    etaTop: number = 0;
    etaBottom: number = 0;
    highlighter: number = 0;
    highlighterIndex: number = 0;
    fittedAtHighlighter: number = 0;
    constructor(values: Partial<Snapshot> = {}) { Object.assign(this, values); }
}
export function emptySnapshot(): Snapshot {
    let nan = NaN;
    return new Snapshot({ value: nan, period: nan, phase: nan, etaTop: nan, etaBottom: nan, highlighter: nan, highlighterIndex: nan, fittedAtHighlighter: nan });
}
export class Engine {
    p: DynamicCycleExplorerParams = undefined as unknown as DynamicCycleExplorerParams;
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
        if (Number.isNaN(cycle.period)) {
            return emptySnapshot();
        }
        let longLength = Math.trunc(cycle.period);
        let shortLength = Math.max(2, Math.trunc(Math.floor(((cycle.period / 2) + 0.5))));
        let delay = Math.trunc(longLength / 2);
        let center = ((this.scanner.prices.length - 1) - delay);
        let highlighter = (centeredMean(this.scanner.prices, center, shortLength) - centeredMean(this.scanner.prices, center, longLength));
        let [index, fitted] = [NaN, NaN];
        if ((!Number.isNaN(highlighter) && !isInfinite(highlighter, 0))) {
            index = ((this.scanner.samplesSeen - 1) - delay);
            fitted = cyclescanner.modeledValue(cycle, -delay);
        }
        return new Snapshot({ value: cycle.value, period: cycle.period, phase: cycle.phase, etaTop: cycle.etaTop, etaBottom: cycle.etaBottom, highlighter: highlighter, highlighterIndex: index, fittedAtHighlighter: fitted });
    }
}
export function centeredMean(values: number[], center: number, length: number): number {
    let half = Math.trunc(length / 2);
    if ((((center - half) < 0) || ((center + half) >= values.length))) {
        return NaN;
    }
    let sum: number = 0;
    if (((length % 2) !== 0)) {
        for (const value of values.slice((center - half), ((center + half) + 1))) {
            sum += value;
        }
        return (sum / length);
    }
    for (const value of values.slice(((center - half) + 1), (center + half))) {
        sum += value;
    }
    return (((((0.5 * values[(center - half)]) + sum) + (0.5 * values[(center + half)]))) / length);
}
