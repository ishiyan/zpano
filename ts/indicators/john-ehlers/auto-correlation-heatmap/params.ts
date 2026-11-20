import { BarComponent, DefaultBarComponent } from '../../../entities/bar-component';
import { DefaultQuoteComponent, QuoteComponent } from '../../../entities/quote-component';
import { DefaultTradeComponent, TradeComponent } from '../../../entities/trade-component';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';

/** Parameters for Ehlers' February 2025 AutoCorrelationHeatmap indicator. */
export interface AutoCorrelationHeatmapParams {
    /** Number of aligned sample pairs per Pearson correlation comparison (>= 2). Default: 20. */
    length?: number;

    /** Ultimate Smoother critical period in bars (>= 1.0). Default: 20. */
    smoothingPeriod?: number;

    /** Inclusive upper correlation lag in [0, 99]; minimum lag is always zero. Default: 99. */
    maxLag?: number;

    /** Bar component used by updateBar(); undefined defaults to Close. */
    barComponent?: BarComponent;

    /** Quote component used by updateQuote(); undefined defaults to Mid. */
    quoteComponent?: QuoteComponent;

    /** Trade component used by updateTrade(); undefined defaults to Price. */
    tradeComponent?: TradeComponent;
}

/** Returns Ehlers' February 2025 article default parameters. */
export function defaultParams(): AutoCorrelationHeatmapParams {
    return {
        length: 20,
        smoothingPeriod: 20,
        maxLag: 99,
    };
}

/** Formats a numeric parameter cleanly without negative zero. */
export function formatNumber(value: number): string {
    const normalized = value === 0 ? 0 : value;
    return String(normalized);
}

/** Builds the mnemonic for AutoCorrelationHeatmap. */
export function autoCorrelationHeatmapMnemonic(
    params?: AutoCorrelationHeatmapParams,
): string {
    const length = params?.length ?? 20;
    const smoothingPeriod = params?.smoothingPeriod ?? 20;
    const maxLag = params?.maxLag ?? 99;
    const barComp = params?.barComponent ?? DefaultBarComponent;
    const quoteComp = params?.quoteComponent ?? DefaultQuoteComponent;
    const tradeComp = params?.tradeComponent ?? DefaultTradeComponent;
    const compSuffix = componentTripleMnemonic(barComp, quoteComp, tradeComp);
    return `ach(${length}, ${formatNumber(smoothingPeriod)}, ${maxLag}${compSuffix})`;
}
