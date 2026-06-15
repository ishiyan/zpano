import { buildMetadata } from '../../core/build-metadata';
import { Bar } from '../../../entities/bar';
import { BarComponent, DefaultBarComponent, barComponentValue } from '../../../entities/bar-component';
import { Quote } from '../../../entities/quote';
import { QuoteComponent, DefaultQuoteComponent, quoteComponentValue } from '../../../entities/quote-component';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { TradeComponent, DefaultTradeComponent, tradeComponentValue } from '../../../entities/trade-component';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { Indicator } from '../../core/indicator';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { CubicVertexParams } from './params';

/** Function to calculate the mnemonic of a __CubicVertex__ indicator. */
export const cubicVertexMnemonic = (params: CubicVertexParams): string => {
  const suffix = componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent);
  if (suffix === '') {
    return 'cvtx';
  }

  return `cvtx(${suffix.slice(2)})`; // strip leading ", "
};

/**
 * CubicVertex is Don Mak's Cubic Vertex (CVTX).
 *
 * It predicts turning points by fitting a cubic polynomial to the 4 most recent price
 * points and computing where the two vertices (extrema) occur relative to the current
 * bar. Given four consecutive prices x(n), x(n-1), x(n-2), x(n-3) (most recent first),
 * the cubic coefficients are (Eq 7.2a-c):
 *
 *   c = (x(n) - 3*x(n-1) + 3*x(n-2) - x(n-3)) / 6
 *   d = (2*x(n) - 5*x(n-1) + 4*x(n-2) - x(n-3)) / 2
 *   e = (11*x(n) - 18*x(n-1) + 9*x(n-2) - 2*x(n-3)) / 6
 *
 * The vertex locations are the roots of 3c*t^2 + 2d*t + e = 0. The near root has the
 * smaller absolute value (more imminent turn); the far root the larger. It works best
 * on pre-smoothed prices.
 *
 * Reference:
 *
 * Mak, Don K. (2003). The Science of Financial Market Trading. Ch 7, Appendix 5.
 */
export class CubicVertex implements Indicator {

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  private readonly buffer: number[] = [0, 0, 0, 0];
  private index = 0;
  private count = 0;
  private primed_ = false;

  private readonly mnemonic_: string;
  private readonly description_: string;

  public constructor(params?: CubicVertexParams) {
    const p = params ?? {};

    const bc = p.barComponent ?? DefaultBarComponent;
    const qc = p.quoteComponent ?? DefaultQuoteComponent;
    const tc = p.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    this.mnemonic_ = cubicVertexMnemonic(p);
    this.description_ = 'Cubic vertex ' + this.mnemonic_;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.primed_;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.CubicVertex,
      this.mnemonic_,
      this.description_,
      [
        { mnemonic: this.mnemonic_ + ' near', description: this.description_ + ' near turn' },
        { mnemonic: this.mnemonic_ + ' far', description: this.description_ + ' far turn' },
      ],
    );
  }

  /**
   * Updates the indicator given the next sample value.
   * Returns [barsToNearTurn, barsToFarTurn].
   */
  public update(sample: number): [number, number] {
    // Store the price in the ring buffer.
    this.buffer[this.index] = sample;
    this.index = (this.index + 1) % 4;
    this.count++;

    if (this.count < 4) {
      this.primed_ = false;
      return [NaN, NaN];
    }

    this.primed_ = true;

    // Extract prices: x[n] (newest), x[n-1], x[n-2], x[n-3] (oldest).
    const xn = this.buffer[((this.index - 1) % 4 + 4) % 4];
    const xn1 = this.buffer[((this.index - 2) % 4 + 4) % 4];
    const xn2 = this.buffer[((this.index - 3) % 4 + 4) % 4];
    const xn3 = this.buffer[((this.index - 4) % 4 + 4) % 4];

    // Cubic polynomial coefficients (Eq 7.2a-c).
    const c = (xn - 3 * xn1 + 3 * xn2 - xn3) / 6;
    const d = (2 * xn - 5 * xn1 + 4 * xn2 - xn3) / 2;
    const e = (11 * xn - 18 * xn1 + 9 * xn2 - 2 * xn3) / 6;

    // Case: c == 0 -- cubic term vanishes, reduces to parabola or line.
    if (c === 0) {
      if (d === 0) {
        return [NaN, NaN];
      }
      const vertex = -e / (2 * d);
      return [vertex, NaN];
    }

    // Full cubic: solve quadratic 3c*t^2 + 2d*t + e = 0.
    const disc = d * d - 3 * c * e;

    if (disc < 0) {
      return [NaN, NaN];
    }

    if (disc === 0) {
      const vertex = -d / (3 * c);
      return [vertex, vertex];
    }

    const sqrtDisc = Math.sqrt(disc);
    const threeC = 3 * c;

    const tPlus = (-d + sqrtDisc) / threeC;
    const tMinus = (-d - sqrtDisc) / threeC;

    if (Math.abs(tPlus) <= Math.abs(tMinus)) {
      return [tPlus, tMinus];
    }

    return [tMinus, tPlus];
  }

  /** Updates the indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    const [near, far] = this.update(sample.value);

    const s0 = new Scalar(); s0.time = sample.time; s0.value = near;
    const s1 = new Scalar(); s1.time = sample.time; s1.value = far;

    return [s0, s1];
  }

  /** Updates the indicator given the next bar sample. */
  public updateBar(sample: Bar): IndicatorOutput {
    const v = this.barComponentFunc(sample);
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = v;
    return this.updateScalar(scalar);
  }

  /** Updates the indicator given the next quote sample. */
  public updateQuote(sample: Quote): IndicatorOutput {
    const v = this.quoteComponentFunc(sample);
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = v;
    return this.updateScalar(scalar);
  }

  /** Updates the indicator given the next trade sample. */
  public updateTrade(sample: Trade): IndicatorOutput {
    const v = this.tradeComponentFunc(sample);
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = v;
    return this.updateScalar(scalar);
  }
}
