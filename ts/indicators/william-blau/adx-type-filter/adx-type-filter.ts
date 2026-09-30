import { buildMetadata } from '../../core/build-metadata';
import { Bar } from '../../../entities/bar';
import { DefaultBarComponent, barComponentValue } from '../../../entities/bar-component';
import { Quote } from '../../../entities/quote';
import { DefaultQuoteComponent, quoteComponentValue } from '../../../entities/quote-component';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { DefaultTradeComponent, tradeComponentValue } from '../../../entities/trade-component';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { Indicator } from '../../core/indicator';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { TrueStrengthIndex } from '../true-strength-index/true-strength-index';
import { AdxTypeFilterParams, AdxTypeFilterSource } from './params';

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

/** Appends a sample to a rolling window holding at most `length` values. */
const push = (window: number[], length: number, sample: number): void => {
  if (window.length >= length) {
    window.shift();
  }
  window.push(sample);
};

const sourceMnemonics: Record<AdxTypeFilterSource, string> = {
  [AdxTypeFilterSource.TsiMomentum]: 'tsi',
  [AdxTypeFilterSource.SmiMomentum]: 'smi',
  [AdxTypeFilterSource.DtiMomentum]: 'dti',
  [AdxTypeFilterSource.TviBalance]: 'tvi',
  [AdxTypeFilterSource.TsiNormalized]: 'tsin',
};

/**
 * AdxTypeFilter is William Blau's ADX-Type Filter (ATF).
 *
 * A non-negative trend-strength filter, analogous to Wilder's ADX, built by
 * rectifying and double-smoothing a bipolar momentum series (book Fig. B-24):
 *
 *   ATF(Price, r, s) = EMA(|EMA(Price, r)|, s)
 *
 * The inner EMA(r) smooths the signed momentum, the absolute value discards the
 * direction and keeps the amplitude, and the outer EMA(s) smooths the amplitude.
 * A rising ATF signals a strengthening trend, a falling ATF a ranging market.
 *
 * The bipolar momentum is selected by the source:
 *   - TsiMomentum:   C - C[q-1]                               (TSI_ATF);
 *   - SmiMomentum:   C - 0.5*(HH(q) + LL(q))                  (SMI_ATF);
 *   - DtiMomentum:   max(H - H[q-1], 0) - max(L[q-1] - L, 0);
 *   - TviBalance:    upticks - downticks (2C - H - L for a bar);
 *   - TsiNormalized: TSI(q, r, 1, 1), which replaces the inner EMA (r = 1).
 *
 * Priming convention (book / EasyLanguage): each EMA stage seeds on its first
 * finite momentum value. A NaN momentum (the q-bar look-back warm-up) is
 * propagated: the output is NaN and the EMAs do not advance. The output is
 * always >= 0.
 *
 * Reference:
 *
 * Blau, William (1995). Momentum, Direction, and Divergence, Appendix B Fig. B-24. Wiley.
 */
export class AdxTypeFilter implements Indicator {

  private readonly source: AdxTypeFilterSource;
  private readonly usesComponents: boolean;
  private readonly q: number;

  private readonly closes: number[] = [];
  private readonly highs: number[] = [];
  private readonly lows: number[] = [];

  private previous = 0;
  private hasPrevious = false;

  private readonly tsi: TrueStrengthIndex | undefined;
  private readonly inner: Ema;
  private readonly outer: Ema;

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  private primed_ = false;

  private readonly mnemonic_: string;
  private readonly description_: string;

  constructor(params?: AdxTypeFilterParams) {
    const p = params ?? {};

    const source = p.source ?? AdxTypeFilterSource.TsiMomentum;
    const name = sourceMnemonics[source];
    if (name === undefined) {
      throw new Error(`invalid adx type filter parameters: unknown source ${source}`);
    }

    let q = Math.floor(p.q ?? 0);
    if (q === 0) {
      q = source === AdxTypeFilterSource.SmiMomentum ? 32 : 2;
    }

    const r = Math.floor(p.r ?? 32);
    const s = Math.floor(p.s ?? 32);

    if (q < 1) {
      throw new Error('invalid adx type filter parameters: q should be greater than 0');
    }

    if (r < 1) {
      throw new Error('invalid adx type filter parameters: r should be greater than 0');
    }

    if (s < 1) {
      throw new Error('invalid adx type filter parameters: s should be greater than 0');
    }

    // Price components are meaningful only for the single-price TSI sources; the
    // other sources use the bar's high/low/close or the single value of the sample
    // (scalar value, quote mid price, trade price).
    this.usesComponents = source === AdxTypeFilterSource.TsiMomentum
      || source === AdxTypeFilterSource.TsiNormalized;

    const bc = this.usesComponents ? p.barComponent : undefined;
    const qc = this.usesComponents ? p.quoteComponent : undefined;
    const tc = this.usesComponents ? p.tradeComponent : undefined;

    this.barComponentFunc = barComponentValue(bc ?? DefaultBarComponent);
    this.quoteComponentFunc = quoteComponentValue(qc ?? DefaultQuoteComponent);
    this.tradeComponentFunc = tradeComponentValue(tc ?? DefaultTradeComponent);

    this.source = source;
    this.q = q;

    // The normalized TSI replaces the inner EMA, which becomes a passthrough.
    if (source === AdxTypeFilterSource.TsiNormalized) {
      this.tsi = new TrueStrengthIndex({ q, r, s: 1, u: 1, ul: 1 });
      this.inner = new Ema(1);
    } else {
      this.tsi = undefined;
      this.inner = new Ema(r);
    }
    this.outer = new Ema(s);

    this.mnemonic_ = source === AdxTypeFilterSource.TviBalance
      ? `atf.${name}(${r},${s})`
      : `atf.${name}(${q},${r},${s}${componentTripleMnemonic(bc, qc, tc)})`;
    this.description_ = 'ADX-Type Filter ' + this.mnemonic_;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.primed_;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.AdxTypeFilter,
      this.mnemonic_,
      this.description_,
      [
        { mnemonic: this.mnemonic_, description: this.description_ },
      ],
    );
  }

  /** Inner smooth -> rectify -> outer smooth; a NaN momentum propagates. */
  private filter(momentum: number): number {
    if (Number.isNaN(momentum)) {
      // The EMAs do not advance until the first finite momentum.
      return Number.NaN;
    }

    this.primed_ = true;

    return this.outer.update(Math.abs(this.inner.update(momentum)));
  }

  /**
   * Updates the indicator given the next single sample value.
   *
   * The SmiMomentum and DtiMomentum sources use the value as the high, the low and
   * the close; the TviBalance source applies the tick rule
   * (balance = value - previous value).
   */
  public update(sample: number): number {
    switch (this.source) {
      case AdxTypeFilterSource.TsiMomentum:
        push(this.closes, this.q, sample);
        if (this.closes.length < this.q) {
          return Number.NaN;
        }

        // mtm_k = C_k - C_(k-(q-1)); the leftmost element is C_(k-(q-1)).
        return this.filter(sample - this.closes[0]);

      case AdxTypeFilterSource.TsiNormalized: {
        const [tsi] = this.tsi!.update(sample);
        return this.filter(tsi);
      }

      case AdxTypeFilterSource.TviBalance: {
        const balance = this.hasPrevious ? sample - this.previous : 0;
        this.previous = sample;
        this.hasPrevious = true;
        return this.filter(balance);
      }

      default:
        return this.updateHighLowClose(sample, sample, sample);
    }
  }

  /**
   * Updates the indicator given the next bar's high, low and close.
   *
   * The TsiMomentum and TsiNormalized sources use the close only.
   */
  public updateHighLowClose(high: number, low: number, close: number): number {
    switch (this.source) {
      case AdxTypeFilterSource.SmiMomentum:
        push(this.highs, this.q, high);
        push(this.lows, this.q, low);
        if (this.highs.length < this.q) {
          return Number.NaN;
        }

        // sm = C - 0.5*(HH(q) + LL(q)).
        return this.filter(close - 0.5 * (Math.max(...this.highs) + Math.min(...this.lows)));

      case AdxTypeFilterSource.DtiMomentum: {
        push(this.highs, this.q, high);
        push(this.lows, this.q, low);
        if (this.highs.length < this.q) {
          return Number.NaN;
        }

        // HMU - LMD; the leftmost elements are H_(k-(q-1)) and L_(k-(q-1)).
        const hmu = Math.max(high - this.highs[0], 0);
        const lmd = Math.max(this.lows[0] - low, 0);
        return this.filter(hmu - lmd);
      }

      case AdxTypeFilterSource.TviBalance:
        // up - down = (C - L) - (H - C) = 2C - H - L.
        return this.filter(2 * close - high - low);

      default:
        return this.update(close);
    }
  }

  /** Wraps the ATF value into the output. */
  private wrap(time: Date, value: number): IndicatorOutput {
    const s0 = new Scalar(); s0.time = time; s0.value = value;

    return [s0];
  }

  /** Updates the indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    return this.wrap(sample.time, this.update(sample.value));
  }

  /**
   * Updates the indicator given the next bar sample.
   *
   * The TSI sources use the bar component; the other sources use the bar's high,
   * low and close.
   */
  public updateBar(sample: Bar): IndicatorOutput {
    const value = this.usesComponents
      ? this.update(this.barComponentFunc(sample))
      : this.updateHighLowClose(sample.high, sample.low, sample.close);

    return this.wrap(sample.time, value);
  }

  /** Updates the indicator given the next quote sample. */
  public updateQuote(sample: Quote): IndicatorOutput {
    return this.wrap(sample.time, this.update(this.quoteComponentFunc(sample)));
  }

  /** Updates the indicator given the next trade sample. */
  public updateTrade(sample: Trade): IndicatorOutput {
    return this.wrap(sample.time, this.update(this.tradeComponentFunc(sample)));
  }
}
