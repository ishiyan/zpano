/**
 * Streaming time-series performance and risk measures, a port of the Python
 * `performance` package.
 *
 * - `Measures`: streaming calculation of ~140 performance and risk measures
 *   for a portfolio and its benchmark.
 * - `PERIODS_PER_ANNUM_*`: common annualization conventions.
 * - `core`: the streaming building blocks used by `Measures`, exported as a
 *   namespace (`import { core } from './performance'`) to keep the top-level
 *   API small, as in the Python package where they live in `performance.core`.
 *
 * @module
 */

export {
    Measures,
    PERIODS_PER_ANNUM_YEAR,
    PERIODS_PER_ANNUM_QUARTER,
    PERIODS_PER_ANNUM_MONTH,
    PERIODS_PER_ANNUM_WEEK,
    PERIODS_PER_ANNUM_DAY,
    PERIODS_PER_ANNUM_MINUTE_US_EQUITIES,
    PERIODS_PER_ANNUM_MINUTE_CRYPTO,
} from './measures';

export * as core from './core';
