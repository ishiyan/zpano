/** Classifies how an indicator uses volume information. */
export enum VolumeUsage {
  /** Does not use volume. */
  NoVolume = 1,

  /** Consumes per-bar aggregated volume. */
  AggregateBarVolume,

  /** Consumes per-trade volume. */
  PerTradeVolume,

  /** Consumes quote-side liquidity (bid/ask sizes). */
  QuoteLiquidityVolume
}
