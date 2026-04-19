import { Bar } from '../../../entities/bar';
import { Scalar } from '../../../entities/scalar';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorType } from '../../core/indicator-type';
import { LineIndicator } from '../../core/line-indicator';
import { OutputType } from '../../core/outputs/output-type';
import { ParabolicStopAndReverseOutput } from './parabolic-stop-and-reverse-output';
import { ParabolicStopAndReverseParams } from './parabolic-stop-and-reverse-params';

const defaultAccelerationInit = 0.02;
const defaultAccelerationStep = 0.02;
const defaultAccelerationMax = 0.20;

/**
 * Welles Wilder's Parabolic Stop And Reverse (SAR) indicator.
 *
 * The Parabolic SAR provides potential entry and exit points. It places dots above or below
 * the price to indicate the direction of the trend. When the dots are below the price, it
 * signals a long (upward) trend; when above, it signals a short (downward) trend.
 *
 * This is the "extended" version (SAREXT) which supports separate acceleration factor
 * parameters for long and short directions, an optional start value to force the initial
 * direction, and a percent offset on reversal. The output is signed: positive values
 * indicate long positions, negative values indicate short positions.
 *
 * Algorithm overview (from Welles Wilder / TA-Lib):
 *
 * The implementation of SAR has been somewhat open to interpretation since Wilder
 * (the original author) did not define a precise algorithm on how to bootstrap it.
 *
 * Initial trade direction:
 *   - If startValue == 0 (auto): the direction is determined by comparing +DM and -DM
 *     between the first and second bars. If -DM > +DM the initial direction is short,
 *     otherwise long. Ties default to long.
 *   - If startValue > 0: force long at the specified SAR value.
 *   - If startValue < 0: force short at abs(startValue) as the initial SAR value.
 *
 * Initial extreme point and SAR:
 *   - For auto mode: the first bar's high/low is used as the initial SAR, and the second
 *     bar's high (long) or low (short) is the initial extreme point. This is the same
 *     approach used by Metastock.
 *   - For forced mode: the SAR is set to the specified start value.
 *
 * On each subsequent bar the SAR is updated by the acceleration factor (AF) times the
 * difference between the extreme point (EP) and the current SAR. The AF starts at the
 * initial value and increases by the step value each time a new EP is reached, up to a
 * maximum. When a reversal occurs (price penetrates the SAR), the position flips, the
 * SAR is reset to the EP, and the AF is reset.
 *
 * Reference:
 *
 * Wilder, J. Welles. "New Concepts in Technical Trading Systems", 1978.
 */
export class ParabolicStopAndReverse extends LineIndicator {
  // Resolved parameters.
  private readonly startValue: number;
  private readonly offsetOnReverse: number;
  private readonly afInitLong: number;
  private readonly afStepLong: number;
  private readonly afMaxLong: number;
  private readonly afInitShort: number;
  private readonly afStepShort: number;
  private readonly afMaxShort: number;

  // State.
  private count: number;
  private isLong: boolean;
  private sar: number;
  private ep: number;
  private afLong: number;
  private afShort: number;
  private prevHigh: number;
  private prevLow: number;
  private newHigh: number;
  private newLow: number;

  /** Constructs an instance given the parameters. */
  public constructor(params?: ParabolicStopAndReverseParams) {
    super();

    const p = params ?? {};

    let afInitLong = p.accelerationInitLong ?? 0;
    let afStepLong = p.accelerationLong ?? 0;
    let afMaxLong = p.accelerationMaxLong ?? 0;
    let afInitShort = p.accelerationInitShort ?? 0;
    let afStepShort = p.accelerationShort ?? 0;
    let afMaxShort = p.accelerationMaxShort ?? 0;

    // Apply defaults for zero values.
    if (afInitLong === 0) afInitLong = defaultAccelerationInit;
    if (afStepLong === 0) afStepLong = defaultAccelerationStep;
    if (afMaxLong === 0) afMaxLong = defaultAccelerationMax;
    if (afInitShort === 0) afInitShort = defaultAccelerationInit;
    if (afStepShort === 0) afStepShort = defaultAccelerationStep;
    if (afMaxShort === 0) afMaxShort = defaultAccelerationMax;

    // Validate.
    if (afInitLong < 0 || afStepLong < 0 || afMaxLong < 0) {
      throw new Error('long acceleration factors must be non-negative');
    }
    if (afInitShort < 0 || afStepShort < 0 || afMaxShort < 0) {
      throw new Error('short acceleration factors must be non-negative');
    }
    if ((p.offsetOnReverse ?? 0) < 0) {
      throw new Error('offset on reverse must be non-negative');
    }

    // Clamp.
    if (afInitLong > afMaxLong) afInitLong = afMaxLong;
    if (afStepLong > afMaxLong) afStepLong = afMaxLong;
    if (afInitShort > afMaxShort) afInitShort = afMaxShort;
    if (afStepShort > afMaxShort) afStepShort = afMaxShort;

    this.startValue = p.startValue ?? 0;
    this.offsetOnReverse = p.offsetOnReverse ?? 0;
    this.afInitLong = afInitLong;
    this.afStepLong = afStepLong;
    this.afMaxLong = afMaxLong;
    this.afInitShort = afInitShort;
    this.afStepShort = afStepShort;
    this.afMaxShort = afMaxShort;

    this.mnemonic = 'sar()';
    this.description = 'Parabolic Stop And Reverse ' + this.mnemonic;

    this.count = 0;
    this.isLong = true;
    this.sar = 0;
    this.ep = 0;
    this.afLong = afInitLong;
    this.afShort = afInitShort;
    this.prevHigh = 0;
    this.prevLow = 0;
    this.newHigh = 0;
    this.newLow = 0;
    this.primed = false;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return {
      type: IndicatorType.ParabolicStopAndReverse,
      mnemonic: this.mnemonic,
      description: this.description,
      outputs: [{
        kind: ParabolicStopAndReverseOutput.ParabolicStopAndReverseValue,
        type: OutputType.Scalar,
        mnemonic: this.mnemonic,
        description: this.description,
      }],
    };
  }

  /** Updates the indicator with the given scalar sample. */
  public update(sample: number): number {
    if (Number.isNaN(sample)) {
      return Number.NaN;
    }
    return this.updateHL(sample, sample);
  }

  /** Updates the indicator with the given high and low values. */
  public updateHL(high: number, low: number): number {
    if (Number.isNaN(high) || Number.isNaN(low)) {
      return Number.NaN;
    }

    this.count++;

    // First bar: store high/low, no output yet.
    if (this.count === 1) {
      this.newHigh = high;
      this.newLow = low;
      return Number.NaN;
    }

    // Second bar: initialize SAR, EP, and direction.
    if (this.count === 2) {
      const prevHigh = this.newHigh;
      const prevLow = this.newLow;

      if (this.startValue === 0) {
        // Auto-detect direction using MINUS_DM logic.
        let minusDM = prevLow - low;
        let plusDM = high - prevHigh;
        if (minusDM < 0) minusDM = 0;
        if (plusDM < 0) plusDM = 0;

        this.isLong = minusDM <= plusDM;

        if (this.isLong) {
          this.ep = high;
          this.sar = prevLow;
        } else {
          this.ep = low;
          this.sar = prevHigh;
        }
      } else if (this.startValue > 0) {
        this.isLong = true;
        this.ep = high;
        this.sar = this.startValue;
      } else {
        this.isLong = false;
        this.ep = low;
        this.sar = Math.abs(this.startValue);
      }

      this.newHigh = high;
      this.newLow = low;
      this.primed = true;
    }

    // Main SAR calculation (bars 2+).
    if (this.count >= 2) {
      this.prevLow = this.newLow;
      this.prevHigh = this.newHigh;
      this.newLow = low;
      this.newHigh = high;

      if (this.count === 2) {
        // On the second call, match TaLib's "cheat" for first iteration.
        this.prevLow = this.newLow;
        this.prevHigh = this.newHigh;
      }

      if (this.isLong) {
        return this.updateLong();
      }
      return this.updateShort();
    }

    return Number.NaN;
  }

  private updateLong(): number {
    // Switch to short if the low penetrates the SAR value.
    if (this.newLow <= this.sar) {
      this.isLong = false;
      this.sar = this.ep;

      if (this.sar < this.prevHigh) this.sar = this.prevHigh;
      if (this.sar < this.newHigh) this.sar = this.newHigh;

      if (this.offsetOnReverse !== 0) {
        this.sar += this.sar * this.offsetOnReverse;
      }

      const result = -this.sar;

      // Reset short AF and set EP.
      this.afShort = this.afInitShort;
      this.ep = this.newLow;

      // Calculate the new SAR.
      this.sar = this.sar + this.afShort * (this.ep - this.sar);
      if (this.sar < this.prevHigh) this.sar = this.prevHigh;
      if (this.sar < this.newHigh) this.sar = this.newHigh;

      return result;
    }

    // No switch — output the current SAR.
    const result = this.sar;

    // Adjust AF and EP.
    if (this.newHigh > this.ep) {
      this.ep = this.newHigh;
      this.afLong += this.afStepLong;
      if (this.afLong > this.afMaxLong) this.afLong = this.afMaxLong;
    }

    // Calculate the new SAR.
    this.sar = this.sar + this.afLong * (this.ep - this.sar);
    if (this.sar > this.prevLow) this.sar = this.prevLow;
    if (this.sar > this.newLow) this.sar = this.newLow;

    return result;
  }

  private updateShort(): number {
    // Switch to long if the high penetrates the SAR value.
    if (this.newHigh >= this.sar) {
      this.isLong = true;
      this.sar = this.ep;

      if (this.sar > this.prevLow) this.sar = this.prevLow;
      if (this.sar > this.newLow) this.sar = this.newLow;

      if (this.offsetOnReverse !== 0) {
        this.sar -= this.sar * this.offsetOnReverse;
      }

      const result = this.sar;

      // Reset long AF and set EP.
      this.afLong = this.afInitLong;
      this.ep = this.newHigh;

      // Calculate the new SAR.
      this.sar = this.sar + this.afLong * (this.ep - this.sar);
      if (this.sar > this.prevLow) this.sar = this.prevLow;
      if (this.sar > this.newLow) this.sar = this.newLow;

      return result;
    }

    // No switch — output the negated SAR.
    const result = -this.sar;

    // Adjust AF and EP.
    if (this.newLow < this.ep) {
      this.ep = this.newLow;
      this.afShort += this.afStepShort;
      if (this.afShort > this.afMaxShort) this.afShort = this.afMaxShort;
    }

    // Calculate the new SAR.
    this.sar = this.sar + this.afShort * (this.ep - this.sar);
    if (this.sar < this.prevHigh) this.sar = this.prevHigh;
    if (this.sar < this.newHigh) this.sar = this.newHigh;

    return result;
  }

  /** Updates the indicator given the next bar sample, extracting high and low. */
  public override updateBar(sample: Bar): IndicatorOutput {
    const v = this.updateHL(sample.high, sample.low);
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = v;
    return [scalar];
  }
}
