/** Parameters of the Directional Movement Plus indicator. */
export interface DirectionalMovementPlusParams {
  /** The smoothing length (the number of time periods). Must be >= 1. The default value is 14. A length of 1 means no smoothing. */
  length: number;
}

export function defaultParams(): DirectionalMovementPlusParams {
    return {
        length: 14,
    };
}
