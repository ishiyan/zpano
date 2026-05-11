import {
    tProduct, tMin, tLukasiewicz,
    sProbabilistic, sMax,
    fNot,
    tProductAll, tMinAll,
} from './operators.ts';

describe('t-norms', () => {
    it('product basic', () => { expect(tProduct(0.8, 0.6)).toBeCloseTo(0.48, 10); });
    it('product identity', () => { expect(tProduct(0.7, 1.0)).toBeCloseTo(0.7, 10); });
    it('product annihilator', () => { expect(tProduct(0.7, 0.0)).toBeCloseTo(0.0, 10); });
    it('product commutativity', () => { expect(tProduct(0.3, 0.8)).toBeCloseTo(tProduct(0.8, 0.3), 10); });
    it('min basic', () => { expect(tMin(0.8, 0.6)).toBe(0.6); });
    it('min identity', () => { expect(tMin(0.7, 1.0)).toBe(0.7); });
    it('min annihilator', () => { expect(tMin(0.7, 0.0)).toBe(0.0); });
    it('lukasiewicz both high', () => { expect(tLukasiewicz(0.9, 0.8)).toBeCloseTo(0.7, 10); });
    it('lukasiewicz one low', () => { expect(tLukasiewicz(0.3, 0.5)).toBeCloseTo(0.0, 10); });
    it('lukasiewicz clamp', () => { expect(tLukasiewicz(0.1, 0.2)).toBe(0.0); });
    it('lukasiewicz identity', () => { expect(tLukasiewicz(0.7, 1.0)).toBeCloseTo(0.7, 10); });
});

describe('s-norms', () => {
    it('probabilistic basic', () => { expect(sProbabilistic(0.8, 0.6)).toBeCloseTo(0.92, 10); });
    it('probabilistic identity', () => { expect(sProbabilistic(0.7, 0.0)).toBeCloseTo(0.7, 10); });
    it('probabilistic annihilator', () => { expect(sProbabilistic(0.7, 1.0)).toBeCloseTo(1.0, 10); });
    it('max basic', () => { expect(sMax(0.8, 0.6)).toBe(0.8); });
    it('max identity', () => { expect(sMax(0.7, 0.0)).toBe(0.7); });
});

describe('negation', () => {
    it('not basic', () => { expect(fNot(0.3)).toBeCloseTo(0.7, 10); });
    it('not zero', () => { expect(fNot(0.0)).toBeCloseTo(1.0, 10); });
    it('not one', () => { expect(fNot(1.0)).toBeCloseTo(0.0, 10); });
    it('double negation', () => { expect(fNot(fNot(0.4))).toBeCloseTo(0.4, 10); });
});

describe('variadic', () => {
    it('product all three', () => { expect(tProductAll(0.8, 0.6, 0.5)).toBeCloseTo(0.24, 10); });
    it('product all single', () => { expect(tProductAll(0.7)).toBeCloseTo(0.7, 10); });
    it('product all empty', () => { expect(tProductAll()).toBeCloseTo(1.0, 10); });
    it('min all three', () => { expect(tMinAll(0.8, 0.6, 0.9)).toBe(0.6); });
    it('min all empty', () => { expect(tMinAll()).toBe(1.0); });
    it('product all five', () => {
        expect(tProductAll(0.9, 0.9, 0.9, 0.9, 0.9)).toBeCloseTo(Math.pow(0.9, 5), 10);
    });
});

describe('duality', () => {
    it('product/probabilistic De Morgan', () => {
        const a = 0.7, b = 0.4;
        expect(tProduct(a, b)).toBeCloseTo(fNot(sProbabilistic(fNot(a), fNot(b))), 10);
    });
    it('min/max De Morgan', () => {
        const a = 0.7, b = 0.4;
        expect(tMin(a, b)).toBeCloseTo(fNot(sMax(fNot(a), fNot(b))), 10);
    });
});
