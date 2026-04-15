import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorType } from '../../core/indicator-type';
import { LineIndicator } from '../../core/line-indicator';
import { OutputType } from '../../core/outputs/output-type';
import { TriangularMovingAverageOutput } from './triangular-moving-average-output';
import { TriangularMovingAverageParams } from './triangular-moving-average-params';

/** Function to calculate mnemonic of a __TriangularMovingAverage__ indicator. */
export const triangularMovingAverageMnemonic = (params: TriangularMovingAverageParams): string =>
  'trima('.concat(params.length.toString(), componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent), ')');

/** Triangular Moving Average line indicator.
 *
 * Computes the triangular moving average (TRIMA) like a weighted moving average.
 * Instead of the WMA who put more weight on the latest sample, the TRIMA puts
 * more weight on the data in the middle of the window.
 *
 * Using algebra, it can be demonstrated that the TRIMA is equivalent
 * to doing a SMA of a SMA. The following explain the rules.
 *
 *     ➊ When the period π is even, TRIMA(x,π) = SMA(SMA(x,π/2), (π/2)+1).
 *
 *     ➋ When the period π is odd, TRIMA(x,π) = SMA(SMA(x,(π+1)/2), (π+1)/2).
 *
 * The SMA of a SMA is the algorithm generally found in books.
 *
 * TradeStation deviate from the generally accepted implementation
 * by making the TRIMA to be as follows:
 *
 *     TRIMA(x,π) = SMA(SMA(x, (int)(π/2)+1), (int)(π/2)+1).
 *
 * This formula is done regardless if the period is even or odd. In other words:
 *
 *    ➊ A period of 4 becomes TRIMA(x,4) = SMA(SMA(x,3), 3).
 *
 *    ➋ A period of 5 becomes TRIMA(x,5) = SMA(SMA(x,3), 3).
 *
 *    ➌ A period of 6 becomes TRIMA(x,6) = SMA(SMA(x,4), 4).
 *
 *    ➍ A period of 7 becomes TRIMA(x,7) = SMA(SMA(x,4), 4).
 *
 * The Metastock implementation is the same as the generally accepted one.
 *
 * To optimize speed, this implementation uses a better algorithm than the usual SMA of a SMA.
 * The calculation from one TRIMA value to the next is done by doing 4 little adjustments.
 *
 * The following show a TRIMA 4-period:
 *
 *    TRIMA at time δ: ((1*α)+(2*β)+(2*γ)+(1*δ)) / 6
 *
 *    TRIMA at time ε: ((1*β)+(2*γ)+(2*δ)+(1*ε)) / 6
 *
 * To go from TRIMA δ to ε, the following is done:
 *
 *    ➊ α and β are subtract from the numerator.
 *
 *    ➋ δ is added to the numerator.
 *
 *    ➌ ε is added to the numerator.
 *
 *    ➍ TRIMA is calculated by doing numerator / 6.
 *
 *    ➎ Sequence is repeated for the next output.
 */
export class TriangularMovingAverage extends LineIndicator {
  private window: Array<number>;
  private windowLength: number;
  private windowLengthHalf: number;
  private windowCount = 0;
  private lastIndex: number;
  private factor: number;
  private numerator = 0;
  private numeratorSub = 0;
  private numeratorAdd = 0;
  private isOdd: boolean;

  /**
   * Constructs an instance given a length in samples.
   * The length should be an integer greater than 1.
   **/
  public constructor(params: TriangularMovingAverageParams){
    super();
    const length = Math.floor(params.length);
    if (length < 2) {
      throw new Error('length should be greater than 1');
    }

    this.mnemonic = triangularMovingAverageMnemonic(params);
    this.description = 'Triangular moving average ' + this.mnemonic;
    this.barComponent = params.barComponent;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;
    this.primed = false;
    this.window = new Array<number>(length);
    this.windowLength = length;
    this.windowLengthHalf = Math.floor(length/2);
    this.lastIndex = length - 1;
    this.isOdd = (length%2) === 1;
    const l = 1 + this.windowLengthHalf;
    if (this.isOdd) {
      // Let period = 5 and l=(int)(period/2), then the formula for a "triangular" series is:
      // 1+2+3+2+1 = l*(l+1) + l+1 = (l+1)*(l+1) = 3*3 = 9.
      this.factor = 1 / (l*l);
    } else {
      // Let period = 6 and l=(int)(period/2), then  the formula for a "triangular" series is:
      // 1+2+3+3+2+1 = l*(l+1) = 3*4 = 12.
      this.factor = 1 / (this.windowLengthHalf*l);
      --this.windowLengthHalf;
    }
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return {
      type: IndicatorType.TriangularMovingAverage,
      mnemonic: this.mnemonic,
      description: this.description,
      outputs: [{
        kind: TriangularMovingAverageOutput.TriangularMovingAverageValue,
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

    if (this.primed) {
      this.numerator -= this.numeratorSub;
      this.numeratorSub -= this.window[0];

      for (let i = 0; i < this.lastIndex; i++) {
        this.window[i] = this.window[i+1];
      }

      this.window[this.lastIndex] = sample;
      const temp = this.window[this.windowLengthHalf];
      this.numeratorSub += temp;

      if (this.isOdd) { // The logic for an odd length.
        this.numerator += this.numeratorAdd;
        this.numeratorAdd -= temp;
      } else { // The logic for an even length.
        this.numeratorAdd -= temp;
        this.numerator += this.numeratorAdd;
      }

      this.numeratorAdd += sample;
      this.numerator += sample;
    } else { // Not primed.
      this.window[this.windowCount] = sample;
      this.windowCount++;

      if (this.windowLength > this.windowCount) {
        return Number.NaN;
      }

      for (let i = this.windowLengthHalf; i >= 0; i--) {
        this.numeratorSub += this.window[i];
        this.numerator += this.numeratorSub;
      }

      for (let i = this.windowLengthHalf + 1; i < this.windowLength; i++) {
        this.numeratorAdd += this.window[i];
        this.numerator += this.numeratorAdd;
      }

      this.primed = true;
    }

    return this.numerator * this.factor;
  }
}
