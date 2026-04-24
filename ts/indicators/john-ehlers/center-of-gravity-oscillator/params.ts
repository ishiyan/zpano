import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the indicator. */
export interface CenterOfGravityOscillatorParams {
  /** Length is the length, l, (the number of time periods) of the Center of Gravity oscillator.
   *
   * The value should be a positive integer, greater or equal to 1.
   * The default value used by Ehlers is 10.
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

export function defaultParams(): CenterOfGravityOscillatorParams {
    return { length: 10 };
}
