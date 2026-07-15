import { RawMomentsKleinKbn } from './raw-moments-klein-kbn';

describe('RawMomentsKleinKbn', () => {

    function almostEqual(a: number, b: number, eps: number): boolean {
        return Math.abs(a - b) < eps;
    }

    const baconData = [
        0.003, 0.026, 0.011, -0.010, 0.015, 0.025, 0.016, 0.067,
        -0.014, 0.040, -0.005, 0.081, 0.040, -0.037, -0.061, 0.017,
        -0.049, -0.022, 0.070, 0.058, -0.065, 0.024, -0.005, -0.009,
    ];

    it('should compute moments for simple data', () => {
        const m = new RawMomentsKleinKbn(0, true, true);
        for (const x of [1.0, 2.0, 3.0, 4.0]) m.update(x);
        expect(m.mean).toBeCloseTo(2.5, 15);
        expect(m.variance).toBeCloseTo(1.25, 15);
        expect(m.skewness).toBeCloseTo(0.0, 14);
        expect(m.kurtosis).toBeCloseTo(-1.36, 13);
    });

    it('should match standard deviation from variance', () => {
        const m = new RawMomentsKleinKbn(0, true, true);
        for (const x of [1.0, 2.0, 3.0, 4.0]) m.update(x);
        expect(m.standardDeviation).toBeCloseTo(Math.sqrt(m.variance), 15);
    });

    it('should apply ddof', () => {
        const m = new RawMomentsKleinKbn(1, true, true);
        for (const x of [1.0, 2.0, 3.0]) m.update(x);
        expect(m.variance).toBeCloseTo(1.0, 15);
    });

    it('should match scipy for bacon data (bias=true)', () => {
        const m = new RawMomentsKleinKbn(0, true, true);
        for (const x of baconData) m.update(x);
        expect(m.mean).toBeCloseTo(0.009000000000000001, 15);
        expect(m.variance).toBeCloseTo(0.0014989166666666666, 14);
        expect(m.skewness).toBeCloseTo(-0.08256245520856798, 14);
        expect(m.kurtosis).toBeCloseTo(-0.5675462058921257, 13);
    });

    it('should match scipy for bacon data (bias=false)', () => {
        const m = new RawMomentsKleinKbn(0, false, true);
        for (const x of baconData) m.update(x);
        expect(m.skewness).toBeCloseTo(-0.08817174934967527, 14);
        expect(m.kurtosis).toBeCloseTo(-0.40766032118608714, 13);
    });

    it('should support revert round-trip', () => {
        const data = [1.0, 2.0, 3.0, 4.0, 5.0];
        const m = new RawMomentsKleinKbn(0, true, true);
        for (const x of data) m.update(x);
        for (let i = data.length - 2; i >= 0; i--) m.revert(data[i + 1]);
        expect(m.n).toBe(1);
        expect(m.mean).toBeCloseTo(1.0, 15);
    });

    it('should match partial revert', () => {
        const data = [10.0, 18.0, 5.0, 12.0, 7.0];
        const full = new RawMomentsKleinKbn(0, true, true);
        const part = new RawMomentsKleinKbn(0, true, true);
        for (const x of data) full.update(x);
        for (const x of data.slice(0, 4)) part.update(x);
        full.revert(data[4]);

        expect(full.mean).toBeCloseTo(part.mean, 15);
        expect(full.variance).toBeCloseTo(part.variance, 15);
        expect(full.skewness).toBeCloseTo(part.skewness, 14);
        expect(full.kurtosis).toBeCloseTo(part.kurtosis, 13);
    });

    it('should support reset', () => {
        const m = new RawMomentsKleinKbn(0, true, true);
        m.update(10.0);
        m.reset();
        expect(m.n).toBe(0);
        expect(isNaN(m.variance)).toBeTrue();
    });
});
