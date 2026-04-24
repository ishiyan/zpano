/** Parameters of the Directional Movement Index indicator. */
export interface DirectionalMovementIndexParams {
  /** The smoothing length (the number of time periods). Must be >= 1. The default value is 14. */
  length: number;
}

export function defaultParams(): DirectionalMovementIndexParams {
    return {
        length: 14,
    };
}
