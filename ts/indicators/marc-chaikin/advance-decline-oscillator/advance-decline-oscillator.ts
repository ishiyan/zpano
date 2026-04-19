import { Bar } from '../../../entities/bar';
import { Scalar } from '../../../entities/scalar';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorType } from '../../core/indicator-type';
import { LineIndicator } from '../../core/line-indicator';
import { OutputType } from '../../core/outputs/output-type';
import { ExponentialMovingAverage } from '../../common/exponential-moving-average/exponential-moving-average';
import { SimpleMovingAverage } from '../../common/simple-moving-average/simple-moving-average';
import { AdvanceDeclineOscillatorOutput } from './advance-decline-oscillator-output';
import { AdvanceDeclineOscillatorParams, MovingAverageType } from './advance-decline-oscillator-params';

/** Interface for an indicator that accepts a scalar and returns a value. */
interface LineUpdater {
  update(sample: number): number;
  isPrimed(): boolean;
}

/** Function to calculate mnemonic of an __AdvanceDeclineOscillator__ indicator. */
export const advanceDeclineOscillatorMnemonic = (params: AdvanceDeclineOscillatorParams): string => {
  const maLabel = params.movingAverageType === MovingAverageType.SMA ? 'SMA' : 'EMA';

  return `adosc(${maLabel}${params.fastLength}/${maLabel}${params.slowLength})`;
};

/**
 * AdvanceDeclineOscillator is Marc Chaikin's Advance-Decline (A/D) Oscillator.
 *
 * The Chaikin Oscillator is the difference between a fast and slow moving average
 * of the Accumulation/Distribution Line. It is used to anticipate changes in the A/D Line
 * by measuring the momentum behind accumulation/distribution movements.
 *
 * The value is calculated as:
 *
 *   CLV = ((Close - Low) - (High - Close)) / (High - Low)
 *   AD  = AD_prev + CLV × Volume
 *   ADOSC = FastMA(AD) - SlowMA(AD)
 *
 * When High equals Low, the A/D value is unchanged (no division by zero).
 *
 * Reference:
 *
 * Chaikin, Marc. "Chaikin Oscillator".
 */
export class AdvanceDeclineOscillator extends LineIndicator {
  private ad: number;
  private fastMA: LineUpdater;
  private slowMA: LineUpdater;
  private value: number;

  /**
   * Constructs an instance given the parameters.
   */
  public constructor(params: AdvanceDeclineOscillatorParams) {
    super();

    const fastLength = Math.floor(params.fastLength);
    const slowLength = Math.floor(params.slowLength);

    if (fastLength < 2) {
      throw new Error('fast length should be greater than 1');
    }

    if (slowLength < 2) {
      throw new Error('slow length should be greater than 1');
    }

    this.mnemonic = advanceDeclineOscillatorMnemonic(params);
    this.description = 'Chaikin Advance-Decline Oscillator ' + this.mnemonic;

    this.ad = 0;

    if (params.movingAverageType === MovingAverageType.SMA) {
      this.fastMA = new SimpleMovingAverage({ length: fastLength });
      this.slowMA = new SimpleMovingAverage({ length: slowLength });
    } else {
      this.fastMA = new ExponentialMovingAverage({
        length: fastLength,
        firstIsAverage: params.firstIsAverage ?? false,
      });
      this.slowMA = new ExponentialMovingAverage({
        length: slowLength,
        firstIsAverage: params.firstIsAverage ?? false,
      });
    }

    this.value = Number.NaN;
    this.primed = false;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return {
      type: IndicatorType.AdvanceDeclineOscillator,
      mnemonic: this.mnemonic,
      description: this.description,
      outputs: [{
        kind: AdvanceDeclineOscillatorOutput.AdvanceDeclineOscillatorValue,
        type: OutputType.Scalar,
        mnemonic: this.mnemonic,
        description: this.description,
      }],
    };
  }

  /** Updates the indicator with the given sample (H=L=C, volume=1, so AD is unchanged). */
  public update(sample: number): number {
    if (Number.isNaN(sample)) {
      return Number.NaN;
    }

    return this.updateHLCV(sample, sample, sample, 1);
  }

  /** Updates the indicator with the given high, low, close, and volume values. */
  public updateHLCV(high: number, low: number, close: number, volume: number): number {
    if (Number.isNaN(high) || Number.isNaN(low) || Number.isNaN(close) || Number.isNaN(volume)) {
      return Number.NaN;
    }

    // Compute cumulative AD.
    const tmp = high - low;
    if (tmp > 0) {
      this.ad += ((close - low) - (high - close)) / tmp * volume;
    }

    // Feed AD to both MAs.
    const fast = this.fastMA.update(this.ad);
    const slow = this.slowMA.update(this.ad);
    this.primed = this.fastMA.isPrimed() && this.slowMA.isPrimed();

    if (Number.isNaN(fast) || Number.isNaN(slow)) {
      this.value = Number.NaN;
      return this.value;
    }

    this.value = fast - slow;
    return this.value;
  }

  /** Updates the indicator given the next bar sample, extracting HLCV. */
  public override updateBar(sample: Bar): IndicatorOutput {
    const v = this.updateHLCV(sample.high, sample.low, sample.close, sample.volume);
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = v;
    return [scalar];
  }
}
