/**
 * A single entry of a Levels output, expressed as a value with an optional bar
 * offset and an optional strength.
 */
export interface Level {
  /** The value (e.g. a price level or a multiplier) at this entry. */
  value: number;

  /**
   * The number of bars back from the Levels' time at which this level occurs
   * (0 = the current bar). For levels not anchored to a past bar, offset is 0.
   */
  offset: number;

  /**
   * An optional significance measure for this level (higher = more significant).
   * It is NaN when not applicable.
   */
  strength: number;
}

/** Creates a level with the given value, offset and strength. */
export function newLevel(value: number, offset: number, strength: number): Level {
  return { value, offset, strength };
}

/**
 * Creates a level with the given value, an offset of 0 and a NaN strength.
 * Convenience for levels that carry only a value (e.g. a theoretical price level).
 */
export function newValueLevel(value: number): Level {
  return { value, offset: 0, strength: NaN };
}

/**
 * Holds a time stamp (anchoring the current bar) and a variable-length set of
 * levels, typically a ranked set of support/resistance price levels.
 *
 * Each update emits a fresh, self-contained Levels; renderers should replace the
 * previous set of this indicator with the new one. This provides an immutable,
 * streaming-friendly model for indicators whose level set may change as new bars
 * arrive (e.g. support/resistance, pivots, Fibonacci grids, quantum price levels).
 */
export class Levels {
  /** The date and time (x) of the bar that anchors this set (offset 0). */
  time!: Date;

  /** The set of levels. May be empty. */
  levels!: Level[];

  /** Creates a populated Levels. Entries are stored as-is. */
  public static newLevels(time: Date, levels: Level[]): Levels {
    const l = new Levels();
    l.time = time;
    l.levels = levels;
    return l;
  }

  /** Creates an empty Levels with no entries. */
  public static newEmptyLevels(time: Date): Levels {
    const l = new Levels();
    l.time = time;
    l.levels = [];
    return l;
  }

  /** Indicates whether this Levels has no entries. */
  public isEmpty(): boolean {
    return !this.levels || this.levels.length === 0;
  }
}
