import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorType } from '../../core/indicator-type';
import { LineIndicator } from '../../core/line-indicator';
import { OutputType } from '../../core/outputs/output-type';
import { ExponentialMovingAverage } from '../../common/exponential-moving-average/exponential-moving-average';
import { SimpleMovingAverage } from '../../common/simple-moving-average/simple-moving-average';
import { PercentagePriceOscillatorOutput } from './percentage-price-oscillator-output';
import { PercentagePriceOscillatorParams, MovingAverageType } from './percentage-price-oscillator-params';

/** Interface for an indicator that accepts a scalar and returns a value. */
interface LineUpdater {
  update(sample: number): number;
  isPrimed(): boolean;
}

/** Function to calculate mnemonic of a __PercentagePriceOscillator__ indicator. */
export const percentagePriceOscillatorMnemonic = (params: PercentagePriceOscillatorParams): string => {
  const cm = componentTripleMnemonic(
    params.barComponent,
    params.quoteComponent,
    params.tradeComponent,
  );

  const maLabel = params.movingAverageType === MovingAverageType.EMA ? 'EMA' : 'SMA';

  return `ppo(${maLabel}${params.fastLength}/${maLabel}${params.slowLength}${cm})`;
};

/**
 * PercentagePriceOscillator is Gerald Appel's Percentage Price Oscillator (PPO).
 *
 * PPO is calculated by subtracting the slower moving average from the faster moving
 * average and then dividing the result by the slower moving average, expressed as a percentage.
 *
 *   PPO = 100 * (fast_ma - slow_ma) / slow_ma
 *
 * Reference:
 *
 * Appel, Gerald (2005). Technical Analysis: Power Tools for Active Investors.
 */
export class PercentagePriceOscillator extends LineIndicator {
  private fastMA: LineUpdater;
  private slowMA: LineUpdater;
  private value: number;

  /**
   * Constructs an instance given the parameters.
   */
  public constructor(params: PercentagePriceOscillatorParams) {
    super();

    const fastLength = Math.floor(params.fastLength);
    const slowLength = Math.floor(params.slowLength);

    if (fastLength < 2) {
      throw new Error('fast length should be greater than 1');
    }

    if (slowLength < 2) {
      throw new Error('slow length should be greater than 1');
    }

    this.mnemonic = percentagePriceOscillatorMnemonic(params);
    this.description = 'Percentage Price Oscillator ' + this.mnemonic;
    this.barComponent = params.barComponent;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;

    if (params.movingAverageType === MovingAverageType.EMA) {
      this.fastMA = new ExponentialMovingAverage({
        length: fastLength,
        firstIsAverage: params.firstIsAverage ?? false,
      });
      this.slowMA = new ExponentialMovingAverage({
        length: slowLength,
        firstIsAverage: params.firstIsAverage ?? false,
      });
    } else {
      this.fastMA = new SimpleMovingAverage({ length: fastLength });
      this.slowMA = new SimpleMovingAverage({ length: slowLength });
    }

    this.value = Number.NaN;
    this.primed = false;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return {
      type: IndicatorType.PercentagePriceOscillator,
      mnemonic: this.mnemonic,
      description: this.description,
      outputs: [{
        kind: PercentagePriceOscillatorOutput.PercentagePriceOscillatorValue,
        type: OutputType.Scalar,
        mnemonic: this.mnemonic,
        description: this.description,
      }],
    };
  }

  /** Updates the value of the indicator given the next sample. */
  public update(sample: number): number {
    const epsilon = 1e-8;

    if (Number.isNaN(sample)) {
      return sample;
    }

    const slow = this.slowMA.update(sample);
    const fast = this.fastMA.update(sample);
    this.primed = this.slowMA.isPrimed() && this.fastMA.isPrimed();

    if (Number.isNaN(fast) || Number.isNaN(slow)) {
      this.value = Number.NaN;
      return this.value;
    }

    if (Math.abs(slow) < epsilon) {
      this.value = 0;
    } else {
      this.value = 100 * (fast - slow) / slow;
    }

    return this.value;
  }
}
