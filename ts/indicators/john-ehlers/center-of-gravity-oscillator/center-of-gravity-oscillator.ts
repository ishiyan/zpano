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
import { IndicatorType } from '../../core/indicator-type';
import { OutputType } from '../../core/outputs/output-type';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { CenterOfGravityOscillatorParams } from './center-of-gravity-oscillator-params';
import { CenterOfGravityOscillatorOutput } from './center-of-gravity-oscillator-output';

/** __Center of Gravity Oscillator__ (Ehler's Center of Gravity oscillator, _COG_) is
 * described in Ehler's book "Cybernetic Analysis for Stocks and Futures" (2004).
 *
 * The center of gravity in a FIR filter is the position of the average price
 * within the filter window length:
 *
 *	CGi = Sigma((i+1) * Pricei) / Sigma(Pricei), where i = 0...l-1, l being a window size.
 *
 * The Center of Gravity oscillator has essentially zero lag and retains the
 * relative cycle amplitude.
 *
 * It moves toward the most recent bar (decreases) when prices rise and moves
 * away from the most recent bar (increases) when prices fall; thus moving
 * exactly opposite to the price direction.
 *
 * The indicator has two outputs: the oscillator value and a trigger line which
 * is the previous value of the oscillator.
 *
 * Reference:
 *
 *	Ehlers, John F. (2004). Cybernetic Analysis for Stocks and Futures. Wiley.
 */
export class CenterOfGravityOscillator implements Indicator {
  private readonly length: number;
  private readonly lengthMinOne: number;
  private windowCount: number = 0;
  private value: number = Number.NaN;
  private valuePrevious: number = Number.NaN;
  private denominatorSum: number = 0;
  private readonly window: number[];
  private mnemonic: string;
  private description: string;
  private mnemonicTrig: string;
  private descriptionTrig: string;
  private primed: boolean = false;

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  /**
   * Constructs an instance given input parameters.
   */
  public constructor(params: CenterOfGravityOscillatorParams) {
    const len = Math.floor(params.length);
    if (len < 1) {
      throw new Error('length should be a positive integer');
    }

    this.length = len;
    this.lengthMinOne = len - 1;
    this.window = new Array(len).fill(0);

    // Resolve component defaults and create component functions.
    // CoG default bar component is Median, not Close.
    const bc = params.barComponent ?? BarComponent.Median;
    const qc = params.quoteComponent ?? DefaultQuoteComponent;
    const tc = params.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    // Build mnemonics matching Go format: cog(%d%s)
    // Since CoG defaults to Median (not the framework default Close),
    // we always pass BarComponent.Median to componentTripleMnemonic when
    // params.barComponent is undefined, so the mnemonic includes ', hl/2'.
    const cm = componentTripleMnemonic(
      params.barComponent ?? BarComponent.Median,
      params.quoteComponent,
      params.tradeComponent,
    );
    this.mnemonic = `cog(${len}${cm})`;
    this.mnemonicTrig = `cogTrig(${len}${cm})`;

    const descr = 'Center of Gravity oscillator ';
    const descrTrig = 'Center of Gravity trigger ';
    this.description = descr + this.mnemonic;
    this.descriptionTrig = descrTrig + this.mnemonicTrig;
  }

  /** Indicates whether an indicator is primed. */
  public isPrimed(): boolean { return this.primed; }

  /** Describes a requested output data of an indicator. */
  public metadata(): IndicatorMetadata {
    return {
      type: IndicatorType.CenterOfGravityOscillator,
      mnemonic: this.mnemonic,
      description: this.description,
      outputs: [
        {
          kind: CenterOfGravityOscillatorOutput.Value,
          type: OutputType.Scalar,
          mnemonic: this.mnemonic,
          description: this.description,
        },
        {
          kind: CenterOfGravityOscillatorOutput.Trigger,
          type: OutputType.Scalar,
          mnemonic: this.mnemonicTrig,
          description: this.descriptionTrig,
        },
      ],
    };
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

  /** Updates the value of the center of gravity oscillator given the next sample. */
  public update(sample: number): number {
    if (Number.isNaN(sample)) {
      return Number.NaN;
    }

    if (this.primed) {
      this.valuePrevious = this.value;
      this.value = this.calculate(sample);

      return this.value;
    }

    // Not primed.
    if (this.length > this.windowCount) {
      this.denominatorSum += sample;
      this.window[this.windowCount] = sample;

      if (this.lengthMinOne === this.windowCount) {
        let sum = 0;
        if (Math.abs(this.denominatorSum) > Number.MIN_VALUE) {
          for (let i = 0; i < this.length; i++) {
            sum += (1 + i) * this.window[i];
          }

          sum /= this.denominatorSum;
        }

        this.valuePrevious = sum;
      }
    } else {
      this.value = this.calculate(sample);
      this.primed = true;

      this.windowCount++;

      return this.value;
    }

    this.windowCount++;

    return Number.NaN;
  }

  private calculate(sample: number): number {
    this.denominatorSum += sample - this.window[0];

    for (let i = 0; i < this.lengthMinOne; i++) {
      this.window[i] = this.window[i + 1];
    }

    this.window[this.lengthMinOne] = sample;

    let sum = 0;
    if (Math.abs(this.denominatorSum) > Number.MIN_VALUE) {
      for (let i = 0; i < this.length; i++) {
        sum += (1 + i) * this.window[i];
      }

      sum /= this.denominatorSum;
    }

    return sum;
  }

  private updateEntity(time: Date, sample: number): IndicatorOutput {
    const cog = this.update(sample);

    let trig = this.valuePrevious;
    if (Number.isNaN(cog)) {
      trig = Number.NaN;
    }

    const scalarCog = new Scalar();
    scalarCog.time = time;
    scalarCog.value = cog;

    const scalarTrig = new Scalar();
    scalarTrig.time = time;
    scalarTrig.value = trig;

    return [scalarCog, scalarTrig];
  }
}
