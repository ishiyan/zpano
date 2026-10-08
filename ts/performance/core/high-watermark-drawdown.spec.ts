import { HighWaterMarkDrawdown } from './high-watermark-drawdown';

describe('HighWaterMarkDrawdown', () => {

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

    /**
     * Independent reference implementation of chronological
     * high-water-mark drawdowns, matching R PerformanceAnalytics
     * Drawdowns(): the high-water mark starts at the initial equity 1.
     */
    function expectedDrawdowns(returns: readonly number[]): number[] {
        let equity = 1.0;
        let peak = 1.0;
        const result: number[] = [];
        for (const ret of returns) {
            equity *= 1.0 + ret;
            let dd: number;
            if (equity >= peak) {
                peak = equity;
                dd = 0.0;
            } else {
                dd = equity / peak - 1.0;
            }
            result.push(dd);
        }
        return result;
    }

    /** Drawdowns of the rolling window ending at index i: a fresh calculation. */
    function expectedRollingDrawdowns(returns: readonly number[], windowSize: number, i: number): number[] {
        const lo = windowSize <= 0 ? 0 : Math.max(0, i - windowSize + 1);
        return expectedDrawdowns(returns.slice(lo, i + 1));
    }

    function assertState(acc: HighWaterMarkDrawdown, expected: readonly number[], places = 14, ctx = ''): void {
        const actual = acc.drawdowns;
        expect(actual.length).withContext(ctx).toBe(expected.length);
        for (let i = 0; i < Math.min(actual.length, expected.length); i++) {
            expect(actual[i]).withContext(`${ctx} [${i}]`).toBeCloseTo(expected[i], places);
        }
        expect(acc.drawdownsCount).withContext(ctx).toBe(expected.length);
        if (expected.length > 0) {
            expect(acc.drawdown).withContext(ctx).toBeCloseTo(expected[expected.length - 1], places);
            expect(acc.maximumDrawdown).withContext(ctx).toBeCloseTo(Math.min(...expected), places);
            const expectedMean = pySum(expected) / expected.length;
            const expectedSquaredMean = pySum(expected.map(x => x * x)) / expected.length;
            expect(acc.drawdownsMean).withContext(ctx).toBeCloseTo(expectedMean, places);
            expect(acc.drawdownsSquaredMean).withContext(ctx).toBeCloseTo(expectedSquaredMean, places);
        } else {
            expect(acc.drawdown).withContext(ctx).toBeNaN();
            expect(acc.maximumDrawdown).withContext(ctx).toBeNaN();
            expect(acc.drawdownsMean).withContext(ctx).toBeNaN();
            expect(acc.drawdownsSquaredMean).withContext(ctx).toBeNaN();
        }
    }

    function feed(acc: HighWaterMarkDrawdown, returns: readonly number[]): HighWaterMarkDrawdown {
        for (const ret of returns) {
            acc.update(ret);
        }
        return acc;
    }

    function gaussList(rng: SeededRandom, n: number): number[] {
        const out: number[] = [];
        for (let i = 0; i < n; i++) {
            out.push(rng.gauss(0.0, 0.03));
        }
        return out;
    }

    // ------------------------------------------------------------------
    // Expanding-window tests
    // ------------------------------------------------------------------

    it('expanding empty', () => {
        assertState(new HighWaterMarkDrawdown(0), []);
    });

    it('expanding all positive returns', () => {
        // Every observation creates a new high-water mark.
        const acc = feed(new HighWaterMarkDrawdown(0), [0.10, 0.05, 0.20]);
        assertState(acc, [0.0, 0.0, 0.0]);
    });

    it('expanding first negative return', () => {
        // The high-water mark starts at the initial equity 1.0 (as in R).
        const acc = feed(new HighWaterMarkDrawdown(0), [-0.05, -0.02, 0.10]);
        assertState(acc, [-0.05, -0.069, 0.0]);
    });

    it('expanding simple drawdown and recovery', () => {
        const acc = feed(new HighWaterMarkDrawdown(0), [0.10, -0.05, 0.10]);
        assertState(acc, [0.0, -0.05, 0.0]);
    });

    it('expanding compounded drawdown', () => {
        // 0.891 / 1.10 - 1 = -19%
        const acc = feed(new HighWaterMarkDrawdown(0), [0.10, -0.10, -0.10]);
        assertState(acc, [0.0, -0.10, -0.19]);
    });

    it('expanding new high-water mark resets drawdown', () => {
        const acc = feed(new HighWaterMarkDrawdown(0), [0.10, -0.05, 0.06, -0.02]);
        assertState(acc, [0.0, -0.05, 0.0, -0.02]);
    });

    it('expanding reset', () => {
        const acc = feed(new HighWaterMarkDrawdown(0), [0.10, -0.05, -0.02]);
        expect(acc.drawdownsCount).toBeGreaterThan(0);
        acc.reset();
        assertState(acc, []);

        // The accumulator can be reused, starting from equity 1.0 again.
        acc.update(-0.05);
        assertState(acc, [-0.05]);
    });

    it('expanding matches reference', () => {
        const rng = new SeededRandom(42);
        const returns = gaussList(rng, 200);
        const acc = feed(new HighWaterMarkDrawdown(0), returns);
        assertState(acc, expectedDrawdowns(returns), 13);
    });

    it('zero size means expanding', () => {
        const returns = [0.10, -0.05, -0.02, 0.05];
        const acc = feed(new HighWaterMarkDrawdown(0), returns);
        assertState(acc, expectedDrawdowns(returns));
    });

    it('negative size means expanding', () => {
        const returns = [0.10, -0.05, -0.02];
        const acc = feed(new HighWaterMarkDrawdown(-10), returns);
        assertState(acc, expectedDrawdowns(returns));
    });

    // ------------------------------------------------------------------
    // Rolling-window tests
    // ------------------------------------------------------------------

    it('rolling window peak eviction', () => {
        const acc = feed(new HighWaterMarkDrawdown(3), [0.10, -0.05, -0.02]);
        assertState(acc, [0.0, -0.05, -0.069]);
        acc.update(0.03);
        assertState(acc, [-0.05, -0.069, -0.04107]);
    });

    it('rolling window evicted peak followed by new peak', () => {
        const acc = feed(new HighWaterMarkDrawdown(3), [0.05, -0.02, 0.10]);
        assertState(acc, [0.0, -0.02, 0.0]);
        acc.update(-0.03);
        assertState(acc, [-0.02, 0.0, -0.03]);
    });

    it('rolling window peak eviction recomputes drawdowns', () => {
        const acc = feed(new HighWaterMarkDrawdown(3), [0.10, -0.05, -0.05]);
        assertState(acc, [0.0, -0.05, -0.0975]);
        acc.update(0.01);
        assertState(acc, [-0.05, -0.0975, -0.088475]);
    });

    it('rolling window all negative returns', () => {
        const returns = [-0.01, -0.02, -0.03, -0.04];
        const acc = new HighWaterMarkDrawdown(3);
        returns.forEach((ret, i) => {
            acc.update(ret);
            assertState(acc, expectedRollingDrawdowns(returns, 3, i), 14, `step ${i}`);
        });
    });

    it('rolling window size one', () => {
        const returns = [0.10, -0.05, -0.02, 0.03, -0.04];
        const acc = new HighWaterMarkDrawdown(1);
        returns.forEach((ret, i) => {
            acc.update(ret);
            assertState(acc, [Math.min(ret, 0.0)], 14, `step ${i}`);
        });
    });

    it('rolling window matches fresh calculation', () => {
        const rng = new SeededRandom(7);
        for (const windowSize of [2, 3, 4, 7, 20]) {
            for (let k = 0; k < 10; k++) {
                const returns = gaussList(rng, 80);
                const acc = new HighWaterMarkDrawdown(windowSize);
                returns.forEach((ret, i) => {
                    acc.update(ret);
                    assertState(acc, expectedRollingDrawdowns(returns, windowSize, i), 13,
                        `windowSize=${windowSize} step=${i}`);
                });
            }
        }
    });

    it('rolling window recompute flag', () => {
        // Evicting an observation can only change the remaining high-water
        // marks if its return was negative.
        const acc = new HighWaterMarkDrawdown(2);
        expect(acc.update(0.10)).toBe(false);
        expect(acc.update(-0.05)).toBe(false);
        // Evicts +10%: the remaining peaks don't change.
        expect(acc.update(-0.02)).toBe(false);
        // Evicts -5% and the new first observation (-2%) is also below
        // the old starting equity: recompute.
        expect(acc.update(0.01)).toBe(true);
        assertState(acc, [-0.02, -0.0102]);
        // Evicts -2%; the new first observation (equity 1.0343) is still
        // below the old starting equity (1.045): recompute.
        expect(acc.update(0.03)).toBe(true);
        assertState(acc, [0.0, 0.0]);
        // Evicts +1%: its equity is above the window's starting equity,
        // so the remaining peaks don't change.
        expect(acc.update(-0.01)).toBe(false);
        assertState(acc, [0.0, -0.01]);
    });
});
