import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Specifies the bipolar momentum the ADX-Type Filter is applied to. */
export enum AdxTypeFilterSource {
  /** The TSI numerator `C - C[q-1]` (named instance TSI_ATF). */
  TsiMomentum = 0,

  /** The SMI raw stochastic momentum `C - 0.5*(HH(q) + LL(q))` (named instance SMI_ATF). */
  SmiMomentum = 1,

  /** The DTI numerator `max(H - H[q-1], 0) - max(L[q-1] - L, 0)`. */
  DtiMomentum = 2,

  /** The TVI tick balance `upticks - downticks` (`2C - H - L` for a bar). */
  TviBalance = 3,

  /** The single-smoothed normalized `TSI(q, r, 1, 1)`, replacing the inner EMA. */
  TsiNormalized = 4,
}

/**
 * Describes parameters to create an instance of the ADX-Type Filter indicator.
 *
 * The parameter names `q`, `r` and `s` are the canonical symbols from William Blau's
 * Momentum, Direction, and Divergence (Wiley, 1995), Appendix B, Figure B-24. They are
 * kept verbatim for fidelity with the book and the test-data naming.
 */
export interface AdxTypeFilterParams {
    /**
     * The bipolar momentum the filter is applied to.
     *
     * If _undefined_, the TSI numerator (TsiMomentum) is used.
     */
    source?: AdxTypeFilterSource;

    /**
     * The momentum look-back period.
     *
     * If _undefined_ or zero, the source default is used: 2 for TsiMomentum,
     * DtiMomentum and TsiNormalized, 32 for SmiMomentum. The look-back is not used
     * by TviBalance. A non-zero value should be greater than 0.
     */
    q?: number;

    /**
     * The period of the inner EMA, applied to the signed momentum.
     *
     * For TsiNormalized this is the smoothing period of the normalized TSI, which
     * replaces the inner EMA. The value should be greater than 0. The default value
     * is 32.
     */
    r?: number;

    /**
     * The period of the outer EMA, applied to the rectified momentum.
     *
     * The value should be greater than 0. The default value is 32.
     */
    s?: number;

    /**
     * A component of a bar to use when updating the indicator with a bar sample.
     *
     * Used only by the TsiMomentum and TsiNormalized sources. If _undefined_, the bar
     * component will have a default value (ClosePrice) and will not be shown in the
     * indicator mnemonic.
     */
    barComponent?: BarComponent;

    /**
     * A component of a quote to use when updating the indicator with a quote sample.
     *
     * Used only by the TsiMomentum and TsiNormalized sources. If _undefined_, the quote
     * component will have a default value and will not be shown in the indicator mnemonic.
     */
    quoteComponent?: QuoteComponent;

    /**
     * A component of a trade to use when updating the indicator with a trade sample.
     *
     * Used only by the TsiMomentum and TsiNormalized sources. If _undefined_, the trade
     * component will have a default value and will not be shown in the indicator mnemonic.
     */
    tradeComponent?: TradeComponent;
}

/**
 * Returns the default parameters. The look-back `q` is left undefined so that it
 * resolves to the default of whichever source is selected.
 */
export function defaultParams(): AdxTypeFilterParams {
    return { source: AdxTypeFilterSource.TsiMomentum, r: 32, s: 32 };
}
