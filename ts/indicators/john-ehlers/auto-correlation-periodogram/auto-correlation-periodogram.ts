import { Bar } from '../../../entities/bar';
import {
    BarComponent,
    DefaultBarComponent,
    barComponentMnemonic,
    barComponentValue,
} from '../../../entities/bar-component';
import { Quote } from '../../../entities/quote';
import {
    DefaultQuoteComponent,
    QuoteComponent,
    quoteComponentMnemonic,
    quoteComponentValue,
} from '../../../entities/quote-component';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import {
    DefaultTradeComponent,
    TradeComponent,
    tradeComponentMnemonic,
    tradeComponentValue,
} from '../../../entities/trade-component';
import { buildMetadata } from '../../core/build-metadata';
import { Indicator } from '../../core/indicator';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { Heatmap } from '../../core/outputs/heatmap';
import {
    AutoCorrelationPeriodogramParams,
    AutoCorrelationPeriodogramSource,
    autoCorrelationPeriodogramDominantCycleMnemonic,
    autoCorrelationPeriodogramMnemonic,
    defaultParams,
    resolveMinPeriod,
} from './params';

export {
    AutoCorrelationPeriodogramParams,
    AutoCorrelationPeriodogramSource,
    autoCorrelationPeriodogramDominantCycleMnemonic,
    autoCorrelationPeriodogramMnemonic,
    defaultParams,
};

const TWO_PI = 2.0 * Math.PI;
const SQRT_TWO = Math.SQRT2;
const INVERSE_SQRT_TWO = 1.0 / SQRT_TWO;
const DFT_LAG_START = 3;
const POWER_SMOOTHING = 0.2;
const POWER_RETENTION = 0.8;
const BOOK_AGC_DECAY = 0.995;
const ARTICLE_THRESHOLD = 0.25;
const BOOK_THRESHOLD = 0.5;

const ALLOWED_KEYS = new Set<string>([
    'source',
    'minPeriod',
    'maxPeriod',
    'averagingLength',
    'enhanceResolution',
    'barComponent',
    'quoteComponent',
    'tradeComponent',
]);

export interface PeriodogramOutput {
    readonly heatmap: Heatmap;
    readonly dominantCycle: number;
}

export interface PeriodogramState {
    readonly highPass: number;
    readonly filtered: number;
    readonly correlations: readonly number[];
    readonly smoothedPower: readonly number[];
    readonly normalizationMaximum: number;
    readonly internalDominantCycle: number;
}

export interface ResolvedAutoCorrelationPeriodogramParams {
    readonly source: AutoCorrelationPeriodogramSource;
    readonly minPeriod: number;
    readonly maxPeriod: number;
    readonly averagingLength: number;
    readonly enhanceResolution: boolean;
}

function validateSource(value: unknown): AutoCorrelationPeriodogramSource {
    if (value === AutoCorrelationPeriodogramSource.Article) {
        return AutoCorrelationPeriodogramSource.Article;
    }
    if (value === AutoCorrelationPeriodogramSource.Book) {
        return AutoCorrelationPeriodogramSource.Book;
    }
    throw new Error('source must be "article" or "book"');
}

function validateInteger(value: unknown, name: string): number {
    if (typeof value !== 'number' || !Number.isInteger(value)) {
        throw new Error(`${name} must be an integer`);
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
        throw new Error('sample must be a real number');
    }
    if (!Number.isFinite(value) && !Number.isNaN(value)) {
        throw new Error('sample must be finite or NaN (skipped)');
    }
    return value;
}

/** Ehlers' Autocorrelation Periodogram indicator with dominant cycle estimation. */
export class AutoCorrelationPeriodogram implements Indicator {
    public readonly parameters: ResolvedAutoCorrelationPeriodogramParams;
    private readonly hp0: number;
    private readonly hp1: number;
    private readonly hp2: number;
    private readonly ss1: number;
    private readonly ss2: number;
    private readonly ss3: number;
    private readonly historyLength: number;
    private readonly binCount: number;
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
    private close1 = 0.0;
    private close2 = 0.0;
    private high1 = 0.0;
    private high2 = 0.0;
    private filtered: number[] = [];
    private correlations: number[] = [];
    private power: number[] = [];
    private maximum = 0.0;
    private dominant = 0.0;
    private hasDominant = false;
    private lastOutput!: PeriodogramOutput;

    public static default(): AutoCorrelationPeriodogram {
        return new AutoCorrelationPeriodogram(defaultParams());
    }

    public static fromParams(
        params?: AutoCorrelationPeriodogramParams,
    ): AutoCorrelationPeriodogram {
        return new AutoCorrelationPeriodogram(params);
    }

    public constructor(params?: AutoCorrelationPeriodogramParams) {
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

        const source = validateSource(
            params?.source ?? AutoCorrelationPeriodogramSource.Article,
        );
        if (params?.minPeriod !== undefined) {
            validateInteger(params.minPeriod, 'minPeriod');
        }
        const minPeriod = validateInteger(
            resolveMinPeriod(source, params?.minPeriod),
            'minPeriod',
        );
        const maxPeriod = validateInteger(params?.maxPeriod ?? 48, 'maxPeriod');
        const averagingLength = validateInteger(
            params?.averagingLength ?? 3,
            'averagingLength',
        );
        const enhanceResolution = validateBoolean(
            params?.enhanceResolution ?? false,
            'enhanceResolution',
        );

        if (minPeriod < 3) {
            throw new Error('minPeriod must be at least 3');
        }
        if (maxPeriod < 5 || maxPeriod <= minPeriod) {
            throw new Error('maxPeriod must be at least 5 and greater than minPeriod');
        }
        if (averagingLength !== 0 && averagingLength < 2) {
            throw new Error('averagingLength must be 0 or at least 2');
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

        let pole: number;
        let hp0: number;
        let hp1: number;
        let hp2: number;
        if (source === AutoCorrelationPeriodogramSource.Article) {
            const theta = TWO_PI / maxPeriod;
            pole = (1.0 - Math.sin(theta)) / Math.cos(theta);
            hp0 = 0.5 * (1.0 + pole);
            hp1 = pole;
            hp2 = 0.0;
        } else {
            const theta = (INVERSE_SQRT_TWO * TWO_PI) / maxPeriod;
            const cosine = Math.cos(theta);
            const alpha = (cosine + Math.sin(theta) - 1.0) / cosine;
            pole = 1.0 - alpha;
            hp0 = (1.0 - alpha / 2.0) * (1.0 - alpha / 2.0);
            hp1 = 2.0 * pole;
            hp2 = pole * pole;
        }
        const decay = Math.exp((-SQRT_TWO * Math.PI) / minPeriod);
        const ss2 = 2.0 * decay * Math.cos((SQRT_TWO * Math.PI) / minPeriod);
        const ss3 = -decay * decay;
        const ss1 = 1.0 - ss2 - ss3;
        if (
            !(pole > 0.0 && pole < 1.0 && decay > 0.0 && decay < 1.0 && ss1 > 0.0) ||
            !Number.isFinite(hp0) ||
            !Number.isFinite(hp1) ||
            !Number.isFinite(hp2) ||
            !Number.isFinite(ss1) ||
            !Number.isFinite(ss2) ||
            !Number.isFinite(ss3)
        ) {
            throw new Error('periods must yield finite stable coefficients');
        }

        this.hp0 = hp0;
        this.hp1 = hp1;
        this.hp2 = hp2;
        this.ss1 = ss1;
        this.ss2 = ss2;
        this.ss3 = ss3;
        this.parameters = {
            source,
            minPeriod,
            maxPeriod,
            averagingLength,
            enhanceResolution,
        };

        const effectiveMaxWindow = averagingLength === 0 ? maxPeriod : averagingLength;
        this.historyLength = Math.max(3, maxPeriod + effectiveMaxWindow);
        this.binCount = maxPeriod - minPeriod + 1;
        this.cosines = [];
        this.sines = [];
        for (let binIndex = 0; binIndex < this.binCount; binIndex++) {
            const period = minPeriod + binIndex;
            const cosRow: number[] = [];
            const sinRow: number[] = [];
            for (let lag = DFT_LAG_START; lag <= maxPeriod; lag++) {
                const angle = (TWO_PI * lag) / period;
                cosRow.push(Math.cos(angle));
                sinRow.push(Math.sin(angle));
            }
            this.cosines.push(cosRow);
            this.sines.push(sinRow);
        }

        this.barComponentFunc = barComponentValue(barComp);
        this.quoteComponentFunc = quoteComponentValue(quoteComp);
        this.tradeComponentFunc = tradeComponentValue(tradeComp);

        const resolvedParams: AutoCorrelationPeriodogramParams = {
            source,
            minPeriod,
            maxPeriod,
            averagingLength,
            enhanceResolution,
            barComponent: barComp,
            quoteComponent: quoteComp,
            tradeComponent: tradeComp,
        };
        this.mnemonicValue = autoCorrelationPeriodogramMnemonic(resolvedParams);
        this.descriptionValue = `Autocorrelation periodogram ${this.mnemonicValue}`;
        this.mnemonicDC = autoCorrelationPeriodogramDominantCycleMnemonic(resolvedParams);
        this.descriptionDC = `Autocorrelation periodogram dominant cycle ${this.mnemonicDC}`;

        this.reset();
    }

    public get minPeriod(): number {
        return this.parameters.minPeriod;
    }

    public get maxPeriod(): number {
        return this.parameters.maxPeriod;
    }

    public get warmupBars(): number {
        return this.historyLength - 1;
    }

    public get samplesSeen(): number {
        return this.count;
    }

    public get state(): PeriodogramState {
        return {
            highPass: this.high1,
            filtered: this.filtered[0],
            correlations: this.correlations.slice(),
            smoothedPower: this.power.slice(),
            normalizationMaximum: this.maximum,
            internalDominantCycle: this.dominant,
        };
    }

    public get output(): PeriodogramOutput {
        return this.lastOutput;
    }

    public reset(): void {
        this.count = 0;
        this.close1 = 0.0;
        this.close2 = 0.0;
        this.high1 = 0.0;
        this.high2 = 0.0;
        this.filtered = new Array<number>(this.historyLength).fill(0.0);
        this.correlations = new Array<number>(this.parameters.maxPeriod + 1).fill(0.0);
        this.power = new Array<number>(this.binCount).fill(0.0);
        this.maximum = 0.0;
        this.dominant = 0.0;
        this.hasDominant = false;
        this.lastOutput = {
            heatmap: Heatmap.newEmptyHeatmap(new Date(0), this.minPeriod, this.maxPeriod, 1.0),
            dominantCycle: Number.NaN,
        };
    }

    public isPrimed(): boolean {
        return this.count >= this.historyLength;
    }

    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.AutoCorrelationPeriodogram,
            this.mnemonicValue,
            this.descriptionValue,
            [
                { mnemonic: this.mnemonicValue, description: this.descriptionValue },
                { mnemonic: this.mnemonicDC, description: this.descriptionDC },
            ],
        );
    }

    public update(sample: number, time: Date = new Date(0)): PeriodogramOutput {
        const price = validateSample(sample);
        const configuration = this.parameters;
        const isBook = configuration.source === AutoCorrelationPeriodogramSource.Book;

        if (Number.isNaN(price)) {
            return {
                heatmap: Heatmap.newEmptyHeatmap(
                    time,
                    configuration.minPeriod,
                    configuration.maxPeriod,
                    1.0,
                ),
                dominantCycle: Number.NaN,
            };
        }

        const high = !isBook
            ? this.hp0 * (price - this.close1) + this.hp1 * this.high1
            : this.hp0 * (price - 2.0 * this.close1 + this.close2) +
              this.hp1 * this.high1 -
              this.hp2 * this.high2;
        const filteredValue =
            (this.ss1 * (high + this.high1)) / 2.0 +
            this.ss2 * this.filtered[0] +
            this.ss3 * this.filtered[1];
        if (!Number.isFinite(high) || !Number.isFinite(filteredValue)) {
            throw new Error('filter arithmetic overflowed');
        }

        const nextFiltered = new Array<number>(this.historyLength);
        nextFiltered[0] = filteredValue;
        for (let index = 1; index < this.historyLength; index++) {
            nextFiltered[index] = this.filtered[index - 1];
        }

        const nextCorrelations = this.correlations.slice();
        for (let lag = 0; lag <= configuration.maxPeriod; lag++) {
            const windowSize =
                configuration.averagingLength !== 0 ? configuration.averagingLength : lag;
            let sx = 0.0;
            let sy = 0.0;
            let sxx = 0.0;
            let syy = 0.0;
            let sxy = 0.0;
            for (let offset = 0; offset < windowSize; offset++) {
                const xSample = nextFiltered[offset];
                const ySample = nextFiltered[lag + offset];
                sx += xSample;
                sy += ySample;
                sxx += xSample * xSample;
                sxy += xSample * ySample;
                syy += ySample * ySample;
            }
            const denominator =
                (windowSize * sxx - sx * sx) * (windowSize * syy - sy * sy);
            if (
                !Number.isFinite(sx) ||
                !Number.isFinite(sy) ||
                !Number.isFinite(sxx) ||
                !Number.isFinite(syy) ||
                !Number.isFinite(sxy) ||
                !Number.isFinite(denominator)
            ) {
                throw new Error('Pearson arithmetic overflowed');
            }
            if (denominator > 0.0) {
                nextCorrelations[lag] =
                    (windowSize * sxy - sx * sy) / Math.sqrt(denominator);
            }
            if (!Number.isFinite(nextCorrelations[lag])) {
                throw new Error('Pearson correlation overflowed');
            }
        }

        const nextPower = new Array<number>(this.binCount);
        const lagSpan = configuration.maxPeriod - DFT_LAG_START + 1;
        for (let binIndex = 0; binIndex < this.binCount; binIndex++) {
            let cosinePart = 0.0;
            let sinePart = 0.0;
            const cosRow = this.cosines[binIndex];
            const sinRow = this.sines[binIndex];
            for (let offset = 0; offset < lagSpan; offset++) {
                const corrValue = nextCorrelations[DFT_LAG_START + offset];
                cosinePart += corrValue * cosRow[offset];
                sinePart += corrValue * sinRow[offset];
            }
            const squaredSum = cosinePart * cosinePart + sinePart * sinePart;
            const smoothed =
                POWER_SMOOTHING * squaredSum * squaredSum +
                POWER_RETENTION * this.power[binIndex];
            if (!Number.isFinite(smoothed)) {
                throw new Error('spectral arithmetic overflowed');
            }
            nextPower[binIndex] = smoothed;
        }

        let maximum = isBook ? BOOK_AGC_DECAY * this.maximum : 0.0;
        for (const powerValue of nextPower) {
            if (powerValue > maximum) {
                maximum = powerValue;
            }
        }
        const normalized = nextPower.map((powerValue) =>
            maximum > 0.0 ? powerValue / maximum : 0.0,
        );

        const weights =
            !isBook && configuration.enhanceResolution
                ? normalized.map((normValue) => normValue * normValue * normValue)
                : normalized;
        const threshold = isBook ? BOOK_THRESHOLD : ARTICLE_THRESHOLD;
        let peak = weights[0];
        for (let binIndex = 1; binIndex < weights.length; binIndex++) {
            if (weights[binIndex] > peak) {
                peak = weights[binIndex];
            }
        }
        let sp = 0.0;
        let spx = 0.0;
        for (let binIndex = 0; binIndex < weights.length; binIndex++) {
            const weight = weights[binIndex];
            if (weight >= threshold && (isBook || peak >= ARTICLE_THRESHOLD)) {
                spx += (configuration.minPeriod + binIndex) * weight;
                sp += weight;
            }
        }

        let dominant = this.dominant;
        let hasDominant = this.hasDominant;
        if (sp !== 0.0 && (isBook || sp >= ARTICLE_THRESHOLD)) {
            dominant = spx / sp;
            hasDominant = true;
        }
        if (!Number.isFinite(dominant)) {
            throw new Error('dominant-cycle arithmetic overflowed');
        }

        const values =
            isBook && configuration.enhanceResolution
                ? normalized.map((normValue) => normValue * normValue)
                : weights.slice();

        this.close2 = this.close1;
        this.close1 = price;
        this.high2 = this.high1;
        this.high1 = high;
        this.filtered = nextFiltered;
        this.correlations = nextCorrelations;
        this.power = nextPower;
        this.maximum = maximum;
        this.dominant = dominant;
        this.hasDominant = hasDominant;
        this.count += 1;

        if (this.isPrimed()) {
            let valueMin = values[0];
            let valueMax = values[0];
            for (let binIndex = 1; binIndex < values.length; binIndex++) {
                const cell = values[binIndex];
                if (cell < valueMin) valueMin = cell;
                if (cell > valueMax) valueMax = cell;
            }
            const publicDominant = hasDominant ? dominant : Number.NaN;
            this.lastOutput = {
                heatmap: Heatmap.newHeatmap(
                    time,
                    configuration.minPeriod,
                    configuration.maxPeriod,
                    1.0,
                    valueMin,
                    valueMax,
                    values.slice(),
                ),
                dominantCycle: publicDominant,
            };
            return {
                heatmap: Heatmap.newHeatmap(
                    time,
                    configuration.minPeriod,
                    configuration.maxPeriod,
                    1.0,
                    valueMin,
                    valueMax,
                    values,
                ),
                dominantCycle: publicDominant,
            };
        }

        const emptyOutput: PeriodogramOutput = {
            heatmap: Heatmap.newEmptyHeatmap(
                time,
                configuration.minPeriod,
                configuration.maxPeriod,
                1.0,
            ),
            dominantCycle: Number.NaN,
        };
        this.lastOutput = emptyOutput;
        return emptyOutput;
    }

    private wrapEntity(time: Date, result: PeriodogramOutput): IndicatorOutput {
        const dcScalar = new Scalar();
        dcScalar.time = time;
        dcScalar.value = result.dominantCycle;
        return [result.heatmap, dcScalar];
    }

    public updateScalar(sample: Scalar): IndicatorOutput {
        return this.wrapEntity(sample.time, this.update(sample.value, sample.time));
    }

    public updateBar(sample: Bar): IndicatorOutput {
        return this.wrapEntity(
            sample.time,
            this.update(this.barComponentFunc(sample), sample.time),
        );
    }

    public updateQuote(sample: Quote): IndicatorOutput {
        return this.wrapEntity(
            sample.time,
            this.update(this.quoteComponentFunc(sample), sample.time),
        );
    }

    public updateTrade(sample: Trade): IndicatorOutput {
        return this.wrapEntity(
            sample.time,
            this.update(this.tradeComponentFunc(sample), sample.time),
        );
    }
}
