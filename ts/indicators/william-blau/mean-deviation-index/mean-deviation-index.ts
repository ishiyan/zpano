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
import { MeanDeviationIndexParams } from './params';

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

/** Function to calculate mnemonic of a __MeanDeviationIndex__ indicator. */
export const meanDeviationIndexMnemonic = (
  r: number, s: number, u: number,
  barComponent?: BarComponent, quoteComponent?: QuoteComponent, tradeComponent?: TradeComponent,
): string => {
  const cm = componentTripleMnemonic(barComponent, quoteComponent, tradeComponent);

  return `mdi(${r},${s},${u}${cm})`;
};

/**
 * MeanDeviationIndex is William Blau's Mean Deviation Index (MDI) indicator.
 *
 * A detrended, double-/triple-smoothed momentum line in raw price units, paired
 * with an EMA signal line (the Ergodic form, Blau ch. 5):
 *
 *   md_k     = price_k - EMA(price, r)_k                  (deviation from trend)
 *   mdi_k    = EMA(EMA(md, s), u)_k                       (the MDI line)
 *   signal_k = EMA(mdi, ul)_k                             (ul-period EMA)
 *
 * The price series is detrended by subtracting its own r-period EMA, then the
 * deviation is smoothed by an s-period EMA and an optional u-period EMA. Blau
 * notes the MDI approximates the MACD when r is long and s is short.
 *
 * The index is NOT normalized: there is no 100 * TEMA/TEMA ratio and no fixed
 * range, so the output is in the same price units as the input and may take any
 * sign or magnitude. Because there is no division there is also no division guard.
 *
 * The indicator produces two outputs:
 *   - MDI: the index line, in raw price units, finite from bar 0;
 *   - Signal: the ul-period EMA of the index (Blau's Ergodic signal line).
 *
 * Priming convention (book / EasyLanguage): each EMA stage seeds on its first
 * received value. The detrending EMA is defined from bar 0, so md_0 = 0 and both
 * smoothing EMAs seed on that 0 -- there is no NaN warm-up region and bar 0 is
 * exactly 0.0. Degenerate case: r=1 makes the detrend a passthrough, so the
 * index is identically 0.0.
 *
 * Reference:
 *
 * Blau, William (1995). Momentum, Direction, and Divergence, ch. 5. Wiley.
 */
export class MeanDeviationIndex implements Indicator {

  private readonly trend: Ema;
  private readonly smoothS: Ema;
  private readonly smoothU: Ema;

  private readonly signalEma: Ema;

  private primed_ = false;

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  private readonly mnemonic_: string;
  private readonly description_: string;

  constructor(params?: MeanDeviationIndexParams) {
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

    const bc = p.barComponent ?? DefaultBarComponent;
    const qc = p.quoteComponent ?? DefaultQuoteComponent;
    const tc = p.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    // The detrending baseline: a single EMA(r) on the price. The mean deviation
    // is the price minus this trend.
    this.trend = new Ema(r);

    // Two chained EMAs smoothing the deviation: EMA(EMA(md, s), u).
    this.smoothS = new Ema(s);
    this.smoothU = new Ema(u);

    // Signal line: a ul-period EMA of the index. The index is finite from bar 0,
    // so this seeds on bar 0 -- no NaN warm-up.
    this.signalEma = new Ema(ul);

    this.mnemonic_ = meanDeviationIndexMnemonic(
      r, s, u,
      p.barComponent, p.quoteComponent, p.tradeComponent,
    );
    this.description_ = 'Mean Deviation Index ' + this.mnemonic_;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.primed_;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.MeanDeviationIndex,
      this.mnemonic_,
      this.description_,
      [
        { mnemonic: this.mnemonic_ + ' mdi', description: this.description_ + ' MDI' },
        { mnemonic: this.mnemonic_ + ' signal', description: this.description_ + ' signal' },
      ],
    );
  }

  /**
   * Updates the indicator given the next sample value.
   * Returns [mdi, signal].
   */
  public update(sample: number): [number, number] {
    // Mean deviation: the price minus its own r-period EMA trend. The baseline
    // EMA seeds on bar 0, so the bar-0 deviation is exactly 0.
    const md = sample - this.trend.update(sample);

    // Smooth the deviation: EMA(EMA(md, s), u). No normalization, no guard.
    const mdi = this.smoothU.update(this.smoothS.update(md));

    // Signal line = EMA(mdi, ul); seeds here on the bar-0 index value.
    const signal = this.signalEma.update(mdi);
    this.primed_ = true;

    return [mdi, signal];
  }

  /** Updates the indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    const [mdi, signal] = this.update(sample.value);

    const s0 = new Scalar(); s0.time = sample.time; s0.value = mdi;
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
