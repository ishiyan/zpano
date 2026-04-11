/** A _trade_ (price and volume) entity. */
export class Trade {
  /** The date and time. */
  time!: Date;

  /** The price. */
  price!: number;

  /** The volume (quantity). */
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

  /*toJSON(data?: any): any {
    data = typeof data === 'object' ? data : {};
    data.time = this.time ? this.time.toISOString() : undefined;
    data.price = this.price;
    data.volume = this.volume;
    return data;
  }*/
}
