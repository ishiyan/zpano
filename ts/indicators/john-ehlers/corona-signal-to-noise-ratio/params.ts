import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the CoronaSignalToNoiseRatio indicator. */
export interface CoronaSignalToNoiseRatioParams {
  /** Length of the heatmap raster (number of intensity bins). Default 50. A zero value is treated as "use default". */
  rasterLength?: number;

  /** Maximum raster intensity value. Default 20. A zero value is treated as "use default". */
  maxRasterValue?: number;

  /** Minimum ordinate (y) value of the heatmap — lower bound of the mapped SNR. Default 1. A zero value is treated as "use default". */
  minParameterValue?: number;

  /** Maximum ordinate (y) value of the heatmap — upper bound of the mapped SNR. Default 11. A zero value is treated as "use default". */
  maxParameterValue?: number;

  /** High-pass filter cutoff used by the inner Corona engine. Default 30. A zero value is treated as "use default". */
  highPassFilterCutoff?: number;

  /** Minimal cycle period covered by the filter bank. Default 6. A zero value is treated as "use default". */
  minimalPeriod?: number;

  /** Maximal cycle period covered by the filter bank. Default 30. A zero value is treated as "use default". */
  maximalPeriod?: number;

  /** A component of a bar to use when updating the indicator with a bar sample.
   *
   * If _undefined_, the default (BarComponent.Median, hl/2) is used, matching
   * Ehlers' reference which operates on (High+Low)/2.
   */
  barComponent?: BarComponent;

  /** A component of a quote to use when updating the indicator with a quote sample. */
  quoteComponent?: QuoteComponent;

  /** A component of a trade to use when updating the indicator with a trade sample. */
  tradeComponent?: TradeComponent;
}
