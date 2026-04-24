/** Parameters of the Average Directional Movement Index Rating indicator. */
export interface AverageDirectionalMovementIndexRatingParams {
  /** The smoothing length (the number of time periods). Must be >= 1. The default value is 14. */
  length: number;
}

export function defaultParams(): AverageDirectionalMovementIndexRatingParams {
    return {
        length: 14,
    };
}
