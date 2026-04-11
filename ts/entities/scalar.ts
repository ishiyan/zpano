/** A _scalar_ (value and time) entity. */
export class Scalar {
  /** The date and time. */
  time!: Date;

  /** The value. */
  value!: number;

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
    data.value = this.value;
    return data;
  }*/
}
