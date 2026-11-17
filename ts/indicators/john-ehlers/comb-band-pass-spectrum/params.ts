import { BarComponent, DefaultBarComponent } from '../../../entities/bar-component';
import { DefaultQuoteComponent, QuoteComponent } from '../../../entities/quote-component';
import { DefaultTradeComponent, TradeComponent } from '../../../entities/trade-component';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';

/** Parameters for Ehlers' CombBandPassSpectrum (2013 Book Listing 10-1). */
export interface CombBandPassSpectrumParams {
    /** Shortest evaluated integer cycle period in bars (>= 3). Default: 10. */
    minPeriod?: number;

    /** Longest evaluated integer cycle period in bars (>= minPeriod). Default: 48. */
    maxPeriod?: number;

    /** Fractional bandwidth of each band-pass filter (0 < bandwidth < minPeriod / 4). Default: 0.3. */
    bandwidth?: number;

    /** Two-pole roofing high-pass cutoff period in bars (>= 5.0). Default: 48. */
    highpassPeriod?: number;

    /** Two-pole SuperSmoother cutoff period in bars (>= 2.0). Default: 10. */
    smoothingPeriod?: number;

    /** Divide each band-pass amplitude by period before squaring and summing. Default: true. */
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

/** Returns Ehlers' Listing 10-1 default parameters. */
export function defaultParams(): CombBandPassSpectrumParams {
    return {
        minPeriod: 10,
        maxPeriod: 48,
        bandwidth: 0.3,
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
    params?: CombBandPassSpectrumParams,
): string {
    const minPeriod = params?.minPeriod ?? 10;
    const maxPeriod = params?.maxPeriod ?? 48;
    const bandwidth = params?.bandwidth ?? 0.3;
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
    return `${prefix}(${minPeriod}, ${maxPeriod}, ${formatNumber(bandwidth)}, ${formatNumber(highpassPeriod)}, ${formatNumber(smoothingPeriod)}${sdcSuffix}${enhSuffix}${compSuffix})`;
}

/** Builds the primary heatmap mnemonic for CombBandPassSpectrum. */
export function combBandPassSpectrumMnemonic(
    params?: CombBandPassSpectrumParams,
): string {
    return buildMnemonic('cbps', true, params);
}

/** Builds the dominant-cycle mnemonic for CombBandPassSpectrum. */
export function combBandPassSpectrumDominantCycleMnemonic(
    params?: CombBandPassSpectrumParams,
): string {
    return buildMnemonic('cbps-dc', false, params);
}
