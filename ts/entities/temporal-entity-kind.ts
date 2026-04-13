/** Enumerates temporal entity kinds. */
export enum TemporalEntityKind {
  /** Opening price, highest price, lowest price, closing price, volume. */
  Bar = 'bar',

  /** Ask price, ask size, bid price, bid size. */
  Quote = 'quote',

  /** Time, price and value. */
  Trade = 'trade',

  /** Time and value. */
  Scalar = 'scalar',
}
