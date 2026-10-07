import { CentralMomentsKleinKbn } from './central-moments-klein-kbn';

describe('CentralMomentsKleinKbn', () => {

    // Bacon, Carl R., Practical Portfolio Performance Measurement and
    // Attribution, 2nd ed., Wiley, 2008, p. 65 (portfolio returns).
    const BACON = [
        0.003, 0.026, 0.011, -0.010,
        0.015, 0.025, 0.016, 0.067,
        -0.014, 0.040, -0.005, 0.081,
        0.040, -0.037, -0.061, 0.017,
        -0.049, -0.022, 0.070, 0.058,
        -0.065, 0.024, -0.005, -0.009];

    // Reference values for BACON, computed with exact rational arithmetic
    // (fractions.Fraction on the binary float inputs, square roots with
    // 50-digit decimal.Decimal), rounded to the nearest float.  The names
    // follow scipy.stats: skew(bias=...), kurtosis(bias=..., fisher=...).
    const MEAN = 0.009000000000000001;
    const VARIANCE_DDOF_0 = 0.0014989166666666668;
    const VARIANCE_DDOF_1 = 0.0015640869565217393;
    const STD_DDOF_0 = 0.03871584516275819;
    const STD_DDOF_1 = 0.039548539246370897;
    const SKEW_BIASED = -0.08256245520856804;            // skew(bias=True)
    const SKEW_UNBIASED = -0.08817174934967535;          // skew(bias=False)
    const KURT_BIASED_FISHER = -0.5675462058921257;      // kurtosis(bias=True, fisher=True)
    const KURT_BIASED_PEARSON = 2.4324537941078743;      // kurtosis(bias=True, fisher=False)
    const KURT_UNBIASED_FISHER = -0.40766032118608714;   // kurtosis(bias=False, fisher=True)
    const KURT_UNBIASED_PEARSON = 2.592339678813913;     // kurtosis(bias=False, fisher=False)

    // Same statistics for BACON.map(x => 1e4 + x), exact for the shifted floats.
    const OFFSET_SKEW_BIASED = -0.08256245521966786;
    const OFFSET_KURT_BIASED_FISHER = -0.5675462058934164;

    // ── Helpers ────────────────────────────────────────────────────────

    function feed(m: CentralMomentsKleinKbn, data: number[]): CentralMomentsKleinKbn {
        for (const x of data) {
            m.update(x);
        }
        return m;
    }

    /**
     * Correctly rounded sum of finite numbers, a port of CPython's
     * math.fsum (Shewchuk's partials with the final half-even correction).
     */
    function fsum(xs: number[]): number {
        const partials: number[] = [];
        for (let x of xs) {
            let i = 0;
            for (let j = 0; j < partials.length; j++) {
                let y = partials[j];
                if (Math.abs(x) < Math.abs(y)) {
                    const t = x;
                    x = y;
                    y = t;
                }
                const hi = x + y;
                const lo = y - (hi - x);
                if (lo !== 0.0) {
                    partials[i++] = lo;
                }
                x = hi;
            }
            partials.length = i;
            partials.push(x);
        }

        let n = partials.length;
        let hi = 0.0;
        if (n > 0) {
            let lo = 0.0;
            hi = partials[--n];
            while (n > 0) {
                const x = hi;
                const y = partials[--n];
                hi = x + y;
                const yr = hi - x;
                lo = y - yr;
                if (lo !== 0.0) {
                    break;
                }
            }
            if (n > 0 && ((lo < 0.0 && partials[n - 1] < 0.0) || (lo > 0.0 && partials[n - 1] > 0.0))) {
                const y = lo * 2.0;
                const x = hi + y;
                const yr = x - hi;
                if (y === yr) {
                    hi = x;
                }
            }
        }
        return hi;
    }

    /** The arithmetic mean, as Python statistics.fmean (fsum / n). */
    function fmean(xs: number[]): number {
        return fsum(xs) / xs.length;
    }

    /**
     * The population variance, approximating Python statistics.pvariance
     * (which uses exact rational arithmetic) with correctly rounded sums.
     */
    function pvariance(xs: number[]): number {
        const m = fmean(xs);
        return fsum(xs.map(x => (x - m) ** 2)) / xs.length;
    }

    // ── Tests ──────────────────────────────────────────────────────────

    it('simple update', () => {
        const m = feed(new CentralMomentsKleinKbn(0), [1.0, 2.0, 3.0, 4.0]);
        expect(m.n).toBe(4);
        expect(m.mean).toBeCloseTo(2.5, 15);
        expect(m.variance).toBeCloseTo(1.25, 15);
        expect(m.skewness).toBeCloseTo(0.0, 14);
        expect(m.kurtosis).toBeCloseTo(-1.36, 13);
    });

    it('bacon mean variance', () => {
        const m0 = feed(new CentralMomentsKleinKbn(0), BACON);
        const m1 = feed(new CentralMomentsKleinKbn(1), BACON);
        expect(m0.mean).toBeCloseTo(MEAN, 16);
        expect(m0.variance).toBeCloseTo(VARIANCE_DDOF_0, 16);
        expect(m1.variance).toBeCloseTo(VARIANCE_DDOF_1, 16);
        expect(m0.standardDeviation).toBeCloseTo(STD_DDOF_0, 15);
        expect(m1.standardDeviation).toBeCloseTo(STD_DDOF_1, 15);
    });

    it('bacon skewness kurtosis', () => {
        const cases: [boolean, boolean, number, number][] = [
            [true, true, SKEW_BIASED, KURT_BIASED_FISHER],
            [true, false, SKEW_BIASED, KURT_BIASED_PEARSON],
            [false, true, SKEW_UNBIASED, KURT_UNBIASED_FISHER],
            [false, false, SKEW_UNBIASED, KURT_UNBIASED_PEARSON],
        ];
        for (const [bias, fisher, skew, kurt] of cases) {
            const ctx = `bias=${bias}, fisher=${fisher}`;
            const m = feed(new CentralMomentsKleinKbn(0, bias, fisher), BACON);
            expect(m.skewness).withContext(ctx).toBeCloseTo(skew, 14);
            expect(m.kurtosis).withContext(ctx).toBeCloseTo(kurt, 13);
        }
    });

    it('large offset', () => {
        // Central moments don't suffer from the cancellation of raw power sums.
        const m = feed(new CentralMomentsKleinKbn(0), BACON.map(x => 1e4 + x));
        expect(m.mean).toBeCloseTo(1e4 + MEAN, 11);
        expect(m.variance).toBeCloseTo(VARIANCE_DDOF_0, 13);
        expect(m.skewness).toBeCloseTo(OFFSET_SKEW_BIASED, 10);
        expect(m.kurtosis).toBeCloseTo(OFFSET_KURT_BIASED_FISHER, 10);
    });

    it('scale invariance', () => {
        const m = feed(new CentralMomentsKleinKbn(0), BACON.map(x => x * 1e-6));
        expect(m.skewness).toBeCloseTo(SKEW_BIASED, 14);
        expect(m.kurtosis).toBeCloseTo(KURT_BIASED_FISHER, 13);
    });

    it('empty', () => {
        const m = new CentralMomentsKleinKbn();
        expect(m.n).toBe(0);
        expect(m.mean).toBe(0.0);
        expect(m.variance).toBeNaN();
        expect(m.standardDeviation).toBeNaN();
        expect(m.skewness).toBeNaN();
        expect(m.kurtosis).toBeNaN();
    });

    it('ddof', () => {
        let m = feed(new CentralMomentsKleinKbn(1), [1.0, 2.0, 3.0]);
        expect(m.variance).toBeCloseTo(1.0, 15);
        expect(m.standardDeviation).toBeCloseTo(1.0, 15);
        m = feed(new CentralMomentsKleinKbn(1), [1.0]);
        expect(m.variance).toBeNaN();
        expect(m.standardDeviation).toBeNaN();
    });

    it('invalid ddof', () => {
        for (const ddof of [-1, 0.5, true as unknown as number]) {
            expect(() => new CentralMomentsKleinKbn(ddof))
                .withContext(`ddof=${ddof}`)
                .toThrowError('ddof must be a nonnegative integer');
        }
    });

    it('minimum sample sizes', () => {
        const data = [1.0, 2.0, 4.0, 8.0];
        // [bias, fisher] -> minimum n for [skewness, kurtosis]
        const cases: [boolean, boolean, number, number][] = [
            [true, true, 2, 2],
            [true, false, 2, 2],
            [false, true, 3, 4],
            [false, false, 3, 4],
        ];
        for (const [bias, fisher, skewN, kurtN] of cases) {
            const m = new CentralMomentsKleinKbn(0, bias, fisher);
            data.forEach((x, i) => {
                m.update(x);
                const n = i + 1;
                const ctx = `bias=${bias}, fisher=${fisher}, n=${n}`;
                expect(isNaN(m.skewness)).withContext(ctx).toBe(n < skewN);
                expect(isNaN(m.kurtosis)).withContext(ctx).toBe(n < kurtN);
            });
        }
    });

    it('constant data', () => {
        const m = feed(new CentralMomentsKleinKbn(0), new Array<number>(5).fill(3.0));
        expect(m.mean).toBe(3.0);
        expect(m.variance).toBe(0.0);
        expect(m.standardDeviation).toBe(0.0);
        expect(m.skewness).toBeNaN();
        expect(m.kurtosis).toBeNaN();
    });

    it('revert lifo simple', () => {
        const data = [10.0, 18.0, 5.0];
        const mFull = feed(new CentralMomentsKleinKbn(0), data);
        const mPart = feed(new CentralMomentsKleinKbn(0), data.slice(0, 2));
        mFull.revert(data[2]);

        expect(mFull.n).toBe(2);
        expect(mFull.mean).toBeCloseTo(mPart.mean, 15);
        expect(mFull.variance).toBeCloseTo(mPart.variance, 15);
        expect(mFull.skewness).toBeCloseTo(mPart.skewness, 14);
        expect(mFull.kurtosis).toBeCloseTo(mPart.kurtosis, 13);
    });

    it('revert lifo bacon', () => {
        const cases: [boolean, boolean][] = [[true, true], [false, false]];
        for (const [bias, fisher] of cases) {
            const ctx = `bias=${bias}, fisher=${fisher}`;
            const mFull = feed(new CentralMomentsKleinKbn(0, bias, fisher), BACON);
            const mPart = feed(new CentralMomentsKleinKbn(0, bias, fisher), BACON.slice(0, -1));
            mFull.revert(BACON[BACON.length - 1]);

            expect(mFull.mean).withContext(ctx).toBeCloseTo(mPart.mean, 15);
            expect(mFull.variance).withContext(ctx).toBeCloseTo(mPart.variance, 15);
            expect(mFull.skewness).withContext(ctx).toBeCloseTo(mPart.skewness, 13);
            expect(mFull.kurtosis).withContext(ctx).toBeCloseTo(mPart.kurtosis, 12);
        }
    });

    it('revert then update', () => {
        const m = feed(new CentralMomentsKleinKbn(0), BACON);
        for (const x of BACON.slice(12).reverse()) {
            m.revert(x);
        }
        feed(m, BACON.slice(12));
        expect(m.mean).toBeCloseTo(MEAN, 15);
        expect(m.variance).toBeCloseTo(VARIANCE_DDOF_0, 15);
        expect(m.skewness).toBeCloseTo(SKEW_BIASED, 12);
        expect(m.kurtosis).toBeCloseTo(KURT_BIASED_FISHER, 12);
    });

    it('revert lifo roundtrip', () => {
        const m = feed(new CentralMomentsKleinKbn(0), BACON);
        for (const x of [...BACON].reverse()) {
            m.revert(x);
        }
        expect(m.n).toBe(0);
        expect(m.mean).toBe(0.0);
        expect(m.variance).toBeNaN();
    });

    it('revert oldest and middle', () => {
        const data = [0.0, 1.0, 2.0, 4.0, 8.0];
        const m = feed(new CentralMomentsKleinKbn(0), data);
        for (const removed of [0.0, 2.0]) {
            m.revert(removed);
            data.splice(data.indexOf(removed), 1);
            const mean = fmean(data);
            const mu2 = fsum(data.map(x => (x - mean) ** 2)) / data.length;
            const mu3 = fsum(data.map(x => (x - mean) ** 3)) / data.length;
            const mu4 = fsum(data.map(x => (x - mean) ** 4)) / data.length;
            expect(m.n).toBe(data.length);
            expect(m.mean).toBeCloseTo(mean, 14);
            expect(m.variance).toBeCloseTo(mu2, 14);
            expect(m.skewness).toBeCloseTo(mu3 / mu2 ** 1.5, 13);
            expect(m.kurtosis).toBeCloseTo(mu4 / mu2 ** 2 - 3, 13);
        }
    });

    it('fifo rolling window', () => {
        const m = new CentralMomentsKleinKbn(0);
        const width = 6;
        BACON.forEach((x, i) => {
            m.update(x);
            if (i >= width) {
                m.revert(BACON[i - width]);
            }
            const window = BACON.slice(Math.max(0, i - width + 1), i + 1);
            expect(m.n).toBe(window.length);
            expect(m.mean).toBeCloseTo(fmean(window), 14);
            expect(m.variance).toBeCloseTo(pvariance(window), 14);
        });
    });

    it('revert empty raises', () => {
        const m = new CentralMomentsKleinKbn();
        expect(() => m.revert(1.0)).toThrowError('Cannot revert from an empty accumulator');
    });

    it('standard deviation is real after revert', () => {
        // Reverting to two equal samples can leave a tiny negative M2.
        const m = new CentralMomentsKleinKbn(0);
        for (const x of [0.1, 0.1, 0.7]) {
            m.update(x);
        }
        m.revert(0.7);
        expect(typeof m.standardDeviation).toBe('number');
        expect(m.variance).toBeGreaterThanOrEqual(0.0);
        expect(m.standardDeviation).toBeCloseTo(0.0, 15);
    });

    it('reset', () => {
        const m = feed(new CentralMomentsKleinKbn(), BACON);
        m.reset();
        expect(m.n).toBe(0);
        expect(m.mean).toBe(0.0);
        expect(m.variance).toBeNaN();
        feed(m, [1.0, 2.0, 3.0]);
        expect(m.variance).toBeCloseTo(1.0, 15);
    });
});
