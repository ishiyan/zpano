import { RawMomentsKleinKbn } from './raw-moments-klein-kbn';

// https://github.com/medo64/Medo/blob/main/tests/Tests.Medo/Math/WelfordVariance.cs
// https://github.com/andrewuhl/RollingWindow/blob/master/src/RollingWindow.cpp
// https://github.com/ajcr/rolling/blob/master/rolling/similarity.py

describe('RawMomentsKleinKbn', () => {

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
    // 50-digit decimal.Decimal), rounded to the nearest float.
    const EXPECTED: Record<string, number> = {
        mean: 0.009000000000000001,
        varianceDdof0: 0.0014989166666666668,
        varianceDdof1: 0.0015640869565217393,
        standardDeviationDdof0: 0.03871584516275819,
        standardDeviationDdof1: 0.039548539246370897,
        skewnessMoment: -0.08256245520856804,          // scipy skew(bias=True)
        skewnessFisher: -0.08817174934967535,          // scipy skew(bias=False)
        skewnessSample: -0.09398413873544505,          // R PerformanceAnalytics "sample"
        kurtosisMoment: 2.4324537941078743,            // scipy kurtosis(bias=True, fisher=False)
        kurtosisExcess: -0.5675462058921257,           // scipy kurtosis(bias=True, fisher=True)
        kurtosisSampleExcess: -0.40766032118608714,    // scipy kurtosis(bias=False, fisher=True)
        kurtosisSample: 2.592339678813913,             // scipy kurtosis(bias=False, fisher=False)
        kurtosisSampleCorrected: 3.027404613878848,    // R PerformanceAnalytics "sample"
        x1Sum: 0.21600000000000003,
        x2Sum: 0.037918,
        x3Sum: 0.0008738040000000003,
        x4Sum: 0.00014466403,
        x1: 0.009000000000000001,
        x2: 0.0015799166666666668,
        x3: 3.640850000000001e-05,
        x4: 6.0276679166666674e-06,
    };

    // Defaults are ddof=1, bias=true, fisher=true.
    const EXPECTED_BY_DISPATCH: Record<string, number> = {
        mean: EXPECTED['mean'],
        variance: EXPECTED['varianceDdof1'],
        skewness: EXPECTED['skewnessMoment'],
        kurtosis: EXPECTED['kurtosisExcess'],
    };

    // ── Helpers ────────────────────────────────────────────────────────

    function feed(m: RawMomentsKleinKbn, data: number[]): RawMomentsKleinKbn {
        for (const x of data) {
            m.update(x);
        }
        return m;
    }

    /** Reads a numeric property by name (the equivalent of Python getattr). */
    function prop(m: RawMomentsKleinKbn, name: string): number {
        return (m as unknown as Record<string, number>)[name];
    }

    // ── Tests ──────────────────────────────────────────────────────────

    it('simple update', () => {
        const m = feed(new RawMomentsKleinKbn(0), [1.0, 2.0, 3.0, 4.0]);
        expect(m.n).toBe(4);
        expect(m.mean).toBeCloseTo(2.5, 15);
        expect(m.variance).toBeCloseTo(1.25, 15);
        expect(m.skewness).toBeCloseTo(0.0, 14);
        expect(m.kurtosis).toBeCloseTo(-1.36, 13);
    });

    it('bacon all properties', () => {
        const m = feed(new RawMomentsKleinKbn(), BACON);
        for (const [name, expected] of Object.entries(EXPECTED)) {
            expect(prop(m, name)).withContext(name).toBeCloseTo(expected, 14);
        }
    });

    it('dispatch', () => {
        const cases: [boolean, boolean, string, string][] = [
            [true, true, 'skewnessMoment', 'kurtosisExcess'],
            [true, false, 'skewnessMoment', 'kurtosisMoment'],
            [false, true, 'skewnessFisher', 'kurtosisSampleExcess'],
            [false, false, 'skewnessFisher', 'kurtosisSample'],
        ];
        for (const [bias, fisher, skew, kurt] of cases) {
            const ctx = `bias=${bias}, fisher=${fisher}`;
            const m = feed(new RawMomentsKleinKbn(1, bias, fisher), BACON);
            expect(m.skewness).withContext(ctx).toBe(prop(m, skew));
            expect(m.kurtosis).withContext(ctx).toBe(prop(m, kurt));
            expect(m.skewness).withContext(ctx).toBeCloseTo(EXPECTED[skew], 14);
            expect(m.kurtosis).withContext(ctx).toBeCloseTo(EXPECTED[kurt], 13);
        }
    });

    it('ddof', () => {
        for (const ddof of [0, 1]) {
            const m = feed(new RawMomentsKleinKbn(ddof), BACON);
            expect(m.variance).toBe(prop(m, `varianceDdof${ddof}`));
            expect(m.standardDeviation).toBe(prop(m, `standardDeviationDdof${ddof}`));
        }
        const m = feed(new RawMomentsKleinKbn(1), [1.0, 2.0, 3.0]);
        expect(m.variance).toBeCloseTo(1.0, 15);
        expect(m.standardDeviation).toBeCloseTo(1.0, 15);
    });

    it('invalid ddof', () => {
        for (const ddof of [-1, 0.5, true as unknown as number]) {
            expect(() => new RawMomentsKleinKbn(ddof))
                .withContext(`ddof=${ddof}`)
                .toThrowError('ddof must be a nonnegative integer');
        }
    });

    it('kurtosis sample corrected difference', () => {
        // kurtosisSampleCorrected - kurtosisSample = (9n-15) / ((n-2)(n-3))
        const m = feed(new RawMomentsKleinKbn(), BACON);
        const n = BACON.length;
        expect(m.kurtosisSampleCorrected - m.kurtosisSample)
            .toBeCloseTo((9 * n - 15) / ((n - 2) * (n - 3)), 14);
    });

    it('empty', () => {
        const m = new RawMomentsKleinKbn();
        expect(m.n).toBe(0);
        expect(m.mean).toBe(0.0);
        for (const name of ['variance', 'standardDeviation', 'skewness', 'kurtosis',
            'x1', 'x2', 'x3', 'x4']) {
            expect(prop(m, name)).withContext(name).toBeNaN();
        }
        expect(m.x1Sum).toBe(0.0);
    });

    it('minimum sample sizes', () => {
        const data = [1.0, 2.0, 4.0, 8.0];
        const minimumN: Record<string, number> = {
            skewnessMoment: 2,
            skewnessFisher: 3,
            skewnessSample: 3,
            kurtosisMoment: 2,
            kurtosisExcess: 2,
            kurtosisSampleExcess: 4,
            kurtosisSample: 4,
            kurtosisSampleCorrected: 4,
        };
        const m = new RawMomentsKleinKbn();
        data.forEach((x, i) => {
            m.update(x);
            const n = i + 1;
            for (const [name, minN] of Object.entries(minimumN)) {
                expect(isNaN(prop(m, name))).withContext(`${name}, n=${n}`).toBe(n < minN);
            }
        });
    });

    it('constant data', () => {
        const m = feed(new RawMomentsKleinKbn(0), new Array<number>(5).fill(0.1));
        expect(m.mean).toBeCloseTo(0.1, 16);
        expect(m.variance).toBeCloseTo(0.0, 16);
        expect(m.skewness).toBeNaN();
        expect(m.kurtosis).toBeNaN();
    });

    it('scale invariance', () => {
        // The cancellation threshold is relative, so tiny values work.
        const m = feed(new RawMomentsKleinKbn(), BACON.map(x => x * 1e-6));
        expect(m.skewnessMoment).toBeCloseTo(EXPECTED['skewnessMoment'], 13);
        expect(m.kurtosisExcess).toBeCloseTo(EXPECTED['kurtosisExcess'], 13);
    });

    it('large offset preserves variance but not higher moments', () => {
        // Welford's variance remains usable when raw-power cancellation
        // makes skewness and kurtosis unreliable.
        const m = feed(new RawMomentsKleinKbn(0), [1e8, 1e8 + 1, 1e8 + 2]);
        expect(m.mean).toBeCloseTo(1e8 + 1, 10);
        expect(m.variance).toBeCloseTo(2 / 3, 14);
        expect(m.skewness).toBeNaN();
        expect(m.kurtosis).toBeNaN();
    });

    it('revert partial', () => {
        const data = [10.0, 18.0, 5.0, 12.0, 7.0];
        const mFull = feed(new RawMomentsKleinKbn(0), data);
        const mPart = feed(new RawMomentsKleinKbn(0), data.slice(0, 4));
        mFull.revert(data[4]);
        expect(mFull.n).toBe(4);
        expect(mFull.mean).toBeCloseTo(mPart.mean, 15);
        expect(mFull.variance).toBeCloseTo(mPart.variance, 15);
        expect(mFull.skewness).toBeCloseTo(mPart.skewness, 14);
        expect(mFull.kurtosis).toBeCloseTo(mPart.kurtosis, 13);
    });

    it('revert not most recent', () => {
        const m = feed(new RawMomentsKleinKbn(), [...BACON, 0.5]);
        m.revert(0.5);
        const m2 = feed(new RawMomentsKleinKbn(), [0.5, ...BACON]);
        m2.revert(0.5); // the oldest sample
        for (const name of ['mean', 'variance', 'skewness', 'kurtosis']) {
            expect(prop(m, name)).withContext(name).toBeCloseTo(EXPECTED_BY_DISPATCH[name], 13);
            expect(prop(m2, name)).withContext(name).toBeCloseTo(EXPECTED_BY_DISPATCH[name], 13);
        }
    });

    it('revert to empty', () => {
        const m = feed(new RawMomentsKleinKbn(0), BACON);
        for (const x of BACON) {
            m.revert(x);
        }
        expect(m.n).toBe(0);
        expect(m.mean).toBe(0.0);
        expect(m.x1Sum).toBe(0.0);
        expect(m.variance).toBeNaN();
        feed(m, [1.0, 2.0, 3.0, 4.0]);
        expect(m.variance).toBeCloseTo(1.25, 15);
    });

    it('revert empty raises', () => {
        const m = new RawMomentsKleinKbn();
        expect(() => m.revert(1.0)).toThrowError('Cannot revert from an empty accumulator');
    });

    it('rolling window', () => {
        const names = ['mean', 'variance', 'standardDeviation', 'skewness', 'kurtosis',
            'skewnessSample', 'kurtosisSampleCorrected', 'x1', 'x2', 'x3', 'x4'];
        const w = 5;
        const m = new RawMomentsKleinKbn(1, false, true);
        BACON.forEach((x, i) => {
            m.update(x);
            if (i >= w) {
                m.revert(BACON[i - w]);
            }
            const ref = feed(new RawMomentsKleinKbn(1, false, true),
                BACON.slice(Math.max(0, i - w + 1), i + 1));
            expect(m.n).toBe(ref.n);
            for (const name of names) {
                const actual = prop(m, name);
                const expected = prop(ref, name);
                const ctx = `step=${i}, name=${name}`;
                if (isNaN(expected)) {
                    expect(actual).withContext(ctx).toBeNaN();
                } else {
                    expect(actual).withContext(ctx).toBeCloseTo(expected, 13);
                }
            }
        });
    });

    it('standard deviation is real after revert', () => {
        const m = new RawMomentsKleinKbn(0);
        for (const x of [0.1, 0.1, 0.7]) {
            m.update(x);
        }
        m.revert(0.7);
        expect(typeof m.standardDeviation).toBe('number');
        expect(m.variance).toBeGreaterThanOrEqual(0.0);
        expect(m.standardDeviation).toBeCloseTo(0.0, 15);
    });

    it('variance getter has no side effects', () => {
        const m = feed(new RawMomentsKleinKbn(0), BACON);
        const v = m.variance;
        void m.standardDeviation;
        expect(m.variance).toBe(v);
        expect(m.n).toBe(BACON.length);
    });

    it('reset', () => {
        const m = feed(new RawMomentsKleinKbn(), BACON);
        m.reset();
        expect(m.n).toBe(0);
        expect(m.mean).toBe(0.0);
        expect(m.x4Sum).toBe(0.0);
        expect(m.variance).toBeNaN();
        feed(m, [1.0, 2.0, 3.0]);
        expect(m.variance).toBeCloseTo(1.0, 15);
    });
});
