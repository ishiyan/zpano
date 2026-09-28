import { buildMetadata } from '../../core/build-metadata';
import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { Indicator } from '../../core/indicator';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { DirectionalTrendIndexParams } from './params';

/**
 * Stateful streaming EMA: alpha = 2/(period+1), seeds e0 = x0.
 *
 * Inlined verbatim from the Blau exponential moving average so the indicator is a
 * standalone porting unit. Do NOT change its numerics.
 *
 * period == 1 -> alpha == 1 -> pure passthrough (output == input).
 */
class Ema {
  private readonly alpha: number;
  private previous = 0;
  private primed = false;

  constructor(period: number) {
    this.alpha = 2 / (period + 1);
  }

  public update(x: number): number {
    if (!this.primed) {
      this.previous = x;
      this.primed = true;
      return this.previous;
    }
    this.previous = this.alpha * x + (1 - this.alpha) * this.previous;
    return this.previous;
  }
}

/** Function to calculate mnemonic of a __DirectionalTrendIndex__ indicator. */
export const directionalTrendIndexMnemonic = (
  q: number, r: number, s: number, u: number, ul: number,
): string => `dti(${q},${r},${s},${u},${ul})`;

/**
 * DirectionalTrendIndex is William Blau's Directional Trend Index (DTI) indicator.
 *
 * A double-/triple-smoothed High-Low Momentum oscillator bounded to [-100, +100],
 * paired with an EMA signal line (the Ergodic form):
 *
 *   dti_k    = 100 * TEMA(HLM, r, s, u)_k / TEMA(|HLM|, r, s, u)_k   (the oscillator)
 *   signal_k = EMA(dti, ul)_k                                       (ul-period EMA)
 *
 * where the High-Low Momentum is built from how far the high rose and the low
 * fell relative to q-1 bars ago:
 *
 *   HMU_k = max(high_k - high_(k-(q-1)), 0)      (upward high movement)
 *   LMD_k = max(low_(k-(q-1)) - low_k, 0)        (downward low movement)
 *   HLM_k = HMU_k - LMD_k                         (composite high-low momentum)
 *   TEMA(x, r, s, u) = EMA(EMA(EMA(x, r), s), u)  (triple EMA cascade)
 *
 * This is the True Strength Index structure applied to HLM instead of price
 * momentum. The inputs are the high and low prices only (no close).
 *
 * The indicator produces two outputs:
 *   - DTI: the oscillator, range [-100, +100];
 *   - Signal: the ul-period EMA of the oscillator.
 *
 * Priming convention (book / EasyLanguage): HLM is valid from bar q-1 (it needs a
 * high/low from q-1 bars ago), so all cascade stages seed there together; both
 * outputs are NaN for bars 0..q-2 and finite from bar q-1. For q = 1 there is no
 * NaN warm-up, but HLM == 0 on every bar, so the division guard yields dti == 0.0
 * for all bars. The signal EMA seeds on the first finite oscillator value.
 * Division guard: denominator == 0 -> oscillator 0.0.
 *
 * Reference:
 *
 * Blau, William (1995). Momentum, Direction, and Divergence. Wiley.
 */
export class DirectionalTrendIndex implements Indicator {

  private readonly q: number;
  private readonly highs: number[];
  private readonly lows: number[];
  private windowCount = 0;
  private windowIndex = 0;

  private readonly numR: Ema;
  private readonly numS: Ema;
  private readonly numU: Ema;
  private readonly denR: Ema;
  private readonly denS: Ema;
  private readonly denU: Ema;

  private readonly signalEma: Ema;

  private primed_ = false;

  private readonly mnemonic_: string;
  private readonly description_: string;

  constructor(params?: DirectionalTrendIndexParams) {
    const p = params ?? {};

    const q = Math.floor(p.q ?? 2);
    const r = Math.floor(p.r ?? 20);
    const s = Math.floor(p.s ?? 5);
    const u = Math.floor(p.u ?? 3);
    const ul = Math.floor(p.ul ?? 3);

    if (q < 1) {
      throw new Error('q should be greater than 0');
    }

    if (r < 1) {
      throw new Error('r should be greater than 0');
    }

    if (s < 1) {
      throw new Error('s should be greater than 0');
    }

    if (u < 1) {
      throw new Error('u should be greater than 0');
    }

    if (ul < 1) {
      throw new Error('ul should be greater than 0');
    }

    this.q = q;
    this.highs = new Array<number>(q).fill(0);
    this.lows = new Array<number>(q).fill(0);

    this.numR = new Ema(r);
    this.numS = new Ema(s);
    this.numU = new Ema(u);
    this.denR = new Ema(r);
    this.denS = new Ema(s);
    this.denU = new Ema(u);

    this.signalEma = new Ema(ul);

    this.mnemonic_ = directionalTrendIndexMnemonic(q, r, s, u, ul);
    this.description_ = 'Directional Trend Index ' + this.mnemonic_;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.primed_;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.DirectionalTrendIndex,
      this.mnemonic_,
      this.description_,
      [
        { mnemonic: this.mnemonic_ + ' dti', description: this.description_ + ' DTI' },
        { mnemonic: this.mnemonic_ + ' signal', description: this.description_ + ' signal' },
      ],
    );
  }

  /**
   * Updates the indicator given the next bar's high and low values.
   * Returns [dti, signal]; both are NaN until q bars have been seen.
   */
  public update(high: number, low: number): [number, number] {
    this.highs[this.windowIndex] = high;
    this.lows[this.windowIndex] = low;
    this.windowIndex = (this.windowIndex + 1) % this.q;

    if (this.windowCount < this.q) {
      this.windowCount++;
    }

    // HLM needs a high/low from q-1 bars ago. Until then neither output
    // exists -- do NOT advance the EMA cascades.
    if (this.windowCount < this.q) {
      return [Number.NaN, Number.NaN];
    }

    // The oldest value in the full window sits at the next write position:
    // high_(k-(q-1)) and low_(k-(q-1)).
    const previousHigh = this.highs[this.windowIndex];
    const previousLow = this.lows[this.windowIndex];

    // Upward high movement and downward low movement, each floored at 0.
    const hmu = Math.max(high - previousHigh, 0);
    const lmd = Math.max(previousLow - low, 0);

    // Composite high-low momentum and its magnitude.
    const hlm = hmu - lmd;
    const absHlm = Math.abs(hlm);

    // Numerator cascade: TEMA(HLM, r, s, u).
    const n = this.numU.update(this.numS.update(this.numR.update(hlm)));
    // Denominator cascade: TEMA(|HLM|, r, s, u).
    const d = this.denU.update(this.denS.update(this.denR.update(absHlm)));

    // Division guard: denominator 0 -> oscillator 0.0.
    const dti = d !== 0 ? (100 * n) / d : 0;

    // Signal line = EMA(dti, ul); seeds on the first finite oscillator value.
    const signal = this.signalEma.update(dti);
    this.primed_ = true;

    return [dti, signal];
  }

  /** Updates the indicator and wraps the two outputs. */
  private updateEntity(time: Date, high: number, low: number): IndicatorOutput {
    const [dti, signal] = this.update(high, low);

    const s0 = new Scalar(); s0.time = time; s0.value = dti;
    const s1 = new Scalar(); s1.time = time; s1.value = signal;

    return [s0, s1];
  }

  /**
   * Updates the indicator given the next scalar sample.
   *
   * A scalar carries a single value, used as both the high and the low.
   */
  public updateScalar(sample: Scalar): IndicatorOutput {
    const v = sample.value;
    return this.updateEntity(sample.time, v, v);
  }

  /** Updates the indicator given the next bar sample. */
  public updateBar(sample: Bar): IndicatorOutput {
    return this.updateEntity(sample.time, sample.high, sample.low);
  }

  /**
   * Updates the indicator given the next quote sample.
   *
   * A quote maps the mid price to both the high and the low.
   */
  public updateQuote(sample: Quote): IndicatorOutput {
    const v = (sample.bidPrice + sample.askPrice) / 2;
    return this.updateEntity(sample.time, v, v);
  }

  /**
   * Updates the indicator given the next trade sample.
   *
   * A trade carries a single price, used as both the high and the low.
   */
  public updateTrade(sample: Trade): IndicatorOutput {
    const v = sample.price;
    return this.updateEntity(sample.time, v, v);
  }
}
