import { alphaCut } from './defuzzify.ts';

describe('alphaCut', () => {
    it('strong bearish', () => { expect(alphaCut(-87.3)).toBe(-100); });
    it('weak bearish', () => { expect(alphaCut(-32.1)).toBe(0); });
    it('strong bullish', () => { expect(alphaCut(92.5)).toBe(100); });
    it('weak bullish', () => { expect(alphaCut(15.0)).toBe(0); });
    it('zero', () => { expect(alphaCut(0.0)).toBe(0); });
    it('strong confirmation', () => { expect(alphaCut(156.8)).toBe(200); });
    it('negative confirmation', () => { expect(alphaCut(-180.0)).toBe(-200); });
    it('high alpha filters more', () => { expect(alphaCut(-87.3, 0.9)).toBe(0); });
    it('high alpha passes strong', () => { expect(alphaCut(-95.0, 0.9)).toBe(-100); });
    it('low alpha passes more', () => { expect(alphaCut(-15.0, 0.1)).toBe(-100); });
    it('alpha zero passes all', () => { expect(alphaCut(-1.0, 0.0)).toBe(-100); });
    it('exactly at threshold', () => { expect(alphaCut(50.0, 0.5)).toBe(100); });
    it('just below threshold', () => { expect(alphaCut(49.9, 0.5)).toBe(0); });
    it('exactly 100', () => { expect(alphaCut(100.0)).toBe(100); });
    it('exactly minus 100', () => { expect(alphaCut(-100.0)).toBe(-100); });
    it('custom scale', () => { expect(alphaCut(-40.0, 0.5, 50.0)).toBe(-50); });
    it('invalid scale', () => { expect(alphaCut(-87.3, 0.5, 0.0)).toBe(0); });
});
