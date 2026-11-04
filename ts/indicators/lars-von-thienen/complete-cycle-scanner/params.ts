import { BarComponent } from '../../../entities/bar-component.js';
import { QuoteComponent } from '../../../entities/quote-component.js';
import { TradeComponent } from '../../../entities/trade-component.js';
/** Prepared reference parameters; explicit zero is never replaced by a default. */
export interface CompleteCycleScannerParams {
    /** The window; default 200. */
    window: number;
    /** The min period; default 5. */
    minPeriod: number;
    /** The max period; default 80. */
    maxPeriod: number;
    /** The smoothing; default 1600.0. */
    smoothing: number;
    /** The genuine threshold; default 49.0. */
    genuineThreshold: number;
    /** The phase cycles; default 2.0. */
    phaseCycles: number;
    /** The top k; default 3. */
    topK: number;
    /** The peak distance; default 2. */
    peakDistance: number;
    /** The max segments; default 5. */
    maxSegments: number;
    /** Optional bar extraction component; omitted defaults are excluded from the mnemonic. */
    barComponent?: BarComponent;
    /** Optional quote extraction component; omitted defaults are excluded from the mnemonic. */
    quoteComponent?: QuoteComponent;
    /** Optional trade extraction component; omitted defaults are excluded from the mnemonic. */
    tradeComponent?: TradeComponent;
}
export function defaultParams(): CompleteCycleScannerParams { return { "window": 200, "minPeriod": 5, "maxPeriod": 80, "smoothing": 1600.0, "genuineThreshold": 49.0, "phaseCycles": 2.0, "topK": 3, "peakDistance": 2, "maxSegments": 5 }; }
