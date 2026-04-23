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
import { RelativeStrengthIndex } from '../../welles-wilder/relative-strength-index/relative-strength-index';
import { StochasticRelativeStrengthIndexParams, MovingAverageType } from './params';

/** Interface for an indicator that accepts a scalar and returns a value. */
interface LineUpdater {
  update(sample: number): number;
  isPrimed(): boolean;
}

/** Passthrough smoother for FastD period of 1. */
class Passthrough implements LineUpdater {
  update(v: number): number { return v; }
  isPrimed(): boolean { return true; }
}

/** Function to calculate mnemonic of a __StochasticRelativeStrengthIndex__ indicator. */
export const stochasticRelativeStrengthIndexMnemonic = (params: StochasticRelativeStrengthIndexParams): string => {
  const cm = componentTripleMnemonic(
    params.barComponent,
    params.quoteComponent,
    params.tradeComponent,
  );

  const maLabel = params.movingAverageType === MovingAverageType.EMA ? 'EMA' : 'SMA';

  return `stochrsi(${params.length}/${params.fastKLength}/${maLabel}${params.fastDLength}${cm})`;
};

/**
 * StochasticRelativeStrengthIndex is Tushar Chande's Stochastic RSI.
 *
 * Stochastic RSI applies the Stochastic oscillator formula to RSI values
 * instead of price data. It oscillates between 0 and 100.
 *
 * The indicator first computes RSI, then applies a stochastic calculation
 * over a rolling window of RSI values to produce Fast-K. Fast-D is a
 * moving average of Fast-K.
 *
 * Reference:
 *
 * Chande, Tushar S. and Kroll, Stanley (1993). "Stochastic RSI and Dynamic
 * Momentum Index". Stock & Commodities V.11:5 (189-199).
 */
export class StochasticRelativeStrengthIndex implements Indicator {

  private readonly rsi: RelativeStrengthIndex;

  private readonly rsiBuf: Float64Array;
  private rsiBufferIndex = 0;
  private rsiCount = 0;

  private readonly fastKLength_: number;
  private readonly fastDMA: LineUpdater;

  private fastK_ = NaN;
  private fastD_ = NaN;
  private primed_ = false;

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;
  private readonly mnemonic_: string;
  private readonly description_: string;

  constructor(params: StochasticRelativeStrengthIndexParams) {
    const length = Math.floor(params.length);
    const fastKLength = Math.floor(params.fastKLength);
    const fastDLength = Math.floor(params.fastDLength);

    if (length < 2) {
      throw new Error('length should be greater than 1');
    }

    if (fastKLength < 1) {
      throw new Error('fast K length should be greater than 0');
    }

    if (fastDLength < 1) {
      throw new Error('fast D length should be greater than 0');
    }

    this.rsi = new RelativeStrengthIndex({ length });
    this.rsiBuf = new Float64Array(fastKLength);
    this.fastKLength_ = fastKLength;

    // Create Fast-D smoother.
    if (fastDLength < 2) {
      this.fastDMA = new Passthrough();
    } else if (params.movingAverageType === MovingAverageType.EMA) {
      this.fastDMA = new ExponentialMovingAverage({
        length: fastDLength,
        firstIsAverage: params.firstIsAverage ?? false,
      });
    } else {
      this.fastDMA = new SimpleMovingAverage({ length: fastDLength });
    }

    this.mnemonic_ = stochasticRelativeStrengthIndexMnemonic(params);
    this.description_ = 'Stochastic Relative Strength Index ' + this.mnemonic_;

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
      IndicatorIdentifier.StochasticRelativeStrengthIndex,
      this.mnemonic_,
      this.description_,
      [
        { mnemonic: this.mnemonic_ + ' fastK', description: this.description_ + ' Fast-K' },
        { mnemonic: this.mnemonic_ + ' fastD', description: this.description_ + ' Fast-D' },
      ],
    );
  }

  /** Updates the indicator given the next sample and returns both FastK and FastD values. */
  public update(sample: number): [number, number] {
    if (isNaN(sample)) {
      return [NaN, NaN];
    }

    // Feed to internal RSI.
    const rsiValue = this.rsi.update(sample);
    if (isNaN(rsiValue)) {
      return [this.fastK_, this.fastD_];
    }

    // Store RSI value in circular buffer.
    this.rsiBuf[this.rsiBufferIndex] = rsiValue;
    this.rsiBufferIndex = (this.rsiBufferIndex + 1) % this.fastKLength_;
    this.rsiCount++;

    // Need at least fastKLength RSI values for stochastic calculation.
    if (this.rsiCount < this.fastKLength_) {
      return [this.fastK_, this.fastD_];
    }

    // Find min and max of RSI values in the window.
    let minRSI = this.rsiBuf[0];
    let maxRSI = this.rsiBuf[0];

    for (let i = 1; i < this.fastKLength_; i++) {
      if (this.rsiBuf[i] < minRSI) {
        minRSI = this.rsiBuf[i];
      }

      if (this.rsiBuf[i] > maxRSI) {
        maxRSI = this.rsiBuf[i];
      }
    }

    // Calculate Fast-K.
    const diff = maxRSI - minRSI;
    if (diff > 0) {
      this.fastK_ = 100 * (rsiValue - minRSI) / diff;
    } else {
      this.fastK_ = 0;
    }

    // Feed Fast-K to Fast-D smoother.
    this.fastD_ = this.fastDMA.update(this.fastK_);

    if (!this.primed_ && this.fastDMA.isPrimed()) {
      this.primed_ = true;
    }

    return [this.fastK_, this.fastD_];
  }

  /** Updates the indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    const [fastK, fastD] = this.update(sample.value);
    const s1 = new Scalar();
    s1.time = sample.time;
    s1.value = fastK;
    const s2 = new Scalar();
    s2.time = sample.time;
    s2.value = fastD;
    return [s1, s2];
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
