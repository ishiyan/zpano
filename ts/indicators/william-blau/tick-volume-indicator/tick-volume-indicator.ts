import { buildMetadata } from '../../core/build-metadata';
import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { Indicator } from '../../core/indicator';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { TickVolumeIndicatorParams } from './params';

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

/** Function to calculate mnemonic of a __TickVolumeIndicator__ indicator. */
export const tickVolumeIndicatorMnemonic = (
  r: number, s: number, u: number,
): string => `tvi(${r},${s},${u})`;

/**
 * TickVolumeIndicator is William Blau's Tick Volume Indicator (TVI).
 *
 * A normalized, double-/triple-smoothed oscillator built from the balance of
 * upticks vs downticks inside each bar, bounded to [-100, +100] (Blau ch.4, ch.10):
 *
 *   tvi_k = 100 * (TEMA(up, r, s, u) - TEMA(down, r, s, u))
 *               / (TEMA(up, r, s, u) + TEMA(down, r, s, u))
 *
 * where TEMA(x, r, s, u) = EMA(EMA(EMA(x, r), s), u). Setting u=1 recovers the
 * book's double-smoothed TVI(r, s), because EMA(., 1) is a passthrough. Because it
 * is built from intra-bar tick direction rather than from the close vs a previous
 * close, the TVI is immune to opening gaps.
 *
 * The inputs are two non-negative series, upticks and downticks. Genuine tick
 * counts are fed through `update`. The entity updates derive a deterministic proxy:
 *   - Bar: up = close - low, down = high - close (the intra-bar range split);
 *   - Scalar, Trade, Quote: a magnitude tick rule against the previous value
 *     (value, price or mid price): up = max(x - previous, 0),
 *     down = max(previous - x, 0). The first sample yields (0, 0). On such a
 *     single-valued series the TVI reduces to a True Strength Index of the
 *     one-step momentum.
 *
 * Priming convention (book / EasyLanguage): each EMA stage seeds on its first
 * received value, so there is no NaN warm-up region and the output is finite for
 * every update. Division guard: denominator 0 (a fully flat market) -> 0.0.
 *
 * Reference:
 *
 * Blau, William (1995). Momentum, Direction, and Divergence, ch. 4, ch. 10. Wiley.
 */
export class TickVolumeIndicator implements Indicator {

  private readonly upR: Ema;
  private readonly upS: Ema;
  private readonly upU: Ema;
  private readonly downR: Ema;
  private readonly downS: Ema;
  private readonly downU: Ema;

  private previous = 0;
  private hasPrevious = false;

  private primed_ = false;

  private readonly mnemonic_: string;
  private readonly description_: string;

  constructor(params?: TickVolumeIndicatorParams) {
    const p = params ?? {};

    const r = Math.floor(p.r ?? 12);
    const s = Math.floor(p.s ?? 12);
    const u = Math.floor(p.u ?? 1);

    if (r < 1) {
      throw new Error('invalid tick volume indicator parameters: r should be greater than 0');
    }

    if (s < 1) {
      throw new Error('invalid tick volume indicator parameters: s should be greater than 0');
    }

    if (u < 1) {
      throw new Error('invalid tick volume indicator parameters: u should be greater than 0');
    }

    this.upR = new Ema(r);
    this.upS = new Ema(s);
    this.upU = new Ema(u);
    this.downR = new Ema(r);
    this.downS = new Ema(s);
    this.downU = new Ema(u);

    this.mnemonic_ = tickVolumeIndicatorMnemonic(r, s, u);
    this.description_ = 'Tick Volume Indicator ' + this.mnemonic_;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.primed_;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.TickVolumeIndicator,
      this.mnemonic_,
      this.description_,
      [
        { mnemonic: this.mnemonic_, description: this.description_ },
      ],
    );
  }

  /**
   * Updates the indicator given the next bar's upticks and downticks.
   * Returns the TVI value.
   */
  public update(upticks: number, downticks: number): number {
    // Upticks cascade: TEMA(up, r, s, u).
    const up = this.upU.update(this.upS.update(this.upR.update(upticks)));
    // Downticks cascade: TEMA(down, r, s, u).
    const down = this.downU.update(this.downS.update(this.downR.update(downticks)));

    this.primed_ = true;

    // Division guard (Appendix B): fully flat smoothed volume -> 0.0.
    const denominator = up + down;
    if (denominator === 0) {
      return 0;
    }

    return (100 * (up - down)) / denominator;
  }

  /**
   * Derives the upticks and downticks from the change of the value vs the
   * previous value and updates the indicator.
   */
  private updateTickRule(value: number): number {
    let upticks = 0;
    let downticks = 0;
    if (this.hasPrevious) {
      const diff = value - this.previous;
      upticks = Math.max(diff, 0);
      downticks = Math.max(-diff, 0);
    }

    this.previous = value;
    this.hasPrevious = true;

    return this.update(upticks, downticks);
  }

  /** Wraps the TVI value into the output. */
  private wrap(time: Date, value: number): IndicatorOutput {
    const s0 = new Scalar(); s0.time = time; s0.value = value;

    return [s0];
  }

  /**
   * Updates the indicator given the next scalar sample.
   *
   * A scalar carries a single value, so the tick rule is applied to it.
   */
  public updateScalar(sample: Scalar): IndicatorOutput {
    return this.wrap(sample.time, this.updateTickRule(sample.value));
  }

  /**
   * Updates the indicator given the next bar sample.
   *
   * A bar splits its range: up = close - low, down = high - close.
   */
  public updateBar(sample: Bar): IndicatorOutput {
    return this.wrap(sample.time, this.update(sample.close - sample.low, sample.high - sample.close));
  }

  /**
   * Updates the indicator given the next quote sample.
   *
   * A quote applies the tick rule to its mid price.
   */
  public updateQuote(sample: Quote): IndicatorOutput {
    return this.wrap(sample.time, this.updateTickRule(sample.mid()));
  }

  /**
   * Updates the indicator given the next trade sample.
   *
   * A trade applies the tick rule to its price.
   */
  public updateTrade(sample: Trade): IndicatorOutput {
    return this.wrap(sample.time, this.updateTickRule(sample.price));
  }
}
