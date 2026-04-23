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

  /** Creates a populated heatmap column. */
  public static newHeatmap(
    time: Date,
    parameterFirst: number,
    parameterLast: number,
    parameterResolution: number,
    valueMin: number,
    valueMax: number,
    values: number[],
  ): Heatmap {
    const h = new Heatmap();
    h.time = time;
    h.parameterFirst = parameterFirst;
    h.parameterLast = parameterLast;
    h.parameterResolution = parameterResolution;
    h.valueMin = valueMin;
    h.valueMax = valueMax;
    h.values = values;
    return h;
  }

  /** Creates an empty heatmap column with valid axis metadata but no values. */
  public static newEmptyHeatmap(
    time: Date,
    parameterFirst: number,
    parameterLast: number,
    parameterResolution: number,
  ): Heatmap {
    const h = new Heatmap();
    h.time = time;
    h.parameterFirst = parameterFirst;
    h.parameterLast = parameterLast;
    h.parameterResolution = parameterResolution;
    h.valueMin = Number.NaN;
    h.valueMax = Number.NaN;
    h.values = [];
    return h;
  }

  /** Indicates whether this heatmap column has no values. */
  public isEmpty(): boolean {
    return !this.values || this.values.length < 1;
  }
}
