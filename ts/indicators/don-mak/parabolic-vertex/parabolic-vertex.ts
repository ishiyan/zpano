import { buildMetadata } from '../../core/build-metadata';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { ParabolicVertexParams } from './params';

/** Function to calculate the mnemonic of a __ParabolicVertex__ indicator. */
export const parabolicVertexMnemonic = (params: ParabolicVertexParams): string => {
  const suffix = componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent);
  if (suffix === '') {
    return 'pvtx';
  }

  return `pvtx(${suffix.slice(2)})`; // strip leading ", "
};

/**
 * ParabolicVertex is Don Mak's Parabolic Vertex (PVTX).
 *
 * It predicts turning points by fitting a parabola to the 3 most recent price points
 * and computing where the vertex (extremum) occurs relative to the current bar. Given
 * three consecutive prices x(n), x(n-1), x(n-2) (most recent first) fitted to
 * x(t) = d*t^2 + e*t + f at t = 0, -1, -2, the vertex is at:
 *
 *   t_v = -(1.5*x(n) - 2*x(n-1) + 0.5*x(n-2)) / (x(n) - 2*x(n-1) + x(n-2))
 *
 * The output is the number of bars from the current bar to the predicted turning
 * point. It works best on pre-smoothed prices.
 *
 * Reference:
 *
 * Mak, Don K. (2003). The Science of Financial Market Trading. Ch 7, Appendix 5.
 */
export class ParabolicVertex extends LineIndicator {

  private readonly buffer: number[] = [0, 0, 0];
  private index = 0;
  private count = 0;

  public constructor(params: ParabolicVertexParams) {
    super();

    this.mnemonic = parabolicVertexMnemonic(params);
    this.description = 'Parabolic vertex ' + this.mnemonic;
    this.barComponent = params.barComponent;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;

    this.primed = false;
  }

  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.ParabolicVertex,
      this.mnemonic,
      this.description,
      [{ mnemonic: this.mnemonic, description: this.description }],
    );
  }

  public update(sample: number): number {
    // Store the price in the ring buffer.
    this.buffer[this.index] = sample;
    this.index = (this.index + 1) % 3;
    this.count++;

    if (this.count < 3) {
      this.primed = false;
      return Number.NaN;
    }

    this.primed = true;

    // Extract prices: x[n] (newest), x[n-1], x[n-2] (oldest).
    const xn = this.buffer[((this.index - 1) % 3 + 3) % 3];
    const xn1 = this.buffer[((this.index - 2) % 3 + 3) % 3];
    const xn2 = this.buffer[((this.index - 3) % 3 + 3) % 3];

    // Denominator = second-order finite difference (proportional to curvature).
    const denom = xn - 2 * xn1 + xn2;
    if (denom === 0) {
      return Number.NaN;
    }

    const numer = 1.5 * xn - 2 * xn1 + 0.5 * xn2;

    return -numer / denom;
  }
}
