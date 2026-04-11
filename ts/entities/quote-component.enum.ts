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
