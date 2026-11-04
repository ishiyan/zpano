import { BarComponent } from '../../../entities/bar-component.js';
import { QuoteComponent } from '../../../entities/quote-component.js';
import { TradeComponent } from '../../../entities/trade-component.js';
/** Prepared reference parameters; explicit zero is never replaced by a default. */
export interface PhasingOscillatorParams {
    /** The window; default 200. */
    window: number;
    /** The short min; default 5. */
    shortMin: number;
    /** The short max; default 30. */
    shortMax: number;
    /** The long min; default 31. */
    longMin: number;
    /** The long max; default 80. */
    longMax: number;
    /** The smoothing; default 1600.0. */
    smoothing: number;
    /** The genuine threshold; default 49.0. */
    genuineThreshold: number;
    /** The phase cycles; default 2.0. */
    phaseCycles: number;
    /** Optional bar extraction component; omitted defaults are excluded from the mnemonic. */
    barComponent?: BarComponent;
    /** Optional quote extraction component; omitted defaults are excluded from the mnemonic. */
    quoteComponent?: QuoteComponent;
    /** Optional trade extraction component; omitted defaults are excluded from the mnemonic. */
    tradeComponent?: TradeComponent;
}
export function defaultParams(): PhasingOscillatorParams { return { "window": 200, "shortMin": 5, "shortMax": 30, "longMin": 31, "longMax": 80, "smoothing": 1600.0, "genuineThreshold": 49.0, "phaseCycles": 2.0 }; }
