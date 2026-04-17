import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorType } from '../../core/indicator-type';
import { LineIndicator } from '../../core/line-indicator';
import { OutputType } from '../../core/outputs/output-type';
import { RelativeStrengthIndexOutput } from './relative-strength-index-output';
import { RelativeStrengthIndexParams } from './relative-strength-index-params';

/** Function to calculate mnemonic of a __RelativeStrengthIndex__ indicator. */
export const relativeStrengthIndexMnemonic = (params: RelativeStrengthIndexParams): string => {
  const cm = componentTripleMnemonic(
    params.barComponent,
    params.quoteComponent,
    params.tradeComponent,
  );

  return `rsi(${params.length}${cm})`;
};

/**
 * RelativeStrengthIndex is Welles Wilder's Relative Strength Index (RSI).
 *
 * RSI measures the magnitude of recent price changes to evaluate overbought
 * or oversold conditions. It oscillates between 0 and 100.
 *
 * Reference:
 *
 * Wilder, J. Welles Jr. (1978). New Concepts in Technical Trading Systems.
 */
export class RelativeStrengthIndex extends LineIndicator {
  private length: number;
  private count: number;
  private previousSample: number;
  private previousGain: number;
  private previousLoss: number;
  private value: number;

  /**
   * Constructs an instance given the parameters.
   */
  public constructor(params: RelativeStrengthIndexParams) {
    super();

    const length = Math.floor(params.length);
    if (length < 2) {
      throw new Error('length should be greater than 1');
    }

    this.length = length;
    this.mnemonic = relativeStrengthIndexMnemonic(params);
    this.description = 'Relative Strength Index ' + this.mnemonic;
    this.barComponent = params.barComponent;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;

    this.count = -1;
    this.previousSample = 0;
    this.previousGain = 0;
    this.previousLoss = 0;
    this.value = Number.NaN;
    this.primed = false;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return {
      type: IndicatorType.RelativeStrengthIndex,
      mnemonic: this.mnemonic,
      description: this.description,
      outputs: [{
        kind: RelativeStrengthIndexOutput.RelativeStrengthIndexValue,
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

    this.count++;

    if (this.count === 0) {
      this.previousSample = sample;
      return this.value;
    }

    const temp = sample - this.previousSample;
    this.previousSample = sample;

    if (!this.primed) {
      // Accumulation phase: count 1..length-1.
      if (temp < 0) {
        this.previousLoss -= temp;
      } else {
        this.previousGain += temp;
      }

      if (this.count < this.length) {
        return this.value;
      }

      // Priming: count === length.
      this.previousGain /= this.length;
      this.previousLoss /= this.length;
      this.primed = true;
    } else {
      // Wilder's smoothing.
      this.previousGain *= (this.length - 1);
      this.previousLoss *= (this.length - 1);

      if (temp < 0) {
        this.previousLoss -= temp;
      } else {
        this.previousGain += temp;
      }

      this.previousGain /= this.length;
      this.previousLoss /= this.length;
    }

    const sum = this.previousGain + this.previousLoss;
    if (sum > epsilon) {
      this.value = 100 * this.previousGain / sum;
    } else {
      this.value = 0;
    }

    return this.value;
  }
}
