import { buildMetadata } from '../../core/build-metadata';
import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { TrueStrengthIndex } from '../true-strength-index/true-strength-index';
import { SlopeDivergenceTsiFilterParams } from './params';

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

/** Function to calculate mnemonic of a __SlopeDivergenceTsiFilter__ indicator. */
export const slopeDivergenceTsiFilterMnemonic = (
  q: number, r: number, s: number, u: number, x: number, y: number,
  barComponent?: BarComponent, quoteComponent?: QuoteComponent, tradeComponent?: TradeComponent,
): string => {
  const cm = componentTripleMnemonic(barComponent, quoteComponent, tradeComponent);

  return `sdtsi(${q},${r},${s},${u},${x},${y}${cm})`;
};

/**
 * SlopeDivergenceTsiFilter is William Blau's Slope Divergence TSI Filter (SD_TSI).
 *
 * A trend/congestion prefilter built on the True Strength Index. It keeps the TSI
 * value only when the slope of the TSI agrees in sign with the slope of a separate
 * double EMA of price; otherwise it outputs 0 (a slope divergence, or congestion
 * zone):
 *
 *   ind_k = TSI(close, q, r, s, u)_k
 *   ref_k = DEMA(close, x, y)_k = EMA(EMA(close, x), y)_k
 *
 *   SD_TSI_k = ind_k   if ind_k - ind_(k-1) > 0 and ref_k - ref_(k-1) > 0
 *            = ind_k   if ind_k - ind_(k-1) < 0 and ref_k - ref_(k-1) < 0
 *            = 0       otherwise
 *
 * The gate is strict (book Fig. B-25): a flat slope on either series yields 0.
 * The output range is [-100, +100].
 *
 * Priming convention (book / EasyLanguage): the TSI is NaN for bars 0..q-2
 * (momentum look-back), so SD_TSI is NaN there. The price DEMA seeds at bar 0 and
 * advances every bar, including through the TSI warm-up. At the first finite TSI
 * bar there is no prior TSI value, hence no slope, so the output is 0.0.
 *
 * Reference:
 *
 * Blau, William (1995). Momentum, Direction, and Divergence, ch. 12, Appendix B
 * Fig. B-25. Wiley.
 */
export class SlopeDivergenceTsiFilter extends LineIndicator {

  private readonly tsi: TrueStrengthIndex;

  private readonly referenceX: Ema;
  private readonly referenceY: Ema;

  private previousTsi: number;
  private hasPreviousTsi: boolean;
  private previousReference: number;

  public constructor(params?: SlopeDivergenceTsiFilterParams) {
    super();

    const p = params ?? {};

    const q = Math.floor(p.q ?? 2);
    const r = Math.floor(p.r ?? 32);
    const s = Math.floor(p.s ?? 32);
    const u = Math.floor(p.u ?? 7);
    const x = Math.floor(p.x ?? 32);
    const y = Math.floor(p.y ?? 7);

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

    if (x < 1) {
      throw new Error('x should be greater than 0');
    }

    if (y < 1) {
      throw new Error('y should be greater than 0');
    }

    this.mnemonic = slopeDivergenceTsiFilterMnemonic(
      q, r, s, u, x, y,
      p.barComponent, p.quoteComponent, p.tradeComponent,
    );
    this.description = 'Slope Divergence TSI Filter ' + this.mnemonic;
    this.barComponent = p.barComponent;
    this.quoteComponent = p.quoteComponent;
    this.tradeComponent = p.tradeComponent;

    // The TSI oscillator (its signal line is unused, so ul=1).
    this.tsi = new TrueStrengthIndex({ q, r, s, u, ul: 1 });

    // Price reference: DEMA(close, x, y) = EMA(EMA(close, x), y).
    this.referenceX = new Ema(x);
    this.referenceY = new Ema(y);

    // Slope state: the previous finite TSI and the previous-bar reference.
    this.previousTsi = 0;
    this.hasPreviousTsi = false;
    this.previousReference = 0;

    this.primed = false;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.SlopeDivergenceTsiFilter,
      this.mnemonic,
      this.description,
      [{ mnemonic: this.mnemonic, description: this.description }],
    );
  }

  /**
   * Updates the value of the indicator given the next sample.
   *
   * The indicator is not primed during the first q-1 updates.
   */
  public update(sample: number): number {
    // The price reference advances every bar (it has no NaN warm-up).
    const reference = this.referenceY.update(this.referenceX.update(sample));

    const [tsi] = this.tsi.update(sample);
    if (Number.isNaN(tsi)) {
      // TSI momentum warm-up: keep the previous-bar reference current.
      this.previousReference = reference;
      return Number.NaN;
    }

    let result = 0;

    // At the first finite TSI there is no prior TSI, hence no slope.
    if (this.hasPreviousTsi) {
      const deltaTsi = tsi - this.previousTsi;
      const deltaReference = reference - this.previousReference;

      // Keep the TSI only when both slopes are strictly same-signed.
      if ((deltaTsi > 0 && deltaReference > 0) || (deltaTsi < 0 && deltaReference < 0)) {
        result = tsi;
      }
    }

    this.previousTsi = tsi;
    this.hasPreviousTsi = true;
    this.previousReference = reference;
    this.primed = true;

    return result;
  }
}
