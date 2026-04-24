/** Parameters of the Normalized Average True Range indicator. */
export interface NormalizedAverageTrueRangeParams {
  /** The number of time periods. Must be >= 1. The default value is 14. */
  length: number;
}

export function defaultParams(): NormalizedAverageTrueRangeParams {
    return {
        length: 14,
    };
}
