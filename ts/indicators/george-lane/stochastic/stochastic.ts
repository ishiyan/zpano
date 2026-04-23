import { buildMetadata } from '../../core/build-metadata';
import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { Indicator } from '../../core/indicator';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { ExponentialMovingAverage } from '../../common/exponential-moving-average/exponential-moving-average';
import { SimpleMovingAverage } from '../../common/simple-moving-average/simple-moving-average';
import { StochasticParams, MovingAverageType } from './params';

/** Interface for an indicator that accepts a scalar and returns a value. */
interface LineUpdater {
  update(sample: number): number;
  isPrimed(): boolean;
}

/** Passthrough smoother for period of 1. */
class Passthrough implements LineUpdater {
  update(v: number): number { return v; }
  isPrimed(): boolean { return true; }
}

/** Creates a moving average line updater. */
function createMA(maType: MovingAverageType | undefined, length: number, firstIsAverage: boolean): [LineUpdater, string] {
  if (length < 2) {
    return [new Passthrough(), 'SMA'];
  }

  if (maType === MovingAverageType.EMA) {
    return [new ExponentialMovingAverage({ length, firstIsAverage }), 'EMA'];
  }

  return [new SimpleMovingAverage({ length }), 'SMA'];
}

/** Function to calculate mnemonic of a __Stochastic__ indicator. */
export const stochasticMnemonic = (params: StochasticParams): string => {
  const slowKLabel = params.slowKMAType === MovingAverageType.EMA ? 'EMA' : 'SMA';
  const slowDLabel = params.slowDMAType === MovingAverageType.EMA ? 'EMA' : 'SMA';
  return `stoch(${params.fastKLength}/${slowKLabel}${params.slowKLength}/${slowDLabel}${params.slowDLength})`;
};

/**
 * George Lane's Stochastic Oscillator.
 *
 * The Stochastic Oscillator measures the position of the close relative to the
 * high-low range over a lookback period. It produces three outputs:
 *   - Fast-K: the raw stochastic value
 *   - Slow-K: a moving average of Fast-K (also known as Fast-D)
 *   - Slow-D: a moving average of Slow-K
 *
 * The indicator requires bar data (high, low, close). For scalar, quote, and
 * trade updates, the single value substitutes for all three.
 *
 * Reference:
 *
 * Lane, George C. (1984). "Lane's Stochastics". Technical Analysis of Stocks & Commodities.
 */
export class Stochastic implements Indicator {

  private readonly fastKLength_: number;

  private readonly highBuf: Float64Array;
  private readonly lowBuf: Float64Array;
  private bufferIndex = 0;
  private count = 0;

  private readonly slowKMA: LineUpdater;
  private readonly slowDMA: LineUpdater;

  private fastK_ = NaN;
  private slowK_ = NaN;
  private slowD_ = NaN;
  private primed_ = false;

  private readonly mnemonic_: string;
  private readonly description_: string;

  constructor(params: StochasticParams) {
    const fastKLength = Math.floor(params.fastKLength);
    const slowKLength = Math.floor(params.slowKLength);
    const slowDLength = Math.floor(params.slowDLength);

    if (fastKLength < 1) {
      throw new Error('fast K length should be greater than 0');
    }

    if (slowKLength < 1) {
      throw new Error('slow K length should be greater than 0');
    }

    if (slowDLength < 1) {
      throw new Error('slow D length should be greater than 0');
    }

    this.fastKLength_ = fastKLength;
    this.highBuf = new Float64Array(fastKLength);
    this.lowBuf = new Float64Array(fastKLength);

    const fia = params.firstIsAverage ?? false;
    [this.slowKMA] = createMA(params.slowKMAType, slowKLength, fia);
    [this.slowDMA] = createMA(params.slowDMAType, slowDLength, fia);

    this.mnemonic_ = stochasticMnemonic(params);
    this.description_ = 'Stochastic Oscillator ' + this.mnemonic_;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.primed_;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.Stochastic,
      this.mnemonic_,
      this.description_,
      [
        { mnemonic: this.mnemonic_ + ' fastK', description: this.description_ + ' Fast-K' },
        { mnemonic: this.mnemonic_ + ' slowK', description: this.description_ + ' Slow-K' },
        { mnemonic: this.mnemonic_ + ' slowD', description: this.description_ + ' Slow-D' },
      ],
    );
  }

  /** Updates the indicator given the next bar's close, high, and low values. Returns [FastK, SlowK, SlowD]. */
  public update(close: number, high: number, low: number): [number, number, number] {
    if (isNaN(close) || isNaN(high) || isNaN(low)) {
      return [NaN, NaN, NaN];
    }

    // Store high and low in circular buffer.
    this.highBuf[this.bufferIndex] = high;
    this.lowBuf[this.bufferIndex] = low;
    this.bufferIndex = (this.bufferIndex + 1) % this.fastKLength_;
    this.count++;

    // Need at least fastKLength bars.
    if (this.count < this.fastKLength_) {
      return [this.fastK_, this.slowK_, this.slowD_];
    }

    // Find highest high and lowest low in the window.
    let hh = this.highBuf[0];
    let ll = this.lowBuf[0];

    for (let i = 1; i < this.fastKLength_; i++) {
      if (this.highBuf[i] > hh) {
        hh = this.highBuf[i];
      }

      if (this.lowBuf[i] < ll) {
        ll = this.lowBuf[i];
      }
    }

    // Calculate Fast-K.
    const diff = hh - ll;
    if (diff > 0) {
      this.fastK_ = 100 * (close - ll) / diff;
    } else {
      this.fastK_ = 0;
    }

    // Feed Fast-K to Slow-K smoother.
    this.slowK_ = this.slowKMA.update(this.fastK_);

    // Feed Slow-K to Slow-D smoother (only when Slow-K MA is primed).
    if (this.slowKMA.isPrimed()) {
      this.slowD_ = this.slowDMA.update(this.slowK_);

      if (!this.primed_ && this.slowDMA.isPrimed()) {
        this.primed_ = true;
      }
    }

    return [this.fastK_, this.slowK_, this.slowD_];
  }

  /** Updates the indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    const v = sample.value;
    const [fastK, slowK, slowD] = this.update(v, v, v);
    const s1 = new Scalar();
    s1.time = sample.time;
    s1.value = fastK;
    const s2 = new Scalar();
    s2.time = sample.time;
    s2.value = slowK;
    const s3 = new Scalar();
    s3.time = sample.time;
    s3.value = slowD;
    return [s1, s2, s3];
  }

  /** Updates the indicator given the next bar sample. */
  public updateBar(sample: Bar): IndicatorOutput {
    const [fastK, slowK, slowD] = this.update(sample.close, sample.high, sample.low);
    const s1 = new Scalar();
    s1.time = sample.time;
    s1.value = fastK;
    const s2 = new Scalar();
    s2.time = sample.time;
    s2.value = slowK;
    const s3 = new Scalar();
    s3.time = sample.time;
    s3.value = slowD;
    return [s1, s2, s3];
  }

  /** Updates the indicator given the next quote sample. */
  public updateQuote(sample: Quote): IndicatorOutput {
    const v = (sample.bid + sample.ask) / 2;
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = v;
    return this.updateScalar(scalar);
  }

  /** Updates the indicator given the next trade sample. */
  public updateTrade(sample: Trade): IndicatorOutput {
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = sample.price;
    return this.updateScalar(scalar);
  }
}
