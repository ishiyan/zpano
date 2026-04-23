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
import { CoronaSpectrumParams } from './params';

const DEFAULT_MIN_RASTER = 6;
const DEFAULT_MAX_RASTER = 20;
const DEFAULT_MIN_PARAM = 6;
const DEFAULT_MAX_PARAM = 30;
const DEFAULT_HP_CUTOFF = 30;

/** __Corona Spectrum__ (Ehlers) heatmap indicator.
 *
 * The Corona Spectrum measures cyclic activity over a cycle period range (default 6..30 bars)
 * in a bank of contiguous bandpass filters. The amplitude of each filter output is compared to
 * the strongest signal and displayed, in decibels, as a heatmap column. The filter having the
 * strongest output is selected as the current dominant cycle period.
 *
 * It exposes three outputs:
 *
 *	- Value: a per-bar heatmap column (decibels across the filter bank).
 *	- DominantCycle: the weighted-center-of-gravity dominant cycle estimate.
 *	- DominantCycleMedian: the 5-sample median of DominantCycle.
 *
 * Reference:
 *
 *	John Ehlers, "Measuring Cycle Periods", Stocks & Commodities, November 2008.
 */
export class CoronaSpectrum implements Indicator {
  private readonly c: Corona;
  private readonly minParameterValue: number;
  private readonly maxParameterValue: number;
  private readonly parameterResolution: number;

  private readonly mnemonicValue: string;
  private readonly descriptionValue: string;
  private readonly mnemonicDC: string;
  private readonly descriptionDC: string;
  private readonly mnemonicDCM: string;
  private readonly descriptionDCM: string;

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  /** Creates an instance with default parameters
   * (minRaster=6, maxRaster=20, minParam=6, maxParam=30, hpCutoff=30, BarComponent.Median). */
  public static default(): CoronaSpectrum {
    return new CoronaSpectrum({});
  }

  /** Creates an instance based on the given parameters. */
  public static fromParams(params: CoronaSpectrumParams): CoronaSpectrum {
    return new CoronaSpectrum(params);
  }

  private constructor(params: CoronaSpectrumParams) {
    const invalid = 'invalid corona spectrum parameters';

    const minRaster = params.minRasterValue !== undefined && params.minRasterValue !== 0
      ? params.minRasterValue : DEFAULT_MIN_RASTER;
    const maxRaster = params.maxRasterValue !== undefined && params.maxRasterValue !== 0
      ? params.maxRasterValue : DEFAULT_MAX_RASTER;
    const minParamRaw = params.minParameterValue !== undefined && params.minParameterValue !== 0
      ? params.minParameterValue : DEFAULT_MIN_PARAM;
    const maxParamRaw = params.maxParameterValue !== undefined && params.maxParameterValue !== 0
      ? params.maxParameterValue : DEFAULT_MAX_PARAM;
    const hpCutoff = params.highPassFilterCutoff !== undefined && params.highPassFilterCutoff !== 0
      ? params.highPassFilterCutoff : DEFAULT_HP_CUTOFF;

    if (minRaster < 0) {
      throw new Error(`${invalid}: MinRasterValue should be >= 0`);
    }
    if (maxRaster <= minRaster) {
      throw new Error(`${invalid}: MaxRasterValue should be > MinRasterValue`);
    }

    // MBST rounds min up and max down to integers.
    const minParam = Math.ceil(minParamRaw);
    const maxParam = Math.floor(maxParamRaw);

    if (minParam < 2) {
      throw new Error(`${invalid}: MinParameterValue should be >= 2`);
    }
    if (maxParam <= minParam) {
      throw new Error(`${invalid}: MaxParameterValue should be > MinParameterValue`);
    }
    if (hpCutoff < 2) {
      throw new Error(`${invalid}: HighPassFilterCutoff should be >= 2`);
    }

    // CoronaSpectrum mirrors Ehlers' reference: BarComponent.Median default.
    const bc = params.barComponent ?? BarComponent.Median;
    const qc = params.quoteComponent ?? DefaultQuoteComponent;
    const tc = params.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    this.c = new Corona({
      highPassFilterCutoff: hpCutoff,
      minimalPeriod: minParam,
      maximalPeriod: maxParam,
      decibelsLowerThreshold: minRaster,
      decibelsUpperThreshold: maxRaster,
    });

    this.minParameterValue = minParam;
    this.maxParameterValue = maxParam;

    // Values slice length = filterBankLength; first sample at minParam, last at maxParam.
    this.parameterResolution = (this.c.filterBankLength - 1) / (maxParam - minParam);

    const cm = componentTripleMnemonic(bc, qc, tc);
    this.mnemonicValue = `cspect(${minRaster}, ${maxRaster}, ${minParam}, ${maxParam}, ${hpCutoff}${cm})`;
    this.mnemonicDC = `cspect-dc(${hpCutoff}${cm})`;
    this.mnemonicDCM = `cspect-dcm(${hpCutoff}${cm})`;

    this.descriptionValue = 'Corona spectrum ' + this.mnemonicValue;
    this.descriptionDC = 'Corona spectrum dominant cycle ' + this.mnemonicDC;
    this.descriptionDCM = 'Corona spectrum dominant cycle median ' + this.mnemonicDCM;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean { return this.c.isPrimed(); }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.CoronaSpectrum,
      this.mnemonicValue,
      this.descriptionValue,
      [
        { mnemonic: this.mnemonicValue, description: this.descriptionValue },
        { mnemonic: this.mnemonicDC, description: this.descriptionDC },
        { mnemonic: this.mnemonicDCM, description: this.descriptionDCM },
      ],
    );
  }

  /** Feeds the next sample to the engine and returns the heatmap column plus the current
   * DominantCycle and DominantCycleMedian estimates.
   *
   * On unprimed bars the heatmap is an empty heatmap (with the indicator's parameter axis)
   * and both scalar values are NaN. On NaN input, state is left unchanged and all outputs
   * are NaN / empty heatmap. */
  public update(sample: number, time: Date): [Heatmap, number, number] {
    if (Number.isNaN(sample)) {
      return [
        Heatmap.newEmptyHeatmap(time, this.minParameterValue, this.maxParameterValue, this.parameterResolution),
        Number.NaN, Number.NaN,
      ];
    }

    const primed = this.c.update(sample);
    if (!primed) {
      return [
        Heatmap.newEmptyHeatmap(time, this.minParameterValue, this.maxParameterValue, this.parameterResolution),
        Number.NaN, Number.NaN,
      ];
    }

    const bank = this.c.filterBank;
    const values = new Array<number>(bank.length);
    let valueMin = Number.POSITIVE_INFINITY;
    let valueMax = Number.NEGATIVE_INFINITY;

    for (let i = 0; i < bank.length; i++) {
      const v = bank[i].decibels;
      values[i] = v;
      if (v < valueMin) valueMin = v;
      if (v > valueMax) valueMax = v;
    }

    const heatmap = Heatmap.newHeatmap(
      time, this.minParameterValue, this.maxParameterValue, this.parameterResolution,
      valueMin, valueMax, values,
    );

    return [heatmap, this.c.dominantCycle, this.c.dominantCycleMedian];
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
    const [heatmap, dc, dcm] = this.update(sample, time);

    const sDc = new Scalar();
    sDc.time = time;
    sDc.value = dc;

    const sDcm = new Scalar();
    sDcm.time = time;
    sDcm.value = dcm;

    return [heatmap, sDc, sDcm];
  }
}
