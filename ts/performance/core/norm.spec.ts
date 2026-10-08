import { normCdf, normPdf, normPpf } from './norm';

describe('norm', () => {

    it('pdf scipy compliance', () => {
        // These values cover symmetry, peak, and tails
        const inputs = [
            -8.0, -5.0, -3.0,
            -2.0, -1.0, 0.0,
            1.0, 2.0, 3.0,
            5.0, 8.0];
        const outputs = [
            5.052271083536893e-15, 1.4867195147342979e-06, 0.0044318484119380075,
            0.053990966513188056, 0.24197072451914337, 0.3989422804014327,
            0.24197072451914337, 0.053990966513188056, 0.0044318484119380075,
            1.4867195147342979e-06, 5.052271083536893e-15];

        inputs.forEach((x, i) => {
            expect(normPdf(x)).withContext(`step ${i}`).toBeCloseTo(outputs[i], 15);
            // Test symmetry property: normPdf(x) == normPdf(-x)
            expect(normPdf(-x)).withContext(`step ${i} symmetry`).toBeCloseTo(normPdf(x), 15);
        });
    });

    it('cdf scipy compliance', () => {
        const inputs = [
            -8.0, -5.0, -3.0,
            -2.0, -1.0, 0.0,
            1.0, 2.0, 3.0,
            5.0, 8.0];
        const outputs = [
            6.22096057427174e-16, 2.866515718791933e-07, 0.0013498980316300933,
            0.022750131948179195, 0.15865525393145707, 0.5,
            0.8413447460685429, 0.9772498680518208, 0.9986501019683699,
            0.9999997133484281, 0.9999999999999993];

        inputs.forEach((x, i) => {
            expect(normCdf(x)).withContext(`step ${i}`).toBeCloseTo(outputs[i], 15);
            // Test symmetry property: normCdf(x) + normCdf(-x) = 1.0
            expect(normCdf(x) + normCdf(-x)).withContext(`step ${i} symmetry`).toBeCloseTo(1.0, 15);
        });
    });

    it('ppf scipy compliance', () => {
        const pValues = [
            // Center
            0.5, 0.75, 0.9,
            // Branch boundaries (values around the transition)
            0.02425, 0.025, 0.975, 0.97575,
            // Common quantiles
            0.001, 0.005, 0.01,
            0.025, 0.05, 0.10,
            0.90, 0.95, 0.975,
            0.99, 0.995, 0.999,
            // Extreme tails (exercise numerical stability)
            1e-12, 1e-10, 1e-8,
            0.99999999, 0.9999999999, 0.999999999999];
        // Hardcoded 1-pValues
        const pSymmetricValues = [
            // Center
            0.5, 0.25, 0.1,
            // Branch boundaries (values around the transition)
            0.97575, 0.975, 0.025, 0.02425,
            // Common quantiles
            0.999, 0.995, 0.99,
            0.975, 0.95, 0.90,
            0.10, 0.05, 0.025,
            0.01, 0.005, 0.001,
            // Extreme tails (exercise numerical stability)
            0.999999999999, 0.9999999999, 0.99999999,
            1e-8, 1e-10, 1e-12];
        const zValues = [
            // Center
            0.0, 0.6744897501960817, 1.2815515655446004,
            // Branch boundaries (values around the transition)
            -1.972961051311885, -1.9599639845400545, 1.959963984540054, 1.972961051311885,
            // Common quantiles
            -3.090232306167813, -2.575829303548901, -2.3263478740408408,
            -1.9599639845400545, -1.6448536269514729, -1.2815515655446004,
            1.2815515655446004, 1.6448536269514722, 1.959963984540054,
            2.3263478740408408, 2.5758293035489004, 3.090232306167,
            // Extreme tails (exercise numerical stability)
            -7.034483825301131, -6.361340902404056, -5.612001244174789,
            5.612001243305505, 6.361340889697422, 7.0344869100478356];

        // Peter Acklam states that relative error < 1.15 × 10^-9
        // over the entire domain.
        pValues.forEach((p, i) => {
            // Extreme tails have much lower accuracy.
            let delta = i < 19 ? 3e-9 : 8e-9;
            expect(Math.abs(normPpf(p) - zValues[i])).withContext(`step ${i}`)
                .toBeLessThanOrEqual(delta);

            // Verify that normPpf is the inverse of normCdf: p ≈ Φ(Φ⁻¹(p))
            expect(Math.abs(normCdf(normPpf(p)) - p)).withContext(`step ${i} roundtrip`)
                .toBeLessThanOrEqual(delta);

            // Test symmetry property: normPpf(p) == -normPpf(1.0 - p)
            delta = i < 19 ? 1e-13 : 4e-6;
            expect(Math.abs(-normPpf(pSymmetricValues[i]) - normPpf(p))).withContext(`step ${i} symmetry`)
                .toBeLessThanOrEqual(delta);
        });
    });

    it('ppf rejects p outside (0, 1)', () => {
        expect(() => normPpf(0)).toThrowError('p must be between 0 and 1 (exclusive)');
        expect(() => normPpf(1)).toThrowError('p must be between 0 and 1 (exclusive)');
    });
});
