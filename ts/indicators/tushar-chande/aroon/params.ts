/** Describes parameters to create an instance of the Aroon indicator. */
export interface AroonParams {
    /**
     * The lookback period for the Aroon calculation.
     *
     * The value should be greater than 1. The default value is 14.
     */
    length: number;
}

export function defaultParams(): AroonParams {
    return { length: 14 };
}
