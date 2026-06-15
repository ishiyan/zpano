import { buildMetadata } from '../../core/build-metadata';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { PolynomialForecastParams } from './params';

/**
 * Computes FIR coefficients for the `order`-th derivative of a degree-`degree`
 * polynomial fit evaluated at the most recent point (Lagrange basis).
 */
function computeCoefficients(degree: number, order: number): number[] {
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
    if (order === 1) {
      for (let ell = 0; ell < others.length; ell++) {
        let term = 1;
        for (let m = 0; m < others.length; m++) {
          if (m !== ell) {
            term *= others[m];
          }
        }
        numerator += term;
      }
    } else {
      for (let ell = 0; ell < others.length; ell++) {
        for (let r = ell + 1; r < others.length; r++) {
          let term = 2;
          for (let m = 0; m < others.length; m++) {
            if (m !== ell && m !== r) {
              term *= others[m];
            }
          }
          numerator += term;
        }
      }
    }

    coefficients.push(numerator / denom);
  }

  return coefficients;
}

/** Function to calculate the mnemonic of a __PolynomialForecast__ indicator. */
export const polynomialForecastMnemonic = (params: PolynomialForecastParams): string => {
  const degree = Math.floor(params.degree ?? 3);
  const order = Math.floor(params.order ?? 1);
  const smoothing = Math.floor(params.smoothing ?? 0);
  const cm = componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent);

  return `pof(${degree},${order},${smoothing}${cm})`;
};

/**
 * PolynomialForecast is Don Mak's Polynomial Forecast (POF).
 *
 * It is a one-step-ahead price forecast using a Taylor series expansion built on
 * polynomial fit derivatives (PFD):
 *
 *   velocity     = PFD(price, degree, order=1)
 *   acceleration = PFD(price, degree, order=2)
 *   order=1:  forecast = price + velocity
 *   order=2:  forecast = price + velocity + 0.5*acceleration
 *
 * Reference:
 *
 * Mak, Don K. (2003). The Science of Financial Market Trading. Ch 10.2.
 */
export class PolynomialForecast extends LineIndicator {

  private readonly degree: number;
  private readonly order: number;
  private readonly smoothing: number;
  private readonly nPoints: number;
  private readonly coeffVel: number[];
  private readonly coeffAcc: number[] | null;

  private readonly emaAlpha: number;
  private emaValue = 0;
  private emaInitialized = false;

  private readonly buf: number[];
  private bufPos = 0;
  private bufCount = 0;

  public constructor(params: PolynomialForecastParams) {
    super();

    const degree = Math.floor(params.degree ?? 3);
    const order = Math.floor(params.order ?? 1);
    const smoothing = Math.floor(params.smoothing ?? 0);

    if (degree < 2) {
      throw new Error('degree should be >= 2');
    }
    if (order < 1 || order > 2) {
      throw new Error('order should be 1 or 2');
    }
    if (smoothing < 0) {
      throw new Error('smoothing should be >= 0');
    }

    this.mnemonic = polynomialForecastMnemonic(params);
    this.description = 'Polynomial forecast ' + this.mnemonic;
    this.barComponent = params.barComponent;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;

    this.degree = degree;
    this.order = order;
    this.smoothing = smoothing;
    this.nPoints = degree + 1;
    this.coeffVel = computeCoefficients(degree, 1);
    this.coeffAcc = order === 2 ? computeCoefficients(degree, 2) : null;

    this.emaAlpha = smoothing > 0 ? 2 / (smoothing + 1) : 0;

    this.buf = new Array<number>(this.nPoints).fill(0);
    this.primed = false;
  }

  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.PolynomialForecast,
      this.mnemonic,
      this.description,
      [{ mnemonic: this.mnemonic, description: this.description }],
    );
  }

  public update(sample: number): number {
    // Optional EMA pre-smoothing.
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

    // Store the smoothed price in the ring buffer.
    this.buf[this.bufPos] = smoothed;
    this.bufPos = (this.bufPos + 1) % this.nPoints;
    this.bufCount++;

    if (this.bufCount < this.nPoints) {
      this.primed = false;
      return Number.NaN;
    }

    this.primed = true;

    // Read buffer most-recent-first and compute velocity (and acceleration).
    let velocity = 0;
    let acceleration = 0;
    for (let k = 0; k < this.nPoints; k++) {
      const idx = ((this.bufPos - 1 - k) % this.nPoints + this.nPoints) % this.nPoints;
      const value = this.buf[idx];
      velocity += this.coeffVel[k] * value;
      if (this.coeffAcc !== null) {
        acceleration += this.coeffAcc[k] * value;
      }
    }

    let forecast = smoothed + velocity;
    if (this.order === 2) {
      forecast += 0.5 * acceleration;
    }

    return forecast;
  }
}
