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
import { AdaptiveTrendAndCycleFilterParams } from './params';
import {
  FATL_COEFFICIENTS, SATL_COEFFICIENTS, RFTL_COEFFICIENTS, RSTL_COEFFICIENTS, RBCI_COEFFICIENTS,
} from './coefficients';

/** Internal FIR engine shared by all five ATCF lines.
 *
 * Holds a fixed-length window (length = coefficients.length) and, once primed,
 * computes Σ window[i]·coeffs[i] on every update. Index 0 of the window holds
 * the oldest sample; the last index holds the newest. */
class FirFilter {
  private readonly window: number[];
  private readonly coeffs: readonly number[];
  private count = 0;
  private primed_ = false;
  private value_ = Number.NaN;

  public constructor(coeffs: readonly number[]) {
    this.coeffs = coeffs;
    this.window = new Array<number>(coeffs.length).fill(0);
  }

  public get primed(): boolean { return this.primed_; }
  public get value(): number { return this.value_; }

  public update(sample: number): number {
    const len = this.window.length;

    if (this.primed_) {
      // Shift left (drop oldest), append newest.
      for (let i = 0; i < len - 1; i++) this.window[i] = this.window[i + 1];
      this.window[len - 1] = sample;

      let sum = 0;
      for (let i = 0; i < len; i++) sum += this.window[i] * this.coeffs[i];
      this.value_ = sum;
      return this.value_;
    }

    this.window[this.count++] = sample;
    if (this.count === len) {
      this.primed_ = true;
      let sum = 0;
      for (let i = 0; i < len; i++) sum += this.window[i] * this.coeffs[i];
      this.value_ = sum;
    }
    return this.value_;
  }
}

/** __Adaptive Trend and Cycle Filter__ (ATCF) suite by Vladimir Kravchuk.
 *
 * A bank of five Finite Impulse Response (FIR) filters applied to the same
 * input series plus three composite outputs derived from them:
 *
 *  - FATL (Fast Adaptive Trend Line)        — 39-tap FIR.
 *  - SATL (Slow Adaptive Trend Line)        — 65-tap FIR.
 *  - RFTL (Reference Fast Trend Line)       — 44-tap FIR.
 *  - RSTL (Reference Slow Trend Line)       — 91-tap FIR.
 *  - RBCI (Range Bound Channel Index)       — 56-tap FIR.
 *  - FTLM (Fast Trend Line Momentum)        = FATL − RFTL.
 *  - STLM (Slow Trend Line Momentum)        = SATL − RSTL.
 *  - PCCI (Perfect Commodity Channel Index) = sample − FATL.
 *
 * Each FIR filter emits NaN until its own window fills; composite values emit
 * NaN until both their components are primed. Indicator-level `isPrimed()`
 * mirrors RSTL (the longest pole, 91 samples).
 *
 * Reference: Vladimir Kravchuk, "New adaptive method of following the tendency
 * and market cycles", Currency Speculator magazine, 2000. */
export class AdaptiveTrendAndCycleFilter implements Indicator {
  private readonly fatlFir: FirFilter;
  private readonly satlFir: FirFilter;
  private readonly rftlFir: FirFilter;
  private readonly rstlFir: FirFilter;
  private readonly rbciFir: FirFilter;

  private ftlmValue = Number.NaN;
  private stlmValue = Number.NaN;
  private pcciValue = Number.NaN;

  private readonly mnemonic_: string;
  private readonly description_: string;

  private readonly mnemonicFatl: string; private readonly descriptionFatl: string;
  private readonly mnemonicSatl: string; private readonly descriptionSatl: string;
  private readonly mnemonicRftl: string; private readonly descriptionRftl: string;
  private readonly mnemonicRstl: string; private readonly descriptionRstl: string;
  private readonly mnemonicRbci: string; private readonly descriptionRbci: string;
  private readonly mnemonicFtlm: string; private readonly descriptionFtlm: string;
  private readonly mnemonicStlm: string; private readonly descriptionStlm: string;
  private readonly mnemonicPcci: string; private readonly descriptionPcci: string;

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  /** Creates an instance with default parameters. */
  public static default(): AdaptiveTrendAndCycleFilter {
    return new AdaptiveTrendAndCycleFilter({});
  }

  /** Creates an instance based on the given parameters. */
  public static fromParams(params: AdaptiveTrendAndCycleFilterParams): AdaptiveTrendAndCycleFilter {
    return new AdaptiveTrendAndCycleFilter(params);
  }

  private constructor(params: AdaptiveTrendAndCycleFilterParams) {
    const bc = params.barComponent ?? DefaultBarComponent;
    const qc = params.quoteComponent ?? DefaultQuoteComponent;
    const tc = params.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    this.fatlFir = new FirFilter(FATL_COEFFICIENTS);
    this.satlFir = new FirFilter(SATL_COEFFICIENTS);
    this.rftlFir = new FirFilter(RFTL_COEFFICIENTS);
    this.rstlFir = new FirFilter(RSTL_COEFFICIENTS);
    this.rbciFir = new FirFilter(RBCI_COEFFICIENTS);

    // componentTripleMnemonic returns "" or ", <comp>[, <comp>...]". Strip the
    // leading ", " so the mnemonic reads "atcf(hl/2)" instead of "atcf(, hl/2)".
    const cm = componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent);
    const arg = cm === '' ? '' : cm.substring(2);

    this.mnemonic_ = `atcf(${arg})`;
    this.description_ = 'Adaptive trend and cycle filter ' + this.mnemonic_;

    const mk = (name: string, full: string): [string, string] => {
      const m = `${name}(${arg})`;
      return [m, `${full} ${m}`];
    };

    [this.mnemonicFatl, this.descriptionFatl] = mk('fatl', 'Fast Adaptive Trend Line');
    [this.mnemonicSatl, this.descriptionSatl] = mk('satl', 'Slow Adaptive Trend Line');
    [this.mnemonicRftl, this.descriptionRftl] = mk('rftl', 'Reference Fast Trend Line');
    [this.mnemonicRstl, this.descriptionRstl] = mk('rstl', 'Reference Slow Trend Line');
    [this.mnemonicRbci, this.descriptionRbci] = mk('rbci', 'Range Bound Channel Index');
    [this.mnemonicFtlm, this.descriptionFtlm] = mk('ftlm', 'Fast Trend Line Momentum');
    [this.mnemonicStlm, this.descriptionStlm] = mk('stlm', 'Slow Trend Line Momentum');
    [this.mnemonicPcci, this.descriptionPcci] = mk('pcci', 'Perfect Commodity Channel Index');
  }

  /** Indicates whether the indicator is primed (RSTL primed). */
  public isPrimed(): boolean { return this.rstlFir.primed; }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.AdaptiveTrendAndCycleFilter,
      this.mnemonic_,
      this.description_,
      [
        { mnemonic: this.mnemonicFatl, description: this.descriptionFatl },
        { mnemonic: this.mnemonicSatl, description: this.descriptionSatl },
        { mnemonic: this.mnemonicRftl, description: this.descriptionRftl },
        { mnemonic: this.mnemonicRstl, description: this.descriptionRstl },
        { mnemonic: this.mnemonicRbci, description: this.descriptionRbci },
        { mnemonic: this.mnemonicFtlm, description: this.descriptionFtlm },
        { mnemonic: this.mnemonicStlm, description: this.descriptionStlm },
        { mnemonic: this.mnemonicPcci, description: this.descriptionPcci },
      ],
    );
  }

  /** Feeds the next sample to all five FIR filters and recomputes the three
   * composite values. Returns the 8 outputs as a tuple in enum order:
   * [FATL, SATL, RFTL, RSTL, RBCI, FTLM, STLM, PCCI].
   *
   * NaN input leaves internal state unchanged and returns all NaN. */
  public update(sample: number): [number, number, number, number, number, number, number, number] {
    if (Number.isNaN(sample)) {
      const n = Number.NaN;
      return [n, n, n, n, n, n, n, n];
    }

    const fatl = this.fatlFir.update(sample);
    const satl = this.satlFir.update(sample);
    const rftl = this.rftlFir.update(sample);
    const rstl = this.rstlFir.update(sample);
    const rbci = this.rbciFir.update(sample);

    if (this.fatlFir.primed && this.rftlFir.primed) this.ftlmValue = fatl - rftl;
    if (this.satlFir.primed && this.rstlFir.primed) this.stlmValue = satl - rstl;
    if (this.fatlFir.primed) this.pcciValue = sample - fatl;

    return [fatl, satl, rftl, rstl, rbci, this.ftlmValue, this.stlmValue, this.pcciValue];
  }

  /** Updates the indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    return this.updateEntity(sample.time, sample.value);
  }

  /** Updates the indicator given the next bar sample. */
  public updateBar(sample: Bar): IndicatorOutput {
    return this.updateEntity(sample.time, this.barComponentFunc(sample));
  }

  /** Updates the indicator given the next quote sample. */
  public updateQuote(sample: Quote): IndicatorOutput {
    return this.updateEntity(sample.time, this.quoteComponentFunc(sample));
  }

  /** Updates the indicator given the next trade sample. */
  public updateTrade(sample: Trade): IndicatorOutput {
    return this.updateEntity(sample.time, this.tradeComponentFunc(sample));
  }

  private updateEntity(time: Date, sample: number): IndicatorOutput {
    const values = this.update(sample);
    const scalars: Scalar[] = new Array<Scalar>(8);
    for (let i = 0; i < 8; i++) {
      const s = new Scalar();
      s.time = time;
      s.value = values[i];
      scalars[i] = s;
    }
    return scalars;
  }
}
