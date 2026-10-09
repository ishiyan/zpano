import { Measures } from './measures';

describe('Performance review regressions', () => {
    it('gain-to-pain is independent of repetition and window length', () => {
        for (const window of [0, 2, 4]) {
            const m = new Measures(1, 0, 0, window);
            expect(Number.isNaN(m.gainToPainRatio)).toBeTrue();
            for (let i = 0; i < 4; i++) {
                m.addReturn(0.1, 0);
                m.addReturn(-0.05, 0);
                expect(m.gainToPainRatio).toBeCloseTo(1, 13);
            }
            m.reset();
            m.addReturn(0.1, 0);
            expect(Number.isNaN(m.gainToPainRatio)).toBeTrue();
        }
    });

    it('modified information ratio uses geometric active premium', () => {
        const m = new Measures(1);
        m.addReturn(0.5, 0.05);
        m.addReturn(-0.3, 0.05);
        expect(m.activePremium).toBeLessThan(0);
        expect(m.informationRatioModified).toBeCloseTo(0.04473320734100739, 13);
    });

    it('rejects invalid arguments before checking data availability', () => {
        const m = new Measures(1);
        for (const populated of [false, true]) {
            if (populated) {
                m.addReturn(-0.1, 0);
                m.addReturn(0.2, 0);
            }
            for (const confidence of [-1e20, -1, 0, 1, 2, NaN, Infinity, -Infinity]) {
                expect(() => m.isNormalDistribution(confidence)).toThrowError('confidence must be between 0 and 1');
                expect(() => m.rewardToConditionalDrawdown(confidence)).toThrowError('confidence must be between 0 and 1');
            }
            for (const multiplier of [0, -1, NaN, -Infinity]) {
                expect(() => m.biasRatio(multiplier)).toThrowError('std_dev_multiplier must be positive');
            }
        }
        expect(Number.isNaN(new Measures(1).rewardToConditionalDrawdown(0.95))).toBeTrue();
    });
});
