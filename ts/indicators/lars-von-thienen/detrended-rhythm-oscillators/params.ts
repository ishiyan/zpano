// SPDX-License-Identifier: MPL-2.0
// Source: Detrended Rhythm Oscillator, © WhenToTrade / Lars von Thienen.
import { BarComponent } from '../../../entities/bar-component.js';
import { QuoteComponent } from '../../../entities/quote-component.js';
import { TradeComponent } from '../../../entities/trade-component.js';
/** Prepared reference parameters; explicit zero is never replaced by a default. */
export interface DetrendedRhythmOscillatorsParams {
    /** The fast length; default 3. */
    fastLength: number;
    /** The slow length; default 50. */
    slowLength: number;
    /** The zig period; default 30. */
    zigPeriod: number;
    /** The cycle memory; default 15. */
    cycleMemory: number;
    /** The detrend; default True. */
    detrend: boolean;
    /** Optional bar extraction component; omitted defaults are excluded from the mnemonic. */
    barComponent?: BarComponent;
    /** Optional quote extraction component; omitted defaults are excluded from the mnemonic. */
    quoteComponent?: QuoteComponent;
    /** Optional trade extraction component; omitted defaults are excluded from the mnemonic. */
    tradeComponent?: TradeComponent;
}
export function defaultParams(): DetrendedRhythmOscillatorsParams { return { "fastLength": 3, "slowLength": 50, "zigPeriod": 30, "cycleMemory": 15, "detrend": true }; }
