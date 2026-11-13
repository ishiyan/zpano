import { BarComponent, DefaultBarComponent } from '../../../entities/bar-component';
import { DefaultQuoteComponent, QuoteComponent } from '../../../entities/quote-component';
import { DefaultTradeComponent, TradeComponent } from '../../../entities/trade-component';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';

/** Describes parameters to create an instance of Ehlers' RoofingFilter indicator. */
export interface RoofingFilterParams {
    /** Shortest (SuperSmoother low-pass) cycle period in bars (>= 2). Default: 10. */
    shortestCyclePeriod?: number;

    /** Longest (high-pass) cycle period in bars (> shortestCyclePeriod). Default: 48. */
    longestCyclePeriod?: number;

    /** Number of poles in the initial high-pass stage (integer 1 or 2). Default: 2. */
    highpassPoles?: number;

    /** Enables a post-smoothing 1-pole high-pass stage (requires highpassPoles === 1). Default: false. */
    zeroMean?: boolean;

    /** Bar component used when updating with a bar sample; undefined defaults to Close. */
    barComponent?: BarComponent;

    /** Quote component used when updating with a quote sample; undefined defaults to Mid. */
    quoteComponent?: QuoteComponent;

    /** Trade component used when updating with a trade sample; undefined defaults to Price. */
    tradeComponent?: TradeComponent;
}

export type Params = RoofingFilterParams;

export function defaultParams(): RoofingFilterParams {
    return {
        shortestCyclePeriod: 10,
        longestCyclePeriod: 48,
        highpassPoles: 2,
        zeroMean: false,
    };
}

export function formatNumber(value: number): string {
    const normalized = Object.is(value, -0) ? 0 : value;
    return String(normalized);
}

export function roofingFilterMnemonic(params?: RoofingFilterParams): string {
    const shortest = params?.shortestCyclePeriod ?? 10;
    const longest = params?.longestCyclePeriod ?? 48;
    const poles = params?.highpassPoles ?? 2;
    const zeroMean = params?.zeroMean ?? false;
    const barComp = params?.barComponent ?? DefaultBarComponent;
    const quoteComp = params?.quoteComponent ?? DefaultQuoteComponent;
    const tradeComp = params?.tradeComponent ?? DefaultTradeComponent;
    const suffix = componentTripleMnemonic(barComp, quoteComp, tradeComp);
    const zmTag = zeroMean ? 'zm' : '';
    return `roof${poles}hp${zmTag}(${formatNumber(shortest)}, ${formatNumber(longest)}${suffix})`;
}
