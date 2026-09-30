import { buildMetadata } from '../../core/build-metadata';
import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { Indicator } from '../../core/indicator';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { TrueStrengthIndex } from '../true-strength-index/true-strength-index';
import { StochasticMomentumIndex } from '../stochastic-momentum-index/stochastic-momentum-index';
import { DirectionalTrendIndex } from '../directional-trend-index/directional-trend-index';
import { TickVolumeIndicator } from '../tick-volume-indicator/tick-volume-indicator';
import { MeanDeviationIndex } from '../mean-deviation-index/mean-deviation-index';
import { CandlestickMomentumIndex } from '../candlestick-momentum-index/candlestick-momentum-index';
import { CandlestickStrengthIndex } from '../candlestick-strength-index/candlestick-strength-index';
import { NonambiguousTrendFilterBase, NonambiguousTrendFilterParams } from './params';

/** The book named-instance defaults of a base. */
interface BaseDefaults {
  name: string;
  q: number;
  r: number;
  s: number;
  u: number;
  usesQ: boolean;
}

const baseDefaults: Record<NonambiguousTrendFilterBase, BaseDefaults> = {
  [NonambiguousTrendFilterBase.Tsi]: { name: 'tsi', q: 2, r: 32, s: 13, u: 3, usesQ: true },
  [NonambiguousTrendFilterBase.Smi]: { name: 'smi', q: 32, r: 64, s: 7, u: 1, usesQ: true },
  [NonambiguousTrendFilterBase.Dti]: { name: 'dti', q: 2, r: 28, s: 28, u: 5, usesQ: true },
  [NonambiguousTrendFilterBase.Tvi]: { name: 'tvi', q: 0, r: 32, s: 32, u: 5, usesQ: false },
  [NonambiguousTrendFilterBase.Mdi]: { name: 'mdi', q: 0, r: 20, s: 5, u: 3, usesQ: false },
  [NonambiguousTrendFilterBase.Cmi]: { name: 'cmi', q: 0, r: 20, s: 5, u: 3, usesQ: false },
  [NonambiguousTrendFilterBase.Csi]: { name: 'csi', q: 0, r: 32, s: 32, u: 1, usesQ: false },
};

/** Returns the value if it is set and non-zero, otherwise the default. */
const orDefault = (value: number | undefined, def: number): number => {
  const v = Math.floor(value ?? 0);
  return v === 0 ? def : v;
};

/**
 * NonambiguousTrendFilter is William Blau's Nonambiguous Trend Filter (_Trade).
 *
 * A post-processing transform applied to a normalized, signed base oscillator X
 * (TSI, SMI, DTI, TVI, MDI, CMI, CSI). It keeps X only where its sign and slope
 * agree and zeroes every ambiguous bar (book Ch. 8, Appendix B Figs. B-20..B-23):
 *
 *   X_Trade[k] = X[k]   if X[k] > 0 and X[k] - X[k-1] > 0   (positive and rising)
 *              = X[k]   if X[k] < 0 and X[k] - X[k-1] < 0   (negative and falling)
 *              = 0      otherwise                           (ambiguous)
 *
 * The nonzero stretches correspond one-to-one with genuine up/down trends;
 * congestion and flat regions are blanked to zero.
 *
 * The filter wraps an instance of the base indicator: every sample is routed to the
 * base with its own entity mapping, and the base's primary output is filtered.
 *
 * Conventions: a NaN base value (the base's own look-back warm-up) yields NaN and
 * leaves the filter state untouched; the first finite base value has no prior slope,
 * so the output is 0.0; a flat step (delta == 0) is neither rising nor falling, so it
 * is zeroed.
 *
 * Reference:
 *
 * Blau, William (1995). Momentum, Direction, and Divergence, ch. 8, Appendix B
 * Figs. B-20..B-23. Wiley.
 */
export class NonambiguousTrendFilter implements Indicator {

  private readonly base: Indicator;

  private previous = 0;
  private primed_ = false;

  private readonly mnemonic_: string;
  private readonly description_: string;

  constructor(params?: NonambiguousTrendFilterParams) {
    const p = params ?? {};

    const base = p.base ?? NonambiguousTrendFilterBase.Tsi;
    const d = baseDefaults[base];
    if (d === undefined) {
      throw new Error(`invalid nonambiguous trend filter parameters: unknown base ${base}`);
    }

    const q = orDefault(p.q, d.q);
    const r = orDefault(p.r, d.r);
    const s = orDefault(p.s, d.s);
    const u = orDefault(p.u, d.u);

    if (d.usesQ && q < 1) {
      throw new Error('invalid nonambiguous trend filter parameters: q should be greater than 0');
    }

    if (r < 1) {
      throw new Error('invalid nonambiguous trend filter parameters: r should be greater than 0');
    }

    if (s < 1) {
      throw new Error('invalid nonambiguous trend filter parameters: s should be greater than 0');
    }

    if (u < 1) {
      throw new Error('invalid nonambiguous trend filter parameters: u should be greater than 0');
    }

    // Price components are meaningful only for the single-price Tsi and Mdi bases.
    const usesComponents = base === NonambiguousTrendFilterBase.Tsi || base === NonambiguousTrendFilterBase.Mdi;
    const barComponent = usesComponents ? p.barComponent : undefined;
    const quoteComponent = usesComponents ? p.quoteComponent : undefined;
    const tradeComponent = usesComponents ? p.tradeComponent : undefined;

    // The base signal line is unused, so its period is 1.
    switch (base) {
      case NonambiguousTrendFilterBase.Tsi:
        this.base = new TrueStrengthIndex({ q, r, s, u, ul: 1, barComponent, quoteComponent, tradeComponent });
        break;
      case NonambiguousTrendFilterBase.Smi:
        this.base = new StochasticMomentumIndex({ q, r, s, u, ul: 1 });
        break;
      case NonambiguousTrendFilterBase.Dti:
        this.base = new DirectionalTrendIndex({ q, r, s, u, ul: 1 });
        break;
      case NonambiguousTrendFilterBase.Tvi:
        this.base = new TickVolumeIndicator({ r, s, u });
        break;
      case NonambiguousTrendFilterBase.Mdi:
        this.base = new MeanDeviationIndex({ r, s, u, ul: 1, barComponent, quoteComponent, tradeComponent });
        break;
      case NonambiguousTrendFilterBase.Cmi:
        this.base = new CandlestickMomentumIndex({ r, s, u, ul: 1 });
        break;
      default:
        this.base = new CandlestickStrengthIndex({ r, s, u, ul: 1 });
        break;
    }

    const cm = componentTripleMnemonic(barComponent, quoteComponent, tradeComponent);
    this.mnemonic_ = d.usesQ
      ? `ntf.${d.name}(${q},${r},${s},${u}${cm})`
      : `ntf.${d.name}(${r},${s},${u}${cm})`;
    this.description_ = 'Nonambiguous Trend Filter ' + this.mnemonic_;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.primed_;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.NonambiguousTrendFilter,
      this.mnemonic_,
      this.description_,
      [
        { mnemonic: this.mnemonic_, description: this.description_ },
      ],
    );
  }

  /** Keeps x when positive-and-rising or negative-and-falling, else 0. */
  private filter(x: number): number {
    if (Number.isNaN(x)) {
      // The base is still warming up: do not touch the filter state.
      return Number.NaN;
    }

    if (!this.primed_) {
      // First finite value: no prior slope, hence ambiguous.
      this.previous = x;
      this.primed_ = true;
      return 0;
    }

    const delta = x - this.previous;
    this.previous = x;

    if (x > 0 && delta > 0) {
      return x; // positive and rising
    }

    if (x < 0 && delta < 0) {
      return x; // negative and falling
    }

    return 0; // ambiguous / flat / congestion
  }

  /** Filters the primary output of the base and wraps it into the output. */
  private wrap(time: Date, output: IndicatorOutput): IndicatorOutput {
    const s0 = new Scalar(); s0.time = time; s0.value = this.filter((output[0] as Scalar).value);

    return [s0];
  }

  /** Updates the indicator given the next single sample value, fed to the base as a scalar. */
  public update(sample: number): number {
    const s = new Scalar(); s.time = new Date(0); s.value = sample;

    return this.filter((this.base.updateScalar(s)[0] as Scalar).value);
  }

  /** Updates the indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    return this.wrap(sample.time, this.base.updateScalar(sample));
  }

  /** Updates the indicator given the next bar sample. */
  public updateBar(sample: Bar): IndicatorOutput {
    return this.wrap(sample.time, this.base.updateBar(sample));
  }

  /** Updates the indicator given the next quote sample. */
  public updateQuote(sample: Quote): IndicatorOutput {
    return this.wrap(sample.time, this.base.updateQuote(sample));
  }

  /** Updates the indicator given the next trade sample. */
  public updateTrade(sample: Trade): IndicatorOutput {
    return this.wrap(sample.time, this.base.updateTrade(sample));
  }
}
