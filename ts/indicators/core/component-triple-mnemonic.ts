import { BarComponent, DefaultBarComponent, barComponentMnemonic } from '../../entities/bar-component';
import { QuoteComponent, DefaultQuoteComponent, quoteComponentMnemonic } from '../../entities/quote-component';
import { TradeComponent, DefaultTradeComponent, tradeComponentMnemonic } from '../../entities/trade-component';

/**
 * Function to calculate mnemonic for a component triple.
 *
 * Default components are omitted from the mnemonic: the same indicator always
 * produces the same mnemonic regardless of whether defaults were explicitly
 * specified or left undefined.
 */
export const componentTripleMnemonic = (barComponent?: BarComponent, quoteComponent?: QuoteComponent, tradeComponent?: TradeComponent): string => {
    let str = '';

    if (barComponent !== undefined && barComponent !== DefaultBarComponent) {
        str += ', ' + barComponentMnemonic(barComponent);
    }

    if (quoteComponent !== undefined && quoteComponent !== DefaultQuoteComponent) {
        str += ', ' + quoteComponentMnemonic(quoteComponent);
    }

    if (tradeComponent !== undefined && tradeComponent !== DefaultTradeComponent) {
        str += ', ' + tradeComponentMnemonic(tradeComponent);
    }

    return str;
};
