import { buildMetadata } from '../../core/build-metadata';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { VelocityCorrectedExponentialMovingAverageParams } from './params';

/**
 * Computes FIR coefficients for the first derivative of a degree-`degree`
 * polynomial fit evaluated at the most recent point (Lagrange basis, order=1).
 */
function computeVelocityCoefficients(degree: number): number[] {
  const nPoints = degree + 1;
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

    let numerator = 0;
    for (let ell = 0; ell < others.length; ell++) {
      let term = 1;
      for (let m = 0; m < others.length; m++) {
        if (m !== ell) {
          term *= others[m];
        }
      }
      numerator += term;
    }

    coefficients.push(numerator / denom);
  }

  return coefficients;
}

/** Function to calculate the mnemonic of a __VelocityCorrectedExponentialMovingAverage__ indicator. */
export const velocityCorrectedExponentialMovingAverageMnemonic =
  (params: VelocityCorrectedExponentialMovingAverageParams): string => {
    const period = Math.floor(params.period ?? 6);
    const degree = Math.floor(params.degree ?? 3);
    const cm = componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent);

    return `vcema(${period},${degree}${cm})`;
  };

/**
 * VelocityCorrectedExponentialMovingAverage is Don Mak's Velocity-Corrected Exponential Moving Average (VCEMA).
 *
 * It is a reduced-lag EMA that pre-corrects price by adding its polynomial velocity
 * before smoothing:
 *
 *   corrected = price + PFD(price, degree, order=1)
 *   VCEMA(n)  = EMA(corrected, n)
 *
 * Reference:
 *
 * Mak, Don K. (2006). Mathematical Techniques in Financial Market Trading. Ch 4.1.
 */
export class VelocityCorrectedExponentialMovingAverage extends LineIndicator {

  private readonly nPoints: number;
  private readonly coefficients: number[];

  private readonly emaAlpha: number;
  private emaValue = 0;
  private emaInitialized = false;

  private readonly buf: number[];
  private bufPos = 0;
  private bufCount = 0;

  public constructor(params: VelocityCorrectedExponentialMovingAverageParams) {
    super();

    const period = Math.floor(params.period ?? 6);
    const degree = Math.floor(params.degree ?? 3);

    if (period < 2) {
      throw new Error('period should be >= 2');
    }
    if (degree < 2) {
      throw new Error('degree should be >= 2');
    }

    this.mnemonic = velocityCorrectedExponentialMovingAverageMnemonic(params);
    this.description = 'Velocity-corrected exponential moving average ' + this.mnemonic;
    this.barComponent = params.barComponent;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;

    this.nPoints = degree + 1;
    this.coefficients = computeVelocityCoefficients(degree);

    this.emaAlpha = 2 / (period + 1);

    this.buf = new Array<number>(this.nPoints).fill(0);
    this.primed = false;
  }

  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.VelocityCorrectedExponentialMovingAverage,
      this.mnemonic,
      this.description,
      [{ mnemonic: this.mnemonic, description: this.description }],
    );
  }

  public update(sample: number): number {
    // Store the raw price in the ring buffer.
    this.buf[this.bufPos] = sample;
    this.bufPos = (this.bufPos + 1) % this.nPoints;
    this.bufCount++;

    if (this.bufCount < this.nPoints) {
      this.primed = false;
      return Number.NaN;
    }

    this.primed = true;

    // Compute the velocity from the raw prices.
    let velocity = 0;
    for (let k = 0; k < this.nPoints; k++) {
      const idx = ((this.bufPos - 1 - k) % this.nPoints + this.nPoints) % this.nPoints;
      velocity += this.coefficients[k] * this.buf[idx];
    }

    // Corrected price = price + velocity.
    const corrected = sample + velocity;

    // Apply the EMA to the corrected price (seed at the first corrected value).
    if (!this.emaInitialized) {
      this.emaValue = corrected;
      this.emaInitialized = true;
    } else {
      this.emaValue = this.emaAlpha * corrected + (1 - this.emaAlpha) * this.emaValue;
    }

    return this.emaValue;
  }
}
