import { buildMetadata } from '../../core/build-metadata';
import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { Indicator } from '../../core/indicator';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { StochasticMomentumIndexParams } from './params';

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
  private prev = 0;
  private primed = false;

  constructor(period: number) {
    this.alpha = 2 / (period + 1);
  }

  public update(x: number): number {
    if (!this.primed) {
      this.prev = x;
      this.primed = true;
      return this.prev;
    }
    this.prev = this.alpha * x + (1 - this.alpha) * this.prev;
    return this.prev;
  }
}

/** Function to calculate mnemonic of a __StochasticMomentumIndex__ indicator. */
export const stochasticMomentumIndexMnemonic = (
  q: number, r: number, s: number, u: number, ul: number,
): string => `smi(${q},${r},${s},${u},${ul})`;

/**
 * StochasticMomentumIndex is William Blau's Stochastic Momentum Index (SMI) indicator.
 *
 * A double-/triple-smoothed stochastic oscillator bounded to [-100, +100],
 * paired with an EMA signal line (the Ergodic form, Blau ch.3.4):
 *
 *   smi_k    = 100 * TEMA(sm, r, s, u) / TEMA(hr, r, s, u)   (the oscillator)
 *   signal_k = EMA(smi, ul)_k                                (ul-period EMA)
 *
 * where, over the last q bars, HH_k is the highest high and LL_k is the lowest low,
 * sm_k = close_k - 0.5*(HH_k + LL_k) is the distance of the close from the range
 * midpoint, hr_k = 0.5*(HH_k - LL_k) >= 0 is the half-range, and
 * TEMA(x, r, s, u) = EMA(EMA(EMA(x, r), s), u).
 *
 * Where the ordinary stochastic measures where the close sits inside the recent
 * high-low range, the SMI measures the close relative to the midpoint of that range.
 * Because |sm| <= hr on every bar, the ratio is bounded to [-100, +100]. With q = 1 it
 * is Blau's one-day stochastic (sentiment indicator). The inputs are the high, low and
 * close prices.
 *
 * The indicator produces two outputs:
 *   - SMI: the oscillator, range [-100, +100];
 *   - Signal: the ul-period EMA of the oscillator (Blau's Ergodic signal line).
 *
 * Priming convention (book / EasyLanguage): sm and hr become valid once q bars of
 * high/low exist, i.e. at bar q-1. All six cascade stages seed there together, so
 * both outputs are NaN for bars 0..q-2 and finite from bar q-1; for q = 1 there is
 * no NaN warm-up. Division guard: denominator <= 0 -> oscillator 0.0.
 *
 * Reference:
 *
 * Blau, William (1995). Momentum, Direction, and Divergence, ch. 3. Wiley.
 */
export class StochasticMomentumIndex implements Indicator {

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

  constructor(params?: StochasticMomentumIndexParams) {
    const p = params ?? {};

    const q = Math.floor(p.q ?? 5);
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

    this.mnemonic_ = stochasticMomentumIndexMnemonic(q, r, s, u, ul);
    this.description_ = 'Stochastic Momentum Index ' + this.mnemonic_;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.primed_;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.StochasticMomentumIndex,
      this.mnemonic_,
      this.description_,
      [
        { mnemonic: this.mnemonic_ + ' smi', description: this.description_ + ' SMI' },
        { mnemonic: this.mnemonic_ + ' signal', description: this.description_ + ' signal' },
      ],
    );
  }

  /**
   * Updates the indicator given the next bar's high, low and close values.
   * Returns [smi, signal]; both are NaN until q bars have been seen.
   */
  public update(high: number, low: number, close: number): [number, number] {
    this.highs[this.windowIndex] = high;
    this.lows[this.windowIndex] = low;
    this.windowIndex = (this.windowIndex + 1) % this.q;

    if (this.windowCount < this.q) {
      this.windowCount++;
    }

    // Need q bars of high/low before the stochastic is defined. Until then
    // neither output exists -- do NOT advance the EMA cascades.
    if (this.windowCount < this.q) {
      return [Number.NaN, Number.NaN];
    }

    // Rolling extremes over the last q bars.
    let hh = this.highs[0];
    let ll = this.lows[0];

    for (let i = 1; i < this.q; i++) {
      hh = Math.max(hh, this.highs[i]);
      ll = Math.min(ll, this.lows[i]);
    }

    // Stochastic momentum (signed) and half-range (non-negative).
    const sm = close - 0.5 * (hh + ll);
    const hr = 0.5 * (hh - ll);

    // Numerator cascade: TEMA(sm, r, s, u).
    const n = this.numU.update(this.numS.update(this.numR.update(sm)));
    // Denominator cascade: TEMA(hr, r, s, u).
    const d = this.denU.update(this.denS.update(this.denR.update(hr)));

    // Division guard: flat window so far -> oscillator 0.0.
    const smi = d > 0 ? (100 * n) / d : 0;

    // Signal line = EMA(smi, ul); seeds on the first finite oscillator value.
    const signal = this.signalEma.update(smi);
    this.primed_ = true;

    return [smi, signal];
  }

  /** Updates the indicator and wraps the two outputs. */
  private updateEntity(time: Date, high: number, low: number, close: number): IndicatorOutput {
    const [smi, signal] = this.update(high, low, close);

    const s0 = new Scalar(); s0.time = time; s0.value = smi;
    const s1 = new Scalar(); s1.time = time; s1.value = signal;

    return [s0, s1];
  }

  /**
   * Updates the indicator given the next scalar sample.
   *
   * A scalar carries a single value, used as the high, the low and the close.
   */
  public updateScalar(sample: Scalar): IndicatorOutput {
    const v = sample.value;
    return this.updateEntity(sample.time, v, v, v);
  }

  /** Updates the indicator given the next bar sample. */
  public updateBar(sample: Bar): IndicatorOutput {
    return this.updateEntity(sample.time, sample.high, sample.low, sample.close);
  }

  /**
   * Updates the indicator given the next quote sample.
   *
   * A quote maps the mid price to the high, the low and the close.
   */
  public updateQuote(sample: Quote): IndicatorOutput {
    const v = (sample.bidPrice + sample.askPrice) / 2;
    return this.updateEntity(sample.time, v, v, v);
  }

  /**
   * Updates the indicator given the next trade sample.
   *
   * A trade carries a single price, used as the high, the low and the close.
   */
  public updateTrade(sample: Trade): IndicatorOutput {
    const v = sample.price;
    return this.updateEntity(sample.time, v, v, v);
  }
}
