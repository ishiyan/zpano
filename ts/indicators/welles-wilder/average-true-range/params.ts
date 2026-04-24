/** Parameters of the Average True Range indicator. */
export interface AverageTrueRangeParams {
  /** The number of time periods. Must be >= 1. The default value is 14. */
  length: number;
}

export function defaultParams(): AverageTrueRangeParams {
    return {
        length: 14,
    };
}
