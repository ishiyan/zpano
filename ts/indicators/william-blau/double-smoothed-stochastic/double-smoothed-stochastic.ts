import { buildMetadata } from '../../core/build-metadata';
import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { Indicator } from '../../core/indicator';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { DoubleSmoothedStochasticParams } from './params';

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

/**
 * Stateful streaming SMA over the last period inputs.
 *
 * Returns the mean of the window's current contents on every update: an expanding
 * window while fewer than period values have arrived, then a rolling period-bar
 * window. There is no NaN warm-up (finite from the first input).
 *
 * period == 1 -> pure passthrough (output == input).
 */
class Sma {
  private readonly period: number;
  private readonly window: number[];
  private count = 0;
  private index = 0;

  constructor(period: number) {
    this.period = period;
    this.window = new Array<number>(period).fill(0);
  }

  public update(x: number): number {
    this.window[this.index] = x;
    this.index = (this.index + 1) % this.period;

    if (this.count < this.period) {
      this.count++;
    }

    // Naive left-to-right sum from the oldest to the newest value (NOT a
    // compensated sum), so that every port reproduces the same values.
    // Once full, the oldest value sits at the next write position.
    const start = this.count === this.period ? this.index : 0;

    let sum = 0;
    for (let i = 0; i < this.count; i++) {
      sum += this.window[(start + i) % this.period];
    }

    return sum / this.count;
  }
}

/** Function to calculate mnemonic of a __DoubleSmoothedStochastic__ indicator. */
export const doubleSmoothedStochasticMnemonic = (
  q: number, r: number, s: number, g: number,
): string => `dss(${q},${r},${s},${g})`;

/**
 * DoubleSmoothedStochastic is William Blau's Double Smoothed Stochastic (DSS) indicator.
 *
 * A classic double-smoothed stochastic oscillator bounded to [0, 100], paired
 * with a short simple-moving-average signal line:
 *
 *   dss_k    = 100 * EMA(EMA(st, r), s) / EMA(EMA(rng, r), s)   (the oscillator)
 *   signal_k = SMA(dss, g)_k                                    (g-period SMA)
 *
 * where, over the last q bars, HH_k is the highest high and LL_k is the lowest low,
 * st_k = close_k - LL_k >= 0 is the raw stochastic (the close above the low), and
 * rng_k = HH_k - LL_k >= 0 is the q-bar range.
 *
 * The raw stochastic and the range are smoothed separately with the same two-stage
 * EMA cascade (r then s), then divided. Because 0 <= st <= rng on every bar, the
 * ratio is bounded to [0, 100]. With q = 1 it is Blau's one-bar HLC index. It is
 * exactly the MQL5 Blau_TStochI with its third EMA period u = 1. The inputs are the
 * high, low and close prices.
 *
 * The indicator produces two outputs:
 *   - DSS: the oscillator, range [0, 100];
 *   - Signal: the g-period SMA of the oscillator.
 *
 * Priming convention (book / EasyLanguage): st and rng become valid once q bars of
 * high/low exist, i.e. at bar q-1. All four cascade stages seed there together, so
 * both outputs are NaN for bars 0..q-2 and finite from bar q-1; for q = 1 there is
 * no NaN warm-up. The signal SMA seeds on the first finite oscillator value and
 * returns the mean of the oscillator values seen so far (expanding window <= g),
 * then the full g-bar rolling mean. Division guard: denominator <= 0 -> oscillator 0.0.
 *
 * Reference:
 *
 * Blau, William (1995). Momentum, Direction, and Divergence. Wiley.
 */
export class DoubleSmoothedStochastic implements Indicator {

  private readonly q: number;
  private readonly highs: number[];
  private readonly lows: number[];
  private windowCount = 0;
  private windowIndex = 0;

  private readonly numR: Ema;
  private readonly numS: Ema;
  private readonly denR: Ema;
  private readonly denS: Ema;

  private readonly signalSma: Sma;

  private primed_ = false;

  private readonly mnemonic_: string;
  private readonly description_: string;

  constructor(params?: DoubleSmoothedStochasticParams) {
    const p = params ?? {};

    const q = Math.floor(p.q ?? 5);
    const r = Math.floor(p.r ?? 7);
    const s = Math.floor(p.s ?? 3);
    const g = Math.floor(p.g ?? 3);

    if (q < 1) {
      throw new Error('q should be greater than 0');
    }

    if (r < 1) {
      throw new Error('r should be greater than 0');
    }

    if (s < 1) {
      throw new Error('s should be greater than 0');
    }

    if (g < 1) {
      throw new Error('g should be greater than 0');
    }

    this.q = q;
    this.highs = new Array<number>(q).fill(0);
    this.lows = new Array<number>(q).fill(0);

    this.numR = new Ema(r);
    this.numS = new Ema(s);
    this.denR = new Ema(r);
    this.denS = new Ema(s);

    this.signalSma = new Sma(g);

    this.mnemonic_ = doubleSmoothedStochasticMnemonic(q, r, s, g);
    this.description_ = 'Double Smoothed Stochastic ' + this.mnemonic_;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.primed_;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.DoubleSmoothedStochastic,
      this.mnemonic_,
      this.description_,
      [
        { mnemonic: this.mnemonic_ + ' dss', description: this.description_ + ' DSS' },
        { mnemonic: this.mnemonic_ + ' signal', description: this.description_ + ' signal' },
      ],
    );
  }

  /**
   * Updates the indicator given the next bar's high, low and close values.
   * Returns [dss, signal]; both are NaN until q bars have been seen.
   */
  public update(high: number, low: number, close: number): [number, number] {
    this.highs[this.windowIndex] = high;
    this.lows[this.windowIndex] = low;
    this.windowIndex = (this.windowIndex + 1) % this.q;

    if (this.windowCount < this.q) {
      this.windowCount++;
    }

    // Need q bars of high/low before the stochastic is defined. Until then
    // neither output exists -- do NOT advance the EMA cascades or the SMA.
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

    // Raw stochastic and range (both non-negative).
    const st = close - ll;
    const rng = hh - ll;

    // Numerator cascade: EMA(EMA(st, r), s).
    const n = this.numS.update(this.numR.update(st));
    // Denominator cascade: EMA(EMA(rng, r), s).
    const d = this.denS.update(this.denR.update(rng));

    // Division guard: flat window so far -> oscillator 0.0.
    const dss = d > 0 ? (100 * n) / d : 0;

    // Signal line = SMA(dss, g); seeds on the first finite oscillator value.
    const signal = this.signalSma.update(dss);
    this.primed_ = true;

    return [dss, signal];
  }

  /** Updates the indicator and wraps the two outputs. */
  private updateEntity(time: Date, high: number, low: number, close: number): IndicatorOutput {
    const [dss, signal] = this.update(high, low, close);

    const s0 = new Scalar(); s0.time = time; s0.value = dss;
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
