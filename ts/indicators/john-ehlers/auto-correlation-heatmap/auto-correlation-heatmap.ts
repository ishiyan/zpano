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
    AutoCorrelationHeatmapParams,
    autoCorrelationHeatmapMnemonic,
    defaultParams,
} from './params';

export {
    AutoCorrelationHeatmapParams,
    autoCorrelationHeatmapMnemonic,
    defaultParams,
};

const SQRT_TWO = Math.SQRT2;

const ALLOWED_KEYS = new Set<string>([
    'length',
    'smoothingPeriod',
    'maxLag',
    'barComponent',
    'quoteComponent',
    'tradeComponent',
]);

export interface AutocorrelationState {
    readonly acceptedSamples: number;
    readonly filtered: number;
    readonly correlations: readonly number[];
}

export interface ResolvedAutoCorrelationHeatmapParams {
    readonly length: number;
    readonly smoothingPeriod: number;
    readonly maxLag: number;
}

function validateIntegerMin(value: unknown, name: string, minimum: number): number {
    if (typeof value !== 'number' || !Number.isInteger(value) || value < minimum) {
        throw new Error(`${name} must be an integer >= ${minimum}`);
    }
    return value;
}

function validateIntegerRange(value: unknown, name: string, minimum: number, maximum: number): number {
    if (typeof value !== 'number' || !Number.isInteger(value) || value < minimum || value > maximum) {
        throw new Error(`${name} must be an integer in [${minimum}, ${maximum}]`);
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
        throw new Error('price must be finite or NaN (missing sample)');
    }
    if (!Number.isFinite(value) && !Number.isNaN(value)) {
        throw new Error('price must be finite or NaN (missing sample)');
    }
    return value;
}

/** Ehlers' February 2025 Autocorrelation Heatmap indicator (ACH). */
export class AutoCorrelationHeatmap implements Indicator {
    public readonly parameters: ResolvedAutoCorrelationHeatmapParams;
    private readonly historyLength: number;
    private readonly primeCount: number;
    private readonly coeffC1: number;
    private readonly coeffC2: number;
    private readonly coeffC3: number;
    private readonly mnemonicValue: string;
    private readonly descriptionValue: string;
    private readonly barComponentFunc: (bar: Bar) => number;
    private readonly quoteComponentFunc: (quote: Quote) => number;
    private readonly tradeComponentFunc: (trade: Trade) => number;

    private count = 0;
    private price1 = 0.0;
    private price2 = 0.0;
    private filtered1 = 0.0;
    private filtered2 = 0.0;
    private history: number[] = [];
    private correlations: number[] = [];
    private lastOutput!: Heatmap;

    public static default(): AutoCorrelationHeatmap {
        return new AutoCorrelationHeatmap(defaultParams());
    }

    public static fromParams(params?: AutoCorrelationHeatmapParams): AutoCorrelationHeatmap {
        return new AutoCorrelationHeatmap(params);
    }

    public constructor(params?: AutoCorrelationHeatmapParams) {
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

        const length = validateIntegerMin(params?.length ?? 20, 'length', 2);
        const smoothingPeriod = validateRealAtLeast(params?.smoothingPeriod ?? 20, 'smoothingPeriod', 1.0);
        const maxLag = validateIntegerRange(params?.maxLag ?? 99, 'maxLag', 0, 99);

        const angle = (SQRT_TWO * Math.PI) / smoothingPeriod;
        const attenuation = Math.exp(-angle);
        const coeffC2 = 2.0 * attenuation * Math.cos(angle);
        const coeffC3 = -attenuation * attenuation;
        const coeffC1 = (1.0 + coeffC2 - coeffC3) / 4.0;
        if (!Number.isFinite(coeffC1) || !Number.isFinite(coeffC2) || !Number.isFinite(coeffC3)) {
            throw new Error('smoothingPeriod produces non-finite filter coefficients');
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
            length,
            smoothingPeriod,
            maxLag,
        };
        this.historyLength = length + maxLag;
        this.primeCount = Math.max(4, this.historyLength);
        this.coeffC1 = coeffC1;
        this.coeffC2 = coeffC2;
        this.coeffC3 = coeffC3;

        this.barComponentFunc = barComponentValue(barComp);
        this.quoteComponentFunc = quoteComponentValue(quoteComp);
        this.tradeComponentFunc = tradeComponentValue(tradeComp);

        const resolvedParams: AutoCorrelationHeatmapParams = {
            length,
            smoothingPeriod,
            maxLag,
            barComponent: barComp,
            quoteComponent: quoteComp,
            tradeComponent: tradeComp,
        };
        this.mnemonicValue = autoCorrelationHeatmapMnemonic(resolvedParams);
        this.descriptionValue = `Autocorrelation heatmap ${this.mnemonicValue}`;

        this.reset();
    }

    public get length(): number {
        return this.parameters.length;
    }

    public get smoothingPeriod(): number {
        return this.parameters.smoothingPeriod;
    }

    public get maxLag(): number {
        return this.parameters.maxLag;
    }

    public get warmupBars(): number {
        return this.primeCount - 1;
    }

    public get samplesSeen(): number {
        return this.count;
    }

    public get state(): AutocorrelationState {
        return {
            acceptedSamples: this.count,
            filtered: this.filtered1,
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
        return Heatmap.newEmptyHeatmap(time, 0.0, this.parameters.maxLag, 1.0);
    }

    public reset(): void {
        this.count = 0;
        this.price1 = 0.0;
        this.price2 = 0.0;
        this.filtered1 = 0.0;
        this.filtered2 = 0.0;
        this.history = new Array<number>(this.historyLength).fill(0.0);
        this.correlations = new Array<number>(this.parameters.maxLag + 1).fill(0.0);
        this.lastOutput = this.emptyHeatmap(new Date(0));
    }

    public isPrimed(): boolean {
        return this.count >= this.primeCount;
    }

    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.AutoCorrelationHeatmap,
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

        let filtered: number;
        if (this.count < 3) {
            filtered = price;
        } else {
            filtered =
                (1.0 - this.coeffC1) * price +
                (2.0 * this.coeffC1 - this.coeffC2) * this.price1 -
                (this.coeffC1 + this.coeffC3) * this.price2 +
                this.coeffC2 * this.filtered1 +
                this.coeffC3 * this.filtered2;
        }
        if (!Number.isFinite(filtered)) {
            throw new Error('Ultimate Smoother calculation overflowed');
        }

        const nextHistory = new Array<number>(this.historyLength);
        nextHistory[0] = filtered;
        for (let offset = 1; offset < this.historyLength; offset++) {
            nextHistory[offset] = this.history[offset - 1];
        }

        const nextCorrelations = this.correlations.slice();
        const length = this.parameters.length;
        const binCount = this.parameters.maxLag + 1;

        for (let lag = 0; lag < binCount; lag++) {
            let sumX = 0.0;
            let sumY = 0.0;
            let sumXX = 0.0;
            let sumXY = 0.0;
            let sumYY = 0.0;
            for (let offset = 0; offset < length; offset++) {
                const sampleX = nextHistory[offset];
                const sampleY = nextHistory[lag + offset];
                sumX += sampleX;
                sumY += sampleY;
                sumXX += sampleX * sampleX;
                sumXY += sampleX * sampleY;
                sumYY += sampleY * sampleY;
            }
            const varX = length * sumXX - sumX * sumX;
            const varY = length * sumYY - sumY * sumY;
            if (
                !Number.isFinite(sumX) ||
                !Number.isFinite(sumY) ||
                !Number.isFinite(sumXX) ||
                !Number.isFinite(sumXY) ||
                !Number.isFinite(sumYY) ||
                !Number.isFinite(varX) ||
                !Number.isFinite(varY)
            ) {
                throw new Error('Pearson sums overflowed; reduce input magnitude');
            }
            if (varX > 0.0 && varY > 0.0) {
                const product = varX * varY;
                const numerator = length * sumXY - sumX * sumY;
                if (product <= 0.0 || !Number.isFinite(product) || !Number.isFinite(numerator)) {
                    throw new Error('Pearson denominator or numerator overflowed/underflowed');
                }
                const correlation = numerator / Math.sqrt(product);
                if (!Number.isFinite(correlation)) {
                    throw new Error('Pearson correlation is not finite');
                }
                nextCorrelations[lag] = correlation;
            }
        }

        this.price2 = this.price1;
        this.price1 = price;
        this.filtered2 = this.filtered1;
        this.filtered1 = filtered;
        this.history = nextHistory;
        this.correlations = nextCorrelations;
        this.count += 1;

        if (!this.isPrimed()) {
            this.lastOutput = this.emptyHeatmap(time);
            return this.emptyHeatmap(time);
        }

        let valueMin = nextCorrelations[0];
        let valueMax = nextCorrelations[0];
        for (let lag = 1; lag < binCount; lag++) {
            const item = nextCorrelations[lag];
            if (item < valueMin) {
                valueMin = item;
            }
            if (item > valueMax) {
                valueMax = item;
            }
        }

        this.lastOutput = Heatmap.newHeatmap(
            time,
            0.0,
            this.parameters.maxLag,
            1.0,
            valueMin,
            valueMax,
            nextCorrelations.slice(),
        );
        return Heatmap.newHeatmap(
            time,
            0.0,
            this.parameters.maxLag,
            1.0,
            valueMin,
            valueMax,
            nextCorrelations.slice(),
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
