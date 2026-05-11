import {
    muLess, muLessEqual, muGreater, muGreaterEqual,
    muNear, muDirection, MembershipShape
} from './membership.ts';

describe('muLess', () => {
    // Sigmoid shape (default)
    it('crossover at threshold', () => {
        expect(muLess(10.0, 10.0, 2.0)).toBeCloseTo(0.5, 10);
    });
    it('well below threshold', () => {
        expect(muLess(8.0, 10.0, 2.0)).toBeGreaterThan(0.99);
    });
    it('well above threshold', () => {
        expect(muLess(12.0, 10.0, 2.0)).toBeLessThan(0.01);
    });
    it('monotonically decreasing', () => {
        const vals = [8.0, 9.0, 10.0, 11.0, 12.0].map(x => muLess(x, 10.0, 2.0));
        for (let i = 0; i < vals.length - 1; i++) {
            expect(vals[i]).toBeGreaterThan(vals[i + 1]);
        }
    });
    it('symmetry', () => {
        const below = muLess(9.0, 10.0, 2.0);
        const above = muLess(11.0, 10.0, 2.0);
        expect(below + above).toBeCloseTo(1.0, 10);
    });

    // Linear shape
    it('linear crossover', () => {
        expect(muLess(10.0, 10.0, 4.0, MembershipShape.LINEAR)).toBeCloseTo(0.5, 10);
    });
    it('linear below range', () => {
        expect(muLess(7.0, 10.0, 4.0, MembershipShape.LINEAR)).toBe(1.0);
    });
    it('linear above range', () => {
        expect(muLess(13.0, 10.0, 4.0, MembershipShape.LINEAR)).toBe(0.0);
    });
    it('linear midpoint', () => {
        expect(muLess(9.0, 10.0, 4.0, MembershipShape.LINEAR)).toBeCloseTo(0.75, 10);
    });

    // Crisp (width=0)
    it('crisp below', () => { expect(muLess(9.0, 10.0, 0.0)).toBe(1.0); });
    it('crisp above', () => { expect(muLess(11.0, 10.0, 0.0)).toBe(0.0); });
    it('crisp at threshold', () => { expect(muLess(10.0, 10.0, 0.0)).toBe(0.5); });

    it('less equal same as less', () => {
        expect(muLessEqual(9.5, 10.0, 2.0)).toBe(muLess(9.5, 10.0, 2.0));
    });
});

describe('muGreater', () => {
    it('complement of less', () => {
        for (const x of [8.0, 9.0, 10.0, 11.0, 12.0]) {
            expect(muGreater(x, 10.0, 2.0) + muLess(x, 10.0, 2.0)).toBeCloseTo(1.0, 10);
        }
    });
    it('crossover', () => {
        expect(muGreater(10.0, 10.0, 2.0)).toBeCloseTo(0.5, 10);
    });
    it('well above', () => { expect(muGreater(12.0, 10.0, 2.0)).toBeGreaterThan(0.99); });
    it('well below', () => { expect(muGreater(8.0, 10.0, 2.0)).toBeLessThan(0.01); });
    it('greater equal complement', () => {
        expect(muGreaterEqual(9.5, 10.0, 2.0) + muLessEqual(9.5, 10.0, 2.0)).toBeCloseTo(1.0, 10);
    });
});

describe('muNear', () => {
    it('peak at target', () => {
        expect(muNear(10.0, 10.0, 2.0)).toBeCloseTo(1.0, 10);
    });
    it('falls off', () => {
        expect(muNear(12.0, 10.0, 2.0)).toBeLessThan(0.05);
    });
    it('symmetric', () => {
        const below = muNear(9.0, 10.0, 2.0);
        const above = muNear(11.0, 10.0, 2.0);
        expect(below).toBeCloseTo(above, 10);
    });
    it('monotonic from center', () => {
        const vals = [0, 0.5, 1.0, 1.5, 2.0].map(d => muNear(10.0 + d, 10.0, 2.0));
        for (let i = 0; i < vals.length - 1; i++) {
            expect(vals[i]).toBeGreaterThan(vals[i + 1]);
        }
    });
    it('linear peak', () => {
        expect(muNear(10.0, 10.0, 2.0, MembershipShape.LINEAR)).toBeCloseTo(1.0, 10);
    });
    it('linear at boundary', () => {
        expect(muNear(12.0, 10.0, 2.0, MembershipShape.LINEAR)).toBe(0.0);
    });
    it('linear midpoint', () => {
        expect(muNear(11.0, 10.0, 2.0, MembershipShape.LINEAR)).toBeCloseTo(0.5, 10);
    });
    it('crisp exact', () => { expect(muNear(10.0, 10.0, 0.0)).toBe(1.0); });
    it('crisp any distance', () => { expect(muNear(10.1, 10.0, 0.0)).toBe(0.0); });
});

describe('muDirection', () => {
    it('large white body', () => { expect(muDirection(100.0, 110.0, 5.0)).toBeGreaterThan(0.95); });
    it('large black body', () => { expect(muDirection(110.0, 100.0, 5.0)).toBeLessThan(-0.95); });
    it('doji', () => { expect(muDirection(100.0, 100.0, 5.0)).toBeCloseTo(0.0, 10); });
    it('tiny white body', () => {
        const d = muDirection(100.0, 100.1, 5.0);
        expect(d).toBeGreaterThan(0.0);
        expect(d).toBeLessThan(0.1);
    });
    it('antisymmetric', () => {
        const d1 = muDirection(100.0, 105.0, 5.0);
        const d2 = muDirection(105.0, 100.0, 5.0);
        expect(d1).toBeCloseTo(-d2, 10);
    });
    it('zero body avg white', () => { expect(muDirection(100.0, 101.0, 0.0)).toBe(1.0); });
    it('zero body avg black', () => { expect(muDirection(101.0, 100.0, 0.0)).toBe(-1.0); });
    it('zero body avg doji', () => { expect(muDirection(100.0, 100.0, 0.0)).toBe(1.0); });
    it('range bounded', () => {
        for (const [o, c, avg] of [[0, 1000, 1], [1000, 0, 1], [50, 50, 100]]) {
            const d = muDirection(o, c, avg);
            expect(d).toBeGreaterThanOrEqual(-1.0);
            expect(d).toBeLessThanOrEqual(1.0);
        }
    });
});

describe('edge cases', () => {
    it('very large x', () => { expect(muLess(1e10, 0.0, 1.0)).toBe(0.0); });
    it('very small x', () => { expect(muLess(-1e10, 0.0, 1.0)).toBe(1.0); });
    it('tiny width', () => { expect(muLess(9.999, 10.0, 0.001)).toBeGreaterThan(0.99); });
    it('huge width', () => {
        const val = muLess(0.0, 10.0, 1000.0);
        expect(val).toBeGreaterThan(0.49);
        expect(val).toBeLessThan(0.60);
    });
});
