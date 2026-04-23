/** Enumerates indicator output data shapes. */
export enum Shape {

  /** Holds a single value. */
  Scalar,

  /** Holds two values representing lower and upper lines of a band. */
  Band,

  /** Holds an array of values representing a heat-map column. */
  Heatmap,

  /** Holds an ordered, variable-length sequence of (offset, value) points. */
  Polyline
}
