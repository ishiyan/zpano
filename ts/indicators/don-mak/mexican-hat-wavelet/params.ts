import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Selects the frequency band of the Mexican Hat Wavelet filter. */
export enum Band {
  /** High-frequency band (a_f = 1.483, period ~ 4.6 bars). */
  High = 0,

  /** Mid-frequency band (a_f = 4.048, period ~ 13.5 bars). */
  Mid = 1,

  /** Low-frequency band (a_f = 15.97, period ~ 54 bars). */
  Low = 2,

  /** User-specified dilation or period. */
  Custom = 3,
}

/** Describes parameters to create an instance of the Mexican Hat Wavelet indicator. */
export interface MexicanHatWaveletParams {
    /**
     * The frequency band selection (High, Mid, Low, Custom).
     *
     * The default value is Band.Mid.
     */
    band?: Band;

    /**
     * The custom dilation parameter a_f, used only when band is Custom.
     *
     * The value should be > 0.
     */
    dilation?: number;

    /**
     * The custom center period in bars, used only when band is Custom.
     *
     * The value should be > 2. Mutually exclusive with dilation.
     */
    period?: number;

    /**
     * A component of a bar to use when updating the indicator with a bar sample.
     *
     * If _undefined_, the bar component will have a default value (ClosePrice)
     * and will not be shown in the indicator mnemonic.
     */
    barComponent?: BarComponent;

    /**
     * A component of a quote to use when updating the indicator with a quote sample.
     *
     * If _undefined_, the quote component will have a default value and will not be shown in the indicator mnemonic.
     */
    quoteComponent?: QuoteComponent;

    /**
     * A component of a trade to use when updating the indicator with a trade sample.
     *
     * If _undefined_, the trade component will have a default value and will not be shown in the indicator mnemonic.
     */
    tradeComponent?: TradeComponent;
}

export function defaultParams(): MexicanHatWaveletParams {
    return { band: Band.Mid };
}
