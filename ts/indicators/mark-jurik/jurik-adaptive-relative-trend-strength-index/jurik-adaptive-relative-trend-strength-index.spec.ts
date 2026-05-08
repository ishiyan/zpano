import { } from 'jasmine';

import { JurikAdaptiveRelativeTrendStrengthIndex } from './jurik-adaptive-relative-trend-strength-index';
import {
    testInput,
    expectedLo2Hi15, expectedLo2Hi30, expectedLo2Hi60,
    expectedLo5Hi15, expectedLo5Hi30, expectedLo5Hi60,
    expectedLo10Hi15, expectedLo10Hi30, expectedLo10Hi60,
} from './testdata';

function almostEqual(a: number, b: number, epsilon: number): boolean {
    if (isNaN(a) && isNaN(b)) { return true; }
    if (isNaN(a) || isNaN(b)) { return false; }
    return Math.abs(a - b) < epsilon;
}

const epsilon = 1e-13;

describe('JurikAdaptiveRelativeTrendStrengthIndex', () => {
    const cases: { loLength: number; hiLength: number; expected: number[] }[] = [
        { loLength: 2, hiLength: 15, expected: expectedLo2Hi15 },
        { loLength: 2, hiLength: 30, expected: expectedLo2Hi30 },
        { loLength: 2, hiLength: 60, expected: expectedLo2Hi60 },
        { loLength: 5, hiLength: 15, expected: expectedLo5Hi15 },
        { loLength: 5, hiLength: 30, expected: expectedLo5Hi30 },
        { loLength: 5, hiLength: 60, expected: expectedLo5Hi60 },
        { loLength: 10, hiLength: 15, expected: expectedLo10Hi15 },
        { loLength: 10, hiLength: 30, expected: expectedLo10Hi30 },
        { loLength: 10, hiLength: 60, expected: expectedLo10Hi60 },
    ];

    for (const tc of cases) {
        it(`should compute JARSX with lo=${tc.loLength} hi=${tc.hiLength}`, () => {
            const ind = new JurikAdaptiveRelativeTrendStrengthIndex({ loLength: tc.loLength, hiLength: tc.hiLength });
            for (let i = 0; i < testInput.length; i++) {
                const result = ind.update(testInput[i]);
                expect(almostEqual(result, tc.expected[i], epsilon))
                    .withContext(`bar ${i}: got ${result}, want ${tc.expected[i]}`)
                    .toBe(true);
            }
        });
    }
});
