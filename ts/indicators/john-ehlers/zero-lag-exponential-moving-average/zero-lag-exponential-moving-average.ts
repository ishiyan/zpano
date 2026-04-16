import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorType } from '../../core/indicator-type';
import { LineIndicator } from '../../core/line-indicator';
import { OutputType } from '../../core/outputs/output-type';
import { ZeroLagExponentialMovingAverageOutput } from './zero-lag-exponential-moving-average-output';
import { ZeroLagExponentialMovingAverageParams } from './zero-lag-exponential-moving-average-params';

/** Function to calculate mnemonic of a __ZeroLagExponentialMovingAverage__ indicator. */
export const zeroLagExponentialMovingAverageMnemonic = (params: ZeroLagExponentialMovingAverageParams): string => {
  const cm = componentTripleMnemonic(
    params.barComponent,
    params.quoteComponent,
    params.tradeComponent,
  );

  const sf = +params.smoothingFactor.toPrecision(4);
  const gf = +params.velocityGainFactor.toPrecision(4);

  return `zema(${sf}, ${gf}, ${params.velocityMomentumLength}${cm})`;
};

/**
 * ZeroLagExponentialMovingAverage (Ehler's ZEMA) is described in Ehler's book
 * "Rocket Science for Traders" (2001).
 *
 * ZEMA = alpha*(Price + gainFactor*(Price - Price[momentumLength ago])) + (1 - alpha)*ZEMA[prev]
 *
 * The indicator is not primed during the first VelocityMomentumLength updates.
 *
 * Reference:
 *
 * Ehlers, John F. (2001). Rocket Science for Traders. Wiley. pp 167-170.
 */
export class ZeroLagExponentialMovingAverage extends LineIndicator {
  private readonly alpha: number;
  private readonly oneMinAlpha: number;
  private readonly gainFactor: number;
  private readonly momentumLength: number;
  private readonly momentumWindow: number[];
  private count: number;
  private value: number;

  /**
   * Constructs an instance given smoothing factor, gain factor and momentum length.
   */
  public constructor(params: ZeroLagExponentialMovingAverageParams) {
    super();

    const sf = params.smoothingFactor;
    if (sf <= 0 || sf > 1) {
      throw new Error('smoothing factor should be in (0, 1]');
    }

    const ml = Math.floor(params.velocityMomentumLength);
    if (ml < 1) {
      throw new Error('velocity momentum length should be positive');
    }

    this.mnemonic = zeroLagExponentialMovingAverageMnemonic(params);
    this.description = 'Zero-lag Exponential Moving Average ' + this.mnemonic;
    this.barComponent = params.barComponent;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;

    this.alpha = sf;
    this.oneMinAlpha = 1 - sf;
    this.gainFactor = params.velocityGainFactor;
    this.momentumLength = ml;
    this.momentumWindow = new Array<number>(ml + 1).fill(0);
    this.count = 0;
    this.value = Number.NaN;
    this.primed = false;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return {
      type: IndicatorType.ZeroLagExponentialMovingAverage,
      mnemonic: this.mnemonic,
      description: this.description,
      outputs: [{
        kind: ZeroLagExponentialMovingAverageOutput.ZeroLagExponentialMovingAverageValue,
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
      // Shift momentum window left by 1.
      for (let i = 0; i < this.momentumLength; i++) {
        this.momentumWindow[i] = this.momentumWindow[i + 1];
      }

      this.momentumWindow[this.momentumLength] = sample;
      this.value = this.calculate(sample);

      return this.value;
    }

    this.momentumWindow[this.count] = sample;
    this.count++;

    if (this.count <= this.momentumLength) {
      this.value = sample;

      return Number.NaN;
    }

    // count === momentumLength + 1: prime the indicator.
    this.value = this.calculate(sample);
    this.primed = true;

    return this.value;
  }

  private calculate(sample: number): number {
    const momentum = sample - this.momentumWindow[0];

    return this.alpha * (sample + this.gainFactor * momentum) + this.oneMinAlpha * this.value;
  }
}
