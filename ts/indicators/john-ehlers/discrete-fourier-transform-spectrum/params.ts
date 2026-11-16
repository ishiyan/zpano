import { BarComponent, DefaultBarComponent } from '../../../entities/bar-component';
import { DefaultQuoteComponent, QuoteComponent } from '../../../entities/quote-component';
import { DefaultTradeComponent, TradeComponent } from '../../../entities/trade-component';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';

/** Parameters for Ehlers' DiscreteFourierTransformSpectrum (2013 Book Listing 9-1). */
export interface DiscreteFourierTransformSpectrumParams {
    /** Shortest evaluated integer cycle period in bars (>= 3). Default: 10. */
    minPeriod?: number;

    /** Longest evaluated integer cycle period in bars (>= minPeriod). Default: 48. */
    maxPeriod?: number;

    /** Sliding roofed DFT window length in bars (>= 2, independent of maxPeriod). Default: 48. */
    windowLength?: number;

    /** Two-pole roofing high-pass cutoff period in bars (>= 5.0). Default: 48. */
    highpassPeriod?: number;

    /** Two-pole SuperSmoother cutoff period in bars (>= 2.0). Default: 10. */
    smoothingPeriod?: number;

    /** Divide each cosine/sine contribution by period before summing/squaring. Default: true. */
    spectralDilationCompensation?: boolean;

    /** Cube normalized display heatmap intensities after centroid extraction. Default: false. */
    enhanceResolution?: boolean;

    /** Bar component used by updateBar(); undefined defaults to Close. */
    barComponent?: BarComponent;

    /** Quote component used by updateQuote(); undefined defaults to Mid. */
    quoteComponent?: QuoteComponent;

    /** Trade component used by updateTrade(); undefined defaults to Price. */
    tradeComponent?: TradeComponent;
}

/** Returns Ehlers' Listing 9-1 default parameters. */
export function defaultParams(): DiscreteFourierTransformSpectrumParams {
    return {
        minPeriod: 10,
        maxPeriod: 48,
        windowLength: 48,
        highpassPeriod: 48,
        smoothingPeriod: 10,
        spectralDilationCompensation: true,
        enhanceResolution: false,
    };
}

/** Formats a numeric parameter cleanly without negative zero. */
export function formatNumber(value: number): string {
    const normalized = value === 0 ? 0 : value;
    return String(normalized);
}

function buildMnemonic(
    prefix: string,
    includeEnhance: boolean,
    params?: DiscreteFourierTransformSpectrumParams,
): string {
    const minPeriod = params?.minPeriod ?? 10;
    const maxPeriod = params?.maxPeriod ?? 48;
    const windowLength = params?.windowLength ?? 48;
    const highpassPeriod = params?.highpassPeriod ?? 48;
    const smoothingPeriod = params?.smoothingPeriod ?? 10;
    const sdc = params?.spectralDilationCompensation ?? true;
    const enhance = params?.enhanceResolution ?? false;
    const barComp = params?.barComponent ?? DefaultBarComponent;
    const quoteComp = params?.quoteComponent ?? DefaultQuoteComponent;
    const tradeComp = params?.tradeComponent ?? DefaultTradeComponent;
    const compSuffix = componentTripleMnemonic(barComp, quoteComp, tradeComp);
    const sdcSuffix = sdc ? '' : ', no-sdc';
    const enhSuffix = includeEnhance && enhance ? ', enh' : '';
    return `${prefix}(${minPeriod}, ${maxPeriod}, ${windowLength}, ${formatNumber(highpassPeriod)}, ${formatNumber(smoothingPeriod)}${sdcSuffix}${enhSuffix}${compSuffix})`;
}

/** Builds the primary heatmap mnemonic for DiscreteFourierTransformSpectrum. */
export function discreteFourierTransformSpectrumMnemonic(
    params?: DiscreteFourierTransformSpectrumParams,
): string {
    return buildMnemonic('dftps', true, params);
}

/** Builds the dominant-cycle mnemonic for DiscreteFourierTransformSpectrum. */
export function discreteFourierTransformSpectrumDominantCycleMnemonic(
    params?: DiscreteFourierTransformSpectrumParams,
): string {
    return buildMnemonic('dftps-dc', false, params);
}
