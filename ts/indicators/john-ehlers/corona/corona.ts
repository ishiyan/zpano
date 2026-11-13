import { Heatmap } from '../../core/outputs/heatmap';
import { CoronaParams, DefaultCoronaParams } from './params';

export const FIRST_BANK_SAMPLE = 13;
export const TWO_PI = 2.0 * Math.PI;
export const SQRT_THREE_OVER_TWO = Math.sqrt(3.0) / 2.0;

const FIR_WEIGHTS: readonly number[] = [1.0, 2.0, 3.0, 3.0, 2.0, 1.0];

export function validateInteger(name: string, value: number, minimum: number): number {
    if (typeof value !== 'number' || !Number.isInteger(value) || value < minimum) {
        throw new Error(`${name} must be an integer >= ${minimum}`);
    }
    return value;
}

export function validateFinite(name: string, value: number): number {
    if (typeof value !== 'number' || !Number.isFinite(value)) {
        throw new Error(`${name} must be finite`);
    }
    return value;
}

export function validPrice(price: number): boolean {
    if (typeof price !== 'number') {
        throw new Error('price must be a number');
    }
    if (Number.isNaN(price)) {
        return false;
    }
    if (!Number.isFinite(price)) {
        throw new Error('price must be finite or NaN');
    }
    return true;
}

export function checked(value: number): number {
    if (!Number.isFinite(value)) {
        throw new Error('calculation overflow; use finite, reasonably scaled prices');
    }
    return value;
}

export function clip(value: number, lower: number, upper: number): number {
    return Math.max(lower, Math.min(upper, value));
}

export function formatNumber(value: number): string {
    const normalized = Object.is(value, -0) ? 0 : value;
    return Number.isInteger(normalized) ? String(normalized) : String(normalized);
}

export function rollingAppend(history: number[], value: number, length: number): [number, number] {
    history.push(value);
    if (history.length > length) {
        history.shift();
    }
    let minimum = history[0];
    let maximum = history[0];
    for (let i = 1; i < history.length; i++) {
        if (history[i] < minimum) minimum = history[i];
        if (history[i] > maximum) maximum = history[i];
    }
    return [minimum, maximum];
}

export interface CoronaFilter {
    inPhase1: number;
    inPhase2: number;
    quadrature1: number;
    quadrature2: number;
    real1: number;
    real2: number;
    imaginary1: number;
    imaginary2: number;
    power: number;
    decibels: number;
}

export interface CoronaSnapshot {
    readonly parameters: Required<CoronaParams>;
    readonly sampleCount: number;
    readonly accepted: boolean;
    readonly primed: boolean;
    readonly price: number;
    readonly time: Date;
    readonly highPass: number;
    readonly filtered: number;
    readonly bandwidth: number;
    readonly dominantCycle: number;
    readonly dominantCycleMedian: number;
    readonly maximumPower: number;
    readonly normalizationPower: number;
    readonly selectedBins: number;
    readonly hasCycleEstimate: boolean;
    readonly power: number[];
    readonly relativePower: number[];
    readonly decibels: number[];
}

function cloneSnapshot(snapshot: CoronaSnapshot, overrides?: Partial<CoronaSnapshot>): CoronaSnapshot {
    return {
        ...snapshot,
        ...overrides,
        parameters: { ...snapshot.parameters },
        power: snapshot.power.slice(),
        relativePower: snapshot.relativePower.slice(),
        decibels: snapshot.decibels.slice(),
    };
}

export function acceptSnapshot(
    snapshot: CoronaSnapshot,
    expectedParameters: Required<CoronaParams>,
    expectedCount: number,
): boolean {
    const p = snapshot.parameters;
    if (
        p.highPassFilterCutoff !== expectedParameters.highPassFilterCutoff ||
        p.minimalPeriod !== expectedParameters.minimalPeriod ||
        p.maximalPeriod !== expectedParameters.maximalPeriod ||
        p.decibelsLowerThreshold !== expectedParameters.decibelsLowerThreshold ||
        p.decibelsUpperThreshold !== expectedParameters.decibelsUpperThreshold
    ) {
        throw new Error('snapshot parameters do not match consumer engine parameters');
    }
    if (!snapshot.accepted) {
        return false;
    }
    if (snapshot.sampleCount !== expectedCount) {
        throw new Error('shared snapshots must be supplied in accepted-sample order');
    }
    return true;
}

export class Corona {
    public readonly parameters: Required<CoronaParams>;
    public readonly minimalPeriod: number;
    public readonly maximalPeriod: number;
    public readonly minimalPeriodTimesTwo: number;
    public readonly maximalPeriodTimesTwo: number;
    public readonly filterBankLength: number;
    public readonly periods: readonly number[];

    private readonly betas: readonly number[];
    private readonly pole: number;
    private _sampleCount = 0;
    private pricePrevious = 0;
    private hpHistory: number[] = [0, 0, 0, 0, 0, 0];
    private filteredPrevious = 0;
    private bank: CoronaFilter[] = [];
    private dc = 0;
    private dcHistory: number[] = [0, 0, 0, 0, 0];
    private _lastSnapshot: CoronaSnapshot | null = null;

    public constructor(params?: CoronaParams) {
        const hpCutoff = validateInteger(
            'highPassFilterCutoff',
            params?.highPassFilterCutoff ?? DefaultCoronaParams.highPassFilterCutoff,
            5,
        );
        const minPeriod = validateInteger(
            'minimalPeriod',
            params?.minimalPeriod ?? DefaultCoronaParams.minimalPeriod,
            6,
        );
        const maxPeriod = validateInteger(
            'maximalPeriod',
            params?.maximalPeriod ?? DefaultCoronaParams.maximalPeriod,
            minPeriod + 1,
        );
        const lowerThreshold = validateFinite(
            'decibelsLowerThreshold',
            params?.decibelsLowerThreshold ?? DefaultCoronaParams.decibelsLowerThreshold,
        );
        const upperThreshold = validateFinite(
            'decibelsUpperThreshold',
            params?.decibelsUpperThreshold ?? DefaultCoronaParams.decibelsUpperThreshold,
        );
        if (!(0.0 <= lowerThreshold && lowerThreshold < upperThreshold && upperThreshold > 0.0)) {
            throw new Error('require 0 <= decibelsLowerThreshold < decibelsUpperThreshold');
        }

        const angle = TWO_PI / hpCutoff;
        const pole = (1.0 - Math.sin(angle)) / Math.cos(angle);
        if (!Number.isFinite(pole) || !(0.0 < pole && pole < 1.0)) {
            throw new Error('high-pass coefficient must have a stable positive pole');
        }

        this.parameters = {
            highPassFilterCutoff: hpCutoff,
            minimalPeriod: minPeriod,
            maximalPeriod: maxPeriod,
            decibelsLowerThreshold: lowerThreshold,
            decibelsUpperThreshold: upperThreshold,
        };
        this.minimalPeriod = minPeriod;
        this.maximalPeriod = maxPeriod;
        this.minimalPeriodTimesTwo = 2 * minPeriod;
        this.maximalPeriodTimesTwo = 2 * maxPeriod;

        const periods: number[] = [];
        const betas: number[] = [];
        for (let n2 = this.minimalPeriodTimesTwo; n2 <= this.maximalPeriodTimesTwo; n2++) {
            const period = n2 / 2.0;
            periods.push(period);
            betas.push(Math.cos(TWO_PI / period));
        }
        this.periods = periods;
        this.betas = betas;
        this.filterBankLength = periods.length;
        this.pole = pole;
        this.reset();
    }

    public reset(): void {
        this._sampleCount = 0;
        this.pricePrevious = 0;
        this.hpHistory = [0, 0, 0, 0, 0, 0];
        this.filteredPrevious = 0;
        this.bank = this.periods.map(() => ({
            inPhase1: 0,
            inPhase2: 0,
            quadrature1: 0,
            quadrature2: 0,
            real1: 0,
            real2: 0,
            imaginary1: 0,
            imaginary2: 0,
            power: 0,
            decibels: 0,
        }));
        this.dc = 0;
        this.dcHistory = [0, 0, 0, 0, 0];
        this._lastSnapshot = null;
    }

    public isPrimed(): boolean {
        return this._sampleCount >= FIRST_BANK_SAMPLE;
    }

    public get sampleCount(): number {
        return this._sampleCount;
    }

    public get dominantCycle(): number {
        return this.dc;
    }

    public get dominantCycleMedian(): number {
        return this._lastSnapshot !== null ? this._lastSnapshot.dominantCycleMedian : this.minimalPeriod;
    }

    public get maximumPower(): number {
        return this._lastSnapshot !== null ? this._lastSnapshot.maximumPower : 0.0;
    }

    public get lastSnapshot(): CoronaSnapshot | null {
        return this._lastSnapshot !== null ? cloneSnapshot(this._lastSnapshot) : null;
    }

    public update(price: number, time: Date = new Date(0)): CoronaSnapshot {
        if (!validPrice(price)) {
            if (this._lastSnapshot !== null) {
                return cloneSnapshot(this._lastSnapshot, { accepted: false, price, time });
            }
            return {
                parameters: { ...this.parameters },
                sampleCount: 0,
                accepted: false,
                primed: false,
                price,
                time,
                highPass: 0.0,
                filtered: 0.0,
                bandwidth: 0.5,
                dominantCycle: 0.0,
                dominantCycleMedian: this.minimalPeriod,
                maximumPower: 0.0,
                normalizationPower: 0.01,
                selectedBins: 0,
                hasCycleEstimate: false,
                power: new Array<number>(this.filterBankLength).fill(0.0),
                relativePower: new Array<number>(this.filterBankLength).fill(0.0),
                decibels: new Array<number>(this.filterBankLength).fill(0.0),
            };
        }

        const count = this._sampleCount + 1;
        const previous = count === 1 ? price : this.pricePrevious;
        const hp = checked(0.5 * (1.0 + this.pole) * (price - previous) + this.pole * this.hpHistory[5]);
        const hpHistory = [
            this.hpHistory[1],
            this.hpHistory[2],
            this.hpHistory[3],
            this.hpHistory[4],
            this.hpHistory[5],
            hp,
        ];
        let firSum = 0.0;
        for (let i = 0; i < 6; i++) {
            firSum += hpHistory[i] * FIR_WEIGHTS[i];
        }
        let filtered = checked(firSum / 12.0);
        if (count < 7) {
            filtered = checked(price - previous);
        }
        if (count === 1) {
            filtered = 0.0;
        }
        const bandwidth = Math.max(0.1, 0.5 - 0.015 * count);

        const nextBank = new Array<CoronaFilter>(this.filterBankLength);
        for (let i = 0; i < this.filterBankLength; i++) {
            const old = this.bank[i];
            if (count < FIRST_BANK_SAMPLE) {
                nextBank[i] = { ...old };
                continue;
            }
            const period = this.periods[i];
            const beta = this.betas[i];
            const gamma = 1.0 / Math.cos((2.0 * TWO_PI * bandwidth) / period);
            const alpha = gamma - Math.sqrt(gamma * gamma - 1.0);
            const inPhase = filtered;
            const quadrature = checked((period / TWO_PI) * (filtered - this.filteredPrevious));
            const realPart = checked(
                0.5 * (1.0 - alpha) * (inPhase - old.inPhase2) +
                    beta * (1.0 + alpha) * old.real1 -
                    alpha * old.real2,
            );
            const imaginaryPart = checked(
                0.5 * (1.0 - alpha) * (quadrature - old.quadrature2) +
                    beta * (1.0 + alpha) * old.imaginary1 -
                    alpha * old.imaginary2,
            );
            const power = checked(realPart * realPart + imaginaryPart * imaginaryPart);
            nextBank[i] = {
                inPhase1: inPhase,
                inPhase2: old.inPhase1,
                quadrature1: quadrature,
                quadrature2: old.quadrature1,
                real1: realPart,
                real2: old.real1,
                imaginary1: imaginaryPart,
                imaginary2: old.imaginary1,
                power,
                decibels: old.decibels,
            };
        }

        let maximum = 0.0;
        for (let i = 0; i < nextBank.length; i++) {
            if (nextBank[i].power > maximum) maximum = nextBank[i].power;
        }
        const firstPower = nextBank[0].power !== 0.0 ? nextBank[0].power : 0.01;
        const normalizer = Math.max(firstPower, maximum);

        let numerator = 0.0;
        let denominator = 0.0;
        let selected = 0;
        const lowerThreshold = this.parameters.decibelsLowerThreshold;
        const upperThreshold = this.parameters.decibelsUpperThreshold;
        const powers = new Array<number>(this.filterBankLength);
        const relativePowers = new Array<number>(this.filterBankLength);
        const decibelsSlice = new Array<number>(this.filterBankLength);

        for (let i = 0; i < this.filterBankLength; i++) {
            const channel = nextBank[i];
            const relative = channel.power / normalizer;
            const raw =
                relative > 0.0 ? 10.0 * Math.log10((1.0 - 0.99 * relative) / 0.01) : channel.decibels;
            const decibels = Math.min(upperThreshold, 0.33 * raw + 0.67 * channel.decibels);
            channel.decibels = decibels;
            powers[i] = channel.power;
            relativePowers[i] = relative;
            decibelsSlice[i] = decibels;
            if (decibels <= lowerThreshold) {
                const weight = upperThreshold - decibels;
                numerator += this.periods[i] * weight;
                denominator += weight;
                selected += 1;
            }
        }

        const dc = denominator > 0.0 ? numerator / denominator : this.dc;
        const dcHistory = [this.dcHistory[1], this.dcHistory[2], this.dcHistory[3], this.dcHistory[4], dc];
        const sortedDC = dcHistory.slice().sort((a, b) => a - b);
        const median = Math.max(this.minimalPeriod, sortedDC[2]);

        const snapshot: CoronaSnapshot = {
            parameters: { ...this.parameters },
            sampleCount: count,
            accepted: true,
            primed: count >= FIRST_BANK_SAMPLE,
            price,
            time,
            highPass: hp,
            filtered,
            bandwidth,
            dominantCycle: dc,
            dominantCycleMedian: median,
            maximumPower: maximum,
            normalizationPower: normalizer,
            selectedBins: selected,
            hasCycleEstimate: maximum > 0.0 && selected > 0,
            power: powers,
            relativePower: relativePowers,
            decibels: decibelsSlice,
        };

        this._sampleCount = count;
        this.pricePrevious = price;
        this.hpHistory = hpHistory;
        this.filteredPrevious = filtered;
        this.bank = nextBank;
        this.dc = dc;
        this.dcHistory = dcHistory;
        this._lastSnapshot = cloneSnapshot(snapshot);
        return snapshot;
    }
}

export interface AdaptiveCycle {
    readonly bandPass: number;
    readonly quadrature: number;
    readonly amplitude: number;
    readonly lead60: number;
}

export class AdaptiveBandPass {
    private count = 0;
    private prices: [number, number] = [0.0, 0.0];
    private bp1 = 0.0;
    private bp2 = 0.0;

    public reset(): void {
        this.count = 0;
        this.prices = [0.0, 0.0];
        this.bp1 = 0.0;
        this.bp2 = 0.0;
    }

    public update(price: number, period: number): AdaptiveCycle {
        this.count += 1;
        if (this.count <= 2) {
            this.prices[this.count - 1] = price;
            return { bandPass: 0.0, quadrature: 0.0, amplitude: 0.0, lead60: 0.0 };
        }
        const omega = TWO_PI / period;
        const beta = Math.cos(omega);
        const gamma = 1.0 / Math.cos(2.0 * omega * 0.1);
        const alpha = gamma - Math.sqrt(gamma * gamma - 1.0);
        const bp = checked(
            0.5 * (1.0 - alpha) * (price - this.prices[0]) +
                beta * (1.0 + alpha) * this.bp1 -
                alpha * this.bp2,
        );
        const quadrature = checked((bp - this.bp1) / omega);
        const amplitude = checked(Math.sqrt(bp * bp + quadrature * quadrature));
        const lead60 = checked(0.5 * bp + SQRT_THREE_OVER_TWO * quadrature);
        this.prices = [this.prices[1], price];
        this.bp2 = this.bp1;
        this.bp1 = bp;
        return { bandPass: bp, quadrature, amplitude, lead60 };
    }
}

export type RasterKind = 'swing' | 'trend' | 'snr';

export class Raster {
    public readonly length: number;
    public readonly lower: number;
    public readonly upper: number;
    public readonly ceiling: number;
    public readonly step: number;
    public readonly first: number;
    public readonly resolution: number;
    private values: number[];

    public constructor(length: number, lower: number, upper: number, ceiling: number) {
        this.length = validateInteger('rasterLength', length, 2);
        this.lower = validateFinite('minParameterValue', lower);
        this.upper = validateFinite('maxParameterValue', upper);
        this.ceiling = validateFinite('maxRasterValue', ceiling);
        if (this.upper <= this.lower || this.ceiling <= 0.0) {
            throw new Error('require maxParameterValue > minParameterValue and maxRasterValue > 0');
        }
        this.step = (this.upper - this.lower) / this.length;
        this.first = this.lower + this.step;
        this.resolution = 1.0 / this.step;
        this.values = new Array<number>(this.length).fill(0.0);
    }

    public reset(): void {
        this.values.fill(0.0);
    }

    public valuesCopy(): number[] {
        return this.values.slice();
    }

    public setValues(values: readonly number[]): void {
        for (let i = 0; i < this.length; i++) {
            this.values[i] = values[i];
        }
    }

    public empty(time: Date): Heatmap {
        return Heatmap.newEmptyHeatmap(time, this.first, this.upper, this.resolution);
    }

    public column(time: Date): Heatmap {
        const copied = this.values.slice();
        let minimum = copied[0];
        let maximum = copied[0];
        for (let i = 1; i < copied.length; i++) {
            if (copied[i] < minimum) minimum = copied[i];
            if (copied[i] > maximum) maximum = copied[i];
        }
        return Heatmap.newHeatmap(time, this.first, this.upper, this.resolution, minimum, maximum, copied);
    }

    public update(position: number, width: number, kind: RasterKind, suppress = false): void {
        const marker = Math.floor(this.length * position + 0.5);
        const [exponent, scale, historyScale]: [number, number, number] =
            kind === 'swing' ? [0.95, 0.5, 0.5] : kind === 'trend' ? [0.85, 0.8, 0.2] : [0.8, 0.5, 1.0];
        for (let i = 0; i < this.length; i++) {
            const old = this.values[i];
            const j = i + 1;
            let value: number;
            if (suppress) {
                value = this.ceiling;
            } else if (j === marker) {
                value = 0.5 * old;
            } else if (width <= 0.0) {
                value = this.ceiling;
            } else {
                const distance = Math.abs(this.ceiling * position - (this.ceiling * j) / this.length) / width;
                if (kind === 'snr' && j > marker && distance <= 1.0) {
                    value = this.ceiling;
                } else {
                    value = scale * (Math.pow(distance, exponent) + historyScale * old);
                }
            }
            this.values[i] = clip(value, 0.0, this.ceiling);
        }
    }
}
