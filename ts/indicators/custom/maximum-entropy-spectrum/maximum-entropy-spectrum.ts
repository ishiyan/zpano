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
import { MaximumEntropySpectrumEstimator } from './estimator';
import { MaximumEntropySpectrumParams } from './params';

const DEFAULT_LENGTH = 60;
const DEFAULT_DEGREE = 30;
const DEFAULT_MIN_PERIOD = 2;
const DEFAULT_MAX_PERIOD = 59;
const DEFAULT_SPECTRUM_RESOLUTION = 1;
const DEFAULT_AGC_DECAY_FACTOR = 0.995;
const AGC_DECAY_EPSILON = 1e-12;

/** __Maximum Entropy Spectrum__ heatmap indicator (MBST port).
 *
 * Displays a power heatmap of cyclic activity over a configurable cycle-period range using
 * Burg's maximum-entropy auto-regressive method. It supports a fast-attack slow-decay
 * automatic gain control and either floating or fixed (0-clamped) intensity normalization.
 *
 * Reference: MBST Mbs.Trading.Indicators.SpectralAnalysis.MaximumEntropySpectrum. */
export class MaximumEntropySpectrum implements Indicator {
  private readonly estimator: MaximumEntropySpectrumEstimator;
  private readonly lastIndex: number;
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
   * (length=60, degree=30, minPeriod=2, maxPeriod=59, spectrumResolution=1,
   * AGC on (decay=0.995), floating normalization, BarComponent.Median). */
  public static default(): MaximumEntropySpectrum {
    return new MaximumEntropySpectrum({});
  }

  /** Creates an instance based on the given parameters. */
  public static fromParams(params: MaximumEntropySpectrumParams): MaximumEntropySpectrum {
    return new MaximumEntropySpectrum(params);
  }

  private constructor(params: MaximumEntropySpectrumParams) {
    const invalid = 'invalid maximum entropy spectrum parameters';

    const length = params.length !== undefined && params.length !== 0
      ? params.length : DEFAULT_LENGTH;
    const degree = params.degree !== undefined && params.degree !== 0
      ? params.degree : DEFAULT_DEGREE;
    const minPeriod = params.minPeriod !== undefined && params.minPeriod !== 0
      ? params.minPeriod : DEFAULT_MIN_PERIOD;
    const maxPeriod = params.maxPeriod !== undefined && params.maxPeriod !== 0
      ? params.maxPeriod : DEFAULT_MAX_PERIOD;
    const spectrumResolution = params.spectrumResolution !== undefined && params.spectrumResolution !== 0
      ? params.spectrumResolution : DEFAULT_SPECTRUM_RESOLUTION;
    const agcDecayFactor = params.automaticGainControlDecayFactor !== undefined
      && params.automaticGainControlDecayFactor !== 0
      ? params.automaticGainControlDecayFactor : DEFAULT_AGC_DECAY_FACTOR;

    const agcOn = !params.disableAutomaticGainControl;
    const floatingNorm = !params.fixedNormalization;

    if (length < 2) {
      throw new Error(`${invalid}: Length should be >= 2`);
    }
    if (degree <= 0 || degree >= length) {
      throw new Error(`${invalid}: Degree should be > 0 and < Length`);
    }
    if (minPeriod < 2) {
      throw new Error(`${invalid}: MinPeriod should be >= 2`);
    }
    if (maxPeriod <= minPeriod) {
      throw new Error(`${invalid}: MaxPeriod should be > MinPeriod`);
    }
    if (maxPeriod > 2 * length) {
      throw new Error(`${invalid}: MaxPeriod should be <= 2 * Length`);
    }
    if (spectrumResolution < 1) {
      throw new Error(`${invalid}: SpectrumResolution should be >= 1`);
    }
    if (agcOn && (agcDecayFactor <= 0 || agcDecayFactor >= 1)) {
      throw new Error(`${invalid}: AutomaticGainControlDecayFactor should be in (0, 1)`);
    }

    // MaximumEntropySpectrum mirrors MBST's reference: BarComponent.Median default.
    const bc = params.barComponent ?? BarComponent.Median;
    const qc = params.quoteComponent ?? DefaultQuoteComponent;
    const tc = params.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    this.estimator = new MaximumEntropySpectrumEstimator(
      length, degree, minPeriod, maxPeriod, spectrumResolution,
      agcOn, agcDecayFactor,
    );
    this.lastIndex = length - 1;
    this.floatingNormalization = floatingNorm;
    this.minParameterValue = minPeriod;
    this.maxParameterValue = maxPeriod;
    this.parameterResolution = spectrumResolution;

    const cm = componentTripleMnemonic(bc, qc, tc);
    const flags = buildFlagTags(agcOn, floatingNorm, agcDecayFactor);
    this.mnemonicValue = `mespect(${length}, ${degree}, ${formatNum(minPeriod)}, ${formatNum(maxPeriod)}, ${spectrumResolution}${flags}${cm})`;
    this.descriptionValue = 'Maximum entropy spectrum ' + this.mnemonicValue;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean { return this.primed; }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.MaximumEntropySpectrum,
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

    const window = this.estimator.inputSeries;

    if (this.primed) {
      for (let i = 0; i < this.lastIndex; i++) {
        window[i] = window[i + 1];
      }
      window[this.lastIndex] = sample;
    } else {
      window[this.windowCount] = sample;
      this.windowCount++;
      if (this.windowCount === this.estimator.length) {
        this.primed = true;
      }
    }

    if (!this.primed) {
      return Heatmap.newEmptyHeatmap(time, this.minParameterValue, this.maxParameterValue, this.parameterResolution);
    }

    this.estimator.calculate();

    const lengthSpectrum = this.estimator.lengthSpectrum;

    const minRef = this.floatingNormalization ? this.estimator.spectrumMin : 0;
    const maxRef = this.estimator.spectrumMax;
    const spectrumRange = maxRef - minRef;

    // MBST fills spectrum[0] at MaxPeriod and spectrum[last] at MinPeriod.
    // The heatmap axis runs MinPeriod -> MaxPeriod, so reverse on output.
    const values = new Array<number>(lengthSpectrum);
    let valueMin = Number.POSITIVE_INFINITY;
    let valueMax = Number.NEGATIVE_INFINITY;

    for (let i = 0; i < lengthSpectrum; i++) {
      const v = (this.estimator.spectrum[lengthSpectrum - 1 - i] - minRef) / spectrumRange;
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

/** Encodes non-default boolean/decay settings as terse override-only tags. Empty when all
 * flags are at their defaults. */
function buildFlagTags(agcOn: boolean, floatingNorm: boolean, agcDecayFactor: number): string {
  let s = '';
  if (!agcOn) s += ', no-agc';
  if (agcOn && Math.abs(agcDecayFactor - DEFAULT_AGC_DECAY_FACTOR) > AGC_DECAY_EPSILON) {
    s += `, agc=${formatNum(agcDecayFactor)}`;
  }
  if (!floatingNorm) s += ', no-fn';
  return s;
}

/** Matches Go fmt.Sprintf("%g") for the common integer and decimal cases used in the mnemonic. */
function formatNum(n: number): string {
  return n.toString();
}
