import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the Jurik Wavelet Sampler indicator. */
export interface JurikWaveletSamplerParams {
    /** Number of wavelet columns (1-18). */
    index: number;

    /** A component of a bar to use when updating the indicator with a bar sample. */
    barComponent?: BarComponent;

    /** A component of a quote to use when updating the indicator with a quote sample. */
    quoteComponent?: QuoteComponent;

    /** A component of a trade to use when updating the indicator with a trade sample. */
    tradeComponent?: TradeComponent;
}

export function defaultParams(): JurikWaveletSamplerParams {
    return { index: 12 };
}
