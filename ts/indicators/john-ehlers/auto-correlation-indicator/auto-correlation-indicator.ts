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
import { AutoCorrelationIndicatorEstimator } from './estimator';
import { AutoCorrelationIndicatorParams } from './params';

const DEFAULT_MIN_LAG = 3;
const DEFAULT_MAX_LAG = 48;
const DEFAULT_SMOOTHING_PERIOD = 10;
const DEFAULT_AVERAGING_LENGTH = 0;

/** __Autocorrelation Indicator__ heatmap (Ehlers).
 *
 * Displays a heatmap of Pearson correlation coefficients between the current filtered
 * series and a lagged copy of itself, across a configurable lag range. The close series
 * is pre-conditioned by a 2-pole Butterworth highpass (cutoff = maxLag) followed by a
 * 2-pole Super Smoother (cutoff = smoothingPeriod) before the correlation bank is
 * evaluated. Each bin's value is rescaled from the Pearson [-1, 1] range into [0, 1]
 * via 0.5*(r + 1) for direct display.
 *
 * This implementation follows John Ehlers' EasyLanguage listing 8-2 from
 * "Cycle Analytics for Traders". It is NOT a port of MBST's AutoCorrelationCoefficients /
 * AutoCorrelationEstimator, which omit the HP + SS pre-filter, use a different Pearson
 * formulation, and have an opposite AverageLength=0 convention.
 *
 * Reference: John F. Ehlers, "Cycle Analytics for Traders", Code Listing 8-2. */
export class AutoCorrelationIndicator implements Indicator {
  private readonly estimator: AutoCorrelationIndicatorEstimator;
  private readonly primeCount: number;
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
   * (minLag=3, maxLag=48, smoothingPeriod=10, averagingLength=0 (M=lag), BarComponent.Median). */
  public static default(): AutoCorrelationIndicator {
    return new AutoCorrelationIndicator({});
  }

  /** Creates an instance based on the given parameters. */
  public static fromParams(params: AutoCorrelationIndicatorParams): AutoCorrelationIndicator {
    return new AutoCorrelationIndicator(params);
  }

  private constructor(params: AutoCorrelationIndicatorParams) {
    const invalid = 'invalid autocorrelation indicator parameters';

    const minLag = params.minLag !== undefined && params.minLag !== 0
      ? params.minLag : DEFAULT_MIN_LAG;
    const maxLag = params.maxLag !== undefined && params.maxLag !== 0
      ? params.maxLag : DEFAULT_MAX_LAG;
    const smoothingPeriod = params.smoothingPeriod !== undefined && params.smoothingPeriod !== 0
      ? params.smoothingPeriod : DEFAULT_SMOOTHING_PERIOD;
    const averagingLength = params.averagingLength ?? DEFAULT_AVERAGING_LENGTH;

    if (minLag < 1) {
      throw new Error(`${invalid}: MinLag should be >= 1`);
    }
    if (maxLag <= minLag) {
      throw new Error(`${invalid}: MaxLag should be > MinLag`);
    }
    if (smoothingPeriod < 2) {
      throw new Error(`${invalid}: SmoothingPeriod should be >= 2`);
    }
    if (averagingLength < 0) {
      throw new Error(`${invalid}: AveragingLength should be >= 0`);
    }

    // AutoCorrelationIndicator mirrors Ehlers' reference: BarComponent.Median default.
    const bc = params.barComponent ?? BarComponent.Median;
    const qc = params.quoteComponent ?? DefaultQuoteComponent;
    const tc = params.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    this.estimator = new AutoCorrelationIndicatorEstimator(
      minLag, maxLag, smoothingPeriod, averagingLength,
    );
    this.primeCount = this.estimator.filtBufferLen;
    this.minParameterValue = minLag;
    this.maxParameterValue = maxLag;
    this.parameterResolution = 1;

    const cm = componentTripleMnemonic(bc, qc, tc);
    const flags = averagingLength !== DEFAULT_AVERAGING_LENGTH ? `, average=${averagingLength}` : '';
    this.mnemonicValue = `aci(${minLag}, ${maxLag}, ${smoothingPeriod}${flags}${cm})`;
    this.descriptionValue = 'Autocorrelation indicator ' + this.mnemonicValue;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean { return this.primed; }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.AutoCorrelationIndicator,
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

    // Estimator spectrum is already in [0, 1] via 0.5*(r + 1); no additional normalization.
    const values = new Array<number>(lengthSpectrum);
    let valueMin = Number.POSITIVE_INFINITY;
    let valueMax = Number.NEGATIVE_INFINITY;

    for (let i = 0; i < lengthSpectrum; i++) {
      const v = this.estimator.spectrum[i];
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
