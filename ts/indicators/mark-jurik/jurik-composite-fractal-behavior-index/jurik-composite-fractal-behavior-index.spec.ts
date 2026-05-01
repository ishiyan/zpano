import { } from 'jasmine';

import { JurikCompositeFractalBehaviorIndex } from './jurik-composite-fractal-behavior-index';
import {
    testInput,
    expectedType1Smooth2, expectedType1Smooth10, expectedType1Smooth50,
    expectedType2Smooth2, expectedType2Smooth10, expectedType2Smooth50,
    expectedType3Smooth2, expectedType3Smooth10, expectedType3Smooth50,
    expectedType4Smooth2, expectedType4Smooth10, expectedType4Smooth50,
} from './testdata';

function almostEqual(a: number, b: number, epsilon: number): boolean {
    if (isNaN(a) && isNaN(b)) { return true; }
    if (isNaN(a) || isNaN(b)) { return false; }
    return Math.abs(a - b) < epsilon;
}

const epsilon = 1e-13;

describe('JurikCompositeFractalBehaviorIndex', () => {
    const cases: { fractalType: number; smooth: number; expected: number[] }[] = [
        { fractalType: 1, smooth: 2, expected: expectedType1Smooth2 },
        { fractalType: 1, smooth: 10, expected: expectedType1Smooth10 },
        { fractalType: 1, smooth: 50, expected: expectedType1Smooth50 },
        { fractalType: 2, smooth: 2, expected: expectedType2Smooth2 },
        { fractalType: 2, smooth: 10, expected: expectedType2Smooth10 },
        { fractalType: 2, smooth: 50, expected: expectedType2Smooth50 },
        { fractalType: 3, smooth: 2, expected: expectedType3Smooth2 },
        { fractalType: 3, smooth: 10, expected: expectedType3Smooth10 },
        { fractalType: 3, smooth: 50, expected: expectedType3Smooth50 },
        { fractalType: 4, smooth: 2, expected: expectedType4Smooth2 },
        { fractalType: 4, smooth: 10, expected: expectedType4Smooth10 },
        { fractalType: 4, smooth: 50, expected: expectedType4Smooth50 },
    ];

    for (const tc of cases) {
        it(`should compute CFB with fractalType=${tc.fractalType}, smooth=${tc.smooth}`, () => {
            const cfb = new JurikCompositeFractalBehaviorIndex({ fractalType: tc.fractalType, smooth: tc.smooth });
            for (let i = 0; i < testInput.length - 1; i++) { // skip last bar (reference aux loop stops at len-2)
                const result = cfb.update(testInput[i]);
                expect(almostEqual(result, tc.expected[i], epsilon))
                    .withContext(`bar ${i}: got ${result}, want ${tc.expected[i]}`)
                    .toBe(true);
            }
        });
    }
});
