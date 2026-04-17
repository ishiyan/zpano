import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorType } from '../../core/indicator-type';
import { LineIndicator } from '../../core/line-indicator';
import { OutputType } from '../../core/outputs/output-type';
import { ExponentialMovingAverage } from '../../common/exponential-moving-average/exponential-moving-average';
import { SimpleMovingAverage } from '../../common/simple-moving-average/simple-moving-average';
import { AbsolutePriceOscillatorOutput } from './absolute-price-oscillator-output';
import { AbsolutePriceOscillatorParams, MovingAverageType } from './absolute-price-oscillator-params';

/** Interface for an indicator that accepts a scalar and returns a value. */
interface LineUpdater {
  update(sample: number): number;
  isPrimed(): boolean;
}

/** Function to calculate mnemonic of an __AbsolutePriceOscillator__ indicator. */
export const absolutePriceOscillatorMnemonic = (params: AbsolutePriceOscillatorParams): string => {
  const cm = componentTripleMnemonic(
    params.barComponent,
    params.quoteComponent,
    params.tradeComponent,
  );

  const maLabel = params.movingAverageType === MovingAverageType.EMA ? 'EMA' : 'SMA';

  return `apo(${maLabel}${params.fastLength}/${maLabel}${params.slowLength}${cm})`;
};

/**
 * AbsolutePriceOscillator is the Absolute Price Oscillator (APO).
 *
 * APO is calculated by subtracting the slower moving average from the faster moving average.
 *
 *   APO = fast_ma - slow_ma
 */
export class AbsolutePriceOscillator extends LineIndicator {
  private fastMA: LineUpdater;
  private slowMA: LineUpdater;
  private value: number;

  /**
   * Constructs an instance given the parameters.
   */
  public constructor(params: AbsolutePriceOscillatorParams) {
    super();

    const fastLength = Math.floor(params.fastLength);
    const slowLength = Math.floor(params.slowLength);

    if (fastLength < 2) {
      throw new Error('fast length should be greater than 1');
    }

    if (slowLength < 2) {
      throw new Error('slow length should be greater than 1');
    }

    this.mnemonic = absolutePriceOscillatorMnemonic(params);
    this.description = 'Absolute Price Oscillator ' + this.mnemonic;
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
      type: IndicatorType.AbsolutePriceOscillator,
      mnemonic: this.mnemonic,
      description: this.description,
      outputs: [{
        kind: AbsolutePriceOscillatorOutput.AbsolutePriceOscillatorValue,
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

    const slow = this.slowMA.update(sample);
    const fast = this.fastMA.update(sample);
    this.primed = this.slowMA.isPrimed() && this.fastMA.isPrimed();

    if (Number.isNaN(fast) || Number.isNaN(slow)) {
      this.value = Number.NaN;
      return this.value;
    }

    this.value = fast - slow;

    return this.value;
  }
}
