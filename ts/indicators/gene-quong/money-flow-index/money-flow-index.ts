import { Bar } from '../../../entities/bar';
import { BarComponent, barComponentValue } from '../../../entities/bar-component';
import { Scalar } from '../../../entities/scalar';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorType } from '../../core/indicator-type';
import { LineIndicator } from '../../core/line-indicator';
import { OutputType } from '../../core/outputs/output-type';
import { MoneyFlowIndexOutput } from './money-flow-index-output';
import { MoneyFlowIndexParams } from './money-flow-index-params';

/** Function to calculate mnemonic of a __MoneyFlowIndex__ indicator. */
export const moneyFlowIndexMnemonic = (params: MoneyFlowIndexParams): string => {
  return `mfi(${params.length})`;
};

/**
 * MoneyFlowIndex is Gene Quong's Money Flow Index (MFI).
 *
 * MFI is a volume-weighted oscillator calculated over ℓ periods, showing money flow
 * on up days as a percentage of the total of up and down days.
 *
 *   TypicalPrice = (High + Low + Close) / 3
 *   MoneyFlow = TypicalPrice × Volume
 *   MFI = 100 × PositiveMoneyFlow / (PositiveMoneyFlow + NegativeMoneyFlow)
 *
 * A value of 80 is generally considered overbought, or a value of 20 oversold.
 *
 * Reference:
 *
 * Quong, Gene, and Soudack, Avrum (1989). "Volume-Weighted RSI: Money Flow Index".
 * Technical Analysis of Stocks and Commodities.
 */
export class MoneyFlowIndex extends LineIndicator {
  private readonly length: number;
  private readonly negativeBuffer: number[];
  private readonly positiveBuffer: number[];
  private negativeSum: number;
  private positiveSum: number;
  private previousSample: number;
  private bufferIndex: number;
  private bufferLowIndex: number;
  private bufferCount: number;
  private value: number;
  private readonly barFunc: (bar: Bar) => number;

  /**
   * Constructs an instance given the parameters.
   */
  public constructor(params: MoneyFlowIndexParams) {
    super();

    const length = Math.floor(params.length);

    if (length < 1) {
      throw new Error('length should be greater than 0');
    }

    this.length = length;
    this.negativeBuffer = new Array<number>(length).fill(0);
    this.positiveBuffer = new Array<number>(length).fill(0);
    this.negativeSum = 0;
    this.positiveSum = 0;
    this.previousSample = 0;
    this.bufferIndex = 0;
    this.bufferLowIndex = 0;
    this.bufferCount = 0;
    this.value = Number.NaN;
    this.primed = false;

    this.mnemonic = moneyFlowIndexMnemonic(params);
    this.description = 'Money Flow Index ' + this.mnemonic;

    // MFI defaults to TypicalPrice, not ClosePrice.
    const bc = params.barComponent ?? BarComponent.Typical;
    this.barFunc = barComponentValue(bc);
    this.barComponent = bc;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return {
      type: IndicatorType.MoneyFlowIndex,
      mnemonic: this.mnemonic,
      description: this.description,
      outputs: [{
        kind: MoneyFlowIndexOutput.MoneyFlowIndexValue,
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

    const lengthMinOne = this.length - 1;

    if (this.primed) {
      this.negativeSum -= this.negativeBuffer[this.bufferLowIndex];
      this.positiveSum -= this.positiveBuffer[this.bufferLowIndex];

      const amount = sample * volume;
      const diff = sample - this.previousSample;

      if (diff < 0) {
        this.negativeBuffer[this.bufferIndex] = amount;
        this.positiveBuffer[this.bufferIndex] = 0;
        this.negativeSum += amount;
      } else if (diff > 0) {
        this.negativeBuffer[this.bufferIndex] = 0;
        this.positiveBuffer[this.bufferIndex] = amount;
        this.positiveSum += amount;
      } else {
        this.negativeBuffer[this.bufferIndex] = 0;
        this.positiveBuffer[this.bufferIndex] = 0;
      }

      const sum = this.positiveSum + this.negativeSum;
      this.value = sum < 1 ? 0 : (100 * this.positiveSum / sum);

      this.bufferIndex++;
      if (this.bufferIndex > lengthMinOne) {
        this.bufferIndex = 0;
      }

      this.bufferLowIndex++;
      if (this.bufferLowIndex > lengthMinOne) {
        this.bufferLowIndex = 0;
      }
    } else if (this.bufferCount === 0) {
      this.bufferCount++;
    } else {
      const amount = sample * volume;
      const diff = sample - this.previousSample;

      if (diff < 0) {
        this.negativeBuffer[this.bufferIndex] = amount;
        this.positiveBuffer[this.bufferIndex] = 0;
        this.negativeSum += amount;
      } else if (diff > 0) {
        this.negativeBuffer[this.bufferIndex] = 0;
        this.positiveBuffer[this.bufferIndex] = amount;
        this.positiveSum += amount;
      } else {
        this.negativeBuffer[this.bufferIndex] = 0;
        this.positiveBuffer[this.bufferIndex] = 0;
      }

      if (this.length === this.bufferCount) {
        const sum = this.positiveSum + this.negativeSum;
        this.value = sum < 1 ? 0 : (100 * this.positiveSum / sum);
        this.primed = true;
      }

      this.bufferIndex++;
      if (this.bufferIndex > lengthMinOne) {
        this.bufferIndex = 0;
      }

      this.bufferCount++;
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
