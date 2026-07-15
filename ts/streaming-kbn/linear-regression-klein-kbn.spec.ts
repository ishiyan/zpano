import { LinearRegressionKleinKbn } from './linear-regression-klein-kbn';

describe('LinearRegressionKleinKbn', () => {

    function almostEqual(a: number, b: number, eps: number): boolean {
        return Math.abs(a - b) < eps;
    }

    it('should compute perfect fit (y = 2x + 1)', () => {
        const r = new LinearRegressionKleinKbn();
        for (let i = 0; i < 5; i++) r.update(i, 2 * i + 1);
        expect(r.slope).toBeCloseTo(2.0, 13);
        expect(r.intercept).toBeCloseTo(1.0, 13);
        expect(r.correlation).toBeCloseTo(1.0, 13);
    });

    it('should return zero slope and NaN correlation for constant y', () => {
        const r = new LinearRegressionKleinKbn();
        for (let i = 0; i < 5; i++) r.update(i, 0.0);
        expect(r.slope).toBeCloseTo(0.0, 13);
        expect(isNaN(r.correlation)).toBeTrue();
    });

    it('should return NaN for single point', () => {
        const r = new LinearRegressionKleinKbn();
        r.update(1.0, 2.0);
        expect(isNaN(r.slope)).toBeTrue();
        expect(isNaN(r.intercept)).toBeTrue();
        expect(isNaN(r.correlation)).toBeTrue();
    });

    it('should compute exact fit for two points', () => {
        const r = new LinearRegressionKleinKbn();
        r.update(0.0, 1.0);
        r.update(2.0, 5.0);
        expect(r.slope).toBeCloseTo(2.0, 13);
        expect(r.intercept).toBeCloseTo(1.0, 13);
        expect(r.correlation).toBeCloseTo(1.0, 13);
    });

    it('should match single update after revert', () => {
        const r = new LinearRegressionKleinKbn();
        r.update(1.0, 2.0);
        r.update(3.0, 4.0);
        r.revert(3.0, 4.0);

        const ref = new LinearRegressionKleinKbn();
        ref.update(1.0, 2.0);

        expect(r.n).toBe(ref.n);
        expect(isNaN(r.slope)).toBeTrue();
        expect(isNaN(ref.slope)).toBeTrue();
    });

    it('should revert to empty', () => {
        const r = new LinearRegressionKleinKbn();
        r.update(1.0, 2.0);
        r.revert(1.0, 2.0);
        expect(r.n).toBe(0);
        expect(isNaN(r.slope)).toBeTrue();
        expect(isNaN(r.intercept)).toBeTrue();
        expect(isNaN(r.correlation)).toBeTrue();
    });

    it('should support rolling window', () => {
        const data: [number, number][] = [[0, 1], [1, 3], [2, 5], [3, 7], [4, 9]];
        const r = new LinearRegressionKleinKbn();
        for (const [x, y] of data) r.update(x, y);

        r.revert(data[0][0], data[0][1]);
        r.revert(data[1][0], data[1][1]);
        r.update(5.0, 11.0);
        r.update(6.0, 13.0);

        const ref = new LinearRegressionKleinKbn();
        for (const [x, y] of data.slice(2)) ref.update(x, y);
        ref.update(5.0, 11.0);
        ref.update(6.0, 13.0);

        expect(r.n).toBe(ref.n);
        expect(almostEqual(r.slope, ref.slope, 1e-12)).toBeTrue();
        expect(almostEqual(r.intercept, ref.intercept, 1e-12)).toBeTrue();
        expect(almostEqual(r.correlation, ref.correlation, 1e-12)).toBeTrue();
    });

    it('should compute negative correlation', () => {
        const r = new LinearRegressionKleinKbn();
        for (let i = 0; i < 5; i++) r.update(i, -2 * i + 1);
        expect(r.slope).toBeCloseTo(-2.0, 13);
        expect(r.intercept).toBeCloseTo(1.0, 13);
        expect(r.correlation).toBeCloseTo(-1.0, 13);
    });

    it('should support reset', () => {
        const r = new LinearRegressionKleinKbn();
        for (let i = 0; i < 5; i++) r.update(i, 2 * i + 1);
        r.reset();
        expect(r.n).toBe(0);
        expect(isNaN(r.slope)).toBeTrue();
        expect(isNaN(r.intercept)).toBeTrue();
        expect(isNaN(r.correlation)).toBeTrue();

        r.update(0.0, 1.0);
        r.update(1.0, 3.0);
        expect(r.slope).toBeCloseTo(2.0, 13);
    });
});
