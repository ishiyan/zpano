import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the Linear Regression indicator. */
export interface LinearRegressionParams {
  /**
   * Length is the number of time periods in the regression window.
   *
   * The value should be a positive integer, greater than 1.
   */
  length: number;

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

export function defaultParams(): LinearRegressionParams {
    return {
        length: 20,
    };
}
