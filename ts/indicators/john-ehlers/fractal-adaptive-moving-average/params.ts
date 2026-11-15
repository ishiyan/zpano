import { BarComponent, DefaultBarComponent } from '../../../entities/bar-component';
import { DefaultQuoteComponent, QuoteComponent } from '../../../entities/quote-component';
import { DefaultTradeComponent, TradeComponent } from '../../../entities/trade-component';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';

/** Describes parameters to create an instance of Ehlers' FractalAdaptiveMovingAverage indicator. */
export interface FractalAdaptiveMovingAverageParams {
    /** Even integer lookback length (>= 2). Default: 16. */
    length?: number;

    /** Slowest boundary smoothing factor (finite, 0 < A <= 1). Default: 0.01. */
    slowestSmoothingFactor?: number;

    /** Alpha exponent scale mode: 1 for full-precision ln(A), 0 for literal 2005 -4.6 (requires A === 0.01). Default: 1. */
    coefficientMode?: number;

    /** Bar component used when updating with a bar sample; undefined defaults to Close. */
    barComponent?: BarComponent;

    /** Quote component used when updating with a quote sample; undefined defaults to Mid. */
    quoteComponent?: QuoteComponent;

    /** Trade component used when updating with a trade sample; undefined defaults to Price. */
    tradeComponent?: TradeComponent;
}

export type Params = FractalAdaptiveMovingAverageParams;

export function defaultParams(): FractalAdaptiveMovingAverageParams {
    return {
        length: 16,
        slowestSmoothingFactor: 0.01,
        coefficientMode: 1,
    };
}

export function formatNumber(value: number): string {
    const normalized = Object.is(value, -0) ? 0 : value;
    return String(normalized);
}

function formatMnemonic(prefix: string, params?: FractalAdaptiveMovingAverageParams): string {
    const length = params?.length ?? 16;
    const slowest = params?.slowestSmoothingFactor ?? 0.01;
    const mode = params?.coefficientMode ?? 1;
    const barComp = params?.barComponent ?? DefaultBarComponent;
    const quoteComp = params?.quoteComponent ?? DefaultQuoteComponent;
    const tradeComp = params?.tradeComponent ?? DefaultTradeComponent;
    const modeTag = mode === 0 ? ', mode=0' : '';
    const suffix = componentTripleMnemonic(barComp, quoteComp, tradeComp);
    return `${prefix}(${length}, ${formatNumber(slowest)}${modeTag}${suffix})`;
}

export function fractalAdaptiveMovingAverageMnemonic(params?: FractalAdaptiveMovingAverageParams): string {
    return formatMnemonic('frama', params);
}

export function fractalAdaptiveMovingAverageDimensionMnemonic(params?: FractalAdaptiveMovingAverageParams): string {
    return formatMnemonic('framaDim', params);
}
