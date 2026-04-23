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
import { TripleExponentialMovingAverageOscillatorParams } from './params';

/** Function to calculate mnemonic of a __TripleExponentialMovingAverageOscillator__ indicator. */
export const tripleExponentialMovingAverageOscillatorMnemonic = (params: TripleExponentialMovingAverageOscillatorParams): string => {
  const cm = componentTripleMnemonic(
    params.barComponent,
    params.quoteComponent,
    params.tradeComponent,
  );

  return `trix(${params.length}${cm})`;
};

/**
 * TripleExponentialMovingAverageOscillator is Jack Hutson's Triple Exponential Moving
 * Average Oscillator (TRIX).
 *
 * TRIX is a 1-day rate-of-change of a triple-smoothed exponential moving average. It applies
 * EMA three times in series (all with the same period and SMA-seeded), then computes:
 *
 *   TRIX = ((EMA3[i] - EMA3[i-1]) / EMA3[i-1]) * 100
 *
 * The indicator oscillates around zero. Positive values indicate upward momentum, negative
 * values indicate downward momentum.
 *
 * Reference:
 *
 * Hutson, Jack K. (1983). "Good TRIX". Technical Analysis of Stocks and Commodities.
 */
export class TripleExponentialMovingAverageOscillator implements Indicator {

  private readonly ema1: ExponentialMovingAverage;
  private readonly ema2: ExponentialMovingAverage;
  private readonly ema3: ExponentialMovingAverage;
  private previousEMA3 = NaN;
  private hasPreviousEMA = false;
  private primed_ = false;

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;
  private readonly mnemonic_: string;
  private readonly description_: string;

  constructor(params: TripleExponentialMovingAverageOscillatorParams) {
    const length = Math.floor(params.length);

    if (length < 1) {
      throw new Error('length should be positive');
    }

    this.ema1 = new ExponentialMovingAverage({ length, firstIsAverage: true });
    this.ema2 = new ExponentialMovingAverage({ length, firstIsAverage: true });
    this.ema3 = new ExponentialMovingAverage({ length, firstIsAverage: true });

    this.mnemonic_ = tripleExponentialMovingAverageOscillatorMnemonic(params);
    this.description_ = 'Triple exponential moving average oscillator ' + this.mnemonic_;

    const bc = params.barComponent ?? DefaultBarComponent;
    const qc = params.quoteComponent ?? DefaultQuoteComponent;
    const tc = params.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.primed_;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.TripleExponentialMovingAverageOscillator,
      this.mnemonic_,
      this.description_,
      [
        { mnemonic: this.mnemonic_, description: this.description_ },
      ],
    );
  }

  /** Updates the indicator given the next sample. */
  public update(sample: number): number {
    if (isNaN(sample)) {
      return sample;
    }

    const v1 = this.ema1.update(sample);
    if (isNaN(v1)) {
      return NaN;
    }

    const v2 = this.ema2.update(v1);
    if (isNaN(v2)) {
      return NaN;
    }

    const v3 = this.ema3.update(v2);
    if (isNaN(v3)) {
      return NaN;
    }

    if (!this.hasPreviousEMA) {
      this.previousEMA3 = v3;
      this.hasPreviousEMA = true;
      return NaN;
    }

    const result = ((v3 - this.previousEMA3) / this.previousEMA3) * 100;
    this.previousEMA3 = v3;

    if (!this.primed_) {
      this.primed_ = true;
    }

    return result;
  }

  /** Updates the indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    const value = this.update(sample.value);
    const s = new Scalar();
    s.time = sample.time;
    s.value = value;
    return [s];
  }

  /** Updates the indicator given the next bar sample. */
  public updateBar(sample: Bar): IndicatorOutput {
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = this.barComponentFunc(sample);
    return this.updateScalar(scalar);
  }

  /** Updates the indicator given the next quote sample. */
  public updateQuote(sample: Quote): IndicatorOutput {
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = this.quoteComponentFunc(sample);
    return this.updateScalar(scalar);
  }

  /** Updates the indicator given the next trade sample. */
  public updateTrade(sample: Trade): IndicatorOutput {
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = this.tradeComponentFunc(sample);
    return this.updateScalar(scalar);
  }
}
