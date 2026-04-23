import { buildMetadata } from '../../core/build-metadata';
import { Bar } from '../../../entities/bar';
import { BarComponent, DefaultBarComponent, barComponentValue } from '../../../entities/bar-component';
import { Quote } from '../../../entities/quote';
import { QuoteComponent, DefaultQuoteComponent, quoteComponentValue } from '../../../entities/quote-component';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { TradeComponent, DefaultTradeComponent, tradeComponentValue } from '../../../entities/trade-component';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { Indicator } from '../../core/indicator';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Band } from '../../core/outputs/band';
import { ExponentialMovingAverage } from '../../common/exponential-moving-average/exponential-moving-average';
import { SimpleMovingAverage } from '../../common/simple-moving-average/simple-moving-average';
import { Variance } from '../../common/variance/variance';
import { BollingerBandsParams, MovingAverageType } from './params';

/** Interface for an indicator that accepts a scalar and returns a value. */
interface LineUpdater {
  update(sample: number): number;
  isPrimed(): boolean;
}

/** Function to calculate mnemonic of a __BollingerBands__ indicator. */
export const bollingerBandsMnemonic = (length: number, upperMultiplier: number, lowerMultiplier: number,
  barComponent?: BarComponent, quoteComponent?: QuoteComponent, tradeComponent?: TradeComponent): string => {
  const cm = componentTripleMnemonic(barComponent, quoteComponent, tradeComponent);
  return `bb(${length},${Math.floor(upperMultiplier)},${Math.floor(lowerMultiplier)}${cm})`;
};

/**
 * John Bollinger's Bollinger Bands indicator.
 *
 * Bollinger Bands consist of a middle band (moving average) and upper/lower bands
 * placed a specified number of standard deviations above and below the middle band.
 *
 * The indicator produces six outputs:
 *   - LowerValue: middleValue - lowerMultiplier * stddev
 *   - MiddleValue: moving average of the input
 *   - UpperValue: middleValue + upperMultiplier * stddev
 *   - BandWidth: (upperValue - lowerValue) / middleValue
 *   - PercentBand: (sample - lowerValue) / (upperValue - lowerValue)
 *   - Band: lower/upper band pair
 *
 * Reference:
 *
 * Bollinger, John (2002). Bollinger on Bollinger Bands. McGraw-Hill.
 */
export class BollingerBands implements Indicator {

  private readonly ma: LineUpdater;
  private readonly variance_: Variance;
  private readonly upperMultiplier_: number;
  private readonly lowerMultiplier_: number;

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  private middleValue_ = NaN;
  private upperValue_ = NaN;
  private lowerValue_ = NaN;
  private bandWidth_ = NaN;
  private percentBand_ = NaN;
  private primed_ = false;

  private readonly mnemonic_: string;
  private readonly description_: string;

  constructor(params: BollingerBandsParams) {
    const length = Math.floor(params.length);

    if (length < 2) {
      throw new Error('length should be greater than 1');
    }

    const upperMultiplier = params.upperMultiplier ?? 2.0;
    const lowerMultiplier = params.lowerMultiplier ?? 2.0;
    const isUnbiased = params.isUnbiased ?? true;

    this.upperMultiplier_ = upperMultiplier;
    this.lowerMultiplier_ = lowerMultiplier;

    const bc = params.barComponent ?? DefaultBarComponent;
    const qc = params.quoteComponent ?? DefaultQuoteComponent;
    const tc = params.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    // Create variance sub-indicator.
    this.variance_ = new Variance({
      length,
      unbiased: isUnbiased,
      barComponent: params.barComponent,
      quoteComponent: params.quoteComponent,
      tradeComponent: params.tradeComponent,
    });

    // Create moving average sub-indicator.
    if (params.movingAverageType === MovingAverageType.EMA) {
      this.ma = new ExponentialMovingAverage({
        length,
        firstIsAverage: params.firstIsAverage ?? false,
      });
    } else {
      this.ma = new SimpleMovingAverage({ length });
    }

    this.mnemonic_ = bollingerBandsMnemonic(length, upperMultiplier, lowerMultiplier,
      params.barComponent, params.quoteComponent, params.tradeComponent);
    this.description_ = 'Bollinger Bands ' + this.mnemonic_;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.primed_;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.BollingerBands,
      this.mnemonic_,
      this.description_,
      [
        { mnemonic: this.mnemonic_ + ' lower', description: this.description_ + ' Lower' },
        { mnemonic: this.mnemonic_ + ' middle', description: this.description_ + ' Middle' },
        { mnemonic: this.mnemonic_ + ' upper', description: this.description_ + ' Upper' },
        { mnemonic: this.mnemonic_ + ' bandWidth', description: this.description_ + ' Band Width' },
        { mnemonic: this.mnemonic_ + ' percentBand', description: this.description_ + ' Percent Band' },
        { mnemonic: this.mnemonic_ + ' band', description: this.description_ + ' Band' },
      ],
    );
  }

  /**
   * Updates the indicator given the next sample value.
   * Returns [lower, middle, upper, bandWidth, percentBand].
   */
  public update(sample: number): [number, number, number, number, number] {
    if (isNaN(sample)) {
      return [NaN, NaN, NaN, NaN, NaN];
    }

    const middle = this.ma.update(sample);
    const v = this.variance_.update(sample);

    this.primed_ = this.ma.isPrimed() && this.variance_.isPrimed();

    if (isNaN(middle) || isNaN(v)) {
      this.middleValue_ = NaN;
      this.upperValue_ = NaN;
      this.lowerValue_ = NaN;
      this.bandWidth_ = NaN;
      this.percentBand_ = NaN;
      return [NaN, NaN, NaN, NaN, NaN];
    }

    const stddev = Math.sqrt(v);
    const upper = middle + this.upperMultiplier_ * stddev;
    const lower = middle - this.lowerMultiplier_ * stddev;

    const epsilon = 1e-10;

    let bw: number;
    if (Math.abs(middle) < epsilon) {
      bw = 0;
    } else {
      bw = (upper - lower) / middle;
    }

    let pctB: number;
    const spread = upper - lower;
    if (Math.abs(spread) < epsilon) {
      pctB = 0;
    } else {
      pctB = (sample - lower) / spread;
    }

    this.middleValue_ = middle;
    this.upperValue_ = upper;
    this.lowerValue_ = lower;
    this.bandWidth_ = bw;
    this.percentBand_ = pctB;

    return [lower, middle, upper, bw, pctB];
  }

  /** Updates the indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    const [lower, middle, upper, bw, pctB] = this.update(sample.value);

    const s0 = new Scalar(); s0.time = sample.time; s0.value = lower;
    const s1 = new Scalar(); s1.time = sample.time; s1.value = middle;
    const s2 = new Scalar(); s2.time = sample.time; s2.value = upper;
    const s3 = new Scalar(); s3.time = sample.time; s3.value = bw;
    const s4 = new Scalar(); s4.time = sample.time; s4.value = pctB;

    const band = new Band();
    band.time = sample.time;
    if (isNaN(lower) || isNaN(upper)) {
      band.lower = NaN;
      band.upper = NaN;
    } else {
      band.lower = lower;
      band.upper = upper;
    }

    return [s0, s1, s2, s3, s4, band];
  }

  /** Updates the indicator given the next bar sample. */
  public updateBar(sample: Bar): IndicatorOutput {
    const v = this.barComponentFunc(sample);
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = v;
    return this.updateScalar(scalar);
  }

  /** Updates the indicator given the next quote sample. */
  public updateQuote(sample: Quote): IndicatorOutput {
    const v = this.quoteComponentFunc(sample);
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = v;
    return this.updateScalar(scalar);
  }

  /** Updates the indicator given the next trade sample. */
  public updateTrade(sample: Trade): IndicatorOutput {
    const v = this.tradeComponentFunc(sample);
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = v;
    return this.updateScalar(scalar);
  }
}
