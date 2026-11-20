import { BarComponent, DefaultBarComponent } from '../../../entities/bar-component';
import { DefaultQuoteComponent, QuoteComponent } from '../../../entities/quote-component';
import { DefaultTradeComponent, TradeComponent } from '../../../entities/trade-component';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';

/** Parameters for Ehlers' TransformedDiscreteFourierTransformSpectrum (January 2007 article). */
export interface TransformedDiscreteFourierTransformSpectrumParams {
    /** Shortest evaluated integer cycle period in bars (>= 8). Default: 8. */
    minPeriod?: number;

    /** Longest evaluated integer cycle period in bars (>= minPeriod). Default: 50. */
    maxPeriod?: number;

    /** Moving observation window length in cleaned samples (>= 1). Default: 50. */
    windowLength?: number;

    /** High-pass filter cutoff period in bars (> 4.0). Default: 40. */
    highpassPeriod?: number;

    /** Bar component used by updateBar(); undefined defaults to Close. */
    barComponent?: BarComponent;

    /** Quote component used by updateQuote(); undefined defaults to Mid. */
    quoteComponent?: QuoteComponent;

    /** Trade component used by updateTrade(); undefined defaults to Price. */
    tradeComponent?: TradeComponent;
}

/** Returns Ehlers' January 2007 article default parameters. */
export function defaultParams(): TransformedDiscreteFourierTransformSpectrumParams {
    return {
        minPeriod: 8,
        maxPeriod: 50,
        windowLength: 50,
        highpassPeriod: 40,
    };
}

/** Formats a numeric parameter cleanly without negative zero. */
export function formatNumber(value: number): string {
    const normalized = value === 0 ? 0 : value;
    return String(normalized);
}

function buildMnemonic(
    prefix: string,
    params?: TransformedDiscreteFourierTransformSpectrumParams,
): string {
    const minPeriod = params?.minPeriod ?? 8;
    const maxPeriod = params?.maxPeriod ?? 50;
    const windowLength = params?.windowLength ?? 50;
    const highpassPeriod = params?.highpassPeriod ?? 40;
    const barComp = params?.barComponent ?? DefaultBarComponent;
    const quoteComp = params?.quoteComponent ?? DefaultQuoteComponent;
    const tradeComp = params?.tradeComponent ?? DefaultTradeComponent;
    const compSuffix = componentTripleMnemonic(barComp, quoteComp, tradeComp);
    return `${prefix}(${minPeriod}, ${maxPeriod}, ${windowLength}, ${formatNumber(highpassPeriod)}${compSuffix})`;
}

/** Builds the primary heatmap mnemonic for TransformedDiscreteFourierTransformSpectrum. */
export function transformedDiscreteFourierTransformSpectrumMnemonic(
    params?: TransformedDiscreteFourierTransformSpectrumParams,
): string {
    return buildMnemonic('tdfts', params);
}

/** Builds the dominant-cycle mnemonic for TransformedDiscreteFourierTransformSpectrum. */
export function transformedDiscreteFourierTransformSpectrumDominantCycleMnemonic(
    params?: TransformedDiscreteFourierTransformSpectrumParams,
): string {
    return buildMnemonic('tdfts-dc', params);
}
