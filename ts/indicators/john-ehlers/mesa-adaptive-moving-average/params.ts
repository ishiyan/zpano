import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';
import { HilbertTransformerCycleEstimatorType } from '../hilbert-transformer/cycle-estimator-type';
import { HilbertTransformerCycleEstimatorParams } from '../hilbert-transformer/cycle-estimator-params';

/** Describes parameters to create an instance of the indicator based on length. */
export interface MesaAdaptiveMovingAverageLengthParams {
  /** The type of cycle estimator to use.
   *
   * The default value is HilbertTransformerCycleEstimatorType.HomodyneDiscriminator.
   */
  estimatorType?: HilbertTransformerCycleEstimatorType;

  /** Parameters to create an instance of the Hilbert transformer cycle estimator. */
  estimatorParams?: HilbertTransformerCycleEstimatorParams;

  /** FastLimitLength is the fastest boundary length, ℓf.
   * The equivalent smoothing factor αf is
   *
   *   αf = 2/(ℓf + 1), 2 ≤ ℓ
   *
   * The value should be greater than 1.
   * The default value is 3 (αf=0.5).
   */
  fastLimitLength: number;

  /** SlowLimitLength is the slowest boundary length, ℓs.
   * The equivalent smoothing factor αs is
   *
   *   αs = 2/(ℓs + 1), 2 ≤ ℓ
   *
   * The value should be greater than 1.
   * The default value is 39 (αs=0.05).
   */
  slowLimitLength: number;

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

/** Describes parameters to create an instance of the indicator based on smoothing factor. */
export interface MesaAdaptiveMovingAverageSmoothingFactorParams {
  /** The type of cycle estimator to use.
   *
   * The default value is HilbertTransformerCycleEstimatorType.HomodyneDiscriminator.
   */
  estimatorType?: HilbertTransformerCycleEstimatorType;

  /** Parameters to create an instance of the Hilbert transformer cycle estimator. */
  estimatorParams?: HilbertTransformerCycleEstimatorParams;

  /** FastLimitSmoothingFactor is the fastest boundary smoothing factor, αf in (0, 1).
   * The equivalent length ℓf is
   *
   *   ℓf = 2/αf - 1, 0 < αf < 1, 1 < ℓf
   *
   * The default value is 0.5 (ℓf=3).
   */
  fastLimitSmoothingFactor: number;

  /** SlowLimitSmoothingFactor is the slowest boundary smoothing factor, αs in (0, 1).
   * The equivalent length ℓs is
   *
   *   ℓs = 2/αs - 1, 0 < αs < 1, 1 < ℓs
   *
   * The default value is 0.05 (ℓs=39).
   */
  slowLimitSmoothingFactor: number;

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
