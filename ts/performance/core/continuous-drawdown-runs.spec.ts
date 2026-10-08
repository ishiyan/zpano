import { ContinuousDrawdownRuns, ddPercent } from './continuous-drawdown-runs';

describe('ContinuousDrawdownRuns', () => {

    /**
     * Port of Python's random.Random(seed) for integer seeds (MT19937 seeded
     * with init_by_array) with Python's random(), gauss() and choice()
     * algorithms, so the generated data match the Python tests exactly.
     */
    class SeededRandom {
        private readonly mt = new Uint32Array(624);
        private mti = 625;
        private gaussNext: number | null = null;

        constructor(seed: number) {
            const mt = this.mt;
            mt[0] = 19650218;
            for (let i = 1; i < 624; i++) {
                const p = mt[i - 1] ^ (mt[i - 1] >>> 30);
                mt[i] = (Math.imul(1812433253, p) + i) >>> 0;
            }
            // init_by_array with key = [seed] (seed < 2**32).
            const key = [seed >>> 0];
            let i = 1;
            let j = 0;
            for (let k = Math.max(624, key.length); k > 0; k--) {
                const p = mt[i - 1] ^ (mt[i - 1] >>> 30);
                mt[i] = ((mt[i] ^ Math.imul(p, 1664525)) + key[j] + j) >>> 0;
                i++;
                j++;
                if (i >= 624) {
                    mt[0] = mt[623];
                    i = 1;
                }
                if (j >= key.length) {
                    j = 0;
                }
            }
            for (let k = 623; k > 0; k--) {
                const p = mt[i - 1] ^ (mt[i - 1] >>> 30);
                mt[i] = ((mt[i] ^ Math.imul(p, 1566083941)) - i) >>> 0;
                i++;
                if (i >= 624) {
                    mt[0] = mt[623];
                    i = 1;
                }
            }
            mt[0] = 0x80000000;
            this.mti = 624;
        }

        private genrandUint32(): number {
            const mt = this.mt;
            if (this.mti >= 624) {
                let kk = 0;
                let y: number;
                for (; kk < 624 - 397; kk++) {
                    y = (mt[kk] & 0x80000000) | (mt[kk + 1] & 0x7fffffff);
                    mt[kk] = mt[kk + 397] ^ (y >>> 1) ^ ((y & 1) ? 0x9908b0df : 0);
                }
                for (; kk < 623; kk++) {
                    y = (mt[kk] & 0x80000000) | (mt[kk + 1] & 0x7fffffff);
                    mt[kk] = mt[kk + (397 - 624)] ^ (y >>> 1) ^ ((y & 1) ? 0x9908b0df : 0);
                }
                y = (mt[623] & 0x80000000) | (mt[0] & 0x7fffffff);
                mt[623] = mt[396] ^ (y >>> 1) ^ ((y & 1) ? 0x9908b0df : 0);
                this.mti = 0;
            }
            let y = mt[this.mti++];
            y ^= y >>> 11;
            y ^= (y << 7) & 0x9d2c5680;
            y ^= (y << 15) & 0xefc60000;
            y ^= y >>> 18;
            return y >>> 0;
        }

        random(): number {
            const a = this.genrandUint32() >>> 5;
            const b = this.genrandUint32() >>> 6;
            return (a * 67108864.0 + b) * (1.0 / 9007199254740992.0);
        }

        private randbelow(n: number): number {
            const k = 32 - Math.clz32(n); // n.bit_length()
            let r = this.genrandUint32() >>> (32 - k);
            while (r >= n) {
                r = this.genrandUint32() >>> (32 - k);
            }
            return r;
        }

        gauss(mu: number, sigma: number): number {
            let z = this.gaussNext;
            this.gaussNext = null;
            if (z === null) {
                const x2pi = this.random() * 2 * Math.PI;
                const g2rad = Math.sqrt(-2.0 * Math.log(1.0 - this.random()));
                z = Math.cos(x2pi) * g2rad;
                this.gaussNext = Math.sin(x2pi) * g2rad;
            }
            return mu + z * sigma;
        }

        choice<T>(seq: readonly T[]): T {
            return seq[this.randbelow(seq.length)];
        }
    }

    /** Compensated (Neumaier) sum, like Python's built-in sum of floats. */
    function pySum(xs: Iterable<number>): number {
        let s = 0.0;
        let c = 0.0;
        for (const x of xs) {
            const t = s + x;
            if (Math.abs(s) >= Math.abs(x)) {
                c += (s - t) + x;
            } else {
                c += (x - t) + s;
            }
            s = t;
        }
        return s + c;
    }

    /** Compounded drawdown of one continuous losing run. */
    function dd(...returns: number[]): number {
        let compounded = 1;
        for (const r of returns) {
            compounded *= 1.0 + r * 0.01;
        }
        return (compounded - 1.0) * 100.0;
    }

    function assertDrawdownsAlmostEqual(actual: readonly number[], expected: readonly number[],
        places = 12, prefix = ''): void {
        expect(actual.length).withContext(prefix).toBe(expected.length);
        const n = Math.min(actual.length, expected.length);
        for (let i = 0; i < n; i++) {
            expect(actual[i]).withContext(`${prefix} step ${i}`).toBeCloseTo(expected[i], places);
        }
    }

    function assertState(acc: ContinuousDrawdownRuns, expectedDrawdowns: readonly number[],
        expectedRunCount: number | null = null, places = 12, prefix = ''): void {
        assertDrawdownsAlmostEqual(acc.drawdowns, expectedDrawdowns, places, prefix);
        if (expectedRunCount !== null) {
            expect(acc.runCount).withContext(prefix).toBe(expectedRunCount);
        }
        const expectedSumSq = pySum(expectedDrawdowns.map(x => x * x));
        expect(acc.sumDrawdownsSquared).withContext(prefix).toBeCloseTo(expectedSumSq, places);
        expect(acc.sqrtSumDrawdownsSquared).withContext(prefix).toBeCloseTo(Math.sqrt(expectedSumSq), places);
    }

    // ------------------------------------------------------------------
    // Expanding-window tests
    // ------------------------------------------------------------------

    it('ddPercent converts a compounded log return into a percentage', () => {
        expect(ddPercent(0)).toBe(0);
        expect(ddPercent(Math.log1p(-0.05))).toBeCloseTo(-5.0, 13);
    });

    it('expanding single losing run', () => {
        // -1%, -2%, -3%: all three returns belong to one continuous losing run.
        const acc = new ContinuousDrawdownRuns();
        acc.update(-1.0);
        assertState(acc, [dd(-1.0)], 1);
        acc.update(-2.0);
        assertState(acc, [dd(-1.0, -2.0)], 1);
        acc.update(-3.0);
        assertState(acc, [dd(-1.0, -2.0, -3.0)], 1);
    });

    it('expanding multiple losing runs', () => {
        // -1%, -2%, +1%, -3%, -4%, +2%, -5% -> [-1%, -2%] [-3%, -4%] [-5%]
        const acc = new ContinuousDrawdownRuns();
        for (const ret of [-1.0, -2.0, 1.0, -3.0, -4.0, 2.0, -5.0]) {
            acc.update(ret);
        }
        assertState(acc, [dd(-1.0, -2.0), dd(-3.0, -4.0), dd(-5.0)], 3);
    });

    it('expanding non-negative returns are separators', () => {
        // Zero and positive returns must separate losing runs.
        const acc = new ContinuousDrawdownRuns();
        for (const ret of [-2.0, 0.0, -3.0, 1.0, -4.0]) {
            acc.update(ret);
        }
        assertState(acc, [dd(-2.0), dd(-3.0), dd(-4.0)], 3);
    });

    it('expanding positive return does not create drawdown', () => {
        const acc = new ContinuousDrawdownRuns();
        for (const ret of [2.0, 3.0, 0.0, 5.0]) {
            acc.update(ret);
        }
        assertState(acc, [], 0);
    });

    it('reset', () => {
        const acc = new ContinuousDrawdownRuns();
        for (const ret of [-1.0, -2.0, 1.0, -3.0]) {
            acc.update(ret);
        }
        expect(acc.runCount).toBeGreaterThan(0);

        acc.reset();
        assertState(acc, [], 0);

        // It must also be possible to use it again after reset.
        acc.update(-4.0);
        assertState(acc, [dd(-4.0)], 1);
    });

    // ------------------------------------------------------------------
    // Rolling-window tests
    // ------------------------------------------------------------------

    it('rolling window eviction from front of losing run', () => {
        // [-1%, -2%, -3%] -> remove -1%, add -4% -> [-2%, -3%, -4%]
        const acc = new ContinuousDrawdownRuns();
        for (const ret of [-1.0, -2.0, -3.0]) {
            acc.update(ret);
        }
        assertState(acc, [dd(-1.0, -2.0, -3.0)], 1);
        acc.revert(-1.0);
        acc.update(-4.0);
        assertState(acc, [dd(-2.0, -3.0, -4.0)], 1);
    });

    it('rolling window eviction of entire losing run', () => {
        // [-1%, -2%, +1%] -> [-2%, +1%, -3%]
        const acc = new ContinuousDrawdownRuns();
        for (const ret of [-1.0, -2.0, 1.0]) {
            acc.update(ret);
        }
        assertState(acc, [dd(-1.0, -2.0)], 1);
        acc.revert(-1.0);
        acc.update(-3.0);
        assertState(acc, [dd(-2.0), dd(-3.0)], 2);
    });

    it('rolling window eviction of separator', () => {
        // [-1%, +1%, -2%] -> [+1%, -2%, -3%]: -2%, -3% form one run.
        const acc = new ContinuousDrawdownRuns();
        for (const ret of [-1.0, 1.0, -2.0]) {
            acc.update(ret);
        }
        assertState(acc, [dd(-1.0), dd(-2.0)], 2);
        acc.revert(-1.0);
        acc.update(-3.0);
        assertState(acc, [dd(-2.0, -3.0)], 1);
    });

    it('rolling window multiple runs', () => {
        // Window size = 5.
        const acc = new ContinuousDrawdownRuns();
        for (const ret of [-1.0, -2.0, 1.0, -3.0, -4.0]) {
            acc.update(ret);
        }
        assertState(acc, [dd(-1.0, -2.0), dd(-3.0, -4.0)], 2);

        // Slide 1: remove -1%, add +2%
        acc.revert(-1.0);
        acc.update(2.0);
        assertState(acc, [dd(-2.0), dd(-3.0, -4.0)], 2);

        // Slide 2: remove -2%, add -5%
        acc.revert(-2.0);
        acc.update(-5.0);
        assertState(acc, [dd(-3.0, -4.0), dd(-5.0)], 2);
    });

    it('rolling window new return extends existing run', () => {
        // [+1%, -2%, -3%] -> [-2%, -3%, -4%]
        const acc = new ContinuousDrawdownRuns();
        for (const ret of [1.0, -2.0, -3.0]) {
            acc.update(ret);
        }
        assertState(acc, [dd(-2.0, -3.0)], 1);
        acc.revert(1.0);
        acc.update(-4.0);
        assertState(acc, [dd(-2.0, -3.0, -4.0)], 1);
    });

    it('rolling window new negative extends run after separator', () => {
        // [-2%, +1%, -3%] -> [+1%, -3%, -4%]
        const acc = new ContinuousDrawdownRuns();
        for (const ret of [-2.0, 1.0, -3.0]) {
            acc.update(ret);
        }
        assertState(acc, [dd(-2.0), dd(-3.0)], 2);
        acc.revert(-2.0);
        acc.update(-4.0);
        assertState(acc, [dd(-3.0, -4.0)], 1);
    });

    it('revert then update order is required', () => {
        const acc = new ContinuousDrawdownRuns();
        for (const ret of [-1.0, -2.0, 2.0]) {
            acc.update(ret);
        }
        assertState(acc, [dd(-1.0, -2.0)], 1);

        // New window: [-2%, +2%, -3%]
        acc.revert(-1.0);
        acc.update(-3.0);
        assertState(acc, [dd(-2.0), dd(-3.0)], 2);
    });

    // ------------------------------------------------------------------
    // Numerical consistency
    // ------------------------------------------------------------------

    it('sqrt sum drawdowns squared', () => {
        const acc = new ContinuousDrawdownRuns();
        for (const ret of [-1.0, -2.0, 1.0, -3.0]) {
            acc.update(ret);
        }

        const dd1 = dd(-1.0, -2.0);
        const dd2 = dd(-3.0);

        const expectedSumSq = dd1 ** 2 + dd2 ** 2;
        const expectedSqrt = Math.sqrt(expectedSumSq);

        expect(acc.sumDrawdownsSquared).toBeCloseTo(expectedSumSq, 12);
        expect(acc.sqrtSumDrawdownsSquared).toBeCloseTo(expectedSqrt, 12);
    });

    // ------------------------------------------------------------------
    // Brute-force rolling-window test
    // ------------------------------------------------------------------

    it('rolling window matches fresh calculation', () => {
        function referenceRuns(window: readonly number[]): number[] {
            const runs: number[] = [];
            let current: number[] = [];
            for (const r of window) {
                if (r < 0) {
                    current.push(r);
                } else if (current.length > 0) {
                    runs.push(dd(...current));
                    current = [];
                }
            }
            if (current.length > 0) {
                runs.push(dd(...current));
            }
            return runs;
        }

        const rng = new SeededRandom(42);
        for (const windowSize of [1, 2, 3, 5, 12]) {
            const returns: number[] = [];
            for (let i = 0; i < 150; i++) {
                returns.push(rng.choice([0.0, rng.gauss(0.0, 3.0)]));
            }
            const acc = new ContinuousDrawdownRuns();
            returns.forEach((ret, i) => {
                if (i >= windowSize) {
                    acc.revert(returns[i - windowSize]);
                }
                acc.update(ret);
                const window = returns.slice(Math.max(0, i - windowSize + 1), i + 1);
                const expected = referenceRuns(window);
                assertState(acc, expected, expected.length, 12, `window ${windowSize} step ${i}`);
            });
        }
    });
});
