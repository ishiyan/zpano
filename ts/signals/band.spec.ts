import { muAboveBand, muBelowBand, muBetweenBands } from './band.ts';

describe('muAboveBand', () => {
    it('well above', () => { expect(muAboveBand(110.0, 100.0, 5.0)).toBeCloseTo(1.0, 2); });
    it('well below', () => { expect(muAboveBand(90.0, 100.0, 5.0)).toBeCloseTo(0.0, 2); });
    it('at band', () => { expect(muAboveBand(100.0, 100.0, 5.0)).toBeCloseTo(0.5, 10); });
    it('crisp', () => {
        expect(muAboveBand(100.1, 100.0, 0.0)).toBe(1.0);
        expect(muAboveBand(99.9, 100.0, 0.0)).toBe(0.0);
    });
});

describe('muBelowBand', () => {
    it('well below', () => { expect(muBelowBand(85.0, 90.0, 5.0)).toBeCloseTo(1.0, 2); });
    it('well above', () => { expect(muBelowBand(100.0, 90.0, 5.0)).toBeCloseTo(0.0, 2); });
    it('at band', () => { expect(muBelowBand(90.0, 90.0, 5.0)).toBeCloseTo(0.5, 10); });
});

describe('muBetweenBands', () => {
    it('centered', () => { expect(muBetweenBands(100.0, 90.0, 110.0)).toBeGreaterThan(0.8); });
    it('at upper band', () => { expect(muBetweenBands(110.0, 90.0, 110.0)).toBeLessThan(0.6); });
    it('at lower band', () => { expect(muBetweenBands(90.0, 90.0, 110.0)).toBeLessThan(0.6); });
    it('outside above', () => { expect(muBetweenBands(130.0, 90.0, 110.0)).toBeLessThan(0.1); });
    it('outside below', () => { expect(muBetweenBands(70.0, 90.0, 110.0)).toBeLessThan(0.1); });
    it('degenerate bands', () => {
        expect(muBetweenBands(100.0, 110.0, 90.0)).toBe(0.0);
        expect(muBetweenBands(100.0, 100.0, 100.0)).toBe(0.0);
    });
    it('monotonic from center', () => {
        const center = muBetweenBands(100.0, 90.0, 110.0);
        const edge = muBetweenBands(108.0, 90.0, 110.0);
        const outside = muBetweenBands(115.0, 90.0, 110.0);
        expect(center).toBeGreaterThan(edge);
        expect(edge).toBeGreaterThan(outside);
    });
});
