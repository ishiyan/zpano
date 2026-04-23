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
import { CoronaTrendVigorParams } from './params';

const DEFAULT_RASTER_LENGTH = 50;
const DEFAULT_MAX_RASTER = 20;
const DEFAULT_MIN_PARAM = -10;
const DEFAULT_MAX_PARAM = 10;
const DEFAULT_HP_CUTOFF = 30;
const DEFAULT_MIN_PERIOD = 6;
const DEFAULT_MAX_PERIOD = 30;

const BP_DELTA = 0.1;
const RATIO_NEW_COEF = 0.33;
const RATIO_PREVIOUS_COEF = 0.67;

const VIGOR_MID_LOW = 0.3;
const VIGOR_MID_HIGH = 0.7;
const VIGOR_MID = 0.5;
const WIDTH_EDGE = 0.01;

const RASTER_BLEND_SCALE = 0.8;
const RASTER_BLEND_PREVIOUS = 0.2;
const RASTER_BLEND_HALF = 0.5;
const RASTER_BLEND_EXPONENT = 0.85;

const RATIO_LIMIT = 10;
const VIGOR_SCALE = 0.05;

/** __Corona Trend Vigor__ (Ehlers) heatmap indicator.
 *
 * Slope of momentum over a full dominant cycle period, normalized by the cycle amplitude and
 * scaled into [-10, 10]. Values between -2 and +2 form the "corona" (do not trade the trend).
 *
 * Reference: John Ehlers, "Measuring Cycle Periods", Stocks & Commodities, November 2008.
 */
export class CoronaTrendVigor implements Indicator {
  private readonly c: Corona;
  private readonly rasterLength: number;
  private readonly rasterStep: number;
  private readonly maxRasterValue: number;
  private readonly minParameterValue: number;
  private readonly maxParameterValue: number;
  private readonly parameterResolution: number;
  private readonly raster: number[];
  private readonly sampleBuffer: number[];

  private readonly mnemonicValue: string;
  private readonly descriptionValue: string;
  private readonly mnemonicTV: string;
  private readonly descriptionTV: string;

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  private sampleCount = 0;
  private samplePrevious = 0;
  private samplePrevious2 = 0;
  private bandPassPrevious = 0;
  private bandPassPrevious2 = 0;
  private ratioPrevious = 0;
  private trendVigor = Number.NaN;

  public static default(): CoronaTrendVigor {
    return new CoronaTrendVigor({});
  }

  public static fromParams(params: CoronaTrendVigorParams): CoronaTrendVigor {
    return new CoronaTrendVigor(params);
  }

  private constructor(params: CoronaTrendVigorParams) {
    const invalid = 'invalid corona trend vigor parameters';

    const rasterLength = params.rasterLength !== undefined && params.rasterLength !== 0
      ? params.rasterLength : DEFAULT_RASTER_LENGTH;
    const maxRaster = params.maxRasterValue !== undefined && params.maxRasterValue !== 0
      ? params.maxRasterValue : DEFAULT_MAX_RASTER;

    let minParam = params.minParameterValue ?? 0;
    let maxParam = params.maxParameterValue ?? 0;
    if (minParam === 0 && maxParam === 0) {
      minParam = DEFAULT_MIN_PARAM;
      maxParam = DEFAULT_MAX_PARAM;
    }

    const hpCutoff = params.highPassFilterCutoff !== undefined && params.highPassFilterCutoff !== 0
      ? params.highPassFilterCutoff : DEFAULT_HP_CUTOFF;
    const minPeriod = params.minimalPeriod !== undefined && params.minimalPeriod !== 0
      ? params.minimalPeriod : DEFAULT_MIN_PERIOD;
    const maxPeriod = params.maximalPeriod !== undefined && params.maximalPeriod !== 0
      ? params.maximalPeriod : DEFAULT_MAX_PERIOD;

    if (rasterLength < 2) throw new Error(`${invalid}: RasterLength should be >= 2`);
    if (maxRaster <= 0) throw new Error(`${invalid}: MaxRasterValue should be > 0`);
    if (maxParam <= minParam) throw new Error(`${invalid}: MaxParameterValue should be > MinParameterValue`);
    if (hpCutoff < 2) throw new Error(`${invalid}: HighPassFilterCutoff should be >= 2`);
    if (minPeriod < 2) throw new Error(`${invalid}: MinimalPeriod should be >= 2`);
    if (maxPeriod <= minPeriod) throw new Error(`${invalid}: MaximalPeriod should be > MinimalPeriod`);

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
    this.sampleBuffer = new Array<number>(this.c.maximalPeriodTimesTwo).fill(0);

    const cm = componentTripleMnemonic(bc, qc, tc);
    this.mnemonicValue = `ctv(${rasterLength}, ${maxRaster}, ${minParam}, ${maxParam}, ${hpCutoff}${cm})`;
    this.mnemonicTV = `ctv-tv(${hpCutoff}${cm})`;
    this.descriptionValue = 'Corona trend vigor ' + this.mnemonicValue;
    this.descriptionTV = 'Corona trend vigor scalar ' + this.mnemonicTV;
  }

  public isPrimed(): boolean { return this.c.isPrimed(); }

  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.CoronaTrendVigor,
      this.mnemonicValue,
      this.descriptionValue,
      [
        { mnemonic: this.mnemonicValue, description: this.descriptionValue },
        { mnemonic: this.mnemonicTV, description: this.descriptionTV },
      ],
    );
  }

  /** Feeds the next sample and returns the heatmap column plus the current TrendVigor. */
  public update(sample: number, time: Date): [Heatmap, number] {
    if (Number.isNaN(sample)) {
      return [
        Heatmap.newEmptyHeatmap(time, this.minParameterValue, this.maxParameterValue, this.parameterResolution),
        Number.NaN,
      ];
    }

    const primed = this.c.update(sample);
    this.sampleCount++;

    const bufLast = this.sampleBuffer.length - 1;

    if (this.sampleCount === 1) {
      this.samplePrevious = sample;
      this.sampleBuffer[bufLast] = sample;
      return [
        Heatmap.newEmptyHeatmap(time, this.minParameterValue, this.maxParameterValue, this.parameterResolution),
        Number.NaN,
      ];
    }

    // Bandpass InPhase filter at dominant cycle median period.
    const omega = 2 * Math.PI / this.c.dominantCycleMedian;
    const beta2 = Math.cos(omega);
    const gamma2 = 1 / Math.cos(omega * 2 * BP_DELTA);
    const alpha2 = gamma2 - Math.sqrt(gamma2 * gamma2 - 1);
    const bandPass = 0.5 * (1 - alpha2) * (sample - this.samplePrevious2)
      + beta2 * (1 + alpha2) * this.bandPassPrevious
      - alpha2 * this.bandPassPrevious2;

    const quadrature2 = (bandPass - this.bandPassPrevious) / omega;

    this.bandPassPrevious2 = this.bandPassPrevious;
    this.bandPassPrevious = bandPass;
    this.samplePrevious2 = this.samplePrevious;
    this.samplePrevious = sample;

    // Left-shift sampleBuffer and append the new sample.
    for (let i = 0; i < bufLast; i++) {
      this.sampleBuffer[i] = this.sampleBuffer[i + 1];
    }
    this.sampleBuffer[bufLast] = sample;

    const amplitude2 = Math.sqrt(bandPass * bandPass + quadrature2 * quadrature2);

    // DominantCycleMedian-1 directly; clamp to [1, sampleBuffer.length].
    let cyclePeriod = Math.trunc(this.c.dominantCycleMedian - 1);
    if (cyclePeriod > this.sampleBuffer.length) cyclePeriod = this.sampleBuffer.length;
    if (cyclePeriod < 1) cyclePeriod = 1;

    let lookback = cyclePeriod;
    if (this.sampleCount < lookback) lookback = this.sampleCount;

    const trend = sample - this.sampleBuffer[this.sampleBuffer.length - lookback];

    let ratio = 0;
    if (Math.abs(trend) > 0 && amplitude2 > 0) {
      ratio = RATIO_NEW_COEF * trend / amplitude2 + RATIO_PREVIOUS_COEF * this.ratioPrevious;
    }

    if (ratio > RATIO_LIMIT) ratio = RATIO_LIMIT;
    else if (ratio < -RATIO_LIMIT) ratio = -RATIO_LIMIT;

    this.ratioPrevious = ratio;

    const vigor = VIGOR_SCALE * (ratio + RATIO_LIMIT);

    let width: number;
    if (vigor >= VIGOR_MID_LOW && vigor < VIGOR_MID) {
      width = vigor - (VIGOR_MID_LOW - WIDTH_EDGE);
    } else if (vigor >= VIGOR_MID && vigor <= VIGOR_MID_HIGH) {
      width = (VIGOR_MID_HIGH + WIDTH_EDGE) - vigor;
    } else {
      width = WIDTH_EDGE;
    }

    this.trendVigor = (this.maxParameterValue - this.minParameterValue) * vigor + this.minParameterValue;

    const vigorScaledToRasterLength = Math.round(this.rasterLength * vigor);
    const vigorScaledToMaxRasterValue = vigor * this.maxRasterValue;

    for (let i = 0; i < this.rasterLength; i++) {
      let value = this.raster[i];

      if (i === vigorScaledToRasterLength) {
        value *= RASTER_BLEND_HALF;
      } else {
        let argument = vigorScaledToMaxRasterValue - this.rasterStep * i;
        if (i > vigorScaledToRasterLength) argument = -argument;
        if (width > 0) {
          value = RASTER_BLEND_SCALE * (Math.pow(argument / width, RASTER_BLEND_EXPONENT) + RASTER_BLEND_PREVIOUS * value);
        }
      }

      if (value < 0) {
        value = 0;
      } else if (value > this.maxRasterValue || vigor < VIGOR_MID_LOW || vigor > VIGOR_MID_HIGH) {
        value = this.maxRasterValue;
      }

      if (Number.isNaN(value)) value = 0;

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

    return [heatmap, this.trendVigor];
  }

  public updateScalar(sample: Scalar): IndicatorOutput { return this.updateEntity(sample.time, sample.value); }
  public updateBar(sample: Bar): IndicatorOutput { return this.updateEntity(sample.time, this.barComponentFunc(sample)); }
  public updateQuote(sample: Quote): IndicatorOutput { return this.updateEntity(sample.time, this.quoteComponentFunc(sample)); }
  public updateTrade(sample: Trade): IndicatorOutput { return this.updateEntity(sample.time, this.tradeComponentFunc(sample)); }

  private updateEntity(time: Date, sample: number): IndicatorOutput {
    const [heatmap, tv] = this.update(sample, time);
    const s = new Scalar();
    s.time = time;
    s.value = tv;
    return [heatmap, s];
  }
}
