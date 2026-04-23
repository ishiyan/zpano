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
import { CyberCycleLengthParams } from './length-params';
import { CyberCycleSmoothingFactorParams } from './smoothing-factor-params';

const guardLength = (object: any): object is CyberCycleLengthParams => 'length' in object;

/** Function to calculate mnemonic of a __CyberCycle__ indicator. */
export const cyberCycleMnemonic =
  (params: CyberCycleLengthParams | CyberCycleSmoothingFactorParams): string => {
  const epsilon = 0.00000001;
  let length: number;
  if (guardLength(params)) {
    length = Math.floor(params.length);
  } else {
    const alpha = params.smoothingFactor;
    if (alpha < epsilon) {
      length = Number.MAX_SAFE_INTEGER;
    } else {
      length = Math.round(2 / alpha) - 1;
    }
  }

  const cm = componentTripleMnemonic(
    params.barComponent ?? BarComponent.Median,
    params.quoteComponent,
    params.tradeComponent,
  );

  return `cc(${length}${cm})`;
};

/** __Cyber Cycle__ (Ehler's Cyber Cycle, CC) is described in Ehler's book
 * "Cybernetic Analysis for Stocks and Futures" (2004):
 *
 *	H(z) = ((1-α/2)²(1 - 2z⁻¹ + z⁻²)) / (1 - 2(1-α)z⁻¹ + (1-α)²z⁻²)
 *
 * which is a complementary high-pass filter found by subtracting the
 * Instantaneous Trend Line low-pass filter from unity.
 *
 * The Cyber Cycle has zero lag and retains the relative cycle amplitude.
 *
 * The indicator has two outputs: the cycle value and a signal line which
 * is an exponential moving average of the cycle value.
 *
 * Reference:
 *
 *	Ehlers, John F. (2004). Cybernetic Analysis for Stocks and Futures. Wiley.
 */
export class CyberCycle implements Indicator {
  private readonly lengthValue: number;
  private readonly smoothingFactorValue: number;
  private readonly signalLagValue: number;
  private readonly coeff1: number;
  private readonly coeff2: number;
  private readonly coeff3: number;
  private readonly coeff4: number;
  private readonly coeff5: number;
  private count: number = 0;
  private previousSample1: number = 0;
  private previousSample2: number = 0;
  private previousSample3: number = 0;
  private smoothed: number = 0;
  private previousSmoothed1: number = 0;
  private previousSmoothed2: number = 0;
  private value: number = Number.NaN;
  private previousValue1: number = 0;
  private previousValue2: number = 0;
  private signal: number = Number.NaN;
  private primed: boolean = false;
  private readonly mnemonicStr: string;
  private readonly descriptionStr: string;
  private readonly mnemonicSignal: string;
  private readonly descriptionSignal: string;

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  /**
   * Constructs an instance given a length or a smoothing factor, along with a signal lag.
   */
  public constructor(params: CyberCycleLengthParams | CyberCycleSmoothingFactorParams) {
    const epsilon = 0.00000001;
    let length: number;
    let alpha: number;

    if (guardLength(params)) {
      length = Math.floor(params.length);
      if (length < 1) {
        throw new Error('length should be a positive integer');
      }

      alpha = 2 / (1 + length);
    } else {
      alpha = params.smoothingFactor;
      if (alpha < 0 || alpha > 1) {
        throw new Error('smoothing factor should be in range [0, 1]');
      }

      if (alpha < epsilon) {
        length = Number.MAX_SAFE_INTEGER;
      } else {
        length = Math.round(2 / alpha) - 1;
      }
    }

    const signalLag = Math.floor(params.signalLag);
    if (signalLag < 1) {
      throw new Error('signal lag should be a positive integer');
    }

    this.lengthValue = length;
    this.smoothingFactorValue = alpha;
    this.signalLagValue = signalLag;

    // Calculate coefficients.
    let x = 1 - alpha / 2;
    this.coeff1 = x * x;

    x = 1 - alpha;
    this.coeff2 = 2 * x;
    this.coeff3 = -(x * x);

    x = 1 / (1 + signalLag);
    this.coeff4 = x;
    this.coeff5 = 1 - x;

    // Resolve component defaults and create component functions.
    // CyberCycle default bar component is Median, not Close.
    const bc = params.barComponent ?? BarComponent.Median;
    const qc = params.quoteComponent ?? DefaultQuoteComponent;
    const tc = params.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    // Build mnemonics.
    const cm = componentTripleMnemonic(
      params.barComponent ?? BarComponent.Median,
      params.quoteComponent,
      params.tradeComponent,
    );

    this.mnemonicStr = `cc(${length}${cm})`;
    this.mnemonicSignal = `ccSignal(${length}${cm})`;

    const descr = 'Cyber Cycle ';
    const descrSignal = 'Cyber Cycle signal ';
    this.descriptionStr = descr + this.mnemonicStr;
    this.descriptionSignal = descrSignal + this.mnemonicSignal;
  }

  /** Indicates whether an indicator is primed. */
  public isPrimed(): boolean { return this.primed; }

  /** Describes a requested output data of an indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.CyberCycle,
      this.mnemonicStr,
      this.descriptionStr,
      [
        { mnemonic: this.mnemonicStr, description: this.descriptionStr },
        { mnemonic: this.mnemonicSignal, description: this.descriptionSignal },
      ],
    );
  }

  /** Updates an indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    return this.updateEntity(sample.time, sample.value);
  }

  /** Updates an indicator given the next bar sample. */
  public updateBar(sample: Bar): IndicatorOutput {
    return this.updateEntity(sample.time, this.barComponentFunc(sample));
  }

  /** Updates an indicator given the next quote sample. */
  public updateQuote(sample: Quote): IndicatorOutput {
    return this.updateEntity(sample.time, this.quoteComponentFunc(sample));
  }

  /** Updates an indicator given the next trade sample. */
  public updateTrade(sample: Trade): IndicatorOutput {
    return this.updateEntity(sample.time, this.tradeComponentFunc(sample));
  }

  /** Updates the value of the cyber cycle given the next sample. */
  public update(sample: number): number {
    if (Number.isNaN(sample)) {
      return Number.NaN;
    }

    if (this.primed) {
      this.previousSmoothed2 = this.previousSmoothed1;
      this.previousSmoothed1 = this.smoothed;
      this.smoothed = (sample + 2 * this.previousSample1 + 2 * this.previousSample2 + this.previousSample3) / 6;

      this.previousValue2 = this.previousValue1;
      this.previousValue1 = this.value;
      this.value = this.coeff1 * (this.smoothed - 2 * this.previousSmoothed1 + this.previousSmoothed2)
        + this.coeff2 * this.previousValue1 + this.coeff3 * this.previousValue2;

      this.signal = this.coeff4 * this.value + this.coeff5 * this.signal;

      this.previousSample3 = this.previousSample2;
      this.previousSample2 = this.previousSample1;
      this.previousSample1 = sample;

      return this.value;
    }

    this.count++;

    switch (this.count) {
      case 1:
        this.previousSample3 = sample;
        return Number.NaN;
      case 2:
        this.previousSample2 = sample;
        return Number.NaN;
      case 3:
        this.signal = this.coeff4 * (sample - 2 * this.previousSample2 + this.previousSample3) / 4;
        this.previousSample1 = sample;
        return Number.NaN;
      case 4:
        this.previousSmoothed2 = (sample + 2 * this.previousSample1
          + 2 * this.previousSample2 + this.previousSample3) / 6;
        this.signal = this.coeff4 * (sample - 2 * this.previousSample1 + this.previousSample2) / 4
          + this.coeff5 * this.signal;

        this.previousSample3 = this.previousSample2;
        this.previousSample2 = this.previousSample1;
        this.previousSample1 = sample;
        return Number.NaN;
      case 5:
        this.previousSmoothed1 = (sample + 2 * this.previousSample1
          + 2 * this.previousSample2 + this.previousSample3) / 6;
        this.signal = this.coeff4 * (sample - 2 * this.previousSample1 + this.previousSample2) / 4
          + this.coeff5 * this.signal;

        this.previousSample3 = this.previousSample2;
        this.previousSample2 = this.previousSample1;
        this.previousSample1 = sample;
        return Number.NaN;
      case 6:
        this.smoothed = (sample + 2 * this.previousSample1
          + 2 * this.previousSample2 + this.previousSample3) / 6;
        this.previousValue2 = (sample - 2 * this.previousSample1 + this.previousSample2) / 4;
        this.signal = this.coeff4 * this.previousValue2 + this.coeff5 * this.signal;

        this.previousSample3 = this.previousSample2;
        this.previousSample2 = this.previousSample1;
        this.previousSample1 = sample;
        return Number.NaN;
      case 7:
        this.previousSmoothed2 = this.previousSmoothed1;
        this.previousSmoothed1 = this.smoothed;
        this.smoothed = (sample + 2 * this.previousSample1
          + 2 * this.previousSample2 + this.previousSample3) / 6;
        this.previousValue1 = (sample - 2 * this.previousSample1 + this.previousSample2) / 4;
        this.signal = this.coeff4 * this.previousValue1 + this.coeff5 * this.signal;

        this.previousSample3 = this.previousSample2;
        this.previousSample2 = this.previousSample1;
        this.previousSample1 = sample;
        return Number.NaN;
      case 8:
        this.previousSmoothed2 = this.previousSmoothed1;
        this.previousSmoothed1 = this.smoothed;
        this.smoothed = (sample + 2 * this.previousSample1
          + 2 * this.previousSample2 + this.previousSample3) / 6;

        this.value = this.coeff1 * (this.smoothed - 2 * this.previousSmoothed1 + this.previousSmoothed2)
          + this.coeff2 * this.previousValue1 + this.coeff3 * this.previousValue2;

        this.signal = this.coeff4 * this.value + this.coeff5 * this.signal;

        this.previousSample3 = this.previousSample2;
        this.previousSample2 = this.previousSample1;
        this.previousSample1 = sample;
        this.primed = true;

        return this.value;
      default:
        return Number.NaN;
    }
  }

  private updateEntity(time: Date, sample: number): IndicatorOutput {
    const v = this.update(sample);

    let signal = this.signal;
    if (Number.isNaN(v)) {
      signal = Number.NaN;
    }

    const scalarValue = new Scalar();
    scalarValue.time = time;
    scalarValue.value = v;

    const scalarSignal = new Scalar();
    scalarSignal.time = time;
    scalarSignal.value = signal;

    return [scalarValue, scalarSignal];
  }
}
