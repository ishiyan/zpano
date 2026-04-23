/**
 * A single vertex of a Polyline, expressed as (offset, value)
 * where offset is the number of bars back from the Polyline's time
 * (0 = the current bar, 1 = the previous bar, and so on).
 */
export interface Point {
  /** The number of bars back from the Polyline's time. */
  offset: number;

  /** The value (y) at this vertex. */
  value: number;
}

/**
 * Holds a time stamp (anchoring the current bar) and an ordered, variable-length
 * sequence of points describing a polyline over recent history.
 *
 * Points are ordered from oldest (largest offset) to newest (offset === 0).
 *
 * Each update emits a fresh, self-contained Polyline; renderers should replace
 * the previous polyline of this indicator with the new one. This provides an
 * immutable, streaming-friendly model for indicators whose historical overlay
 * may change as new bars arrive (e.g. ZigZag, Fibonacci grids, pivot overlays).
 */
export class Polyline {
  /** The date and time (x) of the bar that anchors this polyline (offset 0). */
  time!: Date;

  /** The ordered sequence of polyline vertices, from oldest to newest. May be empty. */
  points!: Point[];

  /**
   * Creates a populated polyline. Points are stored as-is; callers are
   * responsible for supplying them in the documented old-to-new order.
   */
  public static newPolyline(time: Date, points: Point[]): Polyline {
    const p = new Polyline();
    p.time = time;
    p.points = points;
    return p;
  }

  /** Creates an empty polyline with no points. */
  public static newEmptyPolyline(time: Date): Polyline {
    const p = new Polyline();
    p.time = time;
    p.points = [];
    return p;
  }

  /** Indicates whether this polyline has no points. */
  public isEmpty(): boolean {
    return !this.points || this.points.length === 0;
  }
}
