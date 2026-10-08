import { CumulativeReturn } from './cumulative-return';

describe('CumulativeReturn', () => {

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

    function prod(xs: readonly number[]): number {
        let p = 1;
        for (const x of xs) {
            p *= x;
        }
        return p;
    }

    it('annualized return definition', () => {
        const returns = [0.10, -0.05, 0.03, 0.08];
        const periodsPerYear = 12;

        const cr = new CumulativeReturn();
        for (const r of returns) {
            cr.update(r);
        }

        const growth = prod(returns.map(r => 1 + r));
        const expected = growth ** (periodsPerYear / returns.length) - 1;
        expect(cr.annualizedGeometricMeanReturn(periodsPerYear)).toBeCloseTo(expected, 15);
    });

    it('one month return', () => {
        // One monthly return of 1%, expected (1.01)^12-1.
        const cr = new CumulativeReturn();
        cr.update(0.01);
        expect(cr.annualizedGeometricMeanReturn(12)).toBeCloseTo(1.01 ** 12 - 1, 15);
    });

    it('yearly returns', () => {
        // If the observation itself is yearly, then the annualized
        // geometric mean return equals the geometric mean return.
        const cr = new CumulativeReturn();
        for (const r of [0.12, -0.04, 0.08]) {
            cr.update(r);
        }
        expect(cr.annualizedGeometricMeanReturn(1)).toBeCloseTo(cr.geometricMeanReturn, 15);
    });

    it('constant monthly return', () => {
        // ((1+r)^n)^(12/n) = (1+r)^12: the number of observations cancels.
        const cr = new CumulativeReturn();
        for (let i = 0; i < 60; i++) {
            cr.update(0.01);
        }
        expect(cr.annualizedGeometricMeanReturn(12)).toBeCloseTo(1.01 ** 12 - 1, 15);
    });

    it('empty', () => {
        const cr = new CumulativeReturn();
        expect(cr.annualizedGeometricMeanReturn(12)).toBeNaN();
    });

    it('zero returns', () => {
        // log1p(0) == 0 and expm1(0) == 0, so the result is exactly zero.
        const cr = new CumulativeReturn();
        for (let i = 0; i < 100; i++) {
            cr.update(0.0);
        }
        expect(cr.annualizedGeometricMeanReturn(252)).toBeCloseTo(0, 15);
    });

    it('consistency with geometric mean return', () => {
        // 1 + annualized = (1 + geometric mean)^p
        const returns = [0.0010, -0.0005, 0.0003, 0.0008];
        const periodsPerYear = 252;

        const cr = new CumulativeReturn();
        for (const r of returns) {
            cr.update(r);
        }

        const expected = (1 + cr.geometricMeanReturn) ** periodsPerYear - 1;
        expect(cr.annualizedGeometricMeanReturn(periodsPerYear)).toBeCloseTo(expected, 13);
    });

    it('rolling window matches fresh calculation', () => {
        const rng = new SeededRandom(42);
        const returns: number[] = [];
        for (let i = 0; i < 100; i++) {
            returns.push(rng.choice([0.0, rng.gauss(0.0, 0.03)]));
        }
        const windowSize = 7;
        const cr = new CumulativeReturn();
        returns.forEach((r, i) => {
            if (i >= windowSize) {
                cr.revert(returns[i - windowSize]);
            }
            cr.update(r);
            const window = returns.slice(Math.max(0, i - windowSize + 1), i + 1);
            const growth = prod(window.map(x => 1 + x));
            expect(cr.count).withContext(`step ${i}`).toBe(window.length);
            expect(cr.cumulativeGeometricReturn).withContext(`step ${i}`).toBeCloseTo(growth - 1, 14);
            expect(cr.geometricMeanReturn).withContext(`step ${i}`)
                .toBeCloseTo(growth ** (1 / window.length) - 1, 14);
        });
    });

    it('revert empty throws', () => {
        const cr = new CumulativeReturn();
        expect(() => cr.revert(0.01)).toThrowError('Cannot revert from an empty accumulator');
    });

    it('reset', () => {
        const cr = new CumulativeReturn();
        for (const r of [0.1, -0.2]) {
            cr.update(r);
        }
        cr.reset();
        expect(cr.count).toBe(0);
        expect(cr.cumulativeGeometricReturn).toBe(0.0);
        expect(cr.geometricMeanReturn).toBeNaN();
    });
});
