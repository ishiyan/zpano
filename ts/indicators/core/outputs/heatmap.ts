/** Holds a time stamp (x) and an array of values (z) corresponding to parameter (y) range to paint a heatmap column. */
export class Heatmap {
  /** The date and time (x). */
  time!: Date;

  /** The first parameter (y) value of the heatmap. This value is the same for all columns. */
  parameterFirst!: number;

  /** The last parameter (y) value of the heatmap. This value is the same for all columns. */
  parameterLast!: number;

  /** A parameter resolution (positive number). A value of 10 means that heatmap values are evaluated at every 0.1 of parameter range. */
  parameterResolution!: number;

  /** A minimal value (z) of the heatmap column. */
  valueMin!: number;

  /** A maximal value (z) of the heatmap column. */
  valueMax!: number;

  /** The values (z) of the heatmap column. */
  values!: number[];
}
