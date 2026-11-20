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
    RoofedAutoCorrelationHeatmapParams,
    defaultParams,
    roofedAutoCorrelationHeatmapMnemonic,
} from './params';

export {
    RoofedAutoCorrelationHeatmapParams,
    defaultParams,
    roofedAutoCorrelationHeatmapMnemonic,
};

const TWO_PI = 2.0 * Math.PI;
const SQRT_TWO = Math.SQRT2;

const ALLOWED_KEYS = new Set<string>([
    'minLag',
    'maxLag',
    'averagingLength',
    'highpassPeriod',
    'smoothingPeriod',
    'barComponent',
    'quoteComponent',
    'tradeComponent',
]);

export interface AutocorrelationState {
    readonly highPass: number;
    readonly filtered: number;
    readonly correlations: readonly number[];
}

export interface ResolvedRoofedAutoCorrelationHeatmapParams {
    readonly minLag: number;
    readonly maxLag: number;
    readonly averagingLength: number;
    readonly highpassPeriod: number;
    readonly smoothingPeriod: number;
}

function validateIntegerMin(value: unknown, name: string, minimum: number): number {
    if (typeof value !== 'number' || !Number.isInteger(value) || value < minimum) {
        throw new Error(`${name} must be an integer >= ${minimum}`);
    }
    return value;
}

function validateRealAtLeast(value: unknown, name: string, minimum: number): number {
    if (typeof value !== 'number' || !Number.isFinite(value) || value < minimum) {
        throw new Error(`${name} must be a finite real >= ${minimum}`);
    }
    return value;
}

function validateSample(value: unknown): number {
    if (typeof value !== 'number') {
        throw new Error('price must be finite or NaN (missing)');
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
        const hp = validateRealAtLeast(highpassPeriod, 'highpassPeriod', 5.0);
        const lp = validateRealAtLeast(smoothingPeriod, 'smoothingPeriod', 2.0);
        const angle = TWO_PI / (SQRT_TWO * hp);
        const cosine = Math.cos(angle);
        const alpha = (cosine + Math.sin(angle) - 1.0) / cosine;
        const pole = 1.0 - alpha;
        const attenuation = Math.exp((-SQRT_TWO * Math.PI) / lp);
        this.hp0 = (1.0 - alpha / 2.0) * (1.0 - alpha / 2.0);
        this.hp1 = 2.0 * pole;
        this.hp2 = pole * pole;
        this.ss2 = 2.0 * attenuation * Math.cos((SQRT_TWO * Math.PI) / lp);
        this.ss3 = -attenuation * attenuation;
        this.ss1 = 1.0 - this.ss2 - this.ss3;
        if (
            !(pole > 0.0 && pole < 1.0 && attenuation > 0.0 && attenuation < 1.0 && this.ss1 > 0.0) ||
            !Number.isFinite(this.hp0) ||
            !Number.isFinite(this.hp1) ||
            !Number.isFinite(this.hp2) ||
            !Number.isFinite(this.ss1) ||
            !Number.isFinite(this.ss2) ||
            !Number.isFinite(this.ss3)
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

    public candidate(price: number): { readonly high: number; readonly filtered: number } {
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
        return { high, filtered };
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

/** Ehlers' 2013 Roofed Autocorrelation Heatmap indicator (RACH, Book Listing 8-2). */
export class RoofedAutoCorrelationHeatmap implements Indicator {
    public readonly parameters: ResolvedRoofedAutoCorrelationHeatmapParams;
    private readonly historyLength: number;
    private readonly roof: RoofFilter;
    private readonly mnemonicValue: string;
    private readonly descriptionValue: string;
    private readonly barComponentFunc: (bar: Bar) => number;
    private readonly quoteComponentFunc: (quote: Quote) => number;
    private readonly tradeComponentFunc: (trade: Trade) => number;

    private count = 0;
    private history: number[] = [];
    private correlations: number[] = [];
    private lastOutput!: Heatmap;

    public static default(): RoofedAutoCorrelationHeatmap {
        return new RoofedAutoCorrelationHeatmap(defaultParams());
    }

    public static fromParams(
        params?: RoofedAutoCorrelationHeatmapParams,
    ): RoofedAutoCorrelationHeatmap {
        return new RoofedAutoCorrelationHeatmap(params);
    }

    public constructor(params?: RoofedAutoCorrelationHeatmapParams) {
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

        const minLag = validateIntegerMin(params?.minLag ?? 3, 'minLag', 0);
        const maxLag = validateIntegerMin(params?.maxLag ?? 48, 'maxLag', minLag);
        const averagingLength = validateIntegerMin(
            params?.averagingLength ?? 0,
            'averagingLength',
            0,
        );
        const highpassPeriod = validateRealAtLeast(
            params?.highpassPeriod ?? 48.0,
            'highpassPeriod',
            5.0,
        );
        const smoothingPeriod = validateRealAtLeast(
            params?.smoothingPeriod ?? 10.0,
            'smoothingPeriod',
            2.0,
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
            minLag,
            maxLag,
            averagingLength,
            highpassPeriod,
            smoothingPeriod,
        };
        const effectiveWindow = averagingLength > 0 ? averagingLength : maxLag;
        this.historyLength = Math.max(1, maxLag + effectiveWindow);

        this.barComponentFunc = barComponentValue(barComp);
        this.quoteComponentFunc = quoteComponentValue(quoteComp);
        this.tradeComponentFunc = tradeComponentValue(tradeComp);

        const resolvedParams: RoofedAutoCorrelationHeatmapParams = {
            minLag,
            maxLag,
            averagingLength,
            highpassPeriod,
            smoothingPeriod,
            barComponent: barComp,
            quoteComponent: quoteComp,
            tradeComponent: tradeComp,
        };
        this.mnemonicValue = roofedAutoCorrelationHeatmapMnemonic(resolvedParams);
        this.descriptionValue = `Roofed autocorrelation heatmap ${this.mnemonicValue}`;

        this.reset();
    }

    public get minLag(): number {
        return this.parameters.minLag;
    }

    public get maxLag(): number {
        return this.parameters.maxLag;
    }

    public get minPeriod(): number {
        return this.parameters.minLag;
    }

    public get maxPeriod(): number {
        return this.parameters.maxLag;
    }

    public get averagingLength(): number {
        return this.parameters.averagingLength;
    }

    public get highpassPeriod(): number {
        return this.parameters.highpassPeriod;
    }

    public get smoothingPeriod(): number {
        return this.parameters.smoothingPeriod;
    }

    public get warmupBars(): number {
        return this.historyLength - 1;
    }

    public get samplesSeen(): number {
        return this.count;
    }

    public get state(): AutocorrelationState {
        return {
            highPass: this.roof.h1,
            filtered: this.roof.f1,
            correlations: this.correlations.slice(),
        };
    }

    public get output(): Heatmap {
        if (this.lastOutput.isEmpty()) {
            return Heatmap.newEmptyHeatmap(
                this.lastOutput.time,
                this.lastOutput.parameterFirst,
                this.lastOutput.parameterLast,
                this.lastOutput.parameterResolution,
            );
        }
        return Heatmap.newHeatmap(
            this.lastOutput.time,
            this.lastOutput.parameterFirst,
            this.lastOutput.parameterLast,
            this.lastOutput.parameterResolution,
            this.lastOutput.valueMin,
            this.lastOutput.valueMax,
            this.lastOutput.values.slice(),
        );
    }

    private emptyHeatmap(time: Date): Heatmap {
        return Heatmap.newEmptyHeatmap(
            time,
            this.parameters.minLag,
            this.parameters.maxLag,
            1.0,
        );
    }

    public reset(): void {
        this.roof.reset();
        this.count = 0;
        this.history = new Array<number>(this.historyLength).fill(0.0);
        this.correlations = new Array<number>(this.parameters.maxLag + 1).fill(0.0);
        this.lastOutput = this.emptyHeatmap(new Date(0));
    }

    public isPrimed(): boolean {
        return this.count >= this.historyLength;
    }

    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.RoofedAutoCorrelationHeatmap,
            this.mnemonicValue,
            this.descriptionValue,
            [{ mnemonic: this.mnemonicValue, description: this.descriptionValue }],
        );
    }

    public update(sample: number, time: Date = new Date(0)): Heatmap {
        const price = validateSample(sample);
        if (Number.isNaN(price)) {
            return this.emptyHeatmap(time);
        }

        const { high, filtered } = this.roof.candidate(price);

        const nextHistory = new Array<number>(this.historyLength);
        nextHistory[0] = filtered;
        for (let offset = 1; offset < this.historyLength; offset++) {
            nextHistory[offset] = this.history[offset - 1];
        }

        const nextCorrelations = this.correlations.slice();
        const binCount = this.parameters.maxLag + 1;

        for (let lag = 0; lag < binCount; lag++) {
            const window =
                this.parameters.averagingLength > 0
                    ? this.parameters.averagingLength
                    : lag;
            let sumX = 0.0;
            let sumY = 0.0;
            let sumXX = 0.0;
            let sumXY = 0.0;
            let sumYY = 0.0;
            for (let offset = 0; offset < window; offset++) {
                const sampleX = nextHistory[offset];
                const sampleY = nextHistory[lag + offset];
                sumX += sampleX;
                sumY += sampleY;
                sumXX += sampleX * sampleX;
                sumXY += sampleX * sampleY;
                sumYY += sampleY * sampleY;
            }
            const denominator =
                (window * sumXX - sumX * sumX) * (window * sumYY - sumY * sumY);
            if (
                !Number.isFinite(sumX) ||
                !Number.isFinite(sumY) ||
                !Number.isFinite(sumXX) ||
                !Number.isFinite(sumXY) ||
                !Number.isFinite(sumYY) ||
                !Number.isFinite(denominator)
            ) {
                throw new Error('Pearson arithmetic overflowed');
            }
            if (denominator > 0.0) {
                nextCorrelations[lag] =
                    (window * sumXY - sumX * sumY) / Math.sqrt(denominator);
            }
            nextCorrelations[lag] = 0.5 * (nextCorrelations[lag] + 1.0);
            if (!Number.isFinite(nextCorrelations[lag])) {
                throw new Error('correlation arithmetic overflowed');
            }
        }

        this.roof.commit(price, high, filtered);
        this.history = nextHistory;
        this.correlations = nextCorrelations;
        this.count += 1;

        if (!this.isPrimed()) {
            this.lastOutput = this.emptyHeatmap(time);
            return this.emptyHeatmap(time);
        }

        const displayed = nextCorrelations.slice(
            this.parameters.minLag,
            this.parameters.maxLag + 1,
        );
        let valueMin = displayed[0];
        let valueMax = displayed[0];
        for (let index = 1; index < displayed.length; index++) {
            const item = displayed[index];
            if (item < valueMin) {
                valueMin = item;
            }
            if (item > valueMax) {
                valueMax = item;
            }
        }

        this.lastOutput = Heatmap.newHeatmap(
            time,
            this.parameters.minLag,
            this.parameters.maxLag,
            1.0,
            valueMin,
            valueMax,
            displayed.slice(),
        );
        return Heatmap.newHeatmap(
            time,
            this.parameters.minLag,
            this.parameters.maxLag,
            1.0,
            valueMin,
            valueMax,
            displayed,
        );
    }

    public updateScalar(sample: Scalar): IndicatorOutput {
        return [this.update(sample.value, sample.time)];
    }

    public updateBar(sample: Bar): IndicatorOutput {
        return [this.update(this.barComponentFunc(sample), sample.time)];
    }

    public updateQuote(sample: Quote): IndicatorOutput {
        return [this.update(this.quoteComponentFunc(sample), sample.time)];
    }

    public updateTrade(sample: Trade): IndicatorOutput {
        return [this.update(this.tradeComponentFunc(sample), sample.time)];
    }
}
