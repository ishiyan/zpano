import { buildMetadata } from '../../core/build-metadata';
import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { Indicator } from '../../core/indicator';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { CandlestickStrengthIndexParams } from './params';

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

/** Function to calculate mnemonic of a __CandlestickStrengthIndex__ indicator. */
export const candlestickStrengthIndexMnemonic = (
  r: number, s: number, u: number, ul: number,
): string => `csi(${r},${s},${u},${ul})`;

/**
 * CandlestickStrengthIndex is William Blau's Candlestick Strength Index (CSI) indicator,
 * known in the book as the CandleStick Indicator.
 *
 * A double-/triple-smoothed candle-body-vs-range oscillator bounded to [-100, +100],
 * paired with an EMA signal line (the Ergodic form, Blau ch.6.4):
 *
 *   csi_k    = 100 * TEMA(close-open, r, s, u) / TEMA(high-low, r, s, u)   (the oscillator)
 *   signal_k = EMA(csi, ul)_k                                              (ul-period EMA)
 *
 * where the two intra-bar quantities are the signed candle body co_k = close_k - open_k
 * and the bar range hl_k = high_k - low_k >= 0, and TEMA(x, r, s, u) = EMA(EMA(EMA(x, r), s), u).
 *
 * It is the range-normalized sibling of the Candlestick Momentum Index (CMI): both share
 * the signed numerator TEMA(close-open), but the CSI divides by the smoothed range while
 * the CMI divides by the smoothed absolute body. Because every bar has |close-open| <= high-low,
 * the ratio is bounded to [-100, +100]: +100 when closes pin the high while opens pin the low
 * (relentless bullish bodies), -100 in the mirror-image bearish case, and 0 when bodies net out.
 * The inputs are the open, high, low and close prices.
 *
 * The indicator produces two outputs:
 *   - CSI: the oscillator, range [-100, +100];
 *   - Signal: the ul-period EMA of the oscillator (Blau's Ergodic signal line).
 *
 * Priming convention (book / EasyLanguage): each EMA stage seeds on its first
 * received value. Both intra-bar series are defined from bar 0, so there is no NaN
 * warm-up region -- all stages seed on bar 0 and both outputs are finite for every
 * bar. Division guard: denominator <= 0 -> oscillator 0.0.
 *
 * Reference:
 *
 * Blau, William (1995). Momentum, Direction, and Divergence, ch. 6 and Appendix B
 * (Figure B-15). Wiley.
 */
export class CandlestickStrengthIndex implements Indicator {

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

  constructor(params?: CandlestickStrengthIndexParams) {
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

    this.numR = new Ema(r);
    this.numS = new Ema(s);
    this.numU = new Ema(u);
    this.denR = new Ema(r);
    this.denS = new Ema(s);
    this.denU = new Ema(u);

    this.signalEma = new Ema(ul);

    this.mnemonic_ = candlestickStrengthIndexMnemonic(r, s, u, ul);
    this.description_ = 'Candlestick Strength Index ' + this.mnemonic_;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.primed_;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.CandlestickStrengthIndex,
      this.mnemonic_,
      this.description_,
      [
        { mnemonic: this.mnemonic_ + ' csi', description: this.description_ + ' CSI' },
        { mnemonic: this.mnemonic_ + ' signal', description: this.description_ + ' signal' },
      ],
    );
  }

  /**
   * Updates the indicator given the next bar's open, high, low and close values.
   * Returns [csi, signal].
   */
  public update(open: number, high: number, low: number, close: number): [number, number] {
    // Two intra-bar quantities: the signed candle body and the (non-negative) range.
    const co = close - open;
    const hl = high - low;

    // Numerator cascade: TEMA(close-open, r, s, u).
    const n = this.numU.update(this.numS.update(this.numR.update(co)));
    // Denominator cascade: TEMA(high-low, r, s, u).
    const d = this.denU.update(this.denS.update(this.denR.update(hl)));

    // Division guard: zero range so far -> oscillator 0.0.
    const csi = d > 0 ? (100 * n) / d : 0;

    // Signal line = EMA(csi, ul); seeds on bar 0's oscillator value.
    const signal = this.signalEma.update(csi);
    this.primed_ = true;

    return [csi, signal];
  }

  /** Updates the indicator and wraps the two outputs. */
  private updateEntity(
    time: Date, open: number, high: number, low: number, close: number,
  ): IndicatorOutput {
    const [csi, signal] = this.update(open, high, low, close);

    const s0 = new Scalar(); s0.time = time; s0.value = csi;
    const s1 = new Scalar(); s1.time = time; s1.value = signal;

    return [s0, s1];
  }

  /**
   * Updates the indicator given the next scalar sample.
   *
   * A scalar carries a single value, so the candle body and the bar range are both zero.
   */
  public updateScalar(sample: Scalar): IndicatorOutput {
    return this.updateEntity(sample.time, sample.value, sample.value, sample.value, sample.value);
  }

  /** Updates the indicator given the next bar sample. */
  public updateBar(sample: Bar): IndicatorOutput {
    return this.updateEntity(sample.time, sample.open, sample.high, sample.low, sample.close);
  }

  /**
   * Updates the indicator given the next quote sample.
   *
   * A quote maps the bid to the open and the low, and the ask to the close and the high,
   * so the candle body and the bar range are both the spread.
   */
  public updateQuote(sample: Quote): IndicatorOutput {
    return this.updateEntity(
      sample.time, sample.bidPrice, sample.askPrice, sample.bidPrice, sample.askPrice,
    );
  }

  /**
   * Updates the indicator given the next trade sample.
   *
   * A trade carries a single price, so the candle body and the bar range are both zero.
   */
  public updateTrade(sample: Trade): IndicatorOutput {
    return this.updateEntity(sample.time, sample.price, sample.price, sample.price, sample.price);
  }
}
