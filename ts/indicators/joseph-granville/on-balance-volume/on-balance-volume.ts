import { Bar } from '../../../entities/bar';
import { BarComponent, barComponentValue } from '../../../entities/bar-component';
import { Scalar } from '../../../entities/scalar';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorType } from '../../core/indicator-type';
import { LineIndicator } from '../../core/line-indicator';
import { OutputType } from '../../core/outputs/output-type';
import { OnBalanceVolumeOutput } from './on-balance-volume-output';
import { OnBalanceVolumeParams } from './on-balance-volume-params';

/** Function to calculate mnemonic of an __OnBalanceVolume__ indicator. */
export const onBalanceVolumeMnemonic = (): string => {
  return 'obv';
};

/**
 * OnBalanceVolume is Joseph Granville's On-Balance Volume (OBV).
 *
 * OBV is a cumulative volume indicator. On each update, if the price is higher
 * than the previous price, the volume is added to the running total; if the price
 * is lower, the volume is subtracted. If the price is unchanged, the total remains
 * the same.
 *
 * Reference:
 *
 * Granville, Joseph (1963). "Granville's New Key to Stock Market Profits".
 */
export class OnBalanceVolume extends LineIndicator {
  private previousSample: number;
  private value: number;
  private readonly barFunc: (bar: Bar) => number;

  /**
   * Constructs an instance given the parameters.
   */
  public constructor(params?: OnBalanceVolumeParams) {
    super();

    this.previousSample = 0;
    this.value = Number.NaN;
    this.primed = false;

    this.mnemonic = onBalanceVolumeMnemonic();
    this.description = 'On-Balance Volume OBV';

    // OBV defaults to ClosePrice, not TypicalPrice.
    const bc = params?.barComponent ?? BarComponent.Close;
    this.barFunc = barComponentValue(bc);
    this.barComponent = bc;
    this.quoteComponent = params?.quoteComponent;
    this.tradeComponent = params?.tradeComponent;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return {
      type: IndicatorType.OnBalanceVolume,
      mnemonic: this.mnemonic,
      description: this.description,
      outputs: [{
        kind: OnBalanceVolumeOutput.OnBalanceVolumeValue,
        type: OutputType.Scalar,
        mnemonic: this.mnemonic,
        description: this.description,
      }],
    };
  }

  /** Updates the value of the indicator given the next sample (volume = 1). */
  public update(sample: number): number {
    return this.updateWithVolume(sample, 1);
  }

  /** Updates the value of the indicator given the next sample and volume. */
  public updateWithVolume(sample: number, volume: number): number {
    if (Number.isNaN(sample) || Number.isNaN(volume)) {
      return Number.NaN;
    }

    if (!this.primed) {
      this.value = volume;
      this.primed = true;
    } else {
      if (sample > this.previousSample) {
        this.value += volume;
      } else if (sample < this.previousSample) {
        this.value -= volume;
      }
    }

    this.previousSample = sample;

    return this.value;
  }

  /** Updates the indicator given the next bar sample, using bar volume. */
  public override updateBar(sample: Bar): IndicatorOutput {
    const price = this.barFunc(sample);
    const v = this.updateWithVolume(price, sample.volume);
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = v;
    return [scalar];
  }
}
