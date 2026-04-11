/** An [open, high, low, close, volume] bar. */
export class Bar {
  /** The date and time.
   *
   * For _ohlcv_ bar entities it corresponds to the closing time, so that an _ohlcv_ bar accumulates lower-level entities
   * up to the closing date and time.
   */
  time!: Date;

  /** The opening price. */
  open!: number;

  /** The highest price. */
  high!: number;

  /** The lowest price. */
  low!: number;

  /** The closing price. */
  close!: number;

  /** The volume. */
  volume!: number;

  constructor(data?: any) {
    if (data) {
      for (const property in data) {
        if (Object.prototype.hasOwnProperty.call(data, property)) {
          (this as any)[property] = data[property];
        }
      }
    }
  }

  /** Indicates whether this is a rising bar (open < close). */
  isRising(): boolean {
    return this.open < this.close;
  }

  /** Indicates whether this is a falling bar (close < open). */
  isFalling(): boolean {
    return this.close < this.open;
  }

  /** The median price, calculated as _(low + high) / 2_. */
  median(): number {
    return (this.low + this.high) / 2;
  }

  /** The typical price, calculated as _(low + high + close) / 3_. */
  typical(): number {
    return (this.low + this.high + this.close) / 3;
  }

  /** The weighted price, calculated as _(low + high + 2*close) / 4_. */
  weighted(): number {
    return (this.low + this.high + this.close + this.close) / 4;
  }

  /** The average price, calculated as _(low + high + open + close) / 4_. */
  average(): number {
    return (this.low + this.high + this.open + this.close) / 4;
  }
}
