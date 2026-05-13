/** Error metric used by inverse-variance and rank-based methods. */
export enum ErrorMetric {
    /** |signal_i - outcome| */
    ABSOLUTE = 0,
    /** (signal_i - outcome)^2 */
    SQUARED = 1,
}
