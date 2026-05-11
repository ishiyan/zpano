import { muCrossesAbove, muCrossesBelow, muLineCrossesAbove, muLineCrossesBelow } from './crossover.ts';

describe('muCrossesAbove', () => {
    it('clear cross above', () => { expect(muCrossesAbove(25.0, 35.0, 30.0)).toBeCloseTo(1.0, 10); });
    it('no cross both above', () => { expect(muCrossesAbove(35.0, 40.0, 30.0)).toBeCloseTo(0.0, 10); });
    it('no cross both below', () => { expect(muCrossesAbove(25.0, 28.0, 30.0)).toBeCloseTo(0.0, 10); });
    it('cross down not up', () => { expect(muCrossesAbove(35.0, 25.0, 30.0)).toBeCloseTo(0.0, 10); });
    it('fuzzy near threshold', () => {
        const result = muCrossesAbove(29.0, 31.0, 30.0, 5.0);
        expect(result).toBeGreaterThan(0.1);
        expect(result).toBeLessThan(0.9);
    });
    it('at threshold', () => { expect(muCrossesAbove(30.0, 30.0, 30.0)).toBeCloseTo(0.25, 10); });
});

describe('muCrossesBelow', () => {
    it('clear cross below', () => { expect(muCrossesBelow(35.0, 25.0, 30.0)).toBeCloseTo(1.0, 10); });
    it('no cross both below', () => { expect(muCrossesBelow(25.0, 20.0, 30.0)).toBeCloseTo(0.0, 10); });
    it('symmetry', () => {
        const cb = muCrossesBelow(35.0, 25.0, 30.0, 2.0);
        const ca = muCrossesAbove(25.0, 35.0, 30.0, 2.0);
        expect(cb).toBeCloseTo(ca, 10);
    });
});

describe('muLineCrossesAbove', () => {
    it('golden cross', () => {
        expect(muLineCrossesAbove(49.0, 51.0, 50.0, 50.0)).toBeCloseTo(1.0, 10);
    });
    it('no cross', () => {
        expect(muLineCrossesAbove(52.0, 53.0, 50.0, 50.0)).toBeCloseTo(0.0, 10);
    });
    it('fuzzy near cross', () => {
        const result = muLineCrossesAbove(49.5, 50.5, 50.0, 50.0, 2.0);
        expect(result).toBeGreaterThan(0.0);
        expect(result).toBeLessThan(1.0);
    });
});

describe('muLineCrossesBelow', () => {
    it('death cross', () => {
        expect(muLineCrossesBelow(51.0, 49.0, 50.0, 50.0)).toBeCloseTo(1.0, 10);
    });
    it('no cross', () => {
        expect(muLineCrossesBelow(48.0, 47.0, 50.0, 50.0)).toBeCloseTo(0.0, 10);
    });
});
