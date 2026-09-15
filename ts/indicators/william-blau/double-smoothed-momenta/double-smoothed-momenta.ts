import { buildMetadata } from '../../core/build-metadata';
import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { DoubleSmoothedMomentaParams } from './params';

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

/** Function to calculate mnemonic of a __DoubleSmoothedMomenta__ indicator. */
export const doubleSmoothedMomentaMnemonic = (
  a: number, y: number, z: number,
  barComponent?: BarComponent, quoteComponent?: QuoteComponent, tradeComponent?: TradeComponent,
): string => {
  const cm = componentTripleMnemonic(barComponent, quoteComponent, tradeComponent);

  return `dm(${a},${y},${z}${cm})`;
};

/**
 * DoubleSmoothedMomenta is William Blau's Double-Smoothed Momenta (DM) indicator,
 * also known as the Double-Smoothed RSI (DRSI) when the look-back is fixed at 2.
 *
 * A close-based, double-smoothed momentum oscillator bounded to [0, 100]:
 *
 *   LCa_k = min(Close over the last a bars)          (lowest close)
 *   HCa_k = max(Close over the last a bars)          (highest close)
 *   st_k  = Close_k - LCa_k                          (close above the low close)
 *   rng_k = HCa_k - LCa_k                            (a-bar close range)
 *
 *   DM(a, y, z) = 100 * EMA(EMA(st, y), z) / EMA(EMA(rng, y), z)
 *
 * Each of the numerator (st) and denominator (rng) series is double-smoothed by an
 * inner EMA of period y then an outer EMA of period z (Blau's Ez(Ey(.)) ), and the
 * ratio is scaled by 100.
 *
 * This is structurally the Double-Smoothed Stochastic computed on the CLOSE -- it
 * uses the highest/lowest close over a bars instead of the high/low of the bar --
 * and it has no signal line, so it produces a single scalar output per bar.
 *
 * Named instances:
 *   - RSI equivalence:     DM(2, 1, z) == RSI(z), the EMA-form RSI;
 *   - Double-smoothed RSI: DRSI(y, z) = DM(2, y, z).
 *
 * The equivalence DM(2,1,z) == RSI(z) holds for the EMA-form RSI built from the
 * Blau EMA (alpha = 2/(z+1)), NOT Wilder's classic RSI (which uses RMA smoothing,
 * alpha = 1/z).
 *
 * Priming convention (book / EasyLanguage): st/rng are valid once a closes exist
 * (bar a-1); all four EMA stages seed there. DM is NaN for bars 0..a-2 and finite
 * from bar a-1. For a == 1 there is no NaN warm-up, but the a-bar close range is
 * then always 0, so DM is 0.0 on every bar via the guard (a degenerate setting).
 *
 * Division guard: EMA(EMA(rng)) <= 0 -> DM = 0.0.
 *
 * Reference:
 *
 * Blau, William (1995). Momentum, Direction, and Divergence. Wiley.
 */
export class DoubleSmoothedMomenta extends LineIndicator {

  private readonly window: Array<number>;
  private readonly windowLength: number;
  private windowCount: number;
  private readonly lastIndex: number;

  private readonly numeratorY: Ema;
  private readonly numeratorZ: Ema;
  private readonly denominatorY: Ema;
  private readonly denominatorZ: Ema;

  public constructor(params?: DoubleSmoothedMomentaParams) {
    super();

    const p = params ?? {};

    const a = Math.floor(p.a ?? 2);
    const y = Math.floor(p.y ?? 2);
    const z = Math.floor(p.z ?? 14);

    if (a < 1) {
      throw new Error('a should be greater than 0');
    }

    if (y < 1) {
      throw new Error('y should be greater than 0');
    }

    if (z < 1) {
      throw new Error('z should be greater than 0');
    }

    this.mnemonic = doubleSmoothedMomentaMnemonic(
      a, y, z,
      p.barComponent, p.quoteComponent, p.tradeComponent,
    );
    this.description = 'Double-Smoothed Momenta ' + this.mnemonic;
    this.barComponent = p.barComponent;
    this.quoteComponent = p.quoteComponent;
    this.tradeComponent = p.tradeComponent;

    // Rolling window of the last a closes (for the highest/lowest close).
    this.window = new Array<number>(a).fill(0);
    this.windowLength = a;
    this.windowCount = 0;
    this.lastIndex = a - 1;

    // Two independent 2-stage EMA cascades (double smoothing), each wired
    // inner(y) -> outer(z): EMA(EMA(x, y), z).
    this.numeratorY = new Ema(y);
    this.numeratorZ = new Ema(z);
    this.denominatorY = new Ema(y);
    this.denominatorZ = new Ema(z);

    this.primed = false;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.DoubleSmoothedMomenta,
      this.mnemonic,
      this.description,
      [{ mnemonic: this.mnemonic, description: this.description }],
    );
  }

  /**
   * Updates the value of the indicator given the next sample.
   *
   * The indicator is not primed during the first a-1 updates.
   */
  public update(sample: number): number {
    if (this.primed) {
      for (let i = 0; i < this.lastIndex; i++) {
        this.window[i] = this.window[i + 1];
      }

      this.window[this.lastIndex] = sample;
    } else {
      this.window[this.windowCount] = sample;
      this.windowCount++;

      // Need a closes before the highest/lowest close is defined. While
      // unprimed, the EMA cascades must NOT advance (they seed at bar a-1).
      if (this.windowLength > this.windowCount) {
        return Number.NaN;
      }

      this.primed = true;
    }

    // Highest/lowest close over the last a bars.
    let hc = this.window[0];
    let lc = this.window[0];

    for (let i = 1; i < this.windowLength; i++) {
      const v = this.window[i];

      if (v > hc) {
        hc = v;
      }

      if (v < lc) {
        lc = v;
      }
    }

    // Raw close-above-low (>= 0) and a-bar close range (>= 0).
    const st = sample - lc;
    const rng = hc - lc;

    // Double-smooth each separately (inner y, then outer z), then divide.
    const num = this.numeratorZ.update(this.numeratorY.update(st));
    const den = this.denominatorZ.update(this.denominatorY.update(rng));

    // Division guard: smoothed range <= 0 -> DM = 0.0.
    return den <= 0 ? 0 : 100 * num / den;
  }
}
