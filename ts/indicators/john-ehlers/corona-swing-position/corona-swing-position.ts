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
import { CoronaSwingPositionParams } from './params';

const DEFAULT_RASTER_LENGTH = 50;
const DEFAULT_MAX_RASTER = 20;
const DEFAULT_MIN_PARAM = -5;
const DEFAULT_MAX_PARAM = 5;
const DEFAULT_HP_CUTOFF = 30;
const DEFAULT_MIN_PERIOD = 6;
const DEFAULT_MAX_PERIOD = 30;

const MAX_LEAD_LIST_COUNT = 50;
const MAX_POSITION_LIST_COUNT = 20;

// 60° phase-lead coefficients.
const LEAD60_COEF_BP = 0.5;
const LEAD60_COEF_Q = 0.866;

const BP_DELTA = 0.1;

const WIDTH_HIGH_THRESHOLD = 0.85;
const WIDTH_HIGH_SATURATE = 0.8;
const WIDTH_NARROW = 0.01;

const RASTER_BLEND_EXPONENT = 0.95;
const RASTER_BLEND_HALF = 0.5;

/** __Corona Swing Position__ (Ehlers) heatmap indicator.
 *
 * Correlates prices with a perfect sine wave having the dominant cycle period, producing a
 * smooth waveform that lets us better estimate the swing position and impending turning points.
 *
 * Reference: John Ehlers, "Measuring Cycle Periods", Stocks & Commodities, November 2008.
 */
export class CoronaSwingPosition implements Indicator {
  private readonly c: Corona;
  private readonly rasterLength: number;
  private readonly rasterStep: number;
  private readonly maxRasterValue: number;
  private readonly minParameterValue: number;
  private readonly maxParameterValue: number;
  private readonly parameterResolution: number;
  private readonly raster: number[];
  private readonly leadList: number[] = [];
  private readonly positionList: number[] = [];

  private readonly mnemonicValue: string;
  private readonly descriptionValue: string;
  private readonly mnemonicSP: string;
  private readonly descriptionSP: string;

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  private samplePrevious = 0;
  private samplePrevious2 = 0;
  private bandPassPrevious = 0;
  private bandPassPrevious2 = 0;
  private swingPosition = Number.NaN;
  private isStarted = false;

  public static default(): CoronaSwingPosition {
    return new CoronaSwingPosition({});
  }

  public static fromParams(params: CoronaSwingPositionParams): CoronaSwingPosition {
    return new CoronaSwingPosition(params);
  }

  private constructor(params: CoronaSwingPositionParams) {
    const invalid = 'invalid corona swing position parameters';

    const rasterLength = params.rasterLength !== undefined && params.rasterLength !== 0
      ? params.rasterLength : DEFAULT_RASTER_LENGTH;
    const maxRaster = params.maxRasterValue !== undefined && params.maxRasterValue !== 0
      ? params.maxRasterValue : DEFAULT_MAX_RASTER;

    // Only substitute Min/Max when both are 0 (unconfigured), since 0 is a valid user value.
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

    const cm = componentTripleMnemonic(bc, qc, tc);
    this.mnemonicValue = `cswing(${rasterLength}, ${maxRaster}, ${minParam}, ${maxParam}, ${hpCutoff}${cm})`;
    this.mnemonicSP = `cswing-sp(${hpCutoff}${cm})`;
    this.descriptionValue = 'Corona swing position ' + this.mnemonicValue;
    this.descriptionSP = 'Corona swing position scalar ' + this.mnemonicSP;
  }

  public isPrimed(): boolean { return this.c.isPrimed(); }

  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.CoronaSwingPosition,
      this.mnemonicValue,
      this.descriptionValue,
      [
        { mnemonic: this.mnemonicValue, description: this.descriptionValue },
        { mnemonic: this.mnemonicSP, description: this.descriptionSP },
      ],
    );
  }

  /** Feeds the next sample and returns the heatmap column plus the current SwingPosition. */
  public update(sample: number, time: Date): [Heatmap, number] {
    if (Number.isNaN(sample)) {
      return [
        Heatmap.newEmptyHeatmap(time, this.minParameterValue, this.maxParameterValue, this.parameterResolution),
        Number.NaN,
      ];
    }

    const primed = this.c.update(sample);

    if (!this.isStarted) {
      this.samplePrevious = sample;
      this.isStarted = true;
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

    // 60° lead.
    const lead60 = LEAD60_COEF_BP * this.bandPassPrevious2 + LEAD60_COEF_Q * quadrature2;

    let [lowest, highest] = appendRolling(this.leadList, MAX_LEAD_LIST_COUNT, lead60);

    let position = highest - lowest;
    if (position > 0) {
      position = (lead60 - lowest) / position;
    }

    [lowest, highest] = appendRolling(this.positionList, MAX_POSITION_LIST_COUNT, position);
    highest -= lowest;

    let width = 0.15 * highest;
    if (highest > WIDTH_HIGH_THRESHOLD) {
      width = WIDTH_NARROW;
    }

    this.swingPosition = (this.maxParameterValue - this.minParameterValue) * position + this.minParameterValue;

    const positionScaledToRasterLength = Math.round(position * this.rasterLength);
    const positionScaledToMaxRasterValue = position * this.maxRasterValue;

    for (let i = 0; i < this.rasterLength; i++) {
      let value = this.raster[i];

      if (i === positionScaledToRasterLength) {
        value *= RASTER_BLEND_HALF;
      } else {
        let argument = positionScaledToMaxRasterValue - this.rasterStep * i;
        if (i > positionScaledToRasterLength) {
          argument = -argument;
        }
        if (width > 0) {
          value = RASTER_BLEND_HALF * (Math.pow(argument / width, RASTER_BLEND_EXPONENT) + RASTER_BLEND_HALF * value);
        }
      }

      if (value < 0) {
        value = 0;
      } else if (value > this.maxRasterValue) {
        value = this.maxRasterValue;
      }

      if (highest > WIDTH_HIGH_SATURATE) {
        value = this.maxRasterValue;
      }

      if (Number.isNaN(value)) {
        value = 0;
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

    return [heatmap, this.swingPosition];
  }

  public updateScalar(sample: Scalar): IndicatorOutput { return this.updateEntity(sample.time, sample.value); }
  public updateBar(sample: Bar): IndicatorOutput { return this.updateEntity(sample.time, this.barComponentFunc(sample)); }
  public updateQuote(sample: Quote): IndicatorOutput { return this.updateEntity(sample.time, this.quoteComponentFunc(sample)); }
  public updateTrade(sample: Trade): IndicatorOutput { return this.updateEntity(sample.time, this.tradeComponentFunc(sample)); }

  private updateEntity(time: Date, sample: number): IndicatorOutput {
    const [heatmap, sp] = this.update(sample, time);
    const s = new Scalar();
    s.time = time;
    s.value = sp;
    return [heatmap, s];
  }
}

/** Appends v to the list, drops the oldest once len reaches maxCount, returns [lowest, highest]. */
function appendRolling(list: number[], maxCount: number, v: number): [number, number] {
  if (list.length >= maxCount) {
    list.shift();
  }
  list.push(v);

  let lowest = v;
  let highest = v;
  for (let i = 0; i < list.length; i++) {
    const x = list[i];
    if (x < lowest) lowest = x;
    if (x > highest) highest = x;
  }
  return [lowest, highest];
}
