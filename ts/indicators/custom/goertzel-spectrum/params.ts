import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the GoertzelSpectrum indicator.
 *
 * All boolean fields are named so that `undefined` / `false` corresponds to the MBST default
 * behavior. This lets an empty params object produce the default indicator. */
export interface GoertzelSpectrumParams {
  /** Number of time periods in the spectrum window. Determines the minimum and maximum spectrum
   * periods. Must be >= 2. The default value is 64. A zero value is treated as "use default". */
  length?: number;

  /** Minimum cycle period covered by the spectrum. Must be >= 2 (2 corresponds to the Nyquist
   * frequency). The default value is 2. A zero value is treated as "use default". */
  minPeriod?: number;

  /** Maximum cycle period covered by the spectrum. Must be > minPeriod and <= 2 * length.
   * The default value is 64. A zero value is treated as "use default". */
  maxPeriod?: number;

  /** Spectrum resolution (positive integer). A value of 10 means the spectrum is evaluated at
   * every 0.1 of period amplitude. Must be >= 1. The default value is 1. A zero value is
   * treated as "use default". */
  spectrumResolution?: number;

  /** Selects the first-order Goertzel algorithm when true; otherwise the second-order algorithm
   * is used. MBST default behavior uses the second-order algorithm, so the default value is
   * false. */
  isFirstOrder?: boolean;

  /** Disables spectral dilation compensation when true. MBST default behavior is enabled, so the
   * default value is false (compensation on). */
  disableSpectralDilationCompensation?: boolean;

  /** Disables the fast-attack slow-decay automatic gain control when true. MBST default
   * behavior is enabled, so the default value is false (AGC on). */
  disableAutomaticGainControl?: boolean;

  /** Decay factor used by the fast-attack slow-decay automatic gain control. Must be in the
   * open interval (0, 1) when AGC is enabled. The default value is 0.991. A zero value is
   * treated as "use default". */
  automaticGainControlDecayFactor?: number;

  /** Selects fixed (min clamped to 0) normalization when true. MBST default is floating
   * normalization, so the default value is false (floating normalization). */
  fixedNormalization?: boolean;

  /** A component of a bar to use when updating the indicator with a bar sample.
   *
   * If _undefined_, the default (BarComponent.Median, hl/2) is used, matching MBST's reference
   * which operates on (High+Low)/2. Since this differs from the framework default, it is
   * always shown in the mnemonic. */
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

export function defaultParams(): GoertzelSpectrumParams {
    return { length: 64, minPeriod: 2, maxPeriod: 64, spectrumResolution: 1, automaticGainControlDecayFactor: 0.991 };
}
