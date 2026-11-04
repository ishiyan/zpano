import { BarComponent } from '../../../entities/bar-component.js';
import { QuoteComponent } from '../../../entities/quote-component.js';
import { TradeComponent } from '../../../entities/trade-component.js';
/** Prepared reference parameters; explicit zero is never replaced by a default. */
export interface BartelsPeriodicityScoreParams {
    /** The period; default 20. */
    period: number;
    /** The segments; default 5. */
    segments: number;
    /** The smoothing; default 0.0. */
    smoothing: number;
    /** Optional bar extraction component; omitted defaults are excluded from the mnemonic. */
    barComponent?: BarComponent;
    /** Optional quote extraction component; omitted defaults are excluded from the mnemonic. */
    quoteComponent?: QuoteComponent;
    /** Optional trade extraction component; omitted defaults are excluded from the mnemonic. */
    tradeComponent?: TradeComponent;
}
export function defaultParams(): BartelsPeriodicityScoreParams { return { "period": 20, "segments": 5, "smoothing": 0.0 }; }
