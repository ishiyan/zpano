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
import { AdaptiveExponentialMovingAverageParams } from './params';

const ISWP_MIN_PERIOD = 4.0;
const ISWP_MAX_PERIOD = 50.0;
const ISWP_ERROR_THRESHOLD = 20.0;
const ISWP_DX = 0.01;

/**
 * Embedded Instantaneous Sine Wave Period omega estimator (omega-only reduction).
 *
 * Estimates the dominant circular frequency omega of price data by modeling it
 * locally as a single sine wave, combining a 4-point and a 5-point method and
 * selecting the one with the lower estimation error. Inlined so the indicator is a
 * standalone porting unit. Do NOT change its numerics.
 */
class Iswp {
  private readonly smoothing: number;
  private readonly emaAlpha: number;
  private emaValue = 0;
  private emaPrimed = false;
  private readonly buffer: number[] = [0, 0, 0, 0, 0];
  private count = 0;

  constructor(smoothing: number) {
    this.smoothing = smoothing;
    this.emaAlpha = smoothing > 0 ? 2 / (smoothing + 1) : 1;
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
      return [NaN, ISWP_ERROR_THRESHOLD];
    }

    const ratio = (x0 - xm3) / den;

    const sqrtArg = 3 - ratio;
    if (sqrtArg < 0) {
      return [NaN, ISWP_ERROR_THRESHOLD];
    }

    const arg = 0.5 * Math.sqrt(sqrtArg);
    if (arg > 1) {
      return [NaN, ISWP_ERROR_THRESHOLD];
    }

    const omega4 = 2 * Math.asin(arg);

    const dx2 = ISWP_DX * ISWP_DX;

    const denom1 = 1 - 0.25 * sqrtArg;
    if (denom1 <= 0 || sqrtArg === 0) {
      return [omega4, ISWP_ERROR_THRESHOLD];
    }

    const f1 = 1 / (denom1 * sqrtArg);
    const invDen2 = 1 / (den * den);
    const q2 = invDen2 * (dx2 + dx2) + (ratio * ratio) * invDen2 * (dx2 + dx2);

    const product = f1 * q2;
    if (product < 0) {
      return [omega4, ISWP_ERROR_THRESHOLD];
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
      return [NaN, ISWP_ERROR_THRESHOLD];
    }

    const arg = 0.5 * (x0 - xm4) / den1;
    if (Math.abs(arg) > 1) {
      return [NaN, ISWP_ERROR_THRESHOLD];
    }

    const omega5 = Math.acos(arg);

    const dx2 = ISWP_DX * ISWP_DX;

    const denom = 1 - arg * arg;
    if (denom <= 0) {
      return [omega5, ISWP_ERROR_THRESHOLD];
    }

    const f1 = 1 / denom;
    const invDen1Sq = 1 / (den1 * den1);
    const numeratorRatio = (x0 - xm4) / (den1 * den1);
    const r2 = invDen1Sq * (dx2 + dx2) + (numeratorRatio * numeratorRatio) * (dx2 + dx2);

    const product = f1 * r2;
    if (product < 0) {
      return [omega5, ISWP_ERROR_THRESHOLD];
    }

    return [omega5, 0.5 * Math.sqrt(product)];
  }

  public update(price: number): number {
    const smoothed = this.smoothing > 0 ? this.applyEma(price) : price;

    this.pushBuffer(smoothed);
    this.count++;

    if (this.count < 5) {
      return NaN;
    }

    const [omega4, error4] = this.calcOmega4();
    const [omega5, error5] = this.calcOmega5();

    if (error4 >= ISWP_ERROR_THRESHOLD && error5 >= ISWP_ERROR_THRESHOLD) {
      return NaN;
    }

    const omega = error5 < error4 ? omega5 : omega4;

    if (isNaN(omega) || omega <= 0) {
      return NaN;
    }

    const period = (2 * Math.PI) / omega;
    if (period < ISWP_MIN_PERIOD || period > ISWP_MAX_PERIOD) {
      return NaN;
    }

    return omega;
  }
}

/** Function to calculate mnemonic of an __AdaptiveExponentialMovingAverage__ indicator. */
export const adaptiveExponentialMovingAverageMnemonic = (
  alphaMax: number, alphaMin: number, omega0: number, smoothing: number,
  barComponent?: BarComponent, quoteComponent?: QuoteComponent, tradeComponent?: TradeComponent,
): string => {
  const cm = componentTripleMnemonic(barComponent, quoteComponent, tradeComponent);

  return `aema(${alphaMax.toFixed(2)},${alphaMin.toFixed(2)},${omega0.toFixed(2)},${smoothing}${cm})`;
};

/**
 * AdaptiveExponentialMovingAverage is Don Mak's Adaptive Exponential Moving Average (AEMA).
 *
 * It is an EMA with a time-varying smoothing factor alpha that adapts based on the
 * instantaneous frequency of the price data, estimated by an embedded ISWP.
 *
 * The indicator produces three outputs:
 *   - Value: the adaptively smoothed price (never NaN);
 *   - Omega: the instantaneous frequency estimate (may be NaN);
 *   - Alpha: the smoothing factor used for this bar.
 *
 * Reference:
 *
 * Mak, D.K. (2006). Mathematical Techniques in Financial Market Trading.
 */
export class AdaptiveExponentialMovingAverage implements Indicator {

  private readonly alphaMax: number;
  private readonly alphaMin: number;
  private readonly omega0: number;
  private readonly a: number;
  private readonly b: number;

  private readonly iswp: Iswp;

  private emaValue = 0;
  private initialized = false;
  private primed_ = false;

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  private readonly mnemonic_: string;
  private readonly description_: string;

  constructor(params?: AdaptiveExponentialMovingAverageParams) {
    const p = params ?? {};

    const alphaMax = p.alphaMax ?? 0.5;
    const alphaMin = p.alphaMin ?? 0.05;
    const omega0 = p.omega0 ?? 1.0;
    const smoothing = Math.floor(p.smoothing ?? 3);

    if (!(alphaMin > 0 && alphaMin < alphaMax && alphaMax <= 1)) {
      throw new Error('need 0 < alphaMin < alphaMax <= 1');
    }

    if (!(omega0 > 0 && omega0 < Math.PI)) {
      throw new Error('need 0 < omega0 < pi');
    }

    if (smoothing < 0) {
      throw new Error('smoothing should be >= 0');
    }

    const bc = p.barComponent ?? DefaultBarComponent;
    const qc = p.quoteComponent ?? DefaultQuoteComponent;
    const tc = p.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    this.alphaMax = alphaMax;
    this.alphaMin = alphaMin;
    this.omega0 = omega0;

    this.a = (alphaMax - alphaMin) * omega0 * Math.PI / (Math.PI - omega0);
    this.b = alphaMin - this.a / Math.PI;

    this.iswp = new Iswp(smoothing);

    this.mnemonic_ = adaptiveExponentialMovingAverageMnemonic(
      alphaMax, alphaMin, omega0, smoothing,
      p.barComponent, p.quoteComponent, p.tradeComponent,
    );
    this.description_ = 'Adaptive Exponential Moving Average ' + this.mnemonic_;
  }

  private computeAlpha(omega: number): number {
    if (isNaN(omega)) {
      return this.alphaMin;
    }
    if (omega <= this.omega0) {
      return this.alphaMax;
    }
    if (omega >= Math.PI) {
      return this.alphaMin;
    }

    const alpha = this.a / omega + this.b;
    if (alpha > this.alphaMax) {
      return this.alphaMax;
    }
    if (alpha < this.alphaMin) {
      return this.alphaMin;
    }
    return alpha;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.primed_;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.AdaptiveExponentialMovingAverage,
      this.mnemonic_,
      this.description_,
      [
        { mnemonic: this.mnemonic_ + ' value', description: this.description_ + ' Value' },
        { mnemonic: this.mnemonic_ + ' omega', description: this.description_ + ' Omega' },
        { mnemonic: this.mnemonic_ + ' alpha', description: this.description_ + ' Alpha' },
      ],
    );
  }

  /**
   * Updates the indicator given the next sample value.
   * Returns [value, omega, alpha].
   */
  public update(sample: number): [number, number, number] {
    const omega = this.iswp.update(sample);
    const alpha = this.computeAlpha(omega);

    if (!this.initialized) {
      this.emaValue = sample;
      this.initialized = true;
    } else {
      this.emaValue = alpha * sample + (1 - alpha) * this.emaValue;
    }

    if (!isNaN(omega)) {
      this.primed_ = true;
    }

    return [this.emaValue, omega, alpha];
  }

  /** Updates the indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    const [value, omega, alpha] = this.update(sample.value);

    const s0 = new Scalar(); s0.time = sample.time; s0.value = value;
    const s1 = new Scalar(); s1.time = sample.time; s1.value = omega;
    const s2 = new Scalar(); s2.time = sample.time; s2.value = alpha;

    return [s0, s1, s2];
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
