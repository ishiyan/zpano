/** A price _quote_ (bid/ask price and size pair). */
export class Quote {
  /** The date and time. */
  time!: Date;

  /** The bid price. */
  bidPrice!: number;

  /** The bid size. */
  bidSize!: number;

  /** The ask price. */
  askPrice!: number;

  /** The ask size. */
  askSize!: number;

  constructor(data?: any) {
    if (data) {
      for (const property in data) {
        if (Object.prototype.hasOwnProperty.call(data, property)) {
          (this as any)[property] = data[property];
        }
      }
    }
  }

  /** The mid-price, calculated as _(askPrice + bidPrice) / 2_. */
  mid(): number {
    return (this.askPrice + this.bidPrice) / 2;
  }

  /** The weighted price, calculated as _(askPrice*askSize + bidPrice*bidSize) / (askSize + bidSize)_. Returns 0 if total size is 0. */
  weighted(): number {
    const size = this.askSize + this.bidSize;
    return size === 0 ? 0 : (this.askPrice * this.askSize + this.bidPrice * this.bidSize) / size;
  }

  /** The weighted mid-price (micro-price), calculated as _(askPrice*bidSize + bidPrice*askSize) / (askSize + bidSize)_. Returns 0 if total size is 0. */
  weightedMid(): number {
    const size = this.askSize + this.bidSize;
    return size === 0 ? 0 : (this.askPrice * this.bidSize + this.bidPrice * this.askSize) / size;
  }

  /** The spread in basis points, calculated as _20000 * (askPrice - bidPrice) / (askPrice + bidPrice)_. Returns 0 if mid is 0. */
  spreadBp(): number {
    const mid = this.askPrice + this.bidPrice;
    return mid === 0 ? 0 : 20000 * (this.askPrice - this.bidPrice) / mid;
  }
}
