import { BarComponent } from '../../../entities/bar-component.js';
import { QuoteComponent } from '../../../entities/quote-component.js';
import { TradeComponent } from '../../../entities/trade-component.js';
/** Prepared reference parameters; explicit zero is never replaced by a default. */
export interface CompositeCycleParams {
    /** The window; default 100. */
    window: number;
    /** The periods; default [20.0, 40.0]. */
    periods: number[];
    /** The smoothing; default 1600.0. */
    smoothing: number;
    /** Optional bar extraction component; omitted defaults are excluded from the mnemonic. */
    barComponent?: BarComponent;
    /** Optional quote extraction component; omitted defaults are excluded from the mnemonic. */
    quoteComponent?: QuoteComponent;
    /** Optional trade extraction component; omitted defaults are excluded from the mnemonic. */
    tradeComponent?: TradeComponent;
}
export function defaultParams(): CompositeCycleParams { return { "window": 100, "periods": [20.0, 40.0], "smoothing": 1600.0 }; }
