import { BarComponent, DefaultBarComponent } from '../../../entities/bar-component';
import { DefaultQuoteComponent, QuoteComponent } from '../../../entities/quote-component';
import { DefaultTradeComponent, TradeComponent } from '../../../entities/trade-component';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';

/** Parameters for Ehlers' 2013 RoofedAutoCorrelationHeatmap indicator (Book Listing 8-2). */
export interface RoofedAutoCorrelationHeatmapParams {
    /** First displayed correlation lag on the heatmap axis (>= 0). Default: 3. */
    minLag?: number;

    /** Last displayed correlation lag on the heatmap axis (>= minLag). Default: 48. */
    maxLag?: number;

    /** Sample pairs per Pearson correlation comparison (>= 0; 0 selects M = lag). Default: 0. */
    averagingLength?: number;

    /** Two-pole roofing high-pass cutoff period in bars (>= 5.0). Default: 48. */
    highpassPeriod?: number;

    /** Two-pole SuperSmoother cutoff period in bars (>= 2.0). Default: 10. */
    smoothingPeriod?: number;

    /** Bar component used by updateBar(); undefined defaults to Close. */
    barComponent?: BarComponent;

    /** Quote component used by updateQuote(); undefined defaults to Mid. */
    quoteComponent?: QuoteComponent;

    /** Trade component used by updateTrade(); undefined defaults to Price. */
    tradeComponent?: TradeComponent;
}

/** Returns Ehlers' 2013 Book Listing 8-2 default parameters. */
export function defaultParams(): RoofedAutoCorrelationHeatmapParams {
    return {
        minLag: 3,
        maxLag: 48,
        averagingLength: 0,
        highpassPeriod: 48,
        smoothingPeriod: 10,
    };
}

/** Formats a numeric parameter cleanly without negative zero. */
export function formatNumber(value: number): string {
    const normalized = value === 0 ? 0 : value;
    return String(normalized);
}

/** Builds the mnemonic for RoofedAutoCorrelationHeatmap. */
export function roofedAutoCorrelationHeatmapMnemonic(
    params?: RoofedAutoCorrelationHeatmapParams,
): string {
    const minLag = params?.minLag ?? 3;
    const maxLag = params?.maxLag ?? 48;
    const averagingLength = params?.averagingLength ?? 0;
    const highpassPeriod = params?.highpassPeriod ?? 48;
    const smoothingPeriod = params?.smoothingPeriod ?? 10;
    const barComp = params?.barComponent ?? DefaultBarComponent;
    const quoteComp = params?.quoteComponent ?? DefaultQuoteComponent;
    const tradeComp = params?.tradeComponent ?? DefaultTradeComponent;
    const compSuffix = componentTripleMnemonic(barComp, quoteComp, tradeComp);
    return `rach(${minLag}, ${maxLag}, ${averagingLength}, ${formatNumber(highpassPeriod)}, ${formatNumber(smoothingPeriod)}${compSuffix})`;
}
