import { Capture } from './capture';

describe('Capture', () => {

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

    function prod(xs: readonly number[]): number {
        let p = 1;
        for (const x of xs) {
            p *= x;
        }
        return p;
    }

    function nanDiv(a: number, b: number): number {
        return b !== 0 ? a / b : NaN;
    }

    type Pair = [number, number];

    const NAMES = [
        'upsideCaptureRatioGeometric', 'upsideCaptureRatioArithmetic',
        'downsideCaptureRatioGeometric', 'downsideCaptureRatioArithmetic',
        'upNumberRatio', 'downNumberRatio', 'upPercentageRatio', 'downPercentageRatio',
    ] as const;

    /**
     * Naive capture ratios following the PerformanceAnalytics conventions:
     * upside periods have benchmark > 0; downside capture and down-number
     * use benchmark <= 0; down-percentage uses benchmark < 0.
     */
    function reference(pairs: readonly Pair[]): Record<string, number> {
        const up = pairs.filter(([, b]) => b > 0);
        const dn = pairs.filter(([, b]) => b <= 0);
        const dnStrict = pairs.filter(([, b]) => b < 0);
        return {
            upsideCaptureRatioGeometric: nanDiv(
                prod(up.map(([a]) => 1 + a)) - 1, prod(up.map(([, b]) => 1 + b)) - 1),
            upsideCaptureRatioArithmetic: nanDiv(
                pySum(up.map(([a]) => a)), pySum(up.map(([, b]) => b))),
            downsideCaptureRatioGeometric: nanDiv(
                prod(dn.map(([a]) => 1 + a)) - 1, prod(dn.map(([, b]) => 1 + b)) - 1),
            downsideCaptureRatioArithmetic: nanDiv(
                pySum(dn.map(([a]) => a)), pySum(dn.map(([, b]) => b))),
            upNumberRatio: nanDiv(up.filter(([a]) => a > 0).length, up.length),
            downNumberRatio: nanDiv(dn.filter(([a]) => a < 0).length, dn.length),
            upPercentageRatio: nanDiv(up.filter(([a, b]) => a > b).length, up.length),
            downPercentageRatio: nanDiv(dnStrict.filter(([a, b]) => a > b).length, dnStrict.length),
        };
    }

    function prop(c: Capture, name: string): number {
        return (c as unknown as Record<string, number>)[name];
    }

    function assertMatches(c: Capture, pairs: readonly Pair[], places = 13, msg = ''): void {
        const ref = reference(pairs);
        for (const name of NAMES) {
            const expected = ref[name];
            const actual = prop(c, name);
            if (Number.isNaN(expected)) {
                expect(actual).withContext(`${msg} ${name}`).toBeNaN();
            } else {
                expect(actual).withContext(`${msg} ${name}`).toBeCloseTo(expected, places);
            }
        }
    }

    function randomPairs(rng: SeededRandom, n: number): Pair[] {
        const pairs: Pair[] = [];
        for (let i = 0; i < n; i++) {
            pairs.push([rng.choice([0.0, rng.gauss(0.0, 0.03)]), rng.choice([0.0, rng.gauss(0.0, 0.03)])]);
        }
        return pairs;
    }

    it('empty', () => {
        const c = new Capture();
        for (const name of NAMES) {
            expect(prop(c, name)).withContext(name).toBeNaN();
        }
    });

    it('hand computed', () => {
        const pairs: Pair[] = [[0.02, 0.01], [-0.01, -0.02], [0.03, 0.04], [-0.03, 0.0], [0.01, -0.01]];
        const c = new Capture();
        for (const [a, b] of pairs) {
            c.update(a, b);
        }
        // Up periods (b > 0): (0.02, 0.01), (0.03, 0.04)
        expect(c.upsideCaptureRatioArithmetic).toBeCloseTo(0.05 / 0.05, 15);
        expect(c.upsideCaptureRatioGeometric).toBeCloseTo((1.02 * 1.03 - 1) / (1.01 * 1.04 - 1), 14);
        expect(c.upNumberRatio).toBe(1.0);
        expect(c.upPercentageRatio).toBe(0.5);
        // Down periods (b <= 0) include the zero-benchmark period.
        expect(c.downsideCaptureRatioArithmetic)
            .toBeCloseTo((-0.01 - 0.03 + 0.01) / (-0.02 + 0.0 - 0.01), 15);
        expect(c.downNumberRatio).toBeCloseTo(2 / 3, 15);
        // Down-percentage only counts strictly negative benchmark periods.
        expect(c.downPercentageRatio).toBe(1.0);
    });

    it('matches reference', () => {
        const rng = new SeededRandom(42);
        const pairs = randomPairs(rng, 200);
        const c = new Capture();
        for (const [a, b] of pairs) {
            c.update(a, b);
        }
        assertMatches(c, pairs, 12);
    });

    it('rolling window matches reference', () => {
        const rng = new SeededRandom(7);
        const pairs = randomPairs(rng, 120);
        const w = 9;
        const c = new Capture();
        pairs.forEach(([a, b], i) => {
            if (i >= w) {
                c.revert(...pairs[i - w]);
            }
            c.update(a, b);
            assertMatches(c, pairs.slice(Math.max(0, i - w + 1), i + 1), 12, `step ${i}`);
        });
    });

    it('reset', () => {
        const c = new Capture();
        c.update(0.01, 0.02);
        c.update(-0.01, -0.02);
        c.reset();
        for (const name of NAMES) {
            expect(prop(c, name)).withContext(name).toBeNaN();
        }
    });
});
