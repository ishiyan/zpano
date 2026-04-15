import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the indicator. */
export interface FractalAdaptiveMovingAverageParams {
  /** Length is the length, l, (the number of time periods) of the Fractal Adaptive Moving Average.
   *
   * The value should be an even integer, greater or equal to 2.
   * The default value is 16.
   */
  length: number;

  /** SlowestSmoothingFactor is the slowest boundary smoothing factor, as in [0,1].
   * The equivalent length ls is
   *
   *   ls = 2/as - 1, 0 < as <= 1, 1 <= ls
   *
   * The default value is 0.01 (equivalent ls = 199).
    */
  slowestSmoothingFactor: number;

  /**
   * A component of a bar to use when updating the indicator with a bar sample.
   *
   * If _undefined_, the bar component will have a default value and will not be shown in the indicator mnemonic.
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
