import { BarComponent } from '../../../entities/bar-component.js';
import { QuoteComponent } from '../../../entities/quote-component.js';
import { TradeComponent } from '../../../entities/trade-component.js';
/** Prepared reference parameters; explicit zero is never replaced by a default. */
export interface CyclePolarityParams {
    /** The window; default 100. */
    window: number;
    /** The short period; default 20.0. */
    shortPeriod: number;
    /** The long period; default 40.0. */
    longPeriod: number;
    /** The smoothing; default 1600.0. */
    smoothing: number;
    /** The tolerance; default 0.0. */
    tolerance: number;
    /** Optional bar extraction component; omitted defaults are excluded from the mnemonic. */
    barComponent?: BarComponent;
    /** Optional quote extraction component; omitted defaults are excluded from the mnemonic. */
    quoteComponent?: QuoteComponent;
    /** Optional trade extraction component; omitted defaults are excluded from the mnemonic. */
    tradeComponent?: TradeComponent;
}
export function defaultParams(): CyclePolarityParams { return { "window": 100, "shortPeriod": 20.0, "longPeriod": 40.0, "smoothing": 1600.0, "tolerance": 0.0 }; }
