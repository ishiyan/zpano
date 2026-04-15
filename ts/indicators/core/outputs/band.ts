/** Holds two band values and a time stamp. */
export class Band {
  /** The date and time. */
  time!: Date;

  /** A lower value of the band. */
  lower!: number;

  /** A higher value of the band. */
  upper!: number;
}
