import { MembershipShape } from '../fuzzy/index.ts';
import { muAbove, muBelow, muOverbought, muOversold } from './threshold.ts';

describe('muAbove', () => {
    it('well above', () => { expect(muAbove(80.0, 70.0, 5.0)).toBeCloseTo(1.0, 2); });
    it('well below', () => { expect(muAbove(60.0, 70.0, 5.0)).toBeCloseTo(0.0, 2); });
    it('at threshold', () => { expect(muAbove(70.0, 70.0, 5.0)).toBeCloseTo(0.5, 10); });
    it('zero width above', () => { expect(muAbove(70.1, 70.0, 0.0)).toBe(1.0); });
    it('zero width below', () => { expect(muAbove(69.9, 70.0, 0.0)).toBe(0.0); });
    it('zero width equal', () => { expect(muAbove(70.0, 70.0, 0.0)).toBeCloseTo(0.5, 10); });
    it('monotonic', () => {
        const m1 = muAbove(68.0, 70.0, 5.0);
        const m2 = muAbove(70.0, 70.0, 5.0);
        const m3 = muAbove(72.0, 70.0, 5.0);
        expect(m1).toBeLessThan(m2);
        expect(m2).toBeLessThan(m3);
    });
    it('linear shape', () => {
        expect(muAbove(70.0, 70.0, 10.0, MembershipShape.LINEAR)).toBeCloseTo(0.5, 10);
        expect(muAbove(65.0, 70.0, 10.0, MembershipShape.LINEAR)).toBeCloseTo(0.0, 10);
        expect(muAbove(75.0, 70.0, 10.0, MembershipShape.LINEAR)).toBeCloseTo(1.0, 10);
    });
});

describe('muBelow', () => {
    it('well below', () => { expect(muBelow(20.0, 30.0, 5.0)).toBeCloseTo(1.0, 2); });
    it('well above', () => { expect(muBelow(40.0, 30.0, 5.0)).toBeCloseTo(0.0, 2); });
    it('at threshold', () => { expect(muBelow(30.0, 30.0, 5.0)).toBeCloseTo(0.5, 10); });
    it('complement of above', () => {
        for (const v of [25.0, 30.0, 35.0, 50.0]) {
            expect(muBelow(v, 30.0, 5.0) + muAbove(v, 30.0, 5.0)).toBeCloseTo(1.0, 10);
        }
    });
});

describe('overbought/oversold', () => {
    it('overbought high RSI', () => { expect(muOverbought(85.0)).toBeGreaterThan(0.95); });
    it('overbought low RSI', () => { expect(muOverbought(50.0)).toBeLessThan(0.01); });
    it('oversold low RSI', () => { expect(muOversold(15.0)).toBeGreaterThan(0.95); });
    it('oversold high RSI', () => { expect(muOversold(50.0)).toBeLessThan(0.01); });
    it('overbought custom level', () => { expect(muOverbought(80.0, 80.0)).toBeCloseTo(0.5, 10); });
    it('oversold custom level', () => { expect(muOversold(20.0, 20.0)).toBeCloseTo(0.5, 10); });
});
