import { BarComponent } from '../../../entities/bar-component.js';
import { QuoteComponent } from '../../../entities/quote-component.js';
import { TradeComponent } from '../../../entities/trade-component.js';
/** Prepared reference parameters; explicit zero is never replaced by a default. */
export interface CycleSwingMomentumParams {
    /** The window; default 50. */
    window: number;
    /** The fast smoothing; default 160.0. */
    fastSmoothing: number;
    /** The slow smoothing; default 1600.0. */
    slowSmoothing: number;
    /** The band memory; default 34. */
    bandMemory: number;
    /** The leveling; default 10.0. */
    leveling: number;
    /** Optional bar extraction component; omitted defaults are excluded from the mnemonic. */
    barComponent?: BarComponent;
    /** Optional quote extraction component; omitted defaults are excluded from the mnemonic. */
    quoteComponent?: QuoteComponent;
    /** Optional trade extraction component; omitted defaults are excluded from the mnemonic. */
    tradeComponent?: TradeComponent;
}
export function defaultParams(): CycleSwingMomentumParams { return { "window": 50, "fastSmoothing": 160.0, "slowSmoothing": 1600.0, "bandMemory": 34, "leveling": 10.0 }; }
