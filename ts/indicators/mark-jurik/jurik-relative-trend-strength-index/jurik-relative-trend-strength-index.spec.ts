import { } from 'jasmine';

import { JurikRelativeTrendStrengthIndex } from './jurik-relative-trend-strength-index';
import {
    testInput,
    expectedLength2, expectedLength3, expectedLength4, expectedLength5,
    expectedLength6, expectedLength7, expectedLength8, expectedLength9,
    expectedLength10, expectedLength11, expectedLength12, expectedLength13,
    expectedLength14, expectedLength15,
} from './testdata';

function almostEqual(a: number, b: number, epsilon: number): boolean {
    if (isNaN(a) && isNaN(b)) { return true; }
    if (isNaN(a) || isNaN(b)) { return false; }
    return Math.abs(a - b) < epsilon;
}

const epsilon = 1e-13;

describe('JurikRelativeTrendStrengthIndex', () => {

    it('should throw for invalid length', () => {
        expect(() => new JurikRelativeTrendStrengthIndex({ length: 1 })).toThrowError();
    });

    it('should compute RSX with length=2', () => {
        const rsx = new JurikRelativeTrendStrengthIndex({ length: 2 });
        for (let i = 0; i < testInput.length; i++) {
            const result = rsx.update(testInput[i]);
            expect(almostEqual(result, expectedLength2[i], epsilon))
                .withContext(`bar ${i}: got ${result}, want ${expectedLength2[i]}`)
                .toBe(true);
        }
    });

    it('should compute RSX with length=3', () => {
        const rsx = new JurikRelativeTrendStrengthIndex({ length: 3 });
        for (let i = 0; i < testInput.length; i++) {
            const result = rsx.update(testInput[i]);
            expect(almostEqual(result, expectedLength3[i], epsilon))
                .withContext(`bar ${i}: got ${result}, want ${expectedLength3[i]}`)
                .toBe(true);
        }
    });

    it('should compute RSX with length=4', () => {
        const rsx = new JurikRelativeTrendStrengthIndex({ length: 4 });
        for (let i = 0; i < testInput.length; i++) {
            const result = rsx.update(testInput[i]);
            expect(almostEqual(result, expectedLength4[i], epsilon))
                .withContext(`bar ${i}: got ${result}, want ${expectedLength4[i]}`)
                .toBe(true);
        }
    });

    it('should compute RSX with length=5', () => {
        const rsx = new JurikRelativeTrendStrengthIndex({ length: 5 });
        for (let i = 0; i < testInput.length; i++) {
            const result = rsx.update(testInput[i]);
            expect(almostEqual(result, expectedLength5[i], epsilon))
                .withContext(`bar ${i}: got ${result}, want ${expectedLength5[i]}`)
                .toBe(true);
        }
    });

    it('should compute RSX with length=6', () => {
        const rsx = new JurikRelativeTrendStrengthIndex({ length: 6 });
        for (let i = 0; i < testInput.length; i++) {
            const result = rsx.update(testInput[i]);
            expect(almostEqual(result, expectedLength6[i], epsilon))
                .withContext(`bar ${i}: got ${result}, want ${expectedLength6[i]}`)
                .toBe(true);
        }
    });

    it('should compute RSX with length=7', () => {
        const rsx = new JurikRelativeTrendStrengthIndex({ length: 7 });
        for (let i = 0; i < testInput.length; i++) {
            const result = rsx.update(testInput[i]);
            expect(almostEqual(result, expectedLength7[i], epsilon))
                .withContext(`bar ${i}: got ${result}, want ${expectedLength7[i]}`)
                .toBe(true);
        }
    });

    it('should compute RSX with length=8', () => {
        const rsx = new JurikRelativeTrendStrengthIndex({ length: 8 });
        for (let i = 0; i < testInput.length; i++) {
            const result = rsx.update(testInput[i]);
            expect(almostEqual(result, expectedLength8[i], epsilon))
                .withContext(`bar ${i}: got ${result}, want ${expectedLength8[i]}`)
                .toBe(true);
        }
    });

    it('should compute RSX with length=9', () => {
        const rsx = new JurikRelativeTrendStrengthIndex({ length: 9 });
        for (let i = 0; i < testInput.length; i++) {
            const result = rsx.update(testInput[i]);
            expect(almostEqual(result, expectedLength9[i], epsilon))
                .withContext(`bar ${i}: got ${result}, want ${expectedLength9[i]}`)
                .toBe(true);
        }
    });

    it('should compute RSX with length=10', () => {
        const rsx = new JurikRelativeTrendStrengthIndex({ length: 10 });
        for (let i = 0; i < testInput.length; i++) {
            const result = rsx.update(testInput[i]);
            expect(almostEqual(result, expectedLength10[i], epsilon))
                .withContext(`bar ${i}: got ${result}, want ${expectedLength10[i]}`)
                .toBe(true);
        }
    });

    it('should compute RSX with length=11', () => {
        const rsx = new JurikRelativeTrendStrengthIndex({ length: 11 });
        for (let i = 0; i < testInput.length; i++) {
            const result = rsx.update(testInput[i]);
            expect(almostEqual(result, expectedLength11[i], epsilon))
                .withContext(`bar ${i}: got ${result}, want ${expectedLength11[i]}`)
                .toBe(true);
        }
    });

    it('should compute RSX with length=12', () => {
        const rsx = new JurikRelativeTrendStrengthIndex({ length: 12 });
        for (let i = 0; i < testInput.length; i++) {
            const result = rsx.update(testInput[i]);
            expect(almostEqual(result, expectedLength12[i], epsilon))
                .withContext(`bar ${i}: got ${result}, want ${expectedLength12[i]}`)
                .toBe(true);
        }
    });

    it('should compute RSX with length=13', () => {
        const rsx = new JurikRelativeTrendStrengthIndex({ length: 13 });
        for (let i = 0; i < testInput.length; i++) {
            const result = rsx.update(testInput[i]);
            expect(almostEqual(result, expectedLength13[i], epsilon))
                .withContext(`bar ${i}: got ${result}, want ${expectedLength13[i]}`)
                .toBe(true);
        }
    });

    it('should compute RSX with length=14', () => {
        const rsx = new JurikRelativeTrendStrengthIndex({ length: 14 });
        for (let i = 0; i < testInput.length; i++) {
            const result = rsx.update(testInput[i]);
            expect(almostEqual(result, expectedLength14[i], epsilon))
                .withContext(`bar ${i}: got ${result}, want ${expectedLength14[i]}`)
                .toBe(true);
        }
    });

    it('should compute RSX with length=15', () => {
        const rsx = new JurikRelativeTrendStrengthIndex({ length: 15 });
        for (let i = 0; i < testInput.length; i++) {
            const result = rsx.update(testInput[i]);
            expect(almostEqual(result, expectedLength15[i], epsilon))
                .withContext(`bar ${i}: got ${result}, want ${expectedLength15[i]}`)
                .toBe(true);
        }
    });
});
