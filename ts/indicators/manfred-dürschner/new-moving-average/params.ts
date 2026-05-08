import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Type of moving average used in NMA calculation. */
export enum MAType {
    SMA = 0,
    EMA = 1,
    SMMA = 2,
    LWMA = 3,
}

/** Parameters for the New Moving Average indicator. */
export interface NewMovingAverageParams {
    primary_period?: number;
    secondary_period?: number;
    ma_type?: MAType;
    barComponent?: BarComponent;
    quoteComponent?: QuoteComponent;
    tradeComponent?: TradeComponent;
}

/** Returns default parameters. */
export function defaultParams(): NewMovingAverageParams {
    return {
        primary_period: 0,
        secondary_period: 8,
        ma_type: MAType.LWMA,
    };
}
