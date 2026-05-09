// Re-exports shared test data. The phase accumulator uses the same
// TA-Lib reference data as the dual differentiator and homodyne discriminator.
export {
    input,
    expectedSmoothed,
    expectedDetrended,
    expectedQuadrature,
    expectedInPhase,
    expectedPeriod,
} from './testdata-dual-differentiator';
