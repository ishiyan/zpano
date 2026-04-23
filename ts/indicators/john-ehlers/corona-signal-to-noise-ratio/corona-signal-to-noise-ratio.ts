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
import { Corona } from '../corona/corona';
import { CoronaSignalToNoiseRatioParams } from './params';

const DEFAULT_RASTER_LENGTH = 50;
const DEFAULT_MAX_RASTER = 20;
const DEFAULT_MIN_PARAM = 1;
const DEFAULT_MAX_PARAM = 11;
const DEFAULT_HP_CUTOFF = 30;
const DEFAULT_MIN_PERIOD = 6;
const DEFAULT_MAX_PERIOD = 30;

const HIGH_LOW_BUFFER_SIZE = 5;
const HIGH_LOW_BUFFER_SIZE_MIN_ONE = HIGH_LOW_BUFFER_SIZE - 1;
const HIGH_LOW_MEDIAN_INDEX = 2;
const AVERAGE_SAMPLE_ALPHA = 0.1;
const AVERAGE_SAMPLE_ONE_MINUS = 0.9;
const SIGNAL_EMA_ALPHA = 0.2;
const SIGNAL_EMA_ONE_MINUS = 0.9; // Intentional: sums to 1.1, per Ehlers.
const NOISE_EMA_ALPHA = 0.1;
const NOISE_EMA_ONE_MINUS = 0.9;
const RATIO_OFFSET_DB = 3.5;
const RATIO_UPPER_DB = 10;
const DB_GAIN = 20;
const WIDTH_LOW_RATIO_THRESHOLD = 0.5;
const WIDTH_BASELINE = 0.2;
const WIDTH_SLOPE = 0.4;
const RASTER_BLEND_EXPONENT = 0.8;
const RASTER_BLEND_HALF = 0.5;
const RASTER_NEGATIVE_ARG_CUTOFF = 1;

/** __Corona Signal-to-Noise Ratio__ (Ehlers) heatmap indicator.
 *
 * Measures cycle amplitude relative to noise, where "noise" is the average bar height
 * (there is not much trade information within a bar).
 *
 * It exposes two outputs:
 *
 *  - Value: a per-bar heatmap column (intensity raster).
 *  - SignalToNoiseRatio: the current SNR value mapped into [MinParameterValue, MaxParameterValue].
 *
 * Reference: John Ehlers, "Measuring Cycle Periods", Stocks & Commodities, November 2008.
 */
export class CoronaSignalToNoiseRatio implements Indicator {
  private readonly c: Corona;
  private readonly rasterLength: number;
  private readonly rasterStep: number;
  private readonly maxRasterValue: number;
  private readonly minParameterValue: number;
  private readonly maxParameterValue: number;
  private readonly parameterResolution: number;
  private readonly raster: number[];
  private readonly highLowBuffer: number[];
  private readonly hlSorted: number[];

  private readonly mnemonicValue: string;
  private readonly descriptionValue: string;
  private readonly mnemonicSNR: string;
  private readonly descriptionSNR: string;

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  private averageSamplePrevious = 0;
  private signalPrevious = 0;
  private noisePrevious = 0;
  private signalToNoiseRatio = Number.NaN;
  private isStarted = false;

  /** Creates an instance with default parameters. */
  public static default(): CoronaSignalToNoiseRatio {
    return new CoronaSignalToNoiseRatio({});
  }

  /** Creates an instance based on the given parameters. */
  public static fromParams(params: CoronaSignalToNoiseRatioParams): CoronaSignalToNoiseRatio {
    return new CoronaSignalToNoiseRatio(params);
  }

  private constructor(params: CoronaSignalToNoiseRatioParams) {
    const invalid = 'invalid corona signal to noise ratio parameters';

    const rasterLength = params.rasterLength !== undefined && params.rasterLength !== 0
      ? params.rasterLength : DEFAULT_RASTER_LENGTH;
    const maxRaster = params.maxRasterValue !== undefined && params.maxRasterValue !== 0
      ? params.maxRasterValue : DEFAULT_MAX_RASTER;
    const minParam = params.minParameterValue !== undefined && params.minParameterValue !== 0
      ? params.minParameterValue : DEFAULT_MIN_PARAM;
    const maxParam = params.maxParameterValue !== undefined && params.maxParameterValue !== 0
      ? params.maxParameterValue : DEFAULT_MAX_PARAM;
    const hpCutoff = params.highPassFilterCutoff !== undefined && params.highPassFilterCutoff !== 0
      ? params.highPassFilterCutoff : DEFAULT_HP_CUTOFF;
    const minPeriod = params.minimalPeriod !== undefined && params.minimalPeriod !== 0
      ? params.minimalPeriod : DEFAULT_MIN_PERIOD;
    const maxPeriod = params.maximalPeriod !== undefined && params.maximalPeriod !== 0
      ? params.maximalPeriod : DEFAULT_MAX_PERIOD;

    if (rasterLength < 2) {
      throw new Error(`${invalid}: RasterLength should be >= 2`);
    }
    if (maxRaster <= 0) {
      throw new Error(`${invalid}: MaxRasterValue should be > 0`);
    }
    if (minParam < 0) {
      throw new Error(`${invalid}: MinParameterValue should be >= 0`);
    }
    if (maxParam <= minParam) {
      throw new Error(`${invalid}: MaxParameterValue should be > MinParameterValue`);
    }
    if (hpCutoff < 2) {
      throw new Error(`${invalid}: HighPassFilterCutoff should be >= 2`);
    }
    if (minPeriod < 2) {
      throw new Error(`${invalid}: MinimalPeriod should be >= 2`);
    }
    if (maxPeriod <= minPeriod) {
      throw new Error(`${invalid}: MaximalPeriod should be > MinimalPeriod`);
    }

    const bc = params.barComponent ?? BarComponent.Median;
    const qc = params.quoteComponent ?? DefaultQuoteComponent;
    const tc = params.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    this.c = new Corona({
      highPassFilterCutoff: hpCutoff,
      minimalPeriod: minPeriod,
      maximalPeriod: maxPeriod,
    });

    this.rasterLength = rasterLength;
    this.rasterStep = maxRaster / rasterLength;
    this.maxRasterValue = maxRaster;
    this.minParameterValue = minParam;
    this.maxParameterValue = maxParam;
    this.parameterResolution = (rasterLength - 1) / (maxParam - minParam);

    this.raster = new Array<number>(rasterLength).fill(0);
    this.highLowBuffer = new Array<number>(HIGH_LOW_BUFFER_SIZE).fill(0);
    this.hlSorted = new Array<number>(HIGH_LOW_BUFFER_SIZE).fill(0);

    const cm = componentTripleMnemonic(bc, qc, tc);
    this.mnemonicValue = `csnr(${rasterLength}, ${maxRaster}, ${minParam}, ${maxParam}, ${hpCutoff}${cm})`;
    this.mnemonicSNR = `csnr-snr(${hpCutoff}${cm})`;
    this.descriptionValue = 'Corona signal to noise ratio ' + this.mnemonicValue;
    this.descriptionSNR = 'Corona signal to noise ratio scalar ' + this.mnemonicSNR;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean { return this.c.isPrimed(); }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.CoronaSignalToNoiseRatio,
      this.mnemonicValue,
      this.descriptionValue,
      [
        { mnemonic: this.mnemonicValue, description: this.descriptionValue },
        { mnemonic: this.mnemonicSNR, description: this.descriptionSNR },
      ],
    );
  }

  /** Feeds the next sample plus bar extremes and returns the heatmap column and the
   * current SignalToNoiseRatio. On unprimed bars the heatmap is empty and the scalar
   * is NaN. On NaN sample input state is left unchanged. */
  public update(sample: number, sampleLow: number, sampleHigh: number, time: Date): [Heatmap, number] {
    if (Number.isNaN(sample)) {
      return [
        Heatmap.newEmptyHeatmap(time, this.minParameterValue, this.maxParameterValue, this.parameterResolution),
        Number.NaN,
      ];
    }

    const primed = this.c.update(sample);

    if (!this.isStarted) {
      this.averageSamplePrevious = sample;
      this.highLowBuffer[HIGH_LOW_BUFFER_SIZE_MIN_ONE] = sampleHigh - sampleLow;
      this.isStarted = true;

      return [
        Heatmap.newEmptyHeatmap(time, this.minParameterValue, this.maxParameterValue, this.parameterResolution),
        Number.NaN,
      ];
    }

    const maxAmpSq = this.c.maximalAmplitudeSquared;

    const averageSample = AVERAGE_SAMPLE_ALPHA * sample + AVERAGE_SAMPLE_ONE_MINUS * this.averageSamplePrevious;
    this.averageSamplePrevious = averageSample;

    if (Math.abs(averageSample) > 0 || maxAmpSq > 0) {
      this.signalPrevious = SIGNAL_EMA_ALPHA * Math.sqrt(maxAmpSq) + SIGNAL_EMA_ONE_MINUS * this.signalPrevious;
    }

    // Shift H-L ring buffer left; push new value.
    for (let i = 0; i < HIGH_LOW_BUFFER_SIZE_MIN_ONE; i++) {
      this.highLowBuffer[i] = this.highLowBuffer[i + 1];
    }
    this.highLowBuffer[HIGH_LOW_BUFFER_SIZE_MIN_ONE] = sampleHigh - sampleLow;

    let ratio = 0;
    if (Math.abs(averageSample) > 0) {
      for (let i = 0; i < HIGH_LOW_BUFFER_SIZE; i++) {
        this.hlSorted[i] = this.highLowBuffer[i];
      }
      this.hlSorted.sort((a, b) => a - b);
      this.noisePrevious = NOISE_EMA_ALPHA * this.hlSorted[HIGH_LOW_MEDIAN_INDEX] + NOISE_EMA_ONE_MINUS * this.noisePrevious;

      if (Math.abs(this.noisePrevious) > 0) {
        ratio = DB_GAIN * Math.log10(this.signalPrevious / this.noisePrevious) + RATIO_OFFSET_DB;
        if (ratio < 0) {
          ratio = 0;
        } else if (ratio > RATIO_UPPER_DB) {
          ratio = RATIO_UPPER_DB;
        }
        ratio /= RATIO_UPPER_DB; // ∈ [0, 1]
      }
    }

    this.signalToNoiseRatio = (this.maxParameterValue - this.minParameterValue) * ratio + this.minParameterValue;

    // Raster update.
    let width = 0;
    if (ratio <= WIDTH_LOW_RATIO_THRESHOLD) {
      width = WIDTH_BASELINE - WIDTH_SLOPE * ratio;
    }

    const ratioScaledToRasterLength = Math.round(ratio * this.rasterLength);
    const ratioScaledToMaxRasterValue = ratio * this.maxRasterValue;

    for (let i = 0; i < this.rasterLength; i++) {
      let value = this.raster[i];

      if (i === ratioScaledToRasterLength) {
        value *= 0.5;
      } else if (width === 0) {
        // Above the high-ratio threshold: handled by the ratio>0.5 override below.
      } else {
        let argument = (ratioScaledToMaxRasterValue - this.rasterStep * i) / width;
        if (i < ratioScaledToRasterLength) {
          value = RASTER_BLEND_HALF * (Math.pow(argument, RASTER_BLEND_EXPONENT) + value);
        } else {
          argument = -argument;
          if (argument > RASTER_NEGATIVE_ARG_CUTOFF) {
            value = RASTER_BLEND_HALF * (Math.pow(argument, RASTER_BLEND_EXPONENT) + value);
          } else {
            value = this.maxRasterValue;
          }
        }
      }

      if (value < 0) {
        value = 0;
      } else if (value > this.maxRasterValue) {
        value = this.maxRasterValue;
      }

      if (ratio > WIDTH_LOW_RATIO_THRESHOLD) {
        value = this.maxRasterValue;
      }

      this.raster[i] = value;
    }

    if (!primed) {
      return [
        Heatmap.newEmptyHeatmap(time, this.minParameterValue, this.maxParameterValue, this.parameterResolution),
        Number.NaN,
      ];
    }

    const values = new Array<number>(this.rasterLength);
    let valueMin = Number.POSITIVE_INFINITY;
    let valueMax = Number.NEGATIVE_INFINITY;

    for (let i = 0; i < this.rasterLength; i++) {
      const v = this.raster[i];
      values[i] = v;
      if (v < valueMin) valueMin = v;
      if (v > valueMax) valueMax = v;
    }

    const heatmap = Heatmap.newHeatmap(
      time, this.minParameterValue, this.maxParameterValue, this.parameterResolution,
      valueMin, valueMax, values,
    );

    return [heatmap, this.signalToNoiseRatio];
  }

  /** Updates the indicator given the next scalar sample. Since no High/Low is
   * available, the sample is used for both, yielding zero noise. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    return this.updateEntity(sample.time, sample.value, sample.value, sample.value);
  }

  /** Updates the indicator given the next bar sample. */
  public updateBar(sample: Bar): IndicatorOutput {
    return this.updateEntity(sample.time, this.barComponentFunc(sample), sample.low, sample.high);
  }

  /** Updates the indicator given the next quote sample. */
  public updateQuote(sample: Quote): IndicatorOutput {
    const v = this.quoteComponentFunc(sample);
    return this.updateEntity(sample.time, v, v, v);
  }

  /** Updates the indicator given the next trade sample. */
  public updateTrade(sample: Trade): IndicatorOutput {
    const v = this.tradeComponentFunc(sample);
    return this.updateEntity(sample.time, v, v, v);
  }

  private updateEntity(time: Date, sample: number, low: number, high: number): IndicatorOutput {
    const [heatmap, snr] = this.update(sample, low, high, time);

    const s = new Scalar();
    s.time = time;
    s.value = snr;

    return [heatmap, s];
  }
}
