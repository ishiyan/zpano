import { PartialMoments } from './partial-moments';
import { RawPartialMoments } from './partial-moments-raw';

describe('PartialMoments / RawPartialMoments', () => {

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

    function mean(values: readonly number[]): number {
        return values.length ? pySum(values) / values.length : NaN;
    }

    /** Python max(a, b): b only if b > a. */
    function pyMax(a: number, b: number): number {
        return b > a ? b : a;
    }

    /** Naive partial moments about the threshold. */
    function reference(returns: readonly number[], threshold: number): Record<string, number> {
        const lower = returns.filter(r => r < threshold).map(r => threshold - r);
        const upper = returns.filter(r => r > threshold).map(r => r - threshold);
        const n = returns.length;
        const expected: Record<string, number> = {
            totalCount: n,
            lowerExcessCount: lower.length,
            upperExcessCount: upper.length,
            downsideFrequency: n ? lower.length / n : NaN,
            upsideFrequency: n ? upper.length / n : NaN,
            downsidePotential: mean(returns.map(r => pyMax(threshold - r, 0.0))),
            upperExcessMoment1Sum: pySum(upper),
            upperExcessMoment2Sum: pySum(upper.map(x => x * x)),
            lowerExcessMoment2Sum: pySum(lower.map(x => x * x)),
        };
        for (const k of [1, 2, 3, 4]) {
            expected[`lowerPartialMoment${k}`] = mean(returns.map(r => pyMax(threshold - r, 0.0) ** k));
            expected[`higherPartialMoment${k}`] = mean(returns.map(r => pyMax(r - threshold, 0.0) ** k));
            expected[`upperExcessMoment${k}`] = mean(upper.map(x => x ** k));
            expected[`lowerExcessMoment${k}`] = mean(lower.map(x => x ** k));
        }
        return expected;
    }

    /** Naive raw partial moments (sums, threshold 0). */
    function rawReference(returns: readonly number[]): Record<string, number> {
        return {
            count: returns.length,
            lowerPartialMoment1: pySum(returns.map(r => pyMax(-r, 0.0))),
            higherPartialMoment1: pySum(returns.map(r => pyMax(r, 0.0))),
            countNegative: returns.filter(r => r < 0).length,
            sumNegative: pySum(returns.filter(r => r < 0)),
            countPositive: returns.filter(r => r > 0).length,
            sumPositive: pySum(returns.filter(r => r > 0)),
        };
    }

    function assertMatches(obj: object, expected: Record<string, number>, places = 14, msg = ''): void {
        for (const [name, e] of Object.entries(expected)) {
            const a = (obj as unknown as Record<string, number>)[name];
            expect(typeof a).withContext(`${msg} ${name} is defined`).toBe('number');
            if (Number.isNaN(e)) {
                expect(a).withContext(`${msg} ${name}`).toBeNaN();
            } else {
                expect(a).withContext(`${msg} ${name}`).toBeCloseTo(e, places);
            }
        }
    }

    /** Mix in values exactly equal to the threshold and to zero. */
    function randomReturns(seed: number, n: number, threshold: number): number[] {
        const rng = new SeededRandom(seed);
        const out: number[] = [];
        for (let i = 0; i < n; i++) {
            out.push(rng.choice([threshold, 0.0, rng.gauss(0.0, 0.03), rng.gauss(0.0, 0.03)]));
        }
        return out;
    }

    describe('PartialMoments', () => {

        it('empty', () => {
            const pm = new PartialMoments(0.01);
            expect(pm.totalCount).toBe(0);
            expect(pm.downsideFrequency).toBeNaN();
            expect(pm.upsideFrequency).toBeNaN();
            expect(pm.lowerPartialMoment2).toBeNaN();
        });

        it('hand computed', () => {
            const pm = new PartialMoments(0.01);
            for (const r of [0.03, -0.01, 0.01, 0.00]) {
                pm.update(r);
            }
            // Shortfalls below 1%: 0.02, 0.01; excesses above: 0.02.
            expect(pm.lowerPartialMoment1).toBeCloseTo((0.02 + 0.01) / 4, 16);
            expect(pm.lowerPartialMoment2).toBeCloseTo((0.0004 + 0.0001) / 4, 16);
            expect(pm.higherPartialMoment1).toBeCloseTo(0.02 / 4, 16);
            expect(pm.lowerExcessCount).toBe(2);
            expect(pm.upperExcessCount).toBe(1);
            expect(pm.downsideFrequency).toBe(0.5);
            expect(pm.upsideFrequency).toBe(0.25);
        });

        it('matches reference', () => {
            for (const threshold of [0.0, 0.005]) {
                const returns = randomReturns(42, 200, threshold);
                const pm = new PartialMoments(threshold);
                for (const r of returns) {
                    pm.update(r);
                }
                assertMatches(pm, reference(returns, threshold), 14, `threshold ${threshold}`);
            }
        });

        it('rolling window matches reference', () => {
            const threshold = 0.005;
            const returns = randomReturns(7, 120, threshold);
            const w = 8;
            const pm = new PartialMoments(threshold);
            returns.forEach((r, i) => {
                if (i >= w) {
                    pm.revert(returns[i - w]);
                }
                pm.update(r);
                const window = returns.slice(Math.max(0, i - w + 1), i + 1);
                assertMatches(pm, reference(window, threshold), 13, `step ${i}`);
            });
        });

        it('reset', () => {
            const pm = new PartialMoments(0.0);
            for (const r of [0.01, -0.02]) {
                pm.update(r);
            }
            pm.reset();
            expect(pm.totalCount).toBe(0);
            expect(pm.lowerExcessCount).toBe(0);
            expect(pm.upperExcessCount).toBe(0);
        });
    });

    describe('RawPartialMoments', () => {

        it('empty', () => {
            assertMatches(new RawPartialMoments(), rawReference([]));
        });

        it('matches reference', () => {
            const returns = randomReturns(42, 200, 0.0);
            const pm = new RawPartialMoments();
            for (const r of returns) {
                pm.update(r);
            }
            assertMatches(pm, rawReference(returns));
        });

        it('rolling window matches reference', () => {
            const returns = randomReturns(7, 120, 0.0);
            const w = 8;
            const pm = new RawPartialMoments();
            returns.forEach((r, i) => {
                if (i >= w) {
                    pm.revert(returns[i - w]);
                }
                pm.update(r);
                assertMatches(pm, rawReference(returns.slice(Math.max(0, i - w + 1), i + 1)), 14, `step ${i}`);
            });
        });

        it('reset', () => {
            const pm = new RawPartialMoments();
            for (const r of [0.01, -0.02]) {
                pm.update(r);
            }
            pm.reset();
            assertMatches(pm, rawReference([]));
        });
    });
});
