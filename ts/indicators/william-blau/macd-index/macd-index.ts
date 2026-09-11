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
import { MacdIndexParams } from './params';

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

/** Function to calculate mnemonic of a __MacdIndex__ indicator. */
export const macdIndexMnemonic = (
  r: number, s: number, u: number,
  barComponent?: BarComponent, quoteComponent?: QuoteComponent, tradeComponent?: TradeComponent,
): string => {
  const cm = componentTripleMnemonic(barComponent, quoteComponent, tradeComponent);

  return `macdi(${r},${s},${u}${cm})`;
};

/**
 * MacdIndex is William Blau's MACD Index (MACD_I) indicator.
 *
 * Blau's MACD line is the difference of two EMAs of the close, optionally
 * smoothed by a third EMA, paired with an EMA signal line (the Ergodic form,
 * Blau ch. 5):
 *
 *   macd_k   = EMA(close, s)_k - EMA(close, r)_k          (MACD line; s fast, r slow)
 *   macdi_k  = EMA(macd, u)_k                             (the MACD_I line)
 *   signal_k = EMA(macdi, ul)_k                           (ul-period EMA)
 *
 * with the fast period s strictly shorter than the slow period r (s < r).
 * Setting u=1 recovers the book's pure two-EMA MACD line. Blau notes the MACD
 * and the MDI are both double-smoothed momentum indicators with nearly
 * interchangeable shapes (within a scale factor).
 *
 * The index is NOT normalized: there is no 100 * TEMA/TEMA ratio and no fixed
 * range, so the output is in the same price units as the input and may take any
 * sign or magnitude. Because there is no division there is also no division guard.
 *
 * The indicator produces two outputs:
 *   - MACDI: the index line, in raw price units, finite from bar 0;
 *   - Signal: the ul-period EMA of the index (Blau's Ergodic signal line).
 *
 * Priming convention (book / EasyLanguage): each EMA stage seeds on its first
 * received value. Both price EMAs are defined from bar 0, so macd_0 = 0 and the
 * u smoothing and signal EMAs seed on that 0 -- there is no NaN warm-up region
 * and bar 0 is exactly 0.0.
 *
 * Reference:
 *
 * Blau, William (1995). Momentum, Direction, and Divergence, ch. 5. Wiley.
 */
export class MacdIndex implements Indicator {

  private readonly emaFast: Ema;
  private readonly emaSlow: Ema;
  private readonly smoothU: Ema;

  private readonly signalEma: Ema;

  private primed_ = false;

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  private readonly mnemonic_: string;
  private readonly description_: string;

  constructor(params?: MacdIndexParams) {
    const p = params ?? {};

    const r = Math.floor(p.r ?? 20);
    const s = Math.floor(p.s ?? 5);
    const u = Math.floor(p.u ?? 3);
    const ul = Math.floor(p.ul ?? 3);

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

    if (s >= r) {
      throw new Error('s (fast) should be less than r (slow)');
    }

    const bc = p.barComponent ?? DefaultBarComponent;
    const qc = p.quoteComponent ?? DefaultQuoteComponent;
    const tc = p.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    // The two price EMAs forming the MACD line = EMA(close, s) - EMA(close, r).
    this.emaFast = new Ema(s);
    this.emaSlow = new Ema(r);

    // Third smoothing EMA applied to the MACD line: EMA(macd, u).
    this.smoothU = new Ema(u);

    // Signal line: a ul-period EMA of the index. The index is finite from bar 0,
    // so this seeds on bar 0 -- no NaN warm-up.
    this.signalEma = new Ema(ul);

    this.mnemonic_ = macdIndexMnemonic(
      r, s, u,
      p.barComponent, p.quoteComponent, p.tradeComponent,
    );
    this.description_ = 'MACD Index ' + this.mnemonic_;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.primed_;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.MacdIndex,
      this.mnemonic_,
      this.description_,
      [
        { mnemonic: this.mnemonic_ + ' macdi', description: this.description_ + ' MACDI' },
        { mnemonic: this.mnemonic_ + ' signal', description: this.description_ + ' signal' },
      ],
    );
  }

  /**
   * Updates the indicator given the next sample value.
   * Returns [macdi, signal].
   */
  public update(sample: number): [number, number] {
    // MACD line = fast EMA - slow EMA. Both seed at bar 0 (to close_0), so the
    // line is 0.0 on bar 0 and defined on every bar thereafter.
    const macd = this.emaFast.update(sample) - this.emaSlow.update(sample);

    // Smooth the MACD line: EMA(macd, u). No normalization, no guard.
    const macdi = this.smoothU.update(macd);

    // Signal line = EMA(macdi, ul); seeds here on the bar-0 index value.
    const signal = this.signalEma.update(macdi);
    this.primed_ = true;

    return [macdi, signal];
  }

  /** Updates the indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    const [macdi, signal] = this.update(sample.value);

    const s0 = new Scalar(); s0.time = sample.time; s0.value = macdi;
    const s1 = new Scalar(); s1.time = sample.time; s1.value = signal;

    return [s0, s1];
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
