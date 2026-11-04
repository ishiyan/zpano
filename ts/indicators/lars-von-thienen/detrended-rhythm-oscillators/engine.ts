// SPDX-License-Identifier: MPL-2.0
// Source: Detrended Rhythm Oscillator, © WhenToTrade / Lars von Thienen.
import * as harmoniccycle from '../harmonic-cycle/model.js';
import { DetrendedRhythmOscillatorsParams } from "./params.js";
import { validateInteger, validateReal, finite, validateSample, throwError, isInfinite, mod, push } from '../harmonic-cycle/validation.js';
// Native numerical port of the Go implementation; Lars von Thienen, MPL-2.0.
// Research and startup conventions: description.md.
export class Snapshot {
    value: number = 0;
    direction: number = 0;
    cycleLength: number = 0;
    pivotValue: number = 0;
    pivotOffset: number = 0;
    constructor(values: Partial<Snapshot> = {}) { Object.assign(this, values); }
}
export function emptySnapshot(): Snapshot {
    let nan = NaN;
    return new Snapshot({ value: nan, direction: nan, cycleLength: nan, pivotValue: nan, pivotOffset: nan });
}
export class Engine {
    p: DetrendedRhythmOscillatorsParams = undefined as unknown as DetrendedRhythmOscillatorsParams;
    fast: number = 0;
    slow: number = 0;
    window: number[] = [];
    pivots: Pivot[] = [];
    direction: number = 0;
    samplesSeen: number = 0;
    lastCycleLength: number = 0;
    constructor(values: Partial<Engine> = {}) { Object.assign(this, values); }
    reset() {
        this.fast = NaN;
        this.slow = NaN;
        this.window = [];
        this.pivots = [];
        this.direction = 0;
        this.samplesSeen = 0;
        this.lastCycleLength = NaN;
    }
    primed(): boolean {
        return (this.samplesSeen > this.warmup());
    }
    warmup(): number {
        return rhythmWarmup(this.p);
    }
    update(price: number): Snapshot {
        harmoniccycle.validateSample(price);
        if (Number.isNaN(price)) {
            return emptySnapshot();
        }
        let value = price;
        if (this.p.detrend) {
            this.fast = ema(price, this.fast, this.p.fastLength);
            this.slow = ema(price, this.slow, this.p.slowLength);
            value = NaN;
            if ((this.slow !== 0)) {
                value = harmoniccycle.finite(((100 * ((this.fast - this.slow))) / this.slow));
            }
        }
        let index = this.samplesSeen;
        this.samplesSeen++;
        if (Number.isNaN(value)) {
            return emptySnapshot();
        }
        this.window = harmoniccycle.push(this.window, value, this.p.zigPeriod);
        let [maximum, minimum] = [this.window[0], this.window[0]];
        for (const item of this.window) {
            maximum = Math.max(maximum, item);
            minimum = Math.min(minimum, item);
        }
        let [high, low] = [(value === maximum), (value === minimum)];
        let direction = this.direction;
        if ((high && !low)) {
            direction = 1;
        }
        else if ((low && !high)) {
            direction = -1;
        }
        if (((direction !== 0) && ((high || low)))) {
            if (((direction !== this.direction) || this.pivots.length === 0)) {
                this.pivots = [...[new Pivot({ value: value, index: index })], ...this.pivots];
                if ((this.pivots.length > (this.p.cycleMemory + 3))) {
                    this.pivots = this.pivots.slice(0, (this.p.cycleMemory + 3));
                }
            }
            else if ((((direction === 1) && (value > this.pivots[0].value)) || ((direction === -1) && (value < this.pivots[0].value)))) {
                this.pivots[0] = new Pivot({ value: value, index: index });
            }
        }
        this.direction = direction;
        let cycles = [] as number[];
        if ((this.pivots.length >= 5)) {
            for (let i = 1; (i < (this.pivots.length - 2)); i++) {
                cycles = [...cycles, (this.pivots[i].index - this.pivots[(i + 2)].index)];
            }
        }
        if ((cycles.length > this.p.cycleMemory)) {
            cycles = cycles.slice((cycles.length - this.p.cycleMemory));
        }
        this.lastCycleLength = median(cycles);
        if (!this.primed()) {
            return emptySnapshot();
        }
        let [pivotValue, offset] = [NaN, NaN];
        if (this.pivots.length !== 0) {
            pivotValue = this.pivots[0].value;
            offset = (this.pivots[0].index - index);
        }
        return new Snapshot({ value: value, direction: direction, cycleLength: this.lastCycleLength, pivotValue: pivotValue, pivotOffset: offset });
    }
}
export class Pivot {
    value: number = 0;
    index: number = 0;
    constructor(values: Partial<Pivot> = {}) { Object.assign(this, values); }
}
export function ema(value: number, previous: number, length: number): number {
    if (Number.isNaN(previous)) {
        return value;
    }
    let alpha = (2 / (length + 1));
    return harmoniccycle.finite(((alpha * value) + (((1 - alpha)) * previous)));
}
export function rhythmWarmup(p: DetrendedRhythmOscillatorsParams): number {
    if (p.detrend) {
        return (p.slowLength - 1);
    }
    return 0;
}
export function median(values: number[]): number {
    if ((values.length === 0 || (values.length < 1))) {
        return NaN;
    }
    values.sort((a, b) => a - b);
    let middle = Math.trunc(values.length / 2);
    if (((values.length % 2) !== 0)) {
        return values[middle];
    }
    return (((values[(middle - 1)] + values[middle])) / 2);
}
