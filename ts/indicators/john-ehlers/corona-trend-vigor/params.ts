import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the CoronaTrendVigor indicator. */
export interface CoronaTrendVigorParams {
  /** Length of the heatmap raster. Default 50. A zero value is treated as "use default". */
  rasterLength?: number;

  /** Maximum raster intensity value. Default 20. A zero value is treated as "use default". */
  maxRasterValue?: number;

  /** Minimum ordinate (y) value. Default -10. Only substituted when both Min and Max are 0 (unconfigured). */
  minParameterValue?: number;

  /** Maximum ordinate (y) value. Default 10. Only substituted when both Min and Max are 0 (unconfigured). */
  maxParameterValue?: number;

  /** High-pass filter cutoff. Default 30. A zero value is treated as "use default". */
  highPassFilterCutoff?: number;

  /** Minimal cycle period. Default 6. A zero value is treated as "use default". */
  minimalPeriod?: number;

  /** Maximal cycle period. Default 30. A zero value is treated as "use default". */
  maximalPeriod?: number;

  /** A component of a bar to use (default BarComponent.Median). */
  barComponent?: BarComponent;

  /** A component of a quote to use. */
  quoteComponent?: QuoteComponent;

  /** A component of a trade to use. */
  tradeComponent?: TradeComponent;
}
