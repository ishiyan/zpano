import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorType } from '../../core/indicator-type';
import { LineIndicator } from '../../core/line-indicator';
import { OutputType } from '../../core/outputs/output-type';
import { TripleExponentialMovingAverageOutput } from './triple-exponential-moving-average-output';
import { TripleExponentialMovingAverageLengthParams } from './triple-exponential-moving-average-params';
import { TripleExponentialMovingAverageSmoothingFactorParams } from './triple-exponential-moving-average-params';

const guardLength = (object: any): object is TripleExponentialMovingAverageLengthParams => 'length' in object;

/** Function to calculate mnemonic of a __TripleExponentialMovingAverage__ indicator. */
export const tripleExponentialMovingAverageMnemonic =
  (params: TripleExponentialMovingAverageLengthParams | TripleExponentialMovingAverageSmoothingFactorParams): string => {
  if (guardLength(params)) {
    const p = params as TripleExponentialMovingAverageLengthParams;
    return 'tema('.concat(Math.floor(p.length).toString(),
      componentTripleMnemonic(p.barComponent, p.quoteComponent, p.tradeComponent), ')');
  } else {
    const p = params as TripleExponentialMovingAverageSmoothingFactorParams;
    const length = Math.round(2 / p.smoothingFactor) - 1;
    return 'tema('.concat(length.toString(), ', ', p.smoothingFactor.toFixed(8),
      componentTripleMnemonic(p.barComponent, p.quoteComponent, p.tradeComponent), ')');
  }
};

// https://store.traders.com/-v12-c01-smoothi-pdf.html
// https://store.traders.com/-v12-c02-smoothi-pdf.html

/** __Triple Exponential Moving Average__ line indicator computes the triple exponential, or triple exponentially weighted, moving average (_TEMA_).
 *
 * The TEMA was developed by Patrick G. Mulloy and is described in two articles:
 *
 * ❶ Technical Analysis of Stocks &amp; Commodities v.12:1 (11-19), Smoothing Data With Faster Moving Averages.
 *
 * ❷ Technical Analysis of Stocks &amp; Commodities v.12:2 (72-80), Smoothing Data With Less Lag.
 *
 * The calculation is as follows:
 *
 * EMA¹ᵢ = EMA(Pᵢ) = αPᵢ + (1-α)EMA¹ᵢ₋₁ = EMA¹ᵢ₋₁ + α(Pᵢ - EMA¹ᵢ₋₁), 0 < α ≤ 1
 *
 * EMA²ᵢ = EMA(EMA¹ᵢ) = αEMA¹ᵢ + (1-α)EMA²ᵢ₋₁ = EMA²ᵢ₋₁ + α(EMA¹ᵢ - EMA²ᵢ₋₁), 0 < α ≤ 1
 *
 * EMA³ᵢ = EMA(EMA²ᵢ) = αEMA²ᵢ + (1-α)EMA³ᵢ₋₁ = EMA³ᵢ₋₁ + α(EMA²ᵢ - EMA³ᵢ₋₁), 0 < α ≤ 1
 *
 * TEMAᵢ = 3(EMA¹ᵢ - EMA²ᵢ) + EMA³ᵢ
 *
 * The very first EMA value (the seed for subsequent values) is calculated differently.
 * This implementation allows for two algorithms for this seed.
 *
 * ❶ Use a simple average of the first 'period'. This is the most widely documented approach.
 *
 * ❷ Use first sample value as a seed. This is used in Metastock.
 */
export class TripleExponentialMovingAverage extends LineIndicator {
  private readonly smoothingFactor: number;
  private readonly firstIsAverage: boolean;
  private readonly length: number;
  private readonly length2: number;
  private readonly length3: number;
  private sum = 0;
  private count = 0;
  private ema1 = 0;
  private ema2 = 0;
  private ema3 = 0;

  /**
   * Constructs an instance given a length in samples or a smoothing factor in (0, 1).
   **/
  public constructor(params: TripleExponentialMovingAverageLengthParams | TripleExponentialMovingAverageSmoothingFactorParams){
    super();
    let len;
    if (guardLength(params)) {
      const p = params as TripleExponentialMovingAverageLengthParams;
      len = Math.floor(p.length);
      if (len < 2) {
        throw new Error('length should be greater than 1');
      }

      this.length = len;
      this.smoothingFactor = 2 / (len + 1);
      this.firstIsAverage = p.firstIsAverage;

    } else {
      const p = params as TripleExponentialMovingAverageSmoothingFactorParams;
      if (p.smoothingFactor <= 0 || p.smoothingFactor >= 1) {
        throw new Error('smoothing factor should be in range (0, 1)');
      }

      this.smoothingFactor = p.smoothingFactor;
      this.length = Math.round(2 / this.smoothingFactor) - 1;
      this.firstIsAverage = false;
    }

    this.length2 = 2 * this.length - 1;
    this.length3 = 3 * this.length - 2;
    this.mnemonic = tripleExponentialMovingAverageMnemonic(params);
    this.description = 'Triple exponential moving average ' + this.mnemonic;
    this.barComponent = params.barComponent;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;
    this.primed = false;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return {
      type: IndicatorType.TripleExponentialMovingAverage,
      mnemonic: this.mnemonic,
      description: this.description,
      outputs: [{
        kind: TripleExponentialMovingAverageOutput.TripleExponentialMovingAverageValue,
        type: OutputType.Scalar,
        mnemonic: this.mnemonic,
        description: this.description,
      }],
    };
  }

  /** Updates the value of the indicator given the next sample. */
  public update(sample: number): number {
    if (Number.isNaN(sample)) {
      return sample;
    }

    if (this.primed) {
      this.ema1 += (sample - this.ema1) * this.smoothingFactor;
      this.ema2 += (this.ema1 - this.ema2) * this.smoothingFactor;
      this.ema3 += (this.ema2 - this.ema3) * this.smoothingFactor;
      return 3 * (this.ema1 - this.ema2) + this.ema3;
    }

    // Not primed.
    ++this.count;
    if (this.firstIsAverage) { // First is the simple average.
      if (this.count === 1) {
        this.sum = sample;
      } else if (this.length >= this.count) {
        this.sum += sample;
        if (this.length === this.count) {
          this.ema1 = this.sum / this.length;
          this.sum = this.ema1;
        }
      } else if (this.length2 >= this.count) {
        this.ema1 += (sample - this.ema1) * this.smoothingFactor;
        this.sum += this.ema1;
        if (this.length2 === this.count) {
          this.ema2 = this.sum / this.length;
          this.sum = this.ema2;
        }
      } else { //if (this.length3 >= this.count) {
        this.ema1 += (sample - this.ema1) * this.smoothingFactor;
        this.ema2 += (this.ema1 - this.ema2) * this.smoothingFactor;
        this.sum += this.ema2;
        if (this.length3 === this.count) {
          this.primed = true;
          this.ema3 = this.sum / this.length;
          return 3 * (this.ema1 - this.ema2) + this.ema3;
        }
      }
    } else { // firstIsAverage is false, Metastock case.
      if (this.count === 1) {
        this.ema1 = sample;
      } else if (this.length >= this.count) {
        this.ema1 += (sample - this.ema1) * this.smoothingFactor;
        if (this.length === this.count) {
          this.ema2 = this.ema1;
        }
      } else if (this.length2 >= this.count) {
        this.ema1 += (sample - this.ema1) * this.smoothingFactor;
        this.ema2 += (this.ema1 - this.ema2) * this.smoothingFactor;
        if (this.length2 === this.count) {
          this.ema3 = this.ema2;
        }
      } else { //if (this.length3 >= this.count) {
        this.ema1 += (sample - this.ema1) * this.smoothingFactor;
        this.ema2 += (this.ema1 - this.ema2) * this.smoothingFactor;
        this.ema3 += (this.ema2 - this.ema3) * this.smoothingFactor;
        if (this.length3 === this.count) {
          this.primed = true;
          return 3 * (this.ema1 - this.ema2) + this.ema3;
        }
      }
    }

    return Number.NaN;
  }
}
