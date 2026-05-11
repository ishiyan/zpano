import { signalAnd, signalOr, signalNot, signalStrength } from './compose.ts';

describe('signalAnd', () => {
    it('all high', () => { expect(signalAnd(0.9, 0.8, 0.95)).toBeCloseTo(0.9 * 0.8 * 0.95, 10); });
    it('one zero', () => { expect(signalAnd(0.9, 0.0, 0.8)).toBeCloseTo(0.0, 10); });
    it('all one', () => { expect(signalAnd(1.0, 1.0, 1.0)).toBeCloseTo(1.0, 10); });
    it('two args', () => { expect(signalAnd(0.6, 0.7)).toBeCloseTo(0.42, 10); });
});

describe('signalOr', () => {
    it('both high', () => { expect(signalOr(0.8, 0.9)).toBeCloseTo(0.8 + 0.9 - 0.8 * 0.9, 10); });
    it('one zero', () => { expect(signalOr(0.0, 0.7)).toBeCloseTo(0.7, 10); });
    it('both zero', () => { expect(signalOr(0.0, 0.0)).toBeCloseTo(0.0, 10); });
    it('both one', () => { expect(signalOr(1.0, 1.0)).toBeCloseTo(1.0, 10); });
    it('greater than either', () => {
        const a = 0.6, b = 0.7;
        expect(signalOr(a, b)).toBeGreaterThanOrEqual(Math.max(a, b));
    });
});

describe('signalNot', () => {
    it('zero', () => { expect(signalNot(0.0)).toBeCloseTo(1.0, 10); });
    it('one', () => { expect(signalNot(1.0)).toBeCloseTo(0.0, 10); });
    it('half', () => { expect(signalNot(0.5)).toBeCloseTo(0.5, 10); });
    it('complement', () => {
        for (const v of [0.0, 0.3, 0.5, 0.7, 1.0]) {
            expect(signalNot(v)).toBeCloseTo(1.0 - v, 10);
        }
    });
});

describe('signalStrength', () => {
    it('above threshold', () => { expect(signalStrength(0.8, 0.5)).toBe(0.8); });
    it('below threshold', () => { expect(signalStrength(0.3, 0.5)).toBe(0.0); });
    it('at threshold', () => { expect(signalStrength(0.5, 0.5)).toBe(0.5); });
    it('just below', () => { expect(signalStrength(0.499, 0.5)).toBe(0.0); });
    it('default threshold', () => {
        expect(signalStrength(0.6)).toBe(0.6);
        expect(signalStrength(0.4)).toBe(0.0);
    });
});
