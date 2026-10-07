import { KleinKbnAccumulator } from './klein-kbn-accumulator';

describe('KleinKbnAccumulator', () => {

    // ── Helpers ────────────────────────────────────────────────────────

    function naiveSum(data: number[]): number {
        let s = 0.0;
        for (const x of data) {
            s += x;
        }
        return s;
    }

    function kbnSum(data: number[]): number {
        const kbn = new KleinKbnAccumulator();
        for (const x of data) {
            kbn.update(x);
        }
        return kbn.value;
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

    /** Seeded PRNG (Mulberry32) yielding 53-bit doubles in [0, 1). */
    class Rng {
        private state: number;

        constructor(seed: number) {
            this.state = seed >>> 0;
        }

        private nextUint32(): number {
            let z = (this.state = (this.state + 0x6d2b79f5) >>> 0);
            z = Math.imul(z ^ (z >>> 15), z | 1);
            z ^= z + Math.imul(z ^ (z >>> 7), z | 61);
            return (z ^ (z >>> 14)) >>> 0;
        }

        random(): number {
            const a = this.nextUint32() >>> 5;
            const b = this.nextUint32() >>> 6;
            return (a * 67108864 + b) / 9007199254740992;
        }

        uniform(a: number, b: number): number {
            return a + (b - a) * this.random();
        }

        choice<T>(xs: T[]): T {
            return xs[Math.floor(this.random() * xs.length)];
        }
    }

    // https://en.wikipedia.org/wiki/Kahan_summation_algorithm
    // A simple example due to Peters: summing [1.0, +1e100, 1.0, -1e100]
    // in double precision, Kahan's algorithm yields 0.0, whereas
    // Neumaier's algorithm yields the correct value 2.0.
    const petersData = [1.0, +1e100, 1.0, -1e100];

    // https://github.com/numpy/numpy/issues/8786
    // A badly conditioned sum, condition number ~2.188e+14.
    const numpyData = [
        -0.41253261766461263,
        41287272281118.43,
        -1.4727977348624173e-14,
        5670.3302557520055,
        2.119245229045646e-11,
        -0.003679264134906428,
        -6.892634568678797e-14,
        -0.0006984744181630712,
        -4054136.048352595,
        -1003.101760720037,
        -1.4436349910427172e-17,
        -41287268231649.57];
    const numpyExpected = -0.377392919181026;

    // A sequence where the second-level correction is non-zero at more
    // than one step, so it must be accumulated (ccs += cc), not
    // overwritten (ccs = cc).  Overwriting yields -1.0000000000000001e-16.
    const kleinData = [1e-16, -1e16, 1.0, 1e-16, -1.0, -1e-16, -1e-32, 1e16];
    const kleinExpected = 9.999999999999999e-17; // fsum(kleinData)

    // ── Tests ──────────────────────────────────────────────────────────

    it('initial value is zero', () => {
        expect(new KleinKbnAccumulator().value).toBe(0.0);
    });

    it('peters', () => {
        expect(naiveSum(petersData)).toBe(0.0);
        expect(kbnSum(petersData)).toBe(2.0);
    });

    it('numpy issue', () => {
        expect(kbnSum(numpyData)).toBeCloseTo(numpyExpected, 16);
        expect(naiveSum(numpyData)).not.toBeCloseTo(numpyExpected, 3);
    });

    it('second level correction is accumulated', () => {
        expect(kbnSum(kleinData)).toBe(kleinExpected);
        expect(fsum(kleinData)).toBe(kleinExpected);
    });

    it('matches fsum on mixed magnitudes', () => {
        const rng = new Rng(42);
        for (let k = 0; k < 200; k++) {
            const data: number[] = [];
            for (let i = 0; i < 100; i++) {
                data.push(rng.uniform(-1.0, 1.0) * 10.0 ** rng.choice([-8, 0, 8]));
            }
            expect(kbnSum(data)).toBe(fsum(data));
        }
    });

    it('better accuracy than naive', () => {
        // Add and then subtract the same values, so the exact sum is 0.
        const rng = new Rng(42);
        const data: number[] = [];
        for (let i = 0; i < 100000; i++) {
            data.push(rng.uniform(0.0, 1e7));
        }
        const len = data.length;
        for (let i = 0; i < len; i++) {
            data.push(-data[i]);
        }
        const k = kbnSum(data);
        const v = naiveSum(data);
        expect(k).toBe(0.0);
        expect(v).not.toBe(0.0);
    });

    it('update zero keeps compensation', () => {
        const kbn = new KleinKbnAccumulator();
        for (const x of kleinData) {
            kbn.update(x);
        }
        kbn.update(0.0);
        expect(kbn.value).toBe(kleinExpected);
    });

    it('revert', () => {
        const kbn = new KleinKbnAccumulator();
        kbn.update(1.5);
        kbn.update(2.5);
        kbn.revert(2.5);
        expect(kbn.value).toBe(1.5);
        kbn.revert(1.5);
        expect(kbn.value).toBe(0.0);
    });

    it('revert not most recent', () => {
        const kbn = new KleinKbnAccumulator();
        for (const x of petersData) {
            kbn.update(x);
        }
        kbn.revert(1e100); // not the most recent value
        expect(kbn.value).toBe(2.0 - 1e100);
        kbn.revert(-1e100);
        expect(kbn.value).toBe(2.0);
    });

    it('revert restores compensated sum', () => {
        const kbn = new KleinKbnAccumulator();
        for (const x of numpyData) {
            kbn.update(x);
        }
        kbn.update(1e20);
        kbn.revert(1e20);
        expect(kbn.value).toBeCloseTo(numpyExpected, 16);
    });

    it('set', () => {
        const kbn = new KleinKbnAccumulator();
        for (const x of petersData) {
            kbn.update(x);
        }
        kbn.set(5.0);
        expect(kbn.value).toBe(5.0);
        // Compensation terms are cleared, so only the new values count.
        kbn.update(1e100);
        kbn.update(-1e100);
        expect(kbn.value).toBe(5.0);
    });

    it('reset', () => {
        const kbn = new KleinKbnAccumulator();
        for (const x of petersData) {
            kbn.update(x);
        }
        kbn.reset();
        expect(kbn.value).toBe(0.0);
        kbn.update(1.5);
        expect(kbn.value).toBe(1.5);
    });
});
