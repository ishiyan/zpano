import { KleinKbnAccumulator } from './klein-kbn-accumulator';

describe('KleinKbnAccumulator', () => {

    // ── Helpers ────────────────────────────────────────────────────────

    function almostEqual(a: number, b: number, eps: number): boolean {
        return Math.abs(a - b) < eps;
    }

    /** Mulberry32 deterministic PRNG. */
    class SimpleRng {
        private state: number;
        constructor(seed: number) { this.state = seed >>> 0; }
        next(): number {
            let z = (this.state += 0x6d2b79f5) >>> 0;
            z = Math.imul(z ^ (z >>> 15), z | 1);
            z ^= z + Math.imul(z ^ (z >>> 7), z | 61);
            return ((z ^ (z >>> 14)) >>> 0) / 4294967296;
        }
    }

    class NaiveSum {
        private _value = 0.0;
        get value(): number { return this._value; }
        update(x: number): void { this._value += x; }
    }

    // ── Tests ──────────────────────────────────────────────────────────

    it('should sum Peters example correctly (1.0, 1e100, 1.0, -1e100 → 2.0)', () => {
        const data = [1.0, 1e100, 1.0, -1e100];
        const kbn = new KleinKbnAccumulator();
        for (const x of data) kbn.update(x);
        expect(almostEqual(kbn.value, 2.0, 1e-15)).toBeTrue();
    });

    it('should sum badly-conditioned NumPy example', () => {
        const data = [
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
            -41287268231649.57,
        ];
        const expected = -0.377392919181026;
        const kbn = new KleinKbnAccumulator();
        for (const x of data) kbn.update(x);
        expect(kbn.value).toBeCloseTo(expected, 16);
    });

    it('should be more accurate than naive summation', () => {
        const spread = 1e7;
        const naive = new NaiveSum();
        const kbn = new KleinKbnAccumulator();

        const rng = new SimpleRng(42);
        for (let i = 0; i < 1_000_000; i++) {
            const x = rng.next() * spread;
            naive.update(x);
            kbn.update(x);
        }

        const rng2 = new SimpleRng(42);
        for (let i = 0; i < 1_000_000; i++) {
            const x = rng2.next() * spread;
            naive.update(-x);
            kbn.update(-x);
        }

        expect(Math.abs(kbn.value)).toBeLessThanOrEqual(Math.abs(naive.value));
    });

    it('should support revert', () => {
        const kbn = new KleinKbnAccumulator();
        expect(kbn.value).toBeCloseTo(0.0, 15);

        kbn.update(1.5);
        kbn.update(2.5);
        kbn.revert(2.5);
        expect(kbn.value).toBeCloseTo(1.5, 15);
        kbn.revert(1.5);
        expect(kbn.value).toBeCloseTo(0.0, 15);
    });

    it('should support reset', () => {
        const kbn = new KleinKbnAccumulator();
        kbn.update(1.5);
        kbn.reset();
        expect(kbn.value).toBeCloseTo(0.0, 15);

        kbn.update(1.5);
        expect(kbn.value).toBeCloseTo(1.5, 15);
    });
});
