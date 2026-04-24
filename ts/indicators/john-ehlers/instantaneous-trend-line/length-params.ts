import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the InstantaneousTrendLine indicator based on length. */
export interface InstantaneousTrendLineLengthParams {
  /**
   * Length is the length (the number of time periods, \u2113) of the Instantaneous Trend Line.
   *
   * The smoothing factor \u03b1 is derived as \u03b1 = 2/(\u2113+1).
   * The value should be a positive integer, greater or equal to 1.
   */
  length: number;

  /**
   * A component of a bar to use when updating the indicator with a bar sample.
   *
   * If _undefined_, the bar component defaults to BarComponent.Median (median price)
   * and will be shown in the indicator mnemonic as 'hl/2' since it differs from the
   * framework default (close).
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

export function defaultLengthParams(): InstantaneousTrendLineLengthParams {
    return { length: 28 };
}
