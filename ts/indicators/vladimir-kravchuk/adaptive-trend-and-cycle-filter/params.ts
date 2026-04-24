import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the AdaptiveTrendAndCycleFilter indicator.
 *
 * The ATCF suite has no user-tunable numeric parameters: all five FIR filters
 * (FATL, SATL, RFTL, RSTL, RBCI) use fixed coefficient arrays published by
 * Vladimir Kravchuk.
 */
export interface AdaptiveTrendAndCycleFilterParams {
  /** A component of a bar to use when updating the indicator with a bar sample.
   *
   * If _undefined_, a default (BarComponent.Close) is used and the component
   * is not shown in the indicator mnemonic.
   */
  barComponent?: BarComponent;

  /** A component of a quote to use when updating the indicator with a quote sample.
   *
   * If _undefined_, a default (QuoteComponent.Mid) is used and the component
   * is not shown in the indicator mnemonic.
   */
  quoteComponent?: QuoteComponent;

  /** A component of a trade to use when updating the indicator with a trade sample.
   *
   * If _undefined_, a default (TradeComponent.Price) is used and the component
   * is not shown in the indicator mnemonic.
   */
  tradeComponent?: TradeComponent;
}

export function defaultParams(): AdaptiveTrendAndCycleFilterParams {
    return {};
}
