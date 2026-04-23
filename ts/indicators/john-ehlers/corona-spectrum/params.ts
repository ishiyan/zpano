import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the CoronaSpectrum indicator. */
export interface CoronaSpectrumParams {
  /** Minimal raster value (z) of the heatmap, in decibels. Corresponds to the
   * CoronaLowerDecibels threshold.
   *
   * The default value is 6. A zero value is treated as "use default".
   */
  minRasterValue?: number;

  /** Maximal raster value (z) of the heatmap, in decibels. Corresponds to the
   * CoronaUpperDecibels threshold.
   *
   * The default value is 20. A zero value is treated as "use default".
   */
  maxRasterValue?: number;

  /** Minimal ordinate (y) value of the heatmap, representing the minimal cycle
   * period covered by the filter bank. Rounded up to the nearest integer.
   *
   * The default value is 6. A zero value is treated as "use default".
   */
  minParameterValue?: number;

  /** Maximal ordinate (y) value of the heatmap, representing the maximal cycle
   * period covered by the filter bank. Rounded down to the nearest integer.
   *
   * The default value is 30. A zero value is treated as "use default".
   */
  maxParameterValue?: number;

  /** High-pass filter cutoff (de-trending period) used by the inner Corona engine.
   * Suggested values are 20, 30, 100.
   *
   * The default value is 30. A zero value is treated as "use default".
   */
  highPassFilterCutoff?: number;

  /** A component of a bar to use when updating the indicator with a bar sample.
   *
   * If _undefined_, the default (BarComponent.Median, hl/2) is used, matching
   * Ehlers' reference which operates on (High+Low)/2. Since this differs from
   * the framework default, it is always shown in the mnemonic.
   */
  barComponent?: BarComponent;

  /** A component of a quote to use when updating the indicator with a quote sample.
   *
   * If _undefined_, a default value is used and the component is not shown in the indicator mnemonic.
   */
  quoteComponent?: QuoteComponent;

  /** A component of a trade to use when updating the indicator with a trade sample.
   *
   * If _undefined_, a default value is used and the component is not shown in the indicator mnemonic.
   */
  tradeComponent?: TradeComponent;
}
