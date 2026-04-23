import { buildMetadata } from '../../core/build-metadata';
import { Bar } from '../../../entities/bar';
import { BarComponent, barComponentValue } from '../../../entities/bar-component';
import { Quote } from '../../../entities/quote';
import { DefaultQuoteComponent, quoteComponentValue } from '../../../entities/quote-component';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { DefaultTradeComponent, tradeComponentValue } from '../../../entities/trade-component';
import { Indicator } from '../../core/indicator';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { LinearRegressionParams } from './params';

const RAD_TO_DEG = 180.0 / Math.PI;

/** __Linear Regression__ (LINREG) computes the least-squares regression line over a
 * rolling window and produces five outputs per sample:
 *
 *  - Value:     b + m*(period-1)  — the regression value at the last bar of the window
 *  - Forecast:  b + m*period      — the time series forecast (one bar ahead)
 *  - Intercept: b                 — the y-intercept of the regression line
 *  - SlopeRad:  m                 — the slope of the regression line
 *  - SlopeDeg:  atan(m)*180/pi   — the slope expressed in degrees
 *
 * where y = b + m*x is the best-fit line (x = 0 ... period-1).
 *
 * The indicator is not primed during the first (period-1) updates.
 */
export class LinearRegression implements Indicator {
  private readonly length: number;
  private readonly lengthF: number;
  private readonly sumX: number;
  private readonly divisor: number;
  private readonly window: number[];
  private windowCount: number = 0;
  private primed: boolean = false;

  private curValue: number = Number.NaN;
  private curForecast: number = Number.NaN;
  private curIntercept: number = Number.NaN;
  private curSlopeRad: number = Number.NaN;
  private curSlopeDeg: number = Number.NaN;

  private readonly mnemonic: string;
  private readonly description: string;

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  /** Constructs an instance given input parameters. */
  public constructor(params: LinearRegressionParams) {
    const len = Math.floor(params.length);
    if (len < 2) {
      throw new Error('length should be greater than 1');
    }

    this.length = len;
    this.lengthF = len;
    this.window = new Array(len).fill(0);

    const n = len;
    this.sumX = n * (n - 1) * 0.5;
    const sumXSqr = n * (n - 1) * (2 * n - 1) / 6;
    this.divisor = this.sumX * this.sumX - n * sumXSqr;

    const bc = params.barComponent ?? BarComponent.Close;
    const qc = params.quoteComponent ?? DefaultQuoteComponent;
    const tc = params.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    const cm = componentTripleMnemonic(
      params.barComponent,
      params.quoteComponent,
      params.tradeComponent,
    );
    this.mnemonic = `linreg(${len}${cm})`;
    this.description = 'Linear Regression ' + this.mnemonic;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean { return this.primed; }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.LinearRegression,
      this.mnemonic,
      this.description,
      [
        { mnemonic: this.mnemonic, description: this.description + ' value' },
        { mnemonic: this.mnemonic, description: this.description + ' forecast' },
        { mnemonic: this.mnemonic, description: this.description + ' intercept' },
        { mnemonic: this.mnemonic, description: this.description + ' slope' },
        { mnemonic: this.mnemonic, description: this.description + ' angle' },
      ],
    );
  }

  /** Updates the indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    return this.updateEntity(sample.time, sample.value);
  }

  /** Updates the indicator given the next bar sample. */
  public updateBar(sample: Bar): IndicatorOutput {
    return this.updateEntity(sample.time, this.barComponentFunc(sample));
  }

  /** Updates the indicator given the next quote sample. */
  public updateQuote(sample: Quote): IndicatorOutput {
    return this.updateEntity(sample.time, this.quoteComponentFunc(sample));
  }

  /** Updates the indicator given the next trade sample. */
  public updateTrade(sample: Trade): IndicatorOutput {
    return this.updateEntity(sample.time, this.tradeComponentFunc(sample));
  }

  /** Updates the indicator given the next sample and returns the Value output. */
  public update(sample: number): number {
    if (Number.isNaN(sample)) {
      return Number.NaN;
    }

    if (this.primed) {
      this.calculate(sample);
      return this.curValue;
    }

    this.window[this.windowCount] = sample;
    this.windowCount++;

    if (this.windowCount === this.length) {
      this.primed = true;
      this.computeFromWindow();
      return this.curValue;
    }

    return Number.NaN;
  }

  private calculate(sample: number): void {
    for (let i = 0; i < this.length - 1; i++) {
      this.window[i] = this.window[i + 1];
    }

    this.window[this.length - 1] = sample;
    this.computeFromWindow();
  }

  private computeFromWindow(): void {
    let sumXY = 0;
    let sumY = 0;

    for (let i = this.length; i > 0; i--) {
      const v = this.window[this.length - i];
      sumY += v;
      sumXY += (i - 1) * v;
    }

    const m = (this.lengthF * sumXY - this.sumX * sumY) / this.divisor;
    const b = (sumY - m * this.sumX) / this.lengthF;

    this.curSlopeRad = m;
    this.curSlopeDeg = Math.atan(m) * RAD_TO_DEG;
    this.curIntercept = b;
    this.curValue = b + m * (this.lengthF - 1);
    this.curForecast = b + m * this.lengthF;
  }

  private updateEntity(time: Date, sample: number): IndicatorOutput {
    const value = this.update(sample);

    const sValue = new Scalar();
    sValue.time = time;

    const sForecast = new Scalar();
    sForecast.time = time;

    const sIntercept = new Scalar();
    sIntercept.time = time;

    const sSlopeRad = new Scalar();
    sSlopeRad.time = time;

    const sSlopeDeg = new Scalar();
    sSlopeDeg.time = time;

    if (Number.isNaN(value)) {
      sValue.value = Number.NaN;
      sForecast.value = Number.NaN;
      sIntercept.value = Number.NaN;
      sSlopeRad.value = Number.NaN;
      sSlopeDeg.value = Number.NaN;
    } else {
      sValue.value = this.curValue;
      sForecast.value = this.curForecast;
      sIntercept.value = this.curIntercept;
      sSlopeRad.value = this.curSlopeRad;
      sSlopeDeg.value = this.curSlopeDeg;
    }

    return [sValue, sForecast, sIntercept, sSlopeRad, sSlopeDeg];
  }
}
