import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the CyberCycle indicator based on smoothing factor. */
export interface CyberCycleSmoothingFactorParams {
  /**
   * SmoothingFactor is the smoothing factor, α in [0,1], of the Cyber Cycle.
   *
   * The equivalent length ℓ is:
   *
   *     ℓ = round(2/α) - 1, 0<α≤1, 1≤ℓ.
   *
   * If α is near zero (< epsilon), ℓ is set to Number.MAX_SAFE_INTEGER.
   * The default value used by Ehlers is 0.07 (ℓ = 28).
   */
  smoothingFactor: number;

  /**
   * SignalLag is the signal lag (the number of time periods) for the signal line EMA.
   *
   * The signal EMA factor is 1/(signalLag+1).
   * The value should be a positive integer, greater or equal to 1.
   * The default value used by Ehlers is 9.
   */
  signalLag: number;

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
