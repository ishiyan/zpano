export interface HurstDifferenceParams {
    /** Lookback period N for FGDI computation. Must be >= 2. */
    period: number;
    /** Bar component to extract. Defaults to Close. */
    barComponent?: number;
    /** Quote component to extract. Defaults to Mid. */
    quoteComponent?: number;
    /** Trade component to extract. Defaults to Price. */
    tradeComponent?: number;
}

export function defaultParams(): HurstDifferenceParams {
    return { period: 30 };
}
