import { Quote } from './quote';
import { QuoteComponent } from './quote-component.enum';

/** Function for calculating a price component of a _Quote_. */
export const quoteComponentValue = (component: QuoteComponent): (quote: Quote) => number => {
  switch (component) {
    case QuoteComponent.Bid:
      return (q: Quote) => q.bidPrice;
    case QuoteComponent.Ask:
      return (q: Quote) => q.askPrice;
    case QuoteComponent.BidSize:
      return (q: Quote) => q.bidSize;
    case QuoteComponent.AskSize:
      return (q: Quote) => q.askSize;
    case QuoteComponent.Mid:
      return (q: Quote) => q.mid();
    case QuoteComponent.Weighted:
      return (q: Quote) => q.weighted();
    case QuoteComponent.WeightedMid:
      return (q: Quote) => q.weightedMid();
    case QuoteComponent.SpreadBp:
      return (q: Quote) => q.spreadBp();
    default: // Default to mid-price.
      return (q: Quote) => q.mid();
  }
};

/** The mnemonic of a price component of a _Quote_. */
export const quoteComponentMnemonic = (component: QuoteComponent): string => {
  switch (component) {
    case QuoteComponent.Bid:
      return 'b';
    case QuoteComponent.Ask:
      return 'a';
    case QuoteComponent.BidSize:
      return 'bs';
    case QuoteComponent.AskSize:
      return 'as';
    case QuoteComponent.Mid:
      return 'ba/2';
    case QuoteComponent.Weighted:
      return '(bbs+aas)/(bs+as)';
    case QuoteComponent.WeightedMid:
      return '(bas+abs)/(bs+as)';
    case QuoteComponent.SpreadBp:
      return 'spread bp';
    default:
      return '??';
  }
};
