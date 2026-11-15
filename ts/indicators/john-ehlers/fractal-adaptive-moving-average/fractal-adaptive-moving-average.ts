import { Bar } from '../../../entities/bar';
import { DefaultBarComponent, barComponentValue } from '../../../entities/bar-component';
import { Quote } from '../../../entities/quote';
import { DefaultQuoteComponent, quoteComponentValue } from '../../../entities/quote-component';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { DefaultTradeComponent, tradeComponentValue } from '../../../entities/trade-component';
import { buildMetadata } from '../../core/build-metadata';
import { Indicator } from '../../core/indicator';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import {
    FractalAdaptiveMovingAverageParams,
    defaultParams,
    fractalAdaptiveMovingAverageDimensionMnemonic,
    fractalAdaptiveMovingAverageMnemonic,
} from './params';

const LOG_TWO = Math.log(2);

function checkRealNumber(value: unknown, name: string, allowNan = false): number {
    if (typeof value !== 'number') {
        throw new Error(`${name} must be a real number`);
    }
    if (!Number.isFinite(value) && !(allowNan && Number.isNaN(value))) {
        throw new Error(`${name} must be finite${allowNan ? ' or NaN' : ''}`);
    }
    return value;
}

function boxRange(highs: number[], lows: number[], start: number, end: number): number {
    let highest = highs[start];
    let lowest = lows[start];
    for (let index = start + 1; index < end; index++) {
        if (highs[index] > highest) {
            highest = highs[index];
        }
        if (lows[index] < lowest) {
            lowest = lows[index];
        }
    }
    const span = highest - lowest;
    if (!Number.isFinite(span)) {
        throw new Error('price-range arithmetic overflowed');
    }
    return span;
}

/**
 * Ehlers' Fractal Adaptive Moving Average (FRAMA).
 *
 * Returns both the adaptive EMA value and the raw guarded 2005 range-based
 * fractal-dimension estimate used to adapt it.
 */
export class FractalAdaptiveMovingAverage implements Indicator {
    private readonly length: number;
    private readonly halfLength: number;
    private readonly warmupCount: number;
    private readonly slowestSmoothingFactor: number;
    private readonly coefficientMode: number;
    private readonly exponentScale: number;
    private readonly mnemonic: string;
    private readonly description: string;
    private readonly mnemonicFdim: string;
    private readonly descriptionFdim: string;
    private readonly barComponentFunc: (bar: Bar) => number;
    private readonly quoteComponentFunc: (quote: Quote) => number;
    private readonly tradeComponentFunc: (trade: Trade) => number;

    private count = 0;
    private highs: number[] = [];
    private lows: number[] = [];
    private dimensionState = 0;
    private alphaState = 1;
    private filterState = 0;
    private currentValue = Number.NaN;
    private currentDimension = Number.NaN;

    public static default(): FractalAdaptiveMovingAverage {
        return new FractalAdaptiveMovingAverage(defaultParams());
    }

    public static fromParams(params?: FractalAdaptiveMovingAverageParams): FractalAdaptiveMovingAverage {
        return new FractalAdaptiveMovingAverage(params);
    }

    public constructor(params?: FractalAdaptiveMovingAverageParams) {
        const length = params?.length ?? 16;
        if (typeof length !== 'number' || !Number.isInteger(length) || length < 2 || length % 2 !== 0) {
            throw new Error('length must be an even integer >= 2');
        }

        const slowest = checkRealNumber(params?.slowestSmoothingFactor ?? 0.01, 'slowestSmoothingFactor');
        if (!(slowest > 0 && slowest <= 1)) {
            throw new Error('slowestSmoothingFactor must satisfy 0 < A <= 1');
        }

        const mode = params?.coefficientMode ?? 1;
        if (typeof mode !== 'number' || !Number.isInteger(mode) || (mode !== 0 && mode !== 1)) {
            throw new Error('coefficientMode must be 0 (article) or 1 (full precision)');
        }
        if (mode === 0 && slowest !== 0.01) {
            throw new Error('literal article mode requires slowestSmoothingFactor === 0.01');
        }

        this.length = length;
        this.halfLength = Math.floor(length / 2);
        this.warmupCount = length - 1;
        this.slowestSmoothingFactor = slowest;
        this.coefficientMode = mode;
        this.exponentScale = mode === 1 ? Math.log(slowest) : -4.6;

        const barComp = params?.barComponent ?? DefaultBarComponent;
        const quoteComp = params?.quoteComponent ?? DefaultQuoteComponent;
        const tradeComp = params?.tradeComponent ?? DefaultTradeComponent;

        this.barComponentFunc = barComponentValue(barComp);
        this.quoteComponentFunc = quoteComponentValue(quoteComp);
        this.tradeComponentFunc = tradeComponentValue(tradeComp);

        const normalizedParams: FractalAdaptiveMovingAverageParams = {
            length,
            slowestSmoothingFactor: slowest,
            coefficientMode: mode,
            barComponent: params?.barComponent,
            quoteComponent: params?.quoteComponent,
            tradeComponent: params?.tradeComponent,
        };
        this.mnemonic = fractalAdaptiveMovingAverageMnemonic(normalizedParams);
        this.mnemonicFdim = fractalAdaptiveMovingAverageDimensionMnemonic(normalizedParams);

        const prefix = 'Fractal adaptive moving average ';
        this.description = prefix + this.mnemonic;
        this.descriptionFdim = prefix + this.mnemonicFdim;

        this.reset();
    }

    /** Clears all streaming histories and priming state while preserving configuration. */
    public reset(): void {
        this.count = 0;
        this.highs = [];
        this.lows = [];
        this.dimensionState = 0;
        this.alphaState = 1;
        this.filterState = 0;
        this.currentValue = Number.NaN;
        this.currentDimension = Number.NaN;
    }

    /** Returns the number of initial NaN outputs before priming. */
    public warmupBars(): number {
        return this.warmupCount;
    }

    /** Returns the number of accepted finite observations seen since start or reset. */
    public samplesSeen(): number {
        return this.count;
    }

    /** Returns the latest accepted FRAMA value (NaN before priming). */
    public value(): number {
        return this.currentValue;
    }

    /** Returns the latest accepted raw fractal-dimension estimate (NaN before priming). */
    public fractalDimension(): number {
        return this.currentDimension;
    }

    /** Returns the latest diagnostic smoothing factor (NaN before priming). */
    public alpha(): number {
        return this.isPrimed() ? this.alphaState : Number.NaN;
    }

    /** Indicates whether an indicator is primed. */
    public isPrimed(): boolean {
        return this.count >= this.length;
    }

    /** Describes the output data of the indicator. */
    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.FractalAdaptiveMovingAverage,
            this.mnemonic,
            this.description,
            [
                { mnemonic: this.mnemonic, description: this.description },
                { mnemonic: this.mnemonicFdim, description: this.descriptionFdim },
            ],
        );
    }

    /** Updates an indicator given the next scalar sample. */
    public updateScalar(sample: Scalar): IndicatorOutput {
        const extracted = sample.value;
        return this.updateEntity(sample.time, extracted, extracted, extracted);
    }

    /** Updates an indicator given the next bar sample. */
    public updateBar(sample: Bar): IndicatorOutput {
        const extracted = this.barComponentFunc(sample);
        return this.updateEntity(sample.time, extracted, sample.high, sample.low);
    }

    /** Updates an indicator given the next quote sample. */
    public updateQuote(sample: Quote): IndicatorOutput {
        const extracted = this.quoteComponentFunc(sample);
        return this.updateEntity(sample.time, extracted, sample.askPrice, sample.bidPrice);
    }

    /** Updates an indicator given the next trade sample. */
    public updateTrade(sample: Trade): IndicatorOutput {
        const extracted = this.tradeComponentFunc(sample);
        return this.updateEntity(sample.time, extracted, extracted, extracted);
    }

    /** Consumes one scalar or (price, high, low) sample and returns the FRAMA value. */
    public update(price: number, high?: number, low?: number): number {
        const checkedPrice = checkRealNumber(price, 'price', true);
        if ((high === undefined) !== (low === undefined)) {
            throw new Error('supply both high and low, or neither');
        }
        let checkedHigh: number;
        let checkedLow: number;
        if (high === undefined || low === undefined) {
            checkedHigh = checkedPrice;
            checkedLow = checkedPrice;
        } else {
            checkedHigh = checkRealNumber(high, 'high', true);
            checkedLow = checkRealNumber(low, 'low', true);
        }

        if (Number.isNaN(checkedPrice) || Number.isNaN(checkedHigh) || Number.isNaN(checkedLow)) {
            return Number.NaN;
        }
        if (checkedHigh < checkedLow) {
            throw new Error('high must be >= low');
        }

        const nextHighs = [checkedHigh, ...this.highs.slice(0, this.length - 1)];
        const nextLows = [checkedLow, ...this.lows.slice(0, this.length - 1)];
        const nextCount = this.count + 1;
        let dimension = this.dimensionState;
        let alpha = this.alphaState;

        if (nextCount >= this.length) {
            const n1 = boxRange(nextHighs, nextLows, 0, this.halfLength) / this.halfLength;
            const n2 = boxRange(nextHighs, nextLows, this.halfLength, this.length) / this.halfLength;
            const n3 = boxRange(nextHighs, nextLows, 0, this.length) / this.length;

            if (n1 > 0 && n2 > 0 && n3 > 0) {
                const combined = n1 + n2;
                if (!Number.isFinite(combined)) {
                    throw new Error('combined box-range arithmetic overflowed');
                }
                dimension = (Math.log(combined) - Math.log(n3)) / LOG_TWO;
                if (!Number.isFinite(dimension)) {
                    throw new Error('dimension arithmetic overflowed');
                }
            }

            if (dimension <= 1) {
                alpha = 1;
            } else {
                alpha = Math.exp(this.exponentScale * (dimension - 1));
                alpha = Math.min(1, Math.max(this.slowestSmoothingFactor, alpha));
            }
        }

        let nextValue: number;
        if (nextCount <= this.length) {
            nextValue = checkedPrice;
        } else {
            nextValue = alpha * checkedPrice + (1 - alpha) * this.filterState;
            if (!Number.isFinite(nextValue)) {
                throw new Error('FRAMA arithmetic overflowed');
            }
        }

        this.highs = nextHighs;
        this.lows = nextLows;
        this.count = nextCount;
        this.dimensionState = dimension;
        this.alphaState = alpha;
        this.filterState = nextValue;
        if (nextCount >= this.length) {
            this.currentValue = nextValue;
            this.currentDimension = dimension;
            return nextValue;
        }
        return Number.NaN;
    }

    private updateEntity(
        time: Date,
        sample: number,
        sampleHigh: number,
        sampleLow: number,
    ): IndicatorOutput {
        const frama = this.update(sample, sampleHigh, sampleLow);
        const fdim = Number.isNaN(frama) ? Number.NaN : this.currentDimension;
        return [
            new Scalar({ time, value: frama }),
            new Scalar({ time, value: fdim }),
        ];
    }
}
