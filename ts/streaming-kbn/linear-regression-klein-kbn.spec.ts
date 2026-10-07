import { LinearRegressionKleinKbn } from './linear-regression-klein-kbn';

describe('LinearRegressionKleinKbn', () => {

    // Bacon, Carl R., Practical Portfolio Performance Measurement and
    // Attribution, 2nd ed., Wiley, 2008, p. 65 (portfolio) and p. 66 (benchmark).
    const PORTFOLIO = [
        0.003, 0.026, 0.011, -0.010,
        0.015, 0.025, 0.016, 0.067,
        -0.014, 0.040, -0.005, 0.081,
        0.040, -0.037, -0.061, 0.017,
        -0.049, -0.022, 0.070, 0.058,
        -0.065, 0.024, -0.005, -0.009];
    const BENCHMARK = [
        0.002, 0.025, 0.018, -0.011,
        0.014, 0.018, 0.014, 0.065,
        -0.015, 0.042, -0.006, 0.083,
        0.039, -0.038, -0.062, 0.015,
        -0.048, 0.021, 0.060, 0.056,
        -0.067, 0.019, -0.003, 0.000];

    // Reference values for x = BENCHMARK, y = PORTFOLIO, computed with exact
    // rational arithmetic (fractions.Fraction on the binary float inputs,
    // square roots with 50-digit decimal.Decimal), rounded to the nearest float.
    const SLOPE = 0.9988502086225746;
    const INTERCEPT = -0.001030120844918352;
    const CORRELATION = 0.9693858148753051;
    const CO_MOMENT = 0.033844;
    const COVARIANCE = 0.0014101666666666668;
    const VARIANCE_X = 0.0014117899305555557;
    const VARIANCE_Y = 0.0014989166666666668;

    // ── Helpers ────────────────────────────────────────────────────────

    function feed(reg: LinearRegressionKleinKbn, xs: number[], ys: number[]): LinearRegressionKleinKbn {
        const len = Math.min(xs.length, ys.length);
        for (let i = 0; i < len; i++) {
            reg.update(xs[i], ys[i]);
        }
        return reg;
    }

    /** Reads a numeric property by name (the equivalent of Python getattr). */
    function prop(reg: LinearRegressionKleinKbn, name: string): number {
        return (reg as unknown as Record<string, number>)[name];
    }

    function expectAllNaN(reg: LinearRegressionKleinKbn): void {
        for (const name of ['slope', 'intercept', 'correlation']) {
            expect(prop(reg, name)).withContext(name).toBeNaN();
        }
    }

    // ── Tests ──────────────────────────────────────────────────────────

    it('bacon', () => {
        const reg = feed(new LinearRegressionKleinKbn(), BENCHMARK, PORTFOLIO);
        expect(reg.n).toBe(PORTFOLIO.length);
        expect(reg.slope).toBeCloseTo(SLOPE, 14);
        expect(reg.intercept).toBeCloseTo(INTERCEPT, 15);
        expect(reg.correlation).toBeCloseTo(CORRELATION, 14);
        expect(reg.coMoment).toBeCloseTo(CO_MOMENT, 16);
        expect(reg.covariance).toBeCloseTo(COVARIANCE, 16);
        expect(reg.varianceX).toBeCloseTo(VARIANCE_X, 16);
        expect(reg.varianceY).toBeCloseTo(VARIANCE_Y, 16);
    });

    it('perfect fit', () => {
        const reg = new LinearRegressionKleinKbn();
        for (let x = 0; x < 5; x++) {
            reg.update(x, 2 * x + 1);
        }
        expect(reg.slope).toBeCloseTo(2.0, 13);
        expect(reg.intercept).toBeCloseTo(1.0, 13);
        expect(reg.correlation).toBeCloseTo(1.0, 15);
        expect(reg.meanX).toBeCloseTo(2.0, 15);
        expect(reg.meanY).toBeCloseTo(5.0, 15);
        expect(reg.varianceX).toBeCloseTo(2.0, 15);
        expect(reg.coMoment).toBeCloseTo(20.0, 13);
        expect(reg.covariance).toBeCloseTo(4.0, 13);
    });

    it('negative correlation', () => {
        const reg = new LinearRegressionKleinKbn();
        for (let x = 0; x < 5; x++) {
            reg.update(x, -2.0 * x + 1.0);
        }
        expect(reg.slope).toBeCloseTo(-2.0, 13);
        expect(reg.intercept).toBeCloseTo(1.0, 13);
        expect(reg.correlation).toBeCloseTo(-1.0, 15);
        expect(reg.covariance).toBeCloseTo(-4.0, 13);
    });

    it('constant y', () => {
        const reg = new LinearRegressionKleinKbn();
        for (let x = 0; x < 5; x++) {
            reg.update(x, 3.0);
        }
        expect(reg.slope).toBeCloseTo(0.0, 13);
        expect(reg.intercept).toBeCloseTo(3.0, 13);
        expect(reg.covariance).toBe(0.0);
        expect(reg.correlation).toBeNaN();
    });

    it('constant x', () => {
        const reg = new LinearRegressionKleinKbn();
        for (let y = 0; y < 5; y++) {
            reg.update(3.0, y);
        }
        expect(reg.covariance).toBe(0.0);
        expectAllNaN(reg);
    });

    it('empty', () => {
        const reg = new LinearRegressionKleinKbn();
        expect(reg.n).toBe(0);
        expect(reg.coMoment).toBe(0.0);
        expect(reg.covariance).toBeNaN();
        expect(reg.varianceX).toBeNaN();
        expectAllNaN(reg);
    });

    it('single point', () => {
        const reg = new LinearRegressionKleinKbn();
        reg.update(1.0, 2.0);
        expect(reg.covariance).toBe(0.0);
        expectAllNaN(reg);
    });

    it('two points', () => {
        const reg = new LinearRegressionKleinKbn();
        reg.update(0.0, 1.0);
        reg.update(2.0, 5.0);
        expect(reg.slope).toBeCloseTo(2.0, 13);
        expect(reg.intercept).toBeCloseTo(1.0, 13);
        expect(reg.correlation).toBeCloseTo(1.0, 13);
    });

    it('revert most recent', () => {
        const reg = feed(new LinearRegressionKleinKbn(), [...BENCHMARK, 0.5], [...PORTFOLIO, -0.5]);
        reg.revert(0.5, -0.5);
        expect(reg.n).toBe(PORTFOLIO.length);
        expect(reg.slope).toBeCloseTo(SLOPE, 13);
        expect(reg.intercept).toBeCloseTo(INTERCEPT, 14);
        expect(reg.correlation).toBeCloseTo(CORRELATION, 13);
        expect(reg.covariance).toBeCloseTo(COVARIANCE, 15);
    });

    it('revert oldest', () => {
        const reg = feed(new LinearRegressionKleinKbn(), [0.5, ...BENCHMARK], [-0.5, ...PORTFOLIO]);
        reg.revert(0.5, -0.5);
        expect(reg.slope).toBeCloseTo(SLOPE, 13);
        expect(reg.intercept).toBeCloseTo(INTERCEPT, 14);
        expect(reg.correlation).toBeCloseTo(CORRELATION, 13);
        expect(reg.covariance).toBeCloseTo(COVARIANCE, 15);
    });

    it('revert to single', () => {
        const reg = new LinearRegressionKleinKbn();
        reg.update(1.0, 2.0);
        reg.update(3.0, 4.0);
        reg.revert(3.0, 4.0);
        expect(reg.n).toBe(1);
        expect(reg.meanX).toBeCloseTo(1.0, 15);
        expect(reg.meanY).toBeCloseTo(2.0, 15);
        expect(reg.coMoment).toBeCloseTo(0.0, 15);
        expectAllNaN(reg);
    });

    it('revert to empty', () => {
        const reg = new LinearRegressionKleinKbn();
        reg.update(1.0, 2.0);
        reg.revert(1.0, 2.0);
        expect(reg.n).toBe(0);
        expectAllNaN(reg);
    });

    it('revert empty raises', () => {
        const reg = new LinearRegressionKleinKbn();
        expect(() => reg.revert(1.0, 2.0)).toThrowError('Cannot revert from an empty regression');
    });

    it('rolling window', () => {
        const names = ['slope', 'intercept', 'correlation', 'covariance', 'coMoment',
            'varianceX', 'varianceY'];
        const w = 6;
        const reg = new LinearRegressionKleinKbn();
        for (let i = 0; i < BENCHMARK.length; i++) {
            reg.update(BENCHMARK[i], PORTFOLIO[i]);
            if (i >= w) {
                reg.revert(BENCHMARK[i - w], PORTFOLIO[i - w]);
            }
            const lo = Math.max(0, i - w + 1);
            const ref = feed(new LinearRegressionKleinKbn(), BENCHMARK.slice(lo, i + 1), PORTFOLIO.slice(lo, i + 1));
            expect(reg.n).toBe(ref.n);
            for (const name of names) {
                const actual = prop(reg, name);
                const expected = prop(ref, name);
                const ctx = `step=${i}, name=${name}`;
                if (isNaN(expected)) {
                    expect(actual).withContext(ctx).toBeNaN();
                } else {
                    expect(actual).withContext(ctx).toBeCloseTo(expected, 13);
                }
            }
        }
    });

    it('reset', () => {
        const reg = new LinearRegressionKleinKbn();
        for (let x = 0; x < 5; x++) {
            reg.update(x, 2 * x + 1);
        }
        reg.reset();
        expect(reg.n).toBe(0);
        expectAllNaN(reg);
        reg.update(0.0, 1.0);
        reg.update(1.0, 3.0);
        expect(reg.slope).toBeCloseTo(2.0, 13);
    });
});
