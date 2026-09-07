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
import { TrueStrengthIndex } from '../true-strength-index/true-strength-index';
import { ErgodicOscillatorParams } from './params';

/** Function to calculate mnemonic of an __ErgodicOscillator__ indicator. */
export const ergodicOscillatorMnemonic = (
  q: number, r: number, s: number, u: number, ul: number,
  barComponent?: BarComponent, quoteComponent?: QuoteComponent, tradeComponent?: TradeComponent,
): string => {
  const cm = componentTripleMnemonic(barComponent, quoteComponent, tradeComponent);

  return `ergodic(${q},${r},${s},${u},${ul}${cm})`;
};

/**
 * ErgodicOscillator is William Blau's Ergodic Oscillator indicator.
 *
 * The Ergodic is the True Strength Index plotted together with a signal line --
 * the EMA of the oscillator that Blau introduces as the trading vehicle for the
 * TSI (ch. 2, Fig. 2-14):
 *
 *   ergodic_k = TSI(q, r, s, u)_k                    (the oscillator)
 *   signal_k  = EMA(ergodic, ul)_k                   (ul-period EMA of it)
 *
 * The indicator produces two outputs:
 *   - Ergodic: the oscillator, range [-100, +100], NaN during warm-up (bars 0..q-2);
 *   - Signal: the ul-period EMA of the oscillator.
 *
 * The numerics are exactly those of the True Strength Index, so this indicator
 * wraps a `TrueStrengthIndex` instance instead of duplicating the triple EMA
 * cascade. Only the mnemonic, the identifier, and the output naming differ.
 *
 * Priming convention (book / EasyLanguage): both outputs are NaN for bars
 * 0..q-2 and finite from bar q-1 onward; the signal EMA seeds on the first
 * finite oscillator value. Division guard: denominator 0 -> oscillator 0.0.
 *
 * Reference:
 *
 * Blau, William (1995). Momentum, Direction, and Divergence, ch. 2. Wiley.
 */
export class ErgodicOscillator implements Indicator {

  private readonly tsi: TrueStrengthIndex;

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  private readonly mnemonic_: string;
  private readonly description_: string;

  constructor(params?: ErgodicOscillatorParams) {
    const p = params ?? {};

    const q = Math.floor(p.q ?? 2);
    const r = Math.floor(p.r ?? 20);
    const s = Math.floor(p.s ?? 5);
    const u = Math.floor(p.u ?? 3);
    const ul = Math.floor(p.ul ?? 3);

    const bc = p.barComponent ?? DefaultBarComponent;
    const qc = p.quoteComponent ?? DefaultQuoteComponent;
    const tc = p.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    // The oscillator and its signal line are exactly the True Strength Index
    // outputs; wrap an instance rather than duplicating its numerics. The
    // resolved components are passed through so both agree on the mnemonic.
    // Parameter validation is performed by the wrapped indicator.
    this.tsi = new TrueStrengthIndex({
      q, r, s, u, ul,
      barComponent: bc, quoteComponent: qc, tradeComponent: tc,
    });

    this.mnemonic_ = ergodicOscillatorMnemonic(
      q, r, s, u, ul,
      p.barComponent, p.quoteComponent, p.tradeComponent,
    );
    this.description_ = 'Ergodic Oscillator ' + this.mnemonic_;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.tsi.isPrimed();
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.ErgodicOscillator,
      this.mnemonic_,
      this.description_,
      [
        { mnemonic: this.mnemonic_ + ' ergodic', description: this.description_ + ' ergodic' },
        { mnemonic: this.mnemonic_ + ' signal', description: this.description_ + ' signal' },
      ],
    );
  }

  /**
   * Updates the indicator given the next sample value.
   * Returns [ergodic, signal].
   */
  public update(sample: number): [number, number] {
    return this.tsi.update(sample);
  }

  /** Updates the indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    const [ergodic, signal] = this.update(sample.value);

    const s0 = new Scalar(); s0.time = sample.time; s0.value = ergodic;
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
