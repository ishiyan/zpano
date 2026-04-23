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
import { CombBandPassSpectrumEstimator } from './estimator';
import { CombBandPassSpectrumParams } from './params';

const DEFAULT_MIN_PERIOD = 10;
const DEFAULT_MAX_PERIOD = 48;
const DEFAULT_BANDWIDTH = 0.3;
const DEFAULT_AGC_DECAY_FACTOR = 0.995;
const AGC_DECAY_EPSILON = 1e-12;
const BANDWIDTH_EPSILON = 1e-12;

/** __Comb Band-Pass Spectrum__ heatmap indicator (Ehlers).
 *
 * Displays a power heatmap of cyclic activity over a configurable cycle-period range by
 * running a bank of 2-pole band-pass filters, one per integer period in
 * [minPeriod..maxPeriod]. The input series is pre-conditioned by a 2-pole Butterworth
 * highpass (cutoff = maxPeriod) followed by a 2-pole Super Smoother (cutoff = minPeriod)
 * before entering the comb. Each bin's power is the sum of squared band-pass outputs over
 * the last N samples, optionally compensated for spectral dilation (divide by N) and
 * normalized by a fast-attack slow-decay automatic gain control.
 *
 * This implementation follows John Ehlers' EasyLanguage listing 10-1 from
 * "Cycle Analytics for Traders". It is NOT a port of MBST's CombBandPassSpectrumEstimator,
 * which is misnamed and actually implements a plain DFT (see the
 * DiscreteFourierTransformSpectrum indicator for a faithful MBST DFT port).
 *
 * Reference: John F. Ehlers, "Cycle Analytics for Traders", Code Listing 10-1. */
export class CombBandPassSpectrum implements Indicator {
  private readonly estimator: CombBandPassSpectrumEstimator;
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
   * (minPeriod=10, maxPeriod=48, bandwidth=0.3, SDC on, AGC on (decay=0.995),
   * floating normalization, BarComponent.Median). */
  public static default(): CombBandPassSpectrum {
    return new CombBandPassSpectrum({});
  }

  /** Creates an instance based on the given parameters. */
  public static fromParams(params: CombBandPassSpectrumParams): CombBandPassSpectrum {
    return new CombBandPassSpectrum(params);
  }

  private constructor(params: CombBandPassSpectrumParams) {
    const invalid = 'invalid comb band-pass spectrum parameters';

    const minPeriod = params.minPeriod !== undefined && params.minPeriod !== 0
      ? params.minPeriod : DEFAULT_MIN_PERIOD;
    const maxPeriod = params.maxPeriod !== undefined && params.maxPeriod !== 0
      ? params.maxPeriod : DEFAULT_MAX_PERIOD;
    const bandwidth = params.bandwidth !== undefined && params.bandwidth !== 0
      ? params.bandwidth : DEFAULT_BANDWIDTH;
    const agcDecayFactor = params.automaticGainControlDecayFactor !== undefined
      && params.automaticGainControlDecayFactor !== 0
      ? params.automaticGainControlDecayFactor : DEFAULT_AGC_DECAY_FACTOR;

    const sdcOn = !params.disableSpectralDilationCompensation;
    const agcOn = !params.disableAutomaticGainControl;
    const floatingNorm = !params.fixedNormalization;

    if (minPeriod < 2) {
      throw new Error(`${invalid}: MinPeriod should be >= 2`);
    }
    if (maxPeriod <= minPeriod) {
      throw new Error(`${invalid}: MaxPeriod should be > MinPeriod`);
    }
    if (bandwidth <= 0 || bandwidth >= 1) {
      throw new Error(`${invalid}: Bandwidth should be in (0, 1)`);
    }
    if (agcOn && (agcDecayFactor <= 0 || agcDecayFactor >= 1)) {
      throw new Error(`${invalid}: AutomaticGainControlDecayFactor should be in (0, 1)`);
    }

    // CombBandPassSpectrum mirrors Ehlers' reference: BarComponent.Median default.
    const bc = params.barComponent ?? BarComponent.Median;
    const qc = params.quoteComponent ?? DefaultQuoteComponent;
    const tc = params.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    this.estimator = new CombBandPassSpectrumEstimator(
      minPeriod, maxPeriod, bandwidth, sdcOn, agcOn, agcDecayFactor,
    );
    this.primeCount = maxPeriod;
    this.floatingNormalization = floatingNorm;
    this.minParameterValue = minPeriod;
    this.maxParameterValue = maxPeriod;
    this.parameterResolution = 1;

    const cm = componentTripleMnemonic(bc, qc, tc);
    const flags = buildFlagTags(bandwidth, sdcOn, agcOn, floatingNorm, agcDecayFactor);
    this.mnemonicValue = `cbps(${formatNum(minPeriod)}, ${formatNum(maxPeriod)}${flags}${cm})`;
    this.descriptionValue = 'Comb band-pass spectrum ' + this.mnemonicValue;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean { return this.primed; }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.CombBandPassSpectrum,
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
    const maxRef = this.estimator.spectrumMax;
    const spectrumRange = maxRef - minRef;

    // The estimator's spectrum is already in axis order (bin 0 = MinPeriod,
    // bin last = MaxPeriod), matching the heatmap axis.
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

/** Encodes non-default boolean/decay/bandwidth settings as terse override-only tags.
 * Empty when all flags are at their defaults. Emission order matches the Params field
 * order. */
function buildFlagTags(
  bandwidth: number,
  sdcOn: boolean,
  agcOn: boolean,
  floatingNorm: boolean,
  agcDecayFactor: number,
): string {
  let s = '';
  if (Math.abs(bandwidth - DEFAULT_BANDWIDTH) > BANDWIDTH_EPSILON) {
    s += `, bw=${formatNum(bandwidth)}`;
  }
  if (!sdcOn) s += ', no-sdc';
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
