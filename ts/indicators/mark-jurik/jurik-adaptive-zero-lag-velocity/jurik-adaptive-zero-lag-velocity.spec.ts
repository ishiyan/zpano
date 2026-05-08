import { } from 'jasmine';

import { JurikAdaptiveZeroLagVelocity } from './jurik-adaptive-zero-lag-velocity';
import {
    testInput,
    expectedLo2Hi15, expectedLo2Hi30, expectedLo2Hi60,
    expectedLo5Hi15, expectedLo5Hi30, expectedLo5Hi60,
    expectedLo10Hi15, expectedLo10Hi30, expectedLo10Hi60,
    expectedSens05, expectedSens25, expectedSens50,
    expectedPeriod15, expectedPeriod100, expectedPeriod300,
} from './testdata';

function almostEqual(a: number, b: number, epsilon: number): boolean {
    if (isNaN(a) && isNaN(b)) { return true; }
    if (isNaN(a) || isNaN(b)) { return false; }
    return Math.abs(a - b) < epsilon;
}

const epsilon = 1e-13;

describe('JurikAdaptiveZeroLagVelocity', () => {
    const cases: { loLength: number; hiLength: number; sensitivity: number; period: number; expected: number[] }[] = [
        { loLength: 2, hiLength: 15, sensitivity: 1.0, period: 3.0, expected: expectedLo2Hi15 },
        { loLength: 2, hiLength: 30, sensitivity: 1.0, period: 3.0, expected: expectedLo2Hi30 },
        { loLength: 2, hiLength: 60, sensitivity: 1.0, period: 3.0, expected: expectedLo2Hi60 },
        { loLength: 5, hiLength: 15, sensitivity: 1.0, period: 3.0, expected: expectedLo5Hi15 },
        { loLength: 5, hiLength: 30, sensitivity: 1.0, period: 3.0, expected: expectedLo5Hi30 },
        { loLength: 5, hiLength: 60, sensitivity: 1.0, period: 3.0, expected: expectedLo5Hi60 },
        { loLength: 10, hiLength: 15, sensitivity: 1.0, period: 3.0, expected: expectedLo10Hi15 },
        { loLength: 10, hiLength: 30, sensitivity: 1.0, period: 3.0, expected: expectedLo10Hi30 },
        { loLength: 10, hiLength: 60, sensitivity: 1.0, period: 3.0, expected: expectedLo10Hi60 },
        { loLength: 5, hiLength: 30, sensitivity: 0.5, period: 3.0, expected: expectedSens05 },
        { loLength: 5, hiLength: 30, sensitivity: 2.5, period: 3.0, expected: expectedSens25 },
        { loLength: 5, hiLength: 30, sensitivity: 5.0, period: 3.0, expected: expectedSens50 },
        { loLength: 5, hiLength: 30, sensitivity: 1.0, period: 1.5, expected: expectedPeriod15 },
        { loLength: 5, hiLength: 30, sensitivity: 1.0, period: 10.0, expected: expectedPeriod100 },
        { loLength: 5, hiLength: 30, sensitivity: 1.0, period: 30.0, expected: expectedPeriod300 },
    ];

    for (const tc of cases) {
        it(`should compute JAVEL with lo=${tc.loLength} hi=${tc.hiLength} sens=${tc.sensitivity} period=${tc.period}`, () => {
            const ind = new JurikAdaptiveZeroLagVelocity({
                loLength: tc.loLength, hiLength: tc.hiLength,
                sensitivity: tc.sensitivity, period: tc.period,
            });
            for (let i = 0; i < testInput.length; i++) {
                const result = ind.update(testInput[i]);
                expect(almostEqual(result, tc.expected[i], epsilon))
                    .withContext(`bar ${i}: got ${result}, want ${tc.expected[i]}`)
                    .toBe(true);
            }
        });
    }
});
