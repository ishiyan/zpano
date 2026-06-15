import { buildMetadata } from '../../core/build-metadata';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { ModifiedExponentialMovingAverageParams } from './params';

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

/** Function to calculate the mnemonic of a __ModifiedExponentialMovingAverage__ indicator. */
export const modifiedExponentialMovingAverageMnemonic = (params: ModifiedExponentialMovingAverageParams): string => {
  const period = Math.floor(params.period ?? 6);
  const degree = Math.floor(params.degree ?? 3);
  const skip = Math.floor(params.skip ?? 1);
  const cm = componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent);

  return `mema(${period},${degree},${skip}${cm})`;
};

/**
 * ModifiedExponentialMovingAverage is Don Mak's Modified Exponential Moving Average (MEMA / MEMA-D).
 *
 * It is a reduced-lag EMA that adds the EMA's own polynomial velocity back to its
 * output, compensating for smoothing delay:
 *
 *   MEMA(n) = EMA(n) + PFD(EMA, degree, order=1, stride=skip)
 *
 * Reference:
 *
 * Mak, Don K. (2006). Mathematical Techniques in Financial Market Trading. Ch 4.2.
 */
export class ModifiedExponentialMovingAverage extends LineIndicator {

  private readonly degree: number;
  private readonly skip: number;
  private readonly nPoints: number;
  private readonly coefficients: number[];

  private readonly emaAlpha: number;
  private emaValue = 0;
  private emaInitialized = false;

  private readonly buf: number[];
  private readonly bufSize: number;
  private bufPos = 0;
  private bufCount = 0;

  public constructor(params: ModifiedExponentialMovingAverageParams) {
    super();

    const period = Math.floor(params.period ?? 6);
    const degree = Math.floor(params.degree ?? 3);
    const skip = Math.floor(params.skip ?? 1);

    if (period < 2) {
      throw new Error('period should be >= 2');
    }
    if (degree < 2) {
      throw new Error('degree should be >= 2');
    }
    if (skip < 1) {
      throw new Error('skip should be >= 1');
    }

    this.mnemonic = modifiedExponentialMovingAverageMnemonic(params);
    this.description = 'Modified exponential moving average ' + this.mnemonic;
    this.barComponent = params.barComponent;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;

    this.degree = degree;
    this.skip = skip;
    this.nPoints = degree + 1;
    this.coefficients = computeVelocityCoefficients(degree);

    this.emaAlpha = 2 / (period + 1);

    this.bufSize = degree * skip + 1;
    this.buf = new Array<number>(this.bufSize).fill(0);
    this.primed = false;
  }

  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.ModifiedExponentialMovingAverage,
      this.mnemonic,
      this.description,
      [{ mnemonic: this.mnemonic, description: this.description }],
    );
  }

  public update(sample: number): number {
    // EMA recursion (seed at first sample).
    if (!this.emaInitialized) {
      this.emaValue = sample;
      this.emaInitialized = true;
    } else {
      this.emaValue = this.emaAlpha * sample + (1 - this.emaAlpha) * this.emaValue;
    }

    // Store EMA value in the ring buffer.
    this.buf[this.bufPos] = this.emaValue;
    this.bufPos = (this.bufPos + 1) % this.bufSize;
    this.bufCount++;

    if (this.bufCount < this.bufSize) {
      this.primed = false;
      return Number.NaN;
    }

    this.primed = true;

    // Read EMA values at stride positions and compute the velocity correction.
    let velocity = 0;
    for (let k = 0; k < this.nPoints; k++) {
      const offset = k * this.skip;
      const idx = ((this.bufPos - 1 - offset) % this.bufSize + this.bufSize) % this.bufSize;
      velocity += this.coefficients[k] * this.buf[idx];
    }

    return this.emaValue + velocity;
  }
}
