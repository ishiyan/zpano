import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Specifies the base oscillator the Nonambiguous Trend Filter is applied to. */
export enum NonambiguousTrendFilterBase {
  /** True Strength Index (named instance TSI_Trade); defaults q=2, r=32, s=13, u=3. */
  Tsi = 0,

  /** Stochastic Momentum Index (named instance SMI_Trade); defaults q=32, r=64, s=7, u=1. */
  Smi = 1,

  /** Directional Trend Index (named instance DTI_Trade); defaults q=2, r=28, s=28, u=5. */
  Dti = 2,

  /** Tick Volume Indicator (named instance TVI_Trade); defaults r=32, s=32, u=5. */
  Tvi = 3,

  /** Mean Deviation Index (MDI_Trade); defaults r=20, s=5, u=3. */
  Mdi = 4,

  /** Candlestick Momentum Index (CMI_Trade); defaults r=20, s=5, u=3. */
  Cmi = 5,

  /** Candlestick Strength Index (CSI_Trade); defaults r=32, s=32, u=1. */
  Csi = 6,
}

/**
 * Describes parameters to create an instance of the Nonambiguous Trend Filter indicator.
 *
 * The filter itself is parameterless; `q`, `r`, `s` and `u` are the canonical symbols of
 * the base oscillator from William Blau's Momentum, Direction, and Divergence (Wiley,
 * 1995). An _undefined_ or zero value selects the book default of the selected base.
 */
export interface NonambiguousTrendFilterParams {
    /**
     * The base oscillator the filter is applied to.
     *
     * If _undefined_, the True Strength Index (Tsi) is used.
     */
    base?: NonambiguousTrendFilterBase;

    /**
     * The momentum look-back period of the base (Tsi, Smi and Dti only).
     *
     * If _undefined_ or zero, the base default is used. A non-zero value should be
     * greater than 0.
     */
    q?: number;

    /**
     * The period of the 1st EMA of the base smoothing cascade.
     *
     * If _undefined_ or zero, the base default is used. A non-zero value should be
     * greater than 0.
     */
    r?: number;

    /**
     * The period of the 2nd EMA of the base smoothing cascade.
     *
     * If _undefined_ or zero, the base default is used. A non-zero value should be
     * greater than 0.
     */
    s?: number;

    /**
     * The period of the 3rd EMA of the base smoothing cascade.
     *
     * If _undefined_ or zero, the base default is used. A non-zero value should be
     * greater than 0.
     */
    u?: number;

    /**
     * A component of a bar to use when updating the indicator with a bar sample.
     *
     * Used only by the Tsi and Mdi bases. If _undefined_, the bar component will have a
     * default value (ClosePrice) and will not be shown in the indicator mnemonic.
     */
    barComponent?: BarComponent;

    /**
     * A component of a quote to use when updating the indicator with a quote sample.
     *
     * Used only by the Tsi and Mdi bases. If _undefined_, the quote component will have a
     * default value and will not be shown in the indicator mnemonic.
     */
    quoteComponent?: QuoteComponent;

    /**
     * A component of a trade to use when updating the indicator with a trade sample.
     *
     * Used only by the Tsi and Mdi bases. If _undefined_, the trade component will have a
     * default value and will not be shown in the indicator mnemonic.
     */
    tradeComponent?: TradeComponent;
}

/**
 * Returns the default parameters. The base periods are left undefined so that they
 * resolve to the book defaults of whichever base is selected.
 */
export function defaultParams(): NonambiguousTrendFilterParams {
    return { base: NonambiguousTrendFilterBase.Tsi };
}
