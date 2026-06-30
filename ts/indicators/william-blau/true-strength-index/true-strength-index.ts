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
import { TrueStrengthIndexParams } from './params';

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

/** Function to calculate mnemonic of a __TrueStrengthIndex__ indicator. */
export const trueStrengthIndexMnemonic = (
  q: number, r: number, s: number, u: number,
  barComponent?: BarComponent, quoteComponent?: QuoteComponent, tradeComponent?: TradeComponent,
): string => {
  const cm = componentTripleMnemonic(barComponent, quoteComponent, tradeComponent);

  return `tsi(${q},${r},${s},${u}${cm})`;
};

/**
 * TrueStrengthIndex is William Blau's True Strength Index (TSI) indicator.
 *
 * A double-/triple-smoothed momentum oscillator bounded to [-100, +100], paired
 * with an EMA signal line (the Ergodic form, Blau ch.1.4):
 *
 *   tsi_k    = 100 * TEMA(mtm, r, s, u) / TEMA(|mtm|, r, s, u)   (the oscillator)
 *   signal_k = EMA(tsi, ul)_k                                    (ul-period EMA)
 *
 * where mtm_k = C_k - C_(k-(q-1)) and TEMA(x, r, s, u) = EMA(EMA(EMA(x, r), s), u).
 *
 * The indicator produces two outputs:
 *   - TSI: the oscillator, range [-100, +100], NaN during warm-up (bars 0..q-2);
 *   - Signal: the ul-period EMA of the oscillator (Blau's Ergodic signal line).
 *
 * Priming convention (book / EasyLanguage): each EMA stage seeds on its first
 * received value, so all stages seed at bar q-1 together; the signal EMA seeds on
 * the first finite oscillator. Division guard: denominator 0 -> oscillator 0.0.
 *
 * Reference:
 *
 * Blau, William (1995). Momentum, Direction, and Divergence, ch. 2. Wiley.
 */
export class TrueStrengthIndex implements Indicator {

  private readonly q: number;

  private readonly history: number[] = [];

  private readonly numR: Ema;
  private readonly numS: Ema;
  private readonly numU: Ema;
  private readonly denR: Ema;
  private readonly denS: Ema;
  private readonly denU: Ema;

  private readonly signalEma: Ema;

  private primed_ = false;

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  private readonly mnemonic_: string;
  private readonly description_: string;

  constructor(params?: TrueStrengthIndexParams) {
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

    const bc = p.barComponent ?? DefaultBarComponent;
    const qc = p.quoteComponent ?? DefaultQuoteComponent;
    const tc = p.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    this.q = q;

    this.numR = new Ema(r);
    this.numS = new Ema(s);
    this.numU = new Ema(u);
    this.denR = new Ema(r);
    this.denS = new Ema(s);
    this.denU = new Ema(u);

    this.signalEma = new Ema(ul);

    this.mnemonic_ = trueStrengthIndexMnemonic(
      q, r, s, u,
      p.barComponent, p.quoteComponent, p.tradeComponent,
    );
    this.description_ = 'True Strength Index ' + this.mnemonic_;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.primed_;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.TrueStrengthIndex,
      this.mnemonic_,
      this.description_,
      [
        { mnemonic: this.mnemonic_ + ' tsi', description: this.description_ + ' TSI' },
        { mnemonic: this.mnemonic_ + ' signal', description: this.description_ + ' signal' },
      ],
    );
  }

  /**
   * Updates the indicator given the next sample value.
   * Returns [tsi, signal].
   */
  public update(sample: number): [number, number] {
    // Maintain a rolling window of the last q prices; the leftmost element is
    // C_(k-(q-1)).
    if (this.history.length < this.q) {
      this.history.push(sample);
    } else {
      this.history.shift();
      this.history.push(sample);
    }

    // Momentum needs a price from q-1 bars ago, available only once the window
    // holds q prices. Before then neither output is defined and the signal EMA
    // is NOT advanced.
    if (this.history.length < this.q) {
      return [NaN, NaN];
    }

    // mtm_k = C_k - C_(k-(q-1)); the leftmost history element is C_(k-(q-1)).
    const mtm = sample - this.history[0];
    const absMtm = Math.abs(mtm);

    // Numerator cascade: TEMA(mtm, r, s, u).
    const n = this.numU.update(this.numS.update(this.numR.update(mtm)));
    // Denominator cascade: TEMA(|mtm|, r, s, u).
    const d = this.denU.update(this.denS.update(this.denR.update(absMtm)));

    // Division guard (Blau_TSI.mq5): denominator 0 -> oscillator 0.0.
    const tsi = d === 0 ? 0 : (100 * n) / d;

    // Signal line = EMA(tsi, ul); seeds here on the first finite oscillator.
    const signal = this.signalEma.update(tsi);
    this.primed_ = true;

    return [tsi, signal];
  }

  /** Updates the indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    const [tsi, signal] = this.update(sample.value);

    const s0 = new Scalar(); s0.time = sample.time; s0.value = tsi;
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
