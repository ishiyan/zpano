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
import { SchaffTrendCycleParams } from './params';

/**
 * Stateful streaming EMA: alpha = 2/(period+1), seeds e0 = x0.
 *
 * Inlined verbatim from the Blau exponential moving average so the indicator is a
 * standalone porting unit. Do NOT change its numerics.
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

/** A fixed-capacity ring buffer of the last n values, providing min/max. */
class Window {
  private readonly data: number[];
  private readonly size: number;
  private pos = 0;
  private count = 0;

  constructor(n: number) {
    this.size = n;
    this.data = new Array<number>(n).fill(0);
  }

  public push(v: number): void {
    this.data[this.pos] = v;
    this.pos = (this.pos + 1) % this.size;
    if (this.count < this.size) {
      this.count++;
    }
  }

  public minMax(): [number, number] {
    let minVal = this.data[0];
    let maxVal = this.data[0];
    for (let i = 1; i < this.count; i++) {
      const v = this.data[i];
      if (v < minVal) {
        minVal = v;
      }
      if (v > maxVal) {
        maxVal = v;
      }
    }
    return [minVal, maxVal];
  }
}

/** Function to calculate mnemonic of a __SchaffTrendCycle__ indicator. */
export const schaffTrendCycleMnemonic = (
  fast: number, slow: number, tclen: number, factor: number,
  barComponent?: BarComponent, quoteComponent?: QuoteComponent, tradeComponent?: TradeComponent,
): string => {
  const cm = componentTripleMnemonic(barComponent, quoteComponent, tradeComponent);

  return `stc(${fast},${slow},${tclen},${factor.toFixed(2)}${cm})`;
};

/**
 * SchaffTrendCycle is Doug Schaff's Schaff Trend Cycle (STC) indicator.
 *
 * STC runs a MACD line through two cascaded stochastics, each followed by an
 * EMA-style smoothing, producing a cyclical oscillator bounded to [0, 100].
 *
 * The indicator produces three outputs:
 *   - STC: the oscillator, range [0, 100], NaN during warm-up (bars 0..slow);
 *   - MACD: the gated MACD line XMAC (0.0 pre-gate), exposed for stage testing;
 *   - PF: the first smoothed %D (0.0 pre-gate), exposed for stage testing.
 *
 * Reference:
 *
 * Malagrida, F. (2017). Schaff Trend Cycle (schaff-trend-cycle2), ProRealCode.
 */
export class SchaffTrendCycle implements Indicator {

  private readonly emaFast: Ema;
  private readonly emaSlow: Ema;

  private readonly slow: number;
  private readonly tclen: number;
  private readonly factor: number;

  private bar = -1;

  private readonly macdWin: Window;
  private readonly pfWin: Window;

  private frac1 = 0;
  private frac2 = 0;
  private pf = 0;
  private pff = 0;

  private primed_ = false;

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  private readonly mnemonic_: string;
  private readonly description_: string;

  constructor(params?: SchaffTrendCycleParams) {
    const p = params ?? {};

    const fast = Math.floor(p.fast ?? 23);
    const slow = Math.floor(p.slow ?? 50);
    const tclen = Math.floor(p.tclen ?? 10);
    const factor = p.factor ?? 0.5;

    if (fast < 1) {
      throw new Error('fast should be greater than 0');
    }

    if (slow < 1) {
      throw new Error('slow should be greater than 0');
    }

    if (tclen < 1) {
      throw new Error('tclen should be greater than 0');
    }

    if (factor <= 0 || factor > 1) {
      throw new Error('factor should be in (0, 1]');
    }

    const bc = p.barComponent ?? DefaultBarComponent;
    const qc = p.quoteComponent ?? DefaultQuoteComponent;
    const tc = p.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    this.slow = slow;
    this.tclen = tclen;
    this.factor = factor;

    this.emaFast = new Ema(fast);
    this.emaSlow = new Ema(slow);

    this.macdWin = new Window(tclen);
    this.pfWin = new Window(tclen);

    this.mnemonic_ = schaffTrendCycleMnemonic(
      fast, slow, tclen, factor,
      p.barComponent, p.quoteComponent, p.tradeComponent,
    );
    this.description_ = 'Schaff Trend Cycle ' + this.mnemonic_;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.primed_;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.SchaffTrendCycle,
      this.mnemonic_,
      this.description_,
      [
        { mnemonic: this.mnemonic_ + ' stc', description: this.description_ + ' STC' },
        { mnemonic: this.mnemonic_ + ' macd', description: this.description_ + ' MACD' },
        { mnemonic: this.mnemonic_ + ' pf', description: this.description_ + ' PF' },
      ],
    );
  }

  /**
   * Updates the indicator given the next sample value.
   * Returns [stc, macd, pf].
   */
  public update(sample: number): [number, number, number] {
    this.bar++;
    const k = this.bar;

    // Price EMAs always advance (they accumulate over the full history).
    const emaFast = this.emaFast.update(sample);
    const emaSlow = this.emaSlow.update(sample);

    // GATE: XMAC is only assigned while barindex > slow.
    const gateOpen = k > this.slow;
    const macd = gateOpen ? emaFast - emaSlow : 0;
    this.macdWin.push(macd);

    if (!gateOpen) {
      this.pfWin.push(this.pf);
      return [NaN, macd, this.pf];
    }

    // 1st stochastic of the MACD over tclen (guard on the range).
    const [ll1, hh1] = this.macdWin.minMax();
    const rng1 = hh1 - ll1;
    if (rng1 > 0) {
      this.frac1 = ((macd - ll1) / rng1) * 100;
    }

    // 1st smoothing: PF = EMA(Frac1, alpha=factor), seed 0.
    this.pf = this.pf + this.factor * (this.frac1 - this.pf);
    this.pfWin.push(this.pf);

    // 2nd stochastic of PF over tclen.
    const [ll2, hh2] = this.pfWin.minMax();
    const rng2 = hh2 - ll2;
    if (rng2 > 0) {
      this.frac2 = ((this.pf - ll2) / rng2) * 100;
    }

    // 2nd smoothing: STC = PFF = EMA(Frac2, alpha=factor), seed 0.
    this.pff = this.pff + this.factor * (this.frac2 - this.pff);
    this.primed_ = true;

    return [this.pff, macd, this.pf];
  }

  /** Updates the indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    const [stc, macd, pf] = this.update(sample.value);

    const s0 = new Scalar(); s0.time = sample.time; s0.value = stc;
    const s1 = new Scalar(); s1.time = sample.time; s1.value = macd;
    const s2 = new Scalar(); s2.time = sample.time; s2.value = pf;

    return [s0, s1, s2];
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
