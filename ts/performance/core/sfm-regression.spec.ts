import { SFMRegression } from './sfm-regression';

describe('SFMRegression', () => {

    it('full, bull and bear fits use excess returns', () => {
        const regression = new SFMRegression(0.01);
        for (const excessBenchmark of [-0.04, -0.02, 0.0, 0.02, 0.05]) {
            const benchmark = 0.01 + excessBenchmark;
            const portfolio = 0.01 + 0.005 + 2 * excessBenchmark;
            regression.update(portfolio, benchmark);
        }

        expect(regression.alpha).toBeCloseTo(0.005, 14);
        expect(regression.beta).toBeCloseTo(2.0, 14);
        expect(regression.betaBull).toBeCloseTo(2.0, 14);
        expect(regression.betaBear).toBeCloseTo(2.0, 14);
        expect(regression.r2).toBeCloseTo(1.0, 14);

        // Removing an older bear observation leaves too few bear points
        // for a slope, while the full and bull fits remain defined.
        regression.revert(0.01 + 0.005 + 2 * -0.04, 0.01 - 0.04);
        expect(regression.beta).toBeCloseTo(2.0, 14);
        expect(regression.betaBull).toBeCloseTo(2.0, 14);
        expect(regression.betaBear).toBeNaN();
    });

    it('reset and zero excess benchmark', () => {
        const regression = new SFMRegression(0.01);
        regression.update(0.02, 0.01);
        expect(regression.betaBull).toBeNaN();
        expect(regression.betaBear).toBeNaN();
        regression.reset();
        expect(regression.alpha).withContext('alpha').toBeNaN();
        expect(regression.beta).withContext('beta').toBeNaN();
        expect(regression.betaBull).withContext('betaBull').toBeNaN();
        expect(regression.betaBear).withContext('betaBear').toBeNaN();
        expect(regression.r2).withContext('r2').toBeNaN();
    });
});
