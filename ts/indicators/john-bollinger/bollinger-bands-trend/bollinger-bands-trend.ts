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
import { ExponentialMovingAverage } from '../../common/exponential-moving-average/exponential-moving-average';
import { SimpleMovingAverage } from '../../common/simple-moving-average/simple-moving-average';
import { Variance } from '../../common/variance/variance';
import { BollingerBandsTrendParams, MovingAverageType } from './params';

/** Interface for an indicator that accepts a scalar and returns a value. */
interface LineUpdater {
  update(sample: number): number;
  isPrimed(): boolean;
}

/** Holds the sub-components for one Bollinger Band calculation. */
class BBLine {
  private readonly ma: LineUpdater;
  private readonly variance_: Variance;
  private readonly upperMultiplier_: number;
  private readonly lowerMultiplier_: number;
  private primed_ = false;

  constructor(
    length: number,
    upperMultiplier: number,
    lowerMultiplier: number,
    isUnbiased: boolean,
    maType: MovingAverageType | undefined,
    firstIsAverage: boolean,
    barComponent?: BarComponent,
    quoteComponent?: QuoteComponent,
    tradeComponent?: TradeComponent,
  ) {
    this.upperMultiplier_ = upperMultiplier;
    this.lowerMultiplier_ = lowerMultiplier;

    this.variance_ = new Variance({
      length,
      unbiased: isUnbiased,
      barComponent,
      quoteComponent,
      tradeComponent,
    });

    if (maType === MovingAverageType.EMA) {
      this.ma = new ExponentialMovingAverage({ length, firstIsAverage });
    } else {
      this.ma = new SimpleMovingAverage({ length });
    }
  }

  /** Returns [lower, middle, upper, primed]. */
  update(sample: number): [number, number, number, boolean] {
    const middle = this.ma.update(sample);
    const v = this.variance_.update(sample);

    this.primed_ = this.ma.isPrimed() && this.variance_.isPrimed();

    if (isNaN(middle) || isNaN(v)) {
      return [NaN, NaN, NaN, this.primed_];
    }

    const stddev = Math.sqrt(v);
    const upper = middle + this.upperMultiplier_ * stddev;
    const lower = middle - this.lowerMultiplier_ * stddev;

    return [lower, middle, upper, this.primed_];
  }
}

/**
 * John Bollinger's Bollinger Bands Trend indicator.
 *
 * BBTrend measures the difference between the widths of fast and slow Bollinger Bands
 * relative to the fast middle band, indicating trend strength and direction.
 *
 * The indicator produces a single output:
 *
 *   bbtrend = (|fastLower - slowLower| - |fastUpper - slowUpper|) / fastMiddle
 *
 * Reference:
 *
 * Bollinger, John (2002). Bollinger on Bollinger Bands. McGraw-Hill.
 */
export class BollingerBandsTrend implements Indicator {

  private readonly fastBB: BBLine;
  private readonly slowBB: BBLine;

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  private value_ = NaN;
  private primed_ = false;

  private readonly mnemonic_: string;
  private readonly description_: string;

  constructor(params: BollingerBandsTrendParams) {
    const fastLength = Math.floor(params.fastLength);
    const slowLength = Math.floor(params.slowLength);

    if (fastLength < 2) {
      throw new Error('fast length should be greater than 1');
    }

    if (slowLength < 2) {
      throw new Error('slow length should be greater than 1');
    }

    if (slowLength <= fastLength) {
      throw new Error('slow length should be greater than fast length');
    }

    const upperMultiplier = params.upperMultiplier ?? 2.0;
    const lowerMultiplier = params.lowerMultiplier ?? 2.0;
    const isUnbiased = params.isUnbiased ?? true;
    const firstIsAverage = params.firstIsAverage ?? false;

    const bc = params.barComponent ?? DefaultBarComponent;
    const qc = params.quoteComponent ?? DefaultQuoteComponent;
    const tc = params.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    this.fastBB = new BBLine(
      fastLength, upperMultiplier, lowerMultiplier, isUnbiased,
      params.movingAverageType, firstIsAverage,
      params.barComponent, params.quoteComponent, params.tradeComponent,
    );

    this.slowBB = new BBLine(
      slowLength, upperMultiplier, lowerMultiplier, isUnbiased,
      params.movingAverageType, firstIsAverage,
      params.barComponent, params.quoteComponent, params.tradeComponent,
    );

    const cm = componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent);
    this.mnemonic_ = `bbtrend(${fastLength},${slowLength},${Math.floor(upperMultiplier)},${Math.floor(lowerMultiplier)}${cm})`;
    this.description_ = 'Bollinger Bands Trend ' + this.mnemonic_;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.primed_;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.BollingerBandsTrend,
      this.mnemonic_,
      this.description_,
      [
        { mnemonic: this.mnemonic_, description: this.description_ },
      ],
    );
  }

  /** Updates the indicator given the next sample value and returns the BBTrend value. */
  public update(sample: number): number {
    if (isNaN(sample)) {
      return NaN;
    }

    const [fastLower, fastMiddle, fastUpper, fastPrimed] = this.fastBB.update(sample);
    const [slowLower, , slowUpper, slowPrimed] = this.slowBB.update(sample);

    this.primed_ = fastPrimed && slowPrimed;

    if (!this.primed_ || isNaN(fastMiddle) || isNaN(fastLower) || isNaN(slowLower)) {
      this.value_ = NaN;
      return NaN;
    }

    const epsilon = 1e-10;
    const lowerDiff = Math.abs(fastLower - slowLower);
    const upperDiff = Math.abs(fastUpper - slowUpper);

    if (Math.abs(fastMiddle) < epsilon) {
      this.value_ = 0;
      return 0;
    }

    const result = (lowerDiff - upperDiff) / fastMiddle;
    this.value_ = result;

    return result;
  }

  /** Updates the indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    const v = this.update(sample.value);
    const s = new Scalar();
    s.time = sample.time;
    s.value = v;
    return [s];
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
