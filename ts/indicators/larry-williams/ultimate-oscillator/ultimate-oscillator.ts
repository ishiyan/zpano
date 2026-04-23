import { buildMetadata } from '../../core/build-metadata';
import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { Indicator } from '../../core/indicator';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { UltimateOscillatorParams } from './params';

const defaultLength1 = 7;
const defaultLength2 = 14;
const defaultLength3 = 28;
const minLength = 2;
const weight1 = 4.0;
const weight2 = 2.0;
const weight3 = 1.0;
const totalWeight = weight1 + weight2 + weight3; // 7.0

/**
 * Larry Williams' Ultimate Oscillator indicator.
 *
 * The Ultimate Oscillator combines three different time periods into a single
 * oscillator that measures buying pressure relative to true range. The three
 * periods are weighted 4:2:1 (shortest:medium:longest).
 *
 * The indicator requires bar data (high, low, close) and does not use a single
 * bar component. For scalar, quote, and trade updates, the single value is used
 * as a substitute for high, low, and close.
 *
 * Reference:
 *
 * Williams, Larry (1985). "The Ultimate Oscillator". Technical Analysis of Stocks & Commodities.
 */
export class UltimateOscillator implements Indicator {

  private readonly p1: number;
  private readonly p2: number;
  private readonly p3: number;

  private previousClose = NaN;

  private readonly bpBuffer: Float64Array;
  private readonly trBuffer: Float64Array;
  private bufferIndex = 0;

  private bpSum1 = 0;
  private bpSum2 = 0;
  private bpSum3 = 0;
  private trSum1 = 0;
  private trSum2 = 0;
  private trSum3 = 0;

  private count = 0;
  private primed_ = false;
  private readonly mnemonic_: string;

  constructor(params?: UltimateOscillatorParams) {
    const l1 = params?.length1 ?? defaultLength1;
    const l2 = params?.length2 ?? defaultLength2;
    const l3 = params?.length3 ?? defaultLength3;

    if (l1 < minLength) {
      throw new Error(`length1 must be >= ${minLength}, got ${l1}`);
    }

    if (l2 < minLength) {
      throw new Error(`length2 must be >= ${minLength}, got ${l2}`);
    }

    if (l3 < minLength) {
      throw new Error(`length3 must be >= ${minLength}, got ${l3}`);
    }

    // Sort the three periods ascending.
    const sorted = [l1, l2, l3].sort((a, b) => a - b);
    this.p1 = sorted[0];
    this.p2 = sorted[1];
    this.p3 = sorted[2];

    this.bpBuffer = new Float64Array(this.p3);
    this.trBuffer = new Float64Array(this.p3);

    this.mnemonic_ = `ultosc(${l1}, ${l2}, ${l3})`;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.primed_;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    const description = 'Ultimate Oscillator';
    return buildMetadata(
      IndicatorIdentifier.UltimateOscillator,
      this.mnemonic_,
      `${description} ${this.mnemonic_}`,
      [
        { mnemonic: this.mnemonic_, description: `${description} ${this.mnemonic_}` },
      ],
    );
  }

  /** Updates the Ultimate Oscillator given the next bar's close, high, and low values. */
  public update(close: number, high: number, low: number): number {
    if (isNaN(close) || isNaN(high) || isNaN(low)) {
      return NaN;
    }

    // First bar: just store close, return NaN.
    if (isNaN(this.previousClose)) {
      this.previousClose = close;
      return NaN;
    }

    // Calculate buying pressure and true range.
    const trueLow = Math.min(low, this.previousClose);
    const bp = close - trueLow;

    let tr = high - low;
    const diffHigh = Math.abs(this.previousClose - high);
    if (diffHigh > tr) {
      tr = diffHigh;
    }

    const diffLow = Math.abs(this.previousClose - low);
    if (diffLow > tr) {
      tr = diffLow;
    }

    this.previousClose = close;

    this.count++;

    // Remove trailing values BEFORE storing the new value in the circular buffer,
    // because for p3 the old index equals bufferIndex (the buffer wraps exactly).
    if (this.count > this.p1) {
      const oldIndex = (this.bufferIndex - this.p1 + this.p3) % this.p3;
      this.bpSum1 -= this.bpBuffer[oldIndex];
      this.trSum1 -= this.trBuffer[oldIndex];
    }

    if (this.count > this.p2) {
      const oldIndex = (this.bufferIndex - this.p2 + this.p3) % this.p3;
      this.bpSum2 -= this.bpBuffer[oldIndex];
      this.trSum2 -= this.trBuffer[oldIndex];
    }

    if (this.count > this.p3) {
      const oldIndex = (this.bufferIndex - this.p3 + this.p3) % this.p3;
      this.bpSum3 -= this.bpBuffer[oldIndex];
      this.trSum3 -= this.trBuffer[oldIndex];
    }

    // Add to running sums.
    this.bpSum1 += bp;
    this.bpSum2 += bp;
    this.bpSum3 += bp;
    this.trSum1 += tr;
    this.trSum2 += tr;
    this.trSum3 += tr;

    // Store in circular buffer (after subtraction so p3 trailing reads the old value).
    this.bpBuffer[this.bufferIndex] = bp;
    this.trBuffer[this.bufferIndex] = tr;

    // Advance buffer index.
    this.bufferIndex = (this.bufferIndex + 1) % this.p3;

    // Need at least p3 values (the longest period) to produce output.
    if (this.count < this.p3) {
      return NaN;
    }

    this.primed_ = true;

    // Calculate output.
    let output = 0;

    if (this.trSum1 !== 0) {
      output += weight1 * (this.bpSum1 / this.trSum1);
    }

    if (this.trSum2 !== 0) {
      output += weight2 * (this.bpSum2 / this.trSum2);
    }

    if (this.trSum3 !== 0) {
      output += weight3 * (this.bpSum3 / this.trSum3);
    }

    return 100.0 * (output / totalWeight);
  }

  /** Updates the indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    const v = sample.value;
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = this.update(v, v, v);
    return [scalar];
  }

  /** Updates the indicator given the next bar sample. */
  public updateBar(sample: Bar): IndicatorOutput {
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = this.update(sample.close, sample.high, sample.low);
    return [scalar];
  }

  /** Updates the indicator given the next quote sample. */
  public updateQuote(sample: Quote): IndicatorOutput {
    const v = (sample.bid + sample.ask) / 2;
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = this.update(v, v, v);
    return [scalar];
  }

  /** Updates the indicator given the next trade sample. */
  public updateTrade(sample: Trade): IndicatorOutput {
    const v = sample.price;
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = this.update(v, v, v);
    return [scalar];
  }
}
