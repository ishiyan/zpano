import { Trade } from './trade';
import { TradeComponent } from './trade-component.enum';

/** Function for calculating a component of a _Trade_. */
export const tradeComponentValue = (component: TradeComponent): (trade: Trade) => number => {
    switch (component) {
        case TradeComponent.Price:
            return (t: Trade) => t.price;
        case TradeComponent.Volume:
            return (t: Trade) => t.volume;
        default:
            return (t: Trade) => t.price;
    }
};

/** The mnemonic of a component of a _Trade_. */
export const tradeComponentMnemonic = (component: TradeComponent): string => {
    switch (component) {
        case TradeComponent.Price:
            return 'p';
        case TradeComponent.Volume:
            return 'v';
        default:
            return '??';
    }
};
