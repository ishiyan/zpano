/** Enumerates price components of a _Bar_. */
export enum BarComponent {
  /** The opening price. */
  Open,

  /** The highest price. */
  High,

  /** The lowest price. */
  Low,

  /** The closing price. */
  Close,

  /** The volume. */
  Volume,

  /** The median price, calculated as _(high + low) / 2_. */
  Median,

  /** The typical price, calculated as _(high + low + close) / 3_. */
  Typical,

  /** The weighted price, calculated as _(high + low + 2*close) / 4_. */
  Weighted,

  /** The average price, calculated as _(open + high + low + close) / 4_. */
  Average,
}
