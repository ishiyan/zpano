import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the AutoCorrelationIndicator.
 *
 * All fields are named so that `undefined` / `0` / `false` corresponds to the Ehlers
 * reference default behavior. This lets an empty params object produce the default
 * indicator. */
export interface AutoCorrelationIndicatorParams {
  /** Minimum (shortest) correlation lag shown on the heatmap axis. Must be >= 1.
   * The default value is 3 (matching Ehlers' EasyLanguage listing 8-2, which plots
   * lags 3..48). A zero value is treated as "use default". */
  minLag?: number;

  /** Maximum (longest) correlation lag shown on the heatmap axis. Must be > minLag.
   * Also drives the cutoff of the 2-pole Butterworth highpass pre-filter. The default
   * value is 48. A zero value is treated as "use default". */
  maxLag?: number;

  /** Cutoff period of the 2-pole Super Smoother pre-filter applied after the highpass.
   * Must be >= 2. The default value is 10 (matching Ehlers' EasyLanguage listing 8-2,
   * which hardcodes 10). A zero value is treated as "use default". */
  smoothingPeriod?: number;

  /** Number of samples (M) used in each Pearson correlation accumulation. When zero
   * (the Ehlers default), M equals the current lag, making each correlation use the
   * same number of samples as its lag distance. When positive, the same M is used
   * for all lags. Must be >= 0. */
  averagingLength?: number;

  /** A component of a bar to use when updating the indicator with a bar sample.
   *
   * If _undefined_, the default (BarComponent.Median, hl/2) is used, matching Ehlers'
   * reference. Since this differs from the framework default, it is always shown in
   * the mnemonic. */
  barComponent?: BarComponent;

  /** A component of a quote to use when updating the indicator with a quote sample.
   *
   * If _undefined_, a default value is used and the component is not shown in the mnemonic. */
  quoteComponent?: QuoteComponent;

  /** A component of a trade to use when updating the indicator with a trade sample.
   *
   * If _undefined_, a default value is used and the component is not shown in the mnemonic. */
  tradeComponent?: TradeComponent;
}
