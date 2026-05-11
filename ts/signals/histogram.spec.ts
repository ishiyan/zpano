import { muTurnsPositive, muTurnsNegative } from './histogram.ts';

describe('muTurnsPositive', () => {
    it('clear turn positive', () => { expect(muTurnsPositive(-5.0, 5.0)).toBeCloseTo(1.0, 10); });
    it('stays positive', () => { expect(muTurnsPositive(3.0, 5.0)).toBeCloseTo(0.0, 10); });
    it('stays negative', () => { expect(muTurnsPositive(-5.0, -3.0)).toBeCloseTo(0.0, 10); });
    it('turns more negative', () => { expect(muTurnsPositive(5.0, -5.0)).toBeCloseTo(0.0, 10); });
    it('from zero', () => { expect(muTurnsPositive(0.0, 5.0)).toBeCloseTo(0.5, 10); });
    it('fuzzy near zero', () => {
        const result = muTurnsPositive(-0.5, 0.5, 2.0);
        expect(result).toBeGreaterThan(0.1);
        expect(result).toBeLessThan(0.95);
    });
    it('fuzzy width makes softer', () => {
        const narrow = muTurnsPositive(-1.0, 1.0, 0.5);
        const wide = muTurnsPositive(-1.0, 1.0, 10.0);
        expect(narrow).toBeGreaterThan(wide);
    });
});

describe('muTurnsNegative', () => {
    it('clear turn negative', () => { expect(muTurnsNegative(5.0, -5.0)).toBeCloseTo(1.0, 10); });
    it('stays negative', () => { expect(muTurnsNegative(-5.0, -3.0)).toBeCloseTo(0.0, 10); });
    it('stays positive', () => { expect(muTurnsNegative(3.0, 5.0)).toBeCloseTo(0.0, 10); });
    it('symmetry', () => {
        const tn = muTurnsNegative(3.0, -3.0, 1.0);
        const tp = muTurnsPositive(-3.0, 3.0, 1.0);
        expect(tn).toBeCloseTo(tp, 10);
    });
});
