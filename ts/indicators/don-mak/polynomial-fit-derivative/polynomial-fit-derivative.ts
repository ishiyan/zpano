import { buildMetadata } from '../../core/build-metadata';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { PolynomialFitDerivativeParams } from './params';

/**
 * Computes the FIR filter coefficients for the order-th derivative of a
 * degree-`degree` polynomial fit, evaluated at the most recent point.
 *
 * Uses the Lagrange basis with the elementary-symmetric-polynomial identity:
 *   c_i = order! * e_{degree-order}(others) / prod_{j != i} (j - i)
 * where `others` is the set of point positions {0..degree} excluding i.
 */
function computeCoefficients(degree: number, order: number): number[] {
  const nPoints = degree + 1;

  let factorialOrder = 1;
  for (let f = 2; f <= order; f++) {
    factorialOrder *= f;
  }

  const coefficients: number[] = [];

  for (let i = 0; i < nPoints; i++) {
    let denom = 1;
    for (let j = 0; j < nPoints; j++) {
      if (j !== i) {
        denom *= j - i;
      }
    }

    const others: number[] = [];
    for (let j = 0; j < nPoints; j++) {
      if (j !== i) {
        others.push(j);
      }
    }

    const m = others.length; // equals degree

    // Elementary symmetric polynomials e[0..m] of the values in `others`.
    const e = new Array<number>(m + 1).fill(0);
    e[0] = 1;
    for (const v of others) {
      for (let k = m; k >= 1; k--) {
        e[k] += v * e[k - 1];
      }
    }

    const numerator = factorialOrder * e[m - order];
    coefficients.push(numerator / denom);
  }

  return coefficients;
}

/** Function to calculate the mnemonic of a __PolynomialFitDerivative__ indicator. */
export const polynomialFitDerivativeMnemonic = (params: PolynomialFitDerivativeParams): string => {
  const degree = Math.floor(params.degree ?? 3);
  const order = Math.floor(params.order ?? 1);
  const smoothing = Math.floor(params.smoothing ?? 6);
  const cm = componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent);

  return `pfd(${degree},${order},${smoothing}${cm})`;
};

/**
 * PolynomialFitDerivative is Don Mak's Polynomial Fit Derivative (PFD) indicator.
 *
 * It fits a polynomial of degree `degree` to the most recent `degree + 1`
 * (optionally EMA-smoothed) prices and evaluates its `order`-th derivative at the
 * current bar. This is a FIR filter: a dot product of fixed Lagrange-derived
 * coefficients with the last `degree + 1` smoothed prices.
 *
 * Reference:
 *
 * Mak, Don K. (2003). The Science of Financial Market Trading. Ch 6.
 * Mak, Don K. (2006). Mathematical Techniques in Financial Market Trading. Ch 8.
 */
export class PolynomialFitDerivative extends LineIndicator {

  private readonly coefficients: number[];
  private readonly nPoints: number;

  private readonly smoothing: number;
  private readonly emaAlpha: number;
  private emaValue = 0;
  private emaInitialized = false;

  private readonly buf: number[];
  private bufPos = 0;
  private bufCount = 0;

  public constructor(params: PolynomialFitDerivativeParams) {
    super();

    const degree = Math.floor(params.degree ?? 3);
    const order = Math.floor(params.order ?? 1);
    const smoothing = Math.floor(params.smoothing ?? 6);

    if (degree < 2) {
      throw new Error('degree should be >= 2');
    }
    if (order < 1 || order > degree) {
      throw new Error('order should be >= 1 and <= degree');
    }
    if (smoothing < 0) {
      throw new Error('smoothing should be >= 0');
    }

    this.mnemonic = polynomialFitDerivativeMnemonic(params);
    this.description = 'Polynomial fit derivative ' + this.mnemonic;
    this.barComponent = params.barComponent;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;

    this.coefficients = computeCoefficients(degree, order);
    this.nPoints = degree + 1;
    this.smoothing = smoothing;
    this.emaAlpha = smoothing > 0 ? 2 / (smoothing + 1) : 0;
    this.buf = new Array<number>(this.nPoints).fill(0);
    this.primed = false;
  }

  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.PolynomialFitDerivative,
      this.mnemonic,
      this.description,
      [{ mnemonic: this.mnemonic, description: this.description }],
    );
  }

  public update(sample: number): number {
    // Step 1: optional EMA smoothing.
    let smoothed = sample;
    if (this.smoothing > 0) {
      if (!this.emaInitialized) {
        this.emaValue = sample;
        this.emaInitialized = true;
      } else {
        this.emaValue = this.emaAlpha * sample + (1 - this.emaAlpha) * this.emaValue;
      }
      smoothed = this.emaValue;
    }

    // Step 2: push into the ring buffer.
    this.buf[this.bufPos] = smoothed;
    this.bufPos = (this.bufPos + 1) % this.nPoints;
    this.bufCount++;

    // Step 3: not enough data yet.
    if (this.bufCount < this.nPoints) {
      this.primed = false;
      return Number.NaN;
    }

    // Step 4: FIR dot product (coefficients[j] multiplies the j-th most recent).
    let result = 0;
    for (let j = 0; j < this.nPoints; j++) {
      const bufIdx = ((this.bufPos - 1 - j) % this.nPoints + this.nPoints) % this.nPoints;
      result += this.coefficients[j] * this.buf[bufIdx];
    }

    this.primed = true;
    return result;
  }
}
