import { Quote } from './quote';

/** Enumerates price components of a _Quote_. */
export enum QuoteComponent {
  /** The bid price. */
  Bid,

  /** The ask price. */
  Ask,

  /** The bid size. */
  BidSize,

  /** The ask size. */
  AskSize,

  /** The mid-price, calculated as _(ask + bid) / 2_. */
  Mid,

  /** The weighted price, calculated as _(ask*askSize + bid*bidSize) / (askSize + bidSize)_. */
  Weighted,

  /** The weighted mid-price (sometimes called micro-price), calculated as _(ask*bidSize + bid*askSize) / (askSize + bidSize)_. */
  WeightedMid,

  /** The spread in basis points (100 basis points = 1%), calculated as _10000 * (ask - bid) / mid_. */
  SpreadBp,
}

/** The default quote component used when none is specified. */
export const DefaultQuoteComponent = QuoteComponent.Mid;

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
