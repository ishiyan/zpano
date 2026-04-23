import { buildMetadata } from '../../core/build-metadata';
import { Bar } from '../../../entities/bar';
import { BarComponent, barComponentValue } from '../../../entities/bar-component';
import { Scalar } from '../../../entities/scalar';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { OnBalanceVolumeParams } from './params';

/** Function to calculate mnemonic of an __OnBalanceVolume__ indicator. */
export const onBalanceVolumeMnemonic = (params: OnBalanceVolumeParams = {}): string => {
  const suffix = componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent);
  if (suffix === '') {
    return 'obv';
  }
  return `obv(${suffix.slice(2)})`; // strip leading ", "
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

    this.mnemonic = onBalanceVolumeMnemonic(params);
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
    return buildMetadata(
      IndicatorIdentifier.OnBalanceVolume,
      this.mnemonic,
      this.description,
      [
        { mnemonic: this.mnemonic, description: this.description },
      ],
    );
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
