import { BarComponent, DefaultBarComponent } from '../../../entities/bar-component';
import { DefaultQuoteComponent, QuoteComponent } from '../../../entities/quote-component';
import { DefaultTradeComponent, TradeComponent } from '../../../entities/trade-component';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';

const ALPHA_EPSILON = 1e-8;

/** Describes length-based parameters to create an instance of Ehlers' ZeroLagErrorCorrectingExponentialMovingAverage. */
export interface ZeroLagErrorCorrectingExponentialMovingAverageLengthParams {
    /** Equivalent EMA length L (integer >= 1), defining alpha = 2 / (L + 1). Default: 20. */
    length?: number;

    /** Smoothing factor must not be combined with length. */
    smoothingFactor?: never;

    /** Nonnegative integer gain bound in tenths (candidates span [-G/10, +G/10]). Default: 50. */
    gainLimit?: number;

    /** Bar component used when updating with a bar sample; undefined defaults to Close and is omitted from the mnemonic. */
    barComponent?: BarComponent;

    /** Quote component used when updating with a quote sample; undefined defaults to Mid and is omitted from the mnemonic. */
    quoteComponent?: QuoteComponent;

    /** Trade component used when updating with a trade sample; undefined defaults to Price and is omitted from the mnemonic. */
    tradeComponent?: TradeComponent;
}

/** Describes smoothing-factor-based parameters to create an instance of Ehlers' ZeroLagErrorCorrectingExponentialMovingAverage. */
export interface ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams {
    /** Baseline EMA smoothing factor alpha in (0, 1]. Default: 2/21. */
    smoothingFactor: number;

    /** Length must not be combined with smoothingFactor. */
    length?: never;

    /** Nonnegative integer gain bound in tenths (candidates span [-G/10, +G/10]). Default: 50. */
    gainLimit?: number;

    /** Bar component used when updating with a bar sample; undefined defaults to Close and is omitted from the mnemonic. */
    barComponent?: BarComponent;

    /** Quote component used when updating with a quote sample; undefined defaults to Mid and is omitted from the mnemonic. */
    quoteComponent?: QuoteComponent;

    /** Trade component used when updating with a trade sample; undefined defaults to Price and is omitted from the mnemonic. */
    tradeComponent?: TradeComponent;
}

export type ZeroLagErrorCorrectingExponentialMovingAverageParams =
    | ZeroLagErrorCorrectingExponentialMovingAverageLengthParams
    | ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams;

export type Params = ZeroLagErrorCorrectingExponentialMovingAverageLengthParams;

export function defaultLengthParams(): ZeroLagErrorCorrectingExponentialMovingAverageLengthParams {
    return {
        length: 20,
        gainLimit: 50,
    };
}

export function defaultSmoothingFactorParams(): ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams {
    return {
        smoothingFactor: 2 / 21,
        gainLimit: 50,
    };
}

export function defaultParams(): ZeroLagErrorCorrectingExponentialMovingAverageLengthParams {
    return defaultLengthParams();
}

export function equivalentLengthFromAlpha(alpha: number): number {
    if (!(alpha >= ALPHA_EPSILON)) {
        return Number.MAX_SAFE_INTEGER;
    }
    const rounded = Math.floor(2 / alpha + 0.5);
    if (rounded >= Number.MAX_SAFE_INTEGER) {
        return Number.MAX_SAFE_INTEGER;
    }
    return rounded - 1;
}

function isSmoothingFactorParams(
    params: ZeroLagErrorCorrectingExponentialMovingAverageParams,
): params is ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams {
    return 'smoothingFactor' in params && params.smoothingFactor !== undefined;
}

function formatMnemonic(
    prefix: string,
    params?: ZeroLagErrorCorrectingExponentialMovingAverageParams,
): string {
    const source = params ?? defaultLengthParams();
    const gainLimit = source.gainLimit === undefined ? 50 : source.gainLimit;
    const barComp = source.barComponent ?? DefaultBarComponent;
    const quoteComp = source.quoteComponent ?? DefaultQuoteComponent;
    const tradeComp = source.tradeComponent ?? DefaultTradeComponent;
    const suffix = componentTripleMnemonic(barComp, quoteComp, tradeComp);

    if (isSmoothingFactorParams(source)) {
        const equivLength = equivalentLengthFromAlpha(source.smoothingFactor);
        return `${prefix}(${equivLength}, ${source.smoothingFactor.toFixed(8)}, ${gainLimit}${suffix})`;
    }

    const length = source.length === undefined ? 20 : source.length;
    return `${prefix}(${length}, ${gainLimit}${suffix})`;
}

/** Builds the primary EC mnemonic string for ZECEMA. */
export function zeroLagErrorCorrectingExponentialMovingAverageMnemonic(
    params?: ZeroLagErrorCorrectingExponentialMovingAverageParams,
): string {
    return formatMnemonic('zecema', params);
}

/** Builds the baseline EMA mnemonic string for ZECEMA. */
export function zeroLagErrorCorrectingExponentialMovingAverageEmaMnemonic(
    params?: ZeroLagErrorCorrectingExponentialMovingAverageParams,
): string {
    return formatMnemonic('zecemaEma', params);
}
