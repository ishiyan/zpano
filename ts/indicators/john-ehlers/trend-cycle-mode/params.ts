import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';
import { HilbertTransformerCycleEstimatorType } from '../hilbert-transformer/cycle-estimator-type';
import { HilbertTransformerCycleEstimatorParams } from '../hilbert-transformer/cycle-estimator-params';

/** Describes parameters to create an instance of the TrendCycleMode indicator. */
export interface TrendCycleModeParams {
  /** The type of cycle estimator to use.
   *
   * The default value is HilbertTransformerCycleEstimatorType.HomodyneDiscriminator.
   */
  estimatorType?: HilbertTransformerCycleEstimatorType;

  /** Parameters to create an instance of the Hilbert transformer cycle estimator. */
  estimatorParams?: HilbertTransformerCycleEstimatorParams;

  /** The value of α (0 < α ≤ 1) used in EMA for additional smoothing of the instantaneous period.
   *
   * The default value is 0.33.
   */
  alphaEmaPeriodAdditional: number;

  /** The additional WMA smoothing length used to smooth the trend line.
   *
   * The valid values are 2, 3, 4. The default value is 4.
   */
  trendLineSmoothingLength?: number;

  /** The multiplier to the dominant cycle period used to determine the window length to
   * calculate the trend line. Typical values are in [0.5, 1.5].
   *
   * The default value is 1.0. Valid range is (0, 10].
   */
  cyclePartMultiplier?: number;

  /** The threshold (in percent) above which a wide separation between the WMA-smoothed
   * price and the instantaneous trend line forces the trend mode.
   *
   * The default value is 1.5. Valid range is (0, 100].
   */
  separationPercentage?: number;

  /** A component of a bar to use when updating the indicator with a bar sample.
   *
   * If _undefined_, the default (BarComponent.Median, hl/2) is used. Since the default
   * differs from the framework default bar component, it is always shown in the mnemonic.
   */
  barComponent?: BarComponent;

  /** A component of a quote to use when updating the indicator with a quote sample.
   *
   * If _undefined_, a default value is used and the component is not shown in the indicator mnemonic.
   */
  quoteComponent?: QuoteComponent;

  /** A component of a trade to use when updating the indicator with a trade sample.
   *
   * If _undefined_, a default value is used and the component is not shown in the indicator mnemonic.
   */
  tradeComponent?: TradeComponent;
}

export function defaultParams(): TrendCycleModeParams {
    return { alphaEmaPeriodAdditional: 0.33, trendLineSmoothingLength: 4, cyclePartMultiplier: 1.0, separationPercentage: 1.5 };
}
