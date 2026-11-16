import { BarComponent, DefaultBarComponent } from '../../../entities/bar-component';
import { DefaultQuoteComponent, QuoteComponent } from '../../../entities/quote-component';
import { DefaultTradeComponent, TradeComponent } from '../../../entities/trade-component';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';

/** Selects the Ehlers source profile for AutoCorrelationPeriodogram. */
export enum AutoCorrelationPeriodogramSource {
    /** TASC September 2016 "Measuring Market Cycles" sidebar profile (default). */
    Article = 'article',
    /** Corrected Cycle Analytics for Traders (2013), Listing 8-3 profile. */
    Book = 'book',
}

/** Parameters for Ehlers' AutoCorrelationPeriodogram. */
export interface AutoCorrelationPeriodogramParams {
    /** Source profile: 'article' (default) or 'book'. */
    source?: AutoCorrelationPeriodogramSource | 'article' | 'book';

    /** Shortest spectrum bin and SuperSmoother cutoff in bars (>= 3). Undefined selects 8 (article) or 10 (book). */
    minPeriod?: number;

    /** Longest spectrum bin, high-pass cutoff, and largest lag (>= 5, > minPeriod). Default: 48. */
    maxPeriod?: number;

    /** Pearson window length: 0 for lag-dependent (M = lag) or >= 2. Default: 3. */
    averagingLength?: number;

    /** Cube normalized power before centroid (article) or square display power after centroid (book). Default: false. */
    enhanceResolution?: boolean;

    /** Bar component used by updateBar(); undefined defaults to Close. */
    barComponent?: BarComponent;

    /** Quote component used by updateQuote(); undefined defaults to Mid. */
    quoteComponent?: QuoteComponent;

    /** Trade component used by updateTrade(); undefined defaults to Price. */
    tradeComponent?: TradeComponent;
}

/** Returns Ehlers' 2016 article default parameters (minPeriod omitted so source override resolves 8 vs 10). */
export function defaultParams(): AutoCorrelationPeriodogramParams {
    return {
        source: AutoCorrelationPeriodogramSource.Article,
        maxPeriod: 48,
        averagingLength: 3,
        enhanceResolution: false,
    };
}

/** Resolves undefined minPeriod to 8 for article mode or 10 for book mode. */
export function resolveMinPeriod(
    source: AutoCorrelationPeriodogramSource | 'article' | 'book' | undefined,
    minPeriod: number | undefined,
): number {
    if (minPeriod !== undefined) {
        return minPeriod;
    }
    return source === AutoCorrelationPeriodogramSource.Book ? 10 : 8;
}

function buildMnemonic(
    prefix: string,
    includeBookEnhance: boolean,
    params?: AutoCorrelationPeriodogramParams,
): string {
    const source = params?.source ?? AutoCorrelationPeriodogramSource.Article;
    const isBook = source === AutoCorrelationPeriodogramSource.Book;
    const minPeriod = resolveMinPeriod(source, params?.minPeriod);
    const maxPeriod = params?.maxPeriod ?? 48;
    const averagingLength = params?.averagingLength ?? 3;
    const enhance = params?.enhanceResolution ?? false;
    const barComp = params?.barComponent ?? DefaultBarComponent;
    const quoteComp = params?.quoteComponent ?? DefaultQuoteComponent;
    const tradeComp = params?.tradeComponent ?? DefaultTradeComponent;
    const compSuffix = componentTripleMnemonic(barComp, quoteComp, tradeComp);
    const sourceTag = isBook ? ', book' : '';
    const enhTag = enhance && (!isBook || includeBookEnhance) ? ', enh' : '';
    return `${prefix}(${minPeriod}, ${maxPeriod}, ${averagingLength}${sourceTag}${enhTag}${compSuffix})`;
}

/** Builds the primary heatmap mnemonic for AutoCorrelationPeriodogram. */
export function autoCorrelationPeriodogramMnemonic(
    params?: AutoCorrelationPeriodogramParams,
): string {
    return buildMnemonic('acp', true, params);
}

/** Builds the dominant-cycle mnemonic for AutoCorrelationPeriodogram. */
export function autoCorrelationPeriodogramDominantCycleMnemonic(
    params?: AutoCorrelationPeriodogramParams,
): string {
    return buildMnemonic('acp-dc', false, params);
}
