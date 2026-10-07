import { KleinKbnSummator } from './klein-kbn-summator';

describe('KleinKbnSummator', () => {

    // ── Helpers ────────────────────────────────────────────────────────

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

    // ── Tests ──────────────────────────────────────────────────────────

    it('empty', () => {
        const s = new KleinKbnSummator();
        expect(s.n).toBe(0);
        expect(s.value).toBe(0.0);
        expect(s.mean).toBeNaN();
    });

    it('update', () => {
        const s = new KleinKbnSummator();
        for (const x of [1.0, 2.0, 3.0, 4.0]) {
            s.update(x);
        }
        expect(s.n).toBe(4);
        expect(s.value).toBe(10.0);
        expect(s.mean).toBe(2.5);
    });

    it('zero is counted', () => {
        const s = new KleinKbnSummator();
        for (const x of [0.0, 3.0, 0.0]) {
            s.update(x);
        }
        expect(s.n).toBe(3);
        expect(s.value).toBe(3.0);
        expect(s.mean).toBe(1.0);
    });

    it('compensated', () => {
        // Peters' example: naive summation yields 0.0.
        const s = new KleinKbnSummator();
        for (const x of [1.0, 1e100, 1.0, -1e100]) {
            s.update(x);
        }
        expect(s.value).toBe(2.0);
        expect(s.mean).toBe(0.5);
    });

    it('revert', () => {
        const s = new KleinKbnSummator();
        for (const x of [1.0, 2.0, 0.0, 4.0]) {
            s.update(x);
        }
        s.revert(0.0);
        expect(s.n).toBe(3);
        expect(s.value).toBe(7.0);
        s.revert(1.0); // not the most recent value
        expect(s.n).toBe(2);
        expect(s.value).toBe(6.0);
        expect(s.mean).toBe(3.0);
    });

    it('rolling window', () => {
        const data = [0.003, 0.026, 0.011, -0.010, 0.015, 0.025, 0.016, 0.067];
        const w = 3;
        const s = new KleinKbnSummator();
        data.forEach((x, i) => {
            s.update(x);
            if (i >= w) {
                s.revert(data[i - w]);
            }
            const window = data.slice(Math.max(0, i - w + 1), i + 1);
            expect(s.n).toBe(window.length);
            expect(s.value).toBeCloseTo(fsum(window), 17);
        });
    });

    it('revert to empty', () => {
        const s = new KleinKbnSummator();
        s.update(5.0);
        s.revert(5.0);
        expect(s.n).toBe(0);
        expect(s.value).toBe(0.0);
        expect(s.mean).toBeNaN();
    });

    it('revert empty raises', () => {
        const s = new KleinKbnSummator();
        expect(() => s.revert(1.0)).toThrowError('Cannot revert from an empty summator');
    });

    it('revert to empty clears compensation', () => {
        for (const finalZero of [false, true]) {
            const s = new KleinKbnSummator();
            for (const x of [0.1, 1e16, 1e32, 1e48]) {
                s.update(x);
            }
            if (finalZero) {
                s.update(0.0);
            }
            for (const x of [1e16, 0.1, 1e32, 1e48]) {
                s.revert(x);
            }
            if (finalZero) {
                expect(s.n).toBe(1);
                s.revert(0.0);
            }
            expect(s.n).toBe(0);
            expect(s.value).toBe(0.0);
            expect(s.mean).toBeNaN();
            s.update(3.0);
            expect(s.n).toBe(1);
            expect(s.value).toBe(3.0);
            expect(s.mean).toBe(3.0);
        }
    });

    it('reset', () => {
        const s = new KleinKbnSummator();
        for (const x of [1.0, 2.0]) {
            s.update(x);
        }
        s.reset();
        expect(s.n).toBe(0);
        expect(s.value).toBe(0.0);
        expect(s.mean).toBeNaN();
        s.update(3.0);
        expect(s.mean).toBe(3.0);
    });
});
