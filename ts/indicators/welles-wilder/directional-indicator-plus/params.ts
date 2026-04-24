/** Parameters of the Directional Indicator Plus indicator. */
export interface DirectionalIndicatorPlusParams {
  /** The smoothing length (the number of time periods). Must be >= 1. The default value is 14. */
  length: number;
}

export function defaultParams(): DirectionalIndicatorPlusParams {
    return {
        length: 14,
    };
}
