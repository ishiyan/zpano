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
    TransformedDiscreteFourierTransformSpectrumParams,
    defaultParams,
    transformedDiscreteFourierTransformSpectrumDominantCycleMnemonic,
    transformedDiscreteFourierTransformSpectrumMnemonic,
} from './params';

export {
    TransformedDiscreteFourierTransformSpectrumParams,
    defaultParams,
    transformedDiscreteFourierTransformSpectrumDominantCycleMnemonic,
    transformedDiscreteFourierTransformSpectrumMnemonic,
};

const TWO_PI = 2.0 * Math.PI;
const LOG_TEN = Math.LN10;
const DB_CAP = 20.0;
const CENTROID_DB_LIMIT = 3.0;

const ALLOWED_KEYS = new Set<string>([
    'minPeriod',
    'maxPeriod',
    'windowLength',
    'highpassPeriod',
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
    readonly cleanedData: number;
    readonly maximumPower: number;
    readonly dominantCycle: number;
    readonly cosineParts: readonly number[];
    readonly sineParts: readonly number[];
    readonly powers: readonly number[];
    readonly decibels: readonly number[];
}

export interface ResolvedTransformedDiscreteFourierTransformSpectrumParams {
    readonly minPeriod: number;
    readonly maxPeriod: number;
    readonly windowLength: number;
    readonly highpassPeriod: number;
}

function validateInteger(value: unknown, name: string, minimum: number): number {
    if (typeof value !== 'number' || !Number.isInteger(value) || value < minimum) {
        throw new Error(`${name} must be an integer >= ${minimum}`);
    }
    return value;
}

function validateRealAbove(value: unknown, name: string, lowerExclusive: number): number {
    if (typeof value !== 'number' || !Number.isFinite(value) || value <= lowerExclusive) {
        throw new Error(`${name} must be a finite real > ${lowerExclusive}`);
    }
    return value;
}

function validateSample(value: unknown): number {
    if (typeof value !== 'number') {
        throw new Error('input must be a real number; NaN may mark missing data');
    }
    if (!Number.isFinite(value) && !Number.isNaN(value)) {
        throw new Error('input must not be infinite');
    }
    return value;
}

/** Ehlers' Transformed Discrete Fourier Transform Spectrum indicator (January 2007 article). */
export class TransformedDiscreteFourierTransformSpectrum implements Indicator {
    public readonly parameters: ResolvedTransformedDiscreteFourierTransformSpectrumParams;
    private readonly highpassAlpha: number;
    private readonly warmupCount: number;
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
    private previousPrice = 0.0;
    private hpHistory: number[] = [0.0, 0.0, 0.0, 0.0, 0.0, 0.0];
    private cleaned = 0.0;
    private window: number[] = [];
    private cosineParts: number[] = [];
    private sineParts: number[] = [];
    private powers: number[] = [];
    private decibels: number[] = [];
    private maximum = 0.0;
    private dominant = 0.0;
    private lastOutput!: SpectrumOutput;

    public static default(): TransformedDiscreteFourierTransformSpectrum {
        return new TransformedDiscreteFourierTransformSpectrum(defaultParams());
    }

    public static fromParams(
        params?: TransformedDiscreteFourierTransformSpectrumParams,
    ): TransformedDiscreteFourierTransformSpectrum {
        return new TransformedDiscreteFourierTransformSpectrum(params);
    }

    public constructor(params?: TransformedDiscreteFourierTransformSpectrumParams) {
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

        const minPeriod = validateInteger(params?.minPeriod ?? 8, 'minPeriod', 8);
        const maxPeriod = validateInteger(params?.maxPeriod ?? 50, 'maxPeriod', minPeriod);
        const windowLength = validateInteger(params?.windowLength ?? 50, 'windowLength', 1);
        const highpassPeriod = validateRealAbove(params?.highpassPeriod ?? 40, 'highpassPeriod', 4.0);

        const omega = TWO_PI / highpassPeriod;
        const cosineOmega = Math.cos(omega);
        if (cosineOmega === 0.0) {
            throw new Error('highpassPeriod produces singular high-pass coefficient');
        }
        const highpassAlpha = (1.0 - Math.sin(omega)) / cosineOmega;
        if (!Number.isFinite(highpassAlpha)) {
            throw new Error('highpassPeriod produces non-finite high-pass coefficient');
        }

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

        this.parameters = {
            minPeriod,
            maxPeriod,
            windowLength,
            highpassPeriod,
        };
        this.highpassAlpha = highpassAlpha;
        this.warmupCount = Math.max(5, windowLength - 1);

        const channelCount = maxPeriod - minPeriod + 1;
        this.periods = [];
        this.cosines = [];
        this.sines = [];
        for (let channelIndex = 0; channelIndex < channelCount; channelIndex++) {
            const period = minPeriod + channelIndex;
            this.periods.push(period);
            const cosineRow: number[] = [];
            const sineRow: number[] = [];
            for (let lag = 0; lag < windowLength; lag++) {
                const angle = (TWO_PI * lag) / period;
                cosineRow.push(Math.cos(angle));
                sineRow.push(Math.sin(angle));
            }
            this.cosines.push(cosineRow);
            this.sines.push(sineRow);
        }

        this.barComponentFunc = barComponentValue(barComp);
        this.quoteComponentFunc = quoteComponentValue(quoteComp);
        this.tradeComponentFunc = tradeComponentValue(tradeComp);

        const resolvedParams: TransformedDiscreteFourierTransformSpectrumParams = {
            minPeriod,
            maxPeriod,
            windowLength,
            highpassPeriod,
            barComponent: barComp,
            quoteComponent: quoteComp,
            tradeComponent: tradeComp,
        };
        this.mnemonicValue = transformedDiscreteFourierTransformSpectrumMnemonic(resolvedParams);
        this.descriptionValue = `Transformed discrete Fourier transform spectrum ${this.mnemonicValue}`;
        this.mnemonicDC = transformedDiscreteFourierTransformSpectrumDominantCycleMnemonic(resolvedParams);
        this.descriptionDC = `Transformed discrete Fourier transform spectrum dominant cycle ${this.mnemonicDC}`;

        this.reset();
    }

    public get minPeriod(): number {
        return this.parameters.minPeriod;
    }

    public get maxPeriod(): number {
        return this.parameters.maxPeriod;
    }

    public get warmupBars(): number {
        return this.warmupCount;
    }

    public get samplesSeen(): number {
        return this.count;
    }

    public get state(): SpectrumState {
        return {
            highPass: this.hpHistory[0],
            cleanedData: this.cleaned,
            maximumPower: this.maximum,
            dominantCycle: this.dominant,
            cosineParts: this.cosineParts.slice(),
            sineParts: this.sineParts.slice(),
            powers: this.powers.slice(),
            decibels: this.decibels.slice(),
        };
    }

    public get output(): SpectrumOutput {
        return this.lastOutput;
    }

    private unavailableOutput(time: Date): SpectrumOutput {
        return {
            heatmap: Heatmap.newEmptyHeatmap(
                time,
                this.parameters.minPeriod,
                this.parameters.maxPeriod,
                1.0,
            ),
            dominantCycle: Number.NaN,
        };
    }

    public reset(): void {
        const channelCount = this.periods.length;
        this.count = 0;
        this.previousPrice = 0.0;
        this.hpHistory = [0.0, 0.0, 0.0, 0.0, 0.0, 0.0];
        this.cleaned = 0.0;
        this.window = new Array<number>(this.parameters.windowLength).fill(0.0);
        this.cosineParts = new Array<number>(channelCount).fill(0.0);
        this.sineParts = new Array<number>(channelCount).fill(0.0);
        this.powers = new Array<number>(channelCount).fill(0.0);
        this.decibels = new Array<number>(channelCount).fill(0.0);
        this.maximum = 0.0;
        this.dominant = 0.0;
        this.lastOutput = this.unavailableOutput(new Date(0));
    }

    public isPrimed(): boolean {
        return this.count > this.warmupCount;
    }

    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.TransformedDiscreteFourierTransformSpectrum,
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
        if (Number.isNaN(price)) {
            return this.unavailableOutput(time);
        }

        const barNumber = this.count + 1;
        let highPass: number;
        let cleaned: number;
        if (barNumber <= 5) {
            highPass = price;
            cleaned = price;
        } else {
            highPass =
                0.5 * (1.0 + this.highpassAlpha) * (price - this.previousPrice) +
                this.highpassAlpha * this.hpHistory[0];
            cleaned =
                (highPass +
                    2.0 * this.hpHistory[0] +
                    3.0 * this.hpHistory[1] +
                    3.0 * this.hpHistory[2] +
                    2.0 * this.hpHistory[3] +
                    this.hpHistory[4]) /
                12.0;
        }
        if (!Number.isFinite(highPass) || !Number.isFinite(cleaned)) {
            throw new Error('input causes non-finite filter arithmetic; state was not changed');
        }

        const windowLength = this.parameters.windowLength;
        const nextWindow = new Array<number>(windowLength);
        nextWindow[0] = cleaned;
        for (let lag = 1; lag < windowLength; lag++) {
            nextWindow[lag] = this.window[lag - 1];
        }

        const channelCount = this.periods.length;
        const cosineParts = new Array<number>(channelCount);
        const sineParts = new Array<number>(channelCount);
        const powers = new Array<number>(channelCount);
        let maximum = 0.0;
        for (let channelIndex = 0; channelIndex < channelCount; channelIndex++) {
            let cosineSum = 0.0;
            let sineSum = 0.0;
            const cosineRow = this.cosines[channelIndex];
            const sineRow = this.sines[channelIndex];
            for (let lag = 0; lag < windowLength; lag++) {
                cosineSum += nextWindow[lag] * cosineRow[lag];
                sineSum += nextWindow[lag] * sineRow[lag];
            }
            const power = cosineSum * cosineSum + sineSum * sineSum;
            if (!Number.isFinite(cosineSum) || !Number.isFinite(sineSum) || !Number.isFinite(power)) {
                throw new Error('input causes non-finite filter arithmetic; state was not changed');
            }
            cosineParts[channelIndex] = cosineSum;
            sineParts[channelIndex] = sineSum;
            powers[channelIndex] = power;
            if (channelIndex === 0 || power > maximum) {
                maximum = power;
            }
        }

        const decibels = this.decibels.slice();
        for (let channelIndex = 0; channelIndex < channelCount; channelIndex++) {
            const power = powers[channelIndex];
            if (maximum > 0.0 && power > 0.0) {
                decibels[channelIndex] =
                    (-10.0 * Math.log(0.01 / (1.0 - (0.99 * power) / maximum))) / LOG_TEN;
            }
            if (decibels[channelIndex] > DB_CAP) {
                decibels[channelIndex] = DB_CAP;
            }
            if (!Number.isFinite(decibels[channelIndex])) {
                throw new Error('input causes non-finite filter arithmetic; state was not changed');
            }
        }

        let numerator = 0.0;
        let denominator = 0.0;
        for (let channelIndex = 0; channelIndex < channelCount; channelIndex++) {
            const channelDB = decibels[channelIndex];
            if (channelDB < CENTROID_DB_LIMIT) {
                const weight = CENTROID_DB_LIMIT - channelDB;
                numerator += this.periods[channelIndex] * weight;
                denominator += weight;
            }
        }
        const dominant = denominator !== 0.0 ? numerator / denominator : this.dominant;
        if (!Number.isFinite(maximum) || !Number.isFinite(dominant)) {
            throw new Error('input causes non-finite filter arithmetic; state was not changed');
        }

        this.previousPrice = price;
        this.hpHistory = [
            highPass,
            this.hpHistory[0],
            this.hpHistory[1],
            this.hpHistory[2],
            this.hpHistory[3],
            this.hpHistory[4],
        ];
        this.cleaned = cleaned;
        this.window = nextWindow;
        this.cosineParts = cosineParts;
        this.sineParts = sineParts;
        this.powers = powers;
        this.decibels = decibels;
        this.maximum = maximum;
        this.dominant = dominant;
        this.count = barNumber;

        if (this.isPrimed()) {
            const clonedValues = decibels.slice();
            let valueMin = clonedValues[0];
            let valueMax = clonedValues[0];
            for (let channelIndex = 1; channelIndex < clonedValues.length; channelIndex++) {
                const channelValue = clonedValues[channelIndex];
                if (channelValue < valueMin) valueMin = channelValue;
                if (channelValue > valueMax) valueMax = channelValue;
            }
            this.lastOutput = {
                heatmap: Heatmap.newHeatmap(
                    time,
                    this.parameters.minPeriod,
                    this.parameters.maxPeriod,
                    1.0,
                    valueMin,
                    valueMax,
                    clonedValues,
                ),
                dominantCycle: dominant,
            };
        } else {
            this.lastOutput = this.unavailableOutput(time);
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
