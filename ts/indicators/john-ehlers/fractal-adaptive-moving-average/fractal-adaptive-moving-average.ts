import { buildMetadata } from '../../core/build-metadata';
import { Bar } from '../../../entities/bar';
import { BarComponent, DefaultBarComponent, barComponentValue } from '../../../entities/bar-component';
import { Quote } from '../../../entities/quote';
import { QuoteComponent, DefaultQuoteComponent, quoteComponentValue } from '../../../entities/quote-component';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { TradeComponent, DefaultTradeComponent, tradeComponentValue } from '../../../entities/trade-component';
import { Indicator } from '../../core/indicator';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { FractalAdaptiveMovingAverageParams } from './params';

/** __Fractal Adaptive Moving Average__ (Ehler's fractal adaptive moving average, _FRAMA_)
 * is an EMA with the smoothing factor, a, being changed with each new sample:
 *
 *	FRAMAi = aiPi + (1 - ai)*FRAMAi-1,  as <= ai <= 1
 *
 * Here the as is the slowest a (default suggested value is 0.01 or equivalent length of 199 samples).
 *
 * The concept of _FRAMA_ is to relate the fractal dimension FDi, calculated on window
 * samples, to the EMA smoothing factor ai, thus making the EMA adaptive.
 *
 * This dependency is defined as follows:
 *
 *	ai = exp(w(FDi - 1)),  1 <= FDi <= 2,
 *
 *	w = ln(as)
 *
 *	or, given the length ls = 2/as - 1,
 *
 *	w = ln(2/(ls + 1))
 *
 * The fractal dimension varies over the range from 1 to 2.
 *
 * When FDi = 1 (series forms a straight line), the exponent is zero - which means that
 * ai = 1, and the output of the exponential moving average is equal to the input.
 *
 * When FDi = 2 (series fills all plane, excibiting extreme volatility), the exponent
 * is w, which means that ai = as, and the output of the exponential moving average
 * is equal to the output of the slowest moving average with as.
 *
 * The fractal dimension is estimated by using a "box count" method (Falconer 2014 chapter 2).
 *
 *	FDi = (ln(N1+N2) - ln(N3)) / ln(2)
 *
 * Reference:
 *
 *	Falconer, K. (2014). Fractal geometry: Mathematical foundations and applications (3rd ed) Wiley.
 *	Ehlers, John F. (2005). Fractal Adaptive Moving Average. Technical Analysis of Stocks & Commodities, 23(10), 81-82.
 *	Ehlers, John F. (2006). FRAMA - Fractal Adaptive Moving Average, https://www.mesasoftware.com/papers/FRAMA.pdf.
 */
export class FractalAdaptiveMovingAverage implements Indicator {
  private readonly length: number;
  private readonly lengthMinOne: number;
  private readonly halfLength: number;
  private readonly alphaSlowest: number;
  private readonly scalingFactor: number;
  private windowCount: number = 0;
  private value: number = Number.NaN;
  private fractalDimension: number = Number.NaN;
  private readonly windowHigh: number[];
  private readonly windowLow: number[];
  private mnemonic: string;
  private description: string;
  private mnemonicFdim: string;
  private descriptionFdim: string;
  private primed: boolean = false;

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  /**
   * Constructs an instance given input parameters.
   */
  public constructor(params: FractalAdaptiveMovingAverageParams) {
    let len = Math.floor(params.length);
    if (len < 2) {
      throw new Error('length should be an even integer larger than 1');
    }

    if (len % 2 !== 0) {
      len++;
    }

    const alpha = params.slowestSmoothingFactor;
    if (alpha < 0 || alpha > 1) {
      throw new Error('slowest smoothing factor should be in range [0, 1]');
    }

    this.length = len;
    this.lengthMinOne = len - 1;
    this.halfLength = Math.floor(len / 2);
    this.alphaSlowest = alpha;
    this.scalingFactor = Math.log(alpha);
    this.windowHigh = new Array(len).fill(0);
    this.windowLow = new Array(len).fill(0);

    // Resolve component defaults and create component functions.
    const bc = params.barComponent ?? DefaultBarComponent;
    const qc = params.quoteComponent ?? DefaultQuoteComponent;
    const tc = params.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    // Build mnemonics matching Go format: frama(%d, %.3f%s)
    const cm = componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent);
    this.mnemonic = `frama(${len}, ${alpha.toFixed(3)}${cm})`;
    this.mnemonicFdim = `framaDim(${len}, ${alpha.toFixed(3)}${cm})`;

    const descr = 'Fractal adaptive moving average ';
    this.description = descr + this.mnemonic;
    this.descriptionFdim = descr + this.mnemonicFdim;
  }

  /** Indicates whether an indicator is primed. */
  public isPrimed(): boolean { return this.primed; }

  /** Describes a requested output data of an indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.FractalAdaptiveMovingAverage,
      this.mnemonic,
      this.description,
      [
        { mnemonic: this.mnemonic, description: this.description },
        { mnemonic: this.mnemonicFdim, description: this.descriptionFdim },
      ],
    );
  }

  /** Updates an indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    const v = sample.value;

    return this.updateEntity(sample.time, v, v, v);
  }

  /** Updates an indicator given the next bar sample. */
  public updateBar(sample: Bar): IndicatorOutput {
    const v = this.barComponentFunc(sample);

    return this.updateEntity(sample.time, v, sample.high, sample.low);
  }

  /** Updates an indicator given the next quote sample. */
  public updateQuote(sample: Quote): IndicatorOutput {
    const v = this.quoteComponentFunc(sample);

    return this.updateEntity(sample.time, v, sample.askPrice, sample.bidPrice);
  }

  /** Updates an indicator given the next trade sample. */
  public updateTrade(sample: Trade): IndicatorOutput {
    const v = this.tradeComponentFunc(sample);

    return this.updateEntity(sample.time, v, v, v);
  }

  /** Updates the value of the indicator given the next sample with high and low values. */
  public update(sample: number, sampleHigh: number, sampleLow: number): number {
    if (Number.isNaN(sample) || Number.isNaN(sampleHigh) || Number.isNaN(sampleLow)) {
      return Number.NaN;
    }

    if (this.primed) {
      for (let i = 0; i < this.lengthMinOne; i++) {
        const j = i + 1;
        this.windowHigh[i] = this.windowHigh[j];
        this.windowLow[i] = this.windowLow[j];
      }

      this.windowHigh[this.lengthMinOne] = sampleHigh;
      this.windowLow[this.lengthMinOne] = sampleLow;

      this.fractalDimension = this.estimateFractalDimension();
      const a = this.estimateAlpha();
      this.value += (sample - this.value) * a;

      return this.value;
    }

    this.windowHigh[this.windowCount] = sampleHigh;
    this.windowLow[this.windowCount] = sampleLow;

    this.windowCount++;
    if (this.windowCount === this.lengthMinOne) {
      this.value = sample;
    } else if (this.windowCount === this.length) {
      this.fractalDimension = this.estimateFractalDimension();
      const a = this.estimateAlpha();
      this.value += (sample - this.value) * a;
      this.primed = true;

      return this.value;
    }

    return Number.NaN;
  }

  private updateEntity(time: Date, sample: number, sampleHigh: number, sampleLow: number): IndicatorOutput {
    const frama = this.update(sample, sampleHigh, sampleLow);

    let fdim = this.fractalDimension;
    if (Number.isNaN(frama)) {
      fdim = Number.NaN;
    }

    const scalarFrama = new Scalar();
    scalarFrama.time = time;
    scalarFrama.value = frama;

    const scalarFdim = new Scalar();
    scalarFdim.time = time;
    scalarFdim.value = fdim;

    return [scalarFrama, scalarFdim];
  }

  private estimateFractalDimension(): number {
    let minLowHalf = Number.MAX_VALUE;
    let maxHighHalf = -Number.MAX_VALUE;

    for (let i = 0; i < this.halfLength; i++) {
      const l = this.windowLow[i];
      if (minLowHalf > l) {
        minLowHalf = l;
      }

      const h = this.windowHigh[i];
      if (maxHighHalf < h) {
        maxHighHalf = h;
      }
    }

    const rangeN1 = maxHighHalf - minLowHalf;
    let minLowFull = minLowHalf;
    let maxHighFull = maxHighHalf;
    minLowHalf = Number.MAX_VALUE;
    maxHighHalf = -Number.MAX_VALUE;

    for (let i = this.halfLength; i < this.length; i++) {
      const l = this.windowLow[i];
      if (minLowHalf > l) {
        minLowHalf = l;
      }
      if (minLowFull > l) {
        minLowFull = l;
      }

      const h = this.windowHigh[i];
      if (maxHighHalf < h) {
        maxHighHalf = h;
      }
      if (maxHighFull < h) {
        maxHighFull = h;
      }
    }

    const rangeN2 = maxHighHalf - minLowHalf;
    const rangeN3 = maxHighFull - minLowFull;

    const fdim = (Math.log((rangeN1 + rangeN2) / this.halfLength) -
      Math.log(rangeN3 / this.length)) * Math.LOG2E;

    return Math.min(Math.max(fdim, 1), 2);
  }

  private estimateAlpha(): number {
    const factor = this.scalingFactor;

    // We use the fractal dimension to dynamically change the alpha of an exponential moving average.
    // The fractal dimension varies over the range from 1 to 2.
    // Since the prices are log-normal, it seems reasonable to use an exponential function to relate
    // the fractal dimension to alpha.

    // An empirically chosen scaling in Ehlers's method to map fractal dimension (1-2)
    // to the exponential a.
    const alpha = Math.exp(factor * (this.fractalDimension - 1));

    // When the fractal dimension is 1, the exponent is zero - which means that alpha is 1, and
    // the output of the exponential moving average is equal to the input.

    // Limit alpha to vary only from as to 1.
    return Math.min(Math.max(alpha, this.alphaSlowest), 1);
  }
}
