import { buildMetadata } from '../../core/build-metadata';
import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { Indicator } from '../../core/indicator';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { CandlestickMomentumIndexParams } from './params';

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

/** Function to calculate mnemonic of a __CandlestickMomentumIndex__ indicator. */
export const candlestickMomentumIndexMnemonic = (
  r: number, s: number, u: number, ul: number,
): string => `cmi(${r},${s},${u},${ul})`;

/**
 * CandlestickMomentumIndex is William Blau's Candlestick Momentum Index (CMI) indicator.
 *
 * A double-/triple-smoothed intra-bar momentum oscillator bounded to [-100, +100],
 * paired with an EMA signal line (the Ergodic form, Blau ch.6.4):
 *
 *   cmi_k    = 100 * TEMA(cmtm, r, s, u) / TEMA(|cmtm|, r, s, u)   (the oscillator)
 *   signal_k = EMA(cmi, ul)_k                                      (ul-period EMA)
 *
 * where the candle momentum is the signed candle body cmtm_k = close_k - open_k
 * and TEMA(x, r, s, u) = EMA(EMA(EMA(x, r), s), u).
 *
 * This is the True Strength Index structure applied to the candle body instead of
 * price momentum. Because it only looks inside each bar it is immune to inter-bar
 * gaps. The inputs are the open and close prices only.
 *
 * The indicator produces two outputs:
 *   - CMI: the oscillator, range [-100, +100];
 *   - Signal: the ul-period EMA of the oscillator (Blau's Ergodic signal line).
 *
 * Priming convention (book / EasyLanguage): each EMA stage seeds on its first
 * received value. The candle momentum is defined from bar 0, so there is no NaN
 * warm-up region -- all stages seed on bar 0 and both outputs are finite for every
 * bar. Division guard: denominator 0 -> oscillator 0.0.
 *
 * Reference:
 *
 * Blau, William (1995). Momentum, Direction, and Divergence, ch. 6. Wiley.
 */
export class CandlestickMomentumIndex implements Indicator {

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

  constructor(params?: CandlestickMomentumIndexParams) {
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

    this.mnemonic_ = candlestickMomentumIndexMnemonic(r, s, u, ul);
    this.description_ = 'Candlestick Momentum Index ' + this.mnemonic_;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.primed_;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.CandlestickMomentumIndex,
      this.mnemonic_,
      this.description_,
      [
        { mnemonic: this.mnemonic_ + ' cmi', description: this.description_ + ' CMI' },
        { mnemonic: this.mnemonic_ + ' signal', description: this.description_ + ' signal' },
      ],
    );
  }

  /**
   * Updates the indicator given the next bar's open and close values.
   * Returns [cmi, signal].
   */
  public update(open: number, close: number): [number, number] {
    // Candle momentum: the signed body of the candle.
    const cmtm = close - open;
    const absCmtm = Math.abs(cmtm);

    // Numerator cascade: TEMA(cmtm, r, s, u).
    const n = this.numU.update(this.numS.update(this.numR.update(cmtm)));
    // Denominator cascade: TEMA(|cmtm|, r, s, u).
    const d = this.denU.update(this.denS.update(this.denR.update(absCmtm)));

    // Division guard (Blau_CMI.mq5): denominator 0 -> oscillator 0.0.
    const cmi = d === 0 ? 0 : (100 * n) / d;

    // Signal line = EMA(cmi, ul); seeds on bar 0's oscillator value.
    const signal = this.signalEma.update(cmi);
    this.primed_ = true;

    return [cmi, signal];
  }

  /** Updates the indicator and wraps the two outputs. */
  private updateEntity(time: Date, open: number, close: number): IndicatorOutput {
    const [cmi, signal] = this.update(open, close);

    const s0 = new Scalar(); s0.time = time; s0.value = cmi;
    const s1 = new Scalar(); s1.time = time; s1.value = signal;

    return [s0, s1];
  }

  /**
   * Updates the indicator given the next scalar sample.
   *
   * A scalar carries a single value, so open == close and the candle momentum is zero.
   */
  public updateScalar(sample: Scalar): IndicatorOutput {
    return this.updateEntity(sample.time, sample.value, sample.value);
  }

  /** Updates the indicator given the next bar sample. */
  public updateBar(sample: Bar): IndicatorOutput {
    return this.updateEntity(sample.time, sample.open, sample.close);
  }

  /**
   * Updates the indicator given the next quote sample.
   *
   * A quote maps the bid to the open and the ask to the close, so the candle body is the spread.
   */
  public updateQuote(sample: Quote): IndicatorOutput {
    return this.updateEntity(sample.time, sample.bidPrice, sample.askPrice);
  }

  /**
   * Updates the indicator given the next trade sample.
   *
   * A trade carries a single price, so open == close and the candle momentum is zero.
   */
  public updateTrade(sample: Trade): IndicatorOutput {
    return this.updateEntity(sample.time, sample.price, sample.price);
  }
}
