import { BarComponent } from '../../../entities/bar-component.js';
import { QuoteComponent } from '../../../entities/quote-component.js';
import { TradeComponent } from '../../../entities/trade-component.js';
/** Prepared reference parameters; explicit zero is never replaced by a default. */
export interface CyclicSmoothedRelativeStrengthIndexParams {
    /** The dominant cycle; default 20. */
    dominantCycle: number;
    /** The vibration; default 10. */
    vibration: number;
    /** The leveling; default 10.0. */
    leveling: number;
    /** Optional bar extraction component; omitted defaults are excluded from the mnemonic. */
    barComponent?: BarComponent;
    /** Optional quote extraction component; omitted defaults are excluded from the mnemonic. */
    quoteComponent?: QuoteComponent;
    /** Optional trade extraction component; omitted defaults are excluded from the mnemonic. */
    tradeComponent?: TradeComponent;
}
export function defaultParams(): CyclicSmoothedRelativeStrengthIndexParams { return { "dominantCycle": 20, "vibration": 10, "leveling": 10.0 }; }
