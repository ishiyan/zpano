/** Parameters of the Directional Movement Minus indicator. */
export interface DirectionalMovementMinusParams {
  /** The smoothing length (the number of time periods). Must be >= 1. The default value is 14. A length of 1 means no smoothing. */
  length: number;
}

export function defaultParams(): DirectionalMovementMinusParams {
    return {
        length: 14,
    };
}
