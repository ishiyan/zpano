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
import { InstantaneousSineWavePeriodParams } from './params';

/** Function to calculate the mnemonic of an __InstantaneousSineWavePeriod__ indicator. */
export const instantaneousSineWavePeriodMnemonic = (
  smoothing: number, minPeriod: number, maxPeriod: number, errorThreshold: number, dx: number,
  barComponent?: BarComponent, quoteComponent?: QuoteComponent, tradeComponent?: TradeComponent,
): string => {
  const cm = componentTripleMnemonic(barComponent, quoteComponent, tradeComponent);

  return `iswp(${smoothing},${minPeriod.toFixed(2)},${maxPeriod.toFixed(2)},${errorThreshold.toFixed(2)},${dx.toFixed(2)}${cm})`;
};

/**
 * InstantaneousSineWavePeriod is Don Mak's Instantaneous Sine Wave Period (ISWP) indicator.
 *
 * It estimates the dominant cycle period of price data by modeling it locally as a
 * single sine wave superimposed on a constant level, combining a 4-point method
 * (IF4) and a 5-point method (IF5) and selecting the one with the lower estimation
 * error at each bar.
 *
 * The indicator produces seven outputs:
 *   - Period: cycle period in bars (T = 2*pi/omega), NaN if invalid;
 *   - Omega: circular frequency in radians/bar, NaN if invalid;
 *   - Velocity: wave velocity, NaN if invalid;
 *   - Acceleration: wave acceleration, NaN if invalid;
 *   - Amplitude: sine wave amplitude, NaN if invalid;
 *   - Phase: phase angle in radians, NaN if invalid;
 *   - DcLevel: constant level D, NaN if invalid.
 *
 * Reference:
 *
 * Mak, Don K. (2006). Mathematical Techniques in Financial Market Trading.
 */
export class InstantaneousSineWavePeriod implements Indicator {

  private readonly smoothing: number;
  private readonly minPeriod: number;
  private readonly maxPeriod: number;
  private readonly errorThreshold: number;
  private readonly dx: number;

  private readonly emaAlpha: number;
  private emaValue = 0;
  private emaPrimed = false;

  private readonly buffer: number[] = [0, 0, 0, 0, 0];
  private count = 0;

  private primed_ = false;

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  private readonly mnemonic_: string;
  private readonly description_: string;

  constructor(params?: InstantaneousSineWavePeriodParams) {
    const p = params ?? {};

    const smoothing = Math.floor(p.smoothing ?? 0);
    const minPeriod = p.minPeriod ?? 4.0;
    const maxPeriod = p.maxPeriod ?? 50.0;
    const errorThreshold = p.errorThreshold ?? 20.0;
    const dx = p.dx ?? 0.01;

    if (smoothing < 0) {
      throw new Error('smoothing should be >= 0');
    }
    if (minPeriod <= 0) {
      throw new Error('minPeriod should be > 0');
    }
    if (maxPeriod <= minPeriod) {
      throw new Error('maxPeriod should be > minPeriod');
    }
    if (errorThreshold <= 0) {
      throw new Error('errorThreshold should be > 0');
    }
    if (dx <= 0) {
      throw new Error('dx should be > 0');
    }

    const bc = p.barComponent ?? DefaultBarComponent;
    const qc = p.quoteComponent ?? DefaultQuoteComponent;
    const tc = p.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    this.smoothing = smoothing;
    this.minPeriod = minPeriod;
    this.maxPeriod = maxPeriod;
    this.errorThreshold = errorThreshold;
    this.dx = dx;

    this.emaAlpha = smoothing > 0 ? 2 / (smoothing + 1) : 1;

    this.mnemonic_ = instantaneousSineWavePeriodMnemonic(
      smoothing, minPeriod, maxPeriod, errorThreshold, dx,
      p.barComponent, p.quoteComponent, p.tradeComponent,
    );
    this.description_ = 'Instantaneous Sine Wave Period ' + this.mnemonic_;
  }

  private applyEma(price: number): number {
    if (!this.emaPrimed) {
      this.emaValue = price;
      this.emaPrimed = true;
    } else {
      this.emaValue = this.emaAlpha * price + (1 - this.emaAlpha) * this.emaValue;
    }
    return this.emaValue;
  }

  private pushBuffer(value: number): void {
    for (let i = 4; i > 0; i--) {
      this.buffer[i] = this.buffer[i - 1];
    }
    this.buffer[0] = value;
  }

  private calcOmega4(): [number, number] {
    const x0 = this.buffer[0];
    const xm1 = this.buffer[1];
    const xm2 = this.buffer[2];
    const xm3 = this.buffer[3];

    const den = xm1 - xm2;
    if (den === 0) {
      return [NaN, this.errorThreshold];
    }

    const ratio = (x0 - xm3) / den;

    const sqrtArg = 3 - ratio;
    if (sqrtArg < 0) {
      return [NaN, this.errorThreshold];
    }

    const arg = 0.5 * Math.sqrt(sqrtArg);
    if (arg > 1) {
      return [NaN, this.errorThreshold];
    }

    const omega4 = 2 * Math.asin(arg);

    const dx2 = this.dx * this.dx;

    const denom1 = 1 - 0.25 * sqrtArg;
    if (denom1 <= 0 || sqrtArg === 0) {
      return [omega4, this.errorThreshold];
    }

    const f1 = 1 / (denom1 * sqrtArg);
    const invDen2 = 1 / (den * den);
    const q2 = invDen2 * (dx2 + dx2) + (ratio * ratio) * invDen2 * (dx2 + dx2);

    const product = f1 * q2;
    if (product < 0) {
      return [omega4, this.errorThreshold];
    }

    return [omega4, 0.5 * Math.sqrt(product)];
  }

  private calcOmega5(): [number, number] {
    const x0 = this.buffer[0];
    const xm1 = this.buffer[1];
    const xm3 = this.buffer[3];
    const xm4 = this.buffer[4];

    const den1 = xm1 - xm3;
    if (den1 === 0) {
      return [NaN, this.errorThreshold];
    }

    const arg = 0.5 * (x0 - xm4) / den1;
    if (Math.abs(arg) > 1) {
      return [NaN, this.errorThreshold];
    }

    const omega5 = Math.acos(arg);

    const dx2 = this.dx * this.dx;

    const denom = 1 - arg * arg;
    if (denom <= 0) {
      return [omega5, this.errorThreshold];
    }

    const f1 = 1 / denom;
    const invDen1Sq = 1 / (den1 * den1);
    const numeratorRatio = (x0 - xm4) / (den1 * den1);
    const r2 = invDen1Sq * (dx2 + dx2) + (numeratorRatio * numeratorRatio) * (dx2 + dx2);

    const product = f1 * r2;
    if (product < 0) {
      return [omega5, this.errorThreshold];
    }

    return [omega5, 0.5 * Math.sqrt(product)];
  }

  private calcModelParams(omega: number): [number, number, number, number, number] {
    const x0 = this.buffer[0];
    const xm1 = this.buffer[1];
    const xm2 = this.buffer[2];

    const halfW = omega / 2;
    const threeHalfW = 1.5 * omega;

    const sinHW = Math.sin(halfW);
    const cosHW = Math.cos(halfW);
    const sin3HW = Math.sin(threeHalfW);
    const cos3HW = Math.cos(threeHalfW);

    const d0 = sinHW * sinHW * cosHW * sin3HW - sinHW * sinHW * sinHW * cos3HW;

    if (Math.abs(d0) < 1e-15) {
      return [NaN, NaN, NaN, NaN, NaN];
    }

    const invD0 = 1 / d0;

    const dx0M1 = x0 - xm1;
    const dxm1M2 = xm1 - xm2;

    const c = invD0 * (dx0M1 * sinHW * sin3HW - dxm1M2 * sinHW * sinHW);
    const s = invD0 * (dxm1M2 * sinHW * cosHW - dx0M1 * sinHW * cos3HW);

    const amplitude = 0.5 * Math.sqrt(c * c + s * s);
    const phase = Math.atan2(s, c);
    const velocity = amplitude * omega * Math.cos(phase);
    const acceleration = -amplitude * omega * omega * Math.sin(phase);
    const dcLevel = x0 - s / 2;

    return [amplitude, phase, velocity, acceleration, dcLevel];
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.primed_;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.InstantaneousSineWavePeriod,
      this.mnemonic_,
      this.description_,
      [
        { mnemonic: this.mnemonic_ + ' period', description: this.description_ + ' Period' },
        { mnemonic: this.mnemonic_ + ' omega', description: this.description_ + ' Omega' },
        { mnemonic: this.mnemonic_ + ' velocity', description: this.description_ + ' Velocity' },
        { mnemonic: this.mnemonic_ + ' acceleration', description: this.description_ + ' Acceleration' },
        { mnemonic: this.mnemonic_ + ' amplitude', description: this.description_ + ' Amplitude' },
        { mnemonic: this.mnemonic_ + ' phase', description: this.description_ + ' Phase' },
        { mnemonic: this.mnemonic_ + ' dcLevel', description: this.description_ + ' DC Level' },
      ],
    );
  }

  /**
   * Updates the indicator given the next sample value.
   * Returns [period, omega, velocity, acceleration, amplitude, phase, dcLevel].
   */
  public update(sample: number): [number, number, number, number, number, number, number] {
    const smoothed = this.smoothing > 0 ? this.applyEma(sample) : sample;

    this.pushBuffer(smoothed);
    this.count++;

    if (this.count < 5) {
      return [NaN, NaN, NaN, NaN, NaN, NaN, NaN];
    }

    const [omega4, error4] = this.calcOmega4();
    const [omega5, error5] = this.calcOmega5();

    if (error4 >= this.errorThreshold && error5 >= this.errorThreshold) {
      return [NaN, NaN, NaN, NaN, NaN, NaN, NaN];
    }

    const omega = error5 < error4 ? omega5 : omega4;

    if (isNaN(omega) || omega <= 0) {
      return [NaN, NaN, NaN, NaN, NaN, NaN, NaN];
    }

    const period = (2 * Math.PI) / omega;
    if (period < this.minPeriod || period > this.maxPeriod) {
      return [NaN, NaN, NaN, NaN, NaN, NaN, NaN];
    }

    const [amplitude, phase, velocity, acceleration, dcLevel] = this.calcModelParams(omega);

    this.primed_ = true;

    return [period, omega, velocity, acceleration, amplitude, phase, dcLevel];
  }

  /** Updates the indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    const [period, omega, velocity, acceleration, amplitude, phase, dcLevel] = this.update(sample.value);

    const out: Scalar[] = [];
    for (const value of [period, omega, velocity, acceleration, amplitude, phase, dcLevel]) {
      const s = new Scalar();
      s.time = sample.time;
      s.value = value;
      out.push(s);
    }

    return out;
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
