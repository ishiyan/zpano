import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Parameters for the Jurik Relative Trend Strength Index indicator. */
export interface JurikRelativeTrendStrengthIndexParams {
    /** Smoothing length (minimum 2). */
    length: number;
    /** Bar component to extract. */
    barComponent?: BarComponent;
    /** Quote component to extract. */
    quoteComponent?: QuoteComponent;
    /** Trade component to extract. */
    tradeComponent?: TradeComponent;
}

export function defaultParams(): JurikRelativeTrendStrengthIndexParams {
    return { length: 14 };
}
