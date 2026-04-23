import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the CoronaSwingPosition indicator. */
export interface CoronaSwingPositionParams {
  /** Length of the heatmap raster. Default 50. A zero value is treated as "use default". */
  rasterLength?: number;

  /** Maximum raster intensity value. Default 20. A zero value is treated as "use default". */
  maxRasterValue?: number;

  /** Minimum ordinate (y) value of the heatmap — lower bound of the mapped swing position.
   *  Default -5. Only substituted when both Min and Max are 0 (unconfigured). */
  minParameterValue?: number;

  /** Maximum ordinate (y) value of the heatmap — upper bound of the mapped swing position.
   *  Default 5. Only substituted when both Min and Max are 0 (unconfigured). */
  maxParameterValue?: number;

  /** High-pass filter cutoff. Default 30. A zero value is treated as "use default". */
  highPassFilterCutoff?: number;

  /** Minimal cycle period. Default 6. A zero value is treated as "use default". */
  minimalPeriod?: number;

  /** Maximal cycle period. Default 30. A zero value is treated as "use default". */
  maximalPeriod?: number;

  /** A component of a bar to use when updating the indicator with a bar sample.
   *  Default BarComponent.Median (hl/2). */
  barComponent?: BarComponent;

  /** A component of a quote to use when updating the indicator with a quote sample. */
  quoteComponent?: QuoteComponent;

  /** A component of a trade to use when updating the indicator with a trade sample. */
  tradeComponent?: TradeComponent;
}
