import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the CombBandPassSpectrum indicator.
 *
 * All boolean fields are named so that `undefined` / `false` corresponds to the Ehlers
 * reference default behavior. This lets an empty params object produce the default
 * indicator. */
export interface CombBandPassSpectrumParams {
  /** Minimum (shortest) cycle period covered by the spectrum. Must be >= 2 (Nyquist).
   * Also drives the cutoff of the Super Smoother pre-filter. The default value is 10.
   * A zero value is treated as "use default". */
  minPeriod?: number;

  /** Maximum (longest) cycle period covered by the spectrum. Must be > minPeriod. Also
   * drives the cutoff of the Butterworth highpass pre-filter and the band-pass output
   * history length per bin. The default value is 48. A zero value is treated as "use
   * default". */
  maxPeriod?: number;

  /** Fractional bandwidth of each band-pass filter in the comb. Must be in (0, 1).
   * Typical Ehlers values are around 0.3 (default) for medium selectivity. A zero
   * value is treated as "use default". */
  bandwidth?: number;

  /** Disables the spectral dilation compensation (division of each band-pass output by
   * its evaluated period before squaring) when true. Ehlers' default is enabled, so the
   * default value is false (SDC on). */
  disableSpectralDilationCompensation?: boolean;

  /** Disables the fast-attack slow-decay automatic gain control when true. Ehlers'
   * default is enabled, so the default value is false (AGC on). */
  disableAutomaticGainControl?: boolean;

  /** Decay factor used by the fast-attack slow-decay automatic gain control. Must be in
   * the open interval (0, 1) when AGC is enabled. The default value is 0.995 (matching
   * Ehlers' EasyLanguage listing 10-1). A zero value is treated as "use default". */
  automaticGainControlDecayFactor?: number;

  /** Selects fixed (min clamped to 0) normalization when true. The default is floating
   * normalization, consistent with the other zpano spectrum heatmaps. Note that Ehlers'
   * listing 10-1 uses fixed normalization (MaxPwr only); set this to true for exact
   * EL-faithful behavior. */
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
