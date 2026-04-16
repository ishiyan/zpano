import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorType } from '../../core/indicator-type';
import { LineIndicator } from '../../core/line-indicator';
import { OutputType } from '../../core/outputs/output-type';
import { ZeroLagErrorCorrectingExponentialMovingAverageOutput } from './zero-lag-error-correcting-exponential-moving-average-output';
import { ZeroLagErrorCorrectingExponentialMovingAverageParams } from './zero-lag-error-correcting-exponential-moving-average-params';

/** Function to calculate mnemonic of a __ZeroLagErrorCorrectingExponentialMovingAverage__ indicator. */
export const zeroLagErrorCorrectingExponentialMovingAverageMnemonic = (params: ZeroLagErrorCorrectingExponentialMovingAverageParams): string => {
  const cm = componentTripleMnemonic(
    params.barComponent,
    params.quoteComponent,
    params.tradeComponent,
  );

  const sf = +params.smoothingFactor.toPrecision(4);
  const gl = +params.gainLimit.toPrecision(4);
  const gs = +params.gainStep.toPrecision(4);

  return `zecema(${sf}, ${gl}, ${gs}${cm})`;
};

/**
 * ZeroLagErrorCorrectingExponentialMovingAverage (Ehler's ZECEMA) is an adaptive
 * zero-lag error-correcting exponential moving average.
 *
 * The algorithm iterates over gain values in [-gainLimit, gainLimit] with the given
 * gainStep to find the gain that minimizes the error between the sample and the
 * error-corrected EMA value.
 *
 * The indicator is not primed during the first two updates; it primes on the third.
 *
 * Reference:
 *
 * John Ehlers and Ric Way, 'Zero Lag (well, almost)', TASC, 2010, v28.11, pp30-35.
 */
export class ZeroLagErrorCorrectingExponentialMovingAverage extends LineIndicator {
  private readonly alpha: number;
  private readonly oneMinAlpha: number;
  private readonly _gainLimit: number;
  private readonly _gainStep: number;
  private count: number;
  private value: number;
  private emaValue: number;

  /**
   * Constructs an instance given smoothing factor, gain limit and gain step.
   */
  public constructor(params: ZeroLagErrorCorrectingExponentialMovingAverageParams) {
    super();

    const sf = params.smoothingFactor;
    if (sf <= 0 || sf > 1) {
      throw new Error('smoothing factor should be in (0, 1]');
    }

    const gl = params.gainLimit;
    if (gl <= 0) {
      throw new Error('gain limit should be positive');
    }

    const gs = params.gainStep;
    if (gs <= 0) {
      throw new Error('gain step should be positive');
    }

    this.mnemonic = zeroLagErrorCorrectingExponentialMovingAverageMnemonic(params);
    this.description = 'Zero-lag Error-Correcting Exponential Moving Average ' + this.mnemonic;
    this.barComponent = params.barComponent;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;

    this.alpha = sf;
    this.oneMinAlpha = 1 - sf;
    this._gainLimit = gl;
    this._gainStep = gs;
    this.count = 0;
    this.value = Number.NaN;
    this.emaValue = Number.NaN;
    this.primed = false;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return {
      type: IndicatorType.ZeroLagErrorCorrectingExponentialMovingAverage,
      mnemonic: this.mnemonic,
      description: this.description,
      outputs: [{
        kind: ZeroLagErrorCorrectingExponentialMovingAverageOutput.ZeroLagErrorCorrectingExponentialMovingAverageValue,
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
      this.value = this.calculate(sample);

      return this.value;
    }

    this.count++;

    if (this.count === 1) {
      this.emaValue = sample;

      return Number.NaN;
    }

    if (this.count === 2) {
      this.emaValue = this.calculateEma(sample);
      this.value = this.emaValue;

      return Number.NaN;
    }

    // count === 3: prime the indicator.
    this.value = this.calculate(sample);
    this.primed = true;

    return this.value;
  }

  private calculateEma(sample: number): number {
    return this.alpha * sample + this.oneMinAlpha * this.emaValue;
  }

  private calculate(sample: number): number {
    this.emaValue = this.calculateEma(sample);

    let leastError = Number.MAX_VALUE;
    let bestEC = 0;

    for (let gain = -this._gainLimit; gain <= this._gainLimit; gain += this._gainStep) {
      const ec = this.alpha * (this.emaValue + gain * (sample - this.value)) + this.oneMinAlpha * this.value;
      const error = Math.abs(sample - ec);

      if (leastError > error) {
        leastError = error;
        bestEC = ec;
      }
    }

    return bestEC;
  }
}
