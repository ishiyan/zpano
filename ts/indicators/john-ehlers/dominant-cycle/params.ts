import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';
import { HilbertTransformerCycleEstimatorType } from '../hilbert-transformer/cycle-estimator-type';
import { HilbertTransformerCycleEstimatorParams } from '../hilbert-transformer/cycle-estimator-params';

/** Describes parameters to create an instance of the DominantCycle indicator. */
export interface DominantCycleParams {
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

  /** A component of a bar to use when updating the indicator with a bar sample.
   *
   * If _undefined_, a default value is used and the component is not shown in the indicator mnemonic.
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
