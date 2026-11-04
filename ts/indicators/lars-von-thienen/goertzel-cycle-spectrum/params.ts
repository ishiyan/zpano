import { BarComponent } from '../../../entities/bar-component.js';
import { QuoteComponent } from '../../../entities/quote-component.js';
import { TradeComponent } from '../../../entities/trade-component.js';
/** Prepared reference parameters; explicit zero is never replaced by a default. */
export interface GoertzelCycleSpectrumParams {
    /** The window; default 128. */
    window: number;
    /** The min period; default 5. */
    minPeriod: number;
    /** The max period; default 50. */
    maxPeriod: number;
    /** The step; default 1. */
    step: number;
    /** The smoothing; default 1600.0. */
    smoothing: number;
    /** Optional bar extraction component; omitted defaults are excluded from the mnemonic. */
    barComponent?: BarComponent;
    /** Optional quote extraction component; omitted defaults are excluded from the mnemonic. */
    quoteComponent?: QuoteComponent;
    /** Optional trade extraction component; omitted defaults are excluded from the mnemonic. */
    tradeComponent?: TradeComponent;
}
export function defaultParams(): GoertzelCycleSpectrumParams { return { "window": 128, "minPeriod": 5, "maxPeriod": 50, "step": 1, "smoothing": 1600.0 }; }
