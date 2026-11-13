import { BarComponent, DefaultBarComponent } from '../../../entities/bar-component';
import { DefaultQuoteComponent, QuoteComponent } from '../../../entities/quote-component';
import { DefaultTradeComponent, TradeComponent } from '../../../entities/trade-component';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';

/** Describes parameters to create an instance of Ehlers' SuperSmoother indicator. */
export interface SuperSmootherParams {
    /** Shortest cycle period in bars (>= 2). Default: 10. */
    shortestCyclePeriod?: number;

    /** Bar component used when updating with a bar sample; undefined defaults to Close. */
    barComponent?: BarComponent;

    /** Quote component used when updating with a quote sample; undefined defaults to Mid. */
    quoteComponent?: QuoteComponent;

    /** Trade component used when updating with a trade sample; undefined defaults to Price. */
    tradeComponent?: TradeComponent;
}

export type Params = SuperSmootherParams;

export function defaultParams(): SuperSmootherParams {
    return { shortestCyclePeriod: 10 };
}

export function formatNumber(value: number): string {
    const normalized = Object.is(value, -0) ? 0 : value;
    return String(normalized);
}

export function superSmootherMnemonic(params?: SuperSmootherParams): string {
    const shortest = params?.shortestCyclePeriod ?? 10;
    const barComp = params?.barComponent ?? DefaultBarComponent;
    const quoteComp = params?.quoteComponent ?? DefaultQuoteComponent;
    const tradeComp = params?.tradeComponent ?? DefaultTradeComponent;
    const suffix = componentTripleMnemonic(barComp, quoteComp, tradeComp);
    return `ss(${formatNumber(shortest)}${suffix})`;
}
