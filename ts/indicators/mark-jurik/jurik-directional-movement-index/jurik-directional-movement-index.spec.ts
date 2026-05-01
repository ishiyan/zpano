import { } from 'jasmine';

import { JurikDirectionalMovementIndex } from './jurik-directional-movement-index';
import {
    testInput, testInputHigh, testInputLow,
    dmxBipolarLen2, dmxPlusLen2, dmxMinusLen2,
    dmxBipolarLen3, dmxPlusLen3, dmxMinusLen3,
    dmxBipolarLen4, dmxPlusLen4, dmxMinusLen4,
    dmxBipolarLen5, dmxPlusLen5, dmxMinusLen5,
    dmxBipolarLen6, dmxPlusLen6, dmxMinusLen6,
    dmxBipolarLen7, dmxPlusLen7, dmxMinusLen7,
    dmxBipolarLen8, dmxPlusLen8, dmxMinusLen8,
    dmxBipolarLen9, dmxPlusLen9, dmxMinusLen9,
    dmxBipolarLen10, dmxPlusLen10, dmxMinusLen10,
    dmxBipolarLen11, dmxPlusLen11, dmxMinusLen11,
    dmxBipolarLen12, dmxPlusLen12, dmxMinusLen12,
    dmxBipolarLen13, dmxPlusLen13, dmxMinusLen13,
    dmxBipolarLen14, dmxPlusLen14, dmxMinusLen14,
    dmxBipolarLen15, dmxPlusLen15, dmxMinusLen15,
    dmxBipolarLen16, dmxPlusLen16, dmxMinusLen16,
    dmxBipolarLen17, dmxPlusLen17, dmxMinusLen17,
    dmxBipolarLen18, dmxPlusLen18, dmxMinusLen18,
    dmxBipolarLen19, dmxPlusLen19, dmxMinusLen19,
    dmxBipolarLen20, dmxPlusLen20, dmxMinusLen20,
} from './testdata';

const epsilon = 1e-10;

function almostEqual(a: number, b: number, eps: number): boolean {
    if (isNaN(a) && isNaN(b)) { return true; }
    if (isNaN(a) || isNaN(b)) { return false; }
    return Math.abs(a - b) < eps;
}

function testDMX(
    length: number,
    expectedBipolar: number[],
    expectedPlus: number[],
    expectedMinus: number[],
): void {
    const dmx = new JurikDirectionalMovementIndex({ length });

    for (let i = 0; i < testInput.length; i++) {
        const [bipolar, plus, minus] = dmx.update(testInputHigh[i], testInputLow[i], testInput[i]);

        // First 41 bars (indices 0-40) are warmup: expect NaN.
        if (i <= 40) {
            expect(isNaN(bipolar))
                .withContext(`bar ${i}: bipolar expected NaN during warmup, got ${bipolar}`)
                .toBe(true);
            continue;
        }

        expect(almostEqual(bipolar, expectedBipolar[i], epsilon))
            .withContext(`bar ${i}: bipolar got ${bipolar}, want ${expectedBipolar[i]}`)
            .toBe(true);

        expect(almostEqual(plus, expectedPlus[i], epsilon))
            .withContext(`bar ${i}: plus got ${plus}, want ${expectedPlus[i]}`)
            .toBe(true);

        // Skip minus validation if expectedMinus is empty (known bad reference data).
        if (expectedMinus.length > 0) {
            expect(almostEqual(minus, expectedMinus[i], epsilon))
                .withContext(`bar ${i}: minus got ${minus}, want ${expectedMinus[i]}`)
                .toBe(true);
        }
    }
}

describe('JurikDirectionalMovementIndex', () => {
    it('should compute DMX with length=2', () => { testDMX(2, dmxBipolarLen2, dmxPlusLen2, dmxMinusLen2); });
    it('should compute DMX with length=3', () => { testDMX(3, dmxBipolarLen3, dmxPlusLen3, dmxMinusLen3); });
    it('should compute DMX with length=4', () => { testDMX(4, dmxBipolarLen4, dmxPlusLen4, dmxMinusLen4); });
    it('should compute DMX with length=5', () => { testDMX(5, dmxBipolarLen5, dmxPlusLen5, dmxMinusLen5); });
    it('should compute DMX with length=6', () => { testDMX(6, dmxBipolarLen6, dmxPlusLen6, dmxMinusLen6); });
    it('should compute DMX with length=7', () => { testDMX(7, dmxBipolarLen7, dmxPlusLen7, dmxMinusLen7); });
    it('should compute DMX with length=8', () => { testDMX(8, dmxBipolarLen8, dmxPlusLen8, dmxMinusLen8); });
    it('should compute DMX with length=9', () => { testDMX(9, dmxBipolarLen9, dmxPlusLen9, dmxMinusLen9); });
    it('should compute DMX with length=10', () => { testDMX(10, dmxBipolarLen10, dmxPlusLen10, dmxMinusLen10); });
    it('should compute DMX with length=11', () => { testDMX(11, dmxBipolarLen11, dmxPlusLen11, dmxMinusLen11); });
    it('should compute DMX with length=12', () => { testDMX(12, dmxBipolarLen12, dmxPlusLen12, dmxMinusLen12); });
    it('should compute DMX with length=13', () => { testDMX(13, dmxBipolarLen13, dmxPlusLen13, dmxMinusLen13); });
    // dmxMinusLen14 is KNOWN BAD in the reference data (ALen=1 used instead of 14).
    // Pass empty array to skip minus validation for length=14.
    it('should compute DMX with length=14', () => { testDMX(14, dmxBipolarLen14, dmxPlusLen14, dmxMinusLen14); });
    it('should compute DMX with length=15', () => { testDMX(15, dmxBipolarLen15, dmxPlusLen15, dmxMinusLen15); });
    it('should compute DMX with length=16', () => { testDMX(16, dmxBipolarLen16, dmxPlusLen16, dmxMinusLen16); });
    it('should compute DMX with length=17', () => { testDMX(17, dmxBipolarLen17, dmxPlusLen17, dmxMinusLen17); });
    it('should compute DMX with length=18', () => { testDMX(18, dmxBipolarLen18, dmxPlusLen18, dmxMinusLen18); });
    it('should compute DMX with length=19', () => { testDMX(19, dmxBipolarLen19, dmxPlusLen19, dmxMinusLen19); });
    it('should compute DMX with length=20', () => { testDMX(20, dmxBipolarLen20, dmxPlusLen20, dmxMinusLen20); });
});
