import { BarComponent } from '../../../entities/bar-component.js';
import { QuoteComponent } from '../../../entities/quote-component.js';
import { TradeComponent } from '../../../entities/trade-component.js';
/** Prepared reference parameters; explicit zero is never replaced by a default. */
export interface MetonicWilliamsPercentRParams {
    /** The length; default 13. */
    length: number;
    /** The delay1; default 6939. */
    delay1: number;
    /** The delay2; default 4016. */
    delay2: number;
    /** The delay3; default 2922. */
    delay3: number;
}
export function defaultParams(): MetonicWilliamsPercentRParams { return { "length": 13, "delay1": 6939, "delay2": 4016, "delay3": 2922 }; }
