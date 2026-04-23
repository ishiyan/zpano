import { buildMetadata } from '../../core/build-metadata';
import { Bar } from '../../../entities/bar';
import { BarComponent, barComponentValue } from '../../../entities/bar-component';
import { Quote } from '../../../entities/quote';
import { DefaultQuoteComponent, quoteComponentValue } from '../../../entities/quote-component';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { DefaultTradeComponent, tradeComponentValue } from '../../../entities/trade-component';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { Indicator } from '../../core/indicator';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Heatmap } from '../../core/outputs/heatmap';
import { AutoCorrelationPeriodogramEstimator } from './estimator';
import { AutoCorrelationPeriodogramParams } from './params';

const DEFAULT_MIN_PERIOD = 10;
const DEFAULT_MAX_PERIOD = 48;
const DEFAULT_AVERAGING_LENGTH = 3;
const DEFAULT_AGC_DECAY_FACTOR = 0.995;
const AGC_DECAY_EPSILON = 1e-12;

/** __Autocorrelation Periodogram__ heatmap indicator (Ehlers).
 *
 * Displays a power heatmap of cyclic activity by taking a discrete Fourier transform of
 * the autocorrelation function. The input series is pre-conditioned by a 2-pole
 * Butterworth highpass (cutoff = maxPeriod) followed by a 2-pole Super Smoother
 * (cutoff = minPeriod). The autocorrelation function is evaluated at lags 0..maxPeriod
 * using Pearson correlation with a fixed averaging length. Each period bin's squared-sum
 * Fourier magnitude is exponentially smoothed, fast-attack / slow-decay AGC normalized,
 * and displayed.
 *
 * This implementation follows John Ehlers' EasyLanguage listing 8-3 from
 * "Cycle Analytics for Traders". It is NOT a port of MBST's AutoCorrelationSpectrum /
 * AutoCorrelationSpectrumEstimator, which omits the HP + SS pre-filter, uses a different
 * Pearson formulation, and smooths raw SqSum rather than SqSum².
 *
 * Reference: John F. Ehlers, "Cycle Analytics for Traders", Code Listing 8-3. */
export class AutoCorrelationPeriodogram implements Indicator {
  private readonly estimator: AutoCorrelationPeriodogramEstimator;
  private readonly primeCount: number;
  private readonly floatingNormalization: boolean;
  private readonly minParameterValue: number;
  private readonly maxParameterValue: number;
  private readonly parameterResolution: number;

  private readonly mnemonicValue: string;
  private readonly descriptionValue: string;

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  private windowCount = 0;
  private primed = false;

  /** Creates an instance with default parameters
   * (minPeriod=10, maxPeriod=48, averagingLength=3, squaring on, smoothing on,
   * AGC on (decay=0.995), floating normalization, BarComponent.Median). */
  public static default(): AutoCorrelationPeriodogram {
    return new AutoCorrelationPeriodogram({});
  }

  /** Creates an instance based on the given parameters. */
  public static fromParams(params: AutoCorrelationPeriodogramParams): AutoCorrelationPeriodogram {
    return new AutoCorrelationPeriodogram(params);
  }

  private constructor(params: AutoCorrelationPeriodogramParams) {
    const invalid = 'invalid autocorrelation periodogram parameters';

    const minPeriod = params.minPeriod !== undefined && params.minPeriod !== 0
      ? params.minPeriod : DEFAULT_MIN_PERIOD;
    const maxPeriod = params.maxPeriod !== undefined && params.maxPeriod !== 0
      ? params.maxPeriod : DEFAULT_MAX_PERIOD;
    const averagingLength = params.averagingLength !== undefined && params.averagingLength !== 0
      ? params.averagingLength : DEFAULT_AVERAGING_LENGTH;
    const agcDecayFactor = params.automaticGainControlDecayFactor !== undefined
      && params.automaticGainControlDecayFactor !== 0
      ? params.automaticGainControlDecayFactor : DEFAULT_AGC_DECAY_FACTOR;

    const squaringOn = !params.disableSpectralSquaring;
    const smoothingOn = !params.disableSmoothing;
    const agcOn = !params.disableAutomaticGainControl;
    const floatingNorm = !params.fixedNormalization;

    if (minPeriod < 2) {
      throw new Error(`${invalid}: MinPeriod should be >= 2`);
    }
    if (maxPeriod <= minPeriod) {
      throw new Error(`${invalid}: MaxPeriod should be > MinPeriod`);
    }
    if (averagingLength < 1) {
      throw new Error(`${invalid}: AveragingLength should be >= 1`);
    }
    if (agcOn && (agcDecayFactor <= 0 || agcDecayFactor >= 1)) {
      throw new Error(`${invalid}: AutomaticGainControlDecayFactor should be in (0, 1)`);
    }

    // AutoCorrelationPeriodogram mirrors Ehlers' reference: BarComponent.Median default.
    const bc = params.barComponent ?? BarComponent.Median;
    const qc = params.quoteComponent ?? DefaultQuoteComponent;
    const tc = params.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    this.estimator = new AutoCorrelationPeriodogramEstimator(
      minPeriod, maxPeriod, averagingLength,
      squaringOn, smoothingOn, agcOn, agcDecayFactor,
    );
    this.primeCount = this.estimator.filtBufferLen;
    this.floatingNormalization = floatingNorm;
    this.minParameterValue = minPeriod;
    this.maxParameterValue = maxPeriod;
    this.parameterResolution = 1;

    const cm = componentTripleMnemonic(bc, qc, tc);
    const flags = buildFlagTags(
      averagingLength, squaringOn, smoothingOn, agcOn, floatingNorm, agcDecayFactor,
    );
    this.mnemonicValue = `acp(${minPeriod}, ${maxPeriod}${flags}${cm})`;
    this.descriptionValue = 'Autocorrelation periodogram ' + this.mnemonicValue;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean { return this.primed; }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.AutoCorrelationPeriodogram,
      this.mnemonicValue,
      this.descriptionValue,
      [
        { mnemonic: this.mnemonicValue, description: this.descriptionValue },
      ],
    );
  }

  /** Feeds the next sample to the engine and returns the heatmap column.
   *
   * Before priming the heatmap is empty (with the indicator's parameter axis).
   * On a NaN input sample the state is left unchanged and an empty heatmap is returned. */
  public update(sample: number, time: Date): Heatmap {
    if (Number.isNaN(sample)) {
      return Heatmap.newEmptyHeatmap(time, this.minParameterValue, this.maxParameterValue, this.parameterResolution);
    }

    this.estimator.update(sample);

    if (!this.primed) {
      this.windowCount++;
      if (this.windowCount >= this.primeCount) {
        this.primed = true;
      } else {
        return Heatmap.newEmptyHeatmap(time, this.minParameterValue, this.maxParameterValue, this.parameterResolution);
      }
    }

    const lengthSpectrum = this.estimator.lengthSpectrum;

    const minRef = this.floatingNormalization ? this.estimator.spectrumMin : 0;
    // Estimator spectrum is already AGC-normalized in [0, 1]. Apply optional
    // floating-minimum subtraction for display.
    const maxRef = 1.0;
    const spectrumRange = maxRef - minRef;

    const values = new Array<number>(lengthSpectrum);
    let valueMin = Number.POSITIVE_INFINITY;
    let valueMax = Number.NEGATIVE_INFINITY;

    for (let i = 0; i < lengthSpectrum; i++) {
      const v = spectrumRange > 0
        ? (this.estimator.spectrum[i] - minRef) / spectrumRange
        : 0;
      values[i] = v;
      if (v < valueMin) valueMin = v;
      if (v > valueMax) valueMax = v;
    }

    return Heatmap.newHeatmap(
      time, this.minParameterValue, this.maxParameterValue, this.parameterResolution,
      valueMin, valueMax, values,
    );
  }

  /** Updates the indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    return this.updateEntity(sample.time, sample.value);
  }

  /** Updates the indicator given the next bar sample. */
  public updateBar(sample: Bar): IndicatorOutput {
    return this.updateEntity(sample.time, this.barComponentFunc(sample));
  }

  /** Updates the indicator given the next quote sample. */
  public updateQuote(sample: Quote): IndicatorOutput {
    return this.updateEntity(sample.time, this.quoteComponentFunc(sample));
  }

  /** Updates the indicator given the next trade sample. */
  public updateTrade(sample: Trade): IndicatorOutput {
    return this.updateEntity(sample.time, this.tradeComponentFunc(sample));
  }

  private updateEntity(time: Date, sample: number): IndicatorOutput {
    return [this.update(sample, time)];
  }
}

/** Encodes non-default settings as terse override-only tags. Returns an empty
 * string when all flags are at their defaults. Emission order matches the
 * Params field order. */
function buildFlagTags(
  averagingLength: number,
  squaringOn: boolean,
  smoothingOn: boolean,
  agcOn: boolean,
  floatingNorm: boolean,
  agcDecayFactor: number,
): string {
  let s = '';
  if (averagingLength !== DEFAULT_AVERAGING_LENGTH) {
    s += `, average=${averagingLength}`;
  }
  if (!squaringOn) s += ', no-sqr';
  if (!smoothingOn) s += ', no-smooth';
  if (!agcOn) s += ', no-agc';
  if (agcOn && Math.abs(agcDecayFactor - DEFAULT_AGC_DECAY_FACTOR) > AGC_DECAY_EPSILON) {
    s += `, agc=${agcDecayFactor}`;
  }
  if (!floatingNorm) s += ', no-fn';
  return s;
}
