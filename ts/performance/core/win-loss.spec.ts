import { WinLoss } from './win-loss';

describe('WinLoss', () => {

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

    function reference(returns: readonly number[]): Record<string, number> {
        const mean = (values: number[]): number => values.length ? pySum(values) / values.length : NaN;
        const wins = returns.filter(r => r > 0);
        const losses = returns.filter(r => r < 0);
        const nonZero = returns.filter(r => r !== 0);
        return {
            nonZeroReturnsCount: nonZero.length,
            nonZeroReturnsMean: mean(nonZero),
            winningReturnsCount: wins.length,
            winningReturnsSum: pySum(wins),
            winningReturnsMean: mean(wins),
            losingReturnsCount: losses.length,
            losingReturnsSum: pySum(losses),
            losingReturnsMean: mean(losses),
        };
    }

    function assertMatches(wl: WinLoss, returns: readonly number[], places = 15, msg = ''): void {
        for (const [name, expected] of Object.entries(reference(returns))) {
            const actual = (wl as unknown as Record<string, number>)[name];
            if (Number.isNaN(expected)) {
                expect(actual).withContext(`${msg} ${name}`).toBeNaN();
            } else {
                expect(actual).withContext(`${msg} ${name}`).toBeCloseTo(expected, places);
            }
        }
    }

    it('empty', () => {
        assertMatches(new WinLoss(), []);
    });

    it('hand computed', () => {
        const wl = new WinLoss();
        for (const r of [0.02, 0.0, -0.01, 0.04, 0.0, -0.03]) {
            wl.update(r);
        }
        expect(wl.nonZeroReturnsCount).toBe(4);
        expect(wl.nonZeroReturnsMean).toBeCloseTo(0.02 / 4, 16);
        expect(wl.winningReturnsCount).toBe(2);
        expect(wl.winningReturnsMean).toBeCloseTo(0.03, 16);
        expect(wl.losingReturnsCount).toBe(2);
        expect(wl.losingReturnsMean).toBeCloseTo(-0.02, 16);
    });

    it('rolling window matches reference', () => {
        const rng = new SeededRandom(42);
        const returns: number[] = [];
        for (let i = 0; i < 120; i++) {
            returns.push(rng.choice([0.0, rng.gauss(0.0, 0.03)]));
        }
        const w = 6;
        const wl = new WinLoss();
        returns.forEach((r, i) => {
            if (i >= w) {
                wl.revert(returns[i - w]);
            }
            wl.update(r);
            assertMatches(wl, returns.slice(Math.max(0, i - w + 1), i + 1), 15, `step ${i}`);
        });
    });

    it('reset', () => {
        const wl = new WinLoss();
        for (const r of [0.01, -0.02]) {
            wl.update(r);
        }
        wl.reset();
        assertMatches(wl, []);
    });
});
