import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the AutoCorrelationPeriodogram indicator.
 *
 * All boolean fields are named so that `undefined` / `0` / `false` corresponds to the Ehlers
 * reference default behavior. This lets an empty params object produce the default indicator. */
export interface AutoCorrelationPeriodogramParams {
  /** Minimum (shortest) cycle period covered by the periodogram. Must be >= 2 (Nyquist).
   * Also drives the cutoff of the Super Smoother pre-filter. The default value is 10.
   * A zero value is treated as "use default". */
  minPeriod?: number;

  /** Maximum (longest) cycle period covered by the periodogram. Must be > minPeriod. Also
   * drives the cutoff of the Butterworth highpass pre-filter, the autocorrelation lag
   * range, and the DFT basis length. The default value is 48. A zero value is treated
   * as "use default". */
  maxPeriod?: number;

  /** Number of samples (M) used in each Pearson correlation accumulation. Must be >= 1.
   * The default value is 3 (matching Ehlers' EasyLanguage listing 8-3, which hardcodes 3).
   * A zero value is treated as "use default". */
  averagingLength?: number;

  /** Disables squaring the Fourier magnitude before smoothing when true. Ehlers' default
   * EasyLanguage listing 8-3 squares SqSum (R[P] = 0.2·SqSum² + 0.8·R_previous[P]); the default
   * value is false (squaring on). */
  disableSpectralSquaring?: boolean;

  /** Disables the per-bin exponential smoothing when true. Ehlers' default is enabled,
   * so the default value is false (smoothing on). */
  disableSmoothing?: boolean;

  /** Disables the fast-attack slow-decay automatic gain control when true. Ehlers'
   * default is enabled, so the default value is false (AGC on). */
  disableAutomaticGainControl?: boolean;

  /** Decay factor used by the fast-attack slow-decay automatic gain control. Must be in
   * the open interval (0, 1) when AGC is enabled. The default value is 0.995. A zero
   * value is treated as "use default". */
  automaticGainControlDecayFactor?: number;

  /** Selects fixed (min clamped to 0) normalization when true. The default is floating
   * normalization, consistent with the other zpano spectrum heatmaps. */
  fixedNormalization?: boolean;

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

export function defaultParams(): AutoCorrelationPeriodogramParams {
    return { minPeriod: 10, maxPeriod: 48, averagingLength: 3, automaticGainControlDecayFactor: 0.995 };
}
