// Re-exports shared test data. The homodyne discriminator uses the same
// TA-Lib reference data as the dual differentiator and phase accumulator.
export {
    input,
    expectedSmoothed,
    expectedDetrended,
    expectedQuadrature,
    expectedInPhase,
    expectedPeriod,
} from './testdata-dual-differentiator';
