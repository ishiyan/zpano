import { Bar } from '../../../entities/bar';
import { BarComponent, DefaultBarComponent, barComponentMnemonic, barComponentValue } from '../../../entities/bar-component';
import { Quote } from '../../../entities/quote';
import { DefaultQuoteComponent, QuoteComponent, quoteComponentMnemonic, quoteComponentValue } from '../../../entities/quote-component';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { DefaultTradeComponent, TradeComponent, tradeComponentMnemonic, tradeComponentValue } from '../../../entities/trade-component';
import { buildMetadata } from '../../core/build-metadata';
import { Indicator } from '../../core/indicator';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { Heatmap } from '../../core/outputs/heatmap';
import {
    DiscreteFourierTransformSpectrumParams,
    defaultParams,
    discreteFourierTransformSpectrumDominantCycleMnemonic,
    discreteFourierTransformSpectrumMnemonic,
} from './params';

export {
    DiscreteFourierTransformSpectrumParams,
    defaultParams,
    discreteFourierTransformSpectrumDominantCycleMnemonic,
    discreteFourierTransformSpectrumMnemonic,
};

const TWO_PI = 2.0 * Math.PI;
const SQRT_TWO = Math.SQRT2;
const AGC_DECAY = 0.995;
const CENTROID_THRESHOLD = 0.5;

const ALLOWED_KEYS = new Set<string>([
    'minPeriod',
    'maxPeriod',
    'windowLength',
    'highpassPeriod',
    'smoothingPeriod',
    'spectralDilationCompensation',
    'enhanceResolution',
    'barComponent',
    'quoteComponent',
    'tradeComponent',
]);

export interface SpectrumOutput {
    readonly heatmap: Heatmap;
    readonly dominantCycle: number;
}

export interface SpectrumState {
    readonly highPass: number;
    readonly filtered: number;
    readonly rawPower: readonly number[];
    readonly normalizationMaximum: number;
    readonly internalDominantCycle: number;
}

export interface ResolvedDiscreteFourierTransformSpectrumParams {
    readonly minPeriod: number;
    readonly maxPeriod: number;
    readonly windowLength: number;
    readonly highpassPeriod: number;
    readonly smoothingPeriod: number;
    readonly spectralDilationCompensation: boolean;
    readonly enhanceResolution: boolean;
}

function validateInteger(value: unknown, name: string, minimum: number): number {
    if (typeof value !== 'number' || !Number.isInteger(value) || value < minimum) {
        throw new Error(`${name} must be an integer >= ${minimum}`);
    }
    return value;
}

function validateReal(value: unknown, name: string, minimum: number): number {
    if (typeof value !== 'number' || !Number.isFinite(value) || value < minimum) {
        throw new Error(`${name} must be a finite real >= ${minimum}`);
    }
    return value;
}

function validateBoolean(value: unknown, name: string): boolean {
    if (typeof value !== 'boolean') {
        throw new Error(`${name} must be a boolean`);
    }
    return value;
}

function validateSample(value: unknown): number {
    if (typeof value !== 'number') {
        throw new Error('price must be a real number');
    }
    if (!Number.isFinite(value) && !Number.isNaN(value)) {
        throw new Error('price must be finite or NaN (missing)');
    }
    return value;
}

class RoofFilter {
    private readonly hp0: number;
    private readonly hp1: number;
    private readonly hp2: number;
    private readonly ss1: number;
    private readonly ss2: number;
    private readonly ss3: number;
    public x1 = 0.0;
    public x2 = 0.0;
    public h1 = 0.0;
    public h2 = 0.0;
    public f1 = 0.0;
    public f2 = 0.0;

    public constructor(highpassPeriod: number, smoothingPeriod: number) {
        const hp = validateReal(highpassPeriod, 'highpassPeriod', 5.0);
        const lp = validateReal(smoothingPeriod, 'smoothingPeriod', 2.0);
        const angle = TWO_PI / (SQRT_TWO * hp);
        const cosine = Math.cos(angle);
        const alpha = (cosine + Math.sin(angle) - 1.0) / cosine;
        const pole = 1.0 - alpha;
        const a = Math.exp((-SQRT_TWO * Math.PI) / lp);
        this.hp0 = (1.0 - alpha / 2.0) * (1.0 - alpha / 2.0);
        this.hp1 = 2.0 * pole;
        this.hp2 = pole * pole;
        this.ss2 = 2.0 * a * Math.cos((SQRT_TWO * Math.PI) / lp);
        this.ss3 = -a * a;
        this.ss1 = 1.0 - this.ss2 - this.ss3;
        if (
            !(pole > 0.0 && pole < 1.0 && a > 0.0 && a < 1.0 && this.ss1 > 0.0) ||
            !Number.isFinite(this.hp0) ||
            !Number.isFinite(this.ss1)
        ) {
            throw new Error('periods must produce finite resolvable stable coefficients');
        }
    }

    public reset(): void {
        this.x1 = 0.0;
        this.x2 = 0.0;
        this.h1 = 0.0;
        this.h2 = 0.0;
        this.f1 = 0.0;
        this.f2 = 0.0;
    }

    public candidate(price: number): [number, number] {
        const high =
            this.hp0 * (price - 2.0 * this.x1 + this.x2) +
            this.hp1 * this.h1 -
            this.hp2 * this.h2;
        const filtered =
            (this.ss1 * (high + this.h1)) / 2.0 +
            this.ss2 * this.f1 +
            this.ss3 * this.f2;
        if (!Number.isFinite(high) || !Number.isFinite(filtered)) {
            throw new Error('roofing arithmetic overflowed');
        }
        return [high, filtered];
    }

    public commit(price: number, high: number, filtered: number): void {
        this.x2 = this.x1;
        this.x1 = price;
        this.h2 = this.h1;
        this.h1 = high;
        this.f2 = this.f1;
        this.f1 = filtered;
    }
}

function normalizeAndCentroid(
    periods: readonly number[],
    raw: readonly number[],
    previousMax: number,
    previousDC: number,
    hasDC: boolean,
    enhance: boolean,
): { maximum: number; dominant: number; hasDC: boolean; values: number[] } {
    let maximum = AGC_DECAY * previousMax;
    for (const power of raw) {
        if (!Number.isFinite(power) || power < 0.0) {
            throw new Error('spectral power arithmetic overflowed');
        }
        if (power > maximum) {
            maximum = power;
        }
    }
    const normalized = raw.map((power) => (maximum > 0.0 ? power / maximum : 0.0));
    let sp = 0.0;
    let spx = 0.0;
    for (let index = 0; index < periods.length; index++) {
        const power = normalized[index];
        if (power >= CENTROID_THRESHOLD) {
            spx += periods[index] * power;
            sp += power;
        }
    }
    let dominant = previousDC;
    let nextHasDC = hasDC;
    if (sp !== 0.0) {
        dominant = spx / sp;
        nextHasDC = true;
    }
    if (!Number.isFinite(dominant)) {
        throw new Error('centroid arithmetic overflowed');
    }
    const values = enhance
        ? normalized.map((value) => value * value * value)
        : normalized;
    return { maximum, dominant, hasDC: nextHasDC, values };
}

/** Ehlers' Discrete Fourier Transform Spectrum indicator (2013 Book Listing 9-1). */
export class DiscreteFourierTransformSpectrum implements Indicator {
    public readonly parameters: ResolvedDiscreteFourierTransformSpectrumParams;
    private readonly roof: RoofFilter;
    private readonly periods: number[];
    private readonly cosines: number[][];
    private readonly sines: number[][];
    private readonly mnemonicValue: string;
    private readonly descriptionValue: string;
    private readonly mnemonicDC: string;
    private readonly descriptionDC: string;
    private readonly barComponentFunc: (bar: Bar) => number;
    private readonly quoteComponentFunc: (quote: Quote) => number;
    private readonly tradeComponentFunc: (trade: Trade) => number;

    private count = 0;
    private history: number[] = [];
    private raw: number[] = [];
    private maximum = 0.0;
    private dominant = 0.0;
    private hasDominant = false;
    private lastOutput!: SpectrumOutput;

    public static default(): DiscreteFourierTransformSpectrum {
        return new DiscreteFourierTransformSpectrum(defaultParams());
    }

    public static fromParams(params?: DiscreteFourierTransformSpectrumParams): DiscreteFourierTransformSpectrum {
        return new DiscreteFourierTransformSpectrum(params);
    }

    public constructor(params?: DiscreteFourierTransformSpectrumParams) {
        if (params !== undefined && params !== null) {
            if (typeof params !== 'object' || Array.isArray(params)) {
                throw new Error('params must be an object');
            }
            for (const key of Object.keys(params)) {
                if (!ALLOWED_KEYS.has(key)) {
                    throw new Error(`unsupported parameter '${key}'`);
                }
                if ((params as Record<string, unknown>)[key] === null) {
                    throw new Error(`parameter '${key}' cannot be null`);
                }
            }
        }

        const minPeriod = validateInteger(params?.minPeriod ?? 10, 'minPeriod', 3);
        const maxPeriod = validateInteger(params?.maxPeriod ?? 48, 'maxPeriod', minPeriod);
        const windowLength = validateInteger(params?.windowLength ?? 48, 'windowLength', 2);
        const highpassPeriod = validateReal(params?.highpassPeriod ?? 48, 'highpassPeriod', 5.0);
        const smoothingPeriod = validateReal(params?.smoothingPeriod ?? 10, 'smoothingPeriod', 2.0);
        const sdc = validateBoolean(
            params?.spectralDilationCompensation ?? true,
            'spectralDilationCompensation',
        );
        const enhance = validateBoolean(
            params?.enhanceResolution ?? false,
            'enhanceResolution',
        );

        const barComp = params?.barComponent ?? DefaultBarComponent;
        const quoteComp = params?.quoteComponent ?? DefaultQuoteComponent;
        const tradeComp = params?.tradeComponent ?? DefaultTradeComponent;
        if (typeof barComp !== 'number' || barComponentMnemonic(barComp) === '??') {
            throw new Error('invalid barComponent');
        }
        if (typeof quoteComp !== 'number' || quoteComponentMnemonic(quoteComp) === '??') {
            throw new Error('invalid quoteComponent');
        }
        if (typeof tradeComp !== 'number' || tradeComponentMnemonic(tradeComp) === '??') {
            throw new Error('invalid tradeComponent');
        }

        this.roof = new RoofFilter(highpassPeriod, smoothingPeriod);
        this.parameters = {
            minPeriod,
            maxPeriod,
            windowLength,
            highpassPeriod,
            smoothingPeriod,
            spectralDilationCompensation: sdc,
            enhanceResolution: enhance,
        };

        const cells = maxPeriod - minPeriod + 1;
        this.periods = [];
        this.cosines = [];
        this.sines = [];
        for (let index = 0; index < cells; index++) {
            const period = minPeriod + index;
            this.periods.push(period);
            const cosRow: number[] = [];
            const sinRow: number[] = [];
            for (let n = 0; n < windowLength; n++) {
                const angle = (TWO_PI * n) / period;
                cosRow.push(Math.cos(angle));
                sinRow.push(Math.sin(angle));
            }
            this.cosines.push(cosRow);
            this.sines.push(sinRow);
        }

        this.barComponentFunc = barComponentValue(barComp);
        this.quoteComponentFunc = quoteComponentValue(quoteComp);
        this.tradeComponentFunc = tradeComponentValue(tradeComp);

        const resolvedParams: DiscreteFourierTransformSpectrumParams = {
            minPeriod,
            maxPeriod,
            windowLength,
            highpassPeriod,
            smoothingPeriod,
            spectralDilationCompensation: sdc,
            enhanceResolution: enhance,
            barComponent: barComp,
            quoteComponent: quoteComp,
            tradeComponent: tradeComp,
        };
        this.mnemonicValue = discreteFourierTransformSpectrumMnemonic(resolvedParams);
        this.descriptionValue = `Discrete Fourier transform spectrum ${this.mnemonicValue}`;
        this.mnemonicDC = discreteFourierTransformSpectrumDominantCycleMnemonic(resolvedParams);
        this.descriptionDC = `Discrete Fourier transform spectrum dominant cycle ${this.mnemonicDC}`;

        this.reset();
    }

    public get minPeriod(): number {
        return this.parameters.minPeriod;
    }

    public get maxPeriod(): number {
        return this.parameters.maxPeriod;
    }

    public get warmupBars(): number {
        return this.parameters.windowLength - 1;
    }

    public get samplesSeen(): number {
        return this.count;
    }

    public get state(): SpectrumState {
        return {
            highPass: this.roof.h1,
            filtered: this.roof.f1,
            rawPower: this.raw.slice(),
            normalizationMaximum: this.maximum,
            internalDominantCycle: this.dominant,
        };
    }

    public get output(): SpectrumOutput {
        return this.lastOutput;
    }

    public reset(): void {
        this.roof.reset();
        this.count = 0;
        this.history = new Array<number>(this.parameters.windowLength).fill(0.0);
        this.raw = new Array<number>(this.periods.length).fill(0.0);
        this.maximum = 0.0;
        this.dominant = 0.0;
        this.hasDominant = false;
        this.lastOutput = {
            heatmap: Heatmap.newEmptyHeatmap(new Date(0), this.minPeriod, this.maxPeriod, 1.0),
            dominantCycle: Number.NaN,
        };
    }

    public isPrimed(): boolean {
        return this.count >= this.parameters.windowLength;
    }

    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.DiscreteFourierTransformSpectrum,
            this.mnemonicValue,
            this.descriptionValue,
            [
                { mnemonic: this.mnemonicValue, description: this.descriptionValue },
                { mnemonic: this.mnemonicDC, description: this.descriptionDC },
            ],
        );
    }

    public update(sample: number, time: Date = new Date(0)): SpectrumOutput {
        const price = validateSample(sample);
        const params = this.parameters;
        if (Number.isNaN(price)) {
            return {
                heatmap: Heatmap.newEmptyHeatmap(time, params.minPeriod, params.maxPeriod, 1.0),
                dominantCycle: Number.NaN,
            };
        }

        const [high, filtered] = this.roof.candidate(price);
        const history = new Array<number>(params.windowLength);
        history[0] = filtered;
        for (let index = 1; index < params.windowLength; index++) {
            history[index] = this.history[index - 1];
        }

        const raw = new Array<number>(this.periods.length);
        for (let index = 0; index < this.periods.length; index++) {
            const period = this.periods[index];
            const compensation = params.spectralDilationCompensation ? period : 1.0;
            let cosinePart = 0.0;
            let sinePart = 0.0;
            const cosRow = this.cosines[index];
            const sinRow = this.sines[index];
            for (let n = 0; n < params.windowLength; n++) {
                cosinePart += (history[n] * cosRow[n]) / compensation;
                sinePart += (history[n] * sinRow[n]) / compensation;
            }
            raw[index] = cosinePart * cosinePart + sinePart * sinePart;
        }

        const normalizedResult = normalizeAndCentroid(
            this.periods,
            raw,
            this.maximum,
            this.dominant,
            this.hasDominant,
            params.enhanceResolution,
        );

        this.roof.commit(price, high, filtered);
        this.history = history;
        this.raw = raw;
        this.maximum = normalizedResult.maximum;
        this.dominant = normalizedResult.dominant;
        this.hasDominant = normalizedResult.hasDC;
        this.count += 1;

        if (this.isPrimed()) {
            const clonedValues = normalizedResult.values.slice();
            let valueMin = clonedValues[0];
            let valueMax = clonedValues[0];
            for (let index = 1; index < clonedValues.length; index++) {
                const value = clonedValues[index];
                if (value < valueMin) valueMin = value;
                if (value > valueMax) valueMax = value;
            }
            this.lastOutput = {
                heatmap: Heatmap.newHeatmap(
                    time,
                    params.minPeriod,
                    params.maxPeriod,
                    1.0,
                    valueMin,
                    valueMax,
                    clonedValues,
                ),
                dominantCycle: normalizedResult.hasDC ? normalizedResult.dominant : Number.NaN,
            };
        } else {
            this.lastOutput = {
                heatmap: Heatmap.newEmptyHeatmap(time, params.minPeriod, params.maxPeriod, 1.0),
                dominantCycle: Number.NaN,
            };
        }
        return this.lastOutput;
    }

    private wrapEntity(time: Date, result: SpectrumOutput): IndicatorOutput {
        const dcScalar = new Scalar();
        dcScalar.time = time;
        dcScalar.value = result.dominantCycle;
        return [result.heatmap, dcScalar];
    }

    public updateScalar(sample: Scalar): IndicatorOutput {
        return this.wrapEntity(sample.time, this.update(sample.value, sample.time));
    }

    public updateBar(sample: Bar): IndicatorOutput {
        return this.wrapEntity(sample.time, this.update(this.barComponentFunc(sample), sample.time));
    }

    public updateQuote(sample: Quote): IndicatorOutput {
        return this.wrapEntity(sample.time, this.update(this.quoteComponentFunc(sample), sample.time));
    }

    public updateTrade(sample: Trade): IndicatorOutput {
        return this.wrapEntity(sample.time, this.update(this.tradeComponentFunc(sample), sample.time));
    }
}
