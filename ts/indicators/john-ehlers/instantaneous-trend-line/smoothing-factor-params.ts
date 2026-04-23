import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the InstantaneousTrendLine indicator based on smoothing factor. */
export interface InstantaneousTrendLineSmoothingFactorParams {
  /**
   * SmoothingFactor is the smoothing factor, \u03b1 in [0,1], of the Instantaneous Trend Line.
   *
   * The equivalent length \u2113 is:
   *
   *     \u2113 = round(2/\u03b1) - 1, 0<\u03b1\u22641, 1\u2264\u2113.
   *
   * If \u03b1 is near zero (< epsilon), \u2113 is set to Number.MAX_SAFE_INTEGER.
   * The default value used by Ehlers is 0.07 (\u2113 = 28).
   */
  smoothingFactor: number;

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
