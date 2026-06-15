import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Selects the frequency band of the Sinc Wavelet Band-Pass filter. */
export enum Band {
  /** High-frequency band (periods 8-16 bars). */
  High = 0,

  /** Mid-frequency band (periods 16-32 bars). */
  Mid = 1,

  /** Low-frequency band (periods 32-64 bars). */
  Low = 2,

  /** Full band (periods 8-64 bars). */
  Full = 3,
}

/** Describes parameters to create an instance of the Sinc Wavelet Band-Pass indicator. */
export interface SincWaveletBandpassParams {
    /**
     * The frequency band selection (High, Mid, Low, Full).
     *
     * The default value is Band.Mid.
     */
    band?: Band;

    /**
     * Whether a cubic velocity kernel is applied to the band-pass output.
     *
     * The default value is false.
     */
    velocity?: boolean;

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

export function defaultParams(): SincWaveletBandpassParams {
    return { band: Band.Mid, velocity: false };
}
