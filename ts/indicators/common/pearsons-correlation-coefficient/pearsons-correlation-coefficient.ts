import { buildMetadata } from '../../core/build-metadata';
import { Bar } from '../../../entities/bar';
import { Scalar } from '../../../entities/scalar';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { PearsonsCorrelationCoefficientParams } from './params';

/**
 * __Pearson's Correlation Coefficient__ (__CORREL__) computes the Pearson correlation
 * coefficient (r) between two input series X and Y over a rolling window.
 *
 *     r = (n*sumXY - sumX*sumY) / sqrt((n*sumX2 - sumX^2) * (n*sumY2 - sumY^2))
 *
 * The indicator is not primed during the first length-1 updates.
 */
export class PearsonsCorrelationCoefficient extends LineIndicator {
  private readonly windowX: number[];
  private readonly windowY: number[];
  private readonly _length: number;
  private count = 0;
  private pos = 0;
  private sumX = 0;
  private sumY = 0;
  private sumX2 = 0;
  private sumY2 = 0;
  private sumXY = 0;

  /** Constructs an instance given the parameters. */
  public constructor(params: PearsonsCorrelationCoefficientParams) {
    super();
    const length = Math.floor(params.length);
    if (length < 1) {
      throw new Error('length should be positive');
    }

    this._length = length;
    this.windowX = new Array<number>(length).fill(0);
    this.windowY = new Array<number>(length).fill(0);

    const mn = 'correl(' + length.toString() +
      componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent) + ')';

    this.mnemonic = mn;
    this.description = 'Pearsons Correlation Coefficient ' + mn;
    this.barComponent = params.barComponent;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;
    this.primed = false;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.PearsonsCorrelationCoefficient,
      this.mnemonic,
      this.description,
      [
        { mnemonic: this.mnemonic, description: this.description },
      ],
    );
  }

  /**
   * Updates the indicator given a single scalar sample.
   * Both X and Y are set to the same value (degenerate case).
   */
  public update(sample: number): number {
    return this.updatePair(sample, sample);
  }

  /** Updates the indicator given an (x, y) pair. */
  public updatePair(x: number, y: number): number {
    if (Number.isNaN(x) || Number.isNaN(y)) {
      return Number.NaN;
    }

    const n = this._length;

    if (this.primed) {
      const oldX = this.windowX[this.pos];
      const oldY = this.windowY[this.pos];

      this.sumX -= oldX;
      this.sumY -= oldY;
      this.sumX2 -= oldX * oldX;
      this.sumY2 -= oldY * oldY;
      this.sumXY -= oldX * oldY;

      this.windowX[this.pos] = x;
      this.windowY[this.pos] = y;
      this.pos = (this.pos + 1) % n;

      this.sumX += x;
      this.sumY += y;
      this.sumX2 += x * x;
      this.sumY2 += y * y;
      this.sumXY += x * y;

      return this.correlate(n);
    }

    this.windowX[this.count] = x;
    this.windowY[this.count] = y;

    this.sumX += x;
    this.sumY += y;
    this.sumX2 += x * x;
    this.sumY2 += y * y;
    this.sumXY += x * y;

    this.count++;

    if (this.count === n) {
      this.primed = true;
      this.pos = 0;

      return this.correlate(n);
    }

    return Number.NaN;
  }

  /** Shadows the base updateBar to extract high (X) and low (Y) from the bar. */
  public override updateBar(bar: Bar): IndicatorOutput {
    const v = this.updatePair(bar.high, bar.low);
    const scalar = new Scalar();
    scalar.time = bar.time;
    scalar.value = v;
    return [scalar];
  }

  private correlate(n: number): number {
    const varX = this.sumX2 - (this.sumX * this.sumX) / n;
    const varY = this.sumY2 - (this.sumY * this.sumY) / n;
    const tempReal = varX * varY;

    if (tempReal <= 0) {
      return 0;
    }

    return (this.sumXY - (this.sumX * this.sumY) / n) / Math.sqrt(tempReal);
  }
}
